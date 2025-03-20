package handlers

import (
	"net/http"
	"strings"

	stdErrors "errors"

	"github.com/pageza/alchemorsel-v1/internal/dtos"
	almerrors "github.com/pageza/alchemorsel-v1/internal/errors"
	"github.com/pageza/alchemorsel-v1/internal/models"
	"github.com/pageza/alchemorsel-v1/internal/response"
	"github.com/pageza/alchemorsel-v1/internal/services"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

// CreateUser handles POST /v1/users
func CreateUser(c *gin.Context) {
	var user models.User
	if err := c.ShouldBindJSON(&user); err != nil {
		response.RespondError(c, http.StatusBadRequest, "invalid request payload")
		return
	}
	if err := services.CreateUser(c.Request.Context(), &user); err != nil {
		response.RespondError(c, http.StatusInternalServerError, err.Error())
		return
	}
	user.Password = ""
	response.RespondSuccess(c, http.StatusCreated, gin.H{"user": dtos.NewUserResponse(&user)})
}

// GetUser handles GET /v1/users/:id
func GetUser(c *gin.Context) {
	id := c.Param("id")
	user, err := services.GetUser(c.Request.Context(), id)
	if err != nil {
		if stdErrors.Is(err, almerrors.ErrUserNotFound) {
			c.JSON(http.StatusNotFound, gin.H{"error": "user not found"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to get user"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"user": dtos.NewUserResponse(user)})
}

// UpdateUser handles PUT /v1/users/:id
func UpdateUser(c *gin.Context) {
	// TODO: Parse request and update user details via service layer.
	c.JSON(http.StatusOK, gin.H{"message": "UpdateUser endpoint - TODO: implement logic"})
}

// DeleteUser handles DELETE /v1/users/:id
func DeleteUser(c *gin.Context) {
	id := c.Param("id")
	if err := services.DeleteUser(c.Request.Context(), id); err != nil {
		response.RespondError(c, http.StatusInternalServerError, err.Error())
		return
	}
	response.RespondSuccess(c, http.StatusOK, gin.H{"message": "user deleted"})
}

// LoginUser handles POST /v1/users/login
func LoginUser(c *gin.Context) {
	var loginReq models.LoginRequest
	if err := c.ShouldBindJSON(&loginReq); err != nil {
		response.RespondError(c, http.StatusBadRequest, "invalid login payload")
		return
	}
	token, err := services.LoginUser(c.Request.Context(), &loginReq)
	if err != nil {
		response.RespondError(c, http.StatusUnauthorized, "invalid credentials")
		return
	}
	response.RespondSuccess(c, http.StatusOK, gin.H{"token": token})
}

// GetCurrentUser handles GET /v1/users/me by extracting the user ID from the authentication token.
func GetCurrentUser(c *gin.Context) {
	userID, exists := c.Get("currentUser")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
		return
	}

	user, err := services.GetUser(c.Request.Context(), userID.(string))
	if err != nil {
		response.RespondError(c, http.StatusNotFound, "User not found")
		return
	}
	c.JSON(http.StatusOK, gin.H{"user": dtos.NewUserResponse(user)})
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

	if err := services.UpdateUser(c.Request.Context(), currentUser.(string), updatedUser); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to update user"})
		return
	}

	// Retrieve and return the updated user.
	user, err := services.GetUser(c.Request.Context(), currentUser.(string))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to retrieve updated user"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"user": dtos.NewUserResponse(user)})
}

// DeleteCurrentUser handles DELETE /v1/users/me to deactivate the current user's account.
func DeleteCurrentUser(c *gin.Context) {
	currentUser, exists := c.Get("currentUser")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
		return
	}

	if err := services.DeactivateUser(c.Request.Context(), currentUser.(string)); err != nil {
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
	user, err := services.GetUser(c.Request.Context(), currentUser.(string))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to retrieve current user"})
		return
	}
	if !user.IsAdmin {
		c.JSON(http.StatusForbidden, gin.H{"error": "forbidden"})
		return
	}

	users, err := services.GetAllUsers(c.Request.Context())
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

	if err := services.PatchUser(c.Request.Context(), currentUser.(string), patchPayload); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to update user"})
		return
	}

	user, err := services.GetUser(c.Request.Context(), currentUser.(string))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to retrieve updated user"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"user": dtos.NewUserResponse(user)})
}

// VerifyEmail handles GET /v1/users/verify-email/:token
func VerifyEmail(c *gin.Context) {
	token := c.Param("token")
	user, err := services.GetUserByEmailVerificationToken(c.Request.Context(), token)
	if err != nil || user == nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid or expired token"})
		return
	}

	user.EmailVerified = true
	user.EmailVerificationToken = "" // Clear the token after verification
	if err := services.UpdateUser(c.Request.Context(), user.ID, user); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to verify email"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "email verified successfully"})
}

// ForgotPassword handles POST /v1/users/forgot-password.
func ForgotPassword(c *gin.Context) {
	var req struct {
		Email string `json:"email" binding:"required,email"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		response.RespondError(c, http.StatusBadRequest, "invalid payload")
		return
	}
	if err := services.ForgotPassword(c.Request.Context(), req.Email); err != nil {
		zap.L().Error("failed to process forgot password", zap.Error(err))
		response.RespondError(c, http.StatusInternalServerError, "failed to process forgot password")
		return
	}
	// In production, an email would be sent with the reset link.
	response.RespondSuccess(c, http.StatusOK, gin.H{"message": "password reset link sent"})
}

// ResetPassword handles POST /v1/users/reset-password.
func ResetPassword(c *gin.Context) {
	var req struct {
		Token       string `json:"token" binding:"required"`
		NewPassword string `json:"new_password" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		response.RespondError(c, http.StatusBadRequest, "invalid payload")
		return
	}
	if err := services.ResetPassword(c.Request.Context(), req.Token, req.NewPassword); err != nil {
		zap.L().Error("failed to reset password", zap.Error(err))
		response.RespondError(c, http.StatusBadRequest, err.Error())
		return
	}
	response.RespondSuccess(c, http.StatusOK, gin.H{"message": "password has been reset successfully"})
}
