package integration

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"sync"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/pageza/alchemorsel-v1/internal/db"
	"github.com/pageza/alchemorsel-v1/internal/models"
	"github.com/pageza/alchemorsel-v1/internal/repositories"
	"github.com/pageza/alchemorsel-v1/internal/routes"
	"gorm.io/gorm"
)

var integrationTestMutex sync.Mutex // Added to serialize tests that modify the DB.

// resetDB resets the database state before each subtest.
func resetDB(t *testing.T, database *gorm.DB) {
	integrationTestMutex.Lock()         // Lock to ensure sequential DB modifications.
	defer integrationTestMutex.Unlock() // Release after reset.

	if err := database.Exec("DELETE FROM users").Error; err != nil {
		t.Fatalf("failed to clear users table: %v", err)
	}
	// Re-insert dummy user required for certain endpoints.
	dummyUser := models.User{
		ID:       "1",
		Name:     "Dummy User",
		Email:    "dummy@example.com",
		Password: "dummy",
	}
	if err := database.FirstOrCreate(&dummyUser, models.User{ID: "1"}).Error; err != nil {
		t.Fatalf("failed to insert dummy user: %v", err)
	}
}

// generateRandomSuffix returns a unique suffix for test email addresses.
func generateRandomSuffix() string {
	return fmt.Sprintf("%d", time.Now().UnixNano())
}

func TestIntegrationUser(t *testing.T) {
	// Initialize the database
	config := db.NewConfig()
	database, err := db.InitDB(config)
	if err != nil {
		t.Fatalf("Failed to initialize DB: %v", err)
	}

	// Create test logger
	logger := createTestLogger()

	// Initialize the router with the database and logger
	_ = routes.SetupRouter(database, logger)

	// ... rest of test
}

func setupTestDB(t *testing.T) (*gin.Engine, *gorm.DB) {
	// Set environment variables for test
	os.Setenv("DB_DRIVER", "postgres")
	os.Setenv("TEST_MODE", "true")
	os.Setenv("DISABLE_RATE_LIMITER", "true")

	// Initialize the database
	config := db.NewConfig()
	database, err := db.InitDB(config)
	if err != nil {
		t.Fatalf("Failed to initialize DB: %v", err)
	}

	// Run migrations
	if err := repositories.RunMigrations(database); err != nil {
		t.Fatalf("Failed to run migrations: %v", err)
	}

	// Set Gin to test mode
	gin.SetMode(gin.TestMode)

	// Create test logger
	logger := createTestLogger()

	// Initialize the router with the database and logger
	return routes.SetupRouter(database, logger), database
}

func TestUserEndpoints(t *testing.T) {
	router, database := setupTestDB(t)

	t.Run("CreateUser_Success", func(t *testing.T) {
		resetDB(t, database)
		// Test user creation
		req, _ := http.NewRequest("POST", "/v1/users", nil)
		resp := httptest.NewRecorder()
		router.ServeHTTP(resp, req)

		if resp.Code != http.StatusCreated {
			t.Errorf("Expected status 201, got %d", resp.Code)
		}
	})

	t.Run("LoginUser_JWTTokenVerification", func(t *testing.T) {
		resetDB(t, database)
		// Prepare a new user payload with a unique email.
		newUser := map[string]string{
			"name":     "Test User",
			"email":    "testuser+" + generateRandomSuffix() + "@example.com",
			"password": "Password1!",
		}
		// Create user and then perform login tests.
		payload, _ := json.Marshal(newUser)
		req1, _ := http.NewRequest("POST", "/v1/users", bytes.NewBuffer(payload))
		req1.Header.Set("Content-Type", "application/json")
		w1 := httptest.NewRecorder()
		router.ServeHTTP(w1, req1)
		if w1.Code != http.StatusCreated {
			t.Fatalf("failed to create user for login, got status %d", w1.Code)
		}

		loginPayload := map[string]string{
			"email":    "testuser+" + generateRandomSuffix() + "@example.com",
			"password": "Password1!",
		}
		payload, err := json.Marshal(loginPayload)
		if err != nil {
			t.Fatalf("failed to marshal login payload: %v", err)
		}
		req, err := http.NewRequest("POST", "/v1/users/login", bytes.NewBuffer(payload))
		if err != nil {
			t.Fatalf("failed to create login request: %v", err)
		}
		req.Header.Set("Content-Type", "application/json")
		resp := httptest.NewRecorder()
		router.ServeHTTP(resp, req)
		if resp.Code != http.StatusOK {
			t.Fatalf("expected status %d for login, got %d", http.StatusOK, resp.Code)
		}
	})

	t.Run("LoginUser_NoJWTSecret", func(t *testing.T) {
		resetDB(t, database)
		originalSecret := os.Getenv("JWT_SECRET")
		os.Setenv("JWT_SECRET", "")
		defer os.Setenv("JWT_SECRET", originalSecret)

		// Ensure the user exists.
		newUser := map[string]string{
			"name":     "Test User",
			"email":    "testuser@example.com",
			"password": "Password1!",
		}
		payload, _ := json.Marshal(newUser)
		req1, _ := http.NewRequest("POST", "/v1/users", bytes.NewBuffer(payload))
		req1.Header.Set("Content-Type", "application/json")
		w1 := httptest.NewRecorder()
		router.ServeHTTP(w1, req1)
		if w1.Code != http.StatusCreated {
			t.Fatalf("failed to create user for login, got status %d", w1.Code)
		}

		loginPayload := map[string]string{
			"email":    "testuser@example.com",
			"password": "Password1!",
		}
		payload, err := json.Marshal(loginPayload)
		if err != nil {
			t.Fatalf("failed to marshal login payload: %v", err)
		}
		req, err := http.NewRequest("POST", "/v1/users/login", bytes.NewBuffer(payload))
		if err != nil {
			t.Fatalf("failed to create login request: %v", err)
		}
		req.Header.Set("Content-Type", "application/json")
		resp := httptest.NewRecorder()
		router.ServeHTTP(resp, req)
		if resp.Code != http.StatusInternalServerError {
			t.Errorf("expected status %d when JWT secret is missing, got %d", http.StatusInternalServerError, resp.Code)
		}
	})

	t.Run("GetUser_NotFound", func(t *testing.T) {
		resetDB(t, database)
		req, err := http.NewRequest("GET", "/v1/users/nonexistent-id", nil)
		if err != nil {
			t.Fatalf("failed to create GET request: %v", err)
		}
		resp := httptest.NewRecorder()
		router.ServeHTTP(resp, req)
		if resp.Code != http.StatusNotFound {
			t.Errorf("expected status %d for non-existent user, got %d", http.StatusNotFound, resp.Code)
		}
	})
}

func TestAdditionalUserEndpoints(t *testing.T) {
	// Initialize the database
	config := db.NewConfig()
	database, err := db.InitDB(config)
	if err != nil {
		t.Fatalf("Failed to initialize DB: %v", err)
	}

	// Create test logger
	logger := createTestLogger()

	// Initialize the router with the database and logger
	_ = routes.SetupRouter(database, logger)

	// ... rest of test
}
