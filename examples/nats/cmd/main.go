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

// main is the entry point of the NATS load test example.
//
// It initializes the configuration, creates the necessary components,
// and executes the load test. The process exits with a non-zero status if the test fails.
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
		jsonReporter.SetIncludeLatencies(true)
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
