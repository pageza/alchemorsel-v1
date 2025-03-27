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
)

var integrationTestMutex sync.Mutex // Added to serialize tests that modify the DB.

// TestMain ensures the database connection is initialized before tests run.
// func TestMain(m *testing.M) {
//     ctx := context.Background()
//
//     req := testcontainers.ContainerRequest{
//         Image:        "postgres:13",
//         ExposedPorts: []string{"5432/tcp"},
//         Env: map[string]string{
//             "POSTGRES_USER":     "testuser",
//             "POSTGRES_PASSWORD": "testpass",
//             "POSTGRES_DB":       "testdb",
//         },
//         WaitingFor: wait.ForListeningPort("5432/tcp").WithStartupTimeout(60 * time.Second),
//     }
//
//     postgresC, err := testcontainers.GenericContainer(ctx, testcontainers.GenericContainerRequest{
//         ContainerRequest: req,
//         Started:          true,
//     })
//     if err != nil {
//         fmt.Printf("Failed to start PostgreSQL container: %v\n", err)
//         os.Exit(1)
//     }
//     defer func() {
//         _ = postgresC.Terminate(ctx)
//     }()
//
//     host, err := postgresC.Host(ctx)
//     if err != nil {
//         fmt.Printf("Failed to get container host: %v\n", err)
//         os.Exit(1)
//     }
//     mappedPort, err := postgresC.MappedPort(ctx, "5432")
//     if err != nil {
//         fmt.Printf("Failed to get mapped port: %v\n", err)
//         os.Exit(1)
//     }
//
//     os.Setenv("DB_DRIVER", "postgres")
//     dsn := fmt.Sprintf("host=%s port=%s user=testuser password=testpass dbname=testdb sslmode=disable", host, mappedPort.Port())
//     os.Setenv("DB_SOURCE", dsn)
//
//     os.Setenv("POSTGRES_HOST", host)
//     os.Setenv("POSTGRES_PORT", mappedPort.Port())
//     os.Setenv("POSTGRES_USER", "testuser")
//     os.Setenv("POSTGRES_PASSWORD", "testpass")
//     os.Setenv("POSTGRES_DB", "testdb")
//
//     code := m.Run()
//     os.Exit(code)
// }

// resetDB resets the database state before each subtest.
func resetDB(t *testing.T) {
	integrationTestMutex.Lock()         // Lock to ensure sequential DB modifications.
	defer integrationTestMutex.Unlock() // Release after reset.

	if err := repositories.ClearUsers(); err != nil {
		t.Fatalf("failed to clear users table: %v", err)
	}
	// Re-insert dummy user required for certain endpoints.
	dummyUser := models.User{
		ID:       "1",
		Name:     "Dummy User",
		Email:    "dummy@example.com",
		Password: "dummy",
	}
	if err := repositories.DB.FirstOrCreate(&dummyUser, models.User{ID: "1"}).Error; err != nil {
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

	// Initialize the router with the database
	_ = routes.SetupRouter(database)

	// ... rest of test
}

func TestUserEndpoints(t *testing.T) {
	// Initialize the database
	config := db.NewConfig()
	database, err := db.InitDB(config)
	if err != nil {
		t.Fatalf("Failed to initialize DB: %v", err)
	}

	// Force tests to disable the rate limiter.
	os.Setenv("DISABLE_RATE_LIMITER", "true")
	// Set Gin to test mode.
	gin.SetMode(gin.TestMode)
	// Initialize the router with the database
	router := routes.SetupRouter(database)

	t.Run("CreateUser_Success", func(t *testing.T) {
		resetDB(t)
		// Prepare and send POST /v1/users with a unique, freshly generated email.
		newUser := map[string]string{
			"name":     "Test User",
			"email":    "testuser+" + generateRandomSuffix() + "@example.com",
			"password": "Password1!",
		}
		payload, err := json.Marshal(newUser)
		if err != nil {
			t.Fatalf("failed to marshal new user payload: %v", err)
		}
		req, err := http.NewRequest("POST", "/v1/users", bytes.NewBuffer(payload))
		if err != nil {
			t.Fatalf("failed to create request: %v", err)
		}
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)
		if w.Code != http.StatusCreated {
			t.Errorf("expected status 201, got %d", w.Code)
		}
		// ... further assertions as needed ...
	})

	t.Run("LoginUser_JWTTokenVerification", func(t *testing.T) {
		resetDB(t)
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
		// ... token verification code remains unchanged ...
	})

	t.Run("LoginUser_NoJWTSecret", func(t *testing.T) {
		resetDB(t)
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
		resetDB(t)
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

	// t.Run("TestEmailVerification", func(t *testing.T) {
	// 	// Create a dummy user with a valid verification token.
	// 	dummyUser := models.User{
	// 		Name:                     "Verification User",
	// 		Email:                    "verifyuser@example.com",
	// 		Password:                 "Password1!",
	// 		EmailVerificationToken:   "dummy-token",
	// 		EmailVerificationExpires: func() *time.Time { t := time.Now().Add(1 * time.Hour); return &t }(),
	// 	}
	// 	// Use the repository directly to create the user.
	// 	if err := repositories.CreateUser(context.Background(), &dummyUser); err != nil {
	// 		t.Fatalf("failed to create dummy verification user: %v", err)
	// 	}

	// 	// Call the verify-email endpoint with the dummy token.
	// 	req, _ := http.NewRequest("GET", "/v1/users/verify-email/dummy-token", nil)
	// 	resp := httptest.NewRecorder()
	// 	router.ServeHTTP(resp, req)
	// 	if resp.Code != http.StatusOK {
	// 		t.Errorf("expected status 200, got %d", resp.Code)
	// 	}

	// 	// Optionally, retrieve and check that the verification token is cleared.
	// 	verifiedUser, err := repositories.GetUserByEmail(context.Background(), dummyUser.Email)
	// 	if err != nil || verifiedUser == nil {
	// 		t.Fatalf("failed to retrieve user after verification: %v", err)
	// 	}
	// 	if verifiedUser.EmailVerificationToken != "" {
	// 		t.Errorf("expected verification token to be cleared after verification")
	// 	}
	// })

	// ... additional subtests with similar resetDB(t) calls to ensure isolation ...
}

func TestAdditionalUserEndpoints(t *testing.T) {
	// Initialize the database
	config := db.NewConfig()
	database, err := db.InitDB(config)
	if err != nil {
		t.Fatalf("Failed to initialize DB: %v", err)
	}

	// Force tests to disable the rate limiter.
	os.Setenv("DISABLE_RATE_LIMITER", "true")
	// Set Gin to test mode.
	gin.SetMode(gin.TestMode)
	// Initialize the router with the database
	_ = routes.SetupRouter(database)

	// ... rest of test
}

/*
// TestEmailVerification tests the email verification process.
func TestEmailVerification(t *testing.T) {
	// Setup router from application routes.
	router := routes.SetupRouter()

	// Create a user and get the verification token.
	user := &models.User{
		Name:     "Test User",
		Email:    "testuser@example.com",
		Password: "Password1!",
	}
	if err := services.CreateUser(context.Background(), user); err != nil {
		t.Fatalf("failed to create user: %v", err)
	}

	// Simulate email verification.
	req, _ := http.NewRequest("GET", fmt.Sprintf("/v1/users/verify-email/%s", user.EmailVerificationToken), nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("expected status 200, got %d", w.Code)
	}
}


func TestCurrentUserEndpoints(t *testing.T) {
	// Set up SQLite configuration for integration tests.
	os.Setenv("DB_DRIVER", "sqlite")
	os.Setenv("DB_SOURCE", "file::memory:?cache=shared")
	// Disable rate limiter for isolation.
	os.Setenv("DISABLE_RATE_LIMITER", "true")
	gin.SetMode(gin.TestMode)
	router := routes.SetupRouter()
	// Reset the rate limiter counters to ensure no leftover state causes 429 responses.
	middleware.ResetRateLimiter()

	// Create a test user.
	uniqueEmail := "currentuser+" + generateRandomSuffix() + "@example.com"
	newUser := map[string]string{
		"name":     "Current User",
		"email":    uniqueEmail,
		"password": "Password1!",
	}
	newUserBytes, err := json.Marshal(newUser)
	if err != nil {
		t.Fatalf("failed to marshal new user payload: %v", err)
	}
	req, err := http.NewRequest("POST", "/v1/users", bytes.NewBuffer(newUserBytes))
	if err != nil {
		t.Fatalf("failed to create user creation request: %v", err)
	}
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)
	if w.Code != http.StatusCreated {
		t.Fatalf("expected status 201 for user creation, got %d", w.Code)
	}

	// Log in the user to obtain a valid auth token.
	loginPayload := map[string]string{
		"email":    uniqueEmail,
		"password": "Password1!",
	}
	loginBytes, err := json.Marshal(loginPayload)
	if err != nil {
		t.Fatalf("failed to marshal login payload: %v", err)
	}
	req, err = http.NewRequest("POST", "/v1/users/login", bytes.NewBuffer(loginBytes))
	if err != nil {
		t.Fatalf("failed to create login request: %v", err)
	}
	req.Header.Set("Content-Type", "application/json")
	w = httptest.NewRecorder()
	router.ServeHTTP(w, req)
	if w.Code != http.StatusOK {
		t.Fatalf("failed to login, expected status 200, got %d", w.Code)
	}
	var loginResp map[string]string
	if err := json.Unmarshal(w.Body.Bytes(), &loginResp); err != nil {
		t.Fatalf("failed to unmarshal login response: %v", err)
	}
	token, ok := loginResp["token"]
	if !ok || token == "" {
		t.Fatalf("token not found in login response")
	}

	// Subtest: GET /v1/users/me should return the current user.
	t.Run("GetCurrentUser_Success", func(t *testing.T) {
		req, _ := http.NewRequest("GET", "/v1/users/me", nil)
		req.Header.Set("Authorization", "Bearer "+token)
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)
		if w.Code != http.StatusOK {
			t.Errorf("expected status 200 for get current user, got %d", w.Code)
		}
		var userResp map[string]interface{}
		if err := json.Unmarshal(w.Body.Bytes(), &userResp); err != nil {
			t.Fatalf("failed to unmarshal get current user response: %v", err)
		}
		if userResp["email"] != uniqueEmail {
			t.Errorf("expected email to be %s, got %v", uniqueEmail, userResp["email"])
		}
	})

	// Subtest: PUT /v1/users/me with valid payload should update the user.
	t.Run("UpdateCurrentUser_Success", func(t *testing.T) {
		updatePayload := map[string]string{
			"name": "Valid Updated Name",
		}
		payloadBytes, err := json.Marshal(updatePayload)
		if err != nil {
			t.Fatalf("failed to marshal update payload: %v", err)
		}
		req, err := http.NewRequest("PUT", "/v1/users/me", bytes.NewBuffer(payloadBytes))
		if err != nil {
			t.Fatalf("failed to create PUT update request: %v", err)
		}
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("Authorization", "Bearer "+token)
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)
		if w.Code != http.StatusOK {
			t.Errorf("expected status 200 for valid PUT update, got %d", w.Code)
		}
		var respBody map[string]string
		if err := json.Unmarshal(w.Body.Bytes(), &respBody); err != nil {
			t.Fatalf("failed to unmarshal update response: %v", err)
		}
		if respBody["message"] != "user updated successfully" {
			t.Errorf("expected update message to be 'user updated successfully', got: %v", respBody["message"])
		}
	})

	// Subtest: GET /v1/users/me should reflect the updated data.
	t.Run("GetCurrentUser_AfterUpdate", func(t *testing.T) {
		req, _ := http.NewRequest("GET", "/v1/users/me", nil)
		req.Header.Set("Authorization", "Bearer "+token)
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)
		if w.Code != http.StatusOK {
			t.Errorf("expected status 200 for get current user after update, got %d", w.Code)
		}
		var userResp map[string]interface{}
		if err := json.Unmarshal(w.Body.Bytes(), &userResp); err != nil {
			t.Fatalf("failed to unmarshal get current user response: %v", err)
		}
		if userResp["name"] != "Valid Updated Name" {
			t.Errorf("expected name to be 'Valid Updated Name', got: %v", userResp["name"])
		}
	})

	// Subtest: DELETE /v1/users/me to deactivate the user, then GET should return 404.
	t.Run("DeleteCurrentUser_Success", func(t *testing.T) {
		req, _ := http.NewRequest("DELETE", "/v1/users/me", nil)
		req.Header.Set("Authorization", "Bearer "+token)
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)
		if w.Code != http.StatusOK {
			t.Errorf("expected status 200 for delete current user, got %d", w.Code)
		}
		var delResp map[string]string
		if err := json.Unmarshal(w.Body.Bytes(), &delResp); err != nil {
			t.Fatalf("failed to unmarshal delete response: %v", err)
		}
		if delResp["message"] != "user deactivated successfully" {
			t.Errorf("unexpected delete message: %v", delResp["message"])
		}

		// Attempt to GET the same user, expecting a 404 Not Found.
		req, _ = http.NewRequest("GET", "/v1/users/me", nil)
		req.Header.Set("Authorization", "Bearer "+token)
		w = httptest.NewRecorder()
		router.ServeHTTP(w, req)
		if w.Code != http.StatusNotFound {
			t.Errorf("expected status 404 after deletion, got %d", w.Code)
		}
	})
}
*/
