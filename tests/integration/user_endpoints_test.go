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

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v4"
	"github.com/joho/godotenv"
	"github.com/pageza/alchemorsel-v1/internal/db"
	"github.com/pageza/alchemorsel-v1/internal/models"
	"github.com/pageza/alchemorsel-v1/internal/routes"
)

// TestMain ensures the database connection is initialized before tests run.
func TestMain(m *testing.M) {
	// Load configuration from .env.test file for testing.
	if err := godotenv.Load(".env.test"); err != nil {
		// If .env.test is not found, relying on environment variables.
	}

	// Ensure required DB environment variables are set; otherwise, skip tests.
	if os.Getenv("POSTGRES_HOST") == "" ||
		os.Getenv("POSTGRES_PORT") == "" ||
		os.Getenv("POSTGRES_USER") == "" ||
		os.Getenv("POSTGRES_PASSWORD") == "" ||
		os.Getenv("POSTGRES_DB") == "" {
		// Skip tests if the DB is not configured.
		os.Exit(0)
	}

	// Initialize the database connection using environment variables.
	if err := db.Init(); err != nil {
		os.Exit(1)
	}
	os.Exit(m.Run())
}

func TestUserEndpoints(t *testing.T) {
	// Set Gin to test mode.
	gin.SetMode(gin.TestMode)

	// Setup router from the application's routes.
	router := routes.SetupRouter()

	// ------------------------------
	// Test the CreateUser endpoint.
	// ------------------------------
	t.Run("CreateUser_Success", func(t *testing.T) {
		// Prepare a new user payload.
		newUser := models.User{
			ID:       "test-user-1",
			Name:     "Test User",
			Email:    "testuser@example.com",
			Password: "password",
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

		resp := httptest.NewRecorder()
		router.ServeHTTP(resp, req)

		if resp.Code != http.StatusCreated {
			t.Errorf("expected status %d, got %d", http.StatusCreated, resp.Code)
		}

		// Decode the response to obtain the returned user.
		var responseBody map[string]models.User
		if err := json.Unmarshal(resp.Body.Bytes(), &responseBody); err != nil {
			t.Errorf("failed to unmarshal response: %v", err)
		}

		userResp, exists := responseBody["user"]
		if !exists {
			t.Error("response does not contain 'user' field")
		}
		if userResp.Email != newUser.Email {
			t.Errorf("expected email %s, got %s", newUser.Email, userResp.Email)
		}
		if userResp.Password != "" {
			t.Error("expected password to be omitted in the response")
		}
	})

	// ---------------------------------
	// Test the LoginUser endpoint (happy path).
	// ---------------------------------
	t.Run("LoginUser_Success", func(t *testing.T) {
		// Login using the same credentials as the user created above.
		loginPayload := map[string]string{
			"email":    "testuser@example.com",
			"password": "password",
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
			t.Errorf("expected status %d, got %d", http.StatusOK, resp.Code)
		}

		// Decode the login response.
		var responseBody map[string]string
		if err := json.Unmarshal(resp.Body.Bytes(), &responseBody); err != nil {
			t.Errorf("failed to unmarshal login response: %v", err)
		}
		token, exists := responseBody["token"]
		if !exists || token == "" {
			t.Error("expected a non-empty token in the response")
		}
	})

	// -------------------------------
	// Test the LoginUser endpoint (invalid credentials).
	// -------------------------------
	t.Run("LoginUser_InvalidCredentials", func(t *testing.T) {
		loginPayload := map[string]string{
			"email":    "testuser@example.com",
			"password": "wrongpassword",
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

		if resp.Code != http.StatusUnauthorized {
			t.Errorf("expected status %d, got %d", http.StatusUnauthorized, resp.Code)
		}
	})

	// ----------------------------------
	// Test CreateUser endpoint with invalid JSON.
	// ----------------------------------
	t.Run("CreateUser_InvalidJSON", func(t *testing.T) {
		req, err := http.NewRequest("POST", "/v1/users", bytes.NewBuffer([]byte(`{"invalid_json":`)))
		if err != nil {
			t.Fatalf("failed to create request: %v", err)
		}
		req.Header.Set("Content-Type", "application/json")

		resp := httptest.NewRecorder()
		router.ServeHTTP(resp, req)

		if resp.Code != http.StatusBadRequest {
			t.Errorf("expected status %d, got %d", http.StatusBadRequest, resp.Code)
		}
	})

	// ----------------------------------
	// Test LoginUser endpoint with invalid JSON.
	// ----------------------------------
	t.Run("LoginUser_InvalidJSON", func(t *testing.T) {
		req, err := http.NewRequest("POST", "/v1/users/login", bytes.NewBuffer([]byte(`{"invalid_json":`)))
		if err != nil {
			t.Fatalf("failed to create request: %v", err)
		}
		req.Header.Set("Content-Type", "application/json")

		resp := httptest.NewRecorder()
		router.ServeHTTP(resp, req)

		if resp.Code != http.StatusBadRequest {
			t.Errorf("expected status %d, got %d", http.StatusBadRequest, resp.Code)
		}
	})

	// ----------------------------------
	// Test duplicate user registration.
	// ----------------------------------
	t.Run("CreateUser_Duplicate", func(t *testing.T) {
		newUser := models.User{
			ID:       "test-user-dup",
			Name:     "Test User Dup",
			Email:    "testdup@example.com",
			Password: "password",
		}
		payload, err := json.Marshal(newUser)
		if err != nil {
			t.Fatalf("failed to marshal new user payload: %v", err)
		}

		// First registration should succeed.
		req, err := http.NewRequest("POST", "/v1/users", bytes.NewBuffer(payload))
		if err != nil {
			t.Fatalf("failed to create request: %v", err)
		}
		req.Header.Set("Content-Type", "application/json")

		resp := httptest.NewRecorder()
		router.ServeHTTP(resp, req)
		if resp.Code != http.StatusCreated {
			t.Errorf("expected status %d for first registration, got %d", http.StatusCreated, resp.Code)
		}

		// Attempt duplicate registration with the same payload.
		req, err = http.NewRequest("POST", "/v1/users", bytes.NewBuffer(payload))
		if err != nil {
			t.Fatalf("failed to create duplicate request: %v", err)
		}
		req.Header.Set("Content-Type", "application/json")

		resp = httptest.NewRecorder()
		router.ServeHTTP(resp, req)
		if resp.Code != http.StatusConflict {
			t.Errorf("expected status %d for duplicate registration, got %d", http.StatusConflict, resp.Code)
		}
	})

	// ----------------------------------
	// Test user registration with missing fields.
	// ----------------------------------
	t.Run("CreateUser_MissingFields", func(t *testing.T) {
		// Payload missing the required "Email" field.
		newUser := map[string]string{
			"ID":       "test-user-missing",
			"Name":     "Test Missing",
			"Password": "password",
		}
		payload, err := json.Marshal(newUser)
		if err != nil {
			t.Fatalf("failed to marshal payload: %v", err)
		}

		req, err := http.NewRequest("POST", "/v1/users", bytes.NewBuffer(payload))
		if err != nil {
			t.Fatalf("failed to create request: %v", err)
		}
		req.Header.Set("Content-Type", "application/json")

		resp := httptest.NewRecorder()
		router.ServeHTTP(resp, req)
		if resp.Code != http.StatusBadRequest {
			t.Errorf("expected status %d for missing fields, got %d", http.StatusBadRequest, resp.Code)
		}
	})

	// ----------------------------------
	// Test login with missing required fields.
	// ----------------------------------
	t.Run("LoginUser_MissingFields", func(t *testing.T) {
		// Payload missing the required "password" field.
		loginPayload := map[string]string{
			"email": "testuser@example.com",
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
		if resp.Code != http.StatusBadRequest {
			t.Errorf("expected status %d for missing login fields, got %d", http.StatusBadRequest, resp.Code)
		}
	})

	// ----------------------------------
	// Test CreateUser endpoint with an invalid email format.
	// ----------------------------------
	t.Run("CreateUser_InvalidEmail", func(t *testing.T) {
		// Payload with improperly formatted email.
		newUser := map[string]string{
			"ID":       "test-user-invalid-email",
			"Name":     "Invalid Email",
			"Email":    "not-an-email", // invalid format
			"Password": "password",
		}
		payload, err := json.Marshal(newUser)
		if err != nil {
			t.Fatalf("failed to marshal payload: %v", err)
		}

		req, err := http.NewRequest("POST", "/v1/users", bytes.NewBuffer(payload))
		if err != nil {
			t.Fatalf("failed to create request: %v", err)
		}
		req.Header.Set("Content-Type", "application/json")

		resp := httptest.NewRecorder()
		router.ServeHTTP(resp, req)

		if resp.Code != http.StatusBadRequest {
			t.Errorf("expected status %d for invalid email format, got %d", http.StatusBadRequest, resp.Code)
		}
	})

	// ----------------------------------
	// Test that the JWT token returned in LoginUser_Success is valid and has expected claims.
	// ----------------------------------
	t.Run("LoginUser_JWTTokenVerification", func(t *testing.T) {
		loginPayload := map[string]string{
			"email":    "testuser@example.com",
			"password": "password",
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

		var responseBody map[string]string
		if err := json.Unmarshal(resp.Body.Bytes(), &responseBody); err != nil {
			t.Fatalf("failed to unmarshal response: %v", err)
		}
		tokenString, exists := responseBody["token"]
		if !exists || tokenString == "" {
			t.Fatal("expected a non-empty token in the response")
		}

		token, err := jwt.Parse(tokenString, func(t *jwt.Token) (interface{}, error) {
			// Ensure the signing method is HMAC.
			if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
				return nil, fmt.Errorf("unexpected signing method: %v", t.Header["alg"])
			}
			return []byte(os.Getenv("JWT_SECRET")), nil
		})
		if err != nil {
			t.Errorf("failed to parse token: %v", err)
		}
		if !token.Valid {
			t.Error("token is invalid")
		}
		claims, ok := token.Claims.(jwt.MapClaims)
		if !ok {
			t.Error("failed to cast token claims")
		}
		if claims["user_id"] == nil || claims["email"] == nil {
			t.Error("expected claims 'user_id' and 'email' in token")
		}
	})

	// ----------------------------------
	// Test concurrent duplicate registration requests.
	// ----------------------------------
	t.Run("CreateUser_ConcurrentDuplicate", func(t *testing.T) {
		newUser := models.User{
			ID:       "test-user-concurrent",
			Name:     "Concurrent User",
			Email:    "testconcurrent@example.com",
			Password: "password",
		}
		payload, err := json.Marshal(newUser)
		if err != nil {
			t.Fatalf("failed to marshal payload: %v", err)
		}

		numRoutines := 5
		var wg sync.WaitGroup
		resultCh := make(chan int, numRoutines)

		wg.Add(numRoutines)
		for i := 0; i < numRoutines; i++ {
			go func() {
				defer wg.Done()
				req, err := http.NewRequest("POST", "/v1/users", bytes.NewBuffer(payload))
				if err != nil {
					resultCh <- 0
					return
				}
				req.Header.Set("Content-Type", "application/json")
				resp := httptest.NewRecorder()
				router.ServeHTTP(resp, req)
				resultCh <- resp.Code
			}()
		}
		wg.Wait()
		close(resultCh)

		createdCount, conflictCount := 0, 0
		for code := range resultCh {
			if code == http.StatusCreated {
				createdCount++
			} else if code == http.StatusConflict {
				conflictCount++
			}
		}

		if createdCount != 1 {
			t.Errorf("expected exactly one creation, got %d", createdCount)
		}
		if conflictCount != numRoutines-1 {
			t.Errorf("expected %d conflict responses, got %d", numRoutines-1, conflictCount)
		}
	})

	// ----------------------------------
	// Test login when JWT_SECRET is not set.
	// ----------------------------------
	t.Run("LoginUser_NoJWTSecret", func(t *testing.T) {
		originalSecret := os.Getenv("JWT_SECRET")
		os.Setenv("JWT_SECRET", "")
		defer os.Setenv("JWT_SECRET", originalSecret)

		loginPayload := map[string]string{
			"email":    "testuser@example.com",
			"password": "password",
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

		// Expect an internal server error because the JWT secret is missing.
		if resp.Code != http.StatusInternalServerError {
			t.Errorf("expected status %d when JWT secret is missing, got %d", http.StatusInternalServerError, resp.Code)
		}
	})

	// ----------------------------------
	// Test GetUser endpoint for a non-existent user.
	// ----------------------------------
	t.Run("GetUser_NotFound", func(t *testing.T) {
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
