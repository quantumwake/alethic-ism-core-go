package cache

import (
	"container/heap"
	"context"
	"sync"
	"time"
)

// LocalCache implements an in-memory cache with TTL support.
// It uses a map for O(1) lookups and a background goroutine for periodic cleanup.
// This implementation is thread-safe and suitable for single-instance applications.
// For distributed systems, consider using Redis or similar distributed cache solutions.
type LocalCache struct {
	mu        sync.RWMutex           // Protects concurrent access to the items map
	items     map[string]*cacheEntry // Stores all cached entries
	itemsHeap cacheItemsHeap         // Min-heap to track expiration times
	stopChan  chan struct{}          // Signal channel to stop the cleanup goroutine
	config    *Config                // Configuration including default TTL

	//createChanMap sync.Map // map[string]chan struct{} allows us to create per-key channels for entry creation, to prevent master lock contention.
}

type Option func(*LocalCache)

func WithOptionTTL(ttl time.Duration) Option {
	return func(c *LocalCache) {
		c.config.DefaultTTL = ttl
	}
}

func WithOptionCleanupInterval(interval time.Duration) Option {
	return func(c *LocalCache) {
		c.config.CleanupDurationInterval = interval
	}
}

// NewLocalCacheWithOptions creates a new LocalCache instance with functional options.
func NewLocalCacheWithOptions(options ...Option) *LocalCache {
	localCache := NewLocalCache(nil)
	for _, option := range options {
		option(localCache)
	}
	return localCache
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
		items:     make(map[string]*cacheEntry),
		itemsHeap: cacheItemsHeap{},
		stopChan:  make(chan struct{}),
		config:    config,

		// create a channel for creating entries
		//createChanMap: sync.Map{},
	}
	heap.Init(&cache.itemsHeap) // Initialize the heap structure

	// Start background cleanup goroutine
	go cache.cleanupExpired()

	return cache
}

// Len returns the number of items currently stored in the cache.
func (c *LocalCache) Len() int {
	c.mu.RLock()
	defer c.mu.RUnlock()
	count := len(c.items)
	return count
}

//func (c *LocalCache) channelize(key string) chan struct{} {
//	ch, _ := c.createChanMap.LoadOrStore(key, make(chan struct{}))
//	return ch.(chan struct{})
//}

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
func (c *LocalCache) Get(_ context.Context, key string) (any, bool) {
	c.mu.RLock()
	defer c.mu.RUnlock()

	entry, exists := c.items[key]
	if !exists {
		return nil, false
	}

	// Check if entry has expired
	if time.Now().After(entry.evictAt) {
		// Entry exists but is expired, treat as cache miss
		return nil, false
	}

	return entry.value, true
}

// GetCreateOrUpdate retrieves a value from the cache or creates/updates it using fetchFunc.
// It ensures thread-safe access and prevents cache stampedes by using locks.
func (c *LocalCache) GetCreateOrUpdate(ctx context.Context, key string, fetchFunc func(exists bool, existingValue any) (any, error), ttl time.Duration) (any, error) {
	c.mu.RLock() // First attempt to get the value from cache
	entry, exists := c.items[key]
	c.mu.RUnlock() // NOTE BEGIN : at this point the lock is released, so the entry could be modified by other goroutines

	// if cache item exists and its not expired, return immediate
	if exists && time.Now().Before(entry.evictAt) {
		// Value found and not expired
		return entry.value, nil
	}

	// reacquire the master lock (TODO, optimize using per-key locks to prevent master lock contention)
	c.mu.Lock()
	defer c.mu.Unlock()

	// double check since another go routine could have updated the cache while we read lock was released above
	if exists && time.Now().Before(entry.evictAt) {
		return entry.value, nil // value found and not expired
	}

	// currently held value or to be created valued
	var value any = nil
	if exists {
		value = entry.value
	}

	// if the value exists, then it must be expired
	value, err := fetchFunc(exists, value) // pass exists (essentially telling fetchFunc if this is a create or update)
	if err != nil {
		return nil, err
	}

	if value == nil {
		return nil, nil // do not cache nil values
	}

	if exists {
		entry = c.update(key, value, ttl)
	} else {
		entry = c.add(key, value, ttl)
	}

	return entry.value, nil
}

// add creates a new cache entry and adds it to the cache.
// It assumes the caller holds the write lock.
func (c *LocalCache) add(key string, value any, ttl time.Duration) *cacheEntry {
	// Use default TTL if none specified
	if ttl == 0 {
		ttl = c.config.DefaultTTL
	}

	entry := &cacheEntry{
		key:     key,
		value:   value,
		evictAt: time.Now().Add(ttl),
	}
	c.items[key] = entry
	heap.Push(&c.itemsHeap, entry)
	return entry
}

// update modifies an existing cache entry with a new value and TTL.
// It assumes the caller holds the write lock.
func (c *LocalCache) update(key string, value any, ttl time.Duration) *cacheEntry {
	if ttl == 0 {
		ttl = c.config.DefaultTTL
	}
	entry := c.items[key]
	entry.evictAt = time.Now().Add(ttl)
	entry.value = value
	heap.Fix(&c.itemsHeap, entry.index)
	return entry
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
func (c *LocalCache) Set(ctx context.Context, key string, value any, ttl time.Duration) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.add(key, value, ttl)
}

// delete removes a specific key from the cache and ensures evict time is set to now.
// It assumes the caller holds the write lock
func (c *LocalCache) delete(key string) {
	entry, ok := c.items[key]
	if !ok {
		return // key does not exist, noop
	}
	entry.evictAt = time.Now() // set eviction time to now
	entry.value = nil
	delete(c.items, key)
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
func (c *LocalCache) Delete(ctx context.Context, key string) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.delete(key)
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
func (c *LocalCache) Clear(_ context.Context) error {
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
	ticker := time.NewTicker(c.config.CleanupDurationInterval)
	defer ticker.Stop() // Ensure ticker is stopped when goroutine exits

	evictFn := func() {
		if c.itemsHeap.Len() == 0 {
			return
		}

		now := time.Now()

		// acquire read lock to peek at the heap, first item expired then acquire write lock to evict
		c.mu.RLock()
		item := c.itemsHeap[0] // Peek at the item with the earliest eviction time
		c.mu.RUnlock()         // Release read lock before acquiring write lock

		// If the earliest item hasn't expired yet, nothing to do
		if item.evictAt.After(now) {
			return
		}

		// otherwise we obtain the master lock for eviction
		c.mu.Lock()
		defer c.mu.Unlock()

		// iterate and evict all expired items
		for c.itemsHeap.Len() > 0 {
			item = c.itemsHeap[0] // double check, item may have updated.
			if item.evictAt.After(now) {
				return
			}
			c.delete(item.key) // internal call with no lock.
			heap.Pop(&c.itemsHeap)
		}
	}

	for {
		select {
		case <-ticker.C:
			evictFn() // Periodically remove expired entries
		case <-c.stopChan:
			return
		}
	}
}

// removeExpired removes all expired entries from the cache.
// This method is called periodically by the cleanup goroutine.
// It holds a write lock during the operation, so it's designed to be quick.
//func (c *LocalCache) removeExpired() {
//
//	evictFn := func() {
//		c.mu.Lock()
//		defer c.mu.Unlock()
//	}
//
//
//	select {
//
//	}

//now := time.Now()
// Iterate through all entries and remove expired ones
//for key, entry := range c.items {
//	if now.After(entry.evictAt) {
//		delete(c.items, key)
//	}
//}
//}

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
