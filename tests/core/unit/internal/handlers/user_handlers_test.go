package handlers_test

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/pageza/alchemorsel-v1/internal/dtos"
	"github.com/pageza/alchemorsel-v1/internal/handlers"
	"github.com/pageza/alchemorsel-v1/internal/models"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// MockUserService is a mock implementation of the UserService interface
type MockUserService struct {
	mock.Mock
}

func (m *MockUserService) CreateUser(ctx context.Context, user *models.User) error {
	args := m.Called(ctx, user)
	return args.Error(0)
}

func (m *MockUserService) GetUserByID(ctx context.Context, id string) (*models.User, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.User), args.Error(1)
}

func (m *MockUserService) GetUserByEmail(ctx context.Context, email string) (*models.User, error) {
	args := m.Called(ctx, email)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.User), args.Error(1)
}

func (m *MockUserService) DeleteUser(ctx context.Context, id string) error {
	args := m.Called(ctx, id)
	return args.Error(0)
}

func (m *MockUserService) ListUsers(ctx context.Context, page, limit int) ([]models.User, error) {
	args := m.Called(ctx, page, limit)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]models.User), args.Error(1)
}

func (m *MockUserService) VerifyPassword(ctx context.Context, email, password string) error {
	args := m.Called(ctx, email, password)
	return args.Error(0)
}

func (m *MockUserService) GenerateToken(ctx context.Context, user *models.User) (string, error) {
	args := m.Called(ctx, user)
	return args.String(0), args.Error(1)
}

func (m *MockUserService) Authenticate(ctx context.Context, email, password string) (*models.User, error) {
	args := m.Called(ctx, email, password)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.User), args.Error(1)
}

func (m *MockUserService) ForgotPassword(ctx context.Context, email string) error {
	args := m.Called(ctx, email)
	return args.Error(0)
}

func (m *MockUserService) ResetPassword(ctx context.Context, token string, newPassword string) error {
	args := m.Called(ctx, token, newPassword)
	return args.Error(0)
}

func (m *MockUserService) PatchUser(ctx context.Context, id string, updates map[string]interface{}) error {
	args := m.Called(ctx, id, updates)
	return args.Error(0)
}

func (m *MockUserService) GetAllUsers(ctx context.Context) ([]*models.User, error) {
	args := m.Called(ctx)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*models.User), args.Error(1)
}

func (m *MockUserService) GetUser(ctx context.Context, id string) (*models.User, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.User), args.Error(1)
}

func (m *MockUserService) UpdateUser(ctx context.Context, id string, user *models.User) error {
	args := m.Called(ctx, id, user)
	return args.Error(0)
}

func setupUserTest() (*handlers.UserHandler, *gin.Engine, *MockUserService) {
	gin.SetMode(gin.TestMode)
	mockService := new(MockUserService)
	handler := handlers.NewUserHandler(mockService)
	router := gin.New()

	// Set JWT secret for tests
	os.Setenv("JWT_SECRET", "test-secret")

	return handler, router, mockService
}

func TestLoginUser(t *testing.T) {
	handler, router, mockService := setupUserTest()
	router.POST("/login", handler.LoginUser)

	t.Run("successful login", func(t *testing.T) {
		mockUser := &models.User{
			ID:            "1",
			Name:          "Test User",
			Email:         "test@example.com",
			EmailVerified: true,
			CreatedAt:     time.Now(),
			UpdatedAt:     time.Now(),
		}

		loginReq := map[string]string{
			"email":    "test@example.com",
			"password": "password123",
		}

		mockService.On("Authenticate", mock.Anything, "test@example.com", "password123").
			Return(mockUser, nil)

		body, _ := json.Marshal(loginReq)
		w := httptest.NewRecorder()
		req, _ := http.NewRequest("POST", "/login", bytes.NewBuffer(body))
		req.Header.Set("Content-Type", "application/json")
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)

		var response map[string]interface{}
		err := json.Unmarshal(w.Body.Bytes(), &response)
		assert.NoError(t, err)
		assert.Contains(t, response, "token")
	})

	t.Run("invalid credentials", func(t *testing.T) {
		loginReq := map[string]string{
			"email":    "test@example.com",
			"password": "wrongpassword",
		}

		mockService.On("Authenticate", mock.Anything, "test@example.com", "wrongpassword").
			Return(nil, errors.New("user not found"))

		body, _ := json.Marshal(loginReq)
		w := httptest.NewRecorder()
		req, _ := http.NewRequest("POST", "/login", bytes.NewBuffer(body))
		req.Header.Set("Content-Type", "application/json")
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusUnauthorized, w.Code)

		var response dtos.ErrorResponse
		err := json.Unmarshal(w.Body.Bytes(), &response)
		assert.NoError(t, err)
		assert.Equal(t, "UNAUTHORIZED", response.Code)
		assert.Equal(t, "invalid email or password", response.Message)
	})

	t.Run("user not found", func(t *testing.T) {
		loginReq := map[string]string{
			"email":    "nonexistent@example.com",
			"password": "password123",
		}

		mockService.On("Authenticate", mock.Anything, "nonexistent@example.com", "password123").
			Return(nil, errors.New("user not found"))

		body, _ := json.Marshal(loginReq)
		w := httptest.NewRecorder()
		req, _ := http.NewRequest("POST", "/login", bytes.NewBuffer(body))
		req.Header.Set("Content-Type", "application/json")
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusUnauthorized, w.Code)

		var response dtos.ErrorResponse
		err := json.Unmarshal(w.Body.Bytes(), &response)
		assert.NoError(t, err)
		assert.Equal(t, "UNAUTHORIZED", response.Code)
		assert.Equal(t, "invalid email or password", response.Message)
	})

	t.Run("missing required fields", func(t *testing.T) {
		loginReq := map[string]string{
			"email":    "",
			"password": "",
		}

		body, _ := json.Marshal(loginReq)
		w := httptest.NewRecorder()
		req, _ := http.NewRequest("POST", "/login", bytes.NewBuffer(body))
		req.Header.Set("Content-Type", "application/json")
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusBadRequest, w.Code)

		var response dtos.ErrorResponse
		err := json.Unmarshal(w.Body.Bytes(), &response)
		assert.NoError(t, err)
		assert.Equal(t, "BAD_REQUEST", response.Code)
		assert.Equal(t, "email and password are required", response.Message)
	})
}

func TestCreateUser(t *testing.T) {
	handler, router, mockService := setupUserTest()
	router.POST("/users", handler.CreateUser)

	t.Run("successful user creation", func(t *testing.T) {
		userReq := map[string]string{
			"name":     "New User",
			"email":    "newuser@example.com",
			"password": "password123",
		}

		mockService.On("CreateUser", mock.Anything, mock.AnythingOfType("*models.User")).
			Return(nil)

		body, _ := json.Marshal(userReq)
		w := httptest.NewRecorder()
		req, _ := http.NewRequest("POST", "/users", bytes.NewBuffer(body))
		req.Header.Set("Content-Type", "application/json")
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusCreated, w.Code)

		var response dtos.UserResponse
		err := json.Unmarshal(w.Body.Bytes(), &response)
		assert.NoError(t, err)
		assert.Equal(t, "New User", response.Name)
		assert.Equal(t, "newuser@example.com", response.Email)
	})

	t.Run("invalid_request_body", func(t *testing.T) {
		// Setup
		mockUserService := new(MockUserService)
		handler := &handlers.UserHandler{
			Service: mockUserService,
		}
		router := gin.Default()
		router.POST("/v1/users", handler.CreateUser)

		// Invalid JSON request
		w := httptest.NewRecorder()
		req, _ := http.NewRequest("POST", "/v1/users", strings.NewReader("invalid json"))
		req.Header.Set("Content-Type", "application/json")
		router.ServeHTTP(w, req)

		// Assert
		assert.Equal(t, http.StatusBadRequest, w.Code)
		var response map[string]string
		err := json.Unmarshal(w.Body.Bytes(), &response)
		assert.NoError(t, err)
		assert.Equal(t, "BAD_REQUEST", response["code"])
		assert.Contains(t, response["message"], "Invalid request body")
	})

	t.Run("missing_required_fields", func(t *testing.T) {
		// Setup
		mockUserService := new(MockUserService)
		handler := &handlers.UserHandler{
			Service: mockUserService,
		}
		router := gin.Default()
		router.POST("/v1/users", handler.CreateUser)

		// Request with missing fields
		w := httptest.NewRecorder()
		req, _ := http.NewRequest("POST", "/v1/users", strings.NewReader(`{}`))
		req.Header.Set("Content-Type", "application/json")
		router.ServeHTTP(w, req)

		// Assert
		assert.Equal(t, http.StatusBadRequest, w.Code)
		var response map[string]string
		err := json.Unmarshal(w.Body.Bytes(), &response)
		assert.NoError(t, err)
		assert.Equal(t, "BAD_REQUEST", response["code"])
		assert.Contains(t, response["message"], "required")
	})
}

func TestGetUser(t *testing.T) {
	t.Run("successful_get_user", func(t *testing.T) {
		// Setup
		mockUserService := new(MockUserService)
		handler := &handlers.UserHandler{
			Service: mockUserService,
		}
		router := gin.Default()

		// Set up mock expectations
		mockUserService.On("GetUser", mock.Anything, "1").Return(&models.User{
			ID:    "1",
			Email: "test@example.com",
			Name:  "Test User",
		}, nil)

		router.GET("/v1/users/:id", handler.GetUser)

		// Create request
		w := httptest.NewRecorder()
		req, _ := http.NewRequest("GET", "/v1/users/1", nil)
		router.ServeHTTP(w, req)

		// Assert
		assert.Equal(t, 200, w.Code)
		mockUserService.AssertExpectations(t)
	})
}

func TestDeleteCurrentUser(t *testing.T) {
	handler, router, mockService := setupUserTest()
	router.DELETE("/users/:id", handler.DeleteCurrentUser)

	t.Run("successful delete user", func(t *testing.T) {
		mockService.On("DeleteUser", mock.Anything, "1").
			Return(nil)

		w := httptest.NewRecorder()
		req, _ := http.NewRequest("DELETE", "/users/1", nil)
		c, _ := gin.CreateTestContext(w)
		c.Request = req
		c.Set("currentUser", "1")
		handler.DeleteCurrentUser(c)

		assert.Equal(t, http.StatusOK, w.Code)
		var response gin.H
		err := json.Unmarshal(w.Body.Bytes(), &response)
		assert.NoError(t, err)
		assert.Equal(t, "user deleted successfully", response["message"])
	})

	t.Run("user not found", func(t *testing.T) {
		mockService.On("DeleteUser", mock.Anything, "999").
			Return(assert.AnError)

		w := httptest.NewRecorder()
		req, _ := http.NewRequest("DELETE", "/users/999", nil)
		c, _ := gin.CreateTestContext(w)
		c.Request = req
		c.Set("currentUser", "999")
		handler.DeleteCurrentUser(c)

		assert.Equal(t, http.StatusInternalServerError, w.Code)

		var response dtos.ErrorResponse
		err := json.Unmarshal(w.Body.Bytes(), &response)
		assert.NoError(t, err)
		assert.Equal(t, "INTERNAL_ERROR", response.Code)
		assert.Contains(t, response.Message, "Failed to delete user")
	})

	t.Run("unauthorized", func(t *testing.T) {
		w := httptest.NewRecorder()
		req, _ := http.NewRequest("DELETE", "/users/1", nil)
		c, _ := gin.CreateTestContext(w)
		c.Request = req
		handler.DeleteCurrentUser(c)

		assert.Equal(t, http.StatusUnauthorized, w.Code)

		var response dtos.ErrorResponse
		err := json.Unmarshal(w.Body.Bytes(), &response)
		assert.NoError(t, err)
		assert.Equal(t, "UNAUTHORIZED", response.Code)
		assert.Equal(t, "Unauthorized", response.Message)
	})
}

func TestGetCurrentUser(t *testing.T) {
	handler, router, mockService := setupUserTest()
	router.GET("/users/me", handler.GetCurrentUser)

	t.Run("successful get current user", func(t *testing.T) {
		mockUser := &models.User{
			ID:            "1",
			Name:          "Test User",
			Email:         "test@example.com",
			EmailVerified: true,
			CreatedAt:     time.Now(),
			UpdatedAt:     time.Now(),
		}

		mockService.On("GetUser", mock.Anything, "1").
			Return(mockUser, nil)

		w := httptest.NewRecorder()
		req, _ := http.NewRequest("GET", "/users/me", nil)
		c, _ := gin.CreateTestContext(w)
		c.Request = req
		c.Set("currentUser", "1")
		handler.GetCurrentUser(c)

		assert.Equal(t, http.StatusOK, w.Code)

		var response dtos.UserResponse
		err := json.Unmarshal(w.Body.Bytes(), &response)
		assert.NoError(t, err)
		assert.Equal(t, "Test User", response.Name)
		assert.Equal(t, "test@example.com", response.Email)
	})

	t.Run("unauthorized", func(t *testing.T) {
		w := httptest.NewRecorder()
		req, _ := http.NewRequest("GET", "/users/me", nil)
		c, _ := gin.CreateTestContext(w)
		c.Request = req
		handler.GetCurrentUser(c)

		assert.Equal(t, http.StatusUnauthorized, w.Code)

		var response dtos.ErrorResponse
		err := json.Unmarshal(w.Body.Bytes(), &response)
		assert.NoError(t, err)
		assert.Equal(t, "UNAUTHORIZED", response.Code)
		assert.Equal(t, "Unauthorized", response.Message)
	})

	t.Run("user not found", func(t *testing.T) {
		mockService.On("GetUser", mock.Anything, "999").
			Return(nil, assert.AnError)

		w := httptest.NewRecorder()
		req, _ := http.NewRequest("GET", "/users/me", nil)
		c, _ := gin.CreateTestContext(w)
		c.Request = req
		c.Set("currentUser", "999")
		handler.GetCurrentUser(c)

		assert.Equal(t, http.StatusInternalServerError, w.Code)

		var response dtos.ErrorResponse
		err := json.Unmarshal(w.Body.Bytes(), &response)
		assert.NoError(t, err)
		assert.Equal(t, "INTERNAL_ERROR", response.Code)
		assert.Contains(t, response.Message, "Failed to get user")
	})
}

func TestUpdateCurrentUser(t *testing.T) {
	handler, router, mockService := setupUserTest()
	router.PUT("/users/me", handler.UpdateCurrentUser)

	t.Run("successful update current user", func(t *testing.T) {
		userUpdate := &models.User{
			Name:  "Updated User",
			Email: "updated@example.com",
		}

		mockService.On("UpdateUser", mock.Anything, "1", mock.AnythingOfType("*models.User")).
			Return(nil)
		mockService.On("GetUser", mock.Anything, "1").Return(&models.User{
			ID:    "1",
			Name:  "Updated User",
			Email: "updated@example.com",
		}, nil)

		body, _ := json.Marshal(userUpdate)
		w := httptest.NewRecorder()
		req, _ := http.NewRequest("PUT", "/users/me", bytes.NewBuffer(body))
		req.Header.Set("Content-Type", "application/json")
		c, _ := gin.CreateTestContext(w)
		c.Request = req
		c.Set("currentUser", "1")
		handler.UpdateCurrentUser(c)

		assert.Equal(t, http.StatusOK, w.Code)

		var response dtos.UserResponse
		err := json.Unmarshal(w.Body.Bytes(), &response)
		assert.NoError(t, err)
		assert.Equal(t, "Updated User", response.Name)
		assert.Equal(t, "updated@example.com", response.Email)
	})

	t.Run("unauthorized", func(t *testing.T) {
		userUpdate := &models.User{
			Name:  "Updated User",
			Email: "updated@example.com",
		}

		body, _ := json.Marshal(userUpdate)
		w := httptest.NewRecorder()
		req, _ := http.NewRequest("PUT", "/users/me", bytes.NewBuffer(body))
		req.Header.Set("Content-Type", "application/json")
		c, _ := gin.CreateTestContext(w)
		c.Request = req
		handler.UpdateCurrentUser(c)

		assert.Equal(t, http.StatusUnauthorized, w.Code)

		var response dtos.ErrorResponse
		err := json.Unmarshal(w.Body.Bytes(), &response)
		assert.NoError(t, err)
		assert.Equal(t, "UNAUTHORIZED", response.Code)
		assert.Equal(t, "Unauthorized", response.Message)
	})

	t.Run("invalid request body", func(t *testing.T) {
		w := httptest.NewRecorder()
		req, _ := http.NewRequest("PUT", "/users/me", bytes.NewBufferString("invalid json"))
		req.Header.Set("Content-Type", "application/json")
		c, _ := gin.CreateTestContext(w)
		c.Request = req
		c.Set("currentUser", "1")
		handler.UpdateCurrentUser(c)

		assert.Equal(t, http.StatusBadRequest, w.Code)

		var response dtos.ErrorResponse
		err := json.Unmarshal(w.Body.Bytes(), &response)
		assert.NoError(t, err)
		assert.Equal(t, "BAD_REQUEST", response.Code)
		assert.Contains(t, response.Message, "Invalid request body")
	})
}

func TestPatchCurrentUser(t *testing.T) {
	handler, router, mockService := setupUserTest()
	router.PATCH("/users/me", handler.PatchCurrentUser)

	t.Run("successful patch current user", func(t *testing.T) {
		updates := map[string]interface{}{
			"name": "Patched User",
		}

		mockService.On("PatchUser", mock.Anything, "1", updates).
			Return(nil)

		mockUser := &models.User{
			ID:    "1",
			Name:  "Patched User",
			Email: "test@example.com",
		}

		mockService.On("GetUser", mock.Anything, "1").
			Return(mockUser, nil)

		body, _ := json.Marshal(updates)
		w := httptest.NewRecorder()
		req, _ := http.NewRequest("PATCH", "/users/me", bytes.NewBuffer(body))
		req.Header.Set("Content-Type", "application/json")
		c, _ := gin.CreateTestContext(w)
		c.Request = req
		c.Set("currentUser", "1")
		handler.PatchCurrentUser(c)

		assert.Equal(t, http.StatusOK, w.Code)

		var response map[string]interface{}
		err := json.Unmarshal(w.Body.Bytes(), &response)
		assert.NoError(t, err)
		user := response["user"].(map[string]interface{})
		assert.Equal(t, "Patched User", user["name"])
	})

	t.Run("unauthorized", func(t *testing.T) {
		updates := map[string]interface{}{
			"name": "Patched User",
		}

		body, _ := json.Marshal(updates)
		w := httptest.NewRecorder()
		req, _ := http.NewRequest("PATCH", "/users/me", bytes.NewBuffer(body))
		req.Header.Set("Content-Type", "application/json")
		c, _ := gin.CreateTestContext(w)
		c.Request = req
		handler.PatchCurrentUser(c)

		assert.Equal(t, http.StatusUnauthorized, w.Code)

		var response dtos.ErrorResponse
		err := json.Unmarshal(w.Body.Bytes(), &response)
		assert.NoError(t, err)
		assert.Equal(t, "UNAUTHORIZED", response.Code)
		assert.Equal(t, "Unauthorized", response.Message)
	})

	t.Run("invalid request body", func(t *testing.T) {
		w := httptest.NewRecorder()
		req, _ := http.NewRequest("PATCH", "/users/me", bytes.NewBufferString("invalid json"))
		req.Header.Set("Content-Type", "application/json")
		c, _ := gin.CreateTestContext(w)
		c.Request = req
		c.Set("currentUser", "1")
		handler.PatchCurrentUser(c)

		assert.Equal(t, http.StatusBadRequest, w.Code)

		var response dtos.ErrorResponse
		err := json.Unmarshal(w.Body.Bytes(), &response)
		assert.NoError(t, err)
		assert.Equal(t, "BAD_REQUEST", response.Code)
		assert.Contains(t, response.Message, "Invalid request body")
	})

	t.Run("patch_failed", func(t *testing.T) {
		// Reset the mock expectations and calls to ensure no previous calls interfere
		mockService.ExpectedCalls = nil
		mockService.Calls = nil
		updates := map[string]interface{}{
			"name": "Patched User",
		}
		mockService.On("PatchUser", mock.Anything, "1", mock.Anything).
			Return(assert.AnError)
		body, _ := json.Marshal(updates)
		w := httptest.NewRecorder()
		req, _ := http.NewRequest("PATCH", "/users/me", bytes.NewBuffer(body))
		req.Header.Set("Content-Type", "application/json")
		c, _ := gin.CreateTestContext(w)
		c.Request = req
		c.Set("currentUser", "1")
		handler.PatchCurrentUser(c)
		assert.Equal(t, http.StatusInternalServerError, w.Code)
		var response dtos.ErrorResponse
		err := json.Unmarshal(w.Body.Bytes(), &response)
		assert.NoError(t, err)
		assert.Equal(t, "INTERNAL_ERROR", response.Code)
		assert.Contains(t, response.Message, "Failed to update user")
	})
}

func TestForgotPassword(t *testing.T) {
	handler, router, mockService := setupUserTest()
	router.POST("/forgot-password", handler.ForgotPassword)

	t.Run("successful forgot password", func(t *testing.T) {
		mockService.On("ForgotPassword", mock.Anything, "test@example.com").
			Return(nil)

		body, _ := json.Marshal(map[string]string{
			"email": "test@example.com",
		})
		w := httptest.NewRecorder()
		req, _ := http.NewRequest("POST", "/forgot-password", bytes.NewBuffer(body))
		req.Header.Set("Content-Type", "application/json")
		c, _ := gin.CreateTestContext(w)
		c.Request = req
		handler.ForgotPassword(c)

		assert.Equal(t, http.StatusOK, w.Code)

		var response gin.H
		err := json.Unmarshal(w.Body.Bytes(), &response)
		assert.NoError(t, err)
		assert.Equal(t, "reset password instructions sent", response["message"])
	})

	t.Run("invalid request body", func(t *testing.T) {
		w := httptest.NewRecorder()
		req, _ := http.NewRequest("POST", "/forgot-password", bytes.NewBufferString("invalid json"))
		req.Header.Set("Content-Type", "application/json")
		c, _ := gin.CreateTestContext(w)
		c.Request = req
		handler.ForgotPassword(c)

		assert.Equal(t, http.StatusBadRequest, w.Code)

		var response dtos.ErrorResponse
		err := json.Unmarshal(w.Body.Bytes(), &response)
		assert.NoError(t, err)
		assert.Equal(t, "BAD_REQUEST", response.Code)
		assert.Contains(t, response.Message, "Invalid request body")
	})

	t.Run("missing email", func(t *testing.T) {
		body, _ := json.Marshal(map[string]string{})
		w := httptest.NewRecorder()
		req, _ := http.NewRequest("POST", "/forgot-password", bytes.NewBuffer(body))
		req.Header.Set("Content-Type", "application/json")
		c, _ := gin.CreateTestContext(w)
		c.Request = req
		handler.ForgotPassword(c)

		assert.Equal(t, http.StatusBadRequest, w.Code)

		var response dtos.ErrorResponse
		err := json.Unmarshal(w.Body.Bytes(), &response)
		assert.NoError(t, err)
		assert.Equal(t, "BAD_REQUEST", response.Code)
		assert.Contains(t, response.Message, "Email is required")
	})

	t.Run("service error", func(t *testing.T) {
		mockService.On("ForgotPassword", mock.Anything, "error@example.com").
			Return(assert.AnError)

		body, _ := json.Marshal(map[string]string{
			"email": "error@example.com",
		})
		w := httptest.NewRecorder()
		req, _ := http.NewRequest("POST", "/forgot-password", bytes.NewBuffer(body))
		req.Header.Set("Content-Type", "application/json")
		c, _ := gin.CreateTestContext(w)
		c.Request = req
		handler.ForgotPassword(c)

		assert.Equal(t, http.StatusInternalServerError, w.Code)

		var response dtos.ErrorResponse
		err := json.Unmarshal(w.Body.Bytes(), &response)
		assert.NoError(t, err)
		assert.Equal(t, "INTERNAL_ERROR", response.Code)
		assert.Contains(t, response.Message, "Failed to process forgot password request")
	})
}

func TestResetPassword(t *testing.T) {
	handler, router, mockService := setupUserTest()
	router.POST("/reset-password", handler.ResetPassword)

	t.Run("successful reset password", func(t *testing.T) {
		mockService.On("ResetPassword", mock.Anything, "valid-token", "newpassword123").
			Return(nil)

		body, _ := json.Marshal(map[string]string{
			"token":        "valid-token",
			"new_password": "newpassword123",
		})
		w := httptest.NewRecorder()
		req, _ := http.NewRequest("POST", "/reset-password", bytes.NewBuffer(body))
		req.Header.Set("Content-Type", "application/json")
		c, _ := gin.CreateTestContext(w)
		c.Request = req
		handler.ResetPassword(c)

		assert.Equal(t, http.StatusOK, w.Code)

		var response gin.H
		err := json.Unmarshal(w.Body.Bytes(), &response)
		assert.NoError(t, err)
		assert.Equal(t, "password reset successfully", response["message"])
	})

	t.Run("invalid request body", func(t *testing.T) {
		w := httptest.NewRecorder()
		req, _ := http.NewRequest("POST", "/reset-password", bytes.NewBufferString("invalid json"))
		req.Header.Set("Content-Type", "application/json")
		c, _ := gin.CreateTestContext(w)
		c.Request = req
		handler.ResetPassword(c)

		assert.Equal(t, http.StatusBadRequest, w.Code)

		var response dtos.ErrorResponse
		err := json.Unmarshal(w.Body.Bytes(), &response)
		assert.NoError(t, err)
		assert.Equal(t, "BAD_REQUEST", response.Code)
		assert.Contains(t, response.Message, "Invalid request body")
	})

	t.Run("missing token or password", func(t *testing.T) {
		body, _ := json.Marshal(map[string]string{
			"token": "valid-token",
			// Missing new_password
		})
		w := httptest.NewRecorder()
		req, _ := http.NewRequest("POST", "/reset-password", bytes.NewBuffer(body))
		req.Header.Set("Content-Type", "application/json")
		c, _ := gin.CreateTestContext(w)
		c.Request = req
		handler.ResetPassword(c)

		assert.Equal(t, http.StatusBadRequest, w.Code)

		var response dtos.ErrorResponse
		err := json.Unmarshal(w.Body.Bytes(), &response)
		assert.NoError(t, err)
		assert.Equal(t, "BAD_REQUEST", response.Code)
		assert.Contains(t, response.Message, "Missing token or new password")
	})

	t.Run("invalid token", func(t *testing.T) {
		mockService.On("ResetPassword", mock.Anything, "invalid-token", "newpassword123").
			Return(assert.AnError)

		body, _ := json.Marshal(map[string]string{
			"token":        "invalid-token",
			"new_password": "newpassword123",
		})
		w := httptest.NewRecorder()
		req, _ := http.NewRequest("POST", "/reset-password", bytes.NewBuffer(body))
		req.Header.Set("Content-Type", "application/json")
		c, _ := gin.CreateTestContext(w)
		c.Request = req
		handler.ResetPassword(c)

		assert.Equal(t, http.StatusInternalServerError, w.Code)

		var response dtos.ErrorResponse
		err := json.Unmarshal(w.Body.Bytes(), &response)
		assert.NoError(t, err)
		assert.Equal(t, "INTERNAL_ERROR", response.Code)
		assert.Contains(t, response.Message, "Failed to reset password")
	})
}

func TestVerifyEmail(t *testing.T) {
	handler, router, _ := setupUserTest()
	router.GET("/verify-email/:token", handler.VerifyEmail)

	t.Run("successful email verification", func(t *testing.T) {
		w := httptest.NewRecorder()
		req, _ := http.NewRequest("GET", "/verify-email/valid-token", nil)
		c, _ := gin.CreateTestContext(w)
		c.Request = req
		c.Params = []gin.Param{{Key: "token", Value: "valid-token"}}
		handler.VerifyEmail(c)

		assert.Equal(t, http.StatusOK, w.Code)

		var response gin.H
		err := json.Unmarshal(w.Body.Bytes(), &response)
		assert.NoError(t, err)
		assert.Equal(t, "email verified successfully", response["message"])
	})

	t.Run("missing token", func(t *testing.T) {
		w := httptest.NewRecorder()
		req, _ := http.NewRequest("GET", "/verify-email/", nil)
		c, _ := gin.CreateTestContext(w)
		c.Request = req
		handler.VerifyEmail(c)

		assert.Equal(t, http.StatusBadRequest, w.Code)

		var response dtos.ErrorResponse
		err := json.Unmarshal(w.Body.Bytes(), &response)
		assert.NoError(t, err)
		assert.Equal(t, "BAD_REQUEST", response.Code)
		assert.Equal(t, "Missing token", response.Message)
	})
}

func TestGetAllUsers(t *testing.T) {
	handler, router, mockService := setupUserTest()
	router.GET("/users", handler.GetAllUsers)

	t.Run("successful get all users", func(t *testing.T) {
		mockUsers := []*models.User{
			{
				ID:            "1",
				Name:          "User 1",
				Email:         "user1@example.com",
				EmailVerified: true,
				CreatedAt:     time.Now(),
				UpdatedAt:     time.Now(),
			},
			{
				ID:            "2",
				Name:          "User 2",
				Email:         "user2@example.com",
				EmailVerified: true,
				CreatedAt:     time.Now(),
				UpdatedAt:     time.Now(),
			},
		}

		mockService.On("GetAllUsers", mock.Anything).
			Return(mockUsers, nil)

		w := httptest.NewRecorder()
		req, _ := http.NewRequest("GET", "/users", nil)
		c, _ := gin.CreateTestContext(w)
		c.Request = req
		handler.GetAllUsers(c)

		assert.Equal(t, http.StatusOK, w.Code)

		var response map[string]interface{}
		err := json.Unmarshal(w.Body.Bytes(), &response)
		assert.NoError(t, err)
		users := response["users"].([]interface{})
		assert.Len(t, users, 2)
		user1 := users[0].(map[string]interface{})
		assert.Equal(t, "User 1", user1["name"])
		assert.Equal(t, "user1@example.com", user1["email"])
	})

	t.Run("service error", func(t *testing.T) {
		// Reset the mock expectations to avoid interference from previous subtests
		mockService.ExpectedCalls = nil
		mockService.On("GetAllUsers", mock.Anything).
			Return(nil, assert.AnError)

		w := httptest.NewRecorder()
		req, _ := http.NewRequest("GET", "/users", nil)
		c, _ := gin.CreateTestContext(w)
		c.Request = req
		handler.GetAllUsers(c)

		assert.Equal(t, http.StatusInternalServerError, w.Code)

		var response dtos.ErrorResponse
		err := json.Unmarshal(w.Body.Bytes(), &response)
		assert.NoError(t, err)
		assert.Equal(t, "INTERNAL_ERROR", response.Code)
		assert.Contains(t, response.Message, "Failed to get users")
	})
}
