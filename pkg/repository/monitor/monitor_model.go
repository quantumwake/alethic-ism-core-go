package monitor

import "time"

// ProcessorStateRecord represents the database model for processor state
type ProcessorStateRecord struct {
	ID        uint   `gorm:"primaryKey"`
	Type      string `gorm:"not null"`
	RouteID   string `gorm:"not null"`
	State     string `gorm:"not null"`
	CreatedAt time.Time
	UpdatedAt time.Time
}

// MonitorLogRecord represents the database model for monitor logs
type MonitorLogRecord struct {
	ID          uint   `gorm:"primaryKey"`
	ProcessorID uint   `gorm:"not null"`
	Type        string `gorm:"not null"`
	RouteID     string `gorm:"not null"`
	Status      string `gorm:"not null"`
	Exception   string
	Data        string `gorm:"type:jsonb"`
	CreatedAt   time.Time
}
