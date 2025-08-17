package cache

import (
	"context"
	"time"
)

// Cache defines the interface for cache implementations.
// This abstraction allows for easy swapping between different cache backends
// (e.g., in-memory, Redis, Memcached) without changing the consuming code.
type Cache interface {
	// Get retrieves a value from the cache by key.
	// Returns the cached value and true if found, nil and false if not found or expired.
	Get(ctx context.Context, key string) (interface{}, bool)
	
	// Set stores a value in the cache with the specified key and time-to-live (TTL).
	// If ttl is 0, the implementation should use a default TTL.
	Set(ctx context.Context, key string, value interface{}, ttl time.Duration) error
	
	// Delete removes a specific key from the cache.
	// Returns nil even if the key doesn't exist.
	Delete(ctx context.Context, key string) error
	
	// Clear removes all entries from the cache.
	// Use with caution as this affects all cached data.
	Clear(ctx context.Context) error
}

// CacheConfig holds configuration parameters for cache implementations.
// This struct can be extended with additional fields as needed (e.g., max size, eviction policy).
type CacheConfig struct {
	// DefaultTTL is the default time-to-live for cache entries when no specific TTL is provided.
	// After this duration, entries are considered expired and will be evicted.
	DefaultTTL time.Duration
}

// NewDefaultConfig creates a CacheConfig with sensible defaults.
// Default TTL is set to 5 minutes, which provides a good balance between
// reducing database load and ensuring data freshness.
func NewDefaultConfig() *CacheConfig {
	return &CacheConfig{
		DefaultTTL: 5 * time.Minute,
	}
}