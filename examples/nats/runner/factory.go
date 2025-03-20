package runner

import (
	"fmt"
	"log/slog"

	"github.com/mguley/go-loadtest/examples/nats/config"
	"github.com/mguley/go-loadtest/pkg/core"
)

// NatsRunnerFactory creates various NATS runners based on the load test configuration.
//
// Fields:
//   - config: Pointer to the load test configuration.
//   - logger: Logger instance for event logging.
type NatsRunnerFactory struct {
	config *config.LoadTestConfig
	logger *slog.Logger
}

// NewNatsRunnerFactory creates a new instance of NatsRunnerFactory.
//
// Parameters:
//   - config: Pointer to the load test configuration.
//   - logger: Logger instance for logging events.
//
// Returns:
//   - *NatsRunnerFactory: A pointer to the newly created NatsRunnerFactory.
func NewNatsRunnerFactory(config *config.LoadTestConfig, logger *slog.Logger) *NatsRunnerFactory {
	return &NatsRunnerFactory{
		config: config,
		logger: logger,
	}
}

// CreateRunner creates a core.Runner based on the specified test type.
//
// Parameters:
//   - testType: The type of test to execute (e.g., "publish", "subscribe", "combined").
//
// Returns:
//   - core.Runner: An instance of a runner implementing the core.Runner interface.
//   - error: An error if the test type is unknown.
func (f *NatsRunnerFactory) CreateRunner(testType string) (core.Runner, error) {
	natsAddress := fmt.Sprintf("nats://%s:%s", f.config.NatsHost, f.config.NatsPort)

	switch testType {
	case "publish":
		return NewNatsPublishRunner(natsAddress, f.config.Subject, f.config.MessageSize, f.logger), nil
	case "subscribe":
		return NewNatsSubscribeRunner(natsAddress, f.config.Subject, f.config.QueueGroup, f.logger), nil
	case "combined":
		return NewCombinedRunner(natsAddress, f.config.Subject, f.config.QueueGroup, f.config.MessageSize, f.logger), nil
	default:
		return nil, fmt.Errorf("unknown test type: %s", testType)
	}
}

// Name returns a descriptive name for this factory.
//
// Returns:
//   - string: A name identifying the factory.
func (f *NatsRunnerFactory) Name() string {
	return "NATS Runner Factory"
}
