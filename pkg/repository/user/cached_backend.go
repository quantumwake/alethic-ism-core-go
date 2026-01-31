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
	*cache.CachedBackend                 // Embedded generic caching functionality
	base                 *BackendStorage // The underlying user backend
}

// DefaultConfig returns the default TTL configuration for user backend.
func DefaultConfig(baseTTL time.Duration) *cache.MethodTTLConfig {
	config := cache.NewMethodTTLConfig(baseTTL)

	// User profiles are very stable
	config.SetMethodTTL("FindUserByID", 15*time.Minute)

	return config
}

// NewCachedBackend creates a new user backend with caching enabled.
// Uses the default user configuration with provided base TTL.
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

// NewCachedBackendWithConfig creates a user backend with custom TTL configuration.
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

	// Use the configured TTL for this method (set via MethodConfig)
	return cache.CallCached(cb.CachedBackend, ctx, "FindUserByID",
		[]interface{}{id},
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
