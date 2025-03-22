# go-loadtest

A flexible load testing framework for Go applications.

## Overview

go-loadtest is a modular load testing framework designed to help developers measure and understand the performance characteristics of their applications. It provides a common set of interfaces and components for running load tests, collecting metrics, and generating reports.

**Key Features:**
- **Configurable concurrency** - Control the level of parallelism in your load tests.
- **Customizable duration** - Run tests for precise time periods with warmup phases.
- **Rich metrics collection** - Capture latency, throughput, errors, and system resource usage.
- **Extensible reporting** - Output results to the console or as structured JSON data.
- **Modular architecture** - Implement custom runners, collectors, and reporters as needed.

## Installation

```bash
go get github.com/mguley/go-loadtest
```

## Core Components
The framework is built around several key interfaces:
- **Runner** - Executes the actual test operations.
- **MetricsCollector** - Gathers performance metrics during test execution.
- **Reporter** - Outputs test progress and final results.
- **Orchestrator** - Coordinates the entire test process.

## Getting Started
Here's a simple example of how to set up and run a load test:
```
package main

import (
	"log/slog"
	"os"

	"github.com/mguley/go-loadtest/examples/nats/config"
	"github.com/mguley/go-loadtest/examples/nats/runner"
	"github.com/mguley/go-loadtest/pkg"
	"github.com/mguley/go-loadtest/pkg/collector"
	"github.com/mguley/go-loadtest/pkg/core"
	"github.com/mguley/go-loadtest/pkg/reporter"
)

func main() {
	// Load configuration
	cfg := config.GetConfig()

	// Setup logger
	logLevel := getLogLevel(cfg.LogLevel)
	logger := slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{Level: logLevel}))

	// Create test config
	testConfig := &core.TestConfig{
		TestDuration:   cfg.Duration,
		Concurrency:    cfg.Concurrency,
		WarmupDuration: cfg.WarmupDuration,
		ReportInterval: cfg.ReportInterval,
		LogLevel:       cfg.LogLevel,
		Tags:           cfg.Tags,
	}

	// Create orchestrator
	orchestrator := pkg.NewOrchestrator(testConfig, logger)

	// Create runner factory
	runnerFactory := runner.NewNatsRunnerFactory(cfg, logger)

	// Create and add test runner
	testRunner, err := runnerFactory.CreateRunner(cfg.TestType)
	if err != nil {
		logger.Error("Failed to create runner", slog.String("error", err.Error()))
		os.Exit(1)
	}

	orchestrator.AddRunner(testRunner)

	// Add metrics collectors
	sysCollector := collector.NewSystemCollector(logger)
	compositeCollector := collector.NewCompositeCollector(sysCollector)
	orchestrator.AddCollector(compositeCollector)

	// Add reporters
	consoleReporter := reporter.NewConsoleReporter(os.Stdout, cfg.ReportInterval)
	orchestrator.AddReporter(consoleReporter)

	if cfg.OutputPath != "" {
		jsonReporter := reporter.NewJSONReporter(cfg.OutputPath)
		jsonReporter.SetIncludeLatencies(false)
		orchestrator.AddReporter(jsonReporter)
	}

	// Run the test
	logger.Info("Starting NATS load test",
		slog.String("test_type", cfg.TestType),
		slog.String("subject", cfg.Subject),
		slog.Int("concurrency", cfg.Concurrency),
		slog.Any("duration", cfg.Duration.String()))

	if err = orchestrator.Run(); err != nil {
		logger.Error("Load test failed", slog.String("error", err.Error()))
		os.Exit(1)
	}

	logger.Info("Load test completed successfully")
}

// getLogLevel converts a string log level to slog.Level
func getLogLevel(logLevel string) slog.Level {
	switch logLevel {
	case "debug":
		return slog.LevelDebug
	case "info":
		return slog.LevelInfo
	case "warn":
		return slog.LevelWarn
	case "error":
		return slog.LevelError
	default:
		return slog.LevelInfo
	}
}
```

## NATS Example
The library includes a complete example for load testing NATS messaging systems. Here's how to use it:

**Configuration**

First, configure the test parameters through environment variables:
```
# NATS server settings
export NATS_HOST=127.0.0.1
export NATS_PORT=4222

# Load test parameters
export LOAD_TEST_DURATION=1m
export LOAD_TEST_CONCURRENCY=50
export LOAD_TEST_WARMUP=5s
export LOAD_TEST_REPORT_INTERVAL=5s
export LOAD_TEST_LOG_LEVEL=info
export LOAD_TEST_OUTPUT_PATH=./results/load-test-results.json
export LOAD_TEST_TAGS="env=dev,test=nats-publish"
export LOAD_TEST_TYPE=publish
export LOAD_TEST_SUBJECT=load.test
export LOAD_TEST_MESSAGE_SIZE=1024
```

**Starting a NATS Server**

For convenience, you can use Docker to start a NATS server:
```bash
make docker/start
```

**Running Tests**

The example includes several predefined test scenarios:
```
# Basic publish test
make run/publish

# Basic subscribe test
make run/subscribe

# Combined publish/subscribe test
make run/combined

# High concurrency test (100 clients)
make run/high-concurrency

# Long duration test (5 minutes)
make run/long-duration
```

**Sample Output**

During the test, you'll see real-time progress in the console:
```
{"time":"2025-03-22T11:43:38.81066307+02:00","level":"INFO","msg":"Starting NATS load test","test_type":"publish","subject":"load.test","concurrency":50,"duration":"1m0s"}
{"time":"2025-03-22T11:43:38.810722961+02:00","level":"INFO","msg":"Starting load test","duration":"1m0s","concurrency":50,"runners":1,"collectors":1,"reporters":2}
{"time":"2025-03-22T11:43:38.810740728+02:00","level":"INFO","msg":"Starting metrics collector","collector":"Composite Metrics Collector"}
{"time":"2025-03-22T11:43:38.810746339+02:00","level":"INFO","msg":"Starting system metrics collection","interval":"1s"}
{"time":"2025-03-22T11:43:38.810754778+02:00","level":"INFO","msg":"Setting up runner","runner":"NATS Publish Runner"}
{"time":"2025-03-22T11:43:38.812004451+02:00","level":"INFO","msg":"NatsPublishRunner setup complete","subject":"load.test","messageSize":1024}
{"time":"2025-03-22T11:43:38.812030943+02:00","level":"INFO","msg":"Starting warmup period","duration":"5s"}
{"time":"2025-03-22T11:43:43.812293523+02:00","level":"INFO","msg":"Warmup period completed"}
[0.0s] Rate: 2503253 ops/s | Total: 12516265 ops | Errors: 0 | Current: 2502770.51 ops/s
[5.0s] Rate: 2336390 ops/s | Total: 24198216 ops | Errors: 0 | Current: 2419690.20 ops/s
[10.0s] Rate: 2569334 ops/s | Total: 37044887 ops | Errors: 0 | Current: 2469636.73 ops/s
[15.0s] Rate: 2059073 ops/s | Total: 47340253 ops | Errors: 0 | Current: 2366934.32 ops/s
[20.0s] Rate: 2531616 ops/s | Total: 59998333 ops | Errors: 0 | Current: 2399910.62 ops/s
[25.4s] Rate: 2688061 ops/s | Total: 73438638 ops | Errors: 0 | Current: 2414638.84 ops/s
[30.0s] Rate: 2504095 ops/s | Total: 85959115 ops | Errors: 0 | Current: 2455948.05 ops/s
....
```

After completion, the JSON report will contain detailed metrics:
```
{
  "start_time": "2025-03-22T11:43:43+02:00",
  "end_time": "2025-03-22T11:44:38+02:00",
  "test_duration_seconds": 54.998952413,
  "total_operations": 136088798,
  "error_count": 0,
  "error_rate_percent": 0,
  "throughput_ops_per_sec": 2474388.911593759,
  "latency_p50_ms": 0.000083,
  "latency_p90_ms": 0.000095,
  "latency_p95_ms": 0.00015900000000000002,
  "latency_p99_ms": 0.01027603000000119,
  "latency_min_ms": 0.00006599999999999999,
  "latency_max_ms": 71.073603,
  "latency_mean_ms": 0.018006328602729986,
  "cpu_usage_percent": 0,
  "memory_usage_mb": 542.8718335469564,
  "active_goroutines": 5,
  "gc_pause_ms": 0.44637781666666676
}
```

## Creating Custom Runners
To create a custom runner for your specific application, implement the `core.Runner` interface:
```
type MyCustomRunner struct {
    // Your runner state
    logger *slog.Logger
}

func (r *MyCustomRunner) Setup(ctx context.Context) error {
    // Initialize resources
    return nil
}

func (r *MyCustomRunner) Run(ctx context.Context) error {
    // Execute a single operation
    // (This will be called repeatedly for the test duration)
    return nil
}

func (r *MyCustomRunner) Teardown(ctx context.Context) error {
    // Clean up resources
    return nil
}

func (r *MyCustomRunner) Name() string {
    return "My Custom Runner"
}
```

## Working with Metrics Collectors
The framework includes a `SystemCollector` for monitoring system resources. You can also implement custom collectors:
```
type MyCustomCollector struct {
    metrics *core.Metrics
    logger  *slog.Logger
}

func (c *MyCustomCollector) Start() error {
    // Start collecting metrics
    return nil
}

func (c *MyCustomCollector) Stop() error {
    // Stop collecting metrics
    return nil
}

func (c *MyCustomCollector) GetMetrics() *core.Metrics {
    // Return collected metrics
    return c.metrics
}

func (c *MyCustomCollector) Name() string {
    return "My Custom Collector"
}
```

## Custom Reporters
Implement the `core.Reporter` interface to create custom result formats:
```
type MyCustomReporter struct {
    // Reporter state
}

func (r *MyCustomReporter) ReportProgress(snapshot *core.MetricsSnapshot) error {
    // Report progress during the test
    return nil
}

func (r *MyCustomReporter) ReportResults(metrics *core.Metrics) error {
    // Generate final report
    return nil
}

func (r *MyCustomReporter) Name() string {
    return "My Custom Reporter"
}
```
