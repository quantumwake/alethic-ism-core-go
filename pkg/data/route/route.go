package route

import (
	"github.com/quantumwake/alethic-ism-core-go/pkg/data"
	"github.com/quantumwake/alethic-ism-core-go/pkg/data/processor_state"
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
func (da *BackendStorage) FindRouteByID(id string) (*processor_state.ProcessorState, error) {
	var processorState processor_state.ProcessorState
	result := da.DB.Where("id = ?", id).First(&processorState)
	if result.Error != nil {
		return nil, result.Error
	}
	return &processorState, nil
}

// FindRouteByProcessorAndDirection finds all ProcessorStates for a given processor ID and direction
func (da *BackendStorage) FindRouteByProcessorAndDirection(processorID string, direction processor_state.ProcessorStateDirection) ([]processor_state.ProcessorState, error) {
	var processorStates []processor_state.ProcessorState

	result := da.DB.
		Where("processor_id = ? AND direction = ?", processorID, string(direction)).
		Find(&processorStates)

	return processorStates, result.Error
}

// FindRouteByStateAndDirection find routes by state id and the direction it is flowing.
func (da *BackendStorage) FindRouteByStateAndDirection(stateID string, direction processor_state.ProcessorStateDirection) ([]processor_state.ProcessorState, error) {
	var processorStates []processor_state.ProcessorState

	result := da.DB.
		Where("state_id = ? AND direction = ?", stateID, string(direction)).
		Find(&processorStates)

	return processorStates, result.Error
}
