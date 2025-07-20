package project

import (
	"github.com/quantumwake/alethic-ism-core-go/pkg/data"
	"time"
)

type Project struct {
	ID         string    `gorm:"column:project_id;type:varchar(36);primaryKey;not null" json:"project_id"`
	Name       string    `gorm:"column:project_name;type:varchar(36);not null" json:"project_name"` // You may define more specific types here
	UserID     string    `gorm:"column:user_id;type:varchar(36);not null" json:"user_id"`
	Properties data.JSON `gorm:"column:properties;type:jsonb;null" json:"properties"` // Use JSONB for PostgreSQL or JSON for MySQL
	CreatedAt  time.Time `gorm:"column:created_date;type:timestamp;not null;autoCreateTime" json:"created_date"`
	UpdatedAt  time.Time `gorm:"column:updated_date;type:timestamp;not null;autoUpdateTime" json:"updated_date"`
	DeletedAt  time.Time `gorm:"column:deleted_date;type:timestamp;    null" json:"deleted_date"`
}

// TableName sets the table name for the User Project struct
func (Project) TableName() string {
	return "user_project"
}
