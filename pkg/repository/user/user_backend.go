package user

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

// FindUserByID methods for finding user profile data by id.
func (da *BackendStorage) FindUserByID(id string) (*User, error) {
	var user User
	result := da.DB.Where("user_id = ?", id).First(&user)
	if result.Error != nil {
		return nil, result.Error
	}
	return &user, nil
}

// InsertOrUpdate inserts a user if it does not exist or updates the user if it does.
func (da *BackendStorage) InsertOrUpdate(user *User) error {
	result := da.DB.Clauses(clause.OnConflict{
		Columns: []clause.Column{{Name: "user_id"}},
		DoUpdates: clause.AssignmentColumns([]string{
			"email",
			"name",
			"max_agentic_units",
		}),
	}).Create(user)

	return result.Error
}
