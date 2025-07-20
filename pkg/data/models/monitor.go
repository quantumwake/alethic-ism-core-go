package models

import (
	"github.com/quantumwake/alethic-ism-core-go/pkg/repository/processor"
)

// MonitorMessage represents the structure of the monitor message
type MonitorMessage struct {
	Type      MessageType      `json:"type"`
	RouteID   string           `json:"route_id"`
	Status    processor.Status `json:"status"`
	Exception string           `json:"exception,omitempty"`
	Data      interface{}      `json:"data"`
}
