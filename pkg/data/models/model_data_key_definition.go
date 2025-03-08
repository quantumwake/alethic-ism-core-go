package models

// DataKeyDefinition is equivalent to Python's StateDataKeyDefinition
type DataKeyDefinition struct {
	ID       *int64 `json:"id,omitempty"`
	Name     string `json:"name"`
	Alias    string `json:"alias,omitempty"`
	Required *bool  `json:"required,omitempty"` // Default: false
	Callable *bool  `json:"callable,omitempty"` // Default: false
}
