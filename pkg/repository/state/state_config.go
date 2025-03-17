package state

// BaseConfig is equivalent to Python's BaseStateConfig
type BaseConfig struct {
	Name         string `json:"name,omitempty"`
	StorageClass string `json:"storage_class,omitempty"` // Default: "database"

}

type ColumnKeyDefinitions map[DefinitionType][]*ColumnKeyDefinition

//type ConfigAttributes map[StateAttribute]string

// Config is equivalent to StateConfig definition as defined in the alethic-ism-core (python) module, but slightly different representation.
// TODO probably rip the state config out and replace it with something more robust and easier to understand.
type Config struct {
	BaseConfig
	KeyDefinitions ColumnKeyDefinitions `json:"key_definitions,omitempty"`
	Attributes     ConfigAttributes     `json:"attributes"` /// TODO maybe use the ConfigMap (from the vault pkg) instead of having a separate state config map

	//PrimaryKey                        []*StateColumnKeyDefinition `json:"primary_key,omitempty"`
	//QueryStateInheritance             []*StateColumnKeyDefinition `json:"query_state_inheritance,omitempty"`
	//RemapQueryStateColumns            []*StateColumnKeyDefinition `json:"remap_query_state_columns,omitempty"`
	//TemplateColumns                   []*StateColumnKeyDefinition `json:"template_columns,omitempty"`

}

// GetDataKeyDefinitions fetches a list of data key definitions, e.g., primary key data fields from
// the map of available definitions, provided the respective state type defines the definition.
//
// Note, data key definitions are generally optional and depends on whether the state is an output of a certain type
// of processor. For example, in the case of online join functions, the primary key needs to be set such that the
// processor understands what fields to use to join the two stream together.
//
// TODO: will likely change this completely, for now it will do (e.g., use a json block to represent the state configuration)
func (sc *Config) GetDataKeyDefinitions(definitionType DefinitionType) []*ColumnKeyDefinition {
	definitions, ok := sc.KeyDefinitions[definitionType]
	if ok {
		return definitions
	}
	return nil
}

func (sc *Config) BuildStateConfigAttributes(stateID string) []*ConfigAttribute {

	/// TODO build util map function
	attributes := make([]*ConfigAttribute, 0, len(sc.Attributes))
	for _, attr := range sc.Attributes {
		attributes = append(attributes, &ConfigAttribute{
			StateID:   stateID,
			Attribute: attr.Attribute,
			Data:      attr.Data,
		})
	}

	return attributes
}
