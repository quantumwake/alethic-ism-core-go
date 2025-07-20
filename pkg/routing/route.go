package routing

import (
	"context"
	"github.com/nats-io/nats.go"
)

type MessageEnvelop interface {
	Ack(ctx context.Context) error
	MessageRaw() ([]byte, error)
	MessageString() (string, error)
	MessageMap() (map[string]any, error)
}

type Route interface {
	Connect(ctx context.Context) error
	Request(ctx context.Context, msg interface{}) (*nats.Msg, error)
}
