package integration

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"sync"
	"testing"
	"time"

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
		Name:     "Dummy User",
		Email:    "dummy@example.com",
		Password: "Password123!",
	}
	userRepo := repositories.NewUserRepository(database)
	if err := userRepo.CreateUser(context.Background(), &dummyUser); err != nil {
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

	// Create test logger and Redis client
	logger := createTestLogger()
	redisClient := createTestRedisClient()

	// Initialize the router with the database, logger, and Redis client
	_ = routes.SetupRouter(database, logger, redisClient)

	// ... rest of test
}

func TestUserEndpoints(t *testing.T) {
	os.Setenv("JWT_SECRET", "testsecret")
	os.Setenv("DISABLE_RATE_LIMITER", "true")

	// Setup test environment
	router, database := setupTestEnvironment(t)
	defer database.Migrator().DropTable(&models.User{})

	t.Run("CreateUser_Success", func(t *testing.T) {
		resetDB(t, database)
		// Test user creation with required fields
		newUser := models.User{
			Name:     "Test User",
			Email:    "testuser+" + generateRandomSuffix() + "@example.com",
			Password: "Password123!",
		}
		payload, _ := json.Marshal(newUser)
		req, _ := http.NewRequest("POST", "/v1/users", strings.NewReader(string(payload)))
		req.Header.Set("Content-Type", "application/json")
		resp := httptest.NewRecorder()
		router.ServeHTTP(resp, req)

		if resp.Code != http.StatusCreated {
			var errResp map[string]string
			if err := json.NewDecoder(resp.Body).Decode(&errResp); err == nil {
				t.Errorf("Expected status 201, got %d with error: %s", resp.Code, errResp["error"])
			} else {
				t.Errorf("Expected status 201, got %d", resp.Code)
			}
		}
	})

	t.Run("LoginUser_JWTTokenVerification", func(t *testing.T) {
		resetDB(t, database)
		// Prepare a new user payload with a unique email.
		email := "testuser+" + generateRandomSuffix() + "@example.com"
		newUser := models.User{
			Name:     "Test User",
			Email:    email,
			Password: "Password123!",
		}
		// Create user and then perform login tests.
		payload, _ := json.Marshal(newUser)
		req1, _ := http.NewRequest("POST", "/v1/users", strings.NewReader(string(payload)))
		req1.Header.Set("Content-Type", "application/json")
		w1 := httptest.NewRecorder()
		router.ServeHTTP(w1, req1)
		if w1.Code != http.StatusCreated {
			var errResp map[string]string
			if err := json.NewDecoder(w1.Body).Decode(&errResp); err == nil {
				t.Fatalf("failed to create user for login, got status %d with error: %s", w1.Code, errResp["error"])
			} else {
				t.Fatalf("failed to create user for login, got status %d", w1.Code)
			}
		}

		loginPayload := models.LoginRequest{
			Email:    email,
			Password: "Password123!",
		}
		payload, err := json.Marshal(loginPayload)
		if err != nil {
			t.Fatalf("failed to marshal login payload: %v", err)
		}
		req, _ := http.NewRequest("POST", "/v1/users/login", strings.NewReader(string(payload)))
		req.Header.Set("Content-Type", "application/json")
		resp := httptest.NewRecorder()
		router.ServeHTTP(resp, req)
		if resp.Code != http.StatusOK {
			var errResp map[string]string
			if err := json.NewDecoder(resp.Body).Decode(&errResp); err == nil {
				t.Fatalf("expected status %d for login, got %d with error: %s", http.StatusOK, resp.Code, errResp["error"])
			} else {
				t.Fatalf("expected status %d for login, got %d", http.StatusOK, resp.Code)
			}
		}
	})

	t.Run("LoginUser_NoJWTSecret", func(t *testing.T) {
		resetDB(t, database)
		originalSecret := os.Getenv("JWT_SECRET")
		os.Setenv("JWT_SECRET", "")
		defer os.Setenv("JWT_SECRET", originalSecret)

		// Ensure the user exists.
		newUser := models.User{
			Name:     "Test User",
			Email:    "testuser@example.com",
			Password: "Password123!",
		}
		payload, _ := json.Marshal(newUser)
		req1, _ := http.NewRequest("POST", "/v1/users", strings.NewReader(string(payload)))
		req1.Header.Set("Content-Type", "application/json")
		w1 := httptest.NewRecorder()
		router.ServeHTTP(w1, req1)
		if w1.Code != http.StatusCreated {
			var errResp map[string]string
			if err := json.NewDecoder(w1.Body).Decode(&errResp); err == nil {
				t.Fatalf("failed to create user for login, got status %d with error: %s", w1.Code, errResp["error"])
			} else {
				t.Fatalf("failed to create user for login, got status %d", w1.Code)
			}
		}

		loginPayload := models.LoginRequest{
			Email:    "testuser@example.com",
			Password: "Password123!",
		}
		payload, err := json.Marshal(loginPayload)
		if err != nil {
			t.Fatalf("failed to marshal login payload: %v", err)
		}
		req, _ := http.NewRequest("POST", "/v1/users/login", strings.NewReader(string(payload)))
		req.Header.Set("Content-Type", "application/json")
		resp := httptest.NewRecorder()
		router.ServeHTTP(resp, req)
		if resp.Code != http.StatusInternalServerError {
			var errResp map[string]string
			if err := json.NewDecoder(resp.Body).Decode(&errResp); err == nil {
				t.Errorf("expected status %d when JWT secret is missing, got %d with error: %s", http.StatusInternalServerError, resp.Code, errResp["error"])
			} else {
				t.Errorf("expected status %d when JWT secret is missing, got %d", http.StatusInternalServerError, resp.Code)
			}
		}
	})

	t.Run("GetUser_NotFound", func(t *testing.T) {
		resetDB(t, database)
		req, err := http.NewRequest("GET", "/v1/users/123e4567-e89b-12d3-a456-426614174001", nil)
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

	// Create test logger and Redis client
	logger := createTestLogger()
	redisClient := createTestRedisClient()

	// Initialize the router with the database, logger, and Redis client
	_ = routes.SetupRouter(database, logger, redisClient)

	// ... rest of test
}
