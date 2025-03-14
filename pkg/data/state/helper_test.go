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

func helperState(t *testing.T, projectID string) *models.State {
	// insert state for project
	s := &models.State{
		ID:        "0267a05b-8cad-49b7-8c61-49ffc221277d",
		ProjectID: projectID,
		Type:      models.StateBasic,
	}
	err := backendState.UpsertState(s)
	require.NoError(t, err)
	require.NotNil(t, s.ID)
	return s
}

func helperProject(t *testing.T, userID string) *models.Project {
	// insert project for user
	p := &models.Project{
		ID:     "0267a05b-8cad-49b7-8c61-49ffc221277d",
		Name:   "Test Project",
		UserID: userID,
	}
	require.NoError(t, backendProject.InsertOrUpdate(p))
	return p
}

func helperUser(t *testing.T) *models.User {
	// insert a user
	u := &models.User{
		ID:       "0267a05b-8cad-49b7-8c61-49ffc221277d",
		Name:     "Test User",
		Email:    "hello@world.com",
		MaxUnits: 10,
	}
	require.NoError(t, backendUser.InsertOrUpdate(u))
	return u
}
