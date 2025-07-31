package monitor

import (
	"fmt"
	"github.com/quantumwake/alethic-ism-core-go/pkg/repository"
	"time"
)

type BackendStorage struct {
	*repository.Access
}

func NewBackend(dsn string) *BackendStorage {
	return &BackendStorage{
		Access: repository.NewDataAccess(dsn),
	}
}

// AutoMigrate creates the monitor tables if they don't exist
func (mb *BackendStorage) AutoMigrate() error {
	return mb.DB.AutoMigrate(&LogEvent{})
}

// Insert inserts a new monitor log event (append-only, no updates)
func (mb *BackendStorage) Insert(event *LogEvent) (*LogEvent, error) {
	if event == nil {
		return nil, fmt.Errorf("monitor log event cannot be nil")
	}

	event.CreatedAt = time.Now()
	if err := mb.DB.Create(event).Error; err != nil {
		return nil, fmt.Errorf("error creating monitor log event: %v", err)
	}

	return event, nil
}

func (mb *BackendStorage) FindByUserID(userID string) ([]LogEvent, error) {
	var logs []LogEvent
	if err := mb.DB.Find(&LogEvent{}, "user_id = ?", userID).Scan(&logs).Error; err != nil {
		return nil, fmt.Errorf("error finding monitor log event: %v", err)
	}
	return logs, nil
}

func (mb *BackendStorage) FindByProjectID(projectID string) ([]LogEvent, error) {
	var logs []LogEvent
	if err := mb.DB.Find(&LogEvent{}, "project_id = ?", projectID).Scan(&logs).Error; err != nil {
		return nil, fmt.Errorf("error finding monitor log event: %v", err)
	}
	return logs, nil
}

func (mb *BackendStorage) FindBy(userID, projectID string, irID uint64) ([]LogEvent, error) {
	var logs []LogEvent
	if err := mb.DB.Find(&LogEvent{}, "internal_reference_id = ?", irID).Error; err != nil {
		return nil, fmt.Errorf("error finding monitor log event: %v", err)
	}
	return logs, nil
}
