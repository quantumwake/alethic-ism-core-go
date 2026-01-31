package processor

import (
	"database/sql/driver"
	"encoding/json"
	"fmt"
)

// EdgeFunctionType represents the type of edge function
type EdgeFunctionType string

const (
	EdgeFunctionTypeCalibrator  EdgeFunctionType = "CALIBRATOR"
	EdgeFunctionTypeValidator   EdgeFunctionType = "VALIDATOR"
	EdgeFunctionTypeTransformer EdgeFunctionType = "TRANSFORMER"
	EdgeFunctionTypeFilter      EdgeFunctionType = "FILTER"
)

// EdgeFunctionConfig represents the configuration for edge functions on processor state transitions
type EdgeFunctionConfig struct {
	Enabled      bool                   `json:"enabled"`
	FunctionType string                 `json:"function_type"`
	TemplateID   *string                `json:"template_id,omitempty"`
	MaxAttempts  int                    `json:"max_attempts"`
	Config       map[string]interface{} `json:"config,omitempty"`
}

// Scan implements the sql.Scanner interface for GORM
func (e *EdgeFunctionConfig) Scan(value interface{}) error {
	if value == nil {
		return nil
	}

	bytes, ok := value.([]byte)
	if !ok {
		return fmt.Errorf("failed to scan EdgeFunctionConfig: expected []byte, got %T", value)
	}

	if len(bytes) == 0 {
		return nil
	}

	return json.Unmarshal(bytes, e)
}

// Value implements the driver.Valuer interface for GORM
func (e EdgeFunctionConfig) Value() (driver.Value, error) {
	if !e.Enabled {
		return nil, nil
	}
	return json.Marshal(e)
}
