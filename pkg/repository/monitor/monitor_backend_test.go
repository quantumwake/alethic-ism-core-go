package monitor_test

import (
	"testing"

	"github.com/quantumwake/alethic-ism-core-go/pkg/repository/monitor"
	"github.com/quantumwake/alethic-ism-core-go/pkg/repository/test"
	"github.com/stretchr/testify/require"
)

var backendMonitor = monitor.NewBackend(test.DSN)

func init() {
	err := backendMonitor.AutoMigrate()
	if err != nil {
		panic(err)
	}
}

func helperLogEvent(t *testing.T) *monitor.LogEvent {
	return &monitor.LogEvent{
		LogType:             "test_log",
		InternalReferenceID: 123,
		UserID:              "550e8400-e29b-41d4-a716-446655440000",
		ProjectID:           "550e8400-e29b-41d4-a716-446655440001",
		Data:                `{"test": "data"}`,
		Exception:           "",
	}
}

func TestBackendStorage_NewBackend(t *testing.T) {
	backend := monitor.NewBackend(test.DSN)
	require.NotNil(t, backend)
}

func TestBackendStorage_AutoMigrate(t *testing.T) {
	backend := monitor.NewBackend(test.DSN)
	err := backend.AutoMigrate()
	require.NoError(t, err)

	// Test idempotency - should not error when run multiple times
	err = backend.AutoMigrate()
	require.NoError(t, err)
}

func TestBackendStorage_Insert(t *testing.T) {
	event := helperLogEvent(t)

	insertedEvent, err := backendMonitor.Insert(event)
	require.NoError(t, err)
	require.NotNil(t, insertedEvent)
	require.Greater(t, insertedEvent.ID, uint(0))
	require.NotZero(t, insertedEvent.CreatedAt)
	require.Equal(t, event.LogType, insertedEvent.LogType)
	require.Equal(t, event.InternalReferenceID, insertedEvent.InternalReferenceID)
	require.Equal(t, event.UserID, insertedEvent.UserID)
	require.Equal(t, event.ProjectID, insertedEvent.ProjectID)
	require.Equal(t, event.Data, insertedEvent.Data)
}

func TestBackendStorage_Insert_NilEvent(t *testing.T) {
	insertedEvent, err := backendMonitor.Insert(nil)
	require.Error(t, err)
	require.Nil(t, insertedEvent)
	require.Contains(t, err.Error(), "nil event")
}

func TestBackendStorage_Insert_WithException(t *testing.T) {
	event := helperLogEvent(t)
	event.Exception = "Test exception message"

	insertedEvent, err := backendMonitor.Insert(event)
	require.NoError(t, err)
	require.NotNil(t, insertedEvent)
	require.Equal(t, event.Exception, insertedEvent.Exception)
}

func TestBackendStorage_Insert_MultipleEvents(t *testing.T) {
	// Test inserting multiple events to ensure append-only behavior
	event1 := helperLogEvent(t)
	event1.LogType = "type1"

	event2 := helperLogEvent(t)
	event2.LogType = "type2"

	insertedEvent1, err := backendMonitor.Insert(event1)
	require.NoError(t, err)
	require.NotNil(t, insertedEvent1)

	insertedEvent2, err := backendMonitor.Insert(event2)
	require.NoError(t, err)
	require.NotNil(t, insertedEvent2)

	// Ensure different IDs
	require.NotEqual(t, insertedEvent1.ID, insertedEvent2.ID)
	require.Greater(t, insertedEvent2.ID, insertedEvent1.ID)
}

func TestBackendStorage_FindByUserID(t *testing.T) {
	// Create test events with specific user ID
	userID := "test-user-findby-123"

	event1 := helperLogEvent(t)
	event1.UserID = userID
	event1.LogType = "test_type_1"

	event2 := helperLogEvent(t)
	event2.UserID = userID
	event2.LogType = "test_type_2"

	// Insert multiple events for the same user
	_, err := backendMonitor.Insert(event1)
	require.NoError(t, err)

	_, err = backendMonitor.Insert(event2)
	require.NoError(t, err)

	// Find by user ID
	foundEvents, err := backendMonitor.FindByUserID(userID)
	require.NoError(t, err)
	require.NotNil(t, foundEvents)
	require.GreaterOrEqual(t, len(foundEvents), 1)

	// Verify all returned events belong to the user
	for _, event := range foundEvents {
		require.Equal(t, userID, event.UserID)
	}
}

func TestBackendStorage_FindByUserID_NotFound(t *testing.T) {
	// Try to find with non-existent user ID
	foundEvents, err := backendMonitor.FindByUserID("non-existent-user-999")
	require.NoError(t, err)
	if foundEvents == nil {
		foundEvents = []monitor.LogEvent{}
	}
	require.Len(t, foundEvents, 0)
}

func TestBackendStorage_FindByProjectID(t *testing.T) {
	// Create test events with specific project ID
	projectID := "test-proj-456"

	event1 := helperLogEvent(t)
	event1.ProjectID = projectID
	event1.LogType = "project_log_1"

	event2 := helperLogEvent(t)
	event2.ProjectID = projectID
	event2.LogType = "project_log_2"

	// Insert multiple events for the same project
	_, err := backendMonitor.Insert(event1)
	require.NoError(t, err)

	_, err = backendMonitor.Insert(event2)
	require.NoError(t, err)

	// Find by project ID
	foundEvents, err := backendMonitor.FindByProjectID(projectID)
	require.NoError(t, err)
	require.NotNil(t, foundEvents)
	require.GreaterOrEqual(t, len(foundEvents), 1)

	// Verify all returned events belong to the project
	for _, event := range foundEvents {
		require.Equal(t, projectID, event.ProjectID)
	}
}

func TestBackendStorage_FindByProjectID_NotFound(t *testing.T) {
	// Try to find with non-existent project ID
	foundEvents, err := backendMonitor.FindByProjectID("non-existent-project-999")
	require.NoError(t, err)
	if foundEvents == nil {
		foundEvents = []monitor.LogEvent{}
	}
	require.Len(t, foundEvents, 0)
}

func TestBackendStorage_FindBy_NotFound(t *testing.T) {
	// Try to find with non-existent reference ID
	var nonExistentID uint64 = 999999
	foundEvents, err := backendMonitor.FindBy("", "", nonExistentID)
	require.NoError(t, err)
	if foundEvents == nil {
		foundEvents = []monitor.LogEvent{}
	}
	require.Len(t, foundEvents, 0)
}
