package unit

import (
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/pageza/alchemorsel-v1/internal/handlers"
	"github.com/pageza/alchemorsel-v1/internal/models"
	"github.com/stretchr/testify/assert"
)

// Begin: New mock implementation for RecipeServiceInterface

type MockRecipeService struct {
	GetRecipeFunc     func(id string) (*models.Recipe, error)
	SaveRecipeFunc    func(recipe *models.Recipe) error
	ResolveRecipeFunc func(query string, attributes map[string]interface{}) (*models.Recipe, []*models.Recipe, error)
}

func (m *MockRecipeService) GetRecipe(id string) (*models.Recipe, error) {
	if m.GetRecipeFunc != nil {
		return m.GetRecipeFunc(id)
	}
	return nil, errors.New("not implemented")
}

func (m *MockRecipeService) SaveRecipe(recipe *models.Recipe) error {
	if m.SaveRecipeFunc != nil {
		return m.SaveRecipeFunc(recipe)
	}
	return nil
}

func (m *MockRecipeService) ListRecipes() ([]*models.Recipe, error) {
	return nil, errors.New("not implemented")
}

func (m *MockRecipeService) UpdateRecipe(id string, recipe *models.Recipe) error {
	return errors.New("not implemented")
}

func (m *MockRecipeService) DeleteRecipe(id string) error {
	return errors.New("not implemented")
}

func (m *MockRecipeService) ResolveRecipe(query string, attributes map[string]interface{}) (*models.Recipe, []*models.Recipe, error) {
	if m.ResolveRecipeFunc != nil {
		return m.ResolveRecipeFunc(query, attributes)
	}
	return nil, nil, errors.New("not implemented")
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

	// Create a mock service with a GetRecipeFunc
	mockService := &MockRecipeService{
		GetRecipeFunc: func(id string) (*models.Recipe, error) {
			return &models.Recipe{
				ID:          id,
				Title:       "My Recipe",
				Ingredients: []byte("[]"),
				Steps:       []byte("[]"),
				CreatedAt:   time.Now(),
				UpdatedAt:   time.Now(),
			}, nil
		},
	}

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
	// Simulate a proper JSON payload for a new recipe.
	c.Request = httptest.NewRequest("POST", "/v1/recipes",
		strings.NewReader(`{"title": "New Recipe", "ingredients": ["ing1","ing2"], "steps": ["step1","step2"], "approved": true}`))
	c.Request.Header.Set("Content-Type", "application/json")

	// Create a mock service with a SaveRecipeFunc that sets the ID if not present
	mockService := &MockRecipeService{
		SaveRecipeFunc: func(recipe *models.Recipe) error {
			if recipe.ID == "" {
				recipe.ID = uuid.New().String()
			}
			return nil
		},
	}

	handler := handlers.NewRecipeHandler(mockService)
	handler.SaveRecipe(c) // Call the handler

	assert.Equal(t, http.StatusCreated, w.Code)

	var recipe struct {
		ID    string `json:"id"`
		Title string `json:"title"`
	}
	err := json.Unmarshal(w.Body.Bytes(), &recipe)
	assert.NoError(t, err)
	// Ensure the ID is a valid UUID
	assert.NotEmpty(t, recipe.ID)
	_, err = uuid.Parse(recipe.ID)
	assert.NoError(t, err)
	assert.Equal(t, "New Recipe", recipe.Title)
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

	// Create a mock service with a ResolveRecipeFunc
	mockService := &MockRecipeService{
		ResolveRecipeFunc: func(query string, attributes map[string]interface{}) (*models.Recipe, []*models.Recipe, error) {
			dummyRecipe := &models.Recipe{
				ID:          "dummy-id",
				Title:       "Dummy Recipe for " + query,
				Ingredients: []byte("[]"),
				Steps:       []byte("[]"),
			}
			similar := []*models.Recipe{
				{
					ID:          "similar-id",
					Title:       "Similar Recipe for " + query,
					Ingredients: []byte("[]"),
					Steps:       []byte("[]"),
				},
			}
			return dummyRecipe, similar, nil
		},
	}

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
