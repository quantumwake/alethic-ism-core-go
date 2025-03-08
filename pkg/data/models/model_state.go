package models

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

type State0 struct {
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

type StateType string

const (
	StateConfig          = StateType("StateConfig")
	StateConfigLM        = StateType("StateConfigLM")
	StateConfigVisual    = StateType("StateConfigVisual")
	StateConfigCode      = StateType("StateConfigCode")
	StateConfigStream    = StateType("StateConfigStream")
	StateConfigAudio     = StateType("StateConfigAudio")
	StateConfigUserInput = StateType("StateConfigUserInput")
)

type State struct {
	ID        string    `gorm:"column:id;type:varchar(36);primaryKey;not null"`
	ProjectID string    `gorm:"column:project_id;type:varchar(36);not null"`
	StateType StateType `gorm:"column:state_type;type:varchar(32);null"` // Allow null string
	Count     int       `gorm:"column:count;type:integer;not null;default:0"`

	// many-to-one references
	Config  interface{}                      `gorm:"-"` // Allow null JSON
	Columns map[string]*DataColumnDefinition `gorm:"-"` // Ignored by GORM
	Data    map[string]*DataRowColumnData    `gorm:"-"` // Ignored by GORM
	Mapping map[string]*DataColumnIndex      `gorm:"-"` // Ignored by GORM
}

// TableName sets the table name for the State struct
func (State) TableName() string {
	return "state"
}

// DataColumnDefinition represents the structure of the state_column table
type DataColumnDefinition struct {
	// Column: id
	// Type: BIGINT (int64)
	// Constraints: PRIMARY KEY, NOT NULL, AUTO INCREMENT
	ID int64 `gorm:"column:id;primaryKey;autoIncrement;type:bigint;not null" json:"id,omitempty"`

	// Column: state_id
	// Type: VARCHAR(36)
	// Constraints: NOT NULL
	StateID string `gorm:"column:state_id;type:varchar(36);not null" json:"state_id,omitempty"`

	// Column: name
	// Type: VARCHAR(255)
	// Constraints: NOT NULL
	Name string `gorm:"column:name;type:varchar(255);not null" json:"name"`

	// Column: data_type
	// Type: VARCHAR(50)
	// Constraints: NOT NULL, DEFAULT 'str'
	DataType string `gorm:"column:data_type;type:varchar(50);not null;default:str" json:"data_type"`

	// Column: required
	// Type: BOOLEAN
	// Constraints: NULL, DEFAULT TRUE
	Required *bool `gorm:"column:required;type:boolean;default:true" json:"required,omitempty"`

	// Column: callable
	// Type: BOOLEAN
	// Constraints: NULL, DEFAULT FALSE
	Callable *bool `gorm:"column:callable;type:boolean;default:false" json:"callable,omitempty"`

	// Column: min_length
	// Type: INTEGER
	// Constraints: NULL
	MinLength *int `gorm:"column:min_length;type:integer;null" json:"min_length,omitempty"`

	// Column: max_length
	// Type: INTEGER
	// Constraints: NULL
	MaxLength *int `gorm:"column:max_length;type:integer;null" json:"max_length,omitempty"`

	// Column: dimensions
	// Type: INTEGER
	// Constraints: NULL
	Dimensions *int `gorm:"column:dimensions;type:integer;null" json:"dimensions,omitempty"`

	// Column: value
	// Type: TEXT
	// Constraints: NULL
	Value *string `gorm:"column:value;type:text;null" json:"value,omitempty"`

	// Column: source_column_name
	// Type: VARCHAR(255)
	// Constraints: NULL
	SourceColumnName *string `gorm:"column:source_column_name;type:varchar(255);null" json:"source_column_name,omitempty"`
}

func (DataColumnDefinition) TableName() string {
	return "state_column"
}

type DataRowColumnData struct {
	Values []string `json:"values"` // interface{} for flexible types
	Count  int      `json:"count" gorm:"default:0"`
}

func (DataRowColumnData) TableName() string {
	return "state_column_data"
}

type DataColumnIndex struct {
	Key    string        `json:"key"`
	Values []interface{} `json:"values"` // interface{} for flexible types
}
