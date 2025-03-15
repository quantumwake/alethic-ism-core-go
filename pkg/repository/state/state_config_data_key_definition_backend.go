package state

import (
	"gorm.io/gorm/clause"
)

// FindStateConfigKeyDefinitions finds all data key definitions for a given state id.
func (da *BackendStorage) FindStateConfigKeyDefinitions(stateID string) ([]*ColumnKeyDefinition, error) {
	var definitions []*ColumnKeyDefinition
	result := da.DB.Where("state_id = ?", stateID).Find(&definitions)
	if result.Error != nil {
		return nil, result.Error
	}
	return definitions, nil
}

// UpsertStateConfigKeyDefinitions inserts a new or updates an existing data key definition list for a given state.
func (da *BackendStorage) UpsertStateConfigKeyDefinitions(definitions []*ColumnKeyDefinition) error {
	// TODO might be a security risk due to id injection... check it over.
	return da.DB.Clauses(clause.OnConflict{
		Columns: []clause.Column{
			{Name: "name"},
			{Name: "state_id"},
			{Name: "definition_type"},
		},
		DoUpdates: clause.AssignmentColumns([]string{
			"name", "alias", "callable", "required", "definition_type",
		}),
	}).Create(definitions).Error
}
