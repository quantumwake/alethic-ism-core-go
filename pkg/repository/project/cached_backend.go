package project

import (
	"context"
	"time"

	"github.com/quantumwake/alethic-ism-core-go/pkg/cache"
)

// CachedBackendStorage provides a caching layer over the project BackendStorage.
// Projects are accessed frequently but change relatively infrequently,
// making them good candidates for caching with moderate TTLs.
type CachedBackendStorage struct {
	*cache.CachedBackend // Embedded generic caching functionality
	base *BackendStorage  // The underlying project backend
}

// NewCachedBackend creates a new project backend with caching enabled.
// By default, cache entries expire after 5 minutes.
//
// Parameters:
//   - dsn: Database connection string
//   - c: Cache implementation (pass nil for default in-memory cache)
//
// Returns:
//   - A new CachedBackendStorage instance with caching enabled
func NewCachedBackend(dsn string, c cache.Cache) *CachedBackendStorage {
	base := NewBackend(dsn)
	cachedBackend := cache.NewCachedBackend(base, c, 5*time.Minute)

	return &CachedBackendStorage{
		CachedBackend: cachedBackend,
		base:          base,
	}
}

// NewCachedBackendWithTTL creates a new project backend with custom cache TTL.
// Use this when you need different cache expiration times.
//
// Parameters:
//   - dsn: Database connection string
//   - c: Cache implementation (pass nil for default in-memory cache)
//   - ttl: Custom time-to-live for all cached entries
//
// Returns:
//   - A new CachedBackendStorage instance with custom TTL
func NewCachedBackendWithTTL(dsn string, c cache.Cache, ttl time.Duration) *CachedBackendStorage {
	base := NewBackend(dsn)
	cachedBackend := cache.NewCachedBackend(base, c, ttl)

	return &CachedBackendStorage{
		CachedBackend: cachedBackend,
		base:          base,
	}
}

// FindByID retrieves a project by ID with caching.
// Projects are frequently accessed for authorization and context,
// making caching beneficial for reducing database load.
// Cache key format: "FindByID:hash(id)"
//
// Parameters:
//   - id: The project ID to find
//
// Returns:
//   - *Project: The project if found
//   - error: Database error if the operation fails
func (cb *CachedBackendStorage) FindByID(id string) (*Project, error) {
	ctx := context.Background()

	return cache.CallCached(cb.CachedBackend, ctx, "FindByID",
		[]interface{}{id},
		func() (*Project, error) {
			return cb.base.FindByID(id)
		})
}

// FindAllByUserID retrieves all projects for a user with caching.
// User project lists are accessed frequently, especially during project selection
// and authorization checks.
// Cache key format: "FindAllByUserID:hash(userID)"
//
// Parameters:
//   - userID: The user ID to find projects for
//
// Returns:
//   - []Project: List of projects belonging to the user
//   - error: Database error if the operation fails
func (cb *CachedBackendStorage) FindAllByUserID(userID string) ([]Project, error) {
	ctx := context.Background()

	// User project lists change less frequently, so we can use a slightly longer TTL
	return cache.CallCachedWithTTL(cb.CachedBackend, ctx, "FindAllByUserID",
		[]interface{}{userID},
		7*time.Minute, // Slightly longer TTL for user project lists
		func() ([]Project, error) {
			return cb.base.FindAllByUserID(userID)
		})
}

// CreateOrUpdate creates or updates a project and invalidates related caches.
// This ensures that subsequent reads will get the updated data from the database.
//
// Parameters:
//   - project: The project to create or update
//
// Returns:
//   - error: Database error if the operation fails
//
// Cache invalidation:
//   - Invalidates FindByID cache for the project
//   - Invalidates FindAllByUserID cache for the user's project list
func (cb *CachedBackendStorage) CreateOrUpdate(project *Project) error {
	// Perform the database operation first
	err := cb.base.CreateOrUpdate(project)
	if err != nil {
		return err
	}

	ctx := context.Background()

	// Invalidate cache entries that would be affected by this change
	// 1. The specific project cache
	_ = cb.InvalidateMethod(ctx, "FindByID", project.ID)

	// 2. The user's project list cache (since a project was added/updated)
	_ = cb.InvalidateMethod(ctx, "FindAllByUserID", project.UserID)

	return nil
}

// Access returns the underlying BackendStorage for direct database access.
// Use this when you need to bypass the cache layer, for example:
//   - During data migrations
//   - For administrative operations
//   - When debugging cache issues
//   - For operations requiring immediate consistency
//
// Returns:
//   - *BackendStorage: The underlying database backend
func (cb *CachedBackendStorage) Access() *BackendStorage {
	return cb.base
}