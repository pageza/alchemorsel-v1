package unit

import (
	"os"
	"testing"
)

// TestConfig holds configuration for test environment
type TestConfig struct {
	RateLimitEnabled bool
	RateLimitStrict  bool
	TestMode         bool
}

// DefaultTestConfig returns the default test configuration
func DefaultTestConfig() *TestConfig {
	return &TestConfig{
		RateLimitEnabled: false, // Disable rate limiting in tests by default
		RateLimitStrict:  false, // Use relaxed rate limits in tests
		TestMode:         true,  // Enable test mode
	}
}

// SetupTestEnvironment configures the test environment with the given config
func SetupTestEnvironment(config *TestConfig) {
	if config == nil {
		config = DefaultTestConfig()
	}

	// Set environment variables for test configuration
	os.Setenv("TEST_MODE", "true")
	os.Setenv("RATE_LIMIT_ENABLED", "false")
	os.Setenv("RATE_LIMIT_STRICT", "false")
	os.Setenv("JWT_SECRET", "test-secret")
}

// SkipPeripheralTests is a helper function to skip peripheral tests during MVP
func SkipPeripheralTests(t *testing.T) {
	t.Skip("Skipping peripheral tests during MVP phase")
}
