package routing

import (
	"context"
	"github.com/nats-io/nats.go"
	"time"
)

// MessageEnvelop provides an abstraction for message handling across different routing implementations.
// It encapsulates the raw message data and provides methods to acknowledge, retrieve, and convert messages.
type MessageEnvelop interface {
	// Ack acknowledges successful message processing to the message broker.
	// This typically marks the message as consumed and prevents redelivery.
	Ack(ctx context.Context) error

	// NakWithDelay acknowledges the message with a negative acknowledgment,
	// allowing it to be redelivered later after a specified delay.
	NakWithDelay(ctx context.Context, delay time.Duration) error

	// MessageRaw returns the raw message data as a byte slice.
	// Returns an error if the message is empty or cannot be retrieved.
	MessageRaw() ([]byte, error)

	// MessageString returns the message data as a string.
	// This is a convenience method that converts the raw bytes to a string.
	MessageString() (string, error)

	// MessageMap unmarshals the message data into a map[string]any.
	// Useful for handling JSON messages. Returns an error if unmarshaling fails.
	MessageMap() (map[string]any, error)
}

// Route defines the interface for message routing implementations.
// It provides methods for connecting to message brokers, publishing messages,
// subscribing to topics, and managing the connection lifecycle.
type Route interface {
	// Connect establishes a connection to the message broker.
	// For NATS implementation, this initializes the connection and JetStream if enabled.
	// Should be idempotent - safe to call multiple times.
	Connect(ctx context.Context) error

	// Request sends a request message and waits for a reply.
	// This implements the request-reply pattern where the caller blocks until a response is received.
	// The msg parameter can be []byte, string, map[string]any, or any JSON-marshalable type.
	Request(ctx context.Context, msg interface{}) (*nats.Msg, error)

	// Publish sends a message to the configured subject/topic without waiting for a response.
	// Supports both standard NATS and JetStream publishing based on configuration.
	// The msg parameter can be []byte, string, map[string]any, or any JSON-marshalable type.
	Publish(ctx context.Context, msg any) error

	// Subscribe starts listening for messages on the configured subject/topic.
	// Messages are delivered to the callback function set in the route configuration.
	// Supports both regular and queue group subscriptions.
	Subscribe(ctx context.Context) error

	// Unsubscribe stops listening for messages on the subscribed subject/topic.
	// After this call, no new messages will be delivered to the callback.
	Unsubscribe(ctx context.Context) error

	// Disconnect gracefully closes the connection to the message broker.
	// This method drains pending messages before closing to ensure clean shutdown.
	Disconnect(ctx context.Context) error

	// Flush blocks until all pending published messages have been sent to the server.
	// Useful for ensuring message delivery before proceeding.
	Flush() error

	// Drain gracefully shuts down the connection by stopping new operations
	// and processing all pending messages before closing.
	Drain() error
}
