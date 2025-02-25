package usage

import (
	"encoding/json"
	"fmt"
	"github.com/quantumwake/alethic-ism-core-go/pkg/data"
	"time"
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
func (da *BackendStorage) InsertUsage(usage *Usage) error {
	db := da.DB.Create(usage)

	if db.Error != nil {
		return fmt.Errorf("failed to insert trace data, error: %v", db.Error)
	}

	return nil
}

// UnmarshalJSON is a custom unmarshaler for the Usage struct to handle the transaction time field.
func (u *Usage) UnmarshalJSON(data []byte) error {

	// Define an alias struct to handle the transaction time field.
	type Alias Usage

	// Define an auxiliary struct to handle the transaction time field.
	aux := &struct {
		TransactionTime string `json:"transaction_time"`
		*Alias
	}{
		Alias: (*Alias)(u),
	}

	// Unmarshal the data into the auxiliary struct.
	if err := json.Unmarshal(data, &aux); err != nil {
		return err
	}
	var err error

	// Parse the transaction time field into the Usage struct.
	u.TransactionTime, err = time.Parse("2006-01-02T15:04:05.999999", aux.TransactionTime)
	if err != nil {
		return err
	}
	return nil
}
