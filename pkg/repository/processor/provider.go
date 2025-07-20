package processor

import (
	"github.com/quantumwake/alethic-ism-core-go/pkg/data"
	"time"
)

type Class string

const (
	Code            Class = "CodeProcessing"
	NLP             Class = "NaturalLanguageProcessing"
	Data            Class = "DataProcessing"
	Text            Class = "TextProcessing"
	Image           Class = "ImageProcessing"
	Video           Class = "VideoProcessing"
	Audio           Class = "AudioProcessing"
	Signal          Class = "SignalProcessing"
	Database        Class = "DatabaseProcessing"
	MachineLearning Class = "MachineLearning"
	Interactive     Class = "Interactive"
	Proprietary     Class = "Proprietary"
)

type ProviderClass struct {
	ClassName string `json:"class_name" gorm:"column:class_name;type:varchar(255);not null"`
}

func (ProviderClass) TableName() string {
	return "processor_class"
}

type Provider struct {
	ID      string `json:"id" gorm:"column:id;type:varchar(36);primaryKey;not null"`
	Name    string `json:"name" gorm:"column:name;type:varchar(255);not null"`
	Version string `json:"version" gorm:"column:version;type:varchar(32);null"` // Allow null string

	// Class many-to-one
	ClassName Class         `json:"class_name" gorm:"column:class_name;type:string;not null"`
	Class     ProviderClass `json:"-" gorm:"foreignKey:ClassName;references:ClassName"`
	UserID    *string       `json:"user_id" gorm:"column:user_id;type:varchar(36);null"`       // Allow null string
	ProjectID *string       `json:"project_id" gorm:"column:project_id;type:varchar(36);null"` // Allow null string

	CreatedAt time.Time `json:"created_at" gorm:"column:created_date;not null; autoCreateTime"`
	UpdatedAt time.Time `json:"updated_at" gorm:"column:updated_date;not null; autoUpdateTime"`
	Routing   data.JSON `json:"routing" gorm:"column:routing;null"` // Use JSONB for PostgreSQL or JSON for MySQL
}

func (Provider) TableName() string {
	return "processor_provider"
}
