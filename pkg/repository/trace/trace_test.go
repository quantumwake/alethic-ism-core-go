package trace_test

import (
	"github.com/quantumwake/alethic-ism-core-go/pkg/repository/test"
	"github.com/quantumwake/alethic-ism-core-go/pkg/repository/trace"
	"github.com/stretchr/testify/require"
	"testing"
	"time"
)

var (
	ub = trace.NewBackend(test.DSN)
)

func TestAccess_InsertTrace(t *testing.T) {
	entry := &trace.Trace{
		Partition:  "27bce142-8713-413a-930b-fc2783bab872", // for example a project id can be the partition of the logger
		Reference:  "7c2ea117-b281-4b36-add9-e582d1a14fc2", // component being logged, a reference such as (state id, template id, processor id, etc)
		Action:     "some_mock_action",
		ActionTime: time.Now().UTC(),
		Message:    "some mock test message content",
		Level:      trace.LogLevelInfo,
	}

	err := ub.InsertTrace(entry)
	if err != nil {
		t.Errorf("Error: %v", err)
	}

	//
	require.NotNil(t, entry.ID)
	require.NoError(t, ub.DeleteTraceAllByPartition(entry.Partition))
}
