package state

import (
	"gorm.io/gorm/clause"
	"log"
)

// UpsertConfigAttribute inserts or updates a state config attribute.
func (da *BackendStorage) UpsertConfigAttribute(attribute *StateConfigAttribute) error {
	return da.DB.Clauses(clause.OnConflict{
		Columns:   []clause.Column{{Name: "state_id"}, {Name: "attribute"}},
		DoUpdates: clause.AssignmentColumns([]string{"data"}),
	}).Create(&attribute).Error
}

// UpsertConfigAttributes insert or update, if exists, state.config.attributes, by attribute key and state id.
func (da *BackendStorage) UpsertConfigAttributes(attributes []*StateConfigAttribute) error {
	if len(attributes) == 0 {
		// TODO trace logging
		return nil
	}

	return da.DB.Clauses(clause.OnConflict{
		Columns:   []clause.Column{{Name: "state_id"}, {Name: "attribute"}},
		DoUpdates: clause.AssignmentColumns([]string{"data"}),
	}).Create(attributes).Error
}

// FindConfigAttributes retrieves configuration entries by state_id.
func (da *BackendStorage) FindConfigAttributes(stateID string) ([]*StateConfigAttribute, error) {
	var configs []*StateConfigAttribute
	if err := da.DB.Where("state_id = ?", stateID).Find(&configs).Error; err != nil {
		log.Printf("error fetching configs for state_id: %s, error: %v", stateID, err)
		return nil, err
	}
	return configs, nil
}

// DeleteConfigAttributes deletes configuration entries by state_id.
func (da *BackendStorage) DeleteConfigAttributes(stateID string) error {
	return da.DB.Where("state_id = ?", stateID).Delete(&StateConfigAttribute{}).Error
}
