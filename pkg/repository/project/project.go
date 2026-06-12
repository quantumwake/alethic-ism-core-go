package project

import (
	"github.com/quantumwake/alethic-ism-core-go/pkg/repository"
	"gorm.io/gorm/clause"
)

type BackendStorage struct {
	*repository.Access
}

func NewBackend(dsn string) *BackendStorage {
	return &BackendStorage{
		Access: repository.NewDataAccess(dsn),
	}
}

// NewBackendFromAccess reuses an existing connection pool (so a service that
// already holds a *repository.Access doesn't open a second pool).
func NewBackendFromAccess(access *repository.Access) *BackendStorage {
	return &BackendStorage{Access: access}
}

// FindByID methods for finding user profile data by id.
func (da *BackendStorage) FindByID(id string) (*Project, error) {
	var project Project
	result := da.DB.Where("project_id = ?", id).First(&project)
	if result.Error != nil {
		return nil, result.Error
	}
	return &project, nil
}

// FindAllByUserID finds all projects for a given user ID.
func (da *BackendStorage) FindAllByUserID(userID string) ([]Project, error) {
	var projects []Project
	result := da.DB.Where("user_id = ?", userID).Find(&projects)
	return projects, result.Error
}

// CreateOrUpdate inserts a user if it does not exist or updates the user if it does.
func (da *BackendStorage) CreateOrUpdate(project *Project) error {
	return da.DB.Clauses(clause.OnConflict{
		Columns:   []clause.Column{{Name: "project_id"}},
		DoUpdates: clause.AssignmentColumns([]string{"project_name"}),
	}).Create(project).Error
}
