package usage_test

import (
	"github.com/quantumwake/alethic-ism-core-go/pkg/data/test"
	"github.com/quantumwake/alethic-ism-core-go/pkg/data/usage"
	"github.com/quantumwake/alethic-ism-core-go/pkg/model"
	"testing"
	"time"
)

var (
	ub = usage.NewBackend(test.DSN)
)

func TestAccess_InsertTrace(t *testing.T) {
	trace := &model.Trace{
		Partition:  "27bce142-8713-413a-930b-fc2783bab872", // for example a project id can be the partition of the logger
		Reference:  "7c2ea117-b281-4b36-add9-e582d1a14fc2", // component being logged, a reference such as (state id, template id, processor id, etc)
		Action:     "some_mock_action",
		ActionTime: time.Now().UTC(),
		Message:    "some mock test message content",
		Level:      model.LogLevelInfo,
	}

	err := ub.InsertTrace(trace)

	if err != nil {
		t.Errorf("Error: %v", err)
	}
}

func main() {
	// This main function is just a placeholder and won't be executed during testing
}
