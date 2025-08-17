package join

// WindowConfig defines the window configuration for join processors.
// This configuration controls the sliding window behavior for data correlation.
//
// To configure a join processor, add these properties to the processor's Properties field:
//
//	{
//	  "blockCountSoftLimit": 10,
//	  "blockWindowTTL": "1m",
//	  "blockPartMaxJoinCount": 1,
//	  "blockPartMaxAge": "15s"
//	}
type WindowConfig struct {
	// BlockCountSoftLimit defines the maximum number of blocks before eviction starts
	BlockCountSoftLimit *int `json:"blockCountSoftLimit,omitempty"`

	// BlockWindowTTL defines the sliding window TTL for blocks (e.g., "1m", "5m")
	// This resets on each new event arrival
	BlockWindowTTL *string `json:"blockWindowTTL,omitempty"`

	// BlockPartMaxJoinCount defines the maximum number of joins allowed per data part
	// Once reached, the part is evicted
	BlockPartMaxJoinCount *int `json:"blockPartMaxJoinCount,omitempty"`

	// BlockPartMaxAge defines the absolute lifetime of a data part (e.g., "15s", "1m")
	// Parts are evicted after this duration regardless of activity
	BlockPartMaxAge *string `json:"blockPartMaxAge,omitempty"`
}

// DefaultWindowConfig returns the default configuration for join processors
func DefaultWindowConfig() *WindowConfig {
	limit := 10
	ttl := "1m"
	maxJoin := 1
	maxAge := "15s"
	
	return &WindowConfig{
		BlockCountSoftLimit:   &limit,
		BlockWindowTTL:        &ttl,
		BlockPartMaxJoinCount: &maxJoin,
		BlockPartMaxAge:       &maxAge,
	}
}
