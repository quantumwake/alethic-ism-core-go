package processor

import (
	"fmt"
	"github.com/google/uuid"
	"github.com/quantumwake/alethic-ism-core-go/pkg/repository"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
	"log"
)

type BackendStorage struct {
	*repository.Access
}

func NewBackend(dsn string) *BackendStorage {
	return &BackendStorage{
		Access: repository.NewDataAccess(dsn),
	}
}

// NewBackendStorage creates a BackendStorage from an existing gorm.DB instance
func NewBackendStorage(db *gorm.DB) *BackendStorage {
	return &BackendStorage{
		Access: &repository.Access{DB: db},
	}
}

// FindProviderClasses fetches all provider classes from the database.
func (da *BackendStorage) FindProviderClasses() ([]ProviderClass, error) {
	var classes []ProviderClass
	if err := da.DB.Find(&classes).Error; err != nil {
		log.Printf("error fetching provider classes: %v", err)
		return nil, err
	}
	return classes, nil
}

// FindProviders fetches all processor providers for
func (da *BackendStorage) FindProviders(userID, projectID *string) ([]*Provider, error) {
	var configs []*Provider

	if err := da.DB.Where(Provider{UserID: userID, ProjectID: projectID}).Find(&configs).Error; err != nil {
		log.Printf("error fetching configs for state_id: %v, error: %v", userID, projectID)
		return nil, err
	}
	return configs, nil
}

func (da *BackendStorage) FindProviderByClassUserAndProject(className Class, userID, projectID *string) ([]*Provider, error) {
	var providers []*Provider
	if err := da.DB.Where(Provider{ClassName: className, UserID: userID, ProjectID: projectID}).Find(&providers).Error; err != nil {
		return nil, fmt.Errorf("unable to fetch processor providers by class name: %s, error: %v", className, err)
	}
	return providers, nil
}

func (da *BackendStorage) FindProviderByClass(className Class) ([]*Provider, error) {
	return da.FindProviderByClassUserAndProject(className, nil, nil)
}

func (da *BackendStorage) FindProcessorByProjectID(projectID string) ([]*Processor, error) {
	var processors []*Processor
	if err := da.DB.Where(Processor{ProjectID: projectID}).Find(&processors).Error; err != nil {
		log.Printf("error fetching processors for project_id: %s, error: %v", projectID, err)
		return nil, err
	}
	return processors, nil
}

func (da *BackendStorage) CreateOrUpdate(processor *Processor) error {
	if processor == nil {
		return fmt.Errorf("processor cannot be nil")
	}

	err := uuid.Validate(processor.ID)
	if err != nil {
		return fmt.Errorf("invalid processor ID: %s, error: %v", processor.ID, err)
	}

	if err = da.DB.Clauses(clause.OnConflict{
		Columns: []clause.Column{
			{Name: "id"},
		},
		DoUpdates: clause.Assignments(map[string]any{
			"provider_id":  processor.ProviderID,
			"properties":   processor.Properties,
			"updated_date": gorm.Expr("NOW()"),
		}),
	}).Create(processor).Error; err != nil {
		return fmt.Errorf("error upserting processor: %s, error: %v", processor.ID, err)
	}
	return nil
}

func (da *BackendStorage) FindProcessorByID(processorID string) (*Processor, error) {
	var processor Processor
	if err := da.DB.Where(Processor{ID: processorID}).First(&processor).Error; err != nil {
		log.Printf("error fetching processor by ID: %s, error: %v", processorID, err)
		return nil, err
	}
	return &processor, nil
}

func (da *BackendStorage) CreateOrUpdateProvider(provider *Provider) error {
	if provider == nil {
		return fmt.Errorf("provider cannot be nil")
	}

	err := uuid.Validate(provider.ID)
	if err != nil {
		return fmt.Errorf("invalid provider ID: %s, error: %v", provider.ID, err)
	}

	if err = da.DB.Clauses(clause.OnConflict{
		Columns: []clause.Column{
			{Name: "id"},
		},
		DoUpdates: clause.Assignments(map[string]any{
			"name":         provider.Name,
			"class_name":   provider.ClassName,
			"routing":      provider.Routing,
			"updated_date": gorm.Expr("NOW()"),
		}),
	}).Create(provider).Error; err != nil {
		return fmt.Errorf("error upserting processor provider: %s, error: %v", provider.ID, err)
	}
	return nil
}

func (da *BackendStorage) FindProviderByID(providerID string) (*Provider, error) {
	var provider Provider
	if err := da.DB.Where("id = ?", providerID).First(&provider).Error; err != nil {
		log.Printf("error fetching provider by ID: %s, error: %v", providerID, err)
		return nil, err
	}
	return &provider, nil
}
