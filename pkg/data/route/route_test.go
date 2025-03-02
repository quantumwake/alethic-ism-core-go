package route_test

import (
	"github.com/quantumwake/alethic-ism-core-go/pkg/data/models"
	"github.com/quantumwake/alethic-ism-core-go/pkg/data/route"
	"github.com/quantumwake/alethic-ism-core-go/pkg/data/test"
	"testing"
)

var (
	rb = route.NewBackend(test.DSN)
)

func TestAccess_FindByRouteID(t *testing.T) {
	rt, err := rb.FindRouteByID("27bce142-8713-413a-930b-fc2783bab872:7c2ea117-b281-4b36-add9-e582d1a14fc2")

	if err != nil {
		t.Errorf("Error: %v", err)
	}

	println(rt.Direction)
}

func TestAccess_FindRouteByProcessorAndDirection(t *testing.T) {
	rt, err := rb.FindRouteByID("27bce142-8713-413a-930b-fc2783bab872:7c2ea117-b281-4b36-add9-e582d1a14fc2")
	if err != nil {
		t.Errorf("Error: %v", err)
	}

	println(rt.Direction)

	outputRoutes, err := rb.FindRouteByProcessorAndDirection(rt.ProcessorID, models.DirectionOutput)

	if err != nil {
		t.Errorf("Error: %v", err)
	}

	println(outputRoutes[0].Direction)
}
