package models

type MessageType string

const (
	QueryStateEntry       MessageType = "query_state_entry"
	QueryStateDirect      MessageType = "query_state_direct"
	QueryStateRoute       MessageType = "query_state_route"
	MonitorProcessorState MessageType = "models"
)

type RouteMessage struct {
	Type       MessageType              `json:"type"`
	RouteID    string                   `json:"route_id"`
	QueryState []map[string]interface{} `json:"query_state"`
}
