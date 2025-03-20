package runner

import (
	"context"
	"crypto/rand"
	"fmt"
	"log/slog"

	"github.com/nats-io/nats.go"
)

// NatsPublishRunner implements the core.Runner interface for testing NATS publishing.
//
// Fields:
//   - natsAddress: Address of the NATS server.
//   - subject:     Subject used for publishing messages.
//   - messageSize: Size of the message payload in bytes.
//   - payload:     The generated random payload used in publish operations.
//   - natsConn:    Active connection to the NATS server.
//   - logger:      Logger instance for event logging.
type NatsPublishRunner struct {
	natsAddress string
	subject     string
	messageSize int
	payload     []byte
	natsConn    *nats.Conn
	logger      *slog.Logger
}

// NewNatsPublishRunner creates a new instance of NatsPublishRunner.
//
// Parameters:
//   - natsAddress: Address of the NATS server.
//   - subject:     Subject used for NATS publishing.
//   - messageSize: Size of the message payload in bytes.
//   - logger:      Logger instance for logging events.
//
// Returns:
//   - *NatsPublishRunner: A pointer to the newly created NatsPublishRunner.
func NewNatsPublishRunner(
	natsAddress string,
	subject string,
	messageSize int,
	logger *slog.Logger,
) *NatsPublishRunner {
	return &NatsPublishRunner{
		natsAddress: natsAddress,
		subject:     subject,
		messageSize: messageSize,
		logger:      logger,
	}
}

// Setup prepares the NatsPublishRunner by establishing a connection and generating a payload.
//
// Parameters:
//   - ctx: The context for managing cancellation and timeouts.
//
// Returns:
//   - error: An error if setup fails; otherwise, nil.
func (r *NatsPublishRunner) Setup(ctx context.Context) error {
	var err error
	if r.natsConn, err = nats.Connect(r.natsAddress); err != nil {
		return fmt.Errorf("failed to connect to NATS: %w", err)
	}

	// Generate random payload of specified size
	r.payload = make([]byte, r.messageSize)
	if _, err = rand.Read(r.payload); err != nil {
		return fmt.Errorf("failed to generate payload: %w", err)
	}

	r.logger.Info("NatsPublishRunner setup complete",
		slog.String("subject", r.subject),
		slog.Int("messageSize", r.messageSize))

	return nil
}

// Run executes a single publish operation.
//
// Parameters:
//   - ctx: The context for managing cancellation and timeouts.
//
// Returns:
//   - error: An error if the publish fails; otherwise, nil.
func (r *NatsPublishRunner) Run(ctx context.Context) error {
	select {
	case <-ctx.Done():
		return ctx.Err()
	default:
		return r.natsConn.Publish(r.subject, r.payload)
	}
}

// Teardown cleans up resources used by the NatsPublishRunner.
//
// Parameters:
//   - ctx: The context for managing cancellation and timeouts.
//
// Returns:
//   - error: An error if teardown fails; otherwise, nil.
func (r *NatsPublishRunner) Teardown(ctx context.Context) error {
	if r.natsConn != nil {
		r.natsConn.Close()
	}
	return nil
}

// Name returns a descriptive name for this runner.
//
// Returns:
//   - string: A name identifying the runner.
func (r *NatsPublishRunner) Name() string {
	return "NATS Publish Runner"
}
