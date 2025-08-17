# Cache Package

A simple, flexible caching layer for Go backends with TTL management and easy migration path from local to distributed cache systems.

## Features

- **Generic Cache Interface**: Swap between local memory, Redis, or other cache implementations
- **Configurable TTLs**: Global and per-method TTL configuration
- **Type-Safe Generics**: Use `CallCached[T]` for compile-time type safety
- **Smart Invalidation**: Automatic cache invalidation on write operations
- **Method Signature Caching**: Optimized reflection for dynamic method calls
- **Thread-Safe**: Concurrent access with read/write locks

## Quick Start

```go
// Create a cache with 30-second TTL
backendCache := cache.NewLocalCache(&cache.Config{
    DefaultTTL: 30 * time.Second,
})
defer backendCache.Close()

// Initialize backend with cache and base TTL
baseTTL := backendCache.GetDefaultTTL()
processorBackend := processor.NewCachedBackend(dsn, backendCache, baseTTL)
```

For complete backend setup examples, see [BACKEND_SETUP_EXAMPLE.md](BACKEND_SETUP_EXAMPLE.md).

## TTL Configuration

### Method-Specific TTLs

Configure different TTLs for different methods based on data characteristics:

```go
// Create custom configuration
config := cache.NewMethodTTLConfig(30 * time.Second)
config.SetMethodTTL("FindProviderClasses", 10*time.Minute)  // Static data
config.SetMethodTTL("FindUserByID", 15*time.Minute)         // Stable data
config.SetMethodTTL("FindProcessorByID", 30*time.Second)    // Dynamic data

// Apply configuration
backend := processor.NewCachedBackendWithConfig(dsn, backendCache, config)
```

### Default Configurations

Each backend type has sensible defaults:

- **Processor**: Provider classes (10min), Providers (5min), Processors (base TTL)
- **User**: User profiles (15min - very stable)
- **Project**: Projects (base TTL), User project lists (base TTL + 2min)
- **Route**: All methods use base TTL (routes are relatively static)

## Type-Safe Caching

Use generic helper functions for type safety:

```go
// Type-safe caching with generics
user, err := cache.CallCached(cached, ctx, "FindUserByID", []interface{}{userID},
    func() (*User, error) {
        return backend.FindUserByID(userID)
    })

// Custom TTL for specific calls
config, err := cache.CallCachedWithTTL(cached, ctx, "GetConfig", nil, 1*time.Hour,
    func() (*Config, error) {
        return backend.GetConfig()
    })
```

## Cache Invalidation

Write operations automatically invalidate affected cache entries:

```go
// CreateOrUpdate in processor backend
func (cb *CachedBackendStorage) CreateOrUpdate(processor *Processor) error {
    err := cb.base.CreateOrUpdate(processor)
    if err != nil {
        return err
    }
    
    // Automatically invalidate affected cache entries
    ctx := context.Background()
    cb.InvalidateMethod(ctx, "FindProcessorByID", processor.ID)
    cb.InvalidateMethod(ctx, "FindProcessorByProjectID", processor.ProjectID)
    
    return nil
}
```

## Implementation Details

### Local Cache
- In-memory storage using `sync.Map`
- Background goroutine for TTL-based eviction (runs every 5 seconds)
- Thread-safe with read/write locks
- Suitable for single-instance applications

### Cache Interface
Any cache implementation must implement:
```go
type Cache interface {
    Get(ctx context.Context, key string) (interface{}, bool)
    Set(ctx context.Context, key string, value interface{}, ttl time.Duration) error
    Delete(ctx context.Context, key string) error
    Clear(ctx context.Context) error
    GetDefaultTTL() time.Duration
    Close()
}
```

### Migration Path

Easily migrate from local to distributed caching:

```go
// Local cache (development)
cache := cache.NewLocalCache(&cache.Config{DefaultTTL: 30*time.Second})

// Redis cache (future production)
cache := cache.NewRedisCache(&cache.Config{DefaultTTL: 30*time.Second})

// Usage remains the same
backend := processor.NewCachedBackend(dsn, cache, cache.GetDefaultTTL())
```

## Best Practices

1. **Respect Base TTL**: Always use `cache.GetDefaultTTL()` to honor configuration
2. **Method-Specific TTLs**: Override only when data characteristics require it
3. **Cache Invalidation**: Always invalidate on write operations
4. **Resource Cleanup**: Always call `cache.Close()` when done

## Testing

```bash
# Run tests
go test ./pkg/cache/... -v

# Build all packages
go build ./pkg/cache/... ./pkg/repository/...
```