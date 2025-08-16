package processor

// JoinWindowConfig defines the window configuration for join processors
// This configuration controls the sliding window behavior for data correlation
type JoinWindowConfig struct {
	// BlockCountSoftLimit defines the maximum number of blocks before eviction starts
	BlockCountSoftLimit int `json:"blockCountSoftLimit"`
	
	// BlockWindowTTL defines the sliding window TTL for blocks (e.g., "1m", "5m")
	// This resets on each new event arrival
	BlockWindowTTL string `json:"blockWindowTTL"`
	
	// BlockPartMaxJoinCount defines the maximum number of joins allowed per data part
	// Once reached, the part is evicted
	BlockPartMaxJoinCount int `json:"blockPartMaxJoinCount"`
	
	// BlockPartMaxAge defines the absolute lifetime of a data part (e.g., "15s", "1m")
	// Parts are evicted after this duration regardless of activity
	BlockPartMaxAge string `json:"blockPartMaxAge"`
}

// DefaultJoinWindowConfig returns the default configuration for join processors
func DefaultJoinWindowConfig() *JoinWindowConfig {
	return &JoinWindowConfig{
		BlockCountSoftLimit:   10,
		BlockWindowTTL:        "1m",
		BlockPartMaxJoinCount: 1,
		BlockPartMaxAge:       "15s",
	}
}