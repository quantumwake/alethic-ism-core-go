package state_test

import (
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
