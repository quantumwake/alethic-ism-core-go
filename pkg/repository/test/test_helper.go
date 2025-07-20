package test

import (
	"fmt"
	"github.com/aws/smithy-go/ptr"
	"github.com/google/uuid"
	"github.com/quantumwake/alethic-ism-core-go/pkg/repository/processor"
	"github.com/quantumwake/alethic-ism-core-go/pkg/repository/project"
	"github.com/quantumwake/alethic-ism-core-go/pkg/repository/user"
	"github.com/quantumwake/alethic-ism-core-go/pkg/routing/nats"
)

const (
	DSN = "host=localhost port=5432 user=postgres password=postgres1 dbname=postgres sslmode=disable"
)

var (
	userBackend      = user.NewBackend(DSN)
	projectBackend   = project.NewBackend(DSN)
	processorBackend = processor.NewBackend(DSN)
)

func CreateTestUser(id string) (*user.User, error) {
	if id == "" {
		id = uuid.NewString()
	}
	testUser := &user.User{
		ID:       id,
		Name:     "Test User",
		Email:    "hello@world.com",
		MaxUnits: 10,
	}

	// insert the user
	err := userBackend.CreateOrUpdate(testUser)
	return testUser, err
}

func CreateTestProject(userID, id, name string) (*project.Project, error) {
	if id == "" {
		id = uuid.NewString()
	}

	if name == "" {
		name = fmt.Sprintf("Project %s", id[:8])
	}

	testProject := &project.Project{
		ID:     id,
		UserID: userID,
		Name:   name,
	}

	err := projectBackend.CreateOrUpdate(testProject)
	return testProject, err
}

type MockRoute struct {
	*nats.NatConfig
}

func CreateTestProvider(userID, projectID, name *string) (*processor.Provider, error) {
	if name == nil {
		name = ptr.String("Test Provider")
	}

	testProvider := &processor.Provider{
		ID:        uuid.NewString(),
		ClassName: processor.Proprietary,
		UserID:    userID,
		ProjectID: projectID,
		Name:      *name,
		Routing:   map[string]any{},
	}

	err := processorBackend.CreateOrUpdateProvider(testProvider)
	return testProvider, err
}

func CreateTestProcessor(projectID, providerID, id, name string) (*processor.Processor, error) {
	if id == "" {
		id = uuid.NewString()
	}

	if name == "" {
		name = fmt.Sprintf("Processor %s", id[:8])
	}

	testProcessor := &processor.Processor{
		ID:        id,
		ProjectID: projectID,
		Name:      name,
	}

	// Only set ProviderID if it's not empty
	if providerID != "" {
		testProcessor.ProviderID = &providerID
	}

	err := processorBackend.CreateOrUpdate(testProcessor)
	return testProcessor, err
}
