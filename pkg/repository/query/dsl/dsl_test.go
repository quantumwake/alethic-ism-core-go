package dsl

import (
	state_query "github.com/quantumwake/alethic-ism-core-go/pkg/repository/query"
	"github.com/quantumwake/alethic-ism-core-go/pkg/repository/test"
	"github.com/stretchr/testify/assert"
	"testing"
)

var (
	backendQueryState = state_query.NewBackend(test.DSN)
)

func TestDSL(t *testing.T) {
	//userID := "77c17315-3013-5bb8-8c42-32c28618101f"
	stateID := "465884e9-7a08-40d0-acff-148663a7c9cf"
	query := StateQuery{}

	// Define a filter group for ("input" = "xyz" AND "result" = "abc")
	group1 := FilterGroup{GroupLogic: "AND"}
	group1.AddFilter("input", Like, "token")
	group1.AddFilter("result", Like, "%information%")

	// Define another filter group for ("result" = "def")
	group2 := FilterGroup{GroupLogic: "AND"}
	group2.AddFilter("result", Like, "%knowledge%")

	// Add both groups to the query
	query.AddFilterGroup(group1)
	query.AddFilterGroup(group2)

	// query for state data, filter by added filter criteria
	results, err := backendQueryState.Query(stateID, query)
	assert.NoError(t, err)
	assert.Greater(t, len(results), 0)
}
