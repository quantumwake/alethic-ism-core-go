package data

import (
	"fmt"
	"github.com/quantumwake/alethic-ism-core-go/pkg/data"
	"github.com/quantumwake/alethic-ism-core-go/pkg/data/query/dsl"
	"github.com/quantumwake/alethic-ism-core-go/pkg/utils"
)

type Storage interface {
	Query(query dsl.StateQuery) ([]dsl.StateQueryResult, error)
}

type DatabaseStorage struct {
	Storage
	Access *data.Access
}

func (da *DatabaseStorage) Query(query dsl.StateQuery) ([]dsl.StateQueryResult, error) {
	// Validate UUID
	if err := utils.ValidateUUID(query.StateID); err != nil {
		return nil, fmt.Errorf("invalid UUID: %v", err)
	}

	// Build the final SQL query and arguments
	dataSQL, dataArgs, err := query.BuildFinalQuery()
	if err != nil {
		return nil, fmt.Errorf("failed to build final query: %v", err)
	}

	// Execute the final query to get the results
	var results []dsl.StateQueryResult
	if err := da.Access.DB.Raw(dataSQL, dataArgs...).Scan(&results).Error; err != nil {
		return nil, fmt.Errorf("failed to fetch data values: %v", err)
	}

	return results, nil
}
