package route

import (
	"github.com/quantumwake/alethic-ism-core-go/pkg/data"
	"github.com/quantumwake/alethic-ism-core-go/pkg/model"
)

type BackendStorage struct {
	*data.Access
}

func NewBackend(dsn string) *BackendStorage {
	return &BackendStorage{
		Access: data.NewDataAccess(dsn),
	}
}

// FindRouteByID methods
func (da *BackendStorage) FindRouteByID(id string) (*model.ProcessorState, error) {
	var processorState model.ProcessorState
	result := da.DB.Where("id = ?", id).First(&processorState)
	if result.Error != nil {
		return nil, result.Error
	}
	return &processorState, nil
}

// FindRouteByProcessorAndDirection finds all ProcessorStates for a given processor ID and direction
func (da *BackendStorage) FindRouteByProcessorAndDirection(processorID string, direction model.ProcessorStateDirection) ([]model.ProcessorState, error) {
	var processorStates []model.ProcessorState

	result := da.DB.
		Where("processor_id = ? AND direction = ?", processorID, string(direction)).
		Find(&processorStates)

	return processorStates, result.Error
}

// FindRouteByStateAndDirection find routes by state id and the direction it is flowing.
func (da *BackendStorage) FindRouteByStateAndDirection(stateID string, direction model.ProcessorStateDirection) ([]model.ProcessorState, error) {
	var processorStates []model.ProcessorState

	result := da.DB.
		Where("state_id = ? AND direction = ?", stateID, string(direction)).
		Find(&processorStates)

	return processorStates, result.Error
}
