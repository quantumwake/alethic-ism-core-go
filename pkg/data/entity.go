package data

type MessageType string

const (
	QueryStateEntry  MessageType = "query_state_entry"
	QueryStateDirect MessageType = "query_state_direct"
)

type RouteMessage struct {
	Type       MessageType              `json:"type"`
	RouteID    string                   `json:"route_id"`
	QueryState []map[string]interface{} `json:"query_state"`
}

//     response_message = {
//        "type": "query_state_entry",
//        "route_id": route_id,
//        "query_state": [query_state_entry]
//    }
