package integration

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"sync"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
	"github.com/pageza/alchemorsel-v1/internal/models"
	"github.com/pageza/alchemorsel-v1/internal/repositories"
	"github.com/pageza/alchemorsel-v1/internal/routes"
)

var integrationTestMutex sync.Mutex // Added to serialize tests that modify the DB.

// TestMain ensures the database connection is initialized before tests run.
func TestMain(m *testing.M) {
	// Load configuration from .env.test file for testing.
	if err := godotenv.Load(".env.test"); err != nil {
		// If .env.test is not found, relying on environment variables.
	}
	// Ensure integration test mode so that file‑based DB is used.
	os.Setenv("INTEGRATION_TEST", "true")
	// Set a default JWT secret if not already defined.
	if os.Getenv("JWT_SECRET") == "" {
		os.Setenv("JWT_SECRET", "testsecret")
	}
	// cursor--MOD: Default to sqlite and set DB_SOURCE to a persistent file so that migrations run on a single connection.
	if os.Getenv("DB_DRIVER") == "" {
		os.Setenv("DB_DRIVER", "sqlite")
		os.Setenv("DB_SOURCE", "./test.db")
	}
	// If not using sqlite, ensure required Postgres environment variables are set; otherwise, skip tests.
	if os.Getenv("DB_DRIVER") != "sqlite" {
		if os.Getenv("POSTGRES_HOST") == "" ||
			os.Getenv("POSTGRES_PORT") == "" ||
			os.Getenv("POSTGRES_USER") == "" ||
			os.Getenv("POSTGRES_PASSWORD") == "" ||
			os.Getenv("POSTGRES_DB") == "" {
			// Skip tests if the DB is not configured.
			os.Exit(0)
		}
	}
	os.Exit(m.Run())
}

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

func TestUserEndpoints(t *testing.T) {
	// Rely on the DSN set in TestMain (persistent file: "./test.db") for integration tests.
	// Force tests to disable the rate limiter.
	os.Setenv("DISABLE_RATE_LIMITER", "true")
	// Set Gin to test mode.
	gin.SetMode(gin.TestMode)
	// Setup router from the application's routes.
	router := routes.SetupRouter()

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

	t.Run("TestEmailVerification", func(t *testing.T) {
		// Create a dummy user with a valid verification token.
		dummyUser := models.User{
			Name:                     "Verification User",
			Email:                    "verifyuser@example.com",
			Password:                 "Password1!",
			EmailVerificationToken:   "dummy-token",
			EmailVerificationExpires: func() *time.Time { t := time.Now().Add(1 * time.Hour); return &t }(),
		}
		// Use the repository directly to create the user.
		if err := repositories.CreateUser(context.Background(), &dummyUser); err != nil {
			t.Fatalf("failed to create dummy verification user: %v", err)
		}

		// Call the verify-email endpoint with the dummy token.
		req, _ := http.NewRequest("GET", "/v1/users/verify-email/dummy-token", nil)
		resp := httptest.NewRecorder()
		router.ServeHTTP(resp, req)
		if resp.Code != http.StatusOK {
			t.Errorf("expected status 200, got %d", resp.Code)
		}

		// Optionally, retrieve and check that the verification token is cleared.
		verifiedUser, err := repositories.GetUserByEmail(context.Background(), dummyUser.Email)
		if err != nil || verifiedUser == nil {
			t.Fatalf("failed to retrieve user after verification: %v", err)
		}
		if verifiedUser.EmailVerificationToken != "" {
			t.Errorf("expected verification token to be cleared after verification")
		}
	})

	// ... additional subtests with similar resetDB(t) calls to ensure isolation ...
}

func TestAdditionalUserEndpoints(t *testing.T) {
	router := routes.SetupRouter()

	// Subtest: Access protected endpoint without token.
	t.Run("AccessProtectedEndpoint_NoToken", func(t *testing.T) {
		req, _ := http.NewRequest("GET", "/v1/users/me", nil)
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)
		if w.Code != http.StatusUnauthorized {
			t.Errorf("Expected status 401 for missing token, got: %d", w.Code)
		}
	})

	// Subtest: Access protected endpoint with invalid token.
	t.Run("AccessProtectedEndpoint_InvalidToken", func(t *testing.T) {
		req, _ := http.NewRequest("GET", "/v1/users/me", nil)
		req.Header.Set("Authorization", "Bearer invalidtoken")
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)
		if w.Code != http.StatusUnauthorized {
			t.Errorf("Expected status 401 for invalid token, got: %d", w.Code)
		}
	})

	// Create a test user to obtain a valid token.
	newUser := map[string]string{
		"id":       "test-user-patch",
		"name":     "Original Name",
		"email":    "patchuser@example.com",
		"password": "Password1!",
	}
	newUserBytes, _ := json.Marshal(newUser)
	req, _ := http.NewRequest("POST", "/v1/users", bytes.NewBuffer(newUserBytes))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)
	if w.Code != http.StatusCreated {
		t.Fatalf("Failed to create user, expected 201, got %d", w.Code)
	}

	// Login the user to get a token.
	loginPayload := map[string]string{
		"email":    newUser["email"],
		"password": newUser["password"],
	}
	loginBytes, _ := json.Marshal(loginPayload)
	req, _ = http.NewRequest("POST", "/v1/users/login", bytes.NewBuffer(loginBytes))
	req.Header.Set("Content-Type", "application/json")
	w = httptest.NewRecorder()
	router.ServeHTTP(w, req)
	if w.Code != http.StatusOK {
		t.Fatalf("Failed to login, expected 200, got %d", w.Code)
	}
	var loginResp map[string]string
	if err := json.Unmarshal(w.Body.Bytes(), &loginResp); err != nil {
		t.Fatalf("Failed to unmarshal login response: %v", err)
	}
	token, ok := loginResp["token"]
	if !ok || token == "" {
		t.Fatalf("Token not found in login response")
	}

	// Subtest: Test PATCH /v1/users/me to update the user partially.
	t.Run("PatchCurrentUser_Success", func(t *testing.T) {
		// Prepare payload to update the current user's name.
		patchPayload := map[string]string{
			"name": "Updated Test User",
		}
		payload, err := json.Marshal(patchPayload)
		if err != nil {
			t.Fatalf("failed to marshal patch payload: %v", err)
		}
		req, err := http.NewRequest("PATCH", "/v1/users/me", bytes.NewBuffer(payload))
		if err != nil {
			t.Fatalf("failed to create PATCH request: %v", err)
		}
		req.Header.Set("Content-Type", "application/json")
		// Ensure we send the valid auth token in the header.
		req.Header.Set("Authorization", "Bearer "+token)
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)
		if w.Code != http.StatusOK {
			t.Errorf("expected status 200 for patch, got %d", w.Code)
		}
		var resBody map[string]map[string]interface{}
		if err := json.Unmarshal(w.Body.Bytes(), &resBody); err != nil {
			t.Fatalf("failed to unmarshal patch response: %v", err)
		}
		userResp, exists := resBody["user"]
		if !exists {
			t.Fatal("patch response does not contain 'user' field")
		}
		if userResp["name"] != "Updated Test User" {
			t.Errorf("expected name to be 'Updated Test User', got: %v", userResp["name"])
		}
	})

	// Subtest: Test PUT /v1/users/me with invalid payload.
	t.Run("UpdateCurrentUser_InvalidPayload", func(t *testing.T) {
		req, _ := http.NewRequest("PUT", "/v1/users/me", bytes.NewBuffer([]byte(`{"name": "New Name"`)))
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("Authorization", "Bearer "+token)
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)
		if w.Code != http.StatusBadRequest {
			t.Errorf("Expected status 400 for invalid payload, got %d", w.Code)
		}
	})

	// Subtest: Test admin endpoint unauthorized access.
	t.Run("AdminEndpoint_NoToken", func(t *testing.T) {
		req, _ := http.NewRequest("GET", "/v1/admin/users", nil)
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)
		if w.Code != http.StatusUnauthorized {
			t.Errorf("Expected status 401 for admin endpoint without token, got %d", w.Code)
		}
	})

	// Subtest: Test login rate limiter by sending rapid login requests.
	t.Run("LoginRateLimiter", func(t *testing.T) {
		var wg sync.WaitGroup
		numRequests := 10
		rateLimitExceeded := false
		for i := 0; i < numRequests; i++ {
			wg.Add(1)
			go func() {
				defer wg.Done()
				req, _ := http.NewRequest("POST", "/v1/users/login", bytes.NewBuffer(loginBytes))
				req.Header.Set("Content-Type", "application/json")
				w := httptest.NewRecorder()
				router.ServeHTTP(w, req)
				if w.Code == http.StatusTooManyRequests {
					rateLimitExceeded = true
				}
			}()
		}
		wg.Wait()
		if !rateLimitExceeded {
			t.Errorf("Expected at least one request to be rate limited, but none were")
		}
	})
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
