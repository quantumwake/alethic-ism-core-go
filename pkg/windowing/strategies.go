package windowing

import (
	"github.com/quantumwake/alethic-ism-core-go/pkg/data/models"
	"github.com/quantumwake/alethic-ism-core-go/pkg/repository/state"
	"reflect"
	"time"
)

// CombineFunc defines how two BlockParts from different sources are combined.
type CombineFunc func(src1 string, e1 *BlockPart, src2 string, e2 *BlockPart,
	keyDefs state.ColumnKeyDefinitions) (models.Data, error)

// JoinCombine produces a joined output: key fields are copied once,
// non-key fields from both sources are placed side-by-side.
// Each stored x inbound pair produces one output (cross-product via AddData iteration).
func JoinCombine(src1 string, e1 *BlockPart, src2 string, e2 *BlockPart,
	keyDefs state.ColumnKeyDefinitions) (models.Data, error) {

	result := make(models.Data)

	// Copy key fields from one event (assumed identical in both)
	for _, field := range keyDefs {
		if v, ok := e1.Data[field.Name]; ok {
			result[field.Name] = v
		}
	}

	// Helper to add non-key fields
	addFields := func(e models.Data) {
		for k, v := range e {
			skip := false
			for _, field := range keyDefs {
				if k == field.Name {
					skip = true
					break
				}
			}
			if skip {
				continue
			}
			result[k] = v
		}
	}

	addFields(e1.Data)
	addFields(e2.Data)
	result["joinedAt"] = time.Now().Format(time.RFC3339)
	e1.JoinCount++
	e2.JoinCount++
	return result, nil
}

// MergeCombine produces a merged output: all fields from both sources are merged.
// Conflicting values (different non-key values for the same field) become []interface{}{v1, v2}.
// With blockPartMaxJoinCount=1, produces one merged output per key-pair.
func MergeCombine(src1 string, e1 *BlockPart, src2 string, e2 *BlockPart,
	keyDefs state.ColumnKeyDefinitions) (models.Data, error) {

	a := e1.Data
	b := e2.Data
	result := make(models.Data)

	// Process all fields from source 1
	for k, v := range a {
		if bv, exists := b[k]; exists && !reflect.DeepEqual(v, bv) {
			result[k] = []interface{}{v, bv}
		} else {
			result[k] = v
		}
	}

	// Add fields only in source 2
	for k, v := range b {
		if _, exists := a[k]; !exists {
			result[k] = v
		}
	}

	result["mergedAt"] = time.Now().Format(time.RFC3339)
	e1.JoinCount++
	e2.JoinCount++
	return result, nil
}
