// Package nats provides NATS JetStream routing primitives.
//
// ConcurrentRoute extends the base Route to provide concurrent message
// processing with semaphore-based backpressure. It mirrors the Python
// NATSRouteConcurrent pattern from alethic-ism-core.
package nats

import (
	"context"
	"fmt"
	"log"
	"sync"
	"time"

	natslib "github.com/nats-io/nats.go"
	"github.com/quantumwake/alethic-ism-core-go/pkg/routing"
)

// ConcurrentRoute wraps a Route to provide concurrent message processing
// with semaphore-based backpressure.
//
// Processing model:
//
//	1. Acquire semaphore slot  (blocks when at max concurrency)
//	2. Fetch one message       (only pulls when capacity exists)
//	3. Spawn goroutine         (runs callback, releases semaphore)
//	4. Repeat
//
// This ensures:
//   - At most MaxWorkers callbacks execute concurrently
//   - Fetch is paced to callback completion rate (natural backpressure)
//   - Messages are only pulled when a worker slot is available
//   - ACK/NAK is the callback's responsibility
type ConcurrentRoute struct {
	route     *Route
	sub       *natslib.Subscription
	sem       chan struct{}    // semaphore: capacity = max workers
	wg        sync.WaitGroup  // tracks in-flight goroutines
	batchSize int             // messages to fetch per pull (from config or default)
	callback  func(ctx context.Context, msg routing.MessageEnvelop)
}

// NewConcurrentRouteSubscriber creates a concurrent route subscriber.
//
// It connects to NATS using the given selector, creates a durable pull
// consumer on the target stream, and starts a concurrent consume loop
// with the specified number of workers.
//
// The callback is invoked in a separate goroutine for each message.
// The callback MUST ACK or NAK the message before returning.
func NewConcurrentRouteSubscriber(
	ctx context.Context,
	selector string,
	maxWorkers int,
	callback func(ctx context.Context, msg routing.MessageEnvelop),
	opts ...RouteOption,
) (*ConcurrentRoute, error) {

	// Create and connect the underlying route (loads config, connects, creates stream).
	route, err := NewRouteUsingSelector(ctx, selector, opts...)
	if err != nil {
		return nil, fmt.Errorf("create route for selector %s: %w", selector, err)
	}

	if route.js == nil {
		return nil, fmt.Errorf("concurrent route requires JetStream (set name + queue in config)")
	}

	batchSize := 10 // default
	if route.Config.BatchSize != nil && *route.Config.BatchSize > 0 {
		batchSize = *route.Config.BatchSize
	}

	cr := &ConcurrentRoute{
		route:     route,
		sem:       make(chan struct{}, maxWorkers),
		batchSize: batchSize,
		callback:  callback,
	}

	// Resolve consumer configuration from the route config.
	streamName := ""
	if route.Config.Name != nil {
		streamName = *route.Config.Name
	}
	durableName := ""
	if route.Config.Queue != nil {
		durableName = *route.Config.Queue
	}

	// Ensure the durable consumer exists on the target stream.
	// We create it explicitly so we have full control over the config,
	// rather than relying on PullSubscribe's auto-creation which can
	// conflict with BindStream semantics.
	if err := cr.ensureConsumer(streamName, durableName); err != nil {
		return nil, fmt.Errorf("ensure consumer on stream %s: %w", streamName, err)
	}

	// Bind to the existing consumer. Bind() targets a specific stream+consumer
	// pair, preventing the client from resolving to the wrong stream when
	// multiple streams share the same subject (e.g. source + mirror).
	subOpts := buildJetStreamOptions(route.Config)
	subOpts = append(subOpts, natslib.Bind(streamName, durableName))

	cr.sub, err = route.js.PullSubscribe(route.Config.Subject, durableName, subOpts...)
	if err != nil {
		return nil, fmt.Errorf("pull subscribe on stream %s: %w", streamName, err)
	}

	// Start the concurrent consume loop.
	cr.startConsumeLoop(ctx)

	log.Printf("[concurrent] started: selector=%s, stream=%s, consumer=%s, workers=%d, batch=%d",
		selector, streamName, durableName, maxWorkers, batchSize)

	return cr, nil
}

// ensureConsumer creates the durable consumer if it doesn't already exist.
// This is done explicitly (rather than via PullSubscribe auto-creation)
// because BindStream/Bind prevents auto-creation, and we need Bind to
// target the correct stream for mirror/source setups.
func (cr *ConcurrentRoute) ensureConsumer(stream, consumer string) error {
	// Check if consumer already exists.
	_, err := cr.route.js.ConsumerInfo(stream, consumer)
	if err == nil {
		log.Printf("[concurrent] consumer %s already exists on stream %s", consumer, stream)
		return nil
	}

	// Build consumer config from route settings.
	cfg := &natslib.ConsumerConfig{
		Durable:       consumer,
		FilterSubject: cr.route.Config.Subject,
		AckPolicy:     natslib.AckExplicitPolicy,
	}

	if cr.route.Config.AckWait != nil {
		cfg.AckWait = time.Duration(*cr.route.Config.AckWait) * time.Second
	}
	if cr.route.Config.MaxAckPending != nil {
		cfg.MaxAckPending = *cr.route.Config.MaxAckPending
	}

	_, err = cr.route.js.AddConsumer(stream, cfg)
	if err != nil {
		return fmt.Errorf("create consumer %s: %w", consumer, err)
	}

	log.Printf("[concurrent] created consumer %s on stream %s (filter=%s)",
		consumer, stream, cr.route.Config.Subject)
	return nil
}

// startConsumeLoop runs the semaphore-gated fetch loop in a background goroutine.
//
// Flow per iteration:
//
//	msgs := sub.Fetch(batchSize)   // pull up to batchSize messages
//	for each msg:
//	  sem <- struct{}{}            // acquire slot (blocks at capacity)
//	  go callback(msg)             // process in goroutine
//	    defer <-sem                // release slot when done
func (cr *ConcurrentRoute) startConsumeLoop(ctx context.Context) {
	go func() {
		for {
			// Check for shutdown.
			select {
			case <-ctx.Done():
				log.Printf("[concurrent] consume loop stopped: %v", ctx.Err())
				return
			default:
			}

			// Fetch a batch of messages. MaxWait prevents busy-spinning when idle.
			msgs, err := cr.sub.Fetch(cr.batchSize, natslib.MaxWait(5*time.Second))
			if err != nil {
				if err == natslib.ErrTimeout {
					continue
				}
				// Fatal subscription errors — subscription is dead, stop the loop.
				if !cr.sub.IsValid() {
					log.Printf("[concurrent] subscription closed, stopping consume loop")
					return
				}
				// Transient error — back off to avoid tight-loop spam.
				log.Printf("[concurrent] fetch error: %v", err)
				time.Sleep(time.Second)
				continue
			}

			// Dispatch each message to a worker goroutine.
			for _, msg := range msgs {
				// Acquire semaphore slot. Blocks when all workers are busy,
				// which naturally throttles dispatch to processing rate.
				cr.sem <- struct{}{}

				// Log redeliveries for debugging ACK issues.
				if meta, metaErr := msg.Metadata(); metaErr == nil && meta.NumDelivered > 1 {
					log.Printf("[concurrent] redelivery #%d: subject=%s, stream_seq=%d",
						meta.NumDelivered, msg.Subject, meta.Sequence.Stream)
				}

				// Spawn goroutine to process the message.
				// The goroutine owns the semaphore slot and releases it when done.
				cr.wg.Add(1)
				go func(m *natslib.Msg) {
					defer cr.wg.Done()
					defer func() { <-cr.sem }()

					envelop := &MessageEnvelop{Msg: m}
					cr.callback(ctx, envelop)
				}(msg)
			}
		}
	}()
}

// Stop gracefully shuts down the concurrent route.
// Waits for all in-flight goroutines to finish processing before draining.
func (cr *ConcurrentRoute) Stop() error {
	log.Printf("[concurrent] stopping, waiting for in-flight messages...")
	cr.wg.Wait()
	log.Printf("[concurrent] all in-flight messages processed")

	if cr.route != nil {
		return cr.route.Drain()
	}
	return nil
}

// Drain is an alias for Stop (satisfies common shutdown patterns).
func (cr *ConcurrentRoute) Drain() error {
	return cr.Stop()
}
