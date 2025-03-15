package state

// DefinitionType is equivalent to Python's StateDefinitionType
type DefinitionType string

const (
	DefinitionPrimaryKey             DefinitionType = "primary_key"
	DefinitionQueryStateInheritance  DefinitionType = "query_state_inheritance"
	DefinitionRemapQueryStateColumns DefinitionType = "remap_query_state_columns"
	DefinitionTemplateColumns        DefinitionType = "template_columns"
)

// DataKeyDefinition is equivalent to Python's StateDataKeyDefinition
type DataKeyDefinition struct {
	ID             *int64         `gorm:"column:id;primaryKey;autoIncrement" json:"id,omitempty"`
	StateID        string         `gorm:"column:state_id;type:varchar(36);not null" json:"state_id,omitempty"`
	Name           string         `gorm:"column:name;type:varchar(255);not null" json:"name"`
	Alias          string         `gorm:"column:alias;type:varchar(255);null" json:"alias,omitempty"`
	Required       *bool          `gorm:"column:required;type:boolean;default:true" json:"required,omitempty"`
	Callable       *bool          `gorm:"column:callable;type:boolean;default:false" json:"callable,omitempty"`
	DefinitionType DefinitionType `gorm:"column:definition_type;type:varchar(50);not null" json:"definition_type"`
}

// TableName returns the table name for the DataKeyDefinition model
func (DataKeyDefinition) TableName() string {
	return "state_column_key_definition"
}
