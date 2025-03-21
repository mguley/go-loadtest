package pkg

import (
	"context"
	"fmt"
	"log/slog"
	"sync"
	"time"

	"github.com/mguley/go-loadtest/pkg/collector"
	"github.com/mguley/go-loadtest/pkg/core"
)

// Orchestrator coordinates the execution of load tests.
// It sets up the runners, collectors, and reporters, and manages the test lifecycle.
//
// Fields:
//   - config:     Pointer to core.TestConfig containing test configuration parameters.
//   - runners:    Slice of core.Runner used to execute test operations.
//   - collectors: Slice of core.MetricsCollector used to gather metrics during the test.
//   - reporters:  Slice of core.Reporter used for progress and final result reporting.
//   - logger:     Pointer to slog.Logger used for logging events.
type Orchestrator struct {
	config     *core.TestConfig
	runners    []core.Runner
	collectors []core.MetricsCollector
	reporters  []core.Reporter
	logger     *slog.Logger
}

// NewOrchestrator creates a new test orchestrator with the provided configuration and logger.
//
// Parameters:
//   - config: Pointer to core.TestConfig containing load test settings.
//   - logger: Pointer to slog.Logger for logging events.
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

// Run executes the load test managed by the Orchestrator.
// It sets up the collectors, runners, and progress reporting, runs the test operations,
// then cleans up and collects the final results.
//
// Returns:
//   - error: An error if any stage of test execution fails; otherwise nil.
func (o *Orchestrator) Run() error {
	if len(o.runners) == 0 {
		return fmt.Errorf("no test runners configured")
	}

	// Create a context that automatically cancels when the test duration elapses
	ctx, cancel := context.WithTimeout(context.Background(), o.config.TestDuration)
	defer cancel()

	o.logger.Info("Starting load test",
		"duration", o.config.TestDuration.String(),
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

	// Create the base metrics container and record the start time.
	metrics := core.NewMetrics()
	metrics.StartTime = time.Now()

	// Start background progress reporting.
	progressCancel, progressWg := o.startProgressReporting(o.config.ReportInterval, metrics)

	// Run the main test operations.
	o.runOperations(ctx, metrics)

	metrics.EndTime = time.Now()
	progressCancel()
	progressWg.Wait()

	// Clean up runners and collectors.
	o.cleanup(ctx)
	// Merge any additional metrics from collectors and calculate final throughput.
	o.collectData(metrics)

	return nil
}

// startCollectors starts all registered metrics collectors.
//
// Returns:
//   - error: An error if any collector fails to start; otherwise nil.
func (o *Orchestrator) startCollectors() error {
	for _, item := range o.collectors {
		o.logger.Info("Starting metrics collector", "collector", item.Name())
		if err := item.Start(); err != nil {
			return fmt.Errorf("failed to start collector %s: %w", item.Name(), err)
		}
	}
	return nil
}

// setupRunners prepares each test runner for execution.
//
// Parameters:
//   - ctx: The context used for managing runner setup.
//
// Returns:
//   - error: An error if any runner fails to set up; otherwise nil.
func (o *Orchestrator) setupRunners(ctx context.Context) error {
	for _, runner := range o.runners {
		o.logger.Info("Setting up runner", "runner", runner.Name())
		if err := runner.Setup(ctx); err != nil {
			return fmt.Errorf("failed to setup runner %s: %w", runner.Name(), err)
		}
	}
	return nil
}

// warmup executes a warmup period if WarmupDuration is set.
//
// Returns:
//   - error: Nil unless the warmup is canceled by the context.
func (o *Orchestrator) warmup() error {
	if o.config.WarmupDuration <= 0 {
		return nil
	}

	o.logger.Info("Starting warmup period", "duration", o.config.WarmupDuration.String())
	warmupCtx, warmupCancel := context.WithTimeout(context.Background(), o.config.WarmupDuration)
	defer warmupCancel()

	// Run warmup operations without collecting metrics.
	o.runOperations(warmupCtx, nil)
	o.logger.Info("Warmup period completed")

	return nil
}

// startProgressReporting spawns a goroutine that periodically collects and reports progress.
//
// Parameters:
//   - interval: Duration between progress reports.
//   - metrics:  Base metrics container to merge live metrics into.
//
// Returns:
//   - context.CancelFunc: Function to cancel the progress reporting.
//   - *sync.WaitGroup:    WaitGroup that signals when progress reporting has ended.
func (o *Orchestrator) startProgressReporting(
	interval time.Duration,
	metrics *core.Metrics,
) (context.CancelFunc, *sync.WaitGroup) {
	progressCtx, progressCancel := context.WithCancel(context.Background())
	var wg sync.WaitGroup

	wg.Add(1)
	go func() {
		defer wg.Done()
		o.reportProgress(progressCtx, interval, metrics)
	}()

	return progressCancel, &wg
}

// cleanup stops all collectors and tears down all runners.
//
// Parameters:
//   - ctx: The context used to manage cleanup operations.
func (o *Orchestrator) cleanup(ctx context.Context) {
	// Stop all metrics collectors.
	for _, item := range o.collectors {
		o.logger.Info("Stopping metrics collector", "collector", item.Name())
		if err := item.Stop(); err != nil {
			o.logger.Error("Failed to stop collector", "collector", item.Name(), "error", err.Error())
		}
	}

	// Teardown all runners.
	for _, runner := range o.runners {
		o.logger.Info("Tearing down runner", "runner", runner.Name())
		if err := runner.Teardown(ctx); err != nil {
			o.logger.Error("Failed to teardown runner", "runner", runner.Name(), "error", err.Error())
		}
	}
}

// collectData merges metrics from all collectors, calculates throughput,
// and reports the final results using all configured reporters.
//
// Parameters:
//   - metrics: Pointer to core.Metrics containing test results.
func (o *Orchestrator) collectData(metrics *core.Metrics) {
	// Merge metrics from each collector.
	for _, item := range o.collectors {
		collectorMetrics := item.GetMetrics()
		if collectorMetrics != nil {
			metrics.Merge(collectorMetrics)
		}
	}

	// Calculate throughput if the test duration is positive.
	if duration := metrics.EndTime.Sub(metrics.StartTime).Seconds(); duration > 0 {
		metrics.Throughput = float64(metrics.TotalOperations) / duration
	}

	// Generate final reports using all registered reporters.
	for _, reporter := range o.reporters {
		o.logger.Info("Generating final report", "reporter", reporter.Name())
		if err := reporter.ReportResults(metrics); err != nil {
			o.logger.Error("Failed to report results", "reporter", reporter.Name(), "error", err.Error())
		}
	}

	// Format the test duration for final logging.
	d := metrics.EndTime.Sub(metrics.StartTime)
	formatted := fmt.Sprintf("%d minutes %d seconds", int(d.Minutes()), int(d.Seconds())%60)
	o.logger.Info("Load test completed successfully",
		"duration", formatted,
		"operations", metrics.TotalOperations,
		"errors", metrics.ErrorCount,
		"throughput", fmt.Sprintf("%.0f ops/s", metrics.Throughput))
}

// runOperations executes the test operations using the configured runners.
// It spawns worker goroutines per runner based on the configured concurrency level.
//
// Parameters:
//   - ctx:     Context governing test operation execution.
//   - metrics: Pointer to core.Metrics for recording test results; if nil, metrics recording is skipped.
func (o *Orchestrator) runOperations(ctx context.Context, metrics *core.Metrics) {
	var wg sync.WaitGroup

	// Start worker goroutines for each runner.
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
						// Execute the test operation and record its latency.
						start := time.Now()
						err := runner.Run(ctx)
						latency := time.Since(start).Seconds() * 1_000 // milliseconds

						// Update metrics if provided.
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

	// Wait for the context to be canceled (i.e. test duration elapsed) then wait for all workers to finish.
	<-ctx.Done()
	wg.Wait()
}

// reportProgress periodically collects and reports metrics during the test.
// It uses the provided base metrics container to merge live metrics.
//
// Parameters:
//   - ctx:      Context for canceling progress reporting.
//   - interval: Duration between progress reports.
//   - metrics:  Base metrics container to merge live metrics.
func (o *Orchestrator) reportProgress(ctx context.Context, interval time.Duration, metrics *core.Metrics) {
	ticker := time.NewTicker(interval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			snapshot := o.collectMetricsSnapshot(metrics)
			for _, reporter := range o.reporters {
				if err := reporter.ReportProgress(snapshot); err != nil {
					o.logger.Error("Failed to report progress", "reporter", reporter.Name(), "error", err.Error())
				}
			}
		}
	}
}

// collectMetricsSnapshot gathers current metrics from all collectors and merges them into the provided baseMetrics.
// If there is exactly one registered collector, and it is a CompositeCollector, its snapshot is returned directly.
//
// Parameters:
//   - baseMetrics: Pointer to core.Metrics to use as the accumulator for merging.
//
// Returns:
//   - *core.MetricsSnapshot: A snapshot of the current aggregated metrics.
func (o *Orchestrator) collectMetricsSnapshot(baseMetrics *core.Metrics) *core.MetricsSnapshot {
	// If no collectors are registered, return an empty snapshot.
	if len(o.collectors) == 0 {
		return &core.MetricsSnapshot{
			Timestamp: time.Now(),
			Custom:    make(map[string]float64),
		}
	}

	// If there is exactly one collector, and it is a CompositeCollector, return its snapshot.
	if len(o.collectors) == 1 {
		if composite, ok := o.collectors[0].(*collector.CompositeCollector); ok {
			if metrics := composite.GetMetrics(); metrics != nil {
				baseMetrics.Merge(metrics)
				return baseMetrics.GetSnapshot()
			}
		}
	}

	// Otherwise, merge metrics from all collectors.
	mergedMetrics := core.NewMetrics()
	for _, item := range o.collectors {
		if metrics := item.GetMetrics(); metrics != nil {
			mergedMetrics.Merge(metrics)
		}
	}

	baseMetrics.Merge(mergedMetrics)
	return baseMetrics.GetSnapshot()
}
