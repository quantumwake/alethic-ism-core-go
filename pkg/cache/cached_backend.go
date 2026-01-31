package cache

import (
	"context"
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"reflect"
	"sync"
	"time"
)

// CacheableBackend is a marker interface for any backend that can be wrapped with caching.
// Any struct can implement this interface, making it compatible with the caching layer.
type CacheableBackend interface{}

// methodSignature stores reflection metadata about a method to avoid repeated reflection calls.
// This significantly improves performance for dynamic method invocation.
type methodSignature struct {
	Type       reflect.Type                          // The function type
	Value      reflect.Value                         // The function value
	NumIn      int                                   // Number of input parameters
	NumOut     int                                   // Number of output parameters
	IsVariadic bool                                  // Whether the function is variadic
	CallFunc   func([]reflect.Value) []reflect.Value // Cached Call function
}

// CachedBackend provides a generic caching wrapper for any backend implementation.
// It intercepts method calls and caches their results based on configurable TTLs.
// This struct enables consistent caching behavior across different backend types.
type CachedBackend struct {
	backend          CacheableBackend // The underlying backend being wrapped
	cache            Cache            // The cache implementation (local, Redis, etc.)
	defaultTTL       time.Duration    // Default TTL for cached entries
	methodSignatures sync.Map         // Cache of method signatures to avoid reflection overhead
	methodConfigs    sync.Map         // Per-method configuration (TTL, cache behavior)
	keyRegistry      sync.Map         // Maps method:prefixArgs to list of full cache keys
}

// NewCachedBackend creates a new caching wrapper for any backend.
// This is the main entry point for adding caching to existing backends.
//
// Parameters:
//   - backend: The backend to wrap with caching functionality
//   - cache: Cache implementation to use. If nil, creates a local in-memory cache.
//   - defaultTTL: Default time-to-live for cached entries. If 0, uses 5 minutes.
//
// Returns:
//   - A new CachedBackend instance wrapping the provided backend
func NewCachedBackend(backend CacheableBackend, cache Cache, defaultTTL time.Duration) *CachedBackend {
	if cache == nil {
		// Default to local in-memory cache if none provided
		cache = NewLocalCache(NewDefaultConfig())
	}
	if defaultTTL == 0 {
		// Default to 5 minutes if no TTL specified
		defaultTTL = 5 * time.Minute
	}

	cb := &CachedBackend{
		backend:    backend,
		cache:      cache,
		defaultTTL: defaultTTL,
	}

	// Auto-discover and register backend methods if possible
	cb.AutoRegisterMethods(backend)

	return cb
}

// BuildCacheKey generates a deterministic cache key from method name and arguments.
// It uses JSON serialization and SHA256 hashing to ensure consistent keys
// even for complex argument types.
//
// Parameters:
//   - method: The method name being cached
//   - args: Variable arguments that were passed to the method
//
// Returns:
//   - string: A cache key in format "methodName:hashPrefix"
//   - error: If marshaling arguments fails
//
// Example:
//
//	key, _ := BuildCacheKey("FindUserByID", "user-123")
//	// Returns: "FindUserByID:a3b4c5d6"
func (cb *CachedBackend) BuildCacheKey(method string, args ...interface{}) (string, error) {
	keyData := struct {
		Method string        `json:"method"`
		Args   []interface{} `json:"args"`
	}{
		Method: method,
		Args:   args,
	}

	// Serialize to JSON for consistent representation
	jsonBytes, err := json.Marshal(keyData)
	if err != nil {
		return "", fmt.Errorf("failed to marshal cache key: %w", err)
	}

	// Use SHA256 hash to create a fixed-length key prefix
	hash := sha256.Sum256(jsonBytes)
	cacheKey := fmt.Sprintf("%s:%x", method, hash[:8])

	// Register this key with its method and first argument (if any) for prefix invalidation
	if len(args) > 0 {
		// Create a registry key from method and first argument
		registryKey := fmt.Sprintf("%s:%v", method, args[0])

		// Get or create the list of cache keys for this prefix
		val, _ := cb.keyRegistry.LoadOrStore(registryKey, &sync.Map{})
		keySet := val.(*sync.Map)
		keySet.Store(cacheKey, true)
	}

	return cacheKey, nil
}

// GetCached implements the cache-aside pattern for any function.
// It first checks the cache, and if not found, calls the fetch function
// and caches the result with the default TTL.
//
// Parameters:
//   - ctx: Context for cache operations
//   - cacheKey: The cache key to use
//   - fetchFunc: Function to call if cache miss occurs
//
// Returns:
//   - interface{}: The cached or fetched value
//   - error: Any error from the fetch function
//
// Pattern:
//  1. Check cache for existing value
//  2. If found (cache hit), return cached value
//  3. If not found (cache miss), call fetch function
//  4. Cache the result for future requests
//  5. Return the result
func (cb *CachedBackend) GetCached(ctx context.Context, cacheKey string, fetchFunc func() (interface{}, error)) (interface{}, error) {
	// Try to get from cache first
	if cached, found := cb.cache.Get(ctx, cacheKey); found {
		return cached, nil
	}

	// Cache miss - fetch from source
	result, err := fetchFunc()
	if err != nil {
		return nil, err
	}

	// Store in cache for next time (ignore cache set errors)
	cb.cache.Set(ctx, cacheKey, result, cb.defaultTTL)

	return result, nil
}

// GetCachedWithTTL is like GetCached but allows specifying a custom TTL.
// Use this when certain data should be cached for different durations
// (e.g., frequently changing data vs. static configuration).
//
// Parameters:
//   - ctx: Context for cache operations
//   - cacheKey: The cache key to use
//   - ttl: Custom time-to-live for this cache entry
//   - fetchFunc: Function to call if cache miss occurs
//
// Returns:
//   - interface{}: The cached or fetched value
//   - error: Any error from the fetch function
func (cb *CachedBackend) GetCachedWithTTL(ctx context.Context, cacheKey string, ttl time.Duration, fetchFunc func() (interface{}, error)) (interface{}, error) {
	// Try to get from cache first
	if cached, found := cb.cache.Get(ctx, cacheKey); found {
		return cached, nil
	}

	// Cache miss - fetch from source
	result, err := fetchFunc()
	if err != nil {
		return nil, err
	}

	// Store in cache with custom TTL
	cb.cache.Set(ctx, cacheKey, result, ttl)

	return result, nil
}

// InvalidateCache removes cache entries matching the given patterns.
// This is typically called after write operations to ensure cache consistency.
//
// Parameters:
//   - ctx: Context for cache operations
//   - patterns: Cache keys or patterns to invalidate. Use "*" to clear all.
//
// Returns:
//   - error: Any error from cache operations
//
// Example:
//
//	InvalidateCache(ctx, "FindUserByID:123", "FindUsersByRole:admin")
//	InvalidateCache(ctx, "*") // Clear entire cache
func (cb *CachedBackend) InvalidateCache(ctx context.Context, patterns ...string) error {
	for _, pattern := range patterns {
		if pattern == "*" {
			// Special case: clear entire cache
			return cb.cache.Clear(ctx)
		}
		// Delete specific cache key (ignore errors)
		cb.cache.Delete(ctx, pattern)
	}
	return nil
}

// InvalidateMethod invalidates cache for a specific method call.
// This is useful when you know exactly which cached method result needs invalidation.
//
// Parameters:
//   - ctx: Context for cache operations
//   - method: The method name whose cache to invalidate
//   - args: The exact arguments that were used in the cached call
//
// Returns:
//   - error: Any error from building cache key or deletion
//
// Example:
//
//	InvalidateMethod(ctx, "FindUserByID", "user-123")
func (cb *CachedBackend) InvalidateMethod(ctx context.Context, method string, args ...interface{}) error {
	cacheKey, err := cb.BuildCacheKey(method, args...)
	if err != nil {
		return err
	}
	cb.cache.Delete(ctx, cacheKey)
	return nil
}

// InvalidateMethodPrefix invalidates all cache entries for a method that start with given arguments.
// This is useful when you want to invalidate all variations of a method call that share
// the same prefix arguments but may have different trailing arguments.
//
// Parameters:
//   - ctx: Context for cache operations
//   - method: The method name whose cache to invalidate
//   - prefixArgs: The prefix arguments that identify the entries to invalidate
//
// Returns:
//   - error: Any error from cache operations
//
// Example:
//
//	InvalidateMethodPrefix(ctx, "FindStateFull", "state-123")
//	// This would invalidate all entries like:
//	// FindStateFull("state-123", flags1)
//	// FindStateFull("state-123", flags2)
//	// etc.
func (cb *CachedBackend) InvalidateMethodPrefix(ctx context.Context, method string, prefixArgs ...interface{}) error {
	// Build the registry key to find all cache keys for this method+prefix combination
	if len(prefixArgs) == 0 {
		// No prefix args, just delete all entries for this method
		prefix := fmt.Sprintf("%s:", method)
		if deleter, ok := cb.cache.(interface {
			DeleteByPrefix(context.Context, string) error
		}); ok {
			return deleter.DeleteByPrefix(ctx, prefix)
		}
		return nil
	}

	// Look up registered keys for this method and first argument
	registryKey := fmt.Sprintf("%s:%v", method, prefixArgs[0])
	if val, ok := cb.keyRegistry.Load(registryKey); ok {
		keySet := val.(*sync.Map)

		// Delete each registered cache key
		keySet.Range(func(key, _ interface{}) bool {
			cacheKey := key.(string)
			cb.cache.Delete(ctx, cacheKey)
			return true
		})

		// Clear the registry for this prefix
		cb.keyRegistry.Delete(registryKey)
	}

	return nil
}

// GetBackend returns the underlying backend instance.
// This is useful when you need direct access to the backend,
// bypassing the cache layer.
//
// Returns:
//   - CacheableBackend: The wrapped backend instance
func (cb *CachedBackend) GetBackend() CacheableBackend {
	return cb.backend
}

// CallCached is a generic helper function that provides type-safe caching.
// It handles the cache key generation, type conversion, and fallback logic.
// This is the recommended way to add caching to backend methods.
//
// Type Parameters:
//   - T: The return type of the cached method
//
// Parameters:
//   - cb: The CachedBackend instance
//   - ctx: Context for cache operations
//   - method: Name of the method being cached
//   - args: Arguments passed to the method
//   - fetchFunc: Function to call on cache miss
//
// Returns:
//   - T: The typed result from cache or fetch function
//   - error: Any error from fetching
//
// Example:
//
//	user, err := CallCached(cb, ctx, "FindUserByID", []interface{}{userID},
//	  func() (*User, error) {
//	    return backend.FindUserByID(userID)
//	  })
func CallCached[T any](cb *CachedBackend, ctx context.Context, method string, args []interface{}, fetchFunc func() (T, error)) (T, error) {
	var zero T

	// Build cache key from method and arguments
	cacheKey, err := cb.BuildCacheKey(method, args...)
	if err != nil {
		// If cache key generation fails, bypass cache
		return fetchFunc()
	}

	// Try to get from cache
	cached, err := cb.GetCached(ctx, cacheKey, func() (interface{}, error) {
		return fetchFunc()
	})

	if err != nil {
		return zero, err
	}

	// Type assert the cached value
	result, ok := cached.(T)
	if !ok {
		// Type assertion failed, fetch fresh data
		return fetchFunc()
	}

	return result, nil
}

// CallCachedWithTTL is like CallCached but with custom TTL.
// Use this for methods where you want different cache durations.
//
// Type Parameters:
//   - T: The return type of the cached method
//
// Parameters:
//   - cb: The CachedBackend instance
//   - ctx: Context for cache operations
//   - method: Name of the method being cached
//   - args: Arguments passed to the method
//   - ttl: Custom time-to-live for this cache entry
//   - fetchFunc: Function to call on cache miss
//
// Returns:
//   - T: The typed result from cache or fetch function
//   - error: Any error from fetching
//
// Example:
//
//	// Cache static config for 1 hour
//	config, err := CallCachedWithTTL(cb, ctx, "GetConfig", nil, 1*time.Hour,
//	  func() (*Config, error) {
//	    return backend.GetConfig()
//	  })
func CallCachedWithTTL[T any](cb *CachedBackend, ctx context.Context, method string, args []interface{}, ttl time.Duration, fetchFunc func() (T, error)) (T, error) {
	var zero T

	// Build cache key from method and arguments
	cacheKey, err := cb.BuildCacheKey(method, args...)
	if err != nil {
		// If cache key generation fails, bypass cache
		return fetchFunc()
	}

	// Try to get from cache with custom TTL
	cached, err := cb.GetCachedWithTTL(ctx, cacheKey, ttl, func() (interface{}, error) {
		return fetchFunc()
	})

	if err != nil {
		return zero, err
	}

	// Type assert the cached value
	result, ok := cached.(T)
	if !ok {
		// Type assertion failed, fetch fresh data
		return fetchFunc()
	}

	return result, nil
}

// CacheOptions provides configuration for cache operations.
// This can be extended with additional options as needed.
type CacheOptions struct {
	TTL                time.Duration // Custom TTL for the operation
	InvalidatePatterns []string      // Patterns to invalidate after operation
}

// CacheMethod defines caching behavior for a specific method.
// This can be used to configure method-specific caching rules.
type CacheMethod struct {
	Name    string        // Method name
	TTL     time.Duration // Method-specific TTL
	KeyArgs []int         // Which argument positions to include in cache key
	Enabled bool          // Whether caching is enabled for this method
}

// MethodConfig stores configuration for a specific method.
type MethodConfig struct {
	TTL        time.Duration                   // Method-specific TTL (0 means use default)
	Cacheable  bool                            // Whether this method should be cached
	KeyBuilder func(args []interface{}) string // Custom key builder function
}

// RegisterMethod registers a method with its signature for optimized caching.
// This avoids reflection overhead on subsequent calls.
//
// Parameters:
//   - name: The method name for identification
//   - method: The method to register (must be a function)
//   - config: Optional method-specific configuration
//
// Returns:
//   - error: If method is not a function or registration fails
func (cb *CachedBackend) RegisterMethod(name string, method interface{}, config *MethodConfig) error {
	methodValue := reflect.ValueOf(method)
	if methodValue.Kind() != reflect.Func {
		return fmt.Errorf("method must be a function")
	}

	methodType := methodValue.Type()
	sig := &methodSignature{
		Type:       methodType,
		Value:      methodValue,
		NumIn:      methodType.NumIn(),
		NumOut:     methodType.NumOut(),
		IsVariadic: methodType.IsVariadic(),
		CallFunc:   methodValue.Call,
	}

	cb.methodSignatures.Store(name, sig)

	if config != nil {
		cb.methodConfigs.Store(name, config)
	}

	return nil
}

// RegisterMethods registers multiple methods at once.
//
// Parameters:
//   - methods: Map of method names to method functions
//
// Returns:
//   - error: If any method registration fails
func (cb *CachedBackend) RegisterMethods(methods map[string]interface{}) error {
	for name, method := range methods {
		if err := cb.RegisterMethod(name, method, nil); err != nil {
			return fmt.Errorf("failed to register method %s: %w", name, err)
		}
	}
	return nil
}

// AutoRegisterMethods uses reflection to discover and register all exported methods of a backend.
// This is called automatically during backend creation but can be called manually if needed.
//
// Parameters:
//   - backend: The backend whose methods to register
func (cb *CachedBackend) AutoRegisterMethods(backend interface{}) {
	backendValue := reflect.ValueOf(backend)
	backendType := backendValue.Type()

	// Register methods from the backend struct
	for i := 0; i < backendType.NumMethod(); i++ {
		method := backendType.Method(i)
		// Only register exported methods (capitalized names)
		if method.IsExported() {
			methodFunc := backendValue.Method(i)
			_ = cb.RegisterMethod(method.Name, methodFunc.Interface(), nil)
		}
	}
}

// Execute uses cached method signatures to efficiently execute and cache function calls.
// Method signatures are cached on first use to avoid repeated reflection.
//
// Parameters:
//   - ctx: Context for cache operations
//   - methodName: Name for cache key generation and method lookup
//   - args: Arguments for cache key generation
//   - execFunc: The function to execute (must be a function type)
//
// Returns:
//   - interface{}: The function's return value
//   - error: Any error from execution or if execFunc is not a function
func (cb *CachedBackend) Execute(ctx context.Context, methodName string, args []interface{}, execFunc interface{}) (interface{}, error) {
	// Try to get cached method signature
	var sig *methodSignature
	if cachedSig, ok := cb.methodSignatures.Load(methodName); ok {
		sig = cachedSig.(*methodSignature)
	} else {
		// Register the method on first use
		if err := cb.RegisterMethod(methodName, execFunc, nil); err != nil {
			return nil, err
		}
		if cachedSig, ok := cb.methodSignatures.Load(methodName); ok {
			sig = cachedSig.(*methodSignature)
		} else {
			return nil, fmt.Errorf("failed to register method %s", methodName)
		}
	}

	// For function execution, we need to prepare the arguments
	// If the signature was registered from a method, it may need special handling
	var callFunc func() []reflect.Value

	// Check if execFunc is the actual function to call
	execValue := reflect.ValueOf(execFunc)
	if execValue.Kind() == reflect.Func {
		// Use the provided function directly
		callFunc = func() []reflect.Value {
			return execValue.Call(nil)
		}
	} else {
		// Use the cached signature's call function
		callFunc = func() []reflect.Value {
			return sig.CallFunc(nil)
		}
	}

	// Check if method has custom configuration
	var ttl = cb.defaultTTL
	if config, ok := cb.methodConfigs.Load(methodName); ok {
		methodConfig := config.(*MethodConfig)
		if !methodConfig.Cacheable {
			// Method is not cacheable, execute directly
			results := callFunc()
			if len(results) == 2 && !results[1].IsNil() {
				return results[0].Interface(), results[1].Interface().(error)
			}
			if len(results) > 0 {
				return results[0].Interface(), nil
			}
			return nil, nil
		}
		if methodConfig.TTL > 0 {
			ttl = methodConfig.TTL
		}
	}

	// Build cache key
	cacheKey, err := cb.BuildCacheKey(methodName, args...)
	if err != nil {
		// Cache key generation failed, execute directly
		results := callFunc()
		if len(results) == 2 && !results[1].IsNil() {
			return results[0].Interface(), results[1].Interface().(error)
		}
		if len(results) > 0 {
			return results[0].Interface(), nil
		}
		return nil, nil
	}

	// Use cache-aside pattern with configured TTL
	return cb.GetCachedWithTTL(ctx, cacheKey, ttl, func() (interface{}, error) {
		results := callFunc()
		if len(results) == 2 && !results[1].IsNil() {
			return results[0].Interface(), results[1].Interface().(error)
		}
		if len(results) > 0 {
			return results[0].Interface(), nil
		}
		return nil, nil
	})
}

// ExecuteWithArgs is like Execute but allows passing arguments to the function.
// This is useful when the function needs parameters.
//
// Parameters:
//   - ctx: Context for cache operations
//   - methodName: Name for cache key generation and method lookup
//   - cacheArgs: Arguments for cache key generation
//   - funcArgs: Arguments to pass to the function
//   - execFunc: The function to execute
//
// Returns:
//   - interface{}: The function's return value
//   - error: Any error from execution
func (cb *CachedBackend) ExecuteWithArgs(ctx context.Context, methodName string, cacheArgs []interface{}, funcArgs []reflect.Value, execFunc interface{}) (interface{}, error) {
	// Try to get cached method signature
	var sig *methodSignature
	if cachedSig, ok := cb.methodSignatures.Load(methodName); ok {
		sig = cachedSig.(*methodSignature)
	} else {
		// Register the method on first use
		if err := cb.RegisterMethod(methodName, execFunc, nil); err != nil {
			return nil, err
		}
		if cachedSig, ok := cb.methodSignatures.Load(methodName); ok {
			sig = cachedSig.(*methodSignature)
		} else {
			return nil, fmt.Errorf("failed to register method %s", methodName)
		}
	}

	// Build cache key
	cacheKey, err := cb.BuildCacheKey(methodName, cacheArgs...)
	if err != nil {
		// Cache key generation failed, execute directly
		results := sig.CallFunc(funcArgs)
		if len(results) == 2 && !results[1].IsNil() {
			return results[0].Interface(), results[1].Interface().(error)
		}
		return results[0].Interface(), nil
	}

	// Use cache-aside pattern
	return cb.GetCached(ctx, cacheKey, func() (interface{}, error) {
		results := sig.CallFunc(funcArgs)
		if len(results) == 2 && !results[1].IsNil() {
			return results[0].Interface(), results[1].Interface().(error)
		}
		return results[0].Interface(), nil
	})
}

// SetMethodConfig sets or updates the configuration for a specific method.
//
// Parameters:
//   - methodName: The method to configure
//   - config: The configuration to apply
func (cb *CachedBackend) SetMethodConfig(methodName string, config *MethodConfig) {
	cb.methodConfigs.Store(methodName, config)
}

// GetMethodSignature returns the cached method signature if it exists.
//
// Parameters:
//   - methodName: The method name to look up
//
// Returns:
//   - *methodSignature: The cached signature or nil if not found
func (cb *CachedBackend) GetMethodSignature(methodName string) *methodSignature {
	if sig, ok := cb.methodSignatures.Load(methodName); ok {
		return sig.(*methodSignature)
	}
	return nil
}
