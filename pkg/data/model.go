package data

import (
	"encoding/json"
	"time"
)

type UnitType string
type UnitSubType string

const (
	UNIT_TOKENS  UnitType = "TOKEN"
	UNIT_COMPUTE          = "COMPUTE"
)

const (
	UNIT_INPUT  UnitSubType = "INPUT"
	UNIT_OUTPUT             = "OUTPUT"
)

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
func (Usage) TableName() string {
	return "usage"
}

type ProcessorStateDirection string
type ProcessorStatus string

const (
	DirectionInput  ProcessorStateDirection = "INPUT"
	DirectionOutput ProcessorStateDirection = "OUTPUT"
)

type ProcessorState struct {
	ID           string                  `gorm:"column:id;type:varchar(73);default:gen_random_uuid();primaryKey"`
	InternalID   uint                    `gorm:"column:internal_id;autoIncrement;unique"`
	ProcessorID  string                  `gorm:"column:processor_id;type:varchar(36)"`
	StateID      string                  `gorm:"column:state_id;primaryKey;type:varchar(36)"`
	Direction    ProcessorStateDirection `gorm:"column:direction;primaryKey;type:processor_state_direction"`
	Status       ProcessorStatus         `gorm:"column:status;type:processor_status"`
	Count        *int                    `gorm:"column:count"`
	CurrentIndex *int                    `gorm:"column:current_index"`
	MaximumIndex *int                    `gorm:"column:maximum_index"`
}

// TableName overrides the table name used by GORM
func (ProcessorState) TableName() string {
	return "processor_state"
}
