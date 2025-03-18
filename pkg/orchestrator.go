package pkg

import (
	"context"
	"fmt"
	"log/slog"
	"sync"
	"time"

	"github.com/mguley/go-loadtest/pkg/core"
)

// Orchestrator coordinates the execution of load tests.
//
// Fields:
//   - config:     A pointer to core.TestConfig containing test configuration parameters.
//   - runners:    A slice of core.Runner that execute the test operations.
//   - collectors: A slice of core.MetricsCollector that gather metrics during the test.
//   - reporters:  A slice of core.Reporter for reporting progress and final results.
//   - logger:     A pointer to slog.Logger used for logging events.
type Orchestrator struct {
	config     *core.TestConfig
	runners    []core.Runner
	collectors []core.MetricsCollector
	reporters  []core.Reporter
	logger     *slog.Logger
}

// NewOrchestrator creates a new test orchestrator with the provided configuration.
//
// Parameters:
//   - config: A pointer to core.TestConfig containing load test settings.
//   - logger: A pointer to slog.Logger for logging events.
//
// Returns:
//   - *Orchestrator: A pointer to a newly created Orchestrator instance.
func NewOrchestrator(config *core.TestConfig, logger *slog.Logger) *Orchestrator {
	return &Orchestrator{
		config:     config,
		runners:    make([]core.Runner, 0),
		collectors: make([]core.MetricsCollector, 0),
		reporters:  make([]core.Reporter, 0),
		logger:     logger,
	}
}

// AddRunner adds a test runner to the orchestrator.
//
// Parameters:
//   - runner: A core.Runner instance to be added.
func (o *Orchestrator) AddRunner(runner core.Runner) {
	o.runners = append(o.runners, runner)
}

// AddCollector adds a metrics collector to the orchestrator.
//
// Parameters:
//   - collector: A core.MetricsCollector instance to be added.
func (o *Orchestrator) AddCollector(collector core.MetricsCollector) {
	o.collectors = append(o.collectors, collector)
}

// AddReporter adds a reporter to the orchestrator.
//
// Parameters:
//   - reporter: A core.Reporter instance to be added.
func (o *Orchestrator) AddReporter(reporter core.Reporter) {
	o.reporters = append(o.reporters, reporter)
}

// Run executes the load test orchestrated by the Orchestrator.
//
// Returns:
//   - error: An error if any stage of test execution fails, otherwise nil.
func (o *Orchestrator) Run() error {
	if len(o.runners) == 0 {
		return fmt.Errorf("no test runners configured")
	}

	ctx, cancel := context.WithTimeout(context.Background(), o.config.TestDuration)
	defer cancel()

	o.logger.Info("Starting load test",
		"duration", o.config.TestDuration,
		"concurrency", o.config.Concurrency,
		"runners", len(o.runners),
		"collectors", len(o.collectors),
		"reporters", len(o.reporters))

	if err := o.startCollectors(); err != nil {
		return err
	}
	if err := o.setupRunners(ctx); err != nil {
		return err
	}
	if err := o.warmup(); err != nil {
		return err
	}

	// Create metrics for the test
	metrics := core.NewMetrics()
	metrics.StartTime = time.Now()

	// Start progress reporting in background
	progressCancel, progressWg := o.startProgressReporting(o.config.ReportInterval)

	// Run the main test operations
	o.runOperations(ctx, metrics)

	metrics.EndTime = time.Now()
	progressCancel()
	progressWg.Wait()

	// Clean up
	o.cleanup(ctx)
	o.collectData(metrics)

	return nil
}

// startCollectors starts all configured metrics collectors.
//
// Returns:
//   - error: An error if any collector fails to start, otherwise nil.
func (o *Orchestrator) startCollectors() error {
	for _, collector := range o.collectors {
		o.logger.Info("Starting metrics collector", "collector", collector.Name())
		if err := collector.Start(); err != nil {
			return fmt.Errorf("failed to start collector %s: %w", collector.Name(), err)
		}
	}
	return nil
}

// setupRunners prepares each runner before the test starts.
//
// Parameters:
//   - ctx: The context for managing runner setup.
//
// Returns:
//   - error: An error if any runner fails to set up, otherwise nil.
func (o *Orchestrator) setupRunners(ctx context.Context) error {
	for _, runner := range o.runners {
		o.logger.Info("Setting up runner", "runner", runner.Name())
		if err := runner.Setup(ctx); err != nil {
			return fmt.Errorf("failed to setup runner %s: %w", runner.Name(), err)
		}
	}
	return nil
}

// warmup applies a warmup period if WarmupDuration is set.
//
// Returns:
//   - error: Always nil unless warmup is canceled, in which case context error is returned.
func (o *Orchestrator) warmup() error {
	if o.config.WarmupDuration <= 0 {
		return nil
	}

	o.logger.Info("Starting warmup period", "duration", o.config.WarmupDuration)
	warmupCtx, warmupCancel := context.WithTimeout(context.Background(), o.config.WarmupDuration)
	defer warmupCancel()

	// Run warmup operations without collecting metrics
	o.runOperations(warmupCtx, nil)
	o.logger.Info("Warmup period completed")

	return warmupCtx.Err()
}

// startProgressReporting spawns a goroutine that periodically reports progress.
//
// Parameters:
//   - interval: The duration between progress reports.
//
// Returns:
//   - context.CancelFunc: A cancel function to stop progress reporting.
//   - *sync.WaitGroup: A wait group that completes when progress reporting ends.
func (o *Orchestrator) startProgressReporting(interval time.Duration) (context.CancelFunc, *sync.WaitGroup) {
	progressCtx, progressCancel := context.WithCancel(context.Background())
	var wg sync.WaitGroup

	wg.Add(1)
	go func() {
		defer wg.Done()
		o.reportProgress(progressCtx, interval)
	}()

	return progressCancel, &wg
}

// cleanup stops all collectors and tears down all runners.
//
// Parameters:
//   - ctx: The context for managing cleanup operations.
func (o *Orchestrator) cleanup(ctx context.Context) {
	// Stop all configured metrics collectors.
	for _, collector := range o.collectors {
		o.logger.Info("Stopping metrics collector", "collector", collector.Name())
		if err := collector.Stop(); err != nil {
			o.logger.Error("Failed to stop collector", "collector", collector.Name(), "error", err.Error())
		}
	}

	// Teardown all configured runners.
	for _, runner := range o.runners {
		o.logger.Info("Tearing down runner", "runner", runner.Name())
		if err := runner.Teardown(ctx); err != nil {
			o.logger.Error("Failed to teardown runner", "runner", runner.Name(), "error", err.Error())
		}
	}
}

// collectData merges metrics from all collectors, calculates throughput,
// and reports final results using all configured reporters.
//
// Parameters:
//   - metrics: A pointer to core.Metrics containing test results to update.
func (o *Orchestrator) collectData(metrics *core.Metrics) {
	// Merge metrics from all collectors
	for _, collector := range o.collectors {
		collectorMetrics := collector.GetMetrics()
		if collectorMetrics != nil {
			metrics.Merge(collectorMetrics)
		}
	}

	// Calculate throughput
	if duration := metrics.EndTime.Sub(metrics.StartTime).Seconds(); duration > 0 {
		metrics.Throughput = float64(metrics.TotalOperations) / duration
	}

	// Report final results through all configured reporters
	for _, reporter := range o.reporters {
		o.logger.Info("Generating final report", "reporter", reporter.Name())
		if err := reporter.ReportResults(metrics); err != nil {
			o.logger.Error("Failed to report results", "reporter", reporter.Name(), "error", err.Error())
		}
	}

	o.logger.Info("Load test completed successfully",
		"duration", metrics.EndTime.Sub(metrics.StartTime),
		"operations", metrics.TotalOperations,
		"errors", metrics.ErrorCount,
		"throughput", metrics.Throughput)
}

// runOperations executes the test operations using the configured runners.
//
// Parameters:
//   - ctx:     The context governing test operation execution.
//   - metrics: A pointer to core.Metrics for recording test results; if nil, metrics recording is skipped.
func (o *Orchestrator) runOperations(ctx context.Context, metrics *core.Metrics) {
	var wg sync.WaitGroup

	// Start worker goroutines for each runner
	for _, runner := range o.runners {
		for i := 0; i < o.config.Concurrency; i++ {
			wg.Add(1)
			go func(runner core.Runner, workerId int) {
				defer wg.Done()
				defer func() {
					if r := recover(); r != nil {
						o.logger.Warn("Recovered in runner runOperations", "runner", runner.Name())
					}
				}()

				o.logger.Debug("Starting worker", "runner", runner.Name(), "worker_id", workerId)
				for {
					select {
					case <-ctx.Done():
						o.logger.Debug("Worker stopping due to context done",
							"runner", runner.Name(),
							"worker_id", workerId)
						return
					default:
						// Execute the test operation
						start := time.Now()
						err := runner.Run(ctx)
						latency := time.Since(start).Seconds() * 1_000 // convert to milliseconds

						// Record metrics if provided
						if metrics != nil {
							switch {
							case err != nil:
								metrics.IncrementErrors()
							default:
								metrics.IncrementOperations()
								metrics.AddLatency(latency)
							}
						}
					}
				}
			}(runner, i)
		}
	}

	<-ctx.Done()
	wg.Wait()
}

// reportProgress periodically collects and reports metrics during the test.
//
// Parameters:
//   - ctx:      The context for canceling progress reporting.
//   - interval: The duration between progress reports.
func (o *Orchestrator) reportProgress(ctx context.Context, interval time.Duration) {
	ticker := time.NewTicker(interval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			snapshot := o.collectMetricsSnapshot()
			for _, reporter := range o.reporters {
				if err := reporter.ReportProgress(snapshot); err != nil {
					o.logger.Error("Failed to report progress", "reporter", reporter.Name(), "error", err.Error())
				}
			}
		}
	}
}

// collectMetricsSnapshot gathers current metrics from all collectors.
//
// Returns:
//   - *core.MetricsSnapshot: A pointer to a snapshot of the current metrics.
func (o *Orchestrator) collectMetricsSnapshot() *core.MetricsSnapshot {
	var combinedMetrics *core.Metrics

	// Get metrics from all collectors
	for _, collector := range o.collectors {
		if metrics := collector.GetMetrics(); metrics != nil {
			switch {
			case combinedMetrics == nil:
				combinedMetrics = metrics
			default:
				combinedMetrics.Merge(metrics)
			}
		}
	}

	// If no metrics are available, create an empty snapshot
	if combinedMetrics == nil {
		return &core.MetricsSnapshot{
			Timestamp: time.Now(),
			Custom:    make(map[string]float64),
		}
	}

	return combinedMetrics.GetSnapshot()
}
