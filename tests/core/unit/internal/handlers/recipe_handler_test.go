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

	"github.com/gin-gonic/gin"
	"github.com/pageza/alchemorsel-v1/internal/dtos"
	"github.com/pageza/alchemorsel-v1/internal/handlers"
	"github.com/pageza/alchemorsel-v1/internal/middleware"
	"github.com/pageza/alchemorsel-v1/internal/models"
	testhelpers "github.com/pageza/alchemorsel-v1/tests"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"gorm.io/gorm"
)

func init() {
	os.Setenv("JWT_SECRET", "test-secret")
}

// MockRecipeService is a mock implementation of the RecipeService interface
type MockRecipeService struct {
	mock.Mock
}

func (m *MockRecipeService) ListRecipes(ctx context.Context, page, limit int, sort, order string) ([]models.Recipe, error) {
	args := m.Called(ctx, page, limit, sort, order)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]models.Recipe), args.Error(1)
}

func (m *MockRecipeService) GetRecipe(ctx context.Context, id string) (*models.Recipe, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.Recipe), args.Error(1)
}

func (m *MockRecipeService) SaveRecipe(ctx context.Context, recipe *models.Recipe) error {
	args := m.Called(ctx, recipe)
	return args.Error(0)
}

func (m *MockRecipeService) UpdateRecipe(ctx context.Context, recipe *models.Recipe) error {
	args := m.Called(ctx, recipe)
	return args.Error(0)
}

func (m *MockRecipeService) DeleteRecipe(ctx context.Context, id string) error {
	args := m.Called(ctx, id)
	return args.Error(0)
}

func (m *MockRecipeService) SearchRecipes(ctx context.Context, query string, tags []string, difficulty string) ([]models.Recipe, error) {
	args := m.Called(ctx, query, tags, difficulty)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]models.Recipe), args.Error(1)
}

func (m *MockRecipeService) RateRecipe(ctx context.Context, recipeID string, rating float64) error {
	args := m.Called(ctx, recipeID, rating)
	return args.Error(0)
}

func (m *MockRecipeService) GetRecipeRatings(ctx context.Context, recipeID string) ([]float64, error) {
	args := m.Called(ctx, recipeID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]float64), args.Error(1)
}

func (m *MockRecipeService) ResolveRecipe(ctx context.Context, query string, attributes map[string]interface{}) (*models.Recipe, []*models.Recipe, error) {
	args := m.Called(ctx, query, attributes)
	return args.Get(0).(*models.Recipe), args.Get(1).([]*models.Recipe), args.Error(2)
}

func setupTest() (*handlers.RecipeHandler, *gin.Engine, *MockRecipeService) {
	gin.SetMode(gin.TestMode)
	mockService := new(MockRecipeService)
	handler := handlers.NewRecipeHandler(mockService)
	router := gin.New()

	// Add auth middleware
	router.Use(middleware.AuthMiddleware())

	return handler, router, mockService
}

func TestListRecipes(t *testing.T) {
	handler, router, mockService := setupTest()
	router.GET("/recipes", handler.ListRecipes)

	t.Run("error listing recipes", func(t *testing.T) {
		mockService.On("ListRecipes", mock.Anything, 1, 10, "created_at", "desc").
			Return(nil, errors.New("database error"))

		w := httptest.NewRecorder()
		req, _ := http.NewRequest("GET", "/recipes?page=1&limit=10", nil)
		req.Header.Set("Authorization", "Bearer "+testhelpers.GenerateTestToken(nil))
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusInternalServerError, w.Code)

		var response dtos.ErrorResponse
		err := json.Unmarshal(w.Body.Bytes(), &response)
		assert.NoError(t, err)
		assert.Equal(t, "database error", response.Message)
	})

	t.Run("unauthorized access", func(t *testing.T) {
		w := httptest.NewRecorder()
		req, _ := http.NewRequest("GET", "/recipes?page=1&limit=10", nil)
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusUnauthorized, w.Code)

		var response dtos.ErrorResponse
		err := json.Unmarshal(w.Body.Bytes(), &response)
		assert.NoError(t, err)
		assert.Equal(t, "UNAUTHORIZED", response.Code)
		assert.Equal(t, "Missing or invalid authorization token", response.Message)
	})
}

func TestGetRecipe(t *testing.T) {
	handler, router, mockService := setupTest()
	router.GET("/recipes/:id", handler.GetRecipe)

	t.Run("successful get recipe", func(t *testing.T) {
		mockRecipe := &models.Recipe{
			ID:    "1",
			Title: "Test Recipe",
		}

		mockService.On("GetRecipe", mock.Anything, "1").
			Return(mockRecipe, nil)

		w := httptest.NewRecorder()
		req, _ := http.NewRequest("GET", "/recipes/1", nil)
		req.Header.Set("Authorization", "Bearer "+testhelpers.GenerateTestToken(nil))
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)

		var response dtos.RecipeResponse
		err := json.Unmarshal(w.Body.Bytes(), &response)
		assert.NoError(t, err)
		assert.Equal(t, "Test Recipe", response.Title)
	})

	t.Run("recipe not found", func(t *testing.T) {
		mockService.On("GetRecipe", mock.Anything, "999").
			Return(nil, gorm.ErrRecordNotFound)

		w := httptest.NewRecorder()
		req, _ := http.NewRequest("GET", "/recipes/999", nil)
		req.Header.Set("Authorization", "Bearer "+testhelpers.GenerateTestToken(nil))
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusNotFound, w.Code)

		var response dtos.ErrorResponse
		err := json.Unmarshal(w.Body.Bytes(), &response)
		assert.NoError(t, err)
		assert.Equal(t, "NOT_FOUND", response.Code)
		assert.Equal(t, "Recipe not found", response.Message)
	})

	t.Run("unauthorized access", func(t *testing.T) {
		w := httptest.NewRecorder()
		req, _ := http.NewRequest("GET", "/recipes/1", nil)
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusUnauthorized, w.Code)

		var response dtos.ErrorResponse
		err := json.Unmarshal(w.Body.Bytes(), &response)
		assert.NoError(t, err)
		assert.Equal(t, "UNAUTHORIZED", response.Code)
		assert.Equal(t, "Missing or invalid authorization token", response.Message)
	})
}

func TestSaveRecipe(t *testing.T) {
	handler, router, mockService := setupTest()
	router.POST("/recipes", handler.SaveRecipe)

	t.Run("successful save recipe", func(t *testing.T) {
		recipeReq := dtos.RecipeRequest{
			Title:       "New Recipe",
			Ingredients: []dtos.Ingredient{{Name: "Ingredient 1", Amount: "1", Unit: "cup"}},
			Steps:       []dtos.Step{{Order: 1, Description: "Step 1"}},
		}

		mockService.On("SaveRecipe", mock.Anything, mock.AnythingOfType("*models.Recipe")).
			Return(nil)

		body, _ := json.Marshal(recipeReq)
		w := httptest.NewRecorder()
		req, _ := http.NewRequest("POST", "/recipes", bytes.NewBuffer(body))
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("Authorization", "Bearer "+testhelpers.GenerateTestToken(nil))
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusCreated, w.Code)

		var response dtos.RecipeResponse
		err := json.Unmarshal(w.Body.Bytes(), &response)
		assert.NoError(t, err)
		assert.Equal(t, "New Recipe", response.Title)
	})

	t.Run("invalid request body", func(t *testing.T) {
		w := httptest.NewRecorder()
		req, _ := http.NewRequest("POST", "/recipes", bytes.NewBufferString("invalid json"))
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("Authorization", "Bearer "+testhelpers.GenerateTestToken(nil))
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusBadRequest, w.Code)

		var response dtos.ErrorResponse
		err := json.Unmarshal(w.Body.Bytes(), &response)
		assert.NoError(t, err)
		assert.Equal(t, "BAD_REQUEST", response.Code)
		assert.Contains(t, response.Message, "Invalid request body")
	})

	t.Run("missing required fields", func(t *testing.T) {
		recipeReq := dtos.RecipeRequest{
			Title: "", // Missing required title
		}

		body, _ := json.Marshal(recipeReq)
		w := httptest.NewRecorder()
		req, _ := http.NewRequest("POST", "/recipes", bytes.NewBuffer(body))
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("Authorization", "Bearer "+testhelpers.GenerateTestToken(nil))
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusBadRequest, w.Code)

		var response dtos.ErrorResponse
		err := json.Unmarshal(w.Body.Bytes(), &response)
		assert.NoError(t, err)
		assert.Equal(t, "BAD_REQUEST", response.Code)
		assert.Contains(t, response.Message, "Title")
		assert.Contains(t, response.Message, "Ingredients")
		assert.Contains(t, response.Message, "Steps")
	})

	t.Run("unauthorized access", func(t *testing.T) {
		recipeReq := dtos.RecipeRequest{
			Title: "New Recipe",
		}
		body, _ := json.Marshal(recipeReq)
		w := httptest.NewRecorder()
		req, _ := http.NewRequest("POST", "/recipes", bytes.NewBuffer(body))
		req.Header.Set("Content-Type", "application/json")
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusUnauthorized, w.Code)

		var response dtos.ErrorResponse
		err := json.Unmarshal(w.Body.Bytes(), &response)
		assert.NoError(t, err)
		assert.Equal(t, "UNAUTHORIZED", response.Code)
		assert.Equal(t, "Missing or invalid authorization token", response.Message)
	})
}

func TestSaveRecipe_missing_required_fields(t *testing.T) {
	handler, router, _ := setupTest()
	router.POST("/recipes", handler.SaveRecipe)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("POST", "/recipes", strings.NewReader("{}"))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+testhelpers.GenerateTestToken(nil))
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)

	var response dtos.ErrorResponse
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.Equal(t, "BAD_REQUEST", response.Code)
	assert.Contains(t, response.Message, "Title' Error:Field validation for 'Title' failed on the 'required' tag")
	assert.Contains(t, response.Message, "Ingredients' Error:Field validation for 'Ingredients' failed on the 'required' tag")
	assert.Contains(t, response.Message, "Steps' Error:Field validation for 'Steps' failed on the 'required' tag")
}

func TestDeleteRecipe(t *testing.T) {
	handler, router, mockService := setupTest()
	router.DELETE("/recipes/:id", handler.DeleteRecipe)

	t.Run("successful delete recipe", func(t *testing.T) {
		mockService.On("DeleteRecipe", mock.Anything, "1").
			Return(nil)

		w := httptest.NewRecorder()
		req, _ := http.NewRequest("DELETE", "/recipes/1", nil)
		req.Header.Set("Authorization", "Bearer "+testhelpers.GenerateTestToken(nil))
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusNoContent, w.Code)
	})

	t.Run("recipe not found", func(t *testing.T) {
		mockService.On("DeleteRecipe", mock.Anything, "999").
			Return(gorm.ErrRecordNotFound)

		w := httptest.NewRecorder()
		req, _ := http.NewRequest("DELETE", "/recipes/999", nil)
		req.Header.Set("Authorization", "Bearer "+testhelpers.GenerateTestToken(nil))
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusNotFound, w.Code)

		var response dtos.ErrorResponse
		err := json.Unmarshal(w.Body.Bytes(), &response)
		assert.NoError(t, err)
		assert.Equal(t, "NOT_FOUND", response.Code)
		assert.Equal(t, "Recipe not found", response.Message)
	})

	t.Run("unauthorized access", func(t *testing.T) {
		w := httptest.NewRecorder()
		req, _ := http.NewRequest("DELETE", "/recipes/1", nil)
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusUnauthorized, w.Code)

		var response dtos.ErrorResponse
		err := json.Unmarshal(w.Body.Bytes(), &response)
		assert.NoError(t, err)
		assert.Equal(t, "UNAUTHORIZED", response.Code)
		assert.Equal(t, "Missing or invalid authorization token", response.Message)
	})
}

func TestRateRecipe(t *testing.T) {
	t.Skip("Temporarily disabled - rating functionality not implemented yet")
	handler, router, mockService := setupTest()
	router.POST("/recipes/:id/rate", handler.RateRecipe)

	t.Run("successful rate recipe", func(t *testing.T) {
		mockService.On("RateRecipe", mock.Anything, "1", 5.0).
			Return(nil)
		mockService.On("GetRecipe", mock.Anything, "1").
			Return(&models.Recipe{ID: "1"}, nil)

		w := httptest.NewRecorder()
		body := `5.0`
		req, _ := http.NewRequest("POST", "/recipes/1/rate", strings.NewReader(body))
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("Authorization", "Bearer "+testhelpers.GenerateTestToken(nil))
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
	})

	t.Run("invalid rating value - too high", func(t *testing.T) {
		w := httptest.NewRecorder()
		body := `{"rating": 6.0}`
		req, _ := http.NewRequest("POST", "/recipes/1/rate", strings.NewReader(body))
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("Authorization", "Bearer "+testhelpers.GenerateTestToken(nil))
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusBadRequest, w.Code)

		var response dtos.ErrorResponse
		err := json.Unmarshal(w.Body.Bytes(), &response)
		assert.NoError(t, err)
		assert.Equal(t, "BAD_REQUEST", response.Code)
		assert.Contains(t, response.Message, "Rating must be between 0 and 5")
	})

	t.Run("invalid rating value - too low", func(t *testing.T) {
		w := httptest.NewRecorder()
		body := `{"rating": -1.0}`
		req, _ := http.NewRequest("POST", "/recipes/1/rate", strings.NewReader(body))
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("Authorization", "Bearer "+testhelpers.GenerateTestToken(nil))
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusBadRequest, w.Code)

		var response dtos.ErrorResponse
		err := json.Unmarshal(w.Body.Bytes(), &response)
		assert.NoError(t, err)
		assert.Equal(t, "BAD_REQUEST", response.Code)
		assert.Contains(t, response.Message, "Rating must be between 0 and 5")
	})

	t.Run("invalid rating value - not a number", func(t *testing.T) {
		w := httptest.NewRecorder()
		body := `{"rating": "invalid"}`
		req, _ := http.NewRequest("POST", "/recipes/1/rate", strings.NewReader(body))
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("Authorization", "Bearer "+testhelpers.GenerateTestToken(nil))
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusBadRequest, w.Code)

		var response dtos.ErrorResponse
		err := json.Unmarshal(w.Body.Bytes(), &response)
		assert.NoError(t, err)
		assert.Equal(t, "BAD_REQUEST", response.Code)
		assert.Contains(t, response.Message, "Invalid request body")
	})

	t.Run("unauthorized access", func(t *testing.T) {
		w := httptest.NewRecorder()
		body := `{"rating": 5.0}`
		req, _ := http.NewRequest("POST", "/recipes/1/rate", strings.NewReader(body))
		req.Header.Set("Content-Type", "application/json")
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusUnauthorized, w.Code)

		var response dtos.ErrorResponse
		err := json.Unmarshal(w.Body.Bytes(), &response)
		assert.NoError(t, err)
		assert.Equal(t, "UNAUTHORIZED", response.Code)
		assert.Equal(t, "Missing or invalid authorization token", response.Message)
	})
}
