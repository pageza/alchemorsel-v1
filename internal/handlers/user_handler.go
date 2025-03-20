package handlers

import (
	"net/http"

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

	// Call the service layer to create the user.
	if err := services.CreateUser(&user); err != nil {
		zap.L().Error("failed to create user", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to create user"})
		return
	}

	// For security, omit the password from the response.
	user.Password = ""

	c.JSON(http.StatusCreated, gin.H{"user": user})
}

// GetUser handles GET /v1/users/:id
func GetUser(c *gin.Context) {
	// TODO: Retrieve user by ID from the service layer.
	c.JSON(http.StatusOK, gin.H{"message": "GetUser endpoint - TODO: implement logic"})
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
	// Added explicit error checking for JSON binding. This ensures missing fields produce a 400.
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
		c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid credentials"})
		return
	}

	// Return the generated token in the response.
	c.JSON(http.StatusOK, gin.H{"token": token})
}
