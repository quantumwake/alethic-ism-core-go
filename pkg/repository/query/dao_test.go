package query_test

import (
	"github.com/quantumwake/alethic-ism-core-go/pkg/repository/query"
	"github.com/quantumwake/alethic-ism-core-go/pkg/repository/query/dsl"
	"github.com/quantumwake/alethic-ism-core-go/pkg/repository/test"
	"github.com/stretchr/testify/require"
	"testing"
)

func TestSomething(t *testing.T) {
	// This is a placeholder test function.
	// You can add your test cases here.

	stateQuery := dsl.StateQuery{
		DataColumns: []string{
			"scenario",
			"ethical_framework_name",
		},
		FilterGroups: []dsl.FilterGroup{
			{
				Filters: []dsl.Filter{
					{Column: "age_stage", Value: "Infancy & Toddlerhood", Operator: dsl.Equal},
					{Column: "ethical_framework_code", Value: "VIR", Operator: dsl.Equal},
				},
				GroupLogic: "AND",
			},
		},
	}

	dataAccess := query.NewBackend(test.DSN)
	results, err := dataAccess.Query("dd0ce044-53eb-4e88-92e8-bac47cb20d97", stateQuery)
	require.NoError(t, err)
	require.NotNil(t, results)

	rows := results.Pivot()
	require.Greater(t, len(rows), 0)

	_, ok := rows[0]["scenario"]
	require.True(t, ok)
	_, ok = rows[0]["ethical_framework_code"]
	require.False(t, ok)
	_, ok = rows[0]["ethical_framework_name"]
	require.True(t, ok)

	keyLen := len(rows[0])
	require.Equal(t, keyLen, 2)
}

/*

   def pivot_list_of_dicts(self, data):
       result = []  # List to hold the pivoted rows
       current_dict = {}  # Dictionary to hold the current row
       current_index = 1  # Track the current index

       if not data:
           return result

       logger.info("pivoting data to table")
       # Iterate over the data
       for row in data:
           # Check if we're starting a new index
           if row['data_index'] != current_index:
               # If current_dict is not empty, add it to the result list
               if current_dict:
                   result.append(current_dict)

               # Start a new dictionary for the new index
               current_dict = {}
               current_index = row['data_index']

           current_dict = {**current_dict, row['column_name']: row['data_value']}

       logger.info("appending final dictionary to list")
       # Append the final dictionary to the result list
       if current_dict:
           result.append(current_dict)

       return result

*/
