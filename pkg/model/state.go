package model

import "time"

type StateType string

const (
	StateBasic  = StateType("StateConfig")
	StateLM     = StateType("StateConfigLM")
	StateDB     = StateType("StateConfigDB")
	StateCode   = StateType("StateConfigCode")
	StateVisual = StateType("StateConfigVisual")
	StateStream = StateType("StateConfigStream")
)

type State struct {
	ID                string                               `gorm:"primaryKey" json:"id,omitempty"`
	ProjectID         string                               `json:"project_id,omitempty"`
	Config            interface{}                          `json:"config,omitempty"` // You may define more specific types here
	Columns           map[string]StateDataColumnDefinition `json:"columns" gorm:"-"`
	Data              map[string]StateDataRowColumnData    `json:"data" gorm:"-"`
	Mapping           map[string]StateDataColumnIndex      `json:"mapping" gorm:"-"`
	Count             int                                  `json:"count" gorm:"default:0"`
	PersistedPosition int                                  `json:"persisted_position,omitempty" gorm:"default:0"`
	CreateDate        time.Time                            `json:"create_date,omitempty"`
	UpdateDate        time.Time                            `json:"update_date,omitempty"`
	StateType         string                               `json:"state_type,omitempty"`
}

// TableName sets the table name for the State struct
func (State) TableName() string {
	return "state"
}

type StateDataColumnDefinition struct {
	ID               int     `gorm:"primaryKey" json:"id,omitempty"`
	Name             string  `json:"name"`
	DataType         string  `json:"data_type" gorm:"default:str"`
	Required         *bool   `json:"required,omitempty" gorm:"default:true"`
	Callable         *bool   `json:"callable,omitempty" gorm:"default:false"`
	MinLength        *int    `json:"min_length,omitempty"`
	MaxLength        *int    `json:"max_length,omitempty"`
	Dimensions       *int    `json:"dimensions,omitempty"`
	Value            *string `json:"value,omitempty"`
	SourceColumnName *string `json:"source_column_name,omitempty"`
}

func (StateDataColumnDefinition) TableName() string {
	return "state_column"
}

type StateDataRowColumnData struct {
	Values []interface{} `json:"values"` // interface{} for flexible types
	Count  int           `json:"count" gorm:"default:0"`
}

func (StateDataRowColumnData) TableName() string {
	return "state_column_data"
}

type StateDataColumnIndex struct {
	Key    string        `json:"key"`
	Values []interface{} `json:"values"` // interface{} for flexible types
}
