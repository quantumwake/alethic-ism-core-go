package state

import (
	"gorm.io/gorm/clause"
)

// FindStateConfigKeyDefinitions finds all data key definitions for a given state id.
func (da *BackendStorage) FindStateConfigKeyDefinitions(stateID string) (ColumnKeyDefinitions, error) {
	var definitions []*ColumnKeyDefinition
	result := da.DB.Where("state_id = ?", stateID).Find(&definitions)
	if result.Error != nil {
		return nil, result.Error
	}
	return definitions, nil
}

// FindStateConfigKeyDefinitionsGroupByDefinitionType finds all data key definitions for a given state id and groups them by definition type.
func (da *BackendStorage) FindStateConfigKeyDefinitionsGroupByDefinitionType(stateID string) (TypedColumnKeyDefinitions, error) {
	definitions, err := da.FindStateConfigKeyDefinitions(stateID)
	if err != nil {
		return nil, err
	}

	typedColumnKeyDefinitions := TypedColumnKeyDefinitions{}
	for _, definition := range definitions {
		//if columnKeyDefinitions, ok := typedColumnKeyDefinitions[definition.DefinitionType]; !ok {
		//columnKeyDefinitions = ColumnKeyDefinitions{definition}
		//typedColumnKeyDefinitions[definition.DefinitionType] = append(typedColumnKeyDefinitions[definition.DefinitionType], definition)
		//}

		typedColumnKeyDefinitions[definition.DefinitionType] = append(typedColumnKeyDefinitions[definition.DefinitionType], definition)
	}
	return typedColumnKeyDefinitions, nil

	//// Group the definitions using the new utility function
	//return utils.SliceToGroupMap(definitions, func(def *ColumnKeyDefinition) DefinitionType {
	//	return def.DefinitionType
	//}), nil
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
