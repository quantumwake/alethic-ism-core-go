package models

type MessageType string

const (
	QueryStateEntry       MessageType = "query_state_entry"
	QueryStateDirect      MessageType = "query_state_direct"
	QueryStateRoute       MessageType = "query_state_route"
	MonitorProcessorState MessageType = "models"
)

// Data represents an incoming JSON event.
type Data map[string]any

type RouteMessage struct {
	Type    MessageType `json:"type"`
	RouteID string      `json:"route_id"`
	//QueryState []map[string]interface{} `json:"query_state"`
	QueryState []Data `json:"query_state"`
}
