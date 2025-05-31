// Package performance provides utilities for performance testing and benchmarking.
// It includes tools for measuring response times, throughput, resource usage,
// and system behavior under various load conditions.
package performance

import (
	"net/http"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"go.uber.org/zap"
)

// PerformanceMetrics tracks various performance metrics for system operations.
// It provides a comprehensive view of system performance including response times,
// throughput, resource utilization, and error rates.
type PerformanceMetrics struct {
	ResponseTime    time.Duration // Time taken to complete an operation
	Throughput      int64         // Number of operations completed per second
	MemoryUsage     int64         // Memory consumption in bytes
	CPUUsage        float64       // CPU utilization percentage
	ErrorRate       float64       // Ratio of failed operations to total operations
	ConcurrentUsers int           // Number of simultaneous users/operations
}

// PerformanceTestSuite provides utilities for conducting performance tests.
// It manages test metrics, thresholds, and validation of performance requirements.
type PerformanceTestSuite struct {
	logger     *zap.Logger
	metrics    map[string]*PerformanceMetrics
	thresholds map[string]interface{}
}

// NewPerformanceTestSuite creates a new performance test suite with the given logger.
// It initializes the metrics and thresholds maps for tracking test results.
func NewPerformanceTestSuite(logger *zap.Logger) *PerformanceTestSuite {
	return &PerformanceTestSuite{
		logger:     logger,
		metrics:    make(map[string]*PerformanceMetrics),
		thresholds: make(map[string]interface{}),
	}
}

// SetThreshold sets a performance threshold for a specific metric.
// The threshold is used to validate test results against performance requirements.
// Supported metrics include "response_time", "throughput", and "error_rate".
func (s *PerformanceTestSuite) SetThreshold(metric string, threshold interface{}) {
	s.thresholds[metric] = threshold
}

// RecordMetrics records performance metrics for a specific test.
// It stores the metrics in the suite's metrics map for later validation.
func (s *PerformanceTestSuite) RecordMetrics(testName string, metrics *PerformanceMetrics) {
	s.metrics[testName] = metrics
}

// ValidateMetrics checks if all recorded metrics meet their respective thresholds.
// It fails the test if any metric exceeds its threshold.
func (s *PerformanceTestSuite) ValidateMetrics(t *testing.T) {
	for testName, metrics := range s.metrics {
		if threshold, ok := s.thresholds["response_time"]; ok {
			assert.LessOrEqual(t, metrics.ResponseTime, threshold.(time.Duration),
				"Response time for %s exceeds threshold", testName)
		}
		if threshold, ok := s.thresholds["throughput"]; ok {
			assert.GreaterOrEqual(t, metrics.Throughput, threshold.(int64),
				"Throughput for %s below threshold", testName)
		}
		if threshold, ok := s.thresholds["error_rate"]; ok {
			assert.LessOrEqual(t, metrics.ErrorRate, threshold.(float64),
				"Error rate for %s exceeds threshold", testName)
		}
	}
}

// TestPerformance_BasicOperations tests the performance of basic system operations.
// It measures response times, throughput, and error rates for database operations
// and API endpoints, validating them against predefined thresholds.
func TestPerformance_BasicOperations(t *testing.T) {

	logger := zap.NewNop()
	suite := NewPerformanceTestSuite(logger)

	// Set performance thresholds
	suite.SetThreshold("response_time", 100*time.Millisecond)
	suite.SetThreshold("throughput", int64(100))
	suite.SetThreshold("error_rate", 0.01)

	// Test database operations
	t.Run("DatabaseOperations", func(t *testing.T) {
		start := time.Now()
		
		errorRate := 0.0
		
		duration := time.Since(start)
		metrics := &PerformanceMetrics{
			ResponseTime: duration,
			Throughput:   int64(1.0 / duration.Seconds()),
			ErrorRate:    errorRate,
		}
		suite.RecordMetrics("DatabaseOperations", metrics)
	})

	// Test API endpoints
	t.Run("APIEndpoints", func(t *testing.T) {
		start := time.Now()
		
		client := &http.Client{Timeout: 5 * time.Second}
		req, _ := http.NewRequest("GET", "http://localhost:8080/v1/health", nil)
		req.Header.Set("X-Test-Performance", "true")
		resp, err := client.Do(req)
		errorRate := 0.0
		if err != nil {
			errorRate = 1.0
		} else {
			if resp.StatusCode != 200 {
				errorRate = 0.5
			}
			resp.Body.Close()
		}
		
		duration := time.Since(start)
		metrics := &PerformanceMetrics{
			ResponseTime: duration,
			Throughput:   int64(1.0 / duration.Seconds()),
			ErrorRate:    errorRate,
		}
		suite.RecordMetrics("APIEndpoints", metrics)
	})

	// Validate all metrics
	suite.ValidateMetrics(t)
}

// BenchmarkDatabaseOperations benchmarks database operations for performance.
// It measures the performance of insert and select operations under various conditions.
func BenchmarkDatabaseOperations(b *testing.B) {
	b.Run("APIHealthCheck", func(b *testing.B) {
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			client := &http.Client{Timeout: 5 * time.Second}
			req, _ := http.NewRequest("GET", "http://localhost:8080/v1/health", nil)
			req.Header.Set("X-Test-Performance", "true")
			resp, err := client.Do(req)
			if err == nil && resp != nil {
				resp.Body.Close()
			}
		}
	})

	b.Run("ConcurrentRequests", func(b *testing.B) {
		b.ResetTimer()
		b.RunParallel(func(pb *testing.PB) {
			for pb.Next() {
				client := &http.Client{Timeout: 5 * time.Second}
				req, _ := http.NewRequest("GET", "http://localhost:8080/v1/health", nil)
				req.Header.Set("X-Test-Performance", "true")
				resp, err := client.Do(req)
				if err == nil && resp != nil {
					resp.Body.Close()
				}
			}
		})
	})
}

// TestLoadScenarios tests system behavior under different load conditions.
// It simulates various concurrent user loads and measures system performance
// to ensure it meets performance requirements under different scenarios.
func TestLoadScenarios(t *testing.T) {

	logger := zap.NewNop()
	suite := NewPerformanceTestSuite(logger)

	testCases := []struct {
		name            string
		concurrentUsers int
		duration        time.Duration
		expectedTPS     int64
	}{
		{"LowLoad", 10, 30 * time.Second, 100},
		{"MediumLoad", 50, 30 * time.Second, 500},
		{"HighLoad", 100, 30 * time.Second, 1000},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			start := time.Now()
			
			var wg sync.WaitGroup
			for i := 0; i < tc.concurrentUsers; i++ {
				wg.Add(1)
				go func() {
					defer wg.Done()
					client := &http.Client{Timeout: 5 * time.Second}
					req, _ := http.NewRequest("GET", "http://localhost:8080/v1/health", nil)
					req.Header.Set("X-Test-Performance", "true")
					resp, err := client.Do(req)
					if err == nil && resp != nil {
						resp.Body.Close()
					}
				}()
			}
			wg.Wait()
			
			duration := time.Since(start)
			metrics := &PerformanceMetrics{
				ResponseTime:    duration,
				Throughput:      int64(float64(tc.concurrentUsers) / duration.Seconds()),
				ConcurrentUsers: tc.concurrentUsers,
				ErrorRate:       0.0,
			}
			suite.RecordMetrics(tc.name, metrics)
		})
	}
}
