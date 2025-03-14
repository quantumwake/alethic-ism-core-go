package models

// StateAttribute
// TODO this hold thing needs to be refactored, lots of code debt from original ism runtime; note this will impact the python ism modules and pretty much every state configuration, but it can be managed.
type StateAttribute string

const (
	// for language models
	AttributeUserTemplate   = StateAttribute("user_template_id")
	AttributeSystemTemplate = StateAttribute("system_template_id")

	AttributeFlagRequirePrimaryKey             = StateAttribute("flag_require_primary_key")
	AttributeFlagAppendToSession               = StateAttribute("flag_append_to_session")
	AttributeFlagDeDupDropEnabled              = StateAttribute("flag_dedup_drop_enabled")
	AttributeFlagQueryStateInheritanceAll      = StateAttribute("flag_query_state_inheritance_all")
	AttributeFlagQueryStateInheritanceInverse  = StateAttribute("flag_query_state_inheritance_inverse")
	AttributeFlagAutoSaveOutputState           = StateAttribute("flag_auto_save_output_state")
	AttributeFlagAutoRouteOutputState          = StateAttribute("flag_auto_route_output_state")
	AttributeFlagAutoRouteOutputStateAfterSave = StateAttribute("flag_auto_route_output_state_after_save")
)

// BaseConfig is equivalent to Python's BaseStateConfig
type BaseConfig struct {
	Name         string `json:"name,omitempty"`
	StorageClass string `json:"storage_class,omitempty"` // Default: "database"

}

// StateConfig is equivalent to StateConfig definition as defined in the alethic-ism-core (python) module, but slightly different representation.
// TODO probably rip the state config out and replace it with something more robust and easier to understand.
type StateConfig struct {
	BaseConfig
	DataKeyDefinitions map[DefinitionType][]*DataKeyDefinition `json:"key_definitions,omitempty"`
	Attributes         map[StateAttribute]any                  `json:"attributes"` /// TODO maybe use the ConfigMap (from the vault pkg) instead of having a separate state config map

	//PrimaryKey                        []*DataKeyDefinition `json:"primary_key,omitempty"`
	//QueryStateInheritance             []*DataKeyDefinition `json:"query_state_inheritance,omitempty"`
	//RemapQueryStateColumns            []*DataKeyDefinition `json:"remap_query_state_columns,omitempty"`
	//TemplateColumns                   []*DataKeyDefinition `json:"template_columns,omitempty"`

}

// GetDataKeyDefinitions fetches a list of data key definitions, e.g., primary key data fields from
// the map of available definitions, provided the respective state type defines the definition.
//
// Note, data key definitions are generally optional and depends on whether the state is an output of a certain type
// of processor. For example, in the case of online join functions, the primary key needs to be set such that the
// processor understands what fields to use to join the two stream together.
//
// TODO: will likely change this completely, for now it will do (e.g., use a json block to represent the state configuration)
func (sc *StateConfig) GetDataKeyDefinitions(definitionType DefinitionType) []*DataKeyDefinition {
	definitions, ok := sc.DataKeyDefinitions[definitionType]
	if ok {
		return definitions
	}
	return nil
}

type StateConfigAttribute struct {
	StateID   string         `gorm:"column:state_id;type:varchar(255);not null"`
	Attribute StateAttribute `gorm:"column:attribute;type:varchar(255);not null"`
	Data      string         `gorm:"column:data;type:text"` // TODO probably should be a json or *bytes?
}

// TableName sets the table name for the StateConfig struct
func (StateConfigAttribute) TableName() string {
	return "state_config"
}
