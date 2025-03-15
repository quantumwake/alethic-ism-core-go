package state_test

import (
	"github.com/quantumwake/alethic-ism-core-go/pkg/data/models/state"
	"github.com/quantumwake/alethic-ism-core-go/pkg/utils"
	"github.com/stretchr/testify/require"
	"testing"
)

const (
	BoolTrue  = "true"
	BoolFalse = "false"
)

func helperStateConfigAttributes(t *testing.T, stateID string) []*state.StateConfigAttribute {
	attributes := []*state.StateConfigAttribute{
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

func helperStateDataKeyDefinitions(t *testing.T, stateID string) {
	priamryKey := []*state.DataKeyDefinition{
		{ID: nil, DefinitionType: state.DefinitionPrimaryKey, StateID: stateID, Name: "field_a", Required: utils.Bool(false), Callable: utils.Bool(false)},
		{ID: nil, DefinitionType: state.DefinitionPrimaryKey, StateID: stateID, Name: "field_a", Required: utils.Bool(false), Callable: utils.Bool(false)},
	}

}

func TestBackendStorage_InsertStateConfig(t *testing.T) {
	u := helperUser(t)
	p := helperProject(t, u.ID)
	s := helperState(t, p.ID)

	helperStateConfigAttributes(t, s.ID)

	// find the state by ID
	//s2, err := backendState.FindState(s.ID)
	//require.NoError(t, err)
	//require.NotNil(t, s2)
}
