package state

import "github.com/quantumwake/alethic-ism-core-go/pkg/repository/state"

// BaseConfig is equivalent to Python's BaseStateConfig
type BaseConfig struct {
	Name         string `json:"name,omitempty"`
	StorageClass string `json:"storage_class,omitempty"` // Default: "database"

}

// StateConfig is equivalent to StateConfig definition as defined in the alethic-ism-core (python) module, but slightly different representation.
// TODO probably rip the state config out and replace it with something more robust and easier to understand.
type StateConfig struct {
	BaseConfig
	DataKeyDefinitions map[state.DefinitionType][]*state.DataKeyDefinition `json:"key_definitions,omitempty"`
	Attributes         map[state.StateAttribute]any                        `json:"attributes"` /// TODO maybe use the ConfigMap (from the vault pkg) instead of having a separate state config map

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
func (sc *StateConfig) GetDataKeyDefinitions(definitionType state.DefinitionType) []*state.DataKeyDefinition {
	definitions, ok := sc.DataKeyDefinitions[definitionType]
	if ok {
		return definitions
	}
	return nil
}
