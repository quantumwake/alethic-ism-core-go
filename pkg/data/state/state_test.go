package state_test

import (
	"github.com/quantumwake/alethic-ism-core-go/pkg/data/models"
	"github.com/quantumwake/alethic-ism-core-go/pkg/data/project"
	"github.com/quantumwake/alethic-ism-core-go/pkg/data/state"
	"github.com/quantumwake/alethic-ism-core-go/pkg/data/test"
	"github.com/quantumwake/alethic-ism-core-go/pkg/data/user"
	"github.com/stretchr/testify/require"
	"testing"
)

var (
	backendUser    = user.NewBackend(test.DSN)
	backendProject = project.NewBackend(test.DSN)
	backendState   = state.NewBackend(test.DSN)
)

//func TestAccess_FindState(t *testing.T) {
//	id := "0267a05b-8cad-49b7-8c61-49ffc221277d"
//	state, err := backendState.FindStateFull(id)
//	require.NoError(t, err)
//	require.NotNil(t, state)
//}

func TestAccess_InsertState(t *testing.T) {
	// insert a user
	u := &models.User{
		ID:       "0267a05b-8cad-49b7-8c61-49ffc221277d",
		Name:     "Test User",
		Email:    "hello@world.com",
		MaxUnits: 10,
	}
	require.NoError(t, backendUser.InsertOrUpdate(u))

	// insert project for user
	p := &models.Project{
		ID:     "0267a05b-8cad-49b7-8c61-49ffc221277d",
		Name:   "Test Project",
		UserID: u.ID,
	}
	require.NoError(t, backendProject.InsertOrUpdate(p))

	// insert state for project
	s1 := &models.State{
		ID:        "0267a05b-8cad-49b7-8c61-49ffc221277d",
		ProjectID: p.ID,
		StateType: models.StateConfig,
	}
	err := backendState.InsertOrUpdate(s1)
	require.NoError(t, err)
	require.NotNil(t, s1.ID)

	// find the state by ID
	s2, err := backendState.FindState(s1.ID)
	require.NoError(t, err)
	require.NotNil(t, s2)
}
