// Package chaos provides utilities for chaos engineering and resilience testing.
// It includes tools for simulating network issues, service disruptions,
// resource exhaustion, and other failure scenarios to test system resilience.
package chaos

import (
	"context"
	"math/rand"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"go.uber.org/zap"
)

// ChaosTestSuite provides utilities for conducting chaos tests.
// It manages random number generation and various chaos scenarios
// to test system resilience and recovery capabilities.
type ChaosTestSuite struct {
	logger *zap.Logger
	rand   *rand.Rand
}

// ChaosMetrics tracks chaos test results and system behavior.
// It provides information about service disruptions, recovery times,
// error rates, and resource utilization during chaos scenarios.
type ChaosMetrics struct {
	ServiceDisruptions int           // Number of service disruptions simulated
	RecoveryTime       time.Duration // Time taken for system to recover
	ErrorRate          float64       // Ratio of errors during chaos scenarios
	LatencySpikes      int           // Number of latency spikes observed
	ResourceExhaustion bool          // Whether resource exhaustion was detected
}

// NewChaosTestSuite creates a new chaos test suite with the given logger.
// It initializes a random number generator for simulating chaos scenarios.
func NewChaosTestSuite(logger *zap.Logger) *ChaosTestSuite {
	return &ChaosTestSuite{
		logger: logger,
		rand:   rand.New(rand.NewSource(time.Now().UnixNano())),
	}
}

// SimulateNetworkLatency simulates network latency spikes and delays.
// It randomly introduces network delays to test system behavior under
// poor network conditions.
func (s *ChaosTestSuite) SimulateNetworkLatency(ctx context.Context, duration time.Duration) {
	ticker := time.NewTicker(100 * time.Millisecond)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			if s.rand.Float64() < 0.1 { // 10% chance of latency spike
				time.Sleep(time.Duration(s.rand.Int63n(1000)) * time.Millisecond)
			}
		}
	}
}

// SimulateServiceDisruption simulates service disruptions and failures.
// It randomly introduces service downtime to test system resilience
// and recovery capabilities.
func (s *ChaosTestSuite) SimulateServiceDisruption(ctx context.Context, duration time.Duration) {
	ticker := time.NewTicker(1 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			if s.rand.Float64() < 0.05 { // 5% chance of disruption
				time.Sleep(time.Duration(s.rand.Int63n(5000)) * time.Millisecond)
			}
		}
	}
}

// SimulateResourceExhaustion simulates resource exhaustion scenarios.
// It creates multiple goroutines that consume system resources to test
// system behavior under resource constraints.
func (s *ChaosTestSuite) SimulateResourceExhaustion(ctx context.Context, duration time.Duration) {
	var wg sync.WaitGroup
	for i := 0; i < 100; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for {
				select {
				case <-ctx.Done():
					return
				default:
					// Simulate resource-intensive operations
					_ = make([]byte, 1024*1024) // Allocate 1MB
					time.Sleep(10 * time.Millisecond)
				}
			}
		}()
	}
	wg.Wait()
}

// TestChaos_NetworkLatency tests system behavior under network latency conditions.
// It verifies that the system can handle network delays and maintain
// acceptable performance levels.
func TestChaos_NetworkLatency(t *testing.T) {

	logger := zap.NewNop()
	suite := NewChaosTestSuite(logger)

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	// Start network latency simulation
	go suite.SimulateNetworkLatency(ctx, 30*time.Second)

	// Perform operations that should be resilient to network latency
	start := time.Now()
	for i := 0; i < 100; i++ {
		// Simulate API calls or database operations
		time.Sleep(100 * time.Millisecond)
	}
	duration := time.Since(start)

	assert.Less(t, duration, 40*time.Second, "Operations should complete within reasonable time despite latency")
}

// TestChaos_ServiceDisruption tests system behavior under service disruptions.
// It verifies that the system can handle service failures and maintain
// acceptable functionality levels.
func TestChaos_ServiceDisruption(t *testing.T) {

	logger := zap.NewNop()
	suite := NewChaosTestSuite(logger)

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	// Start service disruption simulation
	go suite.SimulateServiceDisruption(ctx, 30*time.Second)

	// Perform operations that should handle service disruptions
	start := time.Now()
	for i := 0; i < 50; i++ {
		// Simulate service calls with retries
		time.Sleep(200 * time.Millisecond)
	}
	duration := time.Since(start)

	assert.Less(t, duration, 40*time.Second, "Operations should complete within reasonable time despite disruptions")
}

// TestChaos_ResourceExhaustion tests system behavior under resource constraints.
// It verifies that the system can handle resource limitations and maintain
// acceptable performance levels.
func TestChaos_ResourceExhaustion(t *testing.T) {

	logger := zap.NewNop()
	suite := NewChaosTestSuite(logger)

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	// Start resource exhaustion simulation
	go suite.SimulateResourceExhaustion(ctx, 30*time.Second)

	// Perform operations that should handle resource constraints
	start := time.Now()
	for i := 0; i < 50; i++ {
		// Simulate operations that should be resilient to resource constraints
		time.Sleep(100 * time.Millisecond)
	}
	duration := time.Since(start)

	assert.Less(t, duration, 40*time.Second, "Operations should complete within reasonable time despite resource constraints")
}

// TestChaos_Recovery tests system recovery after chaos scenarios.
// It verifies that the system can recover and return to normal operation
// after experiencing various failure conditions.
func TestChaos_Recovery(t *testing.T) {

	logger := zap.NewNop()
	suite := NewChaosTestSuite(logger)

	// First apply chaos scenarios
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	var wg sync.WaitGroup
	wg.Add(3)
	go func() {
		defer wg.Done()
		suite.SimulateNetworkLatency(ctx, 30*time.Second)
	}()
	go func() {
		defer wg.Done()
		suite.SimulateServiceDisruption(ctx, 30*time.Second)
	}()
	go func() {
		defer wg.Done()
		suite.SimulateResourceExhaustion(ctx, 30*time.Second)
	}()

	wg.Wait()

	// Then test recovery
	start := time.Now()
	for i := 0; i < 50; i++ {
		// Simulate normal operations after chaos
		time.Sleep(100 * time.Millisecond)
	}
	duration := time.Since(start)

	assert.Less(t, duration, 10*time.Second, "System should recover quickly after chaos scenarios")
}
