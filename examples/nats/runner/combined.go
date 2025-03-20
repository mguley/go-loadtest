package runner

import (
	"context"
	"crypto/rand"
	"fmt"
	"log/slog"
	"sync/atomic"
	"time"

	"github.com/nats-io/nats.go"
)

// CombinedRunner implements a publish-subscribe pattern for testing.
//
// Fields:
//   - natsAddress:     The address of the NATS server.
//   - subject:         The NATS subject used for both publishing and subscribing.
//   - queueGroup:      The queue group name for subscription (optional).
//   - messageSize:     Size of the message payload in bytes.
//   - natsConn:        Active connection to the NATS server.
//   - subscription:    The NATS subscription instance.
//   - payload:         The generated random payload used in publish operations.
//   - logger:          Logger instance for recording events.
//   - msgChan:         Channel to receive NATS messages.
//   - operationToggle: Atomic counter used to alternate between operations.
type CombinedRunner struct {
	natsAddress     string
	subject         string
	queueGroup      string
	messageSize     int
	natsConn        *nats.Conn
	subscription    *nats.Subscription
	payload         []byte
	logger          *slog.Logger
	msgChan         chan *nats.Msg
	operationToggle int32
}

// NewCombinedRunner creates a new CombinedRunner instance.
//
// Parameters:
//   - natsAddress: Address of the NATS server.
//   - subject:     Subject for NATS messages.
//   - queueGroup:  Queue group for subscription (optional).
//   - messageSize: Size of the message payload in bytes.
//   - logger:      Logger instance for event logging.
//
// Returns:
//   - *CombinedRunner: A pointer to the newly created CombinedRunner.
func NewCombinedRunner(
	natsAddress string,
	subject string,
	queueGroup string,
	messageSize int,
	logger *slog.Logger,
) *CombinedRunner {
	return &CombinedRunner{
		natsAddress: natsAddress,
		subject:     subject,
		queueGroup:  queueGroup,
		messageSize: messageSize,
		logger:      logger,
		msgChan:     make(chan *nats.Msg, 1_000),
	}
}

// Setup prepares the CombinedRunner for operation by establishing a connection,
// generating a random payload, and setting up the subscription.
//
// Parameters:
//   - ctx: The context for managing cancellation and timeouts.
//
// Returns:
//   - error: An error if setup fails; otherwise, nil.
func (r *CombinedRunner) Setup(ctx context.Context) error {
	var err error
	if r.natsConn, err = nats.Connect(r.natsAddress); err != nil {
		return fmt.Errorf("failed to connect to NATS: %w", err)
	}

	// Generate random payload of specified size
	r.payload = make([]byte, r.messageSize)
	if _, err = rand.Read(r.payload); err != nil {
		return fmt.Errorf("failed to generate payload: %w", err)
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

	r.logger.Info("CombinedRunner setup complete",
		slog.String("subject", r.subject),
		slog.String("queueGroup", r.queueGroup),
		slog.Int("messageSize", r.messageSize))

	return nil
}

// Run alternates between publishing and subscribing operations to simulate a real-world pattern.
//
// Parameters:
//   - ctx: The context for managing cancellation and timeouts.
//
// Returns:
//   - error: An error if the operation fails or times out; otherwise, nil.
func (r *CombinedRunner) Run(ctx context.Context) error {
	// Toggle between publish and subscribe operations to simulate real-world patterns
	op := atomic.AddInt32(&r.operationToggle, 1) % 2

	switch {
	case op == 0:
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
			return r.natsConn.Publish(r.subject, r.payload)
		}
	default:
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
}

// Teardown cleans up resources after the test completes.
//
// Parameters:
//   - ctx: The context for managing cancellation and timeouts.
//
// Returns:
//   - error: An error if teardown fails; otherwise, nil.
func (r *CombinedRunner) Teardown(ctx context.Context) error {
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
func (r *CombinedRunner) Name() string {
	return "NATS Combined Runner"
}
