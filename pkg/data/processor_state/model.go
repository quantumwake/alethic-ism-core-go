package processor_state

type ProcessorStateDirection string
type ProcessorStatus string

const (
	DirectionInput  ProcessorStateDirection = "INPUT"
	DirectionOutput ProcessorStateDirection = "OUTPUT"
)

type ProcessorState struct {
	ID           string                  `gorm:"column:id;type:varchar(73);default:gen_random_uuid();primaryKey"`
	InternalID   uint                    `gorm:"column:internal_id;autoIncrement;unique"`
	ProcessorID  string                  `gorm:"column:processor_id;type:varchar(36)"`
	StateID      string                  `gorm:"column:state_id;primaryKey;type:varchar(36)"`
	Direction    ProcessorStateDirection `gorm:"column:direction;primaryKey;type:processor_state_direction"`
	Status       ProcessorStatus         `gorm:"column:status;type:processor_status"`
	Count        *int                    `gorm:"column:count"`
	CurrentIndex *int                    `gorm:"column:current_index"`
	MaximumIndex *int                    `gorm:"column:maximum_index"`
}

// TableName overrides the table name used by GORM
func (ProcessorState) TableName() string {
	return "processor_state"
}
