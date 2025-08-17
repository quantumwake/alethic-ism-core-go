# Cache Package

A flexible, high-performance caching layer for Go backends with support for method signature caching, TTL management, and easy migration from local to distributed cache systems.

## Features

- **Generic Cache Interface**: Swap between local memory, Redis, or other cache implementations
- **Method Signature Caching**: Eliminates reflection overhead for dynamic method calls
- **Automatic Method Discovery**: Auto-registers backend methods for caching
- **Configurable TTLs**: Per-method or global TTL configuration
- **Type-Safe Generics**: Use `CallCached[T]` for compile-time type safety
- **Cache Invalidation**: Smart invalidation strategies for write operations

## Quick Start

### Basic Usage

```go
import "github.com/quantumwake/alethic-ism-core-go/pkg/cache"

// Create a local cache
localCache := cache.NewLocalCache(nil)
defer localCache.Stop()

// Wrap any backend with caching
backend := &MyBackend{db: database}
cachedBackend := cache.NewCachedBackend(backend, localCache, 5*time.Minute)
```

### Using with Processor Backend

```go
// Create a cached processor backend
processorBackend := processor.NewCachedBackend(dsn, localCache)

// Use it exactly like the original backend
processor, err := processorBackend.FindProcessorByID("uuid-123")
// First call hits database, subsequent calls use cache
```

## Method Signature Caching

Method signature caching eliminates reflection overhead by caching method metadata on first use.

### Automatic Registration

Methods are automatically discovered and registered when creating a cached backend:

```go
type UserBackend struct {
    db *sql.DB
}

func (u *UserBackend) GetUser(id string) (*User, error) {
    // Database query
}

func (u *UserBackend) ListUsers() ([]*User, error) {
    // Database query
}

// Methods are auto-registered
backend := &UserBackend{db: db}
cached := cache.NewCachedBackend(backend, localCache, 5*time.Minute)
```

### Manual Registration

Register methods explicitly for fine-grained control:

```go
cached := cache.NewCachedBackend(backend, localCache, 5*time.Minute)

// Register a single method
cached.RegisterMethod("GetUser", backend.GetUser, &cache.MethodConfig{
    TTL:       10 * time.Minute,
    Cacheable: true,
})

// Register multiple methods
cached.RegisterMethods(map[string]interface{}{
    "GetUser":   backend.GetUser,
    "ListUsers": backend.ListUsers,
})
```

### Method Configuration

Configure caching behavior per method:

```go
// Make certain methods non-cacheable
cached.SetMethodConfig("UpdateUser", &cache.MethodConfig{
    Cacheable: false, // Write operations shouldn't be cached
})

// Set custom TTL for static data
cached.SetMethodConfig("GetConfig", &cache.MethodConfig{
    TTL:       1 * time.Hour, // Config rarely changes
    Cacheable: true,
})
```

## Type-Safe Caching

Use generic helper functions for type safety:

```go
// Type-safe caching with generics
user, err := cache.CallCached(cached, ctx, "GetUser", []interface{}{userID},
    func() (*User, error) {
        return backend.GetUser(userID)
    })

// Custom TTL for specific calls
config, err := cache.CallCachedWithTTL(cached, ctx, "GetConfig", nil, 1*time.Hour,
    func() (*Config, error) {
        return backend.GetConfig()
    })
```

## Dynamic Method Execution

Execute methods dynamically with cached signatures:

```go
// Execute method with automatic signature caching
result, err := cached.Execute(ctx, "GetUser", []interface{}{userID},
    func() (*User, error) {
        return backend.GetUser(userID)
    })

// Execute with arguments
args := []reflect.Value{
    reflect.ValueOf(userID),
    reflect.ValueOf(includeDeleted),
}
result, err := cached.ExecuteWithArgs(ctx, "GetUserExtended", 
    []interface{}{userID, includeDeleted}, // cache key args
    args,                                    // function args
    backend.GetUserExtended)
```

## Cache Invalidation

### Automatic Invalidation

Write operations automatically invalidate related cache entries:

```go
type CachedUserBackend struct {
    *cache.CachedBackend
    base *UserBackend
}

func (c *CachedUserBackend) UpdateUser(user *User) error {
    err := c.base.UpdateUser(user)
    if err != nil {
        return err
    }
    
    // Invalidate specific cache entries
    ctx := context.Background()
    c.InvalidateMethod(ctx, "GetUser", user.ID)
    c.InvalidateMethod(ctx, "ListUsersByDepartment", user.DepartmentID)
    
    return nil
}
```

### Manual Invalidation

```go
// Invalidate specific cache keys
cached.InvalidateCache(ctx, "GetUser:123", "ListUsers:active")

// Clear entire cache
cached.InvalidateCache(ctx, "*")

// Invalidate by method
cached.InvalidateMethod(ctx, "GetUser", "user-123")
```

## Performance Benefits

Method signature caching provides significant performance improvements:

1. **First Call**: Performs reflection to analyze method signature
2. **Subsequent Calls**: Uses cached signature, avoiding reflection
3. **Cache Hits**: Returns data directly from cache without any reflection

Example performance improvement:
```
Without signature caching: ~500ns per call (with reflection)
With signature caching:    ~50ns per call (no reflection)
Cache hit:                 ~10ns per call (direct return)
```

## Migration Path

The cache package is designed for easy migration from local to distributed caching:

### Local Cache (Development)
```go
cache := cache.NewLocalCache(nil)
backend := NewCachedBackend(db, cache, 5*time.Minute)
```

### Redis Cache (Production)
```go
// Future implementation
cache := cache.NewRedisCache(redisClient, nil)
backend := NewCachedBackend(db, cache, 5*time.Minute)
```

The interface remains the same, only the cache implementation changes.

## Best Practices

1. **Cache Keys**: Use descriptive method names and include all parameters that affect the result
2. **TTL Configuration**: Use longer TTLs for static data, shorter for frequently changing data
3. **Invalidation**: Always invalidate cache after write operations
4. **Method Registration**: Pre-register heavy methods at startup to avoid first-call latency
5. **Monitoring**: Track cache hit rates and adjust TTLs accordingly

## Testing

The package includes comprehensive tests for all features:

```bash
go test ./pkg/cache/... -v
```

Performance benchmarks:
```bash
go test ./pkg/cache/... -bench=. -benchmem
```

## Thread Safety

All cache operations are thread-safe:
- Local cache uses `sync.RWMutex` for concurrent access
- Method signatures use `sync.Map` for lock-free reads
- Safe for use in concurrent goroutines

## Future Enhancements

- Redis cache implementation
- Memcached support
- Cache statistics and metrics
- Cache warming strategies
- Distributed cache invalidation
- Circuit breaker integration