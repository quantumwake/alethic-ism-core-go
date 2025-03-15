package dsl

import (
	"alethic-ism-query-api/pkg/data"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestDSL(t *testing.T) {
	query := StateQuery{
		UserID:  "77c17315-3013-5bb8-8c42-32c28618101f",
		StateID: "465884e9-7a08-40d0-acff-148663a7c9cf",
	}

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

	da, err := data.TestNewDataAccess(t)
	assert.NoError(t, err)

	results, err := da.Query(query)
	assert.NoError(t, err)
	assert.Greater(t, len(results), 0)

	//
	//queryResults, err := da.Query(query)
	//assert.NoError(t, err)
	//assert.Greater(t, len(queryResults), 0)

	//fmt.Println("SQL Query:", sql)
	//fmt.Println("Arguments:", args)
}
