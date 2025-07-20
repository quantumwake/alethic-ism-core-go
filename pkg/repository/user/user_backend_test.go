package user_test

import (
	"github.com/quantumwake/alethic-ism-core-go/pkg/repository/test"
	"github.com/quantumwake/alethic-ism-core-go/pkg/repository/user"
	"github.com/stretchr/testify/require"
	"testing"
)

var (
	ub = user.NewBackend(test.DSN)
)

func TestBackendStorage_InsertOrUpdate(t *testing.T) {
	usr := &user.User{
		ID:       "0267a05b-8cad-49b7-8c61-49ffc221277d",
		Name:     "Test User",
		Email:    "hello@world.com",
		MaxUnits: 10,
	}

	// insert the user
	err := ub.CreateOrUpdate(usr)
	require.NoError(t, err)

	// find the user by ID
	usr, err = ub.FindUserByID(usr.ID)
	require.NoError(t, err)
}
