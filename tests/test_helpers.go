package testhelpers

import (
	"bytes"
	"encoding/json"
	"log"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v4"
	"github.com/pageza/alchemorsel-v1/internal/db"
	"github.com/pageza/alchemorsel-v1/internal/dtos"
	"github.com/pageza/alchemorsel-v1/internal/logging"
	"github.com/pageza/alchemorsel-v1/internal/repositories"
	"github.com/pageza/alchemorsel-v1/internal/routes"
	"github.com/stretchr/testify/assert"
	"gorm.io/gorm"
)

// SetupTestRouter creates a new Gin engine in TestMode and returns it.
func SetupTestRouter() *gin.Engine {
	gin.SetMode(gin.TestMode)
	return gin.New()
}

// MakeTestRequest is a helper function to perform an HTTP request against the provided router.
// The request body is marshaled as JSON if provided and the Content-Type header is set accordingly.
func MakeTestRequest(router *gin.Engine, method, path string, body interface{}) (*httptest.ResponseRecorder, error) {
	var reqBody *bytes.Buffer
	if body != nil {
		jsonData, err := json.Marshal(body)
		if err != nil {
			return nil, err
		}
		reqBody = bytes.NewBuffer(jsonData)
	} else {
		reqBody = &bytes.Buffer{}
	}

	req, err := http.NewRequest(method, path, reqBody)
	if err != nil {
		return nil, err
	}
	if body != nil {
		req.Header.Set("Content-Type", "application/json")
	}

	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)
	return w, nil
}

// GetTestDataDir returns the directory path to store test files. It ensures that the 'tests/.testdata' directory exists.
// This directory should be used to store files generated during tests so that they do not clutter the project root.
func GetTestDataDir(t *testing.T) string {
	testDataDir := filepath.Join("tests", ".testdata")
	err := os.MkdirAll(testDataDir, os.ModePerm)
	if err != nil {
		t.Fatalf("Failed to create test data directory: %v", err)
	}
	return testDataDir
}

// GenerateTestToken returns a valid JWT token for testing purposes.
func GenerateTestToken(claims interface{}) string {
	var tokenClaims jwt.MapClaims
	if claims == nil {
		tokenClaims = jwt.MapClaims{
			"sub": "test-user",
			"exp": time.Now().Add(time.Hour).Unix(),
			"iat": time.Now().Unix(),
		}
	} else {
		tokenClaims = claims.(jwt.MapClaims)
	}
	secret := os.Getenv("JWT_SECRET")
	if secret == "" {
		secret = "test-secret"
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, tokenClaims)
	tokenString, err := token.SignedString([]byte(secret))
	if err != nil {
		return ""
	}
	return tokenString
}

// AssertErrorResponse asserts that the HTTP response recorder contains a JSON error response
// matching the expected status, error code, and error message.
func AssertErrorResponse(t *testing.T, rr *httptest.ResponseRecorder, expectedStatus int, expectedCode, expectedMessage string) {
	assert.Equal(t, expectedStatus, rr.Code, "Unexpected HTTP status code")
	var errResp dtos.ErrorResponse
	err := json.Unmarshal(rr.Body.Bytes(), &errResp)
	assert.NoError(t, err, "Response body should be valid JSON")
	assert.Equal(t, expectedCode, errResp.Code, "Error code does not match")
	assert.Equal(t, expectedMessage, errResp.Message, "Error message does not match")
}

// CreateTestLogger creates and returns a logger configured for tests.
func CreateTestLogger() *logging.Logger {
	config := logging.LogConfig{
		LogDir:            "logs",
		MaxSize:           10,
		MaxBackups:        3,
		MaxAge:            28,
		Compress:          false,
		LogLevel:          "debug",
		RequestIDHeader:   "X-Request-ID",
		LogFormat:         "json",
		EnableConsole:     true,
		EnableFile:        false,
		EnableElastic:     false,
		ElasticURL:        "",
		ElasticIndex:      "",
		EnableCompression: false,
	}
	logger, err := logging.NewLogger(config)
	if err != nil {
		panic("Failed to create test logger: " + err.Error())
	}
	return logger
}

// SetupIntegrationRouter sets up and returns a Gin router configured for integration tests.
func SetupIntegrationRouter(t *testing.T) (*gin.Engine, *gorm.DB) {
	os.Setenv("JWT_SECRET", "testsecret")
	os.Setenv("DISABLE_RATE_LIMITER", "true")

	config := db.NewConfig()
	database, err := db.InitDB(config)
	if err != nil {
		t.Fatalf("Failed to initialize DB: %v", err)
	}

	logger := CreateTestLogger()
	redisClient := setupTestRedis()

	router := routes.SetupRouter(database, logger, redisClient)
	return router, database
}

func setupTestRedis() *repositories.RedisClient {
	// For tests, try to connect to Redis, but don't fail if it's not available
	redisAddr := os.Getenv("REDIS_ADDR")
	if redisAddr == "" {
		redisAddr = "localhost:6379"
	}

	redisClient, err := repositories.NewRedisClient(redisAddr)
	if err != nil {
		log.Printf("Warning: Failed to connect to Redis: %v. Using nil client for tests.", err)
		return nil
	}
	return redisClient
}

// setupTestEnvironment is a helper function to set up all test dependencies
func SetupTestEnvironment(t *testing.T) (*gin.Engine, *gorm.DB) {
	os.Setenv("JWT_SECRET", "testsecret")
	os.Setenv("DISABLE_RATE_LIMITER", "true")

	config := db.NewConfig()
	database, err := db.InitDB(config)
	if err != nil {
		t.Fatalf("Failed to initialize DB: %v", err)
	}

	if err := repositories.RunMigrations(database); err != nil {
		t.Fatalf("Failed to run migrations: %v", err)
	}

	logger := CreateTestLogger()
	redisClient := setupTestRedis()

	router := routes.SetupRouter(database, logger, redisClient)
	return router, database
}
