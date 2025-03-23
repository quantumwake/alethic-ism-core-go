package state

// DefinitionType is equivalent to Python's StateDefinitionType
type DefinitionType string

const (
	DefinitionPrimaryKey             DefinitionType = "primary_key"
	DefinitionStateJoinKey           DefinitionType = "state_join_key"
	DefinitionQueryStateInheritance  DefinitionType = "query_state_inheritance"
	DefinitionRemapQueryStateColumns DefinitionType = "remap_query_state_columns"
	DefinitionTemplateColumns        DefinitionType = "template_columns"
)

// ColumnKeyDefinition is equivalent to Python's StateDataKeyDefinition
type ColumnKeyDefinition struct {
	ID      *int64 `gorm:"column:id;primaryKey;autoIncrement" json:"id,omitempty"`
	StateID string `gorm:"column:state_id;type:varchar(36);not null" json:"state_id,omitempty"`
	Name    string `gorm:"column:name;type:varchar(255);not null" json:"name"`

	// TODO do we need alias?
	Alias string `gorm:"column:alias;type:varchar(255);null" json:"alias,omitempty"`

	// TODO do we need these?
	Required *bool `gorm:"column:required;type:boolean;default:true" json:"required,omitempty"`

	// TODO - I suppose a key column can be a callable field where the value is derived off of a function call.
	Callable       *bool          `gorm:"column:callable;type:boolean;default:false" json:"callable,omitempty"`
	DefinitionType DefinitionType `gorm:"column:definition_type;type:varchar(50);not null" json:"definition_type"`
}

// TableName returns the table name for the StateColumnKeyDefinition model
func (ColumnKeyDefinition) TableName() string {
	return "state_column_key_definition"
}
