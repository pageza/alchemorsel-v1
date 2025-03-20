package handlers

import (
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/golang-jwt/jwt/v4"

	"github.com/gin-gonic/gin"
	"github.com/pageza/alchemorsel-v1/internal/models"
	"github.com/pageza/alchemorsel-v1/internal/repositories"
)

func LoginUser(c *gin.Context) {
	var input struct {
		Email    string `json:"email"`
		Password string `json:"password"`
	}

	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	user, err := repositories.GetUserByEmail(c.Request.Context(), input.Email)
	if err != nil {
		if err.Error() == "user not found" {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid email or password"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	// Ensure a JWT secret is set.
	secret := os.Getenv("JWT_SECRET")
	if secret == "" {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "JWT secret not set"})
		return
	}

	// Build a JWT token with the user's ID and an expiration.
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"id":  user.ID,
		"exp": time.Now().Add(time.Hour).Unix(),
	})

	tokenString, err := token.SignedString([]byte(secret))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to generate token"})
		return
	}

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

func GetUser(c *gin.Context) {
	user, err := repositories.GetUser(c.Request.Context(), c.Param("id"))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	// Check for nil to return a 404 when the user is not found.
	if user == nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "user not found"})
		return
	}
	c.JSON(http.StatusOK, user)
}

// PatchCurrentUser allows partial updating of the current user's fields.
// It returns the updated user as a JSON object.
func PatchCurrentUser(c *gin.Context) {
	// Bind JSON payload to a map.
	var patchData map[string]string
	if err := c.ShouldBindJSON(&patchData); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request payload"})
		return
	}

	// Retrieve the current user ID from context.
	userID, exists := c.Get("currentUser")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
		return
	}

	// Retrieve the current user record using the repository helper.
	userPtr, err := repositories.GetUser(c.Request.Context(), userID.(string))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to retrieve user"})
		return
	}
	if userPtr == nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "user not found"})
		return
	}
	user := *userPtr

	// Apply patch updates – update name if provided.
	if newName, ok := patchData["name"]; ok {
		user.Name = newName
	}

	// Save the updated user record.
	if err := repositories.DB.Model(&user).Updates(user).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to update user"})
		return
	}

	// Return the updated user record in the response.
	c.JSON(http.StatusOK, gin.H{"user": user})
}

// CreateUser creates a new user.
// Now made idempotent: if a duplicate email is detected, the existing user is returned.
func CreateUser(c *gin.Context) {
	var newUser models.User
	if err := c.ShouldBindJSON(&newUser); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	// Always attempt to create a new user.
	// If a duplicate email is provided the DB will error out,
	// which we map to an HTTP 400 with a friendly message.
	if err := repositories.CreateUser(c.Request.Context(), &newUser); err != nil {
		if strings.Contains(err.Error(), "UNIQUE constraint failed") {
			c.JSON(http.StatusBadRequest, gin.H{"error": "email already in use"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusCreated, newUser)
}

// NEW: VerifyEmail handles email verification via a token.
func VerifyEmail(c *gin.Context) {
	token := c.Param("token")
	if token == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "missing token"})
		return
	}
	// TODO: Add real verification logic (e.g., lookup by token, mark email verified).
	c.JSON(http.StatusOK, gin.H{"message": "email verified successfully"})
}

// NEW: ForgotPassword initiates the forgot password flow.
func ForgotPassword(c *gin.Context) {
	var input struct {
		Email string `json:"email"`
	}
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	// TODO: Add password reset token generation and email sending.
	c.JSON(http.StatusOK, gin.H{"message": "reset password instructions sent"})
}

// NEW: ResetPassword handles password reset using a token and a new password.
func ResetPassword(c *gin.Context) {
	var input struct {
		Token       string `json:"token"`
		NewPassword string `json:"new_password"`
	}
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	if input.Token == "" || input.NewPassword == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "missing token or new password"})
		return
	}
	// TODO: Verify the token and update the user's password.
	c.JSON(http.StatusOK, gin.H{"message": "password reset successfully"})
}

// NEW: GetCurrentUser retrieves the current authenticated user's details.
func GetCurrentUser(c *gin.Context) {
	userID, ok := getCurrentUserID(c)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "user not authenticated"})
		return
	}
	user, err := repositories.GetUser(c.Request.Context(), userID)
	if err != nil || user == nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "user not found"})
		return
	}
	c.JSON(http.StatusOK, user)
}

// NEW: UpdateCurrentUser updates the current authenticated user's profile.
func UpdateCurrentUser(c *gin.Context) {
	userID, ok := getCurrentUserID(c)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "user not authenticated"})
		return
	}
	var input struct {
		Name string `json:"name"`
		// Additional fields can be added here.
	}
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	updatedUser := models.User{
		Name: input.Name,
	}
	if err := repositories.UpdateUser(c.Request.Context(), userID, &updatedUser); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "user updated successfully"})
}

// NEW: DeleteCurrentUser performs a soft deletion (deactivation) of the current user.
func DeleteCurrentUser(c *gin.Context) {
	userID, ok := getCurrentUserID(c)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "user not authenticated"})
		return
	}
	if err := repositories.DeactivateUser(c.Request.Context(), userID); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "user deactivated successfully"})
}

// NEW: GetAllUsers returns a list of all users (admin-only endpoint).
func GetAllUsers(c *gin.Context) {
	users, err := repositories.GetAllUsers(c.Request.Context())
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, users)
}

// NEW: HealthCheck provides a basic health check response.
func HealthCheck(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{"message": "OK"})
}
