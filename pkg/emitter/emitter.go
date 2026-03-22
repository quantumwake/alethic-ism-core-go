// Package emitter publishes data batches and status updates via message routes.
package emitter

import (
	"context"
	"fmt"
	"log"

	"github.com/quantumwake/alethic-ism-core-go/pkg/data/models"
	"github.com/quantumwake/alethic-ism-core-go/pkg/repository/processor"
	"github.com/quantumwake/alethic-ism-core-go/pkg/routing"
)

// Emitter publishes data to message routes.
type Emitter struct {
	stateSync   routing.Route
	stateRouter routing.Route
	monitor     routing.Route
}

// New creates an Emitter with the given routes.
func New(stateSync, stateRouter, monitor routing.Route) *Emitter {
	return &Emitter{
		stateSync:   stateSync,
		stateRouter: stateRouter,
		monitor:     monitor,
	}
}

// Publish sends a single batch of data as a RouteMessage to state sync.
func (e *Emitter) Publish(ctx context.Context, routeID string, data []models.Data) (err error) {
	defer func() {
		if flushErr := e.stateSync.Flush(); flushErr != nil && err == nil {
			err = fmt.Errorf("flush state sync: %w", flushErr)
		}
	}()

	msg := models.RouteMessage{
		Type:       models.QueryStateRoute,
		RouteID:    routeID,
		QueryState: data,
	}

	if err = e.stateSync.Publish(ctx, msg); err != nil {
		return fmt.Errorf("publish to state sync: %w", err)
	}
	return nil
}

// PublishBatch splits data into batches and publishes each one.
func (e *Emitter) PublishBatch(ctx context.Context, routeID string, data []models.Data, batchSize int) error {
	if batchSize <= 0 {
		batchSize = 100
	}

	for i := 0; i < len(data); i += batchSize {
		end := i + batchSize
		if end > len(data) {
			end = len(data)
		}
		if err := e.Publish(ctx, routeID, data[i:end]); err != nil {
			return err
		}
	}
	return nil
}

// ReportStatus sends a status update to the monitor route.
func (e *Emitter) ReportStatus(ctx context.Context, routeID string, status processor.Status, exception string) {
	msg := models.MonitorMessage{
		Type:      models.MonitorProcessorState,
		RouteID:   routeID,
		Status:    status,
		Exception: exception,
	}

	if err := e.monitor.Publish(ctx, msg); err != nil {
		log.Printf("error publishing monitor message for route %s: %v", routeID, err)
	}
	if err := e.monitor.Flush(); err != nil {
		log.Printf("error flushing monitor route: %v", err)
	}
}
