package trace

import (
	"fmt"
	"github.com/quantumwake/alethic-ism-core-go/pkg/data"
)

type BackendStorage struct {
	*data.Access
}

func NewBackend(dsn string) *BackendStorage {
	return &BackendStorage{
		Access: data.NewDataAccess(dsn),
	}
}

// InsertTrace insert trace data into the database.
func (da *BackendStorage) InsertTrace(trace *Trace) error {
	db := da.DB.Create(trace)
	if db.Error != nil {
		return fmt.Errorf("failed to insert trace data, error: %v", db.Error)
	}

	return nil
}

// DeleteTraceAllByPartition removes all trace entries given a specific partition id
func (da *BackendStorage) DeleteTraceAllByPartition(partition string) error {
	db := da.DB.Where("partition = ?", partition).Delete(&Trace{})
	return db.Error
}

// FindTraceAllByPartition find all trace records with partition key.
func (da *BackendStorage) FindTraceAllByPartition(partition string) ([]Trace, error) {
	var traces []Trace

	result := da.DB.
		Where("partition = ?", partition).
		Find(&traces)

	return traces, result.Error
}

// FindTraceAllByPartitionAndReference find all trace records with partition and reference keys.
func (da *BackendStorage) FindTraceAllByPartitionAndReference(partition, reference string) ([]Trace, error) {
	var traces []Trace

	result := da.DB.
		Where("partition = ? AND reference = ?", partition, reference).
		Find(&traces)

	return traces, result.Error
}
