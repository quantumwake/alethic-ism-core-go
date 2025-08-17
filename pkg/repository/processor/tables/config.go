package tables

// TableProcessorConfig defines configuration for the tables processor
type TableProcessorConfig struct {
	TableName        *string  `json:"tableName,omitempty"`
	BatchSize        *int     `json:"batchSize,omitempty"`
	BatchWindowTTL   *int     `json:"batchWindowTTL,omitempty"`
	IndexColumns     []string `json:"indexColumns,omitempty"`
	IncludeTimestamp *bool    `json:"includeTimestamp,omitempty"`
}

// DefaultTableProcessorConfig returns the default configuration
func DefaultTableProcessorConfig() *TableProcessorConfig {
	batchSize := 10
	ttl := 15
	tableName := "hello_world"
	includeTimestamp := true
	return &TableProcessorConfig{
		TableName:        &tableName,
		BatchSize:        &batchSize,
		BatchWindowTTL:   &ttl,
		IncludeTimestamp: &includeTimestamp,
		IndexColumns:     []string{},
	}
}
