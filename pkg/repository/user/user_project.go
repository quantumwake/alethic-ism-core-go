package user

import "time"

type Project struct {
	ID          string    `gorm:"column:project_id;type:varchar(36);primaryKey;not null" json:"project_id"`
	Name        string    `gorm:"column:project_name;type:varchar(36);not null" json:"project_name"` // You may define more specific types here
	UserID      string    `gorm:"column:user_id;type:varchar(36);not null" json:"user_id"`
	CreatedDate time.Time `gorm:"column:created_date;type:timestamp;not null;autoCreateTime" json:"created_date"`
}

// TableName sets the table name for the User Project struct
func (Project) TableName() string {
	return "user_project"
}
