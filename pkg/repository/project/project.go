package project

import (
	"github.com/quantumwake/alethic-ism-core-go/pkg/data"
	"github.com/quantumwake/alethic-ism-core-go/pkg/data/models/user"
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

// FindByID methods for finding user profile data by id.
func (da *BackendStorage) FindByID(id string) (*user.User, error) {
	var user user.User
	result := da.DB.Where("project_id = ?", id).First(&user)
	if result.Error != nil {
		return nil, result.Error
	}
	return &user, nil
}

// FindAllByUserID finds all projects for a given user ID.
func (da *BackendStorage) FindAllByUserID(userID string) ([]user.Project, error) {
	var projects []user.Project
	result := da.DB.Where("user_id = ?", userID).Find(&projects)
	return projects, result.Error
}

// InsertOrUpdate inserts a user if it does not exist or updates the user if it does.
func (da *BackendStorage) InsertOrUpdate(project *user.Project) error {
	return da.DB.Clauses(clause.OnConflict{
		Columns:   []clause.Column{{Name: "project_id"}},
		DoUpdates: clause.AssignmentColumns([]string{"project_name"}),
	}).Create(project).Error
}
