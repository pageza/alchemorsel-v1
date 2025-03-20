package handlers

import (
	"net/http"
	"strings"

	stdErrors "errors"

	almerrors "github.com/pageza/alchemorsel-v1/internal/errors"
	"github.com/pageza/alchemorsel-v1/internal/models"
	"github.com/pageza/alchemorsel-v1/internal/services"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

// CreateUser handles POST /v1/users
func CreateUser(c *gin.Context) {
	var user models.User
	// Added explicit error checking for JSON binding
	if err := c.ShouldBindJSON(&user); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid user payload"})
		return
	}

	// Attempt to create the user using the service layer.
	if err := services.CreateUser(&user); err != nil {
		// Check for duplicate entry error based on the error message.
		if strings.Contains(err.Error(), "duplicate key") || strings.Contains(err.Error(), "unique constraint") {
			c.JSON(http.StatusConflict, gin.H{"error": "user already exists"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to create user"})
		return
	}

	// For security, omit the password from the response.
	user.Password = ""

	c.JSON(http.StatusCreated, gin.H{"user": user})
}

// GetUser handles GET /v1/users/:id
func GetUser(c *gin.Context) {
	id := c.Param("id")
	user, err := services.GetUser(id)
	if err != nil {
		if stdErrors.Is(err, almerrors.ErrUserNotFound) {
			c.JSON(http.StatusNotFound, gin.H{"error": "user not found"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to get user"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"user": user})
}

// UpdateUser handles PUT /v1/users/:id
func UpdateUser(c *gin.Context) {
	// TODO: Parse request and update user details via service layer.
	c.JSON(http.StatusOK, gin.H{"message": "UpdateUser endpoint - TODO: implement logic"})
}

// DeleteUser handles DELETE /v1/users/:id
func DeleteUser(c *gin.Context) {
	// TODO: Delete user by ID via service layer.
	c.JSON(http.StatusOK, gin.H{"message": "DeleteUser endpoint - TODO: implement logic"})
}

// LoginUser handles POST /v1/users/login
func LoginUser(c *gin.Context) {
	// Using an inline struct with binding validations for login payload.
	var loginReq struct {
		Email    string `json:"email" binding:"required,email"`
		Password string `json:"password" binding:"required"`
	}
	// Added explicit error checking for JSON binding.
	if err := c.ShouldBindJSON(&loginReq); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid login payload"})
		return
	}

	// Call the service layer to authenticate the user and generate a token.
	token, err := services.LoginUser(&models.LoginRequest{
		Email:    loginReq.Email,
		Password: loginReq.Password,
	})
	if err != nil {
		zap.L().Error("failed to login user", zap.Error(err))
		if stdErrors.Is(err, almerrors.ErrAccountDeactivated) {
			c.JSON(http.StatusForbidden, gin.H{"error": "account is deactivated"})
		} else {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid credentials"})
		}
		return
	}

	// Return the generated token in the response.
	c.JSON(http.StatusOK, gin.H{"token": token})
}

// GetCurrentUser handles GET /v1/users/me by extracting the user ID from the authentication token.
func GetCurrentUser(c *gin.Context) {
	userID, exists := c.Get("currentUser")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
		return
	}

	user, err := services.GetUser(userID.(string))
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "User not found"})
		return
	}

	// Omit the password for security.
	user.Password = ""
	c.JSON(http.StatusOK, gin.H{"user": user})
}

// UpdateCurrentUser handles PUT /v1/users/me to update the current user's profile.
func UpdateCurrentUser(c *gin.Context) {
	currentUser, exists := c.Get("currentUser")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
		return
	}

	// Define a request struct for update payload.
	var updatePayload struct {
		Name  string `json:"name" binding:"omitempty"`
		Email string `json:"email" binding:"omitempty,email"`
	}

	if err := c.ShouldBindJSON(&updatePayload); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid update payload"})
		return
	}

	// NEW: Trim whitespace from incoming fields.
	updatedUser := &models.User{
		Name:  strings.TrimSpace(updatePayload.Name),
		Email: strings.TrimSpace(updatePayload.Email),
	}

	if err := services.UpdateUser(currentUser.(string), updatedUser); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to update user"})
		return
	}

	// Retrieve and return the updated user.
	user, err := services.GetUser(currentUser.(string))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to retrieve updated user"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"user": user})
}

// DeleteCurrentUser handles DELETE /v1/users/me to deactivate the current user's account.
func DeleteCurrentUser(c *gin.Context) {
	currentUser, exists := c.Get("currentUser")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
		return
	}

	if err := services.DeactivateUser(currentUser.(string)); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to deactivate user"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "user deactivated successfully"})
}

// GetAllUsers handles GET /admin/users for admin functionality.
func GetAllUsers(c *gin.Context) {
	currentUser, exists := c.Get("currentUser")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
		return
	}
	user, err := services.GetUser(currentUser.(string))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to retrieve current user"})
		return
	}
	if !user.IsAdmin {
		c.JSON(http.StatusForbidden, gin.H{"error": "forbidden"})
		return
	}

	users, err := services.GetAllUsers()
	if err != nil {
		zap.L().Error("failed to retrieve users", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to retrieve users"})
		return
	}
	// Omit the password for security.
	for i := range users {
		users[i].Password = ""
	}
	c.JSON(http.StatusOK, gin.H{"users": users})
}

// PatchCurrentUser handles PATCH /v1/users/me to update the current user's profile.
func PatchCurrentUser(c *gin.Context) {
	currentUser, exists := c.Get("currentUser")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
		return
	}

	var patchPayload map[string]interface{}
	if err := c.ShouldBindJSON(&patchPayload); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid update payload"})
		return
	}

	// NEW: Filter patchPayload to only allow "name" and "email".
	allowedFields := map[string]bool{
		"name":  true,
		"email": true,
	}
	for key, value := range patchPayload {
		if !allowedFields[key] {
			delete(patchPayload, key)
		} else if str, ok := value.(string); ok {
			patchPayload[key] = strings.TrimSpace(str)
		}
	}

	if err := services.PatchUser(currentUser.(string), patchPayload); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to update user"})
		return
	}

	user, err := services.GetUser(currentUser.(string))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to retrieve updated user"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"user": user})
}

// VerifyEmail handles GET /v1/users/verify-email/:token
func VerifyEmail(c *gin.Context) {
	token := c.Param("token")
	user, err := services.GetUserByEmailVerificationToken(token)
	if err != nil || user == nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid or expired token"})
		return
	}

	user.EmailVerified = true
	user.EmailVerificationToken = "" // Clear the token after verification
	if err := services.UpdateUser(user.ID, user); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to verify email"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "email verified successfully"})
}
