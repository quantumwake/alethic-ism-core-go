package usage_test

import (
	"github.com/quantumwake/alethic-ism-core-go/pkg/data/usage"
	"testing"
	"time"
)

func TestBackendStorage_InsertUsage(t *testing.T) {
	usage := &usage.Usage{
		TransactionTime: time.Now(),
	}

}
