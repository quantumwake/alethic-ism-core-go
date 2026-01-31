package template

import (
	"github.com/quantumwake/alethic-ism-core-go/pkg/repository"
	"gorm.io/gorm"
)

type BackendStorage struct {
	*repository.Access
}

func NewBackend(dsn string) *BackendStorage {
	return &BackendStorage{
		Access: repository.NewDataAccess(dsn),
	}
}

func NewBackendWithDB(db *gorm.DB) *BackendStorage {
	return &BackendStorage{
		Access: &repository.Access{DB: db},
	}
}

// FindByID fetches a template by its ID
func (s *BackendStorage) FindByID(templateID string) (*Template, error) {
	var template Template
	result := s.DB.Where("template_id = ?", templateID).First(&template)
	if result.Error != nil {
		return nil, result.Error
	}
	return &template, nil
}

// FindByProjectID fetches all templates for a project
func (s *BackendStorage) FindByProjectID(projectID string) ([]Template, error) {
	var templates []Template
	result := s.DB.Where("project_id = ?", projectID).Find(&templates)
	return templates, result.Error
}
