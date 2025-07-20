package processor

import (
	"github.com/quantumwake/alethic-ism-core-go/pkg/data"
	"time"
)

type Processor struct {
	ID         string     `json:"id"`
	Name       string     `json:"name" gorm:"column:name;type:varchar(255);not null"`
	ProviderID *string    `json:"provider_id" gorm:"column:provider_id;type:varchar(36);null"`
	ProjectID  string     `json:"project_id" gorm:"column:project_id;type:varchar(36);not null"`
	Properties *data.JSON `json:"properties" gorm:"column:properties;type:jsonb;null"`
	CreatedAt  time.Time  `json:"created_at" gorm:"column:created_date; not null;autoCreateTime"`
	UpdatedAt  time.Time  `json:"updated_at" gorm:"column:updated_date; not null;autoUpdateTime"`
	Status     Status     `json:"status" gorm:"column:status;type:varchar(32);not null;default:'CREATED'"`
}

func (Processor) TableName() string {
	return "processor"
}
