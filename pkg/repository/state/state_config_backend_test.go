package state_test

import (
	"testing"
)

const (
	BoolTrue  = "true"
	BoolFalse = "false"
)

func TestBackendStorage_InsertStateConfig(t *testing.T) {
	userProfile := helperUser(t)
	userProject := helperProject(t, userProfile.ID)
	userProjectState1 := helperState(t, userProject.ID)

	helperStateConfigAttributes(t, userProjectState1.ID)
	helperStateColumnKeyDefinitions(t, userProjectState1.ID)
	helperStateColumns(t, userProjectState1.ID)

	// find the state by ID
	//s2, err := backendState.FindState(s.ID)
	//require.NoError(t, err)
	//require.NotNil(t, s2)
}
