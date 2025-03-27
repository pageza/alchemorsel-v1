package errors

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestErrorTracker(t *testing.T) {
	tracker := NewErrorTracker()

	// Test error tracking
	err := &Error{
		Code:    "TEST_ERROR",
		Message: "Test error message",
	}
	tracker.TrackError(err)

	// Test error rate calculation
	rate := tracker.GetErrorRate("TEST_ERROR")
	if rate <= 0 {
		t.Errorf("Expected error rate > 0, got %f", rate)
	}

	// Test error threshold
	tracker.SetErrorThreshold("TEST_ERROR", 0.1)
	if !tracker.HasExceededThreshold("TEST_ERROR") {
		t.Error("Expected error threshold to be exceeded")
	}

	// Test error count reset
	tracker.ResetErrorCount("TEST_ERROR")
	if tracker.GetErrorRate("TEST_ERROR") != 0 {
		t.Error("Expected error rate to be 0 after reset")
	}
}

func TestErrorTrackerConcurrency(t *testing.T) {
	tracker := NewErrorTracker()
	err := &Error{
		Code:    "CONCURRENT_ERROR",
		Message: "Concurrent test error",
	}

	// Test concurrent error tracking
	done := make(chan bool)
	for i := 0; i < 100; i++ {
		go func() {
			tracker.TrackError(err)
			done <- true
		}()
	}

	// Wait for all goroutines to complete
	for i := 0; i < 100; i++ {
		<-done
	}

	// Verify error count
	rate := tracker.GetErrorRate("CONCURRENT_ERROR")
	if rate != 100 {
		t.Errorf("Expected error rate to be 100, got %f", rate)
	}
}

func TestErrorTrackerTimeWindow(t *testing.T) {
	tracker := NewErrorTracker()
	err := &Error{
		Code:    "TIME_WINDOW_ERROR",
		Message: "Time window test error",
	}

	// Track error
	tracker.TrackError(err)

	// Wait for time window to expire
	time.Sleep(2 * time.Second)

	// Verify error rate is still tracked
	rate := tracker.GetErrorRate("TIME_WINDOW_ERROR")
	if rate <= 0 {
		t.Errorf("Expected error rate > 0, got %f", rate)
	}
}

func TestRecoveryStrategies(t *testing.T) {
	// Test retry strategy
	retryStrategy := RetryStrategy(3, time.Millisecond)
	err := New("RETRY_ERROR", "retry error")
	recoveredErr := retryStrategy(context.Background(), err)
	assert.NoError(t, recoveredErr)

	// Test fallback strategy
	fallbackCalled := false
	fallbackStrategy := FallbackStrategy(func() error {
		fallbackCalled = true
		return nil
	})
	err = New("FALLBACK_ERROR", "fallback error")
	recoveredErr = fallbackStrategy(context.Background(), err)
	assert.NoError(t, recoveredErr)
	assert.True(t, fallbackCalled)

	// Test circuit breaker strategy
	circuitBreaker := CircuitBreakerStrategy(3, time.Second)
	err = New("CIRCUIT_ERROR", "circuit error")
	for i := 0; i < 4; i++ {
		recoveredErr = circuitBreaker(context.Background(), err)
		if i < 3 {
			assert.NoError(t, recoveredErr)
		} else {
			assert.Error(t, recoveredErr)
		}
	}

	// Test timeout strategy
	timeoutStrategy := TimeoutStrategy(time.Millisecond)
	err = New("TIMEOUT_ERROR", "timeout error")
	recoveredErr = timeoutStrategy(context.Background(), err)
	assert.NoError(t, recoveredErr)
}

func TestErrorAnalytics(t *testing.T) {
	tracker := NewErrorTracker()

	// Track multiple errors
	err := New("TEST_ERROR", "test error")
	for i := 0; i < 5; i++ {
		tracker.TrackError(err)
	}

	// Register recovery strategy and recover some errors
	tracker.RegisterRecoveryStrategy("TEST_ERROR", RetryStrategy(3, time.Millisecond))
	for i := 0; i < 3; i++ {
		tracker.TrackError(err)
	}

	// Check analytics
	analytics := tracker.GetAnalytics()
	errorCounts := analytics["error_counts"].(map[string]int64)
	errorRates := analytics["error_rates"].(map[string]float64)
	thresholds := analytics["thresholds"].(map[string]float64)

	assert.Equal(t, int64(8), errorCounts["TEST_ERROR"])
	assert.Greater(t, errorRates["TEST_ERROR"], 0.0)
	assert.Equal(t, 0.0, thresholds["TEST_ERROR"])
}

func TestErrorLatencyTracking(t *testing.T) {
	tracker := NewErrorTracker()

	// Track error with known latency
	err := New("LATENCY_ERROR", "latency error")
	start := time.Now()
	tracker.TrackError(err)
	latency := time.Since(start)

	// Check analytics
	analytics := tracker.GetAnalytics()
	errorRates := analytics["error_rates"].(map[string]float64)
	assert.Greater(t, errorRates["LATENCY_ERROR"], 0.0)
	assert.Less(t, errorRates["LATENCY_ERROR"], float64(latency.Seconds()))
}
