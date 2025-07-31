package route_test

import (
	"github.com/quantumwake/alethic-ism-core-go/pkg/repository/processor"
	"github.com/quantumwake/alethic-ism-core-go/pkg/repository/route"
	"github.com/quantumwake/alethic-ism-core-go/pkg/repository/test"
	"github.com/stretchr/testify/require"
	"testing"
)

var (
	rb = route.NewBackend(test.DSN)
)

func TestAccess_FindByStateID(t *testing.T) {
	routes, err := rb.FindRouteByState("d4edad5e-46e5-43e2-9f5b-6961d55c69bc")
	if err != nil {
		t.Fatal(err)
	}
	require.Equal(t, 1, len(routes))
}

func TestAccess_FindByRouteID(t *testing.T) {
	rt, err := rb.FindRouteByID("d4edad5e-46e5-43e2-9f5b-6961d55c69bc:f6b43729-5f65-48f5-9240-892487cad28f")

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

	outputRoutes, err := rb.FindRouteByProcessorAndDirection(rt.ProcessorID, processor.DirectionOutput)

	if err != nil {
		t.Errorf("Error: %v", err)
	}

	println(outputRoutes[0].Direction)
}
