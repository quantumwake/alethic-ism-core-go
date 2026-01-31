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

func TestRoute_SubscribePublishTest(t *testing.T) {
	ctx := t.Context()

	subscribeRouteCfg := &NatConfig{
		Selector: "test/subscribe/name/*",
		Name:     ptr.String("test_subscribe_name"),
		Queue:    ptr.String("test_subscribe_queue"),
		Subject:  "test.subscribe.queue.*",
		URL:      "nats://localhost:4222",
	}

	var ch = make(chan string, 1)
	subscriberRoute := NewRoute(subscribeRouteCfg, func(ctx context.Context, msg routing.MessageEnvelop) {
		ch <- msg.Subject()
		//fmt.Printf("Received message: %s\n", msg.Subject())
		//require.Equal(t, "test.subscribe.queue.abc", msg.Subject())
		//data, err := msg.MessageString()
		//require.NoError(t, err)
		//require.NotEmpty(t, data)
		//fmt.Print(data)
	}, WithEnableChannels(true))
	err := subscriberRoute.Subscribe(ctx)
	require.NoError(t, err)

	publishRouteCfg := *subscribeRouteCfg
	publishRouteCfg.Subject = "test.subscribe.queue"
	publisherRoute := NewRoute(&publishRouteCfg, nil, WithEnableChannels(true))
	err = publisherRoute.PublishWithSuffix(ctx, "abc", MockData{
		FirstName: "Jane",
		LastName:  "Doe",
	})

	// wait for message or timeout
	var publishedSubject string
	select {
	case publishedSubject = <-ch:
		require.NoError(t, err)
		require.Equal(t, "test.subscribe.queue.abc", publishedSubject)
	case <-time.After(2 * time.Second):
		t.Fatal("Timeout waiting for message")
	}
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
	route := NewRoute(config, func(ctx context.Context, msg routing.MessageEnvelop) {
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
