package cache

import (
	"time"
)

// MethodTTLConfig defines TTL configuration for backend methods.
// This allows centralized configuration of cache behavior without hardcoding in implementations.
type MethodTTLConfig struct {
	// Method name to TTL mapping
	Methods map[string]time.Duration
	
	// Default TTL to use if method not found in map
	DefaultTTL time.Duration
}

// NewMethodTTLConfig creates a new method TTL configuration with the given default TTL.
func NewMethodTTLConfig(defaultTTL time.Duration) *MethodTTLConfig {
	return &MethodTTLConfig{
		Methods:    make(map[string]time.Duration),
		DefaultTTL: defaultTTL,
	}
}

// SetMethodTTL sets the TTL for a specific method.
func (c *MethodTTLConfig) SetMethodTTL(method string, ttl time.Duration) {
	c.Methods[method] = ttl
}

// GetMethodTTL returns the TTL for a specific method, or the default if not configured.
func (c *MethodTTLConfig) GetMethodTTL(method string) time.Duration {
	if ttl, ok := c.Methods[method]; ok {
		return ttl
	}
	return c.DefaultTTL
}

// ApplyToBackend applies the TTL configuration to a CachedBackend.
// This sets up MethodConfig for each configured method.
func (c *MethodTTLConfig) ApplyToBackend(backend *CachedBackend) {
	for method, ttl := range c.Methods {
		backend.SetMethodConfig(method, &MethodConfig{
			TTL:       ttl,
			Cacheable: true,
		})
	}
}

// DefaultProcessorConfig returns the default TTL configuration for processor backend.
func DefaultProcessorConfig(baseTTL time.Duration) *MethodTTLConfig {
	config := NewMethodTTLConfig(baseTTL)
	
	// Provider classes are static configuration
	config.SetMethodTTL("FindProviderClasses", 10*time.Minute)
	
	// Providers change less frequently
	config.SetMethodTTL("FindProviders", 5*time.Minute)
	config.SetMethodTTL("FindProviderByClass", 5*time.Minute)
	config.SetMethodTTL("FindProviderByClassUserAndProject", 5*time.Minute)
	
	// Processors are accessed frequently but change occasionally
	config.SetMethodTTL("FindProcessorByID", baseTTL)
	config.SetMethodTTL("FindProcessorByProjectID", baseTTL)
	
	return config
}

// DefaultRouteConfig returns the default TTL configuration for route backend.
func DefaultRouteConfig(baseTTL time.Duration) *MethodTTLConfig {
	config := NewMethodTTLConfig(baseTTL)
	
	// Routes are relatively static once configured
	config.SetMethodTTL("FindRouteByID", baseTTL)
	config.SetMethodTTL("FindRouteByProcessorAndDirection", baseTTL)
	config.SetMethodTTL("FindRouteByStateAndDirection", baseTTL)
	config.SetMethodTTL("FindRouteByState", baseTTL)
	config.SetMethodTTL("FindRouteWithOutputsByID", baseTTL)
	
	return config
}

// DefaultUserConfig returns the default TTL configuration for user backend.
func DefaultUserConfig(baseTTL time.Duration) *MethodTTLConfig {
	config := NewMethodTTLConfig(baseTTL)
	
	// User profiles are very stable
	config.SetMethodTTL("FindUserByID", 15*time.Minute)
	
	return config
}

// DefaultProjectConfig returns the default TTL configuration for project backend.
func DefaultProjectConfig(baseTTL time.Duration) *MethodTTLConfig {
	config := NewMethodTTLConfig(baseTTL)
	
	// Projects change occasionally
	config.SetMethodTTL("FindByID", baseTTL)
	config.SetMethodTTL("FindAllByUserID", baseTTL+2*time.Minute) // Slightly longer for lists
	
	return config
}