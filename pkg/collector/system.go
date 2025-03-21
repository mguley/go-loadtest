package collector

import (
	"context"
	"log/slog"
	"runtime"
	"sync"
	"time"

	"github.com/mguley/go-loadtest/pkg/core"
	"github.com/mguley/go-loadtest/pkg/util"
)

// SystemCollector collects system-level metrics during load tests.
//
// Fields:
//   - logger:        A pointer to slog.Logger used for logging events.
//   - metrics:       A pointer to core.Metrics to store collected metrics.
//   - ctx:           Context for managing the lifecycle of metrics collection.
//   - cancel:        Function to cancel the context.
//   - interval:      Time duration between consecutive metrics collection intervals.
//   - cpuStats:      A util.Float64Data slice for CPU usage data.
//   - memStats:      A util.Float64Data slice for memory usage data.
//   - goroutineData: A slice of int storing the count of active goroutines.
//   - gcStats:       A util.Float64Data slice for GC pause duration data.
//   - mu:            A sync.Mutex to protect concurrent access to metrics data.
//   - wg:            A sync.WaitGroup to manage the collection goroutine.
type SystemCollector struct {
	logger        *slog.Logger
	metrics       *core.Metrics
	ctx           context.Context
	cancel        context.CancelFunc
	interval      time.Duration
	cpuStats      util.Float64Data
	memStats      util.Float64Data
	goroutineData []int
	gcStats       util.Float64Data
	mu            sync.Mutex
	wg            sync.WaitGroup
}

// NewSystemCollector creates a new system resource metrics collector.
//
// Parameters:
//   - logger: A pointer to slog.Logger used for logging events and metrics.
//
// Returns:
//   - *SystemCollector: A pointer to an instantiated SystemCollector with default settings.
func NewSystemCollector(logger *slog.Logger) *SystemCollector {
	ctx, cancel := context.WithCancel(context.Background())

	return &SystemCollector{
		logger:        logger,
		metrics:       core.NewMetrics(),
		ctx:           ctx,
		cancel:        cancel,
		interval:      time.Second,
		cpuStats:      make(util.Float64Data, 0),
		memStats:      make(util.Float64Data, 0),
		goroutineData: make([]int, 0),
		gcStats:       make(util.Float64Data, 0),
	}
}

// Start begins collecting system metrics at regular intervals.
//
// Returns:
//   - error: An error if metrics collection fails to start, otherwise nil.
func (c *SystemCollector) Start() error {
	c.logger.Info("Starting system metrics collection", "interval", c.interval.String())

	c.wg.Add(1)
	go c.collectMetrics()

	return nil
}

// Stop ends the metrics collection and finalizes the collected metrics.
//
// Returns:
//   - error: An error if encountered during stopping the collection, otherwise nil.
func (c *SystemCollector) Stop() error {
	c.logger.Info("Stopping system metrics collection")
	c.cancel()
	c.wg.Wait()

	// Calculate final metrics
	c.mu.Lock()
	defer c.mu.Unlock()

	if len(c.cpuStats) > 0 {
		c.metrics.ResourceMetrics.CPUUsagePercent = c.cpuStats.Mean()
	}
	if len(c.memStats) > 0 {
		c.metrics.ResourceMetrics.MemoryUsageMB = c.memStats.Mean()
	}
	if len(c.goroutineData) > 0 {
		c.metrics.ResourceMetrics.ActiveGoroutines = c.goroutineData[len(c.goroutineData)-1]
	}
	if len(c.gcStats) > 0 {
		c.metrics.ResourceMetrics.GCPauseMs = c.gcStats.Mean()
	}

	return nil
}

// GetMetrics retrieves the current system metrics.
//
// Returns:
//   - *core.Metrics: A pointer to the collected system metrics.
func (c *SystemCollector) GetMetrics() *core.Metrics {
	c.mu.Lock()
	defer c.mu.Unlock()

	return c.metrics
}

// Name returns the identifier name of this system collector.
//
// Returns:
//   - string: The name "System Resource Collector".
func (c *SystemCollector) Name() string {
	return "System Resource Collector"
}

// SetInterval updates the metrics collection interval.
//
// Parameters:
//   - interval: A time.Duration value representing the new collection interval.
//     Must be a positive duration.
func (c *SystemCollector) SetInterval(interval time.Duration) {
	if interval > 0 {
		c.interval = interval
	}
}

// collectMetrics continuously gathers system metrics at the specified interval.
func (c *SystemCollector) collectMetrics() {
	defer c.wg.Done()

	ticker := time.NewTicker(c.interval)
	defer ticker.Stop()

	for {
		select {
		case <-c.ctx.Done():
			return
		case <-ticker.C:
			runtime.GC()
			var m runtime.MemStats
			runtime.ReadMemStats(&m)

			c.mu.Lock()

			// Collect memory usage in MB
			c.memStats = append(c.memStats, float64(m.Alloc)/(1024*1024))

			// Collect goroutine count
			c.goroutineData = append(c.goroutineData, runtime.NumGoroutine())

			// Collect GC stats (most recent pause in ms)
			if m.NumGC > 0 {
				gcPauseMs := float64(m.PauseNs[(m.NumGC-1)%256]) / 1e6
				c.gcStats = append(c.gcStats, gcPauseMs)
			}

			c.mu.Unlock()
		}
	}
}
