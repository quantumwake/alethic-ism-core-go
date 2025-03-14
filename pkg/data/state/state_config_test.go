package state_test

import (
	"github.com/quantumwake/alethic-ism-core-go/pkg/data/models"
	"github.com/stretchr/testify/require"
	"testing"
)

const (
	BoolTrue  = "true"
	BoolFalse = "false"
)

func helperStateConfigMap(t *testing.T, stateID string) []*models.StateConfigAttribute {
	attributes := []*models.StateConfigAttribute{
		{
			StateID:   stateID,
			Attribute: models.AttributeFlagQueryStateInheritanceAll,
			Data:      BoolTrue,
		},
		{
			StateID:   stateID,
			Attribute: models.AttributeFlagRequirePrimaryKey,
			Data:      BoolTrue,
		},
		{
			StateID:   stateID,
			Attribute: models.AttributeFlagAutoRouteOutputStateAfterSave,
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
		require.Equal(t, attribute.Data, string(BoolFalse))
	}

	return attributes
}

func TestBackendStorage_InsertStateConfig(t *testing.T) {
	u := helperUser(t)
	p := helperProject(t, u.ID)
	s := helperState(t, p.ID)

	helperStateConfigMap(t, s.ID)
	// find the state by ID
	//s2, err := backendState.FindState(s.ID)
	//require.NoError(t, err)
	//require.NotNil(t, s2)
}
