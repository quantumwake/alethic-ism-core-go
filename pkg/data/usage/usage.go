package usage

import (
	"fmt"
	"github.com/quantumwake/alethic-ism-core-go/pkg/data"
	"github.com/quantumwake/alethic-ism-core-go/pkg/model"
)

type BackendStorage struct {
	*data.Access
}

func NewBackend(dsn string) *BackendStorage {
	return &BackendStorage{
		Access: data.NewDataAccess(dsn),
	}
}

// InsertUsage methods
func (da *BackendStorage) InsertUsage(usage *model.Usage) error {
	db := da.DB.Create(usage)

	if db.Error != nil {
		return fmt.Errorf("failed to insert usage data, error: %v", db.Error)
	}

	return nil
}

// InsertTrace
func (da *BackendStorage) InsertTrace(trace *model.Trace) error {
	db := da.DB.Create(trace)
	if db.Error != nil {
		return fmt.Errorf("failed to insert trace data, error: %v", db.Error)
	}

	return nil
}

// FindTraceAllByPartition
func (da *BackendStorage) FindTraceAllByPartition(partition string) ([]model.Trace, error) {
	var traces []model.Trace

	result := da.DB.
		Where("partition = ?", partition).
		Find(&traces)

	return traces, result.Error
}

// FindTraceAllByPartitionAndReference
func (da *BackendStorage) FindTraceAllByPartitionAndReference(partition, reference string) ([]model.Trace, error) {
	var traces []model.Trace

	result := da.DB.
		Where("partition = ? AND reference = ?", partition, reference).
		Find(&traces)

	return traces, result.Error
}
