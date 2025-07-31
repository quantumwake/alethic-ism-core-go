package route

import (
	"github.com/quantumwake/alethic-ism-core-go/pkg/repository"
	"github.com/quantumwake/alethic-ism-core-go/pkg/repository/processor"
)

type BackendStorage struct {
	*repository.Access
}

func NewBackend(dsn string) *BackendStorage {
	return &BackendStorage{
		Access: repository.NewDataAccess(dsn),
	}
}

// FindRouteByID methods
func (da *BackendStorage) FindRouteByID(id string) (*processor.State, error) {
	var processorState processor.State
	result := da.DB.Where("id = ?", id).First(&processorState)
	if result.Error != nil {
		return nil, result.Error
	}
	return &processorState, nil
}

// FindRouteByProcessorAndDirection finds all ProcessorStates for a given processor ID and direction
func (da *BackendStorage) FindRouteByProcessorAndDirection(processorID string, direction processor.StateDirection) ([]processor.State, error) {
	var processorStates []processor.State

	result := da.DB.
		Where("processor_id = ? AND direction = ?", processorID, string(direction)).
		Find(&processorStates)

	return processorStates, result.Error
}

// FindRouteByStateAndDirection find routes by state id and the direction it is flowing.
func (da *BackendStorage) FindRouteByStateAndDirection(stateID string, direction processor.StateDirection) ([]processor.State, error) {
	var processorStates []processor.State

	result := da.DB.
		Where("state_id = ? AND direction = ?", stateID, string(direction)).
		Find(&processorStates)

	return processorStates, result.Error
}

func (da *BackendStorage) FindRouteByState(stateID string) ([]processor.State, error) {
	var processorStates []processor.State
	result := da.DB.
		Where("state_id = ?", stateID).
		Find(&processorStates)
	return processorStates, result.Error
}
