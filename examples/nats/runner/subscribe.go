package runner

import (
	"context"
	"fmt"
	"log/slog"
	"time"

	"github.com/nats-io/nats.go"
)

// NatsSubscribeRunner implements the core.Runner interface for testing NATS subscriptions.
//
// Fields:
//   - natsAddress:  Address of the NATS server.
//   - subject:      Subject to subscribe to.
//   - queueGroup:   Queue group for subscription (optional).
//   - natsConn:     Active connection to the NATS server.
//   - subscription: The NATS subscription instance.
//   - logger:       Logger instance for event logging.
//   - msgChan:      Channel to receive NATS messages.
type NatsSubscribeRunner struct {
	natsAddress  string
	subject      string
	queueGroup   string
	natsConn     *nats.Conn
	subscription *nats.Subscription
	logger       *slog.Logger
	msgChan      chan *nats.Msg
}

// NewNatsSubscribeRunner creates a new instance of NatsSubscribeRunner.
//
// Parameters:
//   - natsAddress: Address of the NATS server.
//   - subject:     Subject to subscribe to.
//   - queueGroup:  Queue group for subscription (optional).
//   - logger:      Logger instance for logging events.
//
// Returns:
//   - *NatsSubscribeRunner: A pointer to the newly created NatsSubscribeRunner.
func NewNatsSubscribeRunner(
	natsAddress string,
	subject string,
	queueGroup string,
	logger *slog.Logger,
) *NatsSubscribeRunner {
	return &NatsSubscribeRunner{
		natsAddress: natsAddress,
		subject:     subject,
		queueGroup:  queueGroup,
		logger:      logger,
		msgChan:     make(chan *nats.Msg, 1_000),
	}
}

// Setup prepares the NatsSubscribeRunner by establishing a connection and setting up a subscription.
//
// Parameters:
//   - ctx: The context for managing cancellation and timeouts.
//
// Returns:
//   - error: An error if setup fails; otherwise, nil.
func (r *NatsSubscribeRunner) Setup(ctx context.Context) error {
	var err error
	if r.natsConn, err = nats.Connect(r.natsAddress); err != nil {
		return fmt.Errorf("failed to connect to NATS: %w", err)
	}

	// Create subscription
	switch {
	case r.queueGroup != "":
		r.subscription, err = r.natsConn.QueueSubscribe(r.subject, r.queueGroup, func(msg *nats.Msg) {
			select {
			case r.msgChan <- msg:
				// Message queued
			default:
				// Channel full, drop message
			}
		})
	default:
		r.subscription, err = r.natsConn.Subscribe(r.subject, func(msg *nats.Msg) {
			select {
			case r.msgChan <- msg:
				// Message queued
			default:
				// Channel full, drop message
			}
		})
	}

	if err != nil {
		return fmt.Errorf("failed to subscribe: %w", err)
	}

	r.logger.Info("NatsSubscribeRunner setup complete",
		slog.String("subject", r.subject),
		slog.String("queueGroup", r.queueGroup))

	return nil
}

// Run executes a single receive operation for a subscription.
//
// Parameters:
//   - ctx: The context for managing cancellation and timeouts.
//
// Returns:
//   - error: An error if no message is received within the timeout or if the context is canceled; otherwise, nil.
func (r *NatsSubscribeRunner) Run(ctx context.Context) error {
	select {
	case <-ctx.Done():
		return ctx.Err()
	case <-r.msgChan:
		// Message received, success
		return nil
	case <-time.After(time.Duration(100) * time.Millisecond):
		// No message received within timeout
		return fmt.Errorf("message timeout")
	}
}

// Teardown cleans up resources used by the NatsSubscribeRunner.
//
// Parameters:
//   - ctx: The context for managing cancellation and timeouts.
//
// Returns:
//   - error: An error if teardown fails; otherwise, nil.
func (r *NatsSubscribeRunner) Teardown(ctx context.Context) error {
	if r.subscription != nil {
		if err := r.subscription.Unsubscribe(); err != nil {
			return fmt.Errorf("failed to unsubscribe: %w", err)
		}
	}
	if r.natsConn != nil {
		r.natsConn.Close()
	}
	close(r.msgChan)
	return nil
}

// Name returns a descriptive name for this runner.
//
// Returns:
//   - string: A name identifying the runner.
func (r *NatsSubscribeRunner) Name() string {
	return "NATS Subscribe Runner"
}
