// Package windowing provides a sliding-window block store for correlating
// data events from multiple sources. It supports pluggable combine strategies
// (e.g. join, merge) and automatic TTL-based eviction of stale blocks and parts.
package windowing

import (
	"github.com/quantumwake/alethic-ism-core-go/pkg/data/models"
	"time"
)

// BlockPart wraps a single data event with TTL tracking and a join/combine counter.
// Each inbound event is stored as a BlockPart within its Block, keyed by source.
type BlockPart struct {
	Data      models.Data // the raw event payload
	ExpireAt  time.Time   // absolute expiry; part is skipped after this time
	JoinCount int         // how many times this part has been combined with another
}
