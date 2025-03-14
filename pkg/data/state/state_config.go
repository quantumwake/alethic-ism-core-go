package state

import (
	"github.com/quantumwake/alethic-ism-core-go/pkg/data/models"
	"gorm.io/gorm/clause"
	"log"
)

//func (da *BackendStorage) FindConfigByStateID(id string) (*models.StateConfigVO, error) {
//var config models.StateConfigVO
//result := da.DB.Where("state_id = ?", id).First(&config)
//if result.Error != nil {
//	return nil, result.Error
//}
//return &config, nil

//return nil, nil
//}

// InsertAttribute inserts or updates a configuration entry.
func (da *BackendStorage) InsertAttribute(attribute *models.StateConfigAttribute) error {
	return da.DB.Clauses(clause.OnConflict{
		Columns:   []clause.Column{{Name: "state_id"}, {Name: "attribute"}},
		DoUpdates: clause.AssignmentColumns([]string{"data"}),
	}).Create(&attribute).Error
}

func (da *BackendStorage) UpsertConfigAttributes(attributes []*models.StateConfigAttribute) error {
	if len(attributes) == 0 {
		// TODO trace logging
		return nil
	}

	// TODO do this in a batch

	//
	return da.DB.Clauses(clause.OnConflict{
		Columns:   []clause.Column{{Name: "state_id"}, {Name: "attribute"}},
		DoUpdates: clause.AssignmentColumns([]string{"data"}),
	}).Create(attributes).Error

	//for _, attribute := range attributes {
	//	if err := da.InsertAttribute(attribute); err != nil {
	//		return err
	//		TODO return proper error her when batched.
	//		return fmt.Errorf()
	//}
	//}

	//return nil
}

// FindConfigAttributes retrieves configuration entries by state_id.
func (da *BackendStorage) FindConfigAttributes(stateID string) ([]*models.StateConfigAttribute, error) {
	var configs []*models.StateConfigAttribute
	if err := da.DB.Where("state_id = ?", stateID).Find(&configs).Error; err != nil {
		log.Printf("error fetching configs for state_id: %s, error: %v", stateID, err)
		return nil, err
	}
	return configs, nil
}

// DeleteStateConfig deletes configuration entries by state_id.
func (da *BackendStorage) DeleteStateConfig(stateID string) error {
	return da.DB.Where("state_id = ?", stateID).Delete(&models.StateConfigAttribute{}).Error
}
