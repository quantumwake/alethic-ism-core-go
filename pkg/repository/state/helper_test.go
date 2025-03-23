package state_test

import (
	"github.com/quantumwake/alethic-ism-core-go/pkg/repository/project"
	"github.com/quantumwake/alethic-ism-core-go/pkg/repository/state"
	"github.com/quantumwake/alethic-ism-core-go/pkg/repository/test"
	"github.com/quantumwake/alethic-ism-core-go/pkg/repository/user"
	"github.com/quantumwake/alethic-ism-core-go/pkg/utils"
	"github.com/stretchr/testify/require"
	"testing"
)

var (
	backendUser    = user.NewBackend(test.DSN)
	backendProject = project.NewBackend(test.DSN)
	backendState   = state.NewBackend(test.DSN)
)

func helperState(t *testing.T, projectID string) *state.State {
	// insert state for project
	s := &state.State{
		ID:        "00000000-0000-0000-0000-00000000000b",
		ProjectID: projectID,
		Type:      state.StateBasic,
	}
	err := backendState.UpsertState(s)
	require.NoError(t, err)
	require.NotNil(t, s.ID)
	return s
}

func helperStateColumns(t *testing.T, stateID string) state.Columns {
	columns := state.Columns{
		"field_a": {StateID: stateID, Name: "field_a", DataType: state.DataTypeString, Required: utils.Bool(false)},
		"field_b": {StateID: stateID, Name: "field_b", DataType: state.DataTypeString, Required: utils.Bool(false)},
	}

	// create new data columns
	require.NoError(t, backendState.UpsertStateColumns(columns))

	// update the column names
	//columns["field_a"].Name = "field_a_updated"
	//require.NoError(t, backendState.UpsertStateColumns(columns))

	// delete the newly created state columns
	require.Equal(t, 2, backendState.DeleteStateColumns(stateID))

	return columns
}

func helperProject(t *testing.T, userID string) *user.Project {
	// insert project for user
	p := &user.Project{
		ID:     "0267a05b-8cad-49b7-8c61-49ffc221277d",
		Name:   "Test Project",
		UserID: userID,
	}
	require.NoError(t, backendProject.InsertOrUpdate(p))
	return p
}

func helperUser(t *testing.T) *user.User {
	// insert a user
	u := &user.User{
		ID:       "0267a05b-8cad-49b7-8c61-49ffc221277d",
		Name:     "Test User",
		Email:    "hello@world.com",
		MaxUnits: 10,
	}
	require.NoError(t, backendUser.InsertOrUpdate(u))
	return u
}

func helperStateConfigAttributes(t *testing.T, stateID string) []*state.ConfigAttribute {
	attributes := []*state.ConfigAttribute{
		{
			StateID:   stateID,
			Attribute: state.AttributeFlagQueryStateInheritanceAll,
			Data:      BoolTrue,
		},
		{
			StateID:   stateID,
			Attribute: state.AttributeFlagRequirePrimaryKey,
			Data:      BoolTrue,
		},
		{
			StateID:   stateID,
			Attribute: state.AttributeFlagAutoRouteOutputStateAfterSave,
			Data:      BoolTrue,
		},
	}

	//
	require.NoError(t, backendState.UpsertConfigAttributes(attributes))

	attributes[0].Data = BoolFalse
	attributes[1].Data = BoolFalse
	attributes[2].Data = BoolFalse
	require.NoError(t, backendState.UpsertConfigAttributes(attributes))

	// fetch attributes by state id and check whether the attribute data is now set to false
	fetchedAttributes, err := backendState.FindConfigAttributes(stateID)
	require.NoError(t, err)

	for _, attribute := range fetchedAttributes {
		require.Equal(t, attribute.Data, BoolFalse)
	}

	return attributes
}

func helperStateColumnKeyDefinitions(t *testing.T, stateID string) {
	definitions := []*state.ColumnKeyDefinition{
		{DefinitionType: state.DefinitionStateJoinKey, StateID: stateID, Name: "field_a", Required: utils.Bool(false), Callable: utils.Bool(false)},
		{DefinitionType: state.DefinitionStateJoinKey, StateID: stateID, Name: "field_b", Required: utils.Bool(false), Callable: utils.Bool(false)},
	}

	require.NoError(t, backendState.UpsertStateConfigKeyDefinitions(definitions))

	// fetch the keys and check them over
	fetchedDefinitions, err := backendState.FindStateConfigKeyDefinitions(stateID)
	require.NoError(t, err)
	require.Len(t, fetchedDefinitions, 2)
	require.Equal(t, definitions[0].Name, fetchedDefinitions[0].Name)
	require.Equal(t, definitions[1].Name, fetchedDefinitions[1].Name)
	require.Equal(t, definitions[0].Required, fetchedDefinitions[0].Required)
	require.Equal(t, definitions[1].Required, fetchedDefinitions[1].Required)
	require.Equal(t, definitions[0].Callable, fetchedDefinitions[0].Callable)
	require.Equal(t, definitions[1].Callable, fetchedDefinitions[1].Callable)
}
