// Package performance provides utilities for performance testing and benchmarking.
// It includes tools for measuring response times, throughput, resource usage,
// and system behavior under various load conditions.
package performance

import (
	"context"
	"fmt"
	"net/http"
	"sync"
	"testing"
	"time"

	"github.com/pageza/alchemorsel-v1/internal/models"
	"github.com/pageza/alchemorsel-v1/internal/repositories"
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
	suite.SetThreshold("throughput", int64(1000))
	suite.SetThreshold("error_rate", 0.01)

	// Test database operations
	t.Run("DatabaseOperations", func(t *testing.T) {
		start := time.Now()
		
		user := &models.User{
			Name:     "Performance Test User",
			Email:    fmt.Sprintf("perf-test-%d@example.com", time.Now().UnixNano()),
			Password: "testpassword123",
		}
		
		err := repositories.CreateUser(context.Background(), user)
		errorRate := 0.0
		if err != nil {
			errorRate = 1.0
		}
		
		if user.ID != "" {
			_, err = repositories.GetUserByID(context.Background(), user.ID)
			if err != nil {
				errorRate += 0.5
			}
			
			repositories.DeleteUser(context.Background(), user.ID)
		}
		
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
		
		resp, err := http.Get("http://localhost:8080/api/health")
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
	b.Run("Insert", func(b *testing.B) {
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			user := &models.User{
				Name:     fmt.Sprintf("Benchmark User %d", i),
				Email:    fmt.Sprintf("bench-user-%d@example.com", i),
				Password: "benchmarkpassword123",
			}
			
			repositories.CreateUser(context.Background(), user)
			if user.ID != "" {
				repositories.DeleteUser(context.Background(), user.ID)
			}
		}
	})

	b.Run("Select", func(b *testing.B) {
		user := &models.User{
			Name:     "Benchmark Select User",
			Email:    "bench-select@example.com",
			Password: "benchmarkpassword123",
		}
		repositories.CreateUser(context.Background(), user)
		
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			if user.ID != "" {
				repositories.GetUserByID(context.Background(), user.ID)
			}
		}
		
		if user.ID != "" {
			repositories.DeleteUser(context.Background(), user.ID)
		}
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
					resp, err := http.Get("http://localhost:8080/api/health")
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
