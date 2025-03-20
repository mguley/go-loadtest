package entity

import "time"

// LoadTestType defines the type of load test to perform.
//
// Values:
//   - PublishTest:   Tests message publishing performance.
//   - SubscribeTest: Tests message subscription performance.
//   - CombinedTest:  Tests both publishing and subscribing.
type LoadTestType string

const (
	// PublishTest tests message publishing performance.
	PublishTest LoadTestType = "publish"
	// SubscribeTest tests message subscription performance.
	SubscribeTest LoadTestType = "subscribe"
	// CombinedTest tests both publishing and subscribing.
	CombinedTest LoadTestType = "combined"
)

// LoadTest represents a load test configuration for NATS operations.
//
// Fields:
//   - Type:            Type of load test to perform.
//   - MessageSize:     Size of the message payload in bytes.
//   - Subject:         NATS subject for the test.
//   - QueueGroup:      Queue group name (if applicable).
//   - PayloadTemplate: Template payload for test messages.
type LoadTest struct {
	Type            LoadTestType
	MessageSize     int
	Subject         string
	QueueGroup      string
	PayloadTemplate []byte
}

// TestResult represents the outcome of a load test.
//
// Fields:
//   - TestType:         Type of load test performed.
//   - Subject:          NATS subject used during the test.
//   - MessageSize:      Size of the message payload in bytes.
//   - TotalMessages:    Total number of messages processed.
//   - Errors:           Number of errors encountered during the test.
//   - ThroughputPerSec: Number of messages processed per second.
//   - AverageLatencyMs: Average latency in milliseconds.
//   - P95LatencyMs:     95th percentile latency in milliseconds.
//   - P99LatencyMs:     99th percentile latency in milliseconds.
//   - Duration:         Total duration of the test.
//   - StartTime:        Timestamp when the test started.
//   - EndTime:          Timestamp when the test ended.
type TestResult struct {
	TestType         LoadTestType
	Subject          string
	MessageSize      int
	TotalMessages    int64
	Errors           int64
	ThroughputPerSec float64
	AverageLatencyMs float64
	P95LatencyMs     float64
	P99LatencyMs     float64
	Duration         time.Duration
	StartTime        time.Time
	EndTime          time.Time
}
