package vault_test

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/quantumwake/alethic-ism-core-go/pkg/data/test"
	"github.com/quantumwake/alethic-ism-core-go/pkg/data/vault"
	"github.com/stretchr/testify/require"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/wait"
	"testing"
	"time"
)

// Global variables accessible to all tests in this package.
var (
	testDSN   string
	container testcontainers.Container
)

// setupPostgresContainer starts a PostgreSQL container and returns it along with a DSN.
func setupPostgresContainer(ctx context.Context) (testcontainers.Container, string, error) {
	req := testcontainers.ContainerRequest{
		Image:        "postgres:13-alpine",
		ExposedPorts: []string{"5432/tcp"},
		Env: map[string]string{
			"POSTGRES_USER":     "testuser",
			"POSTGRES_PASSWORD": "testpassword",
			"POSTGRES_DB":       "testdb",
		},
		WaitingFor: wait.ForListeningPort("5432/tcp"),
	}

	cont, err := testcontainers.GenericContainer(ctx, testcontainers.GenericContainerRequest{
		ContainerRequest: req,
		Started:          true,
	})
	if err != nil {
		return nil, "", err
	}

	// Retrieve host and mapped port to build the DSN.
	host, err := cont.Host(ctx)
	if err != nil {
		return nil, "", err
	}
	mappedPort, err := cont.MappedPort(ctx, "5432")
	if err != nil {
		return nil, "", err
	}

	// Construct DSN for connecting to Postgres.
	dsn := fmt.Sprintf("postgres://testuser:testpassword@%s:%s/testdb?sslmode=disable", host, mappedPort.Port())

	// Optionally wait a little longer to ensure the database is ready.
	time.Sleep(5 * time.Second)

	return cont, dsn, nil
}

//
//// TestMain sets up the container once for all tests, and tears it down afterwards.
//func TestMain(m *testing.M) {
//	ctx := context.Background()
//	var err error
//
//	// Set up the PostgreSQL container.
//	container, testDSN, err = setupPostgresContainer(ctx)
//	if err != nil {
//		log.Fatalf("Failed to set up PostgreSQL container: %v", err)
//	}
//
//	// Run the tests.
//	code := m.Run()
//
//	// Terminate the container.
//	if err := container.Terminate(ctx); err != nil {
//		log.Printf("Failed to terminate container: %v", err)
//	}
//
//	os.Exit(code)
//}

type MockMetaData struct {
	Name string `json:"Name"`
	Age  int    `json:"Age"`
}

func TestBackendStorage_NewDatabaseBackendStorage(t *testing.T) {
	db := vault.NewDatabaseStorage(test.DSN)
	require.NotNil(t, db)
}

func TestBackendStorage_InsertOrUpdateConfigMap(t *testing.T) {
	mockData := MockMetaData{
		Name: "Test Data",
		Age:  52,
	}

	mockDataJSON, err := json.Marshal(mockData)
	require.NoError(t, err)
	require.NotNil(t, mockDataJSON)

	cm := &vault.ConfigMap{
		Name:    "test_mock_configmap",
		Type:    vault.ConfigMapConfig,
		Data:    mockDataJSON,
		OwnerID: "test.user",
	}

	db := vault.NewDatabaseStorage(test.DSN)
	require.NotNil(t, db)

	// clean up the test data first
	err = db.DeleteConfigByOwnerAndName(cm.OwnerID, cm.Name)
	require.NoError(t, err)

	// insert the test data into the database
	err = db.InsertOrUpdateConfig(cm)
	require.NoError(t, err)

	// find the test data in the database
	foundConfigMap, err := db.FindConfig(*cm.ID)
	require.NoError(t, err)
	require.NotNil(t, foundConfigMap)

	// validate the data matches the original dataset
	var foundData MockMetaData
	err = json.Unmarshal(foundConfigMap.Data, &foundData)
	require.NoError(t, err)
	require.Equal(t, mockData, foundData)
}

func TestBackendStorage_FindConfigMap_NotExist(t *testing.T) {
	db := vault.NewDatabaseStorage(test.DSN)
	configMap, err := db.FindConfig("non-existent-ID")
	require.Error(t, err)
	require.Nil(t, configMap)
}

func TestDatabaseBackendStorage_InsertOrUpdateVault(t *testing.T) {
	db := vault.NewDatabaseStorage(test.DSN)
	vaultEntry := &vault.Vault{}
	err := db.InsertOrUpdateVault(vaultEntry)
	require.NoError(t, err)
}
