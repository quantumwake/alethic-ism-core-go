package state_query

import (
	"fmt"
	"github.com/quantumwake/alethic-ism-core-go/pkg/repository"
	"github.com/quantumwake/alethic-ism-core-go/pkg/repository/query/dsl"
	"github.com/quantumwake/alethic-ism-core-go/pkg/utils"
)

type Storage interface {
	Query(query dsl.StateQuery) ([]dsl.StateQueryResult, error)
}

type BackendStorage struct {
	Storage
	Access *repository.Access
}

func NewBackend(dsn string) *BackendStorage {
	storage := &BackendStorage{
		Access: repository.NewDataAccess(dsn),
	}
	return storage
}

func (da *BackendStorage) Query(stateID string, query dsl.StateQuery) ([]dsl.StateQueryResult, error) {
	// Validate UUID
	if err := utils.ValidateUUID(stateID); err != nil {
		return nil, fmt.Errorf("invalid UUID: %v", err)
	}

	// Build the final SQL query and arguments
	dataSQL, dataArgs, err := query.BuildFinalQuery()
	if err != nil {
		return nil, fmt.Errorf("failed to build final query: %v", err)
	}

	// Execute the final query to get the results
	var results []dsl.StateQueryResult
	if err = da.Access.Query(dataSQL, results, dataArgs); err != nil {
		return nil, fmt.Errorf("failed to fetch data values: %v", err)
	}

	//if err := da.Access.DB.Raw(dataSQL, dataArgs...).Scan(&results).Error; err != nil {
	//}

	return results, nil
}
