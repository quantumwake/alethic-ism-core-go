package query

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

func (data StateQueryResults) Pivot() []map[string]any {
	curIdx := 1
	var current map[string]any = nil
	var results = make([]map[string]any, 0)

	for _, cell := range data {
		if cell.DataIndex != curIdx {
			if current != nil {
				results = append(results, current)
			}
			current = map[string]any{}
			curIdx = cell.DataIndex
		}
		current[cell.ColumnName] = cell.DataValue
	}

	return results
}

type StateQueryResults []dsl.StateQueryResult

func (da *BackendStorage) Query(stateID string, query dsl.StateQuery) (StateQueryResults, error) {
	// Validate UUID
	if err := utils.ValidateUUID(stateID); err != nil {
		return nil, fmt.Errorf("invalid UUID: %v", err)
	}

	// Build the final SQL query and arguments
	dataSQL, dataArgs, err := query.BuildFinalQuery(stateID)
	if err != nil {
		return nil, fmt.Errorf("failed to build final query: %v", err)
	}

	// Execute the final query to get the results
	var results []dsl.StateQueryResult
	if err = da.Access.Query(dataSQL, &results, dataArgs...); err != nil {
		return nil, fmt.Errorf("failed to fetch data values: %v", err)
	}

	//if err := da.Access.DB.Raw(dataSQL, dataArgs...).Scan(&results).Error; err != nil {
	//}

	return results, nil
}
