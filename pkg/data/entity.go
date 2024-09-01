package data

import (
	"encoding/json"
)

type MessageType string

const (
	QueryStateEntry  MessageType = "query_state_entry"
	QueryStateDirect MessageType = "query_state_direct"
)

type RouteMessage struct {
	Type       MessageType     `json:"type"`
	RouteID    string          `json:"route_id"`
	QueryState json.RawMessage `json:"query_state"`

	CompositeKey *string `json:"__composite_key__;omitempty"`
}

//     response_message = {
//        "type": "query_state_entry",
//        "route_id": route_id,
//        "query_state": [query_state_entry]
//    }
