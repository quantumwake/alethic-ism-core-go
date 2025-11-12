package usage

import (
	"fmt"
	"github.com/quantumwake/alethic-ism-core-go/pkg/repository"
)

type BackendStorage struct {
	*repository.Access
}

func NewBackend(dsn string) *BackendStorage {
	return &BackendStorage{
		Access: repository.NewDataAccess(dsn),
	}
}

// InsertUsage methods
func (da *BackendStorage) InsertUsage(usage *Usage) error {
	db := da.DB.Create(usage)

	if db.Error != nil {
		return fmt.Errorf("failed to insert trace data, error: %v", db.Error)
	}

	return nil
}

