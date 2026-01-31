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

	// GetCreateOrUpdate retrieves a value from the cache by key, or fetches it using fetchFunc if not present or expired.
	// It does this in a thread-safe manner to prevent cache stampedes.
	GetCreateOrUpdate(ctx context.Context, key string, fetchFunc func(exists bool, value any) (any, error), ttl time.Duration) (any, error)

	// Set stores a value in the cache with the specified key and time-to-live (TTL).
	// If ttl is 0, the implementation should use a default TTL.
	Set(ctx context.Context, key string, value interface{}, ttl time.Duration)

	// Delete removes a specific key from the cache.
	// Returns nil even if the key doesn't exist.
	Delete(ctx context.Context, key string)

	// Clear removes all entries from the cache.
	// Use with caution as this affects all cached data.
	Clear(ctx context.Context) error

	// GetDefaultTTL gets the globally defined default TTL of cache items.
	GetDefaultTTL() time.Duration

	// Close closes any connections and stops any background eviction loops, if any.
	Close()
}

// Config holds configuration parameters for cache implementations.
// This struct can be extended with additional fields as needed (e.g., max size, eviction policy).
type Config struct {
	// DefaultTTL is the default time-to-live for cache entries when no specific TTL is provided.
	// After this duration, entries are considered expired and will be evicted.
	DefaultTTL time.Duration

	// CleanupDurationInterval defines how often the cache should perform cleanup of expired entries.
	CleanupDurationInterval time.Duration
}

// NewDefaultConfig creates a CacheConfig with sensible defaults.
// Default TTL is add to 5 minutes, which provides a good balance between
// reducing database load and ensuring data freshness.
func NewDefaultConfig() *Config {
	return &Config{
		DefaultTTL:              5 * time.Minute,
		CleanupDurationInterval: 10 * time.Minute,
	}
}

// NewConfigWithTTL creates a CacheConfig with a specified default TTL.
// This allows customization of cache expiration behavior based on application needs.
func NewConfigWithTTL(ttl time.Duration) *Config {
	return &Config{
		DefaultTTL:              ttl,
		CleanupDurationInterval: 10 * time.Minute,
	}
}
