package middleware

import (
	"net/http"
	"runtime/debug"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/pageza/alchemorsel-v1/internal/monitoring"
	"github.com/sirupsen/logrus"
)

// ErrorResponse represents a standardized error response
type ErrorResponse struct {
	Error   string `json:"error"`
	Code    int    `json:"code"`
	Message string `json:"message"`
}

// ErrorHandler middleware handles errors globally
func ErrorHandler() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Next()

		// Check if there are any errors
		if len(c.Errors) > 0 {
			err := c.Errors.Last().Err
			logger := logrus.WithFields(logrus.Fields{
				"path":   c.Request.URL.Path,
				"method": c.Request.Method,
				"ip":     c.ClientIP(),
			})

			// Log the error with stack trace in development
			if gin.Mode() == gin.DebugMode {
				logger.WithField("stack", string(debug.Stack())).Error(err)
			} else {
				logger.Error(err)
			}

			// Record metrics
			monitoring.ObserveHTTPRequest(
				c.Request.Method,
				c.Request.URL.Path,
				c.Writer.Status(),
				c.Request.Context().Value("duration").(time.Duration),
			)

			// Determine the appropriate status code
			statusCode := http.StatusInternalServerError
			if c.Writer.Status() != 0 {
				statusCode = c.Writer.Status()
			}

			// Create error response
			response := ErrorResponse{
				Error:   err.Error(),
				Code:    statusCode,
				Message: "An error occurred processing your request",
			}

			// Handle specific error types
			switch e := err.(type) {
			case *RecipeError:
				response.Code = e.Code
				response.Message = e.Message
			case *ValidationError:
				response.Code = http.StatusBadRequest
				response.Message = "Validation failed"
			case *AuthenticationError:
				response.Code = http.StatusUnauthorized
				response.Message = "Authentication failed"
			case *AuthorizationError:
				response.Code = http.StatusForbidden
				response.Message = "Authorization failed"
			}

			c.JSON(statusCode, response)
		}
	}
}

// RecipeError represents a domain-specific error for recipe operations
type RecipeError struct {
	Code    int
	Message string
	Err     error
}

func (e *RecipeError) Error() string {
	if e.Err != nil {
		return e.Message + ": " + e.Err.Error()
	}
	return e.Message
}

// ValidationError represents a validation error
type ValidationError struct {
	Message string
	Fields  map[string]string
}

func (e *ValidationError) Error() string {
	return e.Message
}

// AuthenticationError represents an authentication error
type AuthenticationError struct {
	Message string
}

func (e *AuthenticationError) Error() string {
	return e.Message
}

// AuthorizationError represents an authorization error
type AuthorizationError struct {
	Message string
}

func (e *AuthorizationError) Error() string {
	return e.Message
}

// NewRecipeError creates a new RecipeError
func NewRecipeError(code int, message string, err error) *RecipeError {
	return &RecipeError{
		Code:    code,
		Message: message,
		Err:     err,
	}
}

// NewValidationError creates a new ValidationError
func NewValidationError(message string, fields map[string]string) *ValidationError {
	return &ValidationError{
		Message: message,
		Fields:  fields,
	}
}

// NewAuthenticationError creates a new AuthenticationError
func NewAuthenticationError(message string) *AuthenticationError {
	return &AuthenticationError{
		Message: message,
	}
}

// NewAuthorizationError creates a new AuthorizationError
func NewAuthorizationError(message string) *AuthorizationError {
	return &AuthorizationError{
		Message: message,
	}
}
