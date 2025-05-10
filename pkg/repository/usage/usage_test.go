package usage_test

import (
	"github.com/quantumwake/alethic-ism-core-go/pkg/repository/usage"
	"github.com/stretchr/testify/require"
	"testing"
	"time"
)

func TestBackendStorage_InsertUsage(t *testing.T) {
	backend := usage.NewBackend("sqlite://:memory:")
	err := backend.InsertUsage(&usage.Usage{
		TransactionTime: time.Now(),
	})
	require.NoError(t, err)
}
