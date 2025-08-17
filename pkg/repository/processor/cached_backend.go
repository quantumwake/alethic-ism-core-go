package processor

import (
	"context"
	"time"

	"github.com/quantumwake/alethic-ism-core-go/pkg/cache"
)

// CachedBackendStorage provides a caching layer over the processor BackendStorage.
// It intercepts all read operations and caches their results to reduce database load.
// Write operations automatically invalidate relevant cache entries to maintain consistency.
// This implementation uses the generic cache package, making it easy to switch cache backends.
type CachedBackendStorage struct {
	*cache.CachedBackend               // Embedded generic caching functionality
	base *BackendStorage               // The underlying processor backend
}

// NewCachedBackend creates a new processor backend with caching enabled.
// By default, cache entries expire after 5 minutes.
//
// Parameters:
//   - dsn: Database connection string
//   - c: Cache implementation (pass nil for default in-memory cache)
//
// Returns:
//   - A new CachedBackendStorage instance with caching enabled
//
// Example:
//   cache := cache.NewLocalCache(nil)
//   backend := NewCachedBackend(dsn, cache)
func NewCachedBackend(dsn string, c cache.Cache) *CachedBackendStorage {
	base := NewBackend(dsn)
	cachedBackend := cache.NewCachedBackend(base, c, 5*time.Minute)
	
	return &CachedBackendStorage{
		CachedBackend: cachedBackend,
		base:          base,
	}
}

// NewCachedBackendWithTTL creates a new processor backend with custom cache TTL.
// Use this when you need different cache expiration times.
//
// Parameters:
//   - dsn: Database connection string
//   - c: Cache implementation (pass nil for default in-memory cache)
//   - ttl: Custom time-to-live for all cached entries
//
// Returns:
//   - A new CachedBackendStorage instance with custom TTL
//
// Example:
//   backend := NewCachedBackendWithTTL(dsn, cache, 10*time.Minute)
func NewCachedBackendWithTTL(dsn string, c cache.Cache, ttl time.Duration) *CachedBackendStorage {
	base := NewBackend(dsn)
	cachedBackend := cache.NewCachedBackend(base, c, ttl)
	
	return &CachedBackendStorage{
		CachedBackend: cachedBackend,
		base:          base,
	}
}

// FindProcessorByID retrieves a processor by ID with caching.
// This is one of the most frequently called methods, so caching provides significant benefits.
// Cache key format: "FindProcessorByID:hash(processorID)"
//
// Parameters:
//   - processorID: The UUID of the processor to find
//
// Returns:
//   - *Processor: The processor if found
//   - error: Database error if the operation fails
func (cb *CachedBackendStorage) FindProcessorByID(processorID string) (*Processor, error) {
	ctx := context.Background()
	
	return cache.CallCached(cb.CachedBackend, ctx, "FindProcessorByID", []interface{}{processorID}, 
		func() (*Processor, error) {
			return cb.base.FindProcessorByID(processorID)
		})
}

// FindProcessorByProjectID retrieves all processors for a project with caching.
// Results are cached to avoid repeated database queries for project processor lists.
// Cache key format: "FindProcessorByProjectID:hash(projectID)"
//
// Parameters:
//   - projectID: The project ID to find processors for
//
// Returns:
//   - []*Processor: List of processors belonging to the project
//   - error: Database error if the operation fails
func (cb *CachedBackendStorage) FindProcessorByProjectID(projectID string) ([]*Processor, error) {
	ctx := context.Background()
	
	return cache.CallCached(cb.CachedBackend, ctx, "FindProcessorByProjectID", []interface{}{projectID},
		func() ([]*Processor, error) {
			return cb.base.FindProcessorByProjectID(projectID)
		})
}

// FindProviders retrieves providers for a user and/or project with caching.
// Provider configurations don't change frequently, making them ideal for caching.
// Cache key format: "FindProviders:hash(userID, projectID)"
//
// Parameters:
//   - userID: Optional user ID filter (nil for all users)
//   - projectID: Optional project ID filter (nil for all projects)
//
// Returns:
//   - []*Provider: List of matching providers
//   - error: Database error if the operation fails
func (cb *CachedBackendStorage) FindProviders(userID, projectID *string) ([]*Provider, error) {
	ctx := context.Background()
	
	// Convert pointers to values for consistent cache keys
	var userVal, projectVal string
	if userID != nil {
		userVal = *userID
	}
	if projectID != nil {
		projectVal = *projectID
	}
	
	return cache.CallCached(cb.CachedBackend, ctx, "FindProviders", []interface{}{userVal, projectVal},
		func() ([]*Provider, error) {
			return cb.base.FindProviders(userID, projectID)
		})
}

// FindProviderByClassUserAndProject retrieves providers by class, user, and project with caching.
// This method is used to find specific provider implementations for a given context.
// Cache key format: "FindProviderByClassUserAndProject:hash(className, userID, projectID)"
//
// Parameters:
//   - className: The provider class to filter by
//   - userID: Optional user ID filter (nil for all users)
//   - projectID: Optional project ID filter (nil for all projects)
//
// Returns:
//   - []*Provider: List of matching providers
//   - error: Database error if the operation fails
func (cb *CachedBackendStorage) FindProviderByClassUserAndProject(className Class, userID, projectID *string) ([]*Provider, error) {
	ctx := context.Background()
	
	// Convert pointers to values for consistent cache keys
	var userVal, projectVal string
	if userID != nil {
		userVal = *userID
	}
	if projectID != nil {
		projectVal = *projectID
	}
	
	return cache.CallCached(cb.CachedBackend, ctx, "FindProviderByClassUserAndProject", 
		[]interface{}{className, userVal, projectVal},
		func() ([]*Provider, error) {
			return cb.base.FindProviderByClassUserAndProject(className, userID, projectID)
		})
}

// FindProviderByClass retrieves all providers of a specific class with caching.
// This is a convenience method that finds providers regardless of user or project.
// Cache key format: "FindProviderByClass:hash(className)"
//
// Parameters:
//   - className: The provider class to filter by
//
// Returns:
//   - []*Provider: List of providers with the specified class
//   - error: Database error if the operation fails
func (cb *CachedBackendStorage) FindProviderByClass(className Class) ([]*Provider, error) {
	ctx := context.Background()
	
	return cache.CallCached(cb.CachedBackend, ctx, "FindProviderByClass", []interface{}{className},
		func() ([]*Provider, error) {
			return cb.base.FindProviderByClass(className)
		})
}

// FindProviderClasses retrieves all available provider classes with extended caching.
// Provider classes are static configuration that rarely changes, so we use a longer TTL (10 minutes).
// Cache key format: "FindProviderClasses:hash()"
//
// Returns:
//   - []ProviderClass: List of all provider classes
//   - error: Database error if the operation fails
func (cb *CachedBackendStorage) FindProviderClasses() ([]ProviderClass, error) {
	ctx := context.Background()
	
	// Use longer TTL for static configuration data
	return cache.CallCachedWithTTL(cb.CachedBackend, ctx, "FindProviderClasses", []interface{}{}, 
		10*time.Minute,
		func() ([]ProviderClass, error) {
			return cb.base.FindProviderClasses()
		})
}

// CreateOrUpdate creates or updates a processor and invalidates related cache entries.
// This ensures that subsequent reads will get the updated data from the database.
// Cache invalidation is precise - only affected entries are removed.
//
// Parameters:
//   - processor: The processor to create or update
//
// Returns:
//   - error: Database error if the operation fails
//
// Cache invalidation:
//   - Invalidates FindProcessorByID cache for this processor
//   - Invalidates FindProcessorByProjectID cache for the project
func (cb *CachedBackendStorage) CreateOrUpdate(processor *Processor) error {
	// Perform the database operation first
	err := cb.base.CreateOrUpdate(processor)
	if err != nil {
		return err
	}
	
	ctx := context.Background()
	
	// Invalidate cache entries that would be affected by this change
	_ = cb.InvalidateMethod(ctx, "FindProcessorByID", processor.ID)
	_ = cb.InvalidateMethod(ctx, "FindProcessorByProjectID", processor.ProjectID)
	
	return nil
}

// CreateOrUpdateProvider creates or updates a provider and invalidates related cache entries.
// Provider changes can affect multiple cache entries, so we invalidate all related queries.
//
// Parameters:
//   - provider: The provider to create or update
//
// Returns:
//   - error: Database error if the operation fails
//
// Cache invalidation:
//   - Invalidates FindProviders cache for the user/project combination
//   - Invalidates FindProviderByClass cache for the provider's class
//   - Invalidates FindProviderByClassUserAndProject cache for the specific combination
func (cb *CachedBackendStorage) CreateOrUpdateProvider(provider *Provider) error {
	// Perform the database operation first
	err := cb.base.CreateOrUpdateProvider(provider)
	if err != nil {
		return err
	}
	
	ctx := context.Background()
	
	// Convert pointers to values for cache key generation
	var userVal, projectVal string
	if provider.UserID != nil {
		userVal = *provider.UserID
	}
	if provider.ProjectID != nil {
		projectVal = *provider.ProjectID
	}
	
	// Invalidate all cache entries that could be affected by this provider change
	_ = cb.InvalidateMethod(ctx, "FindProviders", userVal, projectVal)
	_ = cb.InvalidateMethod(ctx, "FindProviderByClass", provider.ClassName)
	_ = cb.InvalidateMethod(ctx, "FindProviderByClassUserAndProject", provider.ClassName, userVal, projectVal)
	
	return nil
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