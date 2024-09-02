package data

type MessageType string

const (
	QueryStateEntry       MessageType = "query_state_entry"
	QueryStateDirect      MessageType = "query_state_direct"
	MonitorProcessorState MessageType = "processor_state"
)

type RouteMessage struct {
	Type       MessageType              `json:"type"`
	RouteID    string                   `json:"route_id"`
	QueryState []map[string]interface{} `json:"query_state"`
}

// ProcessorStatusCode represents the possible statuses of a processor.
type ProcessorStatusCode string

// Enum-like constants for ProcessorStatusCode
const (
	Created   ProcessorStatusCode = "CREATED"
	Route     ProcessorStatusCode = "ROUTE"
	Routed    ProcessorStatusCode = "ROUTED"
	Queued    ProcessorStatusCode = "QUEUED"
	Running   ProcessorStatusCode = "RUNNING"
	Terminate ProcessorStatusCode = "TERMINATE"
	Stopped   ProcessorStatusCode = "STOPPED"
	Completed ProcessorStatusCode = "COMPLETED"
	Failed    ProcessorStatusCode = "FAILED"
)

// MonitorMessage represents the structure of the monitor message
type MonitorMessage struct {
	Type      MessageType         `json:"type"`
	RouteID   string              `json:"route_id"`
	Status    ProcessorStatusCode `json:"status"`
	Exception string              `json:"exception,omitempty"`
	Data      interface{}         `json:"data"`
}

//     response_message = {
//        "type": "query_state_entry",
//        "route_id": route_id,
//        "query_state": [query_state_entry]
//    }
