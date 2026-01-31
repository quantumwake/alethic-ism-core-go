package nats

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/nats-io/nats.go"
	"github.com/quantumwake/alethic-ism-core-go/pkg/cache"
	"github.com/quantumwake/alethic-ism-core-go/pkg/routing"
	"github.com/quantumwake/alethic-ism-core-go/pkg/utils"
	"log"
	"sync"
	"time"
)

var (
	channelTTL = utils.DurationFromEnvWithDefault("SUBJECT_CHANNEL_TTL_DURATION", 10*time.Second)
)

type RouteOptions struct {
	EnableChannels bool
}

// RouteOption defines a function type for configuring Route
type RouteOption func(*Route)

type Route struct {
	routing.Route

	//NatConfig
	Config  *NatConfig
	Options *RouteOptions

	nc   *nats.Conn
	js   nats.JetStreamContext
	sub  *nats.Subscription
	mu   sync.Mutex
	once sync.Once

	Callback func(ctx context.Context, msg routing.MessageEnvelop)
	Channels cache.Cache
}

// MessageEnvelop encapsulates a NATS message for processing.
type MessageEnvelop struct {
	Msg *nats.Msg // The NATS message associated with this envelope
}

// Ack acknowledges the message, indicating successful processing.
func (msg *MessageEnvelop) Ack(_ context.Context) error {
	return msg.Msg.Ack() // TODO need to pass in the opts
}

// NakWithDelay acknowledges the message with a negative acknowledgment, allowing it to be redelivered later.
func (msg *MessageEnvelop) NakWithDelay(_ context.Context, delay time.Duration) error {
	return msg.Msg.NakWithDelay(delay) // Nack with a delay
}

// Subject returns the subject/topic associated with the message.
func (msg *MessageEnvelop) Subject() string {
	return msg.Msg.Subject
}

// MessageRaw return raw message []byte.
func (msg *MessageEnvelop) MessageRaw() ([]byte, error) {
	if msg.Msg.Data == nil {
		return nil, errors.New("message is empty")
	}
	return msg.Msg.Data, nil
}

// MessageString encode raw message bytes in a string.
func (msg *MessageEnvelop) MessageString() (string, error) {
	raw, err := msg.MessageRaw()
	if err != nil {
		return "", fmt.Errorf("failed to encode raw message in string: %w", err)
	}
	return string(raw), nil
}

// MessageMap return raw message in a map[string]any
func (msg *MessageEnvelop) MessageMap() (map[string]any, error) {
	var mapping map[string]any
	err := json.Unmarshal(msg.Msg.Data, &mapping)
	if err != nil {
		return nil, fmt.Errorf("failed to encode raw message in string: %w", err)
	}
	return mapping, nil
}

// WithEnableChannels enables channel-based message routing
func WithEnableChannels(enable bool) RouteOption {
	return func(r *Route) {
		if r.Options == nil {
			r.Options = &RouteOptions{}
		}
		r.Options.EnableChannels = enable
		if enable {
			r.Channels = cache.NewLocalCacheWithOptions(
				cache.WithOptionTTL(5*time.Minute),
				cache.WithOptionCleanupInterval(10*time.Minute),
			)
		}
	}
}

// NewRoute initializes and returns a new Route instance with optional configuration
func NewRoute(config *NatConfig, callback func(ctx context.Context, msg routing.MessageEnvelop), opts ...RouteOption) *Route {
	route := &Route{
		Config:   config,
		Callback: callback,
	}

	// Apply all options
	for _, opt := range opts {
		opt(route)
	}

	return route
}

func NewRouteUsingSelector(ctx context.Context, selector string, opts ...RouteOption) (*Route, error) {
	config, err := LoadConfigFromEnv()
	if err != nil {
		return nil, fmt.Errorf("failed subscribe to selector: %s when loading config: %w", selector, err)
	}

	routeConfig, err := config.FindRouteBySelector(selector)
	if err != nil {
		return nil, fmt.Errorf("error finding route selector: %s; err: %v", selector, err)
	}

	// otherwise subscribe to the route with the callback for when messages are received
	natsRoute := NewRoute(routeConfig, nil, opts...)
	if err = natsRoute.Connect(ctx); err != nil {
		log.Fatalf("error connecting to monitor route: %v", err)
	}
	return natsRoute, nil
}

func NewRouteSubscriberUsingSelector(ctx context.Context, selector string, callback func(ctx context.Context, msg routing.MessageEnvelop), opts ...RouteOption) (*Route, error) {
	natsRoute, err := NewRouteUsingSelector(ctx, selector, opts...)
	if err != nil {
		return nil, fmt.Errorf("failed subscribe to selector: %s; err: %w", selector, err)
	}

	// set the callback, in order to handle messages at the root
	natsRoute.Callback = callback

	// subscribe to the route with the callback for when messages are received
	log.Printf("subscribing on route: %s, selector: %s", natsRoute.Config.Subject, selector)
	if err = natsRoute.Subscribe(ctx); err != nil {
		return nil, fmt.Errorf("unable to subscribe: %v", err)
	}
	log.Printf("subscribed on route: %s, selector: %s", natsRoute.Config.Subject, selector)
	return natsRoute, nil
}

// Connect establishes a connection to the NATS server, initializing JetStream if enabled.
func (r *Route) Connect(ctx context.Context) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	if r.nc != nil && r.nc.IsConnected() {
		return nil // Already connected
	}

	var err error
	r.nc, err = nats.Connect(r.Config.URL)
	if err != nil {
		return fmt.Errorf("failed to connect to NATS: %w", err)
	}

	if r.Config.JetStreamEnabled() {
		r.js, err = r.nc.JetStream()
		if err != nil {
			return fmt.Errorf("failed to initialize JetStream: %w", err)
		}

		//r.js.PullSubscribe()
		if _, err := r.js.StreamInfo(*r.Config.Name); errors.Is(err, nats.ErrStreamNotFound) {
			_, err := r.js.AddStream(&nats.StreamConfig{
				Name:     *r.Config.Name,
				Subjects: []string{r.Config.Subject},
			})
			if err != nil {
				return fmt.Errorf("failed to add stream: %w", err)
			}
		} else if err != nil {
			return fmt.Errorf("failed to get stream info: %w", err)
		}
	}

	log.Printf("Connected to NATS: %v, subject: %s\n", r.Config.Name, r.Config.Subject)
	return nil
}

// Request sends a request and waits for a reply, returning the response.
func (r *Route) Request(ctx context.Context, msg any) (*nats.Msg, error) {
	msgBytes, err := toBytes(msg)
	if err != nil {
		return nil, fmt.Errorf("failed to serialize message: %w", err)
	}

	if err := r.Connect(ctx); err != nil {
		return nil, err
	}

	resp, err := r.nc.RequestWithContext(ctx, r.Config.Subject, msgBytes)
	if err != nil {
		return nil, fmt.Errorf("request failed: %w", err)
	}

	return resp, nil
}

// Publish publishes a message to the subject, either via JetStream or standard NATS.
// func (r *NATSRoute) Publish(ctx context.Context, msg any) error {
func (r *Route) Publish(ctx context.Context, msg any) error {
	data, err := toBytes(msg)
	if err != nil {
		return fmt.Errorf("failed to serialize message: %w", err)
	}

	if err := r.Connect(ctx); err != nil {
		return err
	}

	if r.Config.JetStreamEnabled() {
		_, err := r.js.Publish(r.Config.Subject, data)
		if err != nil {
			return fmt.Errorf("failed to publish message to JetStream: %w", err)
		}
	} else {
		if err := r.nc.Publish(r.Config.Subject, data); err != nil {
			return fmt.Errorf("failed to publish message: %w", err)
		}
	}

	return nil
}

func (r *Route) PublishWithSuffix(ctx context.Context, suffix string, msg any) error {
	var err error
	var data []byte
	data, err = toBytes(msg)
	if err != nil {
		return fmt.Errorf("failed to serialize message: %w", err)
	}

	if err = r.Connect(ctx); err != nil {
		return err
	}

	// construct the subject with the suffix
	subject := fmt.Sprintf("%s.%s", r.Config.Subject, suffix)

	if r.Config.JetStreamEnabled() {
		_, err = r.js.Publish(subject, data)
		if err != nil {
			return fmt.Errorf("failed to publish message to JetStream: %w", err)
		}
	} else {
		if err = r.nc.Publish(subject, data); err != nil {
			return fmt.Errorf("failed to publish message: %w", err)
		}
	}

	return nil
}

// Subscribe subscribes to the subject with an optional callback for handling incoming messages.
func (r *Route) Subscribe(ctx context.Context) error {
	if err := r.Connect(ctx); err != nil {
		return err
	}

	mode := "push" // default mode
	if r.Config.Mode != nil {
		mode = *r.Config.Mode
	}

	if r.Config.JetStreamEnabled() {
		log.Printf("Subscribing to JetStream subject: %s (mode: %s)", r.Config.Subject, mode)
	}

	var err error
	if mode == "pull" {
		err = r.subscribePull(ctx)
	} else {
		err = r.subscribePush(ctx)
	}

	if err != nil {
		return fmt.Errorf("failed to subscribe to subject: %w", err)
	}

	log.Printf("Subscribed to subject: %s\n", r.Config.Subject)
	return nil
}

// subscribePush handles push-based subscription
func (r *Route) subscribePush(ctx context.Context) error {
	callback := func(msg *nats.Msg) {
		if r.Callback == nil {
			log.Printf("no callback function defined for message: %v on subject: %s", msg.Data, msg.Subject)
			return
		}
		envelop := &MessageEnvelop{Msg: msg}
		r.Callback(ctx, envelop)
	}

	var err error
	if r.Config.Queue != nil {
		log.Printf("Subscribing to queue subject: %s", r.Config.Subject)
		opts := buildJetStreamOptions(r.Config)
		r.sub, err = r.js.QueueSubscribe(r.Config.Subject, *r.Config.Queue, callback, opts...)
	} else {
		log.Printf("Subscribing to NATS subject: %s", r.Config.Subject)
		r.sub, err = r.nc.Subscribe(r.Config.Subject, callback)
	}

	return err
}

// subscribePull handles pull-based subscription
func (r *Route) subscribePull(ctx context.Context) error {
	batchSize := 10 // default
	if r.Config.BatchSize != nil {
		batchSize = *r.Config.BatchSize
	}

	opts := buildJetStreamOptions(r.Config)

	// For pull subscribers, the queue name is used as the durable consumer name
	// If not set, creates an ephemeral consumer
	durableName := ""
	if r.Config.Queue != nil {
		durableName = *r.Config.Queue
		log.Printf("Creating durable pull subscriber: %s", durableName)
	} else {
		log.Printf("Creating ephemeral pull subscriber (no durable name)")
	}

	sub, err := r.js.PullSubscribe(r.Config.Subject, durableName, opts...)
	if err != nil {
		return fmt.Errorf("failed to create pull subscriber: %w", err)
	}
	r.sub = sub

	log.Printf("Starting pull consumer with batch size: %d", batchSize)

	// Start background goroutine to pull messages
	go func() {
		for {
			// Check if context is cancelled before fetching
			select {
			case <-ctx.Done():
				log.Printf("Pull subscriber shutting down: %v", ctx.Err())
				return
			default:
				// Continue to fetch
			}

			msgs, err := sub.Fetch(batchSize, nats.MaxWait(5*time.Second))
			if err != nil {
				if errors.Is(err, nats.ErrTimeout) {
					continue // No messages available, keep polling
				}
				log.Printf("Error fetching messages: %v", err)
				continue
			}

			for _, msg := range msgs {
				if r.Callback != nil {
					envelop := &MessageEnvelop{Msg: msg}
					if r.Options != nil && r.Options.EnableChannels {
						r.publishWithChannel(ctx, envelop)
					} else {
						r.Callback(ctx, envelop)
					}
				}
			}
		}
	}()

	return nil
}

func (r *Route) publishWithChannel(ctx context.Context, msg routing.MessageEnvelop) (chan routing.MessageEnvelop, error) {
	subject := msg.Subject()

	//
	ch, err := r.Channels.GetCreateOrUpdate(ctx, subject, func(exists bool, value any) (any, error) {
		// if the channel already exists, then just set it to itself
		if exists {
			return value, nil
		}

		// create channel since it does not exist
		value = make(chan routing.MessageEnvelop, 1)
		return value, nil
	}, channelTTL)

	if err != nil {
		// TODO critical error here
		return nil, fmt.Errorf("failed to get or create channel for subject: %s; err: %v", subject, err)
	}

	//
	msgChan := ch.(chan routing.MessageEnvelop)
	msgChan <- msg
	return msgChan, nil
}

// Unsubscribe unsubscribes from the subject.
func (r *Route) Unsubscribe(ctx context.Context) error {
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
func (r *Route) Disconnect(ctx context.Context) error {
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
func toBytes(msg any) ([]byte, error) {
	var data []byte
	var err error

	switch v := msg.(type) {
	case []byte:
		// If it's already a byte array, use it directly
		v = data
	case string:
		// if a string then turn it into bytes
		return []byte(v), nil
	case map[string]any:
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

func (r *Route) Flush() error {
	if r.nc == nil {
		return errors.New("not connected to NATS")
	}

	return r.nc.Flush()
}

// Drain drains and closes the connection gracefully.
func (r *Route) Drain() error {
	if r.nc == nil || !r.nc.IsConnected() {
		return nil // Not connected, nothing to drain
	}

	err := r.nc.Drain()
	if err != nil {
		return fmt.Errorf("failed to drain connection: %w", err)
	}

	return nil
}

// buildJetStreamOptions builds JetStream consumer options from NatConfig
func buildJetStreamOptions(config *NatConfig) []nats.SubOpt {
	var opts []nats.SubOpt

	if config.MaxAckPending != nil {
		opts = append(opts, nats.MaxAckPending(*config.MaxAckPending))
		log.Printf("Setting MaxAckPending: %d", *config.MaxAckPending)
	}

	if config.AckWait != nil {
		ackWait := time.Duration(*config.AckWait) * time.Second
		opts = append(opts, nats.AckWait(ackWait))
		log.Printf("Setting AckWait: %v", ackWait)
	}

	return opts
}
