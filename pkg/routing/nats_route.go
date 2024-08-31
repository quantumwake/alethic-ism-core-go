package routing

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/nats-io/nats.go"
	"log"
	"sync"
)

type NATSRoute struct {
	route *Route

	nc   *nats.Conn
	js   nats.JetStreamContext
	sub  *nats.Subscription
	mu   sync.Mutex
	once sync.Once

	Callback func(*NATSRoute, *nats.Msg)
	//callback nats.MsgHandler
}

// NewNATSRoute initializes and returns a new NATSRoute instance.
func NewNATSRoute(route *Route, callback func(route *NATSRoute, msg *nats.Msg)) *NATSRoute {
	return &NATSRoute{route: route, Callback: callback}
}

// Connect establishes a connection to the NATS server, initializing JetStream if enabled.
func (r *NATSRoute) Connect(ctx context.Context) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	if r.nc != nil && r.nc.IsConnected() {
		return nil // Already connected
	}

	var err error
	r.nc, err = nats.Connect(r.route.URL)
	if err != nil {
		return fmt.Errorf("failed to connect to NATS: %w", err)
	}

	if r.route.JetStreamEnabled() {
		r.js, err = r.nc.JetStream()
		if err != nil {
			return fmt.Errorf("failed to initialize JetStream: %w", err)
		}

		if _, err := r.js.StreamInfo(*r.route.Name); errors.Is(err, nats.ErrStreamNotFound) {
			_, err := r.js.AddStream(&nats.StreamConfig{
				Name:     *r.route.Name,
				Subjects: []string{r.route.Subject},
			})
			if err != nil {
				return fmt.Errorf("failed to add stream: %w", err)
			}
		} else if err != nil {
			return fmt.Errorf("failed to get stream info: %w", err)
		}
	}

	log.Printf("Connected to NATS: %v, subject: %s\n", r.route.Name, r.route.Subject)
	return nil
}

// Request sends a request and waits for a reply, returning the response.
func (r *NATSRoute) Request(ctx context.Context, msg interface{}) (*nats.Msg, error) {
	msgBytes, err := toBytes(msg)
	if err != nil {
		return nil, fmt.Errorf("failed to serialize message: %w", err)
	}

	if err := r.Connect(ctx); err != nil {
		return nil, err
	}

	resp, err := r.nc.RequestWithContext(ctx, r.route.Subject, msgBytes)
	if err != nil {
		return nil, fmt.Errorf("request failed: %w", err)
	}

	return resp, nil
}

// Publish publishes a message to the subject, either via JetStream or standard NATS.
func (r *NATSRoute) Publish(ctx context.Context, msg interface{}) error {
	msgBytes, err := toBytes(msg)
	if err != nil {
		return fmt.Errorf("failed to serialize message: %w", err)
	}

	if err := r.Connect(ctx); err != nil {
		return err
	}

	if r.route.JetStreamEnabled() {
		_, err := r.js.Publish(r.route.Subject, msgBytes)
		if err != nil {
			return fmt.Errorf("failed to publish message to JetStream: %w", err)
		}
	} else {
		if err := r.nc.Publish(r.route.Subject, msgBytes); err != nil {
			return fmt.Errorf("failed to publish message: %w", err)
		}
	}

	return nil
}

// Subscribe subscribes to the subject with an optional callback for handling incoming messages.
func (r *NATSRoute) Subscribe(ctx context.Context) error {
	if err := r.Connect(ctx); err != nil {
		return err
	}

	// wrap the callback message such that we also get the nats route that it was received on
	callback := func(msg *nats.Msg) {
		if r.Callback == nil {
			log.Printf("no callback function defined for message: %v on subject: %s", msg.Data, msg.Subject)
			return
		}
		r.Callback(r, msg)
	}

	var err error
	if r.route.Queue != nil {
		r.sub, err = r.nc.QueueSubscribe(r.route.Subject, *r.route.Queue, callback)
	} else {
		r.sub, err = r.nc.Subscribe(r.route.Subject, callback)
	}

	if err != nil {
		return fmt.Errorf("failed to subscribe to subject: %w", err)
	}

	log.Printf("Subscribed to subject: %s\n", r.route.Subject)
	return nil
}

// Disconnect drains the connection and closes it.
func (r *NATSRoute) Disconnect(ctx context.Context) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	if r.nc == nil || !r.nc.IsConnected() {
		return errors.New("not connected to NATS")
	}

	err := r.nc.Drain()
	if err != nil {
		return fmt.Errorf("failed to drain connection: %w", err)
	}

	r.nc.Close()
	return nil
}

// toBytes converts a message to a byte slice.
func toBytes(msg interface{}) ([]byte, error) {
	switch v := msg.(type) {
	case string:
		return []byte(v), nil
	case []byte:
		return v, nil
	default:
		return json.Marshal(v)
	}
}

// Drain drains and closes the connection gracefully.
func (r *NATSRoute) Drain(ctx context.Context) error {
	if r.nc == nil || !r.nc.IsConnected() {
		return nil // Not connected, nothing to drain
	}

	err := r.nc.Drain()
	if err != nil {
		return fmt.Errorf("failed to drain connection: %w", err)
	}

	return nil
}
