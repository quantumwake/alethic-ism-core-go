package vault_test

import (
	"encoding/json"
	"github.com/quantumwake/alethic-ism-core-go/pkg/data/test"
	"github.com/quantumwake/alethic-ism-core-go/pkg/data/vault"
	"github.com/quantumwake/alethic-ism-core-go/pkg/model"
	"github.com/stretchr/testify/require"
	"testing"
)

type MockMetaData struct {
	Name string `json:"Name"`
	Age  int    `json:"Age"`
}

func TestBackendStorage_UpsertConfigMap(t *testing.T) {

	mockData := MockMetaData{
		Name: "Test Data",
		Age:  52,
	}

	mockDataJSON, err := json.Marshal(mockData)
	require.NoError(t, err)
	require.NotNil(t, mockDataJSON)

	cm := &model.ConfigMap{
		Name:    "test_mock_configmap",
		Type:    model.Config,
		Data:    mockDataJSON,
		OwnerID: "test.user",
	}

	db := vault.NewBackend(test.DSN)
	require.NotNil(t, db)

	err = db.UpsertConfigMap(cm)
	require.NoError(t, err)
}

func TestBackendStorage_FindConfigMap(t *testing.T) {
	id := "c87630f5-f4ff-4f81-ac6d-a1e0c6511a5c"

	db := vault.NewBackend(test.DSN)
	configMap, err := db.FindConfigMap(id)
	require.NoError(t, err)
	require.NotNil(t, configMap)

}
