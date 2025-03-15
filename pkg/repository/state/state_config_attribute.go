package state

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

type StateConfigAttribute struct {
	StateID   string         `gorm:"column:state_id;type:varchar(255);not null"`
	Attribute StateAttribute `gorm:"column:attribute;type:varchar(255);not null"`
	Data      string         `gorm:"column:data;type:text"` // TODO probably should be a json or *bytes?
}

// TableName sets the table name for the StateConfig struct
func (StateConfigAttribute) TableName() string {
	return "state_config"
}
