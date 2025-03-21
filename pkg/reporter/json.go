package reporter

import (
	"encoding/json"
	"os"
	"path/filepath"
	"time"

	"github.com/mguley/go-loadtest/pkg/core"
	"github.com/mguley/go-loadtest/pkg/util"
)

// JSONReporter outputs test results in JSON format.
//
// Fields:
//   - outputPath:       The file path where JSON output is written.
//   - includeLatencies: A flag indicating whether raw latency data should be included in the output.
type JSONReporter struct {
	outputPath       string
	includeLatencies bool
}

// NewJSONReporter creates a new JSONReporter.
//
// Parameters:
//   - outputPath: The file path where the JSON results will be saved.
//
// Returns:
//   - *JSONReporter: A pointer to the newly created JSONReporter.
func NewJSONReporter(outputPath string) *JSONReporter {
	return &JSONReporter{
		outputPath:       outputPath,
		includeLatencies: false,
	}
}

// ResultOutput defines the JSON structure for test results.
//
// Fields:
//   - StartTime:         The test start time in RFC3339 format.
//   - EndTime:           The test end time in RFC3339 format.
//   - TestDuration:      The duration of the test in seconds.
//   - Tags:              Optional test tags.
//   - TotalOperations:   The total number of operations executed.
//   - ErrorCount:        The total number of errors encountered.
//   - ErrorRate:         The error rate as a percentage.
//   - Throughput:        The throughput in operations per second.
//   - LatencyP50:        The 50th percentile latency in milliseconds.
//   - LatencyP90:        The 90th percentile latency in milliseconds.
//   - LatencyP95:        The 95th percentile latency in milliseconds.
//   - LatencyP99:        The 99th percentile latency in milliseconds.
//   - LatencyMin:        The minimum latency in milliseconds.
//   - LatencyMax:        The maximum latency in milliseconds.
//   - LatencyMean:       The mean latency in milliseconds.
//   - Latencies:         Optional raw latency data.
//   - CPUUsagePercent:   The average CPU usage percentage.
//   - MemoryUsageMB:     The average memory usage in MB.
//   - ActiveGoroutines:  The number of active goroutines.
//   - GCPauseMs:         The average GC pause time in milliseconds.
//   - Custom:            Optional custom metrics.
type ResultOutput struct {
	// Test information
	StartTime    string            `json:"start_time"`
	EndTime      string            `json:"end_time"`
	TestDuration float64           `json:"test_duration_seconds"`
	Tags         map[string]string `json:"tags,omitempty"`

	// Operations metrics
	TotalOperations int64   `json:"total_operations"`
	ErrorCount      int64   `json:"error_count"`
	ErrorRate       float64 `json:"error_rate_percent"`
	Throughput      float64 `json:"throughput_ops_per_sec"`

	// Latency metrics
	LatencyP50  float64   `json:"latency_p50_ms"`
	LatencyP90  float64   `json:"latency_p90_ms"`
	LatencyP95  float64   `json:"latency_p95_ms"`
	LatencyP99  float64   `json:"latency_p99_ms"`
	LatencyMin  float64   `json:"latency_min_ms"`
	LatencyMax  float64   `json:"latency_max_ms"`
	LatencyMean float64   `json:"latency_mean_ms"`
	Latencies   []float64 `json:"latencies_ms,omitempty"`

	// Resource metrics
	CPUUsagePercent  float64 `json:"cpu_usage_percent"`
	MemoryUsageMB    float64 `json:"memory_usage_mb"`
	ActiveGoroutines int     `json:"active_goroutines"`
	GCPauseMs        float64 `json:"gc_pause_ms"`

	// Custom metrics
	Custom map[string]float64 `json:"custom,omitempty"`
}

// ReportProgress does nothing for JSONReporter.
//
// Parameters:
//   - snapshot: A pointer to a core.MetricsSnapshot (unused).
//
// Returns:
//   - error: Always returns nil.
func (r *JSONReporter) ReportProgress(snapshot *core.MetricsSnapshot) error {
	// JSONReporter doesn't handle progress updates.
	return nil
}

// ReportResults writes the final test metrics to a JSON file.
//
// Parameters:
//   - metrics: A pointer to a core.Metrics instance containing the test results.
//
// Returns:
//   - error: An error if JSON marshaling or file writing fails, otherwise nil.
func (r *JSONReporter) ReportResults(metrics *core.Metrics) error {
	latenciesData := util.Float64Data(metrics.Latencies)

	var latP50, latP90, latP95, latP99, latMin, latMax, latMean float64

	if len(latenciesData) > 0 {
		latP50, _ = latenciesData.Percentile(50)
		latP90, _ = latenciesData.Percentile(90)
		latP95, _ = latenciesData.Percentile(95)
		latP99, _ = latenciesData.Percentile(99)
		latMin = latenciesData.Min()
		latMax = latenciesData.Max()
		latMean = latenciesData.Mean()
	}

	// Calculate error rate
	var errorRate float64
	if metrics.TotalOperations > 0 {
		errorRate = float64(metrics.ErrorCount) / float64(metrics.TotalOperations) * 100
	}

	// Create output structure
	result := ResultOutput{
		StartTime:        metrics.StartTime.Format(time.RFC3339),
		EndTime:          metrics.EndTime.Format(time.RFC3339),
		TestDuration:     metrics.EndTime.Sub(metrics.StartTime).Seconds(),
		TotalOperations:  metrics.TotalOperations,
		ErrorCount:       metrics.ErrorCount,
		ErrorRate:        errorRate,
		Throughput:       metrics.Throughput,
		LatencyP50:       latP50,
		LatencyP90:       latP90,
		LatencyP95:       latP95,
		LatencyP99:       latP99,
		LatencyMin:       latMin,
		LatencyMax:       latMax,
		LatencyMean:      latMean,
		CPUUsagePercent:  metrics.ResourceMetrics.CPUUsagePercent,
		MemoryUsageMB:    metrics.ResourceMetrics.MemoryUsageMB,
		ActiveGoroutines: metrics.ResourceMetrics.ActiveGoroutines,
		GCPauseMs:        metrics.ResourceMetrics.GCPauseMs,
		Custom:           metrics.Custom,
	}

	// Include raw latencies if requested
	if r.includeLatencies && len(metrics.Latencies) > 0 {
		result.Latencies = metrics.Latencies
	}

	// Marshall the result to JSON
	jsonData, err := json.MarshalIndent(result, "", "  ")
	if err != nil {
		return err
	}

	// Write to file if path is specified, otherwise return without error
	if r.outputPath != "" {
		dir := filepath.Dir(r.outputPath)
		if err := os.MkdirAll(dir, 0o755); err != nil {
			return err
		}
		return os.WriteFile(r.outputPath, jsonData, 0o600)
	}

	return nil
}

// Name returns the name of this reporter.
//
// Returns:
//   - string: The name "JSON Results Reporter".
func (r *JSONReporter) Name() string {
	return "JSON Results Reporter"
}

// SetIncludeLatencies configures whether to include raw latency data in the output.
//
// Parameters:
//   - include: A boolean flag; true to include raw latencies, false to exclude.
func (r *JSONReporter) SetIncludeLatencies(include bool) {
	r.includeLatencies = include
}
