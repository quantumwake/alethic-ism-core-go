package state

import (
	"database/sql"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
	"log"
)

func (da *BackendStorage) RunTransactionIsolation(fn func(db *gorm.DB) error) error {
	tx := da.DB.Begin(&sql.TxOptions{Isolation: sql.LevelDefault})
	defer tx.Commit()
	return fn(tx)
}

// UpsertConfigAttribute inserts or updates a state config attribute.
func UpsertConfigAttribute(db *gorm.DB, attribute *ConfigAttribute) error {
	return db.Clauses(clause.OnConflict{
		Columns:   []clause.Column{{Name: "state_id"}, {Name: "attribute"}},
		DoUpdates: clause.AssignmentColumns([]string{"data"}),
	}).Create(&attribute).Error
}

// UpsertConfigAttribute inserts or updates a state config attribute.
func (da *BackendStorage) UpsertConfigAttribute(attribute *ConfigAttribute) error {
	return UpsertConfigAttribute(da.DB, attribute)
}

// UpsertConfigAttributes insert or update, if exists, state.config.attributes, by attribute key and state id.
func UpsertConfigAttributes(db *gorm.DB, attributes ConfigAttributes) error {
	if len(attributes) == 0 {
		// TODO trace logging
		return nil
	}

	return db.Clauses(clause.OnConflict{
		Columns:   []clause.Column{{Name: "state_id"}, {Name: "attribute"}},
		DoUpdates: clause.AssignmentColumns([]string{"data"}),
	}).Create(attributes).Error
}

// UpsertConfigAttributes insert or update, if exists, state.config.attributes, by attribute key and state id.
func (da *BackendStorage) UpsertConfigAttributes(attributes ConfigAttributes) error {
	return UpsertConfigAttributes(da.DB, attributes)
}

// FindConfigAttributes retrieves configuration entries by state_id.
func (da *BackendStorage) FindConfigAttributes(stateID string) (ConfigAttributes, error) {
	var configs []*ConfigAttribute
	if err := da.DB.Where("state_id = ?", stateID).Find(&configs).Error; err != nil {
		log.Printf("error fetching configs for state_id: %s, error: %v", stateID, err)
		return nil, err
	}
	return configs, nil
}

// DeleteConfigAttributes deletes configuration entries by state_id.
func (da *BackendStorage) DeleteConfigAttributes(stateID string) error {
	return da.DB.Where("state_id = ?", stateID).Delete(&ConfigAttribute{}).Error
}
