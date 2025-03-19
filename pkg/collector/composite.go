package collector

import (
	"sync"

	"github.com/mguley/go-loadtest/pkg/core"
)

// CompositeCollector aggregates multiple metrics collectors into a single composite collector.
//
// Fields:
//   - collectors: A slice of core.MetricsCollector to be aggregated.
//   - mu:         A sync.Mutex to protect concurrent access to the collectors.
type CompositeCollector struct {
	collectors []core.MetricsCollector
	mu         sync.Mutex
}

// NewCompositeCollector creates a new composite collector.
//
// Parameters:
//   - collectors: Variadic list of core.MetricsCollector to initialize the composite collector.
//
// Returns:
//   - *CompositeCollector: A pointer to a newly created CompositeCollector.
func NewCompositeCollector(collectors ...core.MetricsCollector) *CompositeCollector {
	return &CompositeCollector{
		collectors: collectors,
	}
}

// AddCollector adds a new collector to the composite.
//
// Parameters:
//   - collector: The core.MetricsCollector to be added.
func (c *CompositeCollector) AddCollector(collector core.MetricsCollector) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.collectors = append(c.collectors, collector)
}

// Start begins metrics collection for all child collectors within the composite.
//
// Returns:
//   - error: An error if any of the child collectors fails to start, otherwise nil.
func (c *CompositeCollector) Start() error {
	c.mu.Lock()
	defer c.mu.Unlock()

	for _, collector := range c.collectors {
		if err := collector.Start(); err != nil {
			return err
		}
	}

	return nil
}

// Stop ends metrics collection for all child collectors within the composite.
//
// Returns:
//   - error: The last encountered error during the stop process if any, otherwise nil.
func (c *CompositeCollector) Stop() error {
	c.mu.Lock()
	defer c.mu.Unlock()

	var lastErr error
	for _, collector := range c.collectors {
		if err := collector.Stop(); err != nil {
			lastErr = err
		}
	}

	return lastErr
}

// GetMetrics retrieves and combines metrics from all child collectors.
//
// Returns:
//   - *core.Metrics: A pointer to the aggregated metrics from all collectors.
func (c *CompositeCollector) GetMetrics() *core.Metrics {
	c.mu.Lock()
	defer c.mu.Unlock()

	combined := core.NewMetrics()

	for _, collector := range c.collectors {
		if metrics := collector.GetMetrics(); metrics != nil {
			combined.Merge(metrics)
		}
	}

	return combined
}

// Name returns the identifier name of this composite collector.
//
// Returns:
//   - string: The name "Composite Metrics Collector".
func (c *CompositeCollector) Name() string {
	return "Composite Metrics Collector"
}
