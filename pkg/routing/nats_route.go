package routing

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/nats-io/nats.go"
	"github.com/quantumwake/alethic-ism-core-go/pkg/routing/config"
	"log"
	"sync"
)

// TODO need to make this into an interface and abstract out the backend message routing away from NATS specifics

//type Route interface {
//	func (r *NATSRoute) Request(ctx context.Context, msg interface{}) (*nats.Msg, error) {
//}

type NATSRoute struct {
	//Route
	route *config.Route

	nc   *nats.Conn
	js   nats.JetStreamContext
	sub  *nats.Subscription
	mu   sync.Mutex
	once sync.Once

	Callback func(ctx context.Context, route *NATSRoute, msg *nats.Msg)
	//callback nats.MsgHandler
}

// NewNATSRoute initializes and returns a new NATSRoute instance.
func NewNATSRoute(route *config.Route) *NATSRoute {
	return &NATSRoute{route: route, Callback: nil}
}

// NewNATSRouteWithCallback initializes and returns a new NATSRoute instance.
func NewNATSRouteWithCallback(route *config.Route, callback func(ctx context.Context, route *NATSRoute, msg *nats.Msg)) *NATSRoute {
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

		//r.js.PullSubscribe()
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
// func (r *NATSRoute) Publish(ctx context.Context, msg interface{}) error {
func (r *NATSRoute) Publish(ctx context.Context, msg interface{}) error {
	data, err := toBytes(msg)
	if err != nil {
		return fmt.Errorf("failed to serialize message: %w", err)
	}

	if err := r.Connect(ctx); err != nil {
		return err
	}

	if r.route.JetStreamEnabled() {
		_, err := r.js.Publish(r.route.Subject, data)
		if err != nil {
			return fmt.Errorf("failed to publish message to JetStream: %w", err)
		}
	} else {
		if err := r.nc.Publish(r.route.Subject, data); err != nil {
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
		r.Callback(ctx, r, msg)
	}

	if r.route.JetStreamEnabled() {

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

// Unsubscribe unsubscribes from the subject.
func (r *NATSRoute) Unsubscribe(ctx context.Context) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	if r.nc == nil || r.sub == nil {
		return errors.New("not subscribed to NATS")
	}

	err := r.sub.Unsubscribe()
	if err != nil {
		return fmt.Errorf("failed to unsubscribe from subject: %w", err)
	}

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
	var data []byte
	var err error

	switch v := msg.(type) {
	case []byte:
		// If it's already a byte array, use it directly
		v = data
	case string:
		// if a string then turn it into bytes
		return []byte(v), nil
	case map[string]interface{}:
		// If it's a map, marshal it to JSON
		data, err = json.Marshal(v)
		if err != nil {
			return nil, fmt.Errorf("failed to marshal map to JSON: %w", err)
		}
	default:
		// For any other type, try to marshal it to JSON
		data, err = json.Marshal(v)
		if err != nil {
			return nil, fmt.Errorf("failed to marshal message to JSON: %w", err)
		}
	}
	return data, nil
}

func (r *NATSRoute) Flush() error {
	if r.nc == nil {
		return errors.New("not connected to NATS")
	}

	return r.nc.Flush()
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
