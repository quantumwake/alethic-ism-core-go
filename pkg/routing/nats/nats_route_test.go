package nats

import (
	"context"
	"github.com/aws/smithy-go/ptr"
	"github.com/nats-io/nats.go"
	"testing"
)

func TestRoute_Subscribe(t *testing.T) {
	testRoute := NewRouteWithCallback(&NatConfig{
		Name:     ptr.String("test_subscribe_name"),
		Queue:    ptr.String("test_subscribe_queue"),
		Subject:  "test.subscribe.queue",
		URL:      "nats://localhost:4222",
		Selector: "test/subscribe/name",
	}, func(ctx context.Context, r *Route, msg *nats.Msg) {

	})
}
