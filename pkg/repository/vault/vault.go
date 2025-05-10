package vault

import (
	"github.com/google/uuid"
	"github.com/quantumwake/alethic-ism-core-go/pkg/repository"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

type Storage interface {
	FindVault(id string) (*Vault, error)
	FindConfig(id string) (*ConfigMap, error)
	InsertOrUpdateConfig(configMap *ConfigMap) error
}

// DatabaseStorage is a database backend storage
type DatabaseStorage struct {
	Storage
	*repository.Access
}

// NewDatabaseStorage creates a new database backend storage.
func NewDatabaseStorage(dsn string) *DatabaseStorage {
	return &DatabaseStorage{
		Access: repository.NewDataAccess(dsn),
	}
}

// FindVault finds a vault in the database by its ID.
func (va *DatabaseStorage) FindVault(id string) (*Vault, error) {
	var vault Vault
	result := va.DB.Where("id = ?", id).First(&vault)
	if result.Error != nil {
		return nil, result.Error
	}
	return &vault, nil
}

// DeleteVault deletes a vault from the database.
func (va *DatabaseStorage) DeleteVault(id string) error {
	db := va.DB.Where("id = ?", id).Delete(&Vault{})
	return db.Error
}

// InsertOrUpdateVault inserts or updates a vault in the database.
func (va *DatabaseStorage) InsertOrUpdateVault(vault *Vault) error {
	db := va.DB.Clauses(clause.OnConflict{
		Columns: []clause.Column{
			{Name: "id"},
		},
		DoUpdates: clause.Assignments(map[string]any{
			"name":       vault.Name,
			"metadata":   vault.Metadata,
			"updated_at": gorm.Expr("NOW()"),
		}),
	}).Create(vault)

	return db.Error
}

// FindConfig finds a config map in the database by its ID.
func (va *DatabaseStorage) FindConfig(id string) (*ConfigMap, error) {
	var configMap ConfigMap
	result := va.DB.Where("id = ?", id).First(&configMap)
	if result.Error != nil {
		return nil, result.Error
	}
	return &configMap, nil
}

// InsertOrUpdateConfig inserts or updates a config map in the database.
func (va *DatabaseStorage) InsertOrUpdateConfig(configMap *ConfigMap) error {
	if configMap.ID == nil {
		id := uuid.New().String()
		configMap.ID = &id
	}

	// validate that the id is a uuid
	if _, err := uuid.Parse(*configMap.ID); err != nil {
		return err
	}

	db := va.DB.Clauses(clause.OnConflict{
		Columns: []clause.Column{
			{Name: "id"},
		},
		DoUpdates: clause.Assignments(map[string]any{
			"data":       configMap.Data,
			"updated_at": gorm.Expr("NOW()"),
		}),
	}).Create(configMap)

	return db.Error
}

// DeleteConfig deletes a config map from the database.
func (va *DatabaseStorage) DeleteConfig(id string) error {
	db := va.DB.Where("id = ?", id).Delete(&ConfigMap{})
	return db.Error
}

// DeleteConfigByOwnerAndName deletes a config map from the database by its owner ID and name.
func (va *DatabaseStorage) DeleteConfigByOwnerAndName(ownerID, name string) error {
	db := va.DB.Where("owner_id = ? AND name = ?", ownerID, name).Delete(&ConfigMap{})
	return db.Error
}

// FindConfigByOwnerAll finds all config maps in the database by their owner ID.
func (va *DatabaseStorage) FindConfigByOwnerAll(ownerID string) ([]ConfigMap, error) {
	var configMaps []ConfigMap
	result := va.DB.Where("owner_id = ?", ownerID).Find(&configMaps)
	if result.Error != nil {
		return nil, result.Error
	}
	return configMaps, nil
}
