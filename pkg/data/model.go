package data

import (
	"encoding/json"
	"fmt"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
	"log"
	"time"
)

type UnitType string
type UnitSubType string

const (
	UNIT_TOKENS  UnitType = "TOKEN"
	UNIT_COMPUTE          = "COMPUTE"
)

const (
	UNIT_INPUT  UnitSubType = "INPUT"
	UNIT_OUTPUT             = "OUTPUT"
)

type Usage struct {
	ID              int       `gorm:"not null; column:id; primaryKey" json:"id"`
	TransactionTime time.Time `gorm:"not null; column:transaction_time" json:"transaction_time"`

	// Project Reference
	ProjectID string `gorm:"not null; column:project_id" json:"project_id"`

	// unique resource information such that we can reference it in billing (TODO resource should probably only be logically deleted, e.g. a datasource id or a processor id or compute node id, or something else)
	ResourceID   string `gorm:"not null; column:resource_id; size:255" json:"resource_id"`
	ResourceType string `gorm:"not null; column:resource_type; size:255" json:"resource_type"`

	UnitType    UnitType    `gorm:"not null; column:unit_type" json:"unit_type"`
	UnitSubType UnitSubType `gorm:"not null; column:unit_subtype" json:"unit_subtype"`
	UnitCount   int         `gorm:"not null; column:unit_count" json:"unit_count"`

	Metadata json.RawMessage `gorm:"null; column:metadata" json:"metadata"`
}

// TableName sets the table name for the Usage struct
func (Usage) TableName() string {
	return "usage"
}

type Access struct {
	DSN string
	DB  *gorm.DB
}

func NewDataAccess(dsn string) *Access {
	da := &Access{
		DSN: dsn,
	}
	err := da.Connect()
	if err != nil {
		log.Fatalf("unable to connect to database: %v", err)
	}
	return da
}

func (da *Access) Connect() error {
	var err error
	da.DB, err = gorm.Open(postgres.Open(da.DSN), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Info),
	})
	if err != nil {
		return err
	}
	return nil
}

func (da *Access) Close() error {
	//	TODO TBD
	return nil
}

func (da *Access) InsertUsage(usage *Usage) error {
	db := da.DB.Create(usage)

	if db.Error != nil {
		return fmt.Errorf("failed to insert usage data, error: %v", db.Error)
	}

	return nil
}
