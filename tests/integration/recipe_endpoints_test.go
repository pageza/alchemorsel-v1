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

func TestIntegrationListRecipes(t *testing.T) {
	router := routes.SetupRouter()
	req, _ := http.NewRequest("GET", "/v1/recipes", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	var resp map[string]string
	json.Unmarshal(w.Body.Bytes(), &resp)
	assert.Equal(t, "ListRecipes endpoint - TODO: implement logic", resp["message"])
}

func TestIntegrationGetRecipe(t *testing.T) {
	router := routes.SetupRouter()
	req, _ := http.NewRequest("GET", "/v1/recipes/1", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	var resp map[string]string
	json.Unmarshal(w.Body.Bytes(), &resp)
	assert.Equal(t, "GetRecipe endpoint - TODO: implement logic", resp["message"])
}

func TestIntegrationCreateRecipe(t *testing.T) {
	router := routes.SetupRouter()
	req, _ := http.NewRequest("POST", "/v1/recipes", strings.NewReader("{}"))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusCreated, w.Code)
	var resp map[string]string
	json.Unmarshal(w.Body.Bytes(), &resp)
	assert.Equal(t, "CreateRecipe endpoint - TODO: implement logic", resp["message"])
}

func TestIntegrationUpdateRecipe(t *testing.T) {
	router := routes.SetupRouter()
	req, _ := http.NewRequest("PUT", "/v1/recipes/1", strings.NewReader("{}"))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	var resp map[string]string
	json.Unmarshal(w.Body.Bytes(), &resp)
	assert.Equal(t, "UpdateRecipe endpoint - TODO: implement logic", resp["message"])
}

func TestIntegrationDeleteRecipe(t *testing.T) {
	router := routes.SetupRouter()
	req, _ := http.NewRequest("DELETE", "/v1/recipes/1", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	var resp map[string]string
	json.Unmarshal(w.Body.Bytes(), &resp)
	assert.Equal(t, "DeleteRecipe endpoint - TODO: implement logic", resp["message"])
}
