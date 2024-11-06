package data

import (
	"github.com/quantumwake/alethic-ism-core-go/pkg/model"
	"testing"
	"time"
)

var (
	dataAccess *Access = NewDataAccess("host=localhost port=5432 user=postgres password=postgres1 dbname=postgres sslmode=disable")
)

func TestAccess_FindByRouteID(t *testing.T) {
	route, err := dataAccess.FindRouteByID("27bce142-8713-413a-930b-fc2783bab872:7c2ea117-b281-4b36-add9-e582d1a14fc2")

	if err != nil {
		t.Errorf("Error: %v", err)
	}

	println(route.Direction)
}

func TestAccess_FindRouteByProcessorAndDirection(t *testing.T) {
	route, err := dataAccess.FindRouteByID("27bce142-8713-413a-930b-fc2783bab872:7c2ea117-b281-4b36-add9-e582d1a14fc2")

	if err != nil {
		t.Errorf("Error: %v", err)
	}

	println(route.Direction)

	outputRoutes, err := dataAccess.FindRouteByProcessorAndDirection(route.ProcessorID, model.DirectionOutput)

	if err != nil {
		t.Errorf("Error: %v", err)
	}

	println(outputRoutes[0].Direction)
}

func TestAccess_InsertTrace(t *testing.T) {
	trace := &model.Trace{
		Partition:  "27bce142-8713-413a-930b-fc2783bab872", // for example a project id can be the partition of the logger
		Reference:  "7c2ea117-b281-4b36-add9-e582d1a14fc2", // component being logged, a reference such as (state id, template id, processor id, etc)
		Action:     "some_mock_action",
		ActionTime: time.Now().UTC(),
		Message:    "some mock test message content",
		Level:      model.LogLevelInfo,
	}

	err := dataAccess.InsertTrace(trace)

	if err != nil {
		t.Errorf("Error: %v", err)
	}
}

//
//// TestFindByRouteID tests the FindByRouteID function
//func TestFindByRouteID(t *testing.T) {
//	// Create a new mock database connection
//	//db, mock, err := sqlmock.New()
//	//assert.NoError(t, err)
//	//defer db.Close()
//
//	dataAccess.Insert
//
//	gormDB, err := gorm.Open(postgres.New(postgres.Config{
//		Conn: db,
//	}), &gorm.Config{})
//	assert.NoError(t, err)
//
//	// Set up the expected query and result
//	rows := sqlmock.NewRows([]string{"internal_id", "processor_id", "state_id", "direction", "status", "count", "current_index", "maximum_index", "id"}).
//		AddRow(1, "route-1", "state-1", "input", "active", 10, 5, 20, "uuid-1").
//		AddRow(2, "route-1", "state-2", "output", "inactive", 5, 2, 10, "uuid-2")
//
//	mock.ExpectQuery(`SELECT \* FROM "public"."processor_state" WHERE processor_id = \$1`).
//		WithArgs("route-1").
//		WillReturnRows(rows)
//
//	// Call the function being tested
//	states, err := FindByRouteID(gormDB, "route-1")
//
//	// Assert the results
//	assert.NoError(t, err)
//	assert.Len(t, states, 2)
//	assert.Equal(t, "route-1", states[0].ProcessorID)
//	assert.Equal(t, "state-1", states[0].StateID)
//	assert.Equal(t, DirectionInput, states[0].Direction)
//}
//
//// TestInsertProcessorState tests the Insert method of ProcessorState
//func TestInsertProcessorState(t *testing.T) {
//	// Create a new mock database connection
//	db, mock, err := sqlmock.New()
//	assert.NoError(t, err)
//	defer db.Close()
//
//	gormDB, err := gorm.Open(postgres.New(postgres.Config{
//		Conn: db,
//	}), &gorm.Config{})
//	assert.NoError(t, err)
//
//	// Create a sample ProcessorState
//	ps := &ProcessorState{
//		ProcessorID:  "processor-1",
//		StateID:      "state-1",
//		Direction:    DirectionInput,
//		Status:       "active",
//		Count:        new(int),
//		CurrentIndex: new(int),
//		MaximumIndex: new(int),
//	}
//	*ps.Count = 10
//	*ps.CurrentIndex = 5
//	*ps.MaximumIndex = 20
//
//	// Set up the expected query
//	mock.ExpectBegin()
//	mock.ExpectQuery(`INSERT INTO "public"."processor_state"`).
//		WithArgs(sqlmock.AnyArg(), ps.ProcessorID, ps.StateID, ps.Direction, ps.Status, ps.Count, ps.CurrentIndex, ps.MaximumIndex, sqlmock.AnyArg()).
//		WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow("new-uuid"))
//	mock.ExpectCommit()
//
//	// Call the function being tested
//	err = ps.Insert(gormDB)
//
//	// Assert the results
//	assert.NoError(t, err)
//	assert.NotEmpty(t, ps.ID) // Ensure an ID was generated
//}

func main() {
	// This main function is just a placeholder and won't be executed during testing
}
