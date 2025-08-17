# Backend Setup with Configurable TTLs

This document shows how to set up cached backends with proper TTL configuration that respects the base cache TTL.

## Simple Setup (Recommended)

Use the default configurations with a base TTL from your cache:

```go
package main

import (
    "os"
    "time"
    
    "github.com/quantumwake/alethic-ism-core-go/pkg/cache"
    "github.com/quantumwake/alethic-ism-core-go/pkg/repository/processor"
    "github.com/quantumwake/alethic-ism-core-go/pkg/repository/route"
    "github.com/quantumwake/alethic-ism-core-go/pkg/repository/user"
    "github.com/quantumwake/alethic-ism-core-go/pkg/repository/project"
)

func SetupCachedBackends() error {
    // Create a local cache with an aggressive cache expiration
    backendCache := cache.NewLocalCache(&cache.CacheConfig{
        DefaultTTL: 30 * time.Second,
    })
    defer backendCache.Close()

    // Get the DSN for the database
    dsn := os.Getenv("DSN")
    
    // Get the base TTL from the cache configuration
    baseTTL := backendCache.GetDefaultTTL()

    // Initialize backends with the cache and base TTL
    // Each backend will use method-specific TTLs based on the base TTL
    processorBackend := processor.NewCachedBackend(dsn, backendCache, baseTTL)
    routeBackend := route.NewCachedBackend(dsn, backendCache, baseTTL)
    userBackend := user.NewCachedBackend(dsn, backendCache, baseTTL)
    projectBackend := project.NewCachedBackend(dsn, backendCache, baseTTL)

    return nil
}
```

## Custom Configuration

For fine-grained control over individual method TTLs:

```go
func SetupCustomCachedBackends() error {
    // Create a local cache
    backendCache := cache.NewLocalCache(&cache.CacheConfig{
        DefaultTTL: 30 * time.Second,
    })
    defer backendCache.Close()

    dsn := os.Getenv("DSN")
    baseTTL := backendCache.GetDefaultTTL()

    // Create custom configurations for each backend
    processorConfig := cache.NewMethodTTLConfig(baseTTL)
    processorConfig.SetMethodTTL("FindProviderClasses", 15*time.Minute) // Static data
    processorConfig.SetMethodTTL("FindProcessorByID", 1*time.Minute)    // Override for specific method

    userConfig := cache.NewMethodTTLConfig(baseTTL)
    userConfig.SetMethodTTL("FindUserByID", 10*time.Minute) // User profiles are stable

    // Initialize backends with custom configurations
    processorBackend := processor.NewCachedBackendWithConfig(dsn, backendCache, processorConfig)
    userBackend := user.NewCachedBackendWithConfig(dsn, backendCache, userConfig)

    // Use default configurations for others
    routeBackend := route.NewCachedBackend(dsn, backendCache, baseTTL)
    projectBackend := project.NewCachedBackend(dsn, backendCache, baseTTL)

    return nil
}
```

## Default TTL Configurations

Each backend has sensible defaults based on data characteristics:

### Processor Backend
- `FindProviderClasses`: 10 minutes (static configuration)
- `FindProviders`: 5 minutes (changes occasionally)
- `FindProcessorByID`: base TTL (frequently accessed)

### User Backend
- `FindUserByID`: 15 minutes (very stable data)

### Project Backend
- `FindByID`: base TTL
- `FindAllByUserID`: base TTL + 2 minutes (lists change less frequently)

### Route Backend
- All methods use base TTL (routes are relatively static)

## Key Benefits

1. **Respects Base TTL**: All backends now honor the cache's configured TTL
2. **Method-Specific Overrides**: Can still override TTL for specific methods when needed
3. **Centralized Configuration**: TTL configuration is declarative and easy to manage
4. **No Hardcoded Values**: TTLs are based on the base configuration, not hardcoded

## Migration from Old Setup

If you were using the old setup without TTL parameters:

```go
// Old way (hardcoded TTLs)
processorBackend := processor.NewCachedBackend(dsn, backendCache)

// New way (respects base TTL)
baseTTL := backendCache.GetDefaultTTL()
processorBackend := processor.NewCachedBackend(dsn, backendCache, baseTTL)
```

The new approach ensures that when you configure your cache with a 30-second TTL, the backends actually respect that configuration instead of using their own hardcoded values.