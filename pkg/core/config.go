package core

import "time"

// TestConfig holds common configuration parameters for all load tests.
//
// Fields:
//   - TestDuration:   The total duration to run the test.
//   - Concurrency:    The number of concurrent operations to perform.
//   - WarmupDuration: The duration to run before collecting metrics.
//   - ReportInterval: How often to output progress during the test.
//   - LogLevel:       The logging verbosity during tests.
//   - Tags:           A map to categorize and identify specific test runs.
type TestConfig struct {
	TestDuration   time.Duration
	Concurrency    int
	WarmupDuration time.Duration
	ReportInterval time.Duration
	LogLevel       string
	Tags           map[string]string
}

// NewDefaultConfig creates a TestConfig with sensible default settings.
//
// Returns:
//   - *TestConfig: A pointer to a TestConfig instance with default settings.
func NewDefaultConfig() *TestConfig {
	return &TestConfig{
		TestDuration:   time.Duration(30) * time.Second,
		Concurrency:    10,
		WarmupDuration: time.Duration(5) * time.Second,
		ReportInterval: time.Duration(1) * time.Second,
		LogLevel:       "info",
		Tags:           make(map[string]string),
	}
}
