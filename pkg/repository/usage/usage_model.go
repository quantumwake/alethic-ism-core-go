package usage

import (
	"encoding/json"
	"time"
)

type UnitType string
type UnitSubType string

const (
	UnitToken   UnitType = "TOKEN"
	UnitCompute          = "COMPUTE"
)

const (
	UnitInput  UnitSubType = "INPUT"
	UnitOutput             = "OUTPUT"
)

// Usage struct represents a usage record in the database for a given project and resource type.
type Usage struct {
	ID              int       `gorm:"not null; column:id; primaryKey" json:"id"`
	TransactionTime time.Time `gorm:"not null; column:transaction_time" json:"transaction_time"`

	// Project Reference
	ProjectID string `gorm:"not null; column:project_id" json:"project_id"`

	// unique resource information such that we can reference it in billing (TODO resource should probably only be logically deleted, e.g. a datasource id or a processor id or compute node id, or something else)
	ResourceID   string `gorm:"not null; column:resource_id; size:255" json:"resource_id"`
	ResourceType string `gorm:"not null; column:resource_type; size:255" json:"resource_type"`

	UnitType    UnitType    `gorm:"not null; column:unit_type" json:"unit_type"`
	UnitSubType UnitSubType `gorm:"not null; column:unit_subtype" json:"unit_subtype"`
	UnitCount   int         `gorm:"not null; column:unit_count" json:"unit_count"`

	Metadata json.RawMessage `gorm:"null; column:metadata" json:"metadata"`
}

// TableName sets the table name for the Usage struct
func (*Usage) TableName() string {
	return "usage"
}
