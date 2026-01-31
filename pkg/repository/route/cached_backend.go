package route

import (
	"context"
	"time"

	"github.com/quantumwake/alethic-ism-core-go/pkg/cache"
	"github.com/quantumwake/alethic-ism-core-go/pkg/repository/processor"
)

// CachedBackendStorage provides a caching layer over the route BackendStorage.
// Routes are relatively static once configured, making them ideal for caching.
// This implementation caches all read operations to reduce database load.
type CachedBackendStorage struct {
	*cache.CachedBackend                 // Embedded generic caching functionality
	base                 *BackendStorage // The underlying route backend
}

// DefaultConfig returns the default TTL configuration for route backend.
func DefaultConfig(baseTTL time.Duration) *cache.MethodTTLConfig {
	config := cache.NewMethodTTLConfig(baseTTL)

	// Routes are relatively static once configured
	config.SetMethodTTL("FindRouteByID", baseTTL)
	config.SetMethodTTL("FindRouteByProcessorAndDirection", baseTTL)
	config.SetMethodTTL("FindRouteByStateAndDirection", baseTTL)
	config.SetMethodTTL("FindRouteByState", baseTTL)
	config.SetMethodTTL("FindRouteWithOutputsByID", baseTTL)

	return config
}

// NewCachedBackend creates a new route backend with caching enabled.
// Uses the default route configuration with provided base TTL.
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

// NewCachedBackendWithConfig creates a route backend with custom TTL configuration.
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

// FindRouteByID retrieves a route by ID with caching.
// Individual routes are frequently accessed and benefit from caching.
// Cache key format: "FindRouteByID:hash(id)"
//
// Parameters:
//   - id: The route ID to find
//
// Returns:
//   - *processor.State: The route if found
//   - error: Database error if the operation fails
func (cb *CachedBackendStorage) FindRouteByID(id string) (*processor.State, error) {
	ctx := context.Background()

	return cache.CallCached(cb.CachedBackend, ctx, "FindRouteByID", []interface{}{id},
		func() (*processor.State, error) {
			return cb.base.FindRouteByID(id)
		})
}

// FindRouteByProcessorAndDirection retrieves routes by processor and direction with caching.
// This is commonly used to find all inputs or outputs for a processor.
// Cache key format: "FindRouteByProcessorAndDirection:hash(processorID, direction)"
//
// Parameters:
//   - processorID: The processor ID to find routes for
//   - direction: The direction (input/output) to filter by
//
// Returns:
//   - []processor.State: List of routes matching the criteria
//   - error: Database error if the operation fails
func (cb *CachedBackendStorage) FindRouteByProcessorAndDirection(processorID string, direction processor.StateDirection) ([]processor.State, error) {
	ctx := context.Background()

	return cache.CallCached(cb.CachedBackend, ctx, "FindRouteByProcessorAndDirection",
		[]interface{}{processorID, direction},
		func() ([]processor.State, error) {
			return cb.base.FindRouteByProcessorAndDirection(processorID, direction)
		})
}

// FindRouteByStateAndDirection retrieves routes by state and direction with caching.
// This helps trace data flow through the system.
// Cache key format: "FindRouteByStateAndDirection:hash(stateID, direction)"
//
// Parameters:
//   - stateID: The state ID to find routes for
//   - direction: The direction (input/output) to filter by
//
// Returns:
//   - []processor.State: List of routes matching the criteria
//   - error: Database error if the operation fails
func (cb *CachedBackendStorage) FindRouteByStateAndDirection(stateID string, direction processor.StateDirection) ([]processor.State, error) {
	ctx := context.Background()

	return cache.CallCached(cb.CachedBackend, ctx, "FindRouteByStateAndDirection",
		[]interface{}{stateID, direction},
		func() ([]processor.State, error) {
			return cb.base.FindRouteByStateAndDirection(stateID, direction)
		})
}

// FindRouteByState retrieves all routes for a state with caching.
// This provides all connections for a given state.
// Cache key format: "FindRouteByState:hash(stateID)"
//
// Parameters:
//   - stateID: The state ID to find routes for
//
// Returns:
//   - []processor.State: List of all routes for the state
//   - error: Database error if the operation fails
func (cb *CachedBackendStorage) FindRouteByState(stateID string) ([]processor.State, error) {
	ctx := context.Background()

	return cache.CallCached(cb.CachedBackend, ctx, "FindRouteByState",
		[]interface{}{stateID},
		func() ([]processor.State, error) {
			return cb.base.FindRouteByState(stateID)
		})
}

// FindRouteWithOutputsByID retrieves a route and its processor's outputs with caching.
// This composite method benefits from caching as it consolidates multiple queries.
// Cache key format: "FindRouteWithOutputsByID:hash(routeID)"
//
// Parameters:
//   - routeID: The route ID to find
//
// Returns:
//   - *processor.State: The input route
//   - []processor.State: Output routes for the processor
//   - error: Database error if the operation fails
func (cb *CachedBackendStorage) FindRouteWithOutputsByID(routeID string) (*processor.State, []processor.State, error) {
	ctx := context.Background()

	// We need to handle multiple return values, so we'll use a wrapper struct
	type routeWithOutputs struct {
		InputRoute   *processor.State
		OutputRoutes []processor.State
	}

	result, err := cache.CallCached(cb.CachedBackend, ctx, "FindRouteWithOutputsByID",
		[]interface{}{routeID},
		func() (*routeWithOutputs, error) {
			input, outputs, err := cb.base.FindRouteWithOutputsByID(routeID)
			if err != nil {
				return nil, err
			}
			return &routeWithOutputs{
				InputRoute:   input,
				OutputRoutes: outputs,
			}, nil
		})

	if err != nil {
		return nil, nil, err
	}

	return result.InputRoute, result.OutputRoutes, nil
}

// Access returns the underlying BackendStorage for direct database access.
// Use this when you need to bypass the cache layer, for example:
//   - During data migrations
//   - For administrative operations
//   - When debugging cache issues
//
// Returns:
//   - *BackendStorage: The underlying database backend
func (cb *CachedBackendStorage) Access() *BackendStorage {
	return cb.base
}
