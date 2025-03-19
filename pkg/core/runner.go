package core

import "context"

// Runner defines the interface for executing load tests.
//
// Methods:
//   - Setup:    Prepares the runner before the test begins.
//   - Run:      Executes the test operation repeatedly during the test.
//   - Teardown: Cleans up resources after the test completes.
//   - Name:     Returns a descriptive name for the runner.
type Runner interface {
	// Setup prepares the runner before the test starts.
	// Parameters:
	//   - ctx: The context for managing test execution.
	// Returns:
	//   - error: An error if setup fails, otherwise nil.
	Setup(ctx context.Context) error

	// Run executes the actual test operation.
	// Parameters:
	//   - ctx: The context for managing test execution.
	// Returns:
	//   - error: An error if execution fails, otherwise nil.
	Run(ctx context.Context) error

	// Teardown is called after the test completes to clean up resources.
	// Parameters:
	//   - ctx: The context for managing test execution.
	// Returns:
	//   - error: An error if teardown fails, otherwise nil.
	Teardown(ctx context.Context) error

	// Name returns a descriptive name for this runner.
	// Returns:
	//   - string: The runner's name.
	Name() string
}

// RunnerFactory creates Runner instances based on configuration.
//
// Methods:
//   - Create: Instantiates a new runner using the provided configuration.
//   - Name:   Returns a descriptive name for the factory.
type RunnerFactory interface {
	// Create instantiates a new runner with the given configuration.
	// Parameters:
	//   - config: An interface{} containing configuration parameters.
	// Returns:
	//   - Runner: The created runner instance.
	//   - error:  An error if creation fails, otherwise nil.
	Create(config interface{}) (Runner, error)

	// Name returns a descriptive name for this factory.
	// Returns:
	//   - string: The factory's name.
	Name() string
}

// ConfigurableRunner extends Runner with configuration capabilities.
//
// Methods:
//   - Configure: Applies configuration settings to the runner.
type ConfigurableRunner interface {
	Runner

	// Configure applies configuration to the runner.
	// Parameters:
	//   - config: An interface{} containing configuration parameters.
	// Returns:
	//   - error: An error if configuration fails, otherwise nil.
	Configure(config interface{}) error
}

// Operation represents a single unit of work to be performed during load testing.
//
// Methods:
//   - Execute: Performs the operation and returns the latency along with any error.
//   - Name:    Returns a descriptive name for the operation.
type Operation interface {
	// Execute performs the operation.
	// Parameters:
	//   - ctx: The context for managing execution.
	// Returns:
	//   - latencyMs: The operation latency in milliseconds.
	//   - err:       An error if execution fails, otherwise nil.
	Execute(ctx context.Context) (latencyMs float64, err error)

	// Name returns a descriptive name for this operation.
	// Returns:
	//   - string: The operation's name.
	Name() string
}
