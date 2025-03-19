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
	// Bind JSON from the request to the user struct.
	if err := c.ShouldBindJSON(&user); err != nil {
		zap.L().Error("failed to bind JSON", zap.Error(err))
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request payload"})
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
