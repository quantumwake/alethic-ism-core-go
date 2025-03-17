package state_test

import (
	"github.com/quantumwake/alethic-ism-core-go/pkg/repository/state"
	"github.com/stretchr/testify/require"
	"testing"
)

func TestAccess_InsertState(t *testing.T) {
	u := helperUser(t)
	p := helperProject(t, u.ID)
	s := helperState(t, p.ID)

	// find the state by ID
	s2, err := backendState.FindState(s.ID)
	require.NoError(t, err)
	require.NotNil(t, s2)
}

func TestAccess_FindState(t *testing.T) {
	//u := helperUser(t)
	//p := helperProject(t, u.ID)
	//s := helperState(t, p.ID)

	// find the state by ID
	stateID := "29253fcf-0bb3-4017-a7cc-2435b82273a3"
	//stateID := "4cca3896-a8aa-4e56-91e7-6c57bd38a809"
	s2, err := backendState.FindStateFull(stateID, state.StateLoadFullNoData)
	require.NoError(t, err)
	require.NotNil(t, s2)
}
