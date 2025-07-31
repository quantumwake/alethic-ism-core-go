package monitor

import "time"

// LogEvent represents the model for monitor log events
type LogEvent struct {
	ID                  uint      `gorm:"primaryKey;autoIncrement"`
	LogType             string    `gorm:"column:log_type;not null"`
	InternalReferenceID uint64    `gorm:"column:internal_reference_id"`
	UserID              string    `gorm:"column:user_id;type:varchar(36)"`
	ProjectID           string    `gorm:"column:project_id;type:varchar(36)"`
	Data                string    `gorm:"column:data;type:text"`
	Exception           string    `gorm:"column:exception;type:text"`
	CreatedAt           time.Time `gorm:"column:created_at"`
}

func (LogEvent) TableName() string {
	return "monitor_log_event"
}
