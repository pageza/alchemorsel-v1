package unit

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/pageza/alchemorsel-v1/internal/handlers" // using the existing handler file
	"github.com/stretchr/testify/assert"
)

// These unit tests are written with the intended production behavior in mind.
// For example, GetRecipe should return a JSON object with an "id" (number) and "title".
// Right now, the stubs will cause these tests to fail, which is what we want.

// TestListRecipesHandler expects the response to contain a "data" field with a slice of recipes.
func TestListRecipesHandler(t *testing.T) {
	gin.SetMode(gin.TestMode)
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)

	handlers.ListRecipes(c) // Intended behavior: returns { "data": [<recipe>, ...] } with status 200

	assert.Equal(t, http.StatusOK, w.Code)

	var resp struct {
		Data []interface{} `json:"data"`
	}
	err := json.Unmarshal(w.Body.Bytes(), &resp)
	assert.NoError(t, err)
	// Failing if the "data" key is missing or not an array.
}

// TestGetRecipeHandler expects a JSON object with "id" and "title" keys.
func TestGetRecipeHandler(t *testing.T) {
	gin.SetMode(gin.TestMode)
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Params = []gin.Param{{Key: "id", Value: "1"}}

	handlers.GetRecipe(c) // Intended: returns a recipe object (e.g., { "id": 1, "title": "My Recipe", ... })

	assert.Equal(t, http.StatusOK, w.Code)

	var recipe struct {
		ID    float64 `json:"id"`
		Title string  `json:"title"`
	}
	err := json.Unmarshal(w.Body.Bytes(), &recipe)
	assert.NoError(t, err)
	// Also ensure that the ID and Title are present.
	assert.NotZero(t, recipe.ID)
	assert.NotEmpty(t, recipe.Title)
}

// TestCreateRecipeHandler expects that posting valid JSON creates a new recipe
// and returns the created recipe with a new numeric ID.
func TestCreateRecipeHandler(t *testing.T) {
	gin.SetMode(gin.TestMode)
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	// Simulate a proper JSON payload for a new recipe.
	c.Request = httptest.NewRequest("POST", "/v1/recipes",
		strings.NewReader(`{"title": "New Recipe", "ingredients": ["ing1","ing2"], "steps": ["step1","step2"]}`))
	c.Request.Header.Set("Content-Type", "application/json")

	handlers.CreateRecipe(c) // Intended: returns { "id": <number>, "title": "New Recipe", ... } with status 201

	assert.Equal(t, http.StatusCreated, w.Code)

	var recipe struct {
		ID    float64 `json:"id"`
		Title string  `json:"title"`
	}
	err := json.Unmarshal(w.Body.Bytes(), &recipe)
	assert.NoError(t, err)
	assert.NotZero(t, recipe.ID)
	assert.Equal(t, "New Recipe", recipe.Title)
}

// TestUpdateRecipeHandler expects that updating a recipe returns the updated recipe.
func TestUpdateRecipeHandler(t *testing.T) {
	gin.SetMode(gin.TestMode)
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Params = []gin.Param{{Key: "id", Value: "1"}}
	c.Request = httptest.NewRequest("PUT", "/v1/recipes/1",
		strings.NewReader(`{"title": "Updated Recipe", "ingredients": ["new1"], "steps": ["newstep"]}`))
	c.Request.Header.Set("Content-Type", "application/json")

	handlers.UpdateRecipe(c) // Intended: returns updated recipe with status 200

	assert.Equal(t, http.StatusOK, w.Code)

	var recipe struct {
		ID    float64 `json:"id"`
		Title string  `json:"title"`
	}
	err := json.Unmarshal(w.Body.Bytes(), &recipe)
	assert.NoError(t, err)
	assert.Equal(t, "Updated Recipe", recipe.Title)
}

// TestDeleteRecipeHandler expects a 204 No Content response upon deletion.
func TestDeleteRecipeHandler(t *testing.T) {
	gin.SetMode(gin.TestMode)
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Params = []gin.Param{{Key: "id", Value: "1"}}

	handlers.DeleteRecipe(c) // Intended: returns status 204 and no content

	assert.Equal(t, http.StatusNoContent, w.Code)
	// We could also assert that the response body is empty.
	assert.Empty(t, w.Body.Bytes())
}
