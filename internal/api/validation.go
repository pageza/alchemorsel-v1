package api

import (
	"fmt"
	"net/http"
	"reflect"

	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
)

// ValidationError represents a validation error
type ValidationError struct {
	Field   string `json:"field"`
	Tag     string `json:"tag"`
	Value   string `json:"value"`
	Message string `json:"message"`
}

// ValidationErrors represents multiple validation errors
type ValidationErrors struct {
	Errors []ValidationError `json:"errors"`
}

// Error returns the error message
func (e ValidationErrors) Error() string {
	return fmt.Sprintf("validation failed: %d errors", len(e.Errors))
}

// Validator handles request/response validation
type Validator struct {
	validate *validator.Validate
}

// NewValidator creates a new validator
func NewValidator() *Validator {
	return &Validator{
		validate: validator.New(),
	}
}

// ValidateStruct validates a struct
func (v *Validator) ValidateStruct(obj interface{}) error {
	if err := v.validate.Struct(obj); err != nil {
		if validationErrors, ok := err.(validator.ValidationErrors); ok {
			var errors []ValidationError
			for _, e := range validationErrors {
				errors = append(errors, ValidationError{
					Field:   e.Field(),
					Tag:     e.Tag(),
					Value:   fmt.Sprintf("%v", e.Value()),
					Message: v.getErrorMessage(e),
				})
			}
			return ValidationErrors{Errors: errors}
		}
		return err
	}
	return nil
}

// ValidateRequest validates a request body
func (v *Validator) ValidateRequest(c *gin.Context, obj interface{}) error {
	if err := c.ShouldBindJSON(obj); err != nil {
		return err
	}
	return v.ValidateStruct(obj)
}

// ValidateQuery validates query parameters
func (v *Validator) ValidateQuery(c *gin.Context, obj interface{}) error {
	if err := c.ShouldBindQuery(obj); err != nil {
		return err
	}
	return v.ValidateStruct(obj)
}

// ValidatePath validates path parameters
func (v *Validator) ValidatePath(c *gin.Context, obj interface{}) error {
	if err := c.ShouldBindUri(obj); err != nil {
		return err
	}
	return v.ValidateStruct(obj)
}

// ValidateResponse validates a response body
func (v *Validator) ValidateResponse(obj interface{}) error {
	return v.ValidateStruct(obj)
}

// ValidationMiddleware creates a middleware for request validation
func (v *Validator) ValidationMiddleware(obj interface{}) gin.HandlerFunc {
	return func(c *gin.Context) {
		// Create a new instance of the validation object
		val := reflect.New(reflect.TypeOf(obj).Elem()).Interface()

		// Validate request
		if err := v.ValidateRequest(c, val); err != nil {
			if validationErrors, ok := err.(ValidationErrors); ok {
				c.JSON(http.StatusBadRequest, validationErrors)
				c.Abort()
				return
			}
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			c.Abort()
			return
		}

		// Store validated object in context
		c.Set("validated_request", val)
		c.Next()
	}
}

// getErrorMessage returns a user-friendly error message
func (v *Validator) getErrorMessage(err validator.FieldError) string {
	switch err.Tag() {
	case "required":
		return fmt.Sprintf("%s is required", err.Field())
	case "email":
		return fmt.Sprintf("%s must be a valid email address", err.Field())
	case "min":
		return fmt.Sprintf("%s must be at least %s", err.Field(), err.Param())
	case "max":
		return fmt.Sprintf("%s must be at most %s", err.Field(), err.Param())
	case "len":
		return fmt.Sprintf("%s must be exactly %s characters", err.Field(), err.Param())
	case "url":
		return fmt.Sprintf("%s must be a valid URL", err.Field())
	case "uuid":
		return fmt.Sprintf("%s must be a valid UUID", err.Field())
	case "datetime":
		return fmt.Sprintf("%s must be a valid datetime", err.Field())
	case "oneof":
		return fmt.Sprintf("%s must be one of: %s", err.Field(), err.Param())
	default:
		return fmt.Sprintf("%s is invalid", err.Field())
	}
}

// RegisterCustomValidators registers custom validation rules
func (v *Validator) RegisterCustomValidators() error {
	// Example custom validator for password strength
	return v.validate.RegisterValidation("password", func(fl validator.FieldLevel) bool {
		password := fl.Field().String()
		if len(password) < 8 {
			return false
		}
		// Add more password validation rules as needed
		return true
	})
}
