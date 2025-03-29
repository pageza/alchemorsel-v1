package unit

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/pageza/alchemorsel-v1/internal/handlers"
	"github.com/pageza/alchemorsel-v1/internal/models"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"gorm.io/datatypes"
)

// Begin: New mock implementation for RecipeServiceInterface

type MockRecipeService struct {
	mock.Mock
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

func (m *MockRecipeService) ListRecipes(ctx context.Context) ([]*models.Recipe, error) {
	args := m.Called(ctx)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*models.Recipe), args.Error(1)
}

func (m *MockRecipeService) UpdateRecipe(ctx context.Context, recipe *models.Recipe) error {
	args := m.Called(ctx, recipe)
	return args.Error(0)
}

func (m *MockRecipeService) DeleteRecipe(ctx context.Context, id string) error {
	args := m.Called(ctx, id)
	return args.Error(0)
}

func (m *MockRecipeService) ResolveRecipe(query string, attributes map[string]interface{}) (*models.Recipe, []*models.Recipe, error) {
	args := m.Called(query, attributes)
	if args.Get(0) == nil {
		return nil, nil, args.Error(2)
	}
	return args.Get(0).(*models.Recipe), args.Get(1).([]*models.Recipe), args.Error(2)
}

// End: New mock implementation

// TestListRecipesHandler expects the response to contain a "data" field with a slice of recipes.
// func TestListRecipesHandler(t *testing.T) {
// 	gin.SetMode(gin.TestMode)
// 	w := httptest.NewRecorder()
// 	c, _ := gin.CreateTestContext(w)
//
// 	handlers.ListRecipes(c) // Intended behavior: returns { "data": [<recipe>, ...] } with status 200
//
// 	assert.Equal(t, http.StatusOK, w.Code)
//
// 	var resp struct {
// 		Data []interface{} `json:"data"`
// 	}
// 	err := json.Unmarshal(w.Body.Bytes(), &resp)
// 	assert.NoError(t, err)
// 	// Failing if the "data" key is missing or not an array.
// }

// TestGetRecipeHandler expects a JSON object with "id" and "title" keys, and verifies that id is a valid UUID.
func TestGetRecipeHandler(t *testing.T) {
	gin.SetMode(gin.TestMode)
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	// Using a known valid UUID for testing
	testID := "123e4567-e89b-12d3-a456-426614174000"
	c.Params = []gin.Param{{Key: "id", Value: testID}}
	// Added to ensure c.Request is not nil
	c.Request = httptest.NewRequest("GET", "/v1/recipes/"+testID, nil)

	// Create a mock service
	mockService := new(MockRecipeService)
	mockService.On("GetRecipe", c.Request.Context(), testID).Return(&models.Recipe{
		ID:          testID,
		Title:       "My Recipe",
		Ingredients: []byte("[]"),
		Steps:       []byte("[]"),
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}, nil)

	handler := handlers.NewRecipeHandler(mockService)
	handler.GetRecipe(c) // Call the handler

	assert.Equal(t, http.StatusOK, w.Code)

	var recipe struct {
		ID    string `json:"id"`
		Title string `json:"title"`
	}
	err := json.Unmarshal(w.Body.Bytes(), &recipe)
	assert.NoError(t, err)
	// Ensure that the ID is not empty and is a valid UUID
	assert.NotEmpty(t, recipe.ID)
	_, err = uuid.Parse(recipe.ID)
	assert.NoError(t, err)
	assert.Equal(t, "My Recipe", recipe.Title)
}

// TestSaveRecipeHandler expects that posting valid JSON creates a new recipe and returns it with a valid UUID.
func TestSaveRecipeHandler(t *testing.T) {
	gin.SetMode(gin.TestMode)
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)

	// Set TEST_MODE to true
	os.Setenv("TEST_MODE", "true")
	defer os.Unsetenv("TEST_MODE")

	// Create a valid JSON payload with all required fields
	payload := map[string]interface{}{
		"title": "New Recipe",
		"ingredients": []map[string]interface{}{
			{
				"name":   "ingredient1",
				"amount": "1",
				"unit":   "cup",
			},
		},
		"steps": []map[string]interface{}{
			{
				"order":       1,
				"description": "step1",
			},
		},
		"nutritional_info":   "",
		"allergy_disclaimer": "",
		"embedding":          []float64{},
		"approved":           true,
	}

	payloadBytes, _ := json.Marshal(payload)

	// Initialize the request with the JSON payload
	c.Request = httptest.NewRequest("POST", "/v1/recipes", strings.NewReader(string(payloadBytes)))
	c.Request.Header.Set("Content-Type", "application/json")

	// Create a mock service
	mockService := new(MockRecipeService)
	mockService.On("SaveRecipe", c.Request.Context(), mock.AnythingOfType("*models.Recipe")).Run(func(args mock.Arguments) {
		recipe := args.Get(1).(*models.Recipe)
		recipe.ID = uuid.New().String()
		recipe.Title = "New Recipe"
		recipe.CreatedAt = time.Now()
		recipe.UpdatedAt = time.Now()
		// Set ingredients and steps
		ingredientsJSON, _ := json.Marshal([]map[string]interface{}{
			{
				"name":   "ingredient1",
				"amount": "1",
				"unit":   "cup",
			},
		})
		stepsJSON, _ := json.Marshal([]map[string]interface{}{
			{
				"order":       1,
				"description": "step1",
			},
		})
		recipe.Ingredients = datatypes.JSON(ingredientsJSON)
		recipe.Steps = datatypes.JSON(stepsJSON)
	}).Return(nil)

	handler := handlers.NewRecipeHandler(mockService)
	handler.SaveRecipe(c) // Call the handler

	assert.Equal(t, http.StatusCreated, w.Code)

	var response struct {
		ID    string   `json:"id"`
		Title string   `json:"title"`
		Steps []string `json:"steps"`
	}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.NotEmpty(t, response.ID)
	_, err = uuid.Parse(response.ID)
	assert.NoError(t, err)
	assert.Equal(t, "New Recipe", response.Title)
}

// TestUpdateRecipeHandler expects that updating a recipe returns the updated recipe.
// func TestUpdateRecipeHandler(t *testing.T) {
// 	gin.SetMode(gin.TestMode)
// 	w := httptest.NewRecorder()
// 	c, _ := gin.CreateTestContext(w)
// 	c.Params = []gin.Param{{Key: "id", Value: "1"}}
// 	c.Request = httptest.NewRequest("PUT", "/v1/recipes/1",
// 		strings.NewReader(`{"title": "Updated Recipe", "ingredients": ["new1"], "steps": ["newstep"]}`))
// 	c.Request.Header.Set("Content-Type", "application/json")
//
// 	handlers.UpdateRecipe(c) // Intended: returns updated recipe with status 200
//
// 	assert.Equal(t, http.StatusOK, w.Code)
//
// 	var recipe struct {
// 		ID    float64 `json:"id"`
// 		Title string  `json:"title"`
// 	}
// 	err := json.Unmarshal(w.Body.Bytes(), &recipe)
// 	assert.NoError(t, err)
// 	assert.Equal(t, "Updated Recipe", recipe.Title)
// }

// TestDeleteRecipeHandler expects a 204 No Content response upon deletion.
// func TestDeleteRecipeHandler(t *testing.T) {
// 	gin.SetMode(gin.TestMode)
// 	w := httptest.NewRecorder()
// 	c, _ := gin.CreateTestContext(w)
// 	c.Params = []gin.Param{{Key: "id", Value: "1"}}
//
// 	handlers.DeleteRecipe(c) // Intended: returns status 204 and no content
//
// 	assert.Equal(t, http.StatusNoContent, w.Code)
// 	// We could also assert that the response body is empty.
// 	assert.Empty(t, w.Body.Bytes())
// }

// TestResolveRecipeHandler verifies that the ResolveRecipe endpoint returns a resolved recipe and similar recipes.
func TestResolveRecipeHandler(t *testing.T) {
	gin.SetMode(gin.TestMode)
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)

	// Prepare JSON payload for the request
	payload := `{"query": "chocolate cake", "attributes": {"approved": true}}`
	c.Request = httptest.NewRequest("POST", "/v1/recipes/resolve", strings.NewReader(payload))
	c.Request.Header.Set("Content-Type", "application/json")

	// Create a mock service
	mockService := new(MockRecipeService)
	dummyRecipe := &models.Recipe{
		ID:          "dummy-id",
		Title:       "Dummy Recipe for chocolate cake",
		Ingredients: []byte("[]"),
		Steps:       []byte("[]"),
	}
	similar := []*models.Recipe{
		{
			ID:          "similar-id",
			Title:       "Similar Recipe for chocolate cake",
			Ingredients: []byte("[]"),
			Steps:       []byte("[]"),
		},
	}
	mockService.On("ResolveRecipe", "chocolate cake", map[string]interface{}{"approved": true}).Return(dummyRecipe, similar, nil)

	// Initialize the handler with the mock service
	handler := handlers.NewRecipeHandler(mockService)

	// Call the ResolveRecipe handler
	handler.ResolveRecipe(c)

	// Verify the HTTP status code
	assert.Equal(t, http.StatusOK, w.Code)

	// Parse the response JSON
	var resp struct {
		Resolved *models.Recipe   `json:"resolved"`
		Similar  []*models.Recipe `json:"similar"`
	}
	err := json.Unmarshal(w.Body.Bytes(), &resp)
	assert.NoError(t, err)

	// Ensure the resolved recipe is not nil and has the expected title
	assert.NotNil(t, resp.Resolved)
	assert.Equal(t, "Dummy Recipe for chocolate cake", resp.Resolved.Title)

	// Check that exactly one similar recipe is returned with the expected title
	assert.Len(t, resp.Similar, 1)
	if len(resp.Similar) > 0 {
		assert.Equal(t, "Similar Recipe for chocolate cake", resp.Similar[0].Title)
	}
}
