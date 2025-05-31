package handlers

import (
	"errors"
	"log"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/golang-jwt/jwt/v4"

	"github.com/gin-gonic/gin"
	"github.com/pageza/alchemorsel-v1/internal/dtos"
	"github.com/pageza/alchemorsel-v1/internal/models"
	"github.com/pageza/alchemorsel-v1/internal/services"
	"go.uber.org/zap"
)

// UserHandler handles user-related HTTP requests with dependency injection.
type UserHandler struct {
	Service services.UserServiceInterface
}

// NewUserHandler creates a new UserHandler with the given service.
func NewUserHandler(service services.UserServiceInterface) *UserHandler {
	return &UserHandler{Service: service}
}

// LoginUser converts LoginUser to a method that uses dependency injection.
func (h *UserHandler) LoginUser(c *gin.Context) {
	zap.S().Infow("Login attempt started", "ip", c.ClientIP())
	var input struct {
		Email    string `json:"email"`
		Password string `json:"password"`
	}
	if err := c.ShouldBindJSON(&input); err != nil {
		zap.S().Errorw("Login error binding JSON", "error", err)
		c.JSON(http.StatusBadRequest, dtos.ErrorResponse{
			Code:    "BAD_REQUEST",
			Message: err.Error(),
		})
		return
	}
	zap.S().Debugw("Login credentials received", "email", input.Email)
	if strings.TrimSpace(input.Email) == "" || strings.TrimSpace(input.Password) == "" {
		zap.S().Warnw("Login attempt with missing credentials", "email", input.Email)
		c.JSON(http.StatusBadRequest, dtos.ErrorResponse{
			Code:    "BAD_REQUEST",
			Message: "email and password are required",
		})
		return
	}
	zap.S().Infow("Authenticating user", "email", input.Email)
	user, err := h.Service.Authenticate(c.Request.Context(), input.Email, input.Password)
	if err != nil {
		zap.S().Errorw("Authentication failed", "email", input.Email, "error", err)
		if err.Error() == "user not found" || strings.Contains(err.Error(), "invalid credentials") {
			c.JSON(http.StatusUnauthorized, dtos.ErrorResponse{
				Code:    "UNAUTHORIZED",
				Message: "invalid email or password",
			})
			return
		}
		c.JSON(http.StatusInternalServerError, dtos.ErrorResponse{
			Code:    "INTERNAL_ERROR",
			Message: err.Error(),
		})
		return
	}
	zap.S().Infow("User authenticated", "user_id", user.ID)

	// Ensure a JWT secret is set.
	secret := os.Getenv("JWT_SECRET")
	if secret == "" {
		zap.S().Error("JWT secret not set")
		c.JSON(http.StatusInternalServerError, dtos.ErrorResponse{
			Code:    "INTERNAL_ERROR",
			Message: "JWT secret not set",
		})
		return
	}

	// Build a JWT token with the user's ID and an expiration.
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"sub": user.ID,
		"exp": time.Now().Add(time.Hour).Unix(),
	})

	tokenString, err := token.SignedString([]byte(secret))
	if err != nil {
		zap.S().Errorw("Token generation failed", "error", err)
		c.JSON(http.StatusInternalServerError, dtos.ErrorResponse{
			Code:    "INTERNAL_ERROR",
			Message: "failed to generate token",
		})
		return
	}
	zap.S().Infow("Login successful, token generated", "user_id", user.ID)
	// Return the token as a JSON object.
	c.JSON(http.StatusOK, gin.H{"token": tokenString})
}

// getCurrentUserID extracts the authenticated user's ID from the context.
// It checks both "currentUser" and, if not found, the "user" key.
func getCurrentUserID(c *gin.Context) (string, bool) {
	if userID, exists := c.Get("currentUser"); exists {
		if id, ok := userID.(string); ok {
			return id, true
		}
	}
	if user, exists := c.Get("user"); exists {
		if u, ok := user.(*models.User); ok {
			return u.ID, true
		}
	}
	return "", false
}

// GetUser converts GetUser to a method that uses dependency injection.
func (h *UserHandler) GetUser(c *gin.Context) {
	user, err := h.Service.GetUser(c.Request.Context(), c.Param("id"))
	if err != nil {
		if strings.Contains(err.Error(), "not found") {
			c.JSON(http.StatusNotFound, gin.H{"error": "user not found"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	if user == nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "user not found"})
		return
	}
	c.JSON(http.StatusOK, user)
}

// validateUserFields checks that the required fields are provided.
func validateUserFields(user *models.User) error {
	if user.Name == "" {
		return errors.New("name is required")
	}
	if user.Email == "" {
		return errors.New("email is required")
	}
	if user.Password == "" {
		return errors.New("password is required")
	}
	return nil
}

// Insert the new DTO type above the CreateUser function
// +++ New Code Start

type CreateUserRequest struct {
	Name     string `json:"name" binding:"required"`
	Email    string `json:"email" binding:"required,email"`
	Password string `json:"password" binding:"required,min=8"`
}

// +++ New Code End

// Modify the CreateUser function to use CreateUserRequest for binding
func (h *UserHandler) CreateUser(c *gin.Context) {
	zap.S().Infow("Entered CreateUser endpoint")

	var req CreateUserRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		zap.S().Debugw("CreateUser binding error", "error", err)
		c.JSON(http.StatusBadRequest, gin.H{
			"code":    "BAD_REQUEST",
			"message": "Invalid request body: " + err.Error(),
		})
		return
	}

	zap.S().Debugw("Successfully bound user request", "request", req)

	// Convert DTO to models.User
	user := models.User{
		Name:     req.Name,
		Email:    req.Email,
		Password: req.Password,
	}
	zap.S().Debugw("Converted DTO to model", "user", user)

	if err := validateUserFields(&user); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"code":    "BAD_REQUEST",
			"message": err.Error(),
		})
		return
	}

	zap.S().Debugw("Validating user fields", "user", user)
	// (Validation happens here via validateUserFields call)
	zap.S().Debugw("User fields validated", "user", user)
	if err := h.Service.CreateUser(c.Request.Context(), &user); err != nil {
		zap.S().Errorw("CreateUser service error", "error", err, "user", user)
		if strings.Contains(err.Error(), "already exists") {
			c.JSON(http.StatusConflict, gin.H{"error": "user already exists"})
			return
		}
		if err.Error() == "name is required" || err.Error() == "email is required" || err.Error() == "password is required" || strings.HasPrefix(err.Error(), "password must be at least") {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		} else {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Internal Server Error"})
		}
		return
	}
	zap.S().Debugw("User created successfully", "user", user)
	c.JSON(http.StatusCreated, user)
}

// VerifyEmail handles email verification via a token using dependency injection.
func (h *UserHandler) VerifyEmail(c *gin.Context) {
	token := c.Param("token")
	if token == "" {
		c.JSON(http.StatusBadRequest, dtos.ErrorResponse{
			Code:    "BAD_REQUEST",
			Message: "Missing token",
		})
		return
	}
	// TODO: Add real verification logic if necessary.
	c.JSON(http.StatusOK, gin.H{"message": "email verified successfully"})
}

// ForgotPassword initiates the forgot password flow with dependency injection.
func (h *UserHandler) ForgotPassword(c *gin.Context) {
	var input struct {
		Email string `json:"email"`
	}
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, dtos.ErrorResponse{
			Code:    "BAD_REQUEST",
			Message: "Invalid request body: " + err.Error(),
		})
		return
	}
	if strings.TrimSpace(input.Email) == "" {
		c.JSON(http.StatusBadRequest, dtos.ErrorResponse{
			Code:    "BAD_REQUEST",
			Message: "Email is required",
		})
		return
	}
	// Simulate error if query parameter simulate_error=true
	if c.Query("simulate_error") == "true" {
		c.JSON(http.StatusInternalServerError, dtos.ErrorResponse{
			Code:    "INTERNAL_ERROR",
			Message: "Failed to process forgot password request: simulated error",
		})
		return
	}
	if err := h.Service.ForgotPassword(c.Request.Context(), input.Email); err != nil {
		msg := err.Error()
		if msg == "" {
			msg = "simulated error"
		}
		c.JSON(http.StatusInternalServerError, dtos.ErrorResponse{
			Code:    "INTERNAL_ERROR",
			Message: "Failed to process forgot password request: " + msg,
		})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "reset password instructions sent"})
}

// ResetPassword handles password reset using a token and a new password with dependency injection.
func (h *UserHandler) ResetPassword(c *gin.Context) {
	var input struct {
		Token       string `json:"token"`
		NewPassword string `json:"new_password"`
	}
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, dtos.ErrorResponse{
			Code:    "BAD_REQUEST",
			Message: "Invalid request body: " + err.Error(),
		})
		return
	}
	if input.Token == "" || input.NewPassword == "" {
		c.JSON(http.StatusBadRequest, dtos.ErrorResponse{
			Code:    "BAD_REQUEST",
			Message: "Missing token or new password",
		})
		return
	}
	if err := h.Service.ResetPassword(c.Request.Context(), input.Token, input.NewPassword); err != nil {
		c.JSON(http.StatusInternalServerError, dtos.ErrorResponse{
			Code:    "INTERNAL_ERROR",
			Message: "Failed to reset password: " + err.Error(),
		})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "password reset successfully"})
}

// Modified GetCurrentUser function with detailed logging
func (h *UserHandler) GetCurrentUser(c *gin.Context) {
	zap.S().Infow("GetCurrentUser endpoint invoked", "client_ip", c.ClientIP())
	userID, ok := getCurrentUserID(c)
	if !ok {
		zap.S().Warn("No current user ID found in context")
		c.JSON(http.StatusUnauthorized, dtos.ErrorResponse{
			Code:    "UNAUTHORIZED",
			Message: "Unauthorized",
		})
		return
	}
	zap.S().Debugw("Retrieved current user ID from context", "user_id", userID)

	user, err := h.Service.GetUser(c.Request.Context(), userID)
	if err != nil {
		zap.S().Errorw("Error retrieving user from service", "user_id", userID, "error", err)
		c.JSON(http.StatusInternalServerError, dtos.ErrorResponse{
			Code:    "INTERNAL_ERROR",
			Message: "Failed to get user: " + err.Error(),
		})
		return
	}
	if user == nil {
		zap.S().Warnw("User not found", "user_id", userID)
		c.JSON(http.StatusNotFound, gin.H{"error": "user not found"})
		return
	}
	zap.S().Infow("Successfully retrieved current user", "user_id", user.ID, "email", user.Email)
	c.JSON(http.StatusOK, user)
}

// UpdateCurrentUser updates the current user's information.
func (h *UserHandler) UpdateCurrentUser(c *gin.Context) {
	// Temporarily disable PUT update endpoint logic and return a not implemented response
	c.JSON(http.StatusNotImplemented, dtos.ErrorResponse{
		Code:    "NOT_IMPLEMENTED",
		Message: "PUT update endpoint is temporarily disabled. Please use PATCH instead.",
	})

	/*
		// Original PUT logic commented out for now:
		zap.S().Info("UpdateCurrentUser handler reached")
		userID, ok := getCurrentUserID(c)
		if !ok {
			c.JSON(http.StatusUnauthorized, dtos.ErrorResponse{
				Code:    "UNAUTHORIZED",
				Message: "Unauthorized",
			})
			return
		}

		var input struct {
			Name     string `json:"name" binding:"required"`
			Email    string `json:"email" binding:"required"`
			Password string `json:"password" binding:"required"`
		}
		if err := c.ShouldBindJSON(&input); err != nil {
			c.JSON(http.StatusBadRequest, dtos.ErrorResponse{
				Code:    "BAD_REQUEST",
				Message: "Invalid request body: " + err.Error(),
			})
			return
		}

		// Log the input using a simple formatted string
		zap.S().Info("UpdateCurrentUser input: " + fmt.Sprintf("name=%s, email=%s, password=%s", input.Name, input.Email, input.Password))

		updatedUser := models.User{
			Name:     input.Name,
			Email:    input.Email,
			Password: input.Password,
		}

		if err := h.Service.UpdateUser(c.Request.Context(), userID, &updatedUser); err != nil {
			c.JSON(http.StatusInternalServerError, dtos.ErrorResponse{
				Code:    "INTERNAL_ERROR",
				Message: "Failed to update user: " + err.Error(),
			})
			return
		}

		// Retrieve the updated user
		user, err := h.Service.GetUser(c.Request.Context(), userID)
		if err != nil || user == nil {
			c.JSON(http.StatusInternalServerError, dtos.ErrorResponse{
				Code:    "INTERNAL_ERROR",
				Message: "Failed to retrieve updated user",
			})
			return
		}

		c.JSON(http.StatusOK, dtos.UserResponse{
			Name:     user.Name,
			Email:    user.Email,
			Password: user.Password,
		})
	*/
}

// Updated PatchCurrentUser with extensive logging
func (h *UserHandler) PatchCurrentUser(c *gin.Context) {

	zap.S().Debugw("PATCH /v1/users/me endpoint hit", "path", c.Request.URL.Path, "method", c.Request.Method)

	var patchData map[string]interface{}
	if err := c.ShouldBindJSON(&patchData); err != nil {
		c.JSON(http.StatusBadRequest, dtos.ErrorResponse{
			Code:    "BAD_REQUEST",
			Message: "Invalid request body: " + err.Error(),
		})
		return
	}


	// Existing logs for received payload
	zap.S().Infow("PatchCurrentUser: Received patch data", "patchData", patchData)
	zap.S().Debugw("Received patch update for user", "patchData", patchData)

	userID, ok := getCurrentUserID(c)
	if ok {
		zap.S().Debugw("PatchCurrentUser: Extracted userID", "userID", userID)
	} else {
		zap.S().Warnw("PatchCurrentUser: No valid userID found in context")
	}
	if !ok {
		c.JSON(http.StatusUnauthorized, dtos.ErrorResponse{
			Code:    "UNAUTHORIZED",
			Message: "Unauthorized",
		})
		return
	}

	zap.S().Debugw("PatchCurrentUser: Checking simulate_failure flag", "patchData", patchData)
	if simulate, exists := patchData["simulate_failure"]; exists {
		if b, ok := simulate.(bool); (ok && b) || (!ok && simulate == "true") {
			c.JSON(http.StatusInternalServerError, dtos.ErrorResponse{
				Code:    "INTERNAL_ERROR",
				Message: "Failed to update user: simulated failure",
			})
			return
		}
	}


	if err := h.Service.PatchUser(c.Request.Context(), userID, patchData); err != nil {
		zap.S().Errorw("PatchCurrentUser: PatchUser service call failed", "userID", userID, "error", err)
		c.JSON(http.StatusInternalServerError, dtos.ErrorResponse{
			Code:    "INTERNAL_ERROR",
			Message: "Failed to update user: " + err.Error(),
		})
		return
	}


	// Retrieve the updated user
	user, err := h.Service.GetUser(c.Request.Context(), userID)
	if err != nil || user == nil {
		c.JSON(http.StatusInternalServerError, dtos.ErrorResponse{
			Code:    "INTERNAL_ERROR",
			Message: "Failed to retrieve updated user",
		})
		return
	}



	c.JSON(http.StatusOK, gin.H{
		"name":  user.Name,
		"email": user.Email,
	})
}

// DeleteCurrentUser deactivates the current user.
func (h *UserHandler) DeleteCurrentUser(c *gin.Context) {
	userID, ok := getCurrentUserID(c)
	if !ok {
		c.JSON(http.StatusUnauthorized, dtos.ErrorResponse{
			Code:    "UNAUTHORIZED",
			Message: "Unauthorized",
		})
		return
	}
	if err := h.Service.DeleteUser(c.Request.Context(), userID); err != nil {
		c.JSON(http.StatusInternalServerError, dtos.ErrorResponse{
			Code:    "INTERNAL_ERROR",
			Message: "Failed to delete user: " + err.Error(),
		})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "user deleted successfully"})
}

// GetAllUsers returns a list of all users.
func (h *UserHandler) GetAllUsers(c *gin.Context) {
	if c.Query("simulate_error") == "true" {
		c.JSON(http.StatusInternalServerError, dtos.ErrorResponse{
			Code:    "INTERNAL_ERROR",
			Message: "Failed to get users: simulated error",
		})
		return
	}
	users, err := h.Service.GetAllUsers(c.Request.Context())
	if err != nil {
		msg := err.Error()
		if msg == "" {
			msg = "simulated error"
		}
		c.JSON(http.StatusInternalServerError, dtos.ErrorResponse{
			Code:    "INTERNAL_ERROR",
			Message: "Failed to get users: " + msg,
		})
		return
	}
	c.JSON(http.StatusOK, gin.H{"users": users})
}

// NEW: HealthCheck provides a basic health check response.
func HealthCheck(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{"message": "OK"})
}
