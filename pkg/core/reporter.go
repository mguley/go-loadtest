package core

// Reporter defines the interface for outputting test results.
//
// Methods:
//   - ReportProgress: Outputs progress information during test execution.
//   - ReportResults:  Outputs the final test results after completion.
//   - Name:           Returns a descriptive name for the reporter.
type Reporter interface {
	// ReportProgress outputs information during the test execution.
	// Parameters:
	//   - snapshot: A pointer to a MetricsSnapshot representing the current test metrics.
	// Returns:
	//   - error: An error if progress reporting fails, otherwise nil.
	ReportProgress(snapshot *MetricsSnapshot) error

	// ReportResults outputs the final test results.
	// Parameters:
	//   - metrics: A pointer to a Metrics object containing the complete test results.
	// Returns:
	//   - error: An error if result reporting fails, otherwise nil.
	ReportResults(metrics *Metrics) error

	// Name returns a descriptive name for this reporter.
	// Returns:
	//   - string: The reporter's name.
	Name() string
}

// MetricsCollector defines the interface for collecting metrics during a test.
//
// Methods:
//   - Start:      Begins metrics collection.
//   - Stop:       Ends metrics collection.
//   - GetMetrics: Retrieves the current metrics.
//   - Name:       Returns a descriptive name for the collector.
type MetricsCollector interface {
	// Start begins metrics collection.
	// Returns:
	//   - error: An error if starting fails, otherwise nil.
	Start() error

	// Stop ends metrics collection.
	// Returns:
	//   - error: An error if stopping fails, otherwise nil.
	Stop() error

	// GetMetrics returns the current metrics.
	// Returns:
	//   - *Metrics: A pointer to the collected metrics.
	GetMetrics() *Metrics

	// Name returns a descriptive name for this collector.
	// Returns:
	//   - string: The collector's name.
	Name() string
}

// OutputFormat represents the format in which test results should be reported.
type OutputFormat string

const (
	// FormatTable specifies a human-readable tabular output format.
	FormatTable OutputFormat = "table"

	// FormatJSON specifies a machine-readable JSON output format.
	FormatJSON OutputFormat = "json"

	// FormatCSV specifies a comma-separated values output format.
	FormatCSV OutputFormat = "csv"
)

// ReporterConfig contains configuration options for test reporters.
//
// Fields:
//   - Formats:          A slice of OutputFormat specifying which formats to use.
//   - OutputPath:       The file path for saving results.
//   - Quiet:            A flag to suppress progress output during the test.
//   - IncludeLatencies: A flag indicating whether to include raw latency data in the output.
type ReporterConfig struct {
	Formats          []OutputFormat
	OutputPath       string
	Quiet            bool
	IncludeLatencies bool
}

// NewDefaultReporterConfig creates a ReporterConfig with sensible default settings.
//
// Returns:
//   - *ReporterConfig: A pointer to a ReporterConfig instance with default settings
func NewDefaultReporterConfig() *ReporterConfig {
	return &ReporterConfig{
		Formats:          []OutputFormat{FormatTable},
		Quiet:            false,
		IncludeLatencies: false,
	}
}
