package processor_test

import (
	"github.com/quantumwake/alethic-ism-core-go/pkg/repository/processor"
	"github.com/quantumwake/alethic-ism-core-go/pkg/repository/test"
	"github.com/stretchr/testify/require"
	"testing"
)

var (
	backend = processor.NewBackend(test.DSN)
)

func TestBackendStorage_FindProviderClasses(t *testing.T) {
	classes, err := backend.FindProviderClasses()
	require.NoError(t, err)
	require.Greater(t, len(classes), 0, "Expected to find provider classes, but got none")
}

func TestBackendStorage_FindProcessorProviders(t *testing.T) {
	provider, err := backend.FindProviders(nil, nil)
	require.NoError(t, err)
	require.Greater(t, len(provider), 0, "Expected to find processor providers, but got none")
}

func TestBackendStorage_FindProviderByClassUserAndProject(t *testing.T) {
	testUser, err := test.CreateTestUser("")
	require.NoError(t, err)
	require.NotNil(t, testUser)

	testProject, err := test.CreateTestProject(testUser.ID, "", "")
	require.NoError(t, err)
	require.NotNil(t, testProject)

	testProvider, err := test.CreateTestProvider(&testUser.ID, &testProject.ID, nil)
	require.NoError(t, err)
	require.NotNil(t, testProvider)

	foundProviders, err := backend.FindProviderByClassUserAndProject(testProvider.ClassName, &testUser.ID, &testProject.ID)
	if err != nil {
		return
	}
	require.Len(t, foundProviders, 1)
}

func TestBackendStorage_FindProcessorByProjectID(t *testing.T) {
	testUser, err := test.CreateTestUser("")
	require.NoError(t, err)
	require.NotNil(t, testUser)

	testProject, err := test.CreateTestProject(testUser.ID, "", "")
	require.NoError(t, err)
	require.NotNil(t, testProject)

	testProcessor, err := test.CreateTestProcessor(testProject.ID, "", "", "test processor")
	require.NoError(t, err)
	require.NotNil(t, testProcessor)
}

func TestBackendStorage_FindProcessorByID(t *testing.T) {
	//testUser, err := test.CreateTestUser("")
	//require.NoError(t, err)
	//require.NotNil(t, testUser)
	//

	//processor, err := backend.FindProcessorByID("f6b43729-5f65-48f5-9240-892487cad28f")
	//require.NoError(t, err)
	//require.NotNil(t, processor)
}
