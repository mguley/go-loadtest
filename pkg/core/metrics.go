package core

import (
	"sync"
	"sync/atomic"
	"time"
)

// MetricsSnapshot represents a point-in-time snapshot of metrics during test execution.
//
// Fields:
//   - Timestamp:     The time when the snapshot was taken.
//   - Operations:    The number of operations completed so far.
//   - Errors:        The count of errors encountered so far.
//   - RatePerSecond: The calculated throughput (operations per second).
//   - Custom:        Custom metrics provided by specific test implementations.
type MetricsSnapshot struct {
	Timestamp     time.Time
	Operations    int64
	Errors        int64
	RatePerSecond float64
	Custom        map[string]float64
}

// Metrics holds the complete metrics collected during a load test.
//
// Fields:
//   - StartTime:       The time when the test began.
//   - EndTime:         The time when the test completed.
//   - TotalOperations: The total number of operations executed during the test.
//   - ErrorCount:      The total number of errors encountered.
//   - Latencies:       A slice of latency measurements (in milliseconds) for each operation.
//   - Throughput:      The calculated operations per second over the test duration.
//   - ResourceMetrics: System resource usage data collected during the test.
//   - Custom:          A map for additional, test-specific metrics.
//   - mu:              A mutex to protect concurrent access to slices and maps.
type Metrics struct {
	StartTime       time.Time
	EndTime         time.Time
	TotalOperations int64
	ErrorCount      int64
	Latencies       []float64
	Throughput      float64
	ResourceMetrics ResourceMetrics
	Custom          map[string]float64
	mu              sync.Mutex
}

// ResourceMetrics contains data about system resource utilization during the test.
//
// Fields:
//   - CPUUsagePercent:  The average CPU utilization percentage.
//   - MemoryUsageMB:    The average memory usage in megabytes.
//   - ActiveGoroutines: The number of active goroutines at test completion.
//   - GCPauseMs:        The average garbage collection pause time in milliseconds.
type ResourceMetrics struct {
	CPUUsagePercent  float64
	MemoryUsageMB    float64
	ActiveGoroutines int
	GCPauseMs        float64
}

// NewMetrics creates an empty metrics container.
//
// Returns:
//   - *Metrics: A pointer to an empty Metrics instance.
func NewMetrics() *Metrics {
	return &Metrics{
		Latencies: make([]float64, 0),
		Custom:    make(map[string]float64),
	}
}

// AddLatency safely adds a latency measurement to the metrics.
//
// Parameters:
//   - latencyMs: A float64 representing the latency in milliseconds to add.
func (m *Metrics) AddLatency(latencyMs float64) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.Latencies = append(m.Latencies, latencyMs)
}

// IncrementOperations atomically increments the operation counter.
func (m *Metrics) IncrementOperations() {
	atomic.AddInt64(&m.TotalOperations, 1)
}

// IncrementErrors atomically increments the error counter.
func (m *Metrics) IncrementErrors() {
	atomic.AddInt64(&m.ErrorCount, 1)
}

// SetCustomMetric safely sets a custom metric value.
//
// Parameters:
//   - name:  The name of the custom metric.
//   - value: The value to assign to the custom metric.
func (m *Metrics) SetCustomMetric(name string, value float64) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.Custom[name] = value
}

// GetSnapshot returns a point-in-time snapshot of current metrics.
//
// Returns:
//   - *MetricsSnapshot: A pointer to a snapshot containing operations, errors, throughput, and custom metrics.
func (m *Metrics) GetSnapshot() *MetricsSnapshot {
	operations := atomic.LoadInt64(&m.TotalOperations)
	errors := atomic.LoadInt64(&m.ErrorCount)
	now := time.Now()

	var elapsedSeconds float64 = 1 // Default to 1 to avoid division by zero
	if !m.StartTime.IsZero() {
		elapsedSeconds = now.Sub(m.StartTime).Seconds()
		if elapsedSeconds <= 0 {
			elapsedSeconds = 1
		}
	}

	ratePerSecond := float64(operations) / elapsedSeconds

	// Create a snapshot of custom metrics
	m.mu.Lock()
	customCopy := make(map[string]float64, len(m.Custom))
	for k, v := range m.Custom {
		customCopy[k] = v
	}
	m.mu.Unlock()

	return &MetricsSnapshot{
		Timestamp:     now,
		Operations:    operations,
		Errors:        errors,
		RatePerSecond: ratePerSecond,
		Custom:        customCopy,
	}
}

// Merge combines metrics from another Metrics instance into the current one.
//
// Parameters:
//   - other: A pointer to another Metrics instance to merge.
func (m *Metrics) Merge(other *Metrics) {
	// If the other has a non-zero start time, and it's earlier than the current start time.
	if !other.StartTime.IsZero() {
		if m.StartTime.IsZero() || other.StartTime.Before(m.StartTime) {
			m.StartTime = other.StartTime
		}
	}

	// If the other has a non-zero end time, and it's later than the current end time.
	if !other.EndTime.IsZero() && other.EndTime.After(m.EndTime) {
		m.EndTime = other.EndTime
	}

	// Add counters
	atomic.AddInt64(&m.TotalOperations, atomic.LoadInt64(&other.TotalOperations))
	atomic.AddInt64(&m.ErrorCount, atomic.LoadInt64(&other.ErrorCount))

	// Merge latencies and custom metrics
	m.mu.Lock()
	defer m.mu.Unlock()

	other.mu.Lock()
	defer other.mu.Unlock()

	// Append latencies
	m.Latencies = append(m.Latencies, other.Latencies...)

	// Merge custom metrics (last writer wins for duplicates)
	for k, v := range other.Custom {
		m.Custom[k] = v
	}

	// Recalculate throughput if we have valid time data
	if !m.StartTime.IsZero() && !m.EndTime.IsZero() {
		duration := m.EndTime.Sub(m.StartTime).Seconds()
		if duration > 0 {
			m.Throughput = float64(m.TotalOperations) / duration
		}
	}
}
