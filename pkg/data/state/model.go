package state

import "time"

type Type string

const (
	StateBasic  = Type("StateConfig")
	StateLM     = Type("StateConfigLM")
	StateDB     = Type("StateConfigDB")
	StateCode   = Type("StateConfigCode")
	StateVisual = Type("StateConfigVisual")
	StateStream = Type("StateConfigStream")
)

type State struct {
	ID                string                          `gorm:"primaryKey" json:"id,omitempty"`
	ProjectID         string                          `json:"project_id,omitempty"`
	Config            interface{}                     `json:"config,omitempty"` // You may define more specific types here
	Columns           map[string]DataColumnDefinition `json:"columns" gorm:"-"`
	Data              map[string]DataRowColumnData    `json:"data" gorm:"-"`
	Mapping           map[string]DataColumnIndex      `json:"mapping" gorm:"-"`
	Count             int                             `json:"count" gorm:"default:0"`
	PersistedPosition int                             `json:"persisted_position,omitempty" gorm:"default:0"`
	CreateDate        time.Time                       `json:"create_date,omitempty"`
	UpdateDate        time.Time                       `json:"update_date,omitempty"`
	StateType         string                          `json:"state_type,omitempty"`
}

// TableName sets the table name for the State struct
func (State) TableName() string {
	return "state"
}

type DataColumnDefinition struct {
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

func (DataColumnDefinition) TableName() string {
	return "state_column"
}

type DataRowColumnData struct {
	Values []interface{} `json:"values"` // interface{} for flexible types
	Count  int           `json:"count" gorm:"default:0"`
}

func (DataRowColumnData) TableName() string {
	return "state_column_data"
}

type DataColumnIndex struct {
	Key    string        `json:"key"`
	Values []interface{} `json:"values"` // interface{} for flexible types
}
