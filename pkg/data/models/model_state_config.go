package models

// BaseConfig is equivalent to Python's BaseStateConfig
type BaseConfig struct {
	Name                 string `json:"name,omitempty"`
	StorageClass         string `json:"storage_class,omitempty"` // Default: "database"
	FlagAppendToSession  *bool  `json:"flag_append_to_session,omitempty"`
	FlagDedupDropEnabled *bool  `json:"flag_dedup_drop_enabled,omitempty"`
}

// Config is equivalent to Python's StateConfig
type Config struct {
	BaseConfig
	PrimaryKey                        []*DataKeyDefinition `json:"primary_key,omitempty"`
	QueryStateInheritance             []*DataKeyDefinition `json:"query_state_inheritance,omitempty"`
	RemapQueryStateColumns            []*DataKeyDefinition `json:"remap_query_state_columns,omitempty"`
	TemplateColumns                   []*DataKeyDefinition `json:"template_columns,omitempty"`
	FlagRequirePrimaryKey             *bool                `json:"flag_require_primary_key,omitempty"`
	FlagQueryStateInheritanceAll      *bool                `json:"flag_query_state_inheritance_all,omitempty"`
	FlagQueryStateInheritanceInverse  *bool                `json:"flag_query_state_inheritance_inverse,omitempty"`
	FlagAutoSaveOutputState           *bool                `json:"flag_auto_save_output_state,omitempty"`
	FlagAutoRouteOutputState          *bool                `json:"flag_auto_route_output_state,omitempty"`
	FlagAutoRouteOutputStateAfterSave *bool                `json:"flag_auto_route_output_state_after_save,omitempty"`
}

func (Config) TableName() string {
	return "state_config"
}
