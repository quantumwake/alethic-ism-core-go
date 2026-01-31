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
	*cache.CachedBackend                 // Embedded generic caching functionality
	base                 *BackendStorage // The underlying project backend
}

// DefaultConfig returns the default TTL configuration for project backend.
func DefaultConfig(baseTTL time.Duration) *cache.MethodTTLConfig {
	config := cache.NewMethodTTLConfig(baseTTL)

	// Projects change occasionally
	config.SetMethodTTL("FindByID", baseTTL)
	config.SetMethodTTL("FindAllByUserID", baseTTL+2*time.Minute) // Slightly longer for lists

	return config
}

// NewCachedBackend creates a new project backend with caching enabled.
// Uses the default project configuration with provided base TTL.
//
// Parameters:
//   - dsn: Database connection string
//   - c: Cache implementation
//   - baseTTL: Base TTL for cache entries
//
// Returns:
//   - A new CachedBackendStorage instance with caching enabled
func NewCachedBackend(dsn string, c cache.Cache, baseTTL time.Duration) *CachedBackendStorage {
	config := DefaultConfig(baseTTL)
	return NewCachedBackendWithConfig(dsn, c, config)
}

// NewCachedBackendWithConfig creates a project backend with custom TTL configuration.
// Use this for fine-grained control over method-specific cache TTLs.
//
// Parameters:
//   - dsn: Database connection string
//   - c: Cache implementation
//   - config: Method-specific TTL configuration
//
// Returns:
//   - A new CachedBackendStorage instance with configured TTLs
func NewCachedBackendWithConfig(dsn string, c cache.Cache, config *cache.MethodTTLConfig) *CachedBackendStorage {
	base := NewBackend(dsn)
	cachedBackend := cache.NewCachedBackend(base, c, config.DefaultTTL)

	// Apply method-specific TTL configuration
	config.ApplyToBackend(cachedBackend)

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

	// Use the configured TTL for this method (set via MethodConfig)
	return cache.CallCached(cb.CachedBackend, ctx, "FindAllByUserID",
		[]interface{}{userID},
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
