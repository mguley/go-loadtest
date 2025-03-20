package config

import (
	"os"
	"strconv"
	"strings"
	"sync"
	"time"
)

// LoadTestConfig holds configuration for NATS load tests.
//
// Fields:
//   - Duration:       Total duration of the load test.
//   - Concurrency:    Number of concurrent operations during the test.
//   - WarmupDuration: Duration of the warmup period before the test starts.
//   - ReportInterval: Interval between test reports.
//   - LogLevel:       Logging level (e.g., "info", "debug").
//   - OutputPath:     File path to output JSON reports.
//   - Tags:           Custom tags for the load test.
//   - TestType:       Type of load test (e.g., "publish", "subscribe").
//   - NatsHost:       Hostname for the NATS server.
//   - NatsPort:       Port for the NATS server.
//   - Subject:        NATS subject for the load test messages.
//   - QueueGroup:     Queue group name for load testing.
//   - MessageSize:    Size of the message payload in bytes.
type LoadTestConfig struct {
	// Common test configuration
	Duration       time.Duration
	Concurrency    int
	WarmupDuration time.Duration
	ReportInterval time.Duration
	LogLevel       string
	OutputPath     string
	Tags           map[string]string

	// NATS specific configuration
	TestType    string
	NatsHost    string
	NatsPort    string
	Subject     string
	QueueGroup  string
	MessageSize int
}

var (
	once   sync.Once
	config *LoadTestConfig
)

// GetConfig retrieves the load test configuration.
//
// Returns:
//   - *LoadTestConfig: A pointer to the load test configuration instance.
func GetConfig() *LoadTestConfig {
	once.Do(func() {
		config = loadConfig()
	})
	return config
}

// loadConfig loads configuration from environment variables.
//
// Returns:
//   - *LoadTestConfig: A pointer to the populated load test configuration.
func loadConfig() *LoadTestConfig {
	return &LoadTestConfig{
		// Common test configuration with defaults
		Duration:       getDurationEnv("LOAD_TEST_DURATION", time.Duration(30)*time.Second),
		Concurrency:    getIntEnv("LOAD_TEST_CONCURRENCY", 10),
		WarmupDuration: getDurationEnv("LOAD_TEST_WARMUP", time.Duration(5)*time.Second),
		ReportInterval: getDurationEnv("LOAD_TEST_REPORT_INTERVAL", time.Duration(1)*time.Second),
		LogLevel:       getEnv("LOAD_TEST_LOG_LEVEL", "info"),
		OutputPath:     getEnv("LOAD_TEST_OUTPUT_PATH", ""),
		Tags:           parseTags(getEnv("LOAD_TEST_TAGS", "")),

		// NATS specific configuration
		TestType:    getEnv("LOAD_TEST_TYPE", "publish"),
		NatsHost:    getEnv("NATS_HOST", "localhost"),
		NatsPort:    getEnv("NATS_PORT", "4222"),
		Subject:     getEnv("LOAD_TEST_SUBJECT", "load.test"),
		QueueGroup:  getEnv("LOAD_TEST_QUEUE_GROUP", ""),
		MessageSize: getIntEnv("LOAD_TEST_MESSAGE_SIZE", 1024),
	}
}

// getEnv retrieves the value of an environment variable or returns a fallback value.
//
// Parameters:
//   - key:      The environment variable name.
//   - fallback: Default value if the environment variable is not set.
//
// Returns:
//   - string: Value of the environment variable or the fallback.
func getEnv(key, fallback string) string {
	if v, ok := os.LookupEnv(key); ok {
		return v
	}
	return fallback
}

// getIntEnv retrieves an integer value from an environment variable.
//
// Parameters:
//   - key:      The environment variable name.
//   - fallback: Default integer value if parsing fails or variable is not set.
//
// Returns:
//   - int: The parsed integer value or the fallback.
func getIntEnv(key string, fallback int) int {
	if v, ok := os.LookupEnv(key); ok {
		if i, err := strconv.Atoi(v); err == nil {
			return i
		}
	}
	return fallback
}

// getDurationEnv retrieves a time.Duration value from an environment variable.
//
// Parameters:
//   - key:      The environment variable name.
//   - fallback: Default duration if parsing fails or variable is not set.
//
// Returns:
//   - time.Duration: The parsed duration or the fallback.
func getDurationEnv(key string, fallback time.Duration) time.Duration {
	if v, ok := os.LookupEnv(key); ok {
		if d, err := time.ParseDuration(v); err == nil {
			return d
		}
	}
	return fallback
}

// parseTags converts a comma-separated key=value string into a map of tags.
//
// Parameters:
//   - tagsStr: Comma-separated string of tags (e.g., "env=prod,version=1.0").
//
// Returns:
//   - map[string]string: A map containing tag keys and values.
func parseTags(tagsStr string) map[string]string {
	tags := make(map[string]string)
	if tagsStr == "" {
		return tags
	}

	pairs := strings.Split(tagsStr, ",")
	for _, pair := range pairs {
		kv := strings.SplitN(pair, "=", 2)
		if len(kv) == 2 {
			tags[kv[0]] = kv[1]
		}
	}

	return tags
}
