package data

import (
	"fmt"
	"github.com/quantumwake/alethic-ism-core-go/pkg/model"
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
func (da *Access) FindRouteByID(id string) (*model.ProcessorState, error) {
	var processorState model.ProcessorState
	result := da.DB.Where("id = ?", id).First(&processorState)
	if result.Error != nil {
		return nil, result.Error
	}
	return &processorState, nil
}

// FindByProcessorIDAndDirection finds all ProcessorStates for a given processor ID and direction
func (da *Access) FindRouteByProcessorAndDirection(processorID string, direction model.ProcessorStateDirection) ([]model.ProcessorState, error) {
	var processorStates []model.ProcessorState

	result := da.DB.
		Where("processor_id = ? AND direction = ?", processorID, string(direction)).
		Find(&processorStates)

	return processorStates, result.Error
}

func (da *Access) FindRouteByStateAndDirection(stateID string, direction model.ProcessorStateDirection) ([]model.ProcessorState, error) {
	var processorStates []model.ProcessorState

	result := da.DB.
		Where("state_id = ? AND direction = ?", stateID, string(direction)).
		Find(&processorStates)

	return processorStates, result.Error
}

// Usage methods
func (da *Access) InsertUsage(usage *model.Usage) error {
	db := da.DB.Create(usage)

	if db.Error != nil {
		return fmt.Errorf("failed to insert usage data, error: %v", db.Error)
	}

	return nil
}

func (da *Access) InsertTrace(trace *model.Trace) error {
	db := da.DB.Create(trace)
	if db.Error != nil {
		return fmt.Errorf("failed to insert trace data, error: %v", db.Error)
	}

	return nil
}

func (da *Access) FindTraceAllByPartition(partition string) ([]model.Trace, error) {
	var traces []model.Trace

	result := da.DB.
		Where("partition = ?", partition).
		Find(&traces)

	return traces, result.Error
}

func (da *Access) FindTraceAllByPartitionAndReference(partition, reference string) ([]model.Trace, error) {
	var traces []model.Trace

	result := da.DB.
		Where("partition = ? AND reference = ?", partition, reference).
		Find(&traces)

	return traces, result.Error
}
