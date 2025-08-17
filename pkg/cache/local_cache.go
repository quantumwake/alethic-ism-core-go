package cache

import (
	"context"
	"sync"
	"time"
)

// cacheEntry represents a single cached item with its expiration time.
// This internal structure tracks both the cached value and when it should expire.
type cacheEntry struct {
	value      interface{} // The actual cached value
	expireTime time.Time   // When this entry expires and should be evicted
}

// LocalCache implements an in-memory cache with TTL support.
// It uses a map for O(1) lookups and a background goroutine for periodic cleanup.
// This implementation is thread-safe and suitable for single-instance applications.
// For distributed systems, consider using Redis or similar distributed cache solutions.
type LocalCache struct {
	mu       sync.RWMutex           // Protects concurrent access to the items map
	items    map[string]*cacheEntry // Stores all cached entries
	stopChan chan struct{}          // Signal channel to stop the cleanup goroutine
	config   *Config                // Configuration including default TTL
}

// NewLocalCache creates a new in-memory cache instance.
// It starts a background goroutine that periodically removes expired entries.
// Remember to call Close() when the cache is no longer needed to prevent goroutine leaks.
//
// Parameters:
//   - config: Cache configuration. If nil, default configuration is used.
//
// Returns:
//   - A new LocalCache instance with background cleanup running.
func NewLocalCache(config *Config) *LocalCache {
	if config == nil {
		config = NewDefaultConfig()
	}

	cache := &LocalCache{
		items:    make(map[string]*cacheEntry),
		stopChan: make(chan struct{}),
		config:   config,
	}

	// Start background cleanup goroutine
	go cache.cleanupExpired()

	return cache
}

// Get retrieves a value from the cache.
// It performs expiration checking and returns false for expired entries.
// This method is thread-safe and uses read locks for better concurrent performance.
//
// Parameters:
//   - ctx: Context for the operation (currently unused but kept for interface compatibility)
//   - key: The cache key to look up
//
// Returns:
//   - value: The cached value if found and not expired
//   - found: true if the key exists and hasn't expired, false otherwise
func (c *LocalCache) Get(ctx context.Context, key string) (interface{}, bool) {
	c.mu.RLock()
	defer c.mu.RUnlock()

	entry, exists := c.items[key]
	if !exists {
		return nil, false
	}

	// Check if entry has expired
	if time.Now().After(entry.expireTime) {
		// Entry exists but is expired, treat as cache miss
		return nil, false
	}

	return entry.value, true
}

// Set stores a value in the cache with the specified TTL.
// If TTL is 0, the default TTL from the configuration is used.
// This method overwrites any existing value for the same key.
//
// Parameters:
//   - ctx: Context for the operation (currently unused but kept for interface compatibility)
//   - key: The cache key
//   - value: The value to cache (can be any type)
//   - ttl: Time-to-live for this entry. Use 0 for default TTL.
//
// Returns:
//   - error: Always nil for this implementation, but kept for interface compatibility
func (c *LocalCache) Set(ctx context.Context, key string, value interface{}, ttl time.Duration) error {
	c.mu.Lock()
	defer c.mu.Unlock()

	// Use default TTL if none specified
	if ttl == 0 {
		ttl = c.config.DefaultTTL
	}

	c.items[key] = &cacheEntry{
		value:      value,
		expireTime: time.Now().Add(ttl),
	}

	return nil
}

// Delete removes a specific key from the cache.
// This operation is idempotent - deleting a non-existent key is not an error.
//
// Parameters:
//   - ctx: Context for the operation (currently unused but kept for interface compatibility)
//   - key: The cache key to delete
//
// Returns:
//   - error: Always nil for this implementation
func (c *LocalCache) Delete(ctx context.Context, key string) error {
	c.mu.Lock()
	defer c.mu.Unlock()

	delete(c.items, key)
	return nil
}

// DeleteByPrefix removes all cache entries whose keys start with the given prefix.
// This is useful for invalidating groups of related cache entries.
//
// Parameters:
//   - ctx: Context for the operation (currently unused but kept for interface compatibility)
//   - prefix: The key prefix to match for deletion
//
// Returns:
//   - error: Always nil for this implementation
func (c *LocalCache) DeleteByPrefix(ctx context.Context, prefix string) error {
	c.mu.Lock()
	defer c.mu.Unlock()

	// Collect keys to delete (can't delete while iterating)
	keysToDelete := make([]string, 0)
	for key := range c.items {
		if len(key) >= len(prefix) && key[:len(prefix)] == prefix {
			keysToDelete = append(keysToDelete, key)
		}
	}

	// Delete the collected keys
	for _, key := range keysToDelete {
		delete(c.items, key)
	}

	return nil
}

// Clear removes all entries from the cache.
// This is useful for cache invalidation scenarios or testing.
// Use with caution in production as it affects all cached data.
//
// Parameters:
//   - ctx: Context for the operation (currently unused but kept for interface compatibility)
//
// Returns:
//   - error: Always nil for this implementation
func (c *LocalCache) Clear(ctx context.Context) error {
	c.mu.Lock()
	defer c.mu.Unlock()

	// Create a new map to clear all entries
	c.items = make(map[string]*cacheEntry)
	return nil
}

// cleanupExpired runs in a background goroutine and periodically removes expired entries.
// This prevents memory leaks from accumulating expired entries.
// The cleanup runs every 30 seconds to balance between memory efficiency and CPU usage.
// This goroutine stops when Stop() is called or stopChan is closed.
func (c *LocalCache) cleanupExpired() {
	ticker := time.NewTicker(5 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			// Periodically remove expired entries
			c.removeExpired()
		case <-c.stopChan:
			// Close signal received, exit goroutine
			return
		}
	}
}

// removeExpired removes all expired entries from the cache.
// This method is called periodically by the cleanup goroutine.
// It holds a write lock during the operation, so it's designed to be quick.
func (c *LocalCache) removeExpired() {
	c.mu.Lock()
	defer c.mu.Unlock()

	now := time.Now()
	// Iterate through all entries and remove expired ones
	for key, entry := range c.items {
		if now.After(entry.expireTime) {
			delete(c.items, key)
		}
	}
}

// GetDefaultTTL returns the default TTL configured for this cache.
// This is useful for backends that need to know the base TTL.
func (c *LocalCache) GetDefaultTTL() time.Duration {
	if c.config != nil {
		return c.config.DefaultTTL
	}
	return 5 * time.Minute // Fallback default
}

// Close gracefully shuts down the cache by stopping the background cleanup goroutine.
// This should be called when the cache is no longer needed to prevent goroutine leaks.
// After calling Stop, the cache can still be used but expired entries won't be automatically cleaned up.
func (c *LocalCache) Close() {
	close(c.stopChan)
}
