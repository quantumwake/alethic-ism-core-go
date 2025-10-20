package nats

import (
	"encoding/json"
	"fmt"
	"github.com/quantumwake/alethic-ism-core-go/pkg/data"
	"github.com/quantumwake/alethic-ism-core-go/pkg/repository/processor"
)

// LoadConfigFromProvider loads NATS routing configuration from a provider's routing field
// The provider's Routing field should contain a single route configuration:
//
// Example routing JSON structure (single route per provider):
//
//	{
//	  "selector": "task.process",
//	  "name": "task-stream",      // optional - JetStream stream name
//	  "queue": "task-workers",     // optional - queue group name
//	  "subject": "task.process",
//	  "url": "nats://localhost:4222"
//	}
func LoadConfigFromProvider(provider *processor.Provider) (*Config, error) {
	if provider == nil {
		return nil, fmt.Errorf("provider cannot be nil")
	}

	// Check if routing field is populated
	if provider.Routing == nil || len(provider.Routing) == 0 {
		return nil, fmt.Errorf("provider %s has no routing configuration", provider.ID)
	}

	// Unmarshal the routing JSON into Config
	config, err := UnmarshalProviderRouting(provider.Routing)
	if err != nil {
		return nil, fmt.Errorf("failed to unmarshal routing config for provider %s: %w", provider.ID, err)
	}

	return config, nil
}

// UnmarshalProviderRouting converts the provider's routing JSON field (data.JSON) into a Config struct
// The routing JSON should contain a single NatConfig object (not an array)
// This creates a Config with a single route for compatibility with the existing Config structure
func UnmarshalProviderRouting(routingJSON data.JSON) (*Config, error) {
	// Marshal the routing JSON to bytes
	jsonBytes, err := json.Marshal(routingJSON)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal routing JSON: %w", err)
	}

	// Unmarshal into a single NatConfig
	var natConfig NatConfig
	if err := json.Unmarshal(jsonBytes, &natConfig); err != nil {
		return nil, fmt.Errorf("failed to unmarshal into NatConfig: %w", err)
	}

	// Create a Config with this single route
	config := &Config{
		MessageConfig: MessageConfig{
			Routes: []NatConfig{natConfig},
		},
	}

	// Build the route maps for fast lookups
	config.BuildRouteMaps()

	return config, nil
}