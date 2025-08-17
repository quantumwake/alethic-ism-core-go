package user

import (
	"context"
	"time"

	"github.com/quantumwake/alethic-ism-core-go/pkg/cache"
)

// CachedBackendStorage provides a caching layer over the user BackendStorage.
// User profiles don't change frequently, making them excellent candidates for caching.
// This implementation caches read operations and invalidates on updates.
type CachedBackendStorage struct {
	*cache.CachedBackend // Embedded generic caching functionality
	base *BackendStorage  // The underlying user backend
}

// NewCachedBackend creates a new user backend with caching enabled.
// By default, cache entries expire after 10 minutes (longer than other backends
// because user data changes less frequently).
//
// Parameters:
//   - dsn: Database connection string
//   - c: Cache implementation (pass nil for default in-memory cache)
//
// Returns:
//   - A new CachedBackendStorage instance with caching enabled
func NewCachedBackend(dsn string, c cache.Cache) *CachedBackendStorage {
	base := NewBackend(dsn)
	// User profiles are relatively static, so use a longer TTL
	cachedBackend := cache.NewCachedBackend(base, c, 10*time.Minute)

	return &CachedBackendStorage{
		CachedBackend: cachedBackend,
		base:          base,
	}
}

// NewCachedBackendWithTTL creates a new user backend with custom cache TTL.
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

// FindUserByID retrieves a user profile by ID with caching.
// User profiles are frequently accessed for authentication and authorization,
// making caching highly beneficial.
// Cache key format: "FindUserByID:hash(id)"
//
// Parameters:
//   - id: The user ID to find
//
// Returns:
//   - *User: The user profile if found
//   - error: Database error if the operation fails
func (cb *CachedBackendStorage) FindUserByID(id string) (*User, error) {
	ctx := context.Background()

	// User profiles are stable data, could use even longer TTL if needed
	return cache.CallCachedWithTTL(cb.CachedBackend, ctx, "FindUserByID",
		[]interface{}{id},
		15*time.Minute, // Slightly longer TTL for user profiles
		func() (*User, error) {
			return cb.base.FindUserByID(id)
		})
}

// CreateOrUpdate creates or updates a user profile and invalidates the cache.
// This ensures that subsequent reads will get the updated data from the database.
//
// Parameters:
//   - user: The user profile to create or update
//
// Returns:
//   - error: Database error if the operation fails
//
// Cache invalidation:
//   - Invalidates FindUserByID cache for the affected user
func (cb *CachedBackendStorage) CreateOrUpdate(user *User) error {
	// Perform the database operation first
	err := cb.base.CreateOrUpdate(user)
	if err != nil {
		return err
	}

	ctx := context.Background()

	// Invalidate the cache for this user
	// This ensures the next read will get fresh data from the database
	_ = cb.InvalidateMethod(ctx, "FindUserByID", user.ID)

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