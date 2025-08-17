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