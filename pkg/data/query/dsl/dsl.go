package dsl

import (
	"fmt"
	"strings"
)

type Operator string

const (
	Equal       Operator = "="
	Like        Operator = "like"
	NotEqual    Operator = "!="
	GreaterThan Operator = ">"
	LessThan    Operator = "<"
)

// Filter represents a single filter condition
type Filter struct {
	Column   string   `json:"column" example:"input"`
	Operator Operator `json:"operator" example:"="`
	Value    string   `json:"value" example:"xyz"`
}

// FilterGroup represents a group of filters combined with AND or OR logic
type FilterGroup struct {
	Filters    []Filter `json:"filters"`
	GroupLogic string   `json:"group_logic" example:"AND"` // "AND" or "OR"
}

// StateQuery represents the main query structure with multiple FilterGroups
type StateQuery struct {
	UserID       string        `json:"user_id" example:"77c17315-3013-5bb8-8c42-32c28618101f"`
	StateID      string        `json:"state_id" example:"465884e9-7a08-40d0-acff-148663a7c9cf"`
	FilterGroups []FilterGroup `json:"filter_groups"`
}

// StateQueryResult represents a single result from the query
type StateQueryResult struct {
	ColumnName string `json:"column_name"`
	DataValue  string `json:"data_value"`
	DataIndex  int    `json:"data_index"`
}

// AddFilter adds a single filter to a group
func (fg *FilterGroup) AddFilter(column string, operator Operator, value string) {
	fg.Filters = append(fg.Filters, Filter{
		Column:   column,
		Operator: operator,
		Value:    value,
	})
}

// AddFilterGroup adds a new FilterGroup to the main query
func (q *StateQuery) AddFilterGroup(group FilterGroup) {
	q.FilterGroups = append(q.FilterGroups, group)
}

func (q *StateQuery) BuildIndexQuery() (string, []interface{}) {
	var args []interface{}

	// Base SQL to select distinct indexes
	sql := `
        SELECT d.data_index 
          FROM state_column_data d
         INNER JOIN state_column c 
            ON c.id = d.column_id
         WHERE c.state_id = ?
    `
	args = append(args, q.StateID)

	// TODO add paging limits

	// If no filter groups are provided, return the base SQL
	if q.FilterGroups == nil || len(q.FilterGroups) == 0 {
		return sql, args
	}

	// Build conditions for each filter group using GROUP BY and HAVING
	var groupConditions []string
	for _, group := range q.FilterGroups {
		var filters []string
		for _, filter := range group.Filters {
			// Build individual filter condition for each column-value pair
			condition := fmt.Sprintf("SUM(CASE WHEN c.name = ? AND d.data_value %s ? THEN 1 ELSE 0 END) > 0", filter.Operator)
			filters = append(filters, condition)
			args = append(args, filter.Column, filter.Value)
		}
		// Join conditions within the group with AND logic
		groupCondition := "(" + strings.Join(filters, " AND ") + ")"
		groupConditions = append(groupConditions, groupCondition)
	}

	// Combine all group conditions with OR in the HAVING clause
	sql += `
        GROUP BY d.data_index
        HAVING ` + strings.Join(groupConditions, " OR ") + `
    `

	return sql, args
}

func (q *StateQuery) BuildFinalQuery() (string, []interface{}, error) {
	// Get the index subquery and arguments
	indexSQL, indexArgs := q.BuildIndexQuery()

	// Base SQL to fetch all columns and values for the matching indexes
	sql := fmt.Sprintf(`
        SELECT c.name AS column_name, d.data_value AS data_value, data_index as data_index
          FROM state_column_data d
         INNER JOIN state_column c 
            ON c.id = d.column_id
         WHERE c.state_id = ? AND d.data_index IN (%s)
        ORDER BY d.data_index, c.name
    `, indexSQL)

	// Final arguments list includes the state_id and the arguments for the index subquery
	args := append([]interface{}{
		q.StateID,
	}, indexArgs...)

	return sql, args, nil
}
