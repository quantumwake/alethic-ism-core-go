package state

import (
	"fmt"
	"github.com/quantumwake/alethic-ism-core-go/pkg/data"
	"github.com/quantumwake/alethic-ism-core-go/pkg/data/models/state"
	"gorm.io/gorm/clause"
	"log"
)

type BackendStorage struct {
	*data.Access
}

func NewBackend(dsn string) *BackendStorage {
	return &BackendStorage{
		Access: data.NewDataAccess(dsn),
	}
}

// FindState methods for finding state data.
func (da *BackendStorage) FindState(id string) (*state.State, error) {
	var state state.State
	result := da.DB.Where("id = ?", id).First(&state)
	if result.Error != nil {
		return nil, result.Error
	}
	return &state, nil
}

// UpsertState inserts a state if it does not exist or updates the state if it does.
func (da *BackendStorage) UpsertState(state *state.State) error {
	return da.DB.Clauses(clause.OnConflict{
		Columns:   []clause.Column{{Name: "id"}},
		DoUpdates: clause.AssignmentColumns([]string{"state_type", "count"}),
	}).Create(state).Error
}

// FindDataRowColumnDataByColumnID finds DataRowColumnData by column ID.
//func (da *BackendStorage) FindDataRowColumnDataByColumnID(id int64) ([]*models.DataRowColumnData, error) {
//	var columnData []models.DataRowColumnData
//	result := da.DB.Where("column_id = ?", id).First(&columnData)
//	if result.Error != nil {
//		return nil, result.Error
//	}
//
//	return &columnData, nil
//}

// FindDataRowColumnDataByColumnID retrieves all values for a column ID in order by index.
func (da *BackendStorage) FindDataRowColumnDataByColumnID(id int64) (*state.DataRowColumnData, error) {
	var values []string

	// Query the column_value directly, ordered by column_index
	result := da.DB.Table("state_column_data").
		Select("data_value").
		Where("column_id = ?", id).
		Order("data_index ASC").
		Pluck("data_value", &values)

	if result.Error != nil {
		return nil, result.Error
	}

	// Create the DataRowColumnData with the ordered values
	columnData := &state.DataRowColumnData{
		Values: values,
		Count:  len(values),
	}

	return columnData, nil
}

// FindDataColumnDefinitionsByStateID finds all DataColumnDefinitions for a given state ID.
func (da *BackendStorage) FindDataColumnDefinitionsByStateID(id string) (map[string]*state.DataColumnDefinition, error) {
	var definitions []*state.DataColumnDefinition
	result := da.DB.Where("state_id = ?", id).Find(&definitions)
	if result.Error != nil {
		return nil, result.Error
	}

	// Create a map of column name to DataColumnDefinition
	definitionsMap := make(map[string]*state.DataColumnDefinition)
	for _, definition := range definitions {
		definitionsMap[definition.Name] = definition
	}

	return definitionsMap, nil
}

// FindStateFull finds a state and all associated data columns and data rows
func (da *BackendStorage) FindStateFull(id string) (*state.State, error) {
	state, err := da.FindState(id)
	if err != nil {
		return nil, fmt.Errorf("failed to find state, error: %v", err)
	}

	// Find the data columns for the state and add them to the state
	columns, err := da.FindDataColumnDefinitionsByStateID(id)
	if err != nil {
		return nil, fmt.Errorf("failed to find state data, error: %v", err)
	}
	state.Columns = columns

	// Find the data for each column and add it to the state data map
	state.Data = make(map[string]*state.DataRowColumnData)
	for _, column := range columns {
		columnData, err := da.FindDataRowColumnDataByColumnID(column.ID)
		if err != nil {
			return nil, fmt.Errorf("failed to find state data, error: %v", err)
		}
		state.Data[column.Name] = columnData
		if state.Count != columnData.Count {
			// TODO log warning, generally this should not happen, unless we have a serious persistent issue, which can in theory happen (check python db code, we need a new storage solution, this needs to be done in a single transaction maybe?)
			log.Printf("state count %v does not match column data count %v, column data needs to be rebalanced or cut out from maximum position index", state.Count, columnData.Count)
		}
	}

	return state, nil
}

//// UnmarshalJSON is a custom unmarshaler for the Usage struct to handle the transaction time field.
//func (u *BackendStorage) UnmarshalJSON(data []byte) error {
//
//	// Define an alias struct to handle the transaction time field.
//	type Alias models.State
//
//	// Define an auxiliary struct to handle the transaction time field.
//	aux := &struct {
//		TransactionTime string `json:"transaction_time"`
//		*Alias
//	}{
//		Alias: (*Alias)(u),
//	}
//
//	// Unmarshal the data into the auxiliary struct.
//	if err := json.Unmarshal(data, &aux); err != nil {
//		return err
//	}
//	var err error
//
//	// Parse the transaction time field into the Usage struct.
//	u.TransactionTime, err = time.Parse("2006-01-02T15:04:05.999999", aux.TransactionTime)
//	if err != nil {
//		return err
//	}
//	return nil
//}
