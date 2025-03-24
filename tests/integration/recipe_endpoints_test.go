package integration

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/pageza/alchemorsel-v1/internal/routes" // using the existing routes configuration
	"github.com/stretchr/testify/assert"
)

// These integration tests perform complete flows including creation.
// They now expect production-like responses (e.g. JSON objects with proper resource fields)
// which will fail until the endpoints are fully implemented.

// TestIntegrationListRecipes creates a recipe and then expects GET /v1/recipes to return a list
/*
func TestIntegrationListRecipes(t *testing.T) {
	router := routes.SetupRouter()

	// First, create a recipe
	postBody := `{"title": "Integration Test Recipe", "ingredients": ["ing1", "ing2"], "steps": ["s1", "s2"]}`
	createReq, _ := http.NewRequest("POST", "/v1/recipes", strings.NewReader(postBody))
	createReq.Header.Set("Content-Type", "application/json")
	createResp := httptest.NewRecorder()
	router.ServeHTTP(createResp, createReq)
	assert.Equal(t, http.StatusCreated, createResp.Code)

	// Now, list recipes
	req, _ := http.NewRequest("GET", "/v1/recipes", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)
	assert.Equal(t, http.StatusOK, w.Code)

	var resp struct {
		Data []map[string]interface{} `json:"data"`
	}
	err := json.Unmarshal(w.Body.Bytes(), &resp)
	assert.NoError(t, err)
	// We expect at least one recipe in the list
	assert.GreaterOrEqual(t, len(resp.Data), 1)
}
*/

// TestIntegrationGetRecipe creates a recipe then retrieves it by ID.
/*
func TestIntegrationGetRecipe(t *testing.T) {
	router := routes.SetupRouter()

	// Create a recipe
	postBody := `{"title": "Integration Test Recipe Get", "ingredients": ["ing"], "steps": ["s"]}`
	createReq, _ := http.NewRequest("POST", "/v1/recipes", strings.NewReader(postBody))
	createReq.Header.Set("Content-Type", "application/json")
	createResp := httptest.NewRecorder()
	router.ServeHTTP(createResp, createReq)
	assert.Equal(t, http.StatusCreated, createResp.Code)

	var created struct {
		ID    float64 `json:"id"`
		Title string  `json:"title"`
	}
	err := json.Unmarshal(createResp.Body.Bytes(), &created)
	assert.NoError(t, err)
	assert.NotZero(t, created.ID)

	// Retrieve the created recipe
	getURL := fmt.Sprintf("/v1/recipes/%v", created.ID)
	req, _ := http.NewRequest("GET", getURL, nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)
	assert.Equal(t, http.StatusOK, w.Code)

	var recipe struct {
		ID    float64 `json:"id"`
		Title string  `json:"title"`
	}
	err = json.Unmarshal(w.Body.Bytes(), &recipe)
	assert.NoError(t, err)
	assert.Equal(t, created.ID, recipe.ID)
	assert.Equal(t, created.Title, recipe.Title)
}
*/

// TestIntegrationSaveRecipe expects the POST endpoint to return a saved recipe with a valid ID.
func TestIntegrationSaveRecipe(t *testing.T) {
	router := routes.SetupRouter()

	reqBody := `{"title": "Integration Created Recipe", "ingredients": ["ing1"], "steps": ["step1"], "approved": true}`
	req, _ := http.NewRequest("POST", "/v1/recipes", strings.NewReader(reqBody))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusCreated, w.Code)
	var recipe struct {
		ID    float64 `json:"id"`
		Title string  `json:"title"`
	}
	err := json.Unmarshal(w.Body.Bytes(), &recipe)
	assert.NoError(t, err)
	assert.NotZero(t, recipe.ID)
	assert.Equal(t, "Integration Created Recipe", recipe.Title)
}

// TestIntegrationUpdateRecipe expects the PUT endpoint to update and return the recipe.
/*
func TestIntegrationUpdateRecipe(t *testing.T) {
	router := routes.SetupRouter()

	// First, create a recipe to update
	postBody := `{"title": "Recipe to Update", "ingredients": ["ing"], "steps": ["s"]}`
	createReq, _ := http.NewRequest("POST", "/v1/recipes", strings.NewReader(postBody))
	createReq.Header.Set("Content-Type", "application/json")
	createResp := httptest.NewRecorder()
	router.ServeHTTP(createResp, createReq)
	assert.Equal(t, http.StatusCreated, createResp.Code)

	var created struct {
		ID    float64 `json:"id"`
		Title string  `json:"title"`
	}
	err := json.Unmarshal(createResp.Body.Bytes(), &created)
	assert.NoError(t, err)

	// Now update the recipe
	updateBody := `{"title": "Updated Recipe Title", "ingredients": ["ing1", "ing2"], "steps": ["step1", "step2"]}`
	updateURL := fmt.Sprintf("/v1/recipes/%v", created.ID)
	updateReq, _ := http.NewRequest("PUT", updateURL, strings.NewReader(updateBody))
	updateReq.Header.Set("Content-Type", "application/json")
	updateResp := httptest.NewRecorder()
	router.ServeHTTP(updateResp, updateReq)
	assert.Equal(t, http.StatusOK, updateResp.Code)

	var updatedRecipe struct {
		ID    float64 `json:"id"`
		Title string  `json:"title"`
	}
	err = json.Unmarshal(updateResp.Body.Bytes(), &updatedRecipe)
	assert.NoError(t, err)
	assert.Equal(t, created.ID, updatedRecipe.ID)
	assert.Equal(t, "Updated Recipe Title", updatedRecipe.Title)
}
*/

// TestIntegrationDeleteRecipe expects the DELETE endpoint to return status 204 No Content.
// func TestIntegrationDeleteRecipe(t *testing.T) {
// 	router := routes.SetupRouter()

// 	// Create a recipe to delete
// 	postBody := `{"title": "Recipe to Delete", "ingredients": ["ing"], "steps": ["s"]}`
// 	createReq, _ := http.NewRequest("POST", "/v1/recipes", strings.NewReader(postBody))
// 	createReq.Header.Set("Content-Type", "application/json")
// 	createResp := httptest.NewRecorder()
// 	router.ServeHTTP(createResp, createReq)
// 	assert.Equal(t, http.StatusCreated, createResp.Code)

// 	var created struct {
// 		ID    float64 `json:"id"`
// 		Title string  `json:"title"`
// 	}
// 	err := json.Unmarshal(createResp.Body.Bytes(), &created)
// 	assert.NoError(t, err)

// 	// Delete the recipe
// 	deleteURL := fmt.Sprintf("/v1/recipes/%v", created.ID)
// 	req, _ := http.NewRequest("DELETE", deleteURL, nil)
// 	w := httptest.NewRecorder()
// 	router.ServeHTTP(w, req)
// 	assert.Equal(t, http.StatusNoContent, w.Code)
// 	// Optionally, check that the response body is empty
// 	assert.Empty(t, w.Body.Bytes())
// }
