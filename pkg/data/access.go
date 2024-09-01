package data

import (
	"fmt"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
	"log"
)

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

// ProcessorState methods
func (da *Access) FindRouteByID(id string) (*ProcessorState, error) {
	var processorState ProcessorState
	result := da.DB.Where("id = ?", id).First(&processorState)
	if result.Error != nil {
		return nil, result.Error
	}
	return &processorState, nil
}

// FindByProcessorIDAndDirection finds all ProcessorStates for a given processor ID and direction
func (da *Access) FindRouteByProcessorAndDirection(processorID string, direction ProcessorStateDirection) ([]ProcessorState, error) {
	var processorStates []ProcessorState

	result := da.DB.
		Where("processor_id = ? AND direction = ?", processorID, string(direction)).
		Find(&processorStates)

	return processorStates, result.Error
}

// Usage methods
func (da *Access) InsertUsage(usage *Usage) error {
	db := da.DB.Create(usage)

	if db.Error != nil {
		return fmt.Errorf("failed to insert usage data, error: %v", db.Error)
	}

	return nil
}
