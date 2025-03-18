package reporter

import (
	"fmt"
	"io"
	"time"

	"github.com/mguley/go-loadtest/pkg/core"
)

// ConsoleReporter outputs test progress to an io.Writer (usually stdout).
//
// Fields:
//   - writer:    The io.Writer destination for output (typically stdout).
//   - startTime: The time when reporting started.
//   - lastOps:   The last recorded number of operations.
//   - lastErrs:  The last recorded number of errors.
//   - interval:  The duration used for rate calculation.
//   - quiet:     A flag that, when true, suppresses progress output.
type ConsoleReporter struct {
	writer    io.Writer
	startTime time.Time
	lastOps   int64
	lastErrs  int64
	interval  time.Duration
	quiet     bool
}

// NewConsoleReporter creates a new ConsoleReporter.
//
// Parameters:
//   - writer:   The io.Writer to which output is written.
//   - interval: The duration that defines how frequently the rate is calculated.
//
// Returns:
//   - *ConsoleReporter: A pointer to the newly created ConsoleReporter instance.
func NewConsoleReporter(writer io.Writer, interval time.Duration) *ConsoleReporter {
	return &ConsoleReporter{
		writer:   writer,
		interval: interval,
		quiet:    false,
	}
}

// ReportProgress outputs the current test progress to the console.
//
// Parameters:
//   - snapshot: A pointer to a core.MetricsSnapshot containing the current test metrics.
//
// Returns:
//   - error: An error if output fails, otherwise nil.
func (r *ConsoleReporter) ReportProgress(snapshot *core.MetricsSnapshot) error {
	if r.quiet {
		return nil
	}

	if r.startTime.IsZero() {
		r.startTime = snapshot.Timestamp
	}

	elapsed := snapshot.Timestamp.Sub(r.startTime).Seconds()

	// Calculate delta since last report for rate calculation
	opsDelta := snapshot.Operations - r.lastOps

	// Store current values for next calculation
	r.lastOps = snapshot.Operations
	r.lastErrs = snapshot.Errors

	// Calculate current rate based on the interval
	currentRate := float64(opsDelta) / r.interval.Seconds()

	// Format custom metrics if present
	customMetrics := ""
	if len(snapshot.Custom) > 0 {
		customMetrics = " | "
		for k, v := range snapshot.Custom {
			customMetrics += fmt.Sprintf("%s: %.2f ", k, v)
		}
	}

	// Print progress line
	_, err := fmt.Fprintf(r.writer,
		"[%.1fs] Rate: %d ops/s | Total: %d ops | Errors: %d | Current: %.2f ops/s%s\n",
		elapsed,
		int(currentRate),
		snapshot.Operations,
		snapshot.Errors,
		snapshot.RatePerSecond,
		customMetrics,
	)

	return err
}

// ReportResults does nothing for ConsoleReporter.
//
// Parameters:
//   - metrics: A pointer to a core.Metrics instance (unused).
//
// Returns:
//   - error: Always returns nil.
func (r *ConsoleReporter) ReportResults(metrics *core.Metrics) error {
	// ConsoleReporter is only responsible for progress updates
	return nil
}

// Name returns the name of this reporter.
//
// Returns:
//   - string: The name "Console Progress Reporter".
func (r *ConsoleReporter) Name() string {
	return "Console Progress Reporter"
}

// SetQuiet enables or disables progress output.
//
// Parameters:
//   - quiet: A boolean flag; true to suppress output, false to enable.
func (r *ConsoleReporter) SetQuiet(quiet bool) {
	r.quiet = quiet
}
