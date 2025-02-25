package trace

import "time"

type LogLevel string

const (
	LogLevelInfo  LogLevel = "INFO"
	LogLevelWarn  LogLevel = "WARN"
	LogLevelError LogLevel = "ERROR"
	LogLevelDebug LogLevel = "DEBUG"
)

type Trace struct {
	ID int64 `gorm:"not null; column:id; primaryKey" json:"id"`

	// Action the logger is performing
	Action string `gorm:"not null; column:action" json:"action"`

	// ActionTime the action was performed at
	ActionTime time.Time `gorm:"not null; column:action_time" json:"action_time"`

	// Partition the logger is operating in (e.g. a project id, a user id, etc)
	Partition string `gorm:"not null; column:partition; index" json:"partition"`

	// Reference to the entity the logger is operating on (e.g. a state id, a processor id, etc.)
	Reference string `gorm:"not null; column:reference; index" json:"reference"`

	// Level of the log message
	Level LogLevel `gorm:"not null; column:level" json:"level"`

	// Message to log
	Message string `gorm:"not null; column:message" json:"message"`
}

func (Trace) TableName() string {
	return "trace"
}
