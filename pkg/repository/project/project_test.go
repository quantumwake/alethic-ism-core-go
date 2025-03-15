package project_test

import (
	"github.com/quantumwake/alethic-ism-core-go/pkg/data/models/user"
	"github.com/quantumwake/alethic-ism-core-go/pkg/data/project"
	"github.com/quantumwake/alethic-ism-core-go/pkg/data/test"
	"github.com/quantumwake/alethic-ism-core-go/pkg/data/user"
	user2 "github.com/quantumwake/alethic-ism-core-go/pkg/repository/user"
	"github.com/stretchr/testify/require"
	"testing"
)

var (
	backendUser    = user.NewBackend(test.DSN)
	backendProject = project.NewBackend(test.DSN)
)

func TestBackendStorage_InsertOrUpdate(t *testing.T) {
	usr := &user2.User{
		ID:       "0267a05b-8cad-49b7-8c61-49ffc221277d",
		Name:     "Test User",
		Email:    "hello@world.com",
		MaxUnits: 10,
	}

	// insert the user
	err := backendUser.InsertOrUpdate(usr)
	require.NoError(t, err)

	// create a a list of projects
	projects := []user2.Project{
		{ID: "0267a05b-8cad-49b7-8c61-49ffc221277a", UserID: usr.ID, Name: "Test Project 1"},
		{ID: "0267a05b-8cad-49b7-8c61-49ffc221277b", UserID: usr.ID, Name: "Test Project 2"},
	}

	// insert the projects
	for _, prj := range projects {
		err := backendProject.InsertOrUpdate(&prj)
		require.NoError(t, err)
	}

	// find all projects by id
	prjs, err := backendProject.FindAllByUserID(usr.ID)
	require.NoError(t, err)
	require.Len(t, prjs, 2)

}
