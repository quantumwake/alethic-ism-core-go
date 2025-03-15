package processor

import "github.com/quantumwake/alethic-ism-core-go/pkg/data/models"

// MonitorMessage represents the structure of the monitor message
type MonitorMessage struct {
	Type      models.MessageType `json:"type"`
	RouteID   string             `json:"route_id"`
	Status    ProcessorStatus    `json:"status"`
	Exception string             `json:"exception,omitempty"`
	Data      interface{}        `json:"data"`
}
