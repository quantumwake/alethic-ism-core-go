package usage_test

import (
	"encoding/json"
	"fmt"
	"github.com/quantumwake/alethic-ism-core-go/pkg/repository/usage"
	"github.com/stretchr/testify/require"
	"testing"
	"time"
)

func TestUsageUnit_String(t *testing.T) {
	data := []byte(`{"transaction_time":"2025-11-07T21:54:44.669032Z"}`)
	var u usage.Usage
	if err := json.Unmarshal(data, &u); err != nil {
		panic(err)
	}
	fmt.Println(u.TransactionTime)
}

func TestBackendStorage_InsertUsage(t *testing.T) {
	backend := usage.NewBackend("sqlite://:memory:")
	err := backend.InsertUsage(&usage.Usage{
		TransactionTime: time.Now(),
	})
	require.NoError(t, err)
}
