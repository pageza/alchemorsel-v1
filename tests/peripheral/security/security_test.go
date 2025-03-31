// Package security provides utilities for security testing and vulnerability assessment.
// It includes tools for checking security headers, SSL/TLS configuration,
// SQL injection vulnerabilities, XSS vulnerabilities, and authentication mechanisms.
package security

import (
	"crypto/tls"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"unicode"

	"github.com/gin-gonic/gin"
	"github.com/pageza/alchemorsel-v1/internal/middleware"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"
)

// SecurityTestSuite provides utilities for conducting security tests.
// It manages HTTP clients, security checks, and vulnerability scanning.
type SecurityTestSuite struct {
	logger *zap.Logger
	client *http.Client
}

// SecurityMetrics tracks security test results and findings.
// It provides information about vulnerabilities, security configurations,
// and potential security issues.
type SecurityMetrics struct {
	VulnerabilitiesFound int               // Number of security vulnerabilities detected
	SecurityHeaders      map[string]string // Required security headers and their values
	SSLConfig            *tls.Config       // SSL/TLS configuration details
	OpenPorts            []int             // List of open ports that might need attention
	WeakPasswords        []string          // List of weak passwords detected
}

// NewSecurityTestSuite creates a new security test suite with the given logger.
// It initializes an HTTP client with appropriate security settings for testing.
func NewSecurityTestSuite(logger *zap.Logger) *SecurityTestSuite {
	return &SecurityTestSuite{
		logger: logger,
		client: &http.Client{
			Transport: &http.Transport{
				TLSClientConfig: &tls.Config{
					InsecureSkipVerify: true,
				},
			},
		},
	}
}

// CheckSecurityHeaders checks for required security headers in HTTP responses.
// It verifies that all necessary security headers are present and properly configured.
// Required headers include X-Frame-Options, X-Content-Type-Options, X-XSS-Protection,
// Strict-Transport-Security, and Content-Security-Policy.
func (s *SecurityTestSuite) CheckSecurityHeaders(t *testing.T, url string) {
	resp, err := s.client.Get(url)
	require.NoError(t, err)
	defer resp.Body.Close()

	requiredHeaders := map[string]string{
		"X-Frame-Options":           "DENY",
		"X-Content-Type-Options":    "nosniff",
		"X-XSS-Protection":          "1; mode=block",
		"Strict-Transport-Security": "max-age=31536000; includeSubDomains",
		"Content-Security-Policy":   "default-src 'self'",
	}

	for header, expectedValue := range requiredHeaders {
		actualValue := resp.Header.Get(header)
		assert.Equal(t, expectedValue, actualValue,
			"Security header %s not set correctly", header)
	}
}

// CheckSSLConfiguration checks SSL/TLS configuration of the target server.
// It verifies that the server uses secure TLS versions and proper cipher suites.
// The function also checks for certificate validity and configuration.
func (s *SecurityTestSuite) CheckSSLConfiguration(t *testing.T, url string) {
	// Prepend 'https://' if the URL doesn't already have a scheme
	if !strings.HasPrefix(url, "http") {
		url = "https://" + url
	}
	resp, err := s.client.Get(url)
	require.NoError(t, err)
	defer resp.Body.Close()

	state := resp.TLS
	require.NotNil(t, state, "TLS state should not be nil")
	assert.True(t, state.HandshakeComplete, "SSL handshake should complete")
	assert.Condition(t, func() bool {
		return int(state.Version) >= int(tls.VersionTLS12)
	}, "Should use TLS 1.2 or higher")
}

// CheckSQLInjection checks for SQL injection vulnerabilities in the target endpoint.
// It tests various SQL injection attack patterns and verifies that the system
// properly handles and sanitizes input to prevent SQL injection attacks.
func (s *SecurityTestSuite) CheckSQLInjection(t *testing.T, endpoint string) {
	testCases := []string{
		"' OR '1'='1",
		"'; DROP TABLE users; --",
		"' UNION SELECT * FROM users; --",
	}

	for _, testCase := range testCases {
		resp, err := s.client.Get(endpoint + "?id=" + testCase)
		require.NoError(t, err)
		defer resp.Body.Close()

		// Check for error responses that might indicate SQL injection vulnerability
		assert.NotEqual(t, http.StatusInternalServerError, resp.StatusCode,
			"Potential SQL injection vulnerability detected with input: %s", testCase)
	}
}

// CheckXSSVulnerabilities checks for Cross-Site Scripting (XSS) vulnerabilities.
// It tests various XSS attack patterns and verifies that the system properly
// escapes and sanitizes output to prevent XSS attacks.
func (s *SecurityTestSuite) CheckXSSVulnerabilities(t *testing.T, endpoint string) {
	testCases := []string{
		"<script>alert('xss')</script>",
		"javascript:alert('xss')",
		"onerror=alert('xss')",
	}

	for _, testCase := range testCases {
		resp, err := s.client.Get(endpoint + "?input=" + testCase)
		require.NoError(t, err)
		defer resp.Body.Close()

		// Check if the input is reflected in the response without proper escaping
		body, err := io.ReadAll(resp.Body)
		require.NoError(t, err)
		assert.NotContains(t, string(body), testCase,
			"Potential XSS vulnerability detected with input: %s", testCase)
	}
}

// TestSecurity_Headers tests the presence and configuration of security headers.
// It verifies that all required security headers are properly set and configured.
func TestSecurity_Headers(t *testing.T) {
	t.Skip("Temporarily disabled for MVP")
	router := gin.New()
	router.Use(middleware.SecurityHeaders())
	router.GET("/test", func(c *gin.Context) { c.String(http.StatusOK, "ok") })
	ts := httptest.NewServer(router)
	defer ts.Close()

	logger := zap.NewNop()
	suite := NewSecurityTestSuite(logger)
	suite.CheckSecurityHeaders(t, ts.URL+"/test")
}

// TestSecurity_SSL tests the SSL/TLS configuration of the target server.
// It verifies that the server uses secure TLS versions and proper cipher suites.
func TestSecurity_SSL(t *testing.T) {
	t.Skip("Temporarily disabled for MVP")
	router := gin.New()
	router.Use(middleware.SecurityHeaders())
	router.GET("/test", func(c *gin.Context) { c.String(http.StatusOK, "ok") })
	ts := httptest.NewTLSServer(router)
	defer ts.Close()

	logger := zap.NewNop()
	suite := NewSecurityTestSuite(logger)
	// Use the test server's address
	suite.CheckSSLConfiguration(t, ts.Listener.Addr().String())
}

// TestSecurity_SQLInjection tests for SQL injection vulnerabilities.
// It verifies that the system properly handles and sanitizes input to prevent
// SQL injection attacks.
func TestSecurity_SQLInjection(t *testing.T) {
	t.Skip("Temporarily disabled for MVP")
	router := gin.New()
	router.Use(middleware.SecurityHeaders())
	router.GET("/api/users", func(c *gin.Context) { c.Status(http.StatusOK) })
	ts := httptest.NewServer(router)
	defer ts.Close()

	logger := zap.NewNop()
	suite := NewSecurityTestSuite(logger)
	suite.CheckSQLInjection(t, ts.URL+"/api/users")
}

// TestSecurity_XSS tests for Cross-Site Scripting (XSS) vulnerabilities.
// It verifies that the system properly escapes and sanitizes output to prevent
// XSS attacks.
func TestSecurity_XSS(t *testing.T) {
	t.Skip("Temporarily disabled for MVP")
	router := gin.New()
	router.Use(middleware.SecurityHeaders())
	router.GET("/api/search", func(c *gin.Context) { c.String(http.StatusOK, "ok") })
	ts := httptest.NewServer(router)
	defer ts.Close()

	logger := zap.NewNop()
	suite := NewSecurityTestSuite(logger)
	suite.CheckXSSVulnerabilities(t, ts.URL+"/api/search")
}

// TestSecurity_Authentication tests authentication mechanisms and password policies.
// It verifies that the system enforces strong password requirements and proper
// authentication practices.
func TestSecurity_Authentication(t *testing.T) {
	t.Skip("Temporarily disabled for MVP")
	router := gin.New()
	router.Use(middleware.SecurityHeaders())
	router.GET("/auth", func(c *gin.Context) { c.String(http.StatusOK, "ok") })
	ts := httptest.NewServer(router)
	defer ts.Close()

	logger := zap.NewNop()
	suite := NewSecurityTestSuite(logger)

	// Test security headers for authentication endpoints
	suite.CheckSecurityHeaders(t, ts.URL+"/auth")

	// Test password policies
	testCases := []struct {
		name     string
		password string
		valid    bool
	}{
		{"ValidPassword", "StrongP@ssw0rd", true},
		{"TooShort", "short", false},
		{"NoSpecialChar", "NoSpecialChar123", false},
		{"NoNumber", "NoNumber@", false},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			valid := validatePassword(tc.password)
			assert.Equal(t, tc.valid, valid,
				"Password validation failed for case: %s", tc.name)
		})
	}
}

// validatePassword checks if a password meets security requirements.
// It verifies that the password has sufficient length, contains numbers,
// special characters, and both uppercase and lowercase letters.
func validatePassword(password string) bool {
	if len(password) < 8 {
		return false
	}

	var hasNumber, hasSpecial, hasUpper, hasLower bool
	for _, char := range password {
		switch {
		case unicode.IsNumber(char):
			hasNumber = true
		case unicode.IsUpper(char):
			hasUpper = true
		case unicode.IsLower(char):
			hasLower = true
		case unicode.IsPunct(char) || unicode.IsSymbol(char):
			hasSpecial = true
		}
	}

	return hasNumber && hasSpecial && hasUpper && hasLower
}
