package vault

import (
	"github.com/quantumwake/alethic-ism-core-go/pkg/data"
	"github.com/quantumwake/alethic-ism-core-go/pkg/model"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

type BackendStorage struct {
	*data.Access
}

func NewBackend(dsn string) *BackendStorage {
	return &BackendStorage{
		Access: data.NewDataAccess(dsn),
	}
}

func (va *BackendStorage) FindVault(id string) (*model.Vault, error) {
	var vault model.Vault
	result := va.DB.Where("id = ?", id).First(&vault)
	if result.Error != nil {
		return nil, result.Error
	}
	return &vault, nil
}

func (va *BackendStorage) FindConfigMap(id string) (*model.ConfigMap, error) {
	var configMap model.ConfigMap
	result := va.DB.Where("id = ?", id).First(&configMap)
	if result.Error != nil {
		return nil, result.Error
	}
	return &configMap, nil
}

func (va *BackendStorage) UpsertConfigMap(configMap *model.ConfigMap) error {
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
