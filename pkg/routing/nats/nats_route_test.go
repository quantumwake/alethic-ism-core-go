package nats

import (
	"context"
	"fmt"
	"github.com/aws/smithy-go/ptr"
	"github.com/quantumwake/alethic-ism-core-go/pkg/routing"
	"github.com/stretchr/testify/require"
	"testing"
	"time"
)

type MockData struct {
	FirstName string `json:"first_name"`
	LastName  string `json:"last_name"`
}

func TestRoute_Subscribe(t *testing.T) {

	ctx := context.Background()
	config := &NatConfig{
		Name:     ptr.String("test_subscribe_name"),
		Queue:    ptr.String("test_subscribe_queue"),
		Subject:  "test.subscribe.queue",
		URL:      "nats://localhost:4222",
		Selector: "test/subscribe/name",
	}

	updated := false
	route := NewRoute(config, func(msg routing.MessageEnvelop) {
		data, err := msg.MessageString()
		require.NoError(t, err)
		require.NotEmpty(t, data)
		fmt.Print(data)
		updated = true
	})
	err := route.Subscribe(ctx)
	require.NoError(t, err)

	require.NoError(t, route.Publish(ctx, MockData{
		FirstName: "John",
		LastName:  "Doe",
	}))

	time.Sleep(1 * time.Second)
	require.True(t, updated)
}
