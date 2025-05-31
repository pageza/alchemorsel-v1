// Package load provides utilities for load testing and stress testing.
// It includes tools for simulating concurrent users, measuring system performance
// under load, and validating system behavior under various stress conditions.
package load

import (
	"context"

	"net/http"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"go.uber.org/zap"
)

// LoadTestSuite provides utilities for conducting load tests.
// It manages concurrent user simulation, metrics collection, and test execution.
type LoadTestSuite struct {
	logger     *zap.Logger
	metrics    *LoadMetrics
	concurrent int
	duration   time.Duration
}

// LoadMetrics tracks various metrics during load testing.
// It provides detailed information about request patterns, response times,
// and system performance under load.
type LoadMetrics struct {
	TotalRequests       int64         // Total number of requests made
	SuccessfulRequests  int64         // Number of successful requests
	FailedRequests      int64         // Number of failed requests
	AverageResponseTime time.Duration // Average time taken to process requests
	MaxResponseTime     time.Duration // Maximum response time observed
	MinResponseTime     time.Duration // Minimum response time observed
	RequestsPerSecond   float64       // Average number of requests processed per second
	ErrorRate           float64       // Ratio of failed requests to total requests
}

// NewLoadTestSuite creates a new load test suite with the given parameters.
// It initializes the metrics tracking and sets up the test environment.
func NewLoadTestSuite(logger *zap.Logger, concurrent int, duration time.Duration) *LoadTestSuite {
	return &LoadTestSuite{
		logger:     logger,
		metrics:    &LoadMetrics{},
		concurrent: concurrent,
		duration:   duration,
	}
}

// SimulateUser simulates a single user's behavior during the load test.
// It runs in a goroutine and performs operations until the context is cancelled.
// The function records metrics for each operation performed.
func (s *LoadTestSuite) SimulateUser(ctx context.Context, wg *sync.WaitGroup) {
	defer wg.Done()

	for {
		select {
		case <-ctx.Done():
			return
		default:
			start := time.Now()
			
			client := &http.Client{Timeout: 5 * time.Second}
			req, _ := http.NewRequest("GET", "http://localhost:8080/v1/health", nil)
			req.Header.Set("X-Test-Load", "true")
			resp, err := client.Do(req)
			responseTime := time.Since(start)
			
			success := err == nil && resp != nil && resp.StatusCode == 200
			if resp != nil {
				resp.Body.Close()
			}
			
			s.recordMetrics(responseTime, success)
			
			time.Sleep(10 * time.Millisecond)
		}
	}
}

// recordMetrics records metrics for a single request.
// It updates various metrics including response times, success/failure counts,
// and calculates running averages.
func (s *LoadTestSuite) recordMetrics(responseTime time.Duration, success bool) {
	s.metrics.TotalRequests++
	if success {
		s.metrics.SuccessfulRequests++
	} else {
		s.metrics.FailedRequests++
	}

	// Update response time metrics
	if s.metrics.MinResponseTime == 0 || responseTime < s.metrics.MinResponseTime {
		s.metrics.MinResponseTime = responseTime
	}
	if responseTime > s.metrics.MaxResponseTime {
		s.metrics.MaxResponseTime = responseTime
	}

	// Update average response time
	s.metrics.AverageResponseTime = time.Duration(
		float64(s.metrics.AverageResponseTime)*float64(s.metrics.TotalRequests-1)+
			float64(responseTime),
	) / time.Duration(s.metrics.TotalRequests)
}

// RunLoadTest executes the load test with the configured parameters.
// It starts multiple goroutines to simulate concurrent users, runs for the
// specified duration, and validates the results.
func (s *LoadTestSuite) RunLoadTest(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), s.duration)
	defer cancel()

	var wg sync.WaitGroup
	wg.Add(s.concurrent)

	// Start concurrent users
	for i := 0; i < s.concurrent; i++ {
		go s.SimulateUser(ctx, &wg)
	}

	// Wait for all users to complete
	wg.Wait()

	// Calculate final metrics
	s.metrics.RequestsPerSecond = float64(s.metrics.TotalRequests) / s.duration.Seconds()
	s.metrics.ErrorRate = float64(s.metrics.FailedRequests) / float64(s.metrics.TotalRequests)

	// Validate metrics
	s.validateMetrics(t)
}

// validateMetrics checks if the load test metrics meet performance requirements.
// It verifies that the system maintains acceptable performance under load.
func (s *LoadTestSuite) validateMetrics(t *testing.T) {
	assert.Greater(t, s.metrics.RequestsPerSecond, 0.0, "Requests per second should be greater than 0")
	assert.Less(t, s.metrics.ErrorRate, 0.01, "Error rate should be less than 1%")
	assert.Greater(t, s.metrics.SuccessfulRequests, int64(0), "Should have successful requests")
}

// TestLoad_ConcurrentUsers tests system behavior under different levels of concurrent user load.
// It verifies that the system can handle increasing numbers of simultaneous users
// while maintaining acceptable performance.
func TestLoad_ConcurrentUsers(t *testing.T) {

	logger := zap.NewNop()

	testCases := []struct {
		name       string
		concurrent int
		duration   time.Duration
	}{
		{"LowConcurrency", 10, 30 * time.Second},
		{"MediumConcurrency", 50, 30 * time.Second},
		{"HighConcurrency", 100, 30 * time.Second},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			suite := NewLoadTestSuite(logger, tc.concurrent, tc.duration)
			suite.RunLoadTest(t)
		})
	}
}

// TestLoad_Stress tests system behavior under extreme stress conditions.
// It simulates a high number of concurrent users over an extended period
// to verify system stability and performance under stress.
func TestLoad_Stress(t *testing.T) {

	logger := zap.NewNop()
	suite := NewLoadTestSuite(logger, 200, 60*time.Second)
	suite.RunLoadTest(t)
}

// TestLoad_Recovery tests system recovery after high load conditions.
// It first applies high load to stress the system, then verifies that
// the system can recover and maintain normal performance levels.
func TestLoad_Recovery(t *testing.T) {

	logger := zap.NewNop()

	// First apply high load
	suite := NewLoadTestSuite(logger, 200, 30*time.Second)
	suite.RunLoadTest(t)

	// Then test recovery with normal load
	suite = NewLoadTestSuite(logger, 50, 30*time.Second)
	suite.RunLoadTest(t)
}
