package state

import (
	"fmt"
	"sort"
	"github.com/quantumwake/alethic-ism-core-go/pkg/repository"
	"github.com/quantumwake/alethic-ism-core-go/pkg/utils"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
	"log"
)

type BackendStorage struct {
	*repository.Access
}

func NewBackend(dsn string) *BackendStorage {
	return &BackendStorage{
		Access: repository.NewDataAccess(dsn),
	}
}

// FindState methods for finding state data.
func (da *BackendStorage) FindState(id string) (*State, error) {
	var state State
	result := da.DB.Where("id = ?", id).First(&state)
	if result.Error != nil {
		return nil, result.Error
	}
	return &state, nil
}

// UpsertState inserts a state if it does not exist or updates the state if it does.
func UpsertState(db *gorm.DB, state *State) error {
	return db.Clauses(clause.OnConflict{
		Columns:   []clause.Column{{Name: "id"}},
		DoUpdates: clause.AssignmentColumns([]string{"state_type", "count"}),
	}).Create(state).Error
}

func (da *BackendStorage) UpsertState(state *State) error {
	return UpsertState(da.DB, state)
}

func (da *BackendStorage) UpsertStateComplete(state *State) error {
	return da.RunTransactionIsolation(func(db *gorm.DB) error {
		// persist the state in first
		if err := UpsertState(db, state); err != nil {
			return fmt.Errorf("unable to store state: %v", err)
		}

		//utils.MapReduce()

		/// TODO build util map function
		var attributes []*ConfigAttribute
		for _, attr := range state.Config.Attributes {
			attributes = append(attributes, &ConfigAttribute{
				StateID:   state.ID,
				Attribute: attr.Attribute,
				Data:      attr.Data,
			})
		}

		//if err = UpsertConfigAttributes(db, state.Config.Attributes); err != nil {
		//
		//}

		return nil
	})
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
func (da *BackendStorage) FindDataRowColumnDataByColumnID(id *int64) (*DataRowColumnData, error) {
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
	columnData := &DataRowColumnData{
		Values: values,
		Count:  len(values),
	}

	return columnData, nil
}

// FindDataColumnDefinitionsByStateID finds all DataColumnDefinitions for a given state ID.
func (da *BackendStorage) FindDataColumnDefinitionsByStateID(id string) (Columns, error) {
	var definitions []*DataColumnDefinition
	result := da.DB.Where("state_id = ?", id).Find(&definitions)
	if result.Error != nil {
		return nil, result.Error
	}

	// Create a map of column name to DataColumnDefinition
	definitionsMap := make(map[string]*DataColumnDefinition)
	for _, definition := range definitions {
		definitionsMap[definition.Name] = definition
	}

	return definitionsMap, nil
}

// FindStateFull finds a state and all associated data columns and data rows
func (da *BackendStorage) FindStateFull(id string, flags StateLoadFlags) (*State, error) {
	state, err := da.FindState(id)
	if err != nil {
		return nil, fmt.Errorf("failed to find state, error: %v", err)
	}

	// If we only need the basic state data, return it now
	if flags == StateLoadBasic {
		return state, nil
	}

	// Find the key definitions for the state and add them to the state
	var keyDefinitions TypedColumnKeyDefinitions
	if flags&StateLoadConfigKeyDefinitions != 0 {
		// Find the key definitions for the state and add them to the state
		keyDefinitions, err = da.FindStateConfigKeyDefinitionsGroupByDefinitionType(id)
		if err != nil {
			return nil, fmt.Errorf("failed to find state data, error: %v", err)
		}
	}

	// Find the config attributes for the state and add them to the state
	var configAttributes ConfigAttributes
	if flags&StateLoadConfigAttributes != 0 {
		// Find the key definitions for the state and add them to the state
		configAttributes, err = da.FindConfigAttributes(id)
		if err != nil {
			return nil, fmt.Errorf("failed to find state data, error: %v", err)
		}
	}

	// Derive state config
	state.Config = &Config{
		Attributes:          configAttributes, // TODO fetch the config attributes
		TypedKeyDefinitions: keyDefinitions,
	}

	// Find the data columns for the state and add them to the state
	var columns Columns
	if flags&StateLoadColumns != 0 {
		// Find the data columns for the state and add them to the state
		columns, err = da.FindDataColumnDefinitionsByStateID(id)
		if err != nil {
			return nil, fmt.Errorf("failed to find state data, error: %v", err)
		}
		state.Columns = columns
	}

	if flags&StateLoadData != 0 {
		// Find the data for each column and add it to the state data map
		state.Data = make(map[string]*DataRowColumnData)
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
	}

	return state, nil
}

// UpsertStateColumns insert a map of DataColumnDefinition if it does not exist or updates the DataColumnDefinition if it does.
func (da *BackendStorage) UpsertStateColumns(columns Columns) error {
	insertColumns := utils.MapValues(columns, func(column *DataColumnDefinition) *DataColumnDefinition {
		return column
	})

	// TODO figure this out, needs to be able to handle both create and updates to the name.

	return da.DB.Clauses(clause.OnConflict{
		Columns: []clause.Column{
			{Name: "name"},
			{Name: "state_id"},
		},
		//Where: clause.Where{
		//	Exprs: c.Gt{Column: "id", Value: 0},
		//},
		DoUpdates: clause.AssignmentColumns([]string{
			"name",
			"data_type",
			"required",
			"callable",
			"value",
		}),
	}).Create(insertColumns).Error
}

// DeleteStateColumns deletes all DataColumnDefinitions for a given state ID.
func (da *BackendStorage) DeleteStateColumns(stateID string) int {
	result := da.DB.Where("state_id = ?", stateID).Delete(&DataColumnDefinition{})
	return int(result.RowsAffected)
}

// DeleteStateColumn deletes a DataColumnDefinition by ID.
func (da *BackendStorage) DeleteStateColumn(id int64) bool {
	result := da.DB.Delete(&DataColumnDefinition{}, id)
	return result.RowsAffected > 0
}

// ChunkRow represents a single (column_name, data_index, value) tuple from the
// export query. Used internally by FetchDataChunk.
type ChunkRow struct {
	Name      string  `gorm:"column:name"`
	DataIndex int64   `gorm:"column:data_index"`
	DataValue *string `gorm:"column:data_value"`
}

// FetchDataChunk retrieves state data for the given index range [offset, offset+limit)
// and returns it as a slice of row maps, pivoted from the columnar storage.
// This mirrors alethic-ism-api's fetch_state_data_chunk_for_export.
func (da *BackendStorage) FetchDataChunk(stateID string, offset, limit int64) ([]map[string]any, error) {
	var rows []ChunkRow
	err := da.DB.Raw(`
		SELECT sc.name, sd.data_index,
		       CASE WHEN sc.data_type = 'json'
		            THEN sd.data_json_value::text
		            ELSE sd.data_value
		       END AS data_value
		FROM state_column sc
		LEFT JOIN state_column_data sd ON sc.id = sd.column_id
		WHERE sc.state_id = ?
		  AND sd.data_index >= ?
		  AND sd.data_index < ?
		ORDER BY sd.data_index, sc.name
	`, stateID, offset, offset+limit).Scan(&rows).Error

	if err != nil {
		return nil, fmt.Errorf("fetch data chunk: %w", err)
	}

	// Pivot: group by data_index into records.
	recordMap := make(map[int64]map[string]any)
	var indices []int64

	for _, r := range rows {
		rec, exists := recordMap[r.DataIndex]
		if !exists {
			rec = make(map[string]any)
			recordMap[r.DataIndex] = rec
			indices = append(indices, r.DataIndex)
		}
		if r.DataValue != nil {
			rec[r.Name] = *r.DataValue
		}
	}

	records := make([]map[string]any, 0, len(indices))
	for _, idx := range indices {
		records = append(records, recordMap[idx])
	}

	return records, nil
}

// streamWindowSize is the width of the data_index range fetched per query in
// StreamData. It bounds BOTH the per-query sort (Postgres pgsql_tmp spill) AND
// the result the client driver buffers, keeping StreamData's memory flat
// regardless of state size. A single full-table ORDER BY (the previous
// implementation) sorted the whole EAV result — carrying large data_value
// payloads — which spilled the DB's temp space and was accumulated client-side
// by the driver, OOMing the migrator.
const streamWindowSize = 1000

// StreamData streams every row of a state to fn in data_index order, pivoting
// the columnar (EAV) storage on the fly. It pages through the state in fixed
// data_index range windows of streamWindowSize, so neither the database nor the
// client ever materializes the whole state. Each window owns a contiguous,
// non-overlapping index range, so every logical row's columns land in exactly
// one window (never split), and gaps in data_index simply yield emptier
// windows. fn is invoked once per logical row, in ascending data_index order;
// returning an error aborts.
func (da *BackendStorage) StreamData(stateID string, fn func(record map[string]any) error) error {
	// One cheap bounds query to size the loop. NULL max → the state has no data.
	var bounds struct {
		MinIdx *int64 `gorm:"column:min_idx"`
		MaxIdx *int64 `gorm:"column:max_idx"`
	}
	if err := da.DB.Raw(`
		SELECT MIN(scd.data_index) AS min_idx, MAX(scd.data_index) AS max_idx
		FROM state_column sc
		JOIN state_column_data scd ON sc.id = scd.column_id
		WHERE sc.state_id = ? AND scd.data_index IS NOT NULL
	`, stateID).Scan(&bounds).Error; err != nil {
		return fmt.Errorf("stream data bounds: %w", err)
	}
	if bounds.MaxIdx == nil {
		return nil // no data
	}

	for start := *bounds.MinIdx; start <= *bounds.MaxIdx; start += streamWindowSize {
		end := start + streamWindowSize // exclusive upper bound

		// One bounded window: all cells whose data_index is in [start, end),
		// ordered row-major (data_index first) so the pivot can emit on the index
		// boundary. The sort and the returned rows are bounded to this window.
		var batch []ChunkRow
		// Column-major order (column_id, data_index) matches the
		// (column_id, data_index) index, so this is an index range-scan with NO
		// server-side sort — the fast path. The data_index range predicate bounds
		// it to one window.
		if err := da.DB.Raw(`
			SELECT sc.name, scd.data_index,
			       CASE WHEN sc.data_type = 'json'
			            THEN scd.data_json_value::text
			            ELSE scd.data_value
			       END AS data_value
			FROM state_column sc
			JOIN state_column_data scd ON sc.id = scd.column_id
			WHERE sc.state_id = ?
			  AND scd.data_index >= ? AND scd.data_index < ?
			ORDER BY scd.column_id, scd.data_index
		`, stateID, start, end).Scan(&batch).Error; err != nil {
			return fmt.Errorf("stream data window [%d,%d): %w", start, end, err)
		}

		// Pivot in memory. The window is bounded (streamWindowSize indices), so
		// grouping its cells by data_index and sorting the keys is cheap — and it
		// frees the DB from a data_index-major sort. A record is created for any
		// data_index that has cells (even if all values are NULL); absent columns
		// simply stay out of the map (sparse rows + sparse columns both handled).
		recs := make(map[int64]map[string]any, streamWindowSize)
		for i := range batch {
			r := batch[i]
			rec := recs[r.DataIndex]
			if rec == nil {
				rec = make(map[string]any)
				recs[r.DataIndex] = rec
			}
			if r.DataValue != nil {
				rec[r.Name] = *r.DataValue
			}
		}

		// Emit logical rows in ascending data_index order.
		indices := make([]int64, 0, len(recs))
		for idx := range recs {
			indices = append(indices, idx)
		}
		sort.Slice(indices, func(i, j int) bool { return indices[i] < indices[j] })
		for _, idx := range indices {
			if err := fn(recs[idx]); err != nil {
				return err
			}
		}
		// recs and batch are eligible for GC before the next window's query.
	}
	return nil
}

// ListStates returns all state IDs with their name and row count.
func (da *BackendStorage) ListStates() ([]State, error) {
	var states []State
	result := da.DB.Find(&states)
	if result.Error != nil {
		return nil, result.Error
	}
	return states, nil
}

// FindStatesByProjectID returns all states belonging to a project (basic rows;
// no columns/data). Mirrors processor.FindProcessorByProjectID.
func (da *BackendStorage) FindStatesByProjectID(projectID string) ([]State, error) {
	var states []State
	result := da.DB.Where("project_id = ?", projectID).Find(&states)
	if result.Error != nil {
		return nil, result.Error
	}
	return states, nil
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
