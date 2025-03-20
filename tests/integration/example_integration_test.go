package integration

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/pageza/alchemorsel-v1/internal/routes"
)

func TestHealthCheck(t *testing.T) {
	// Initialize the router first to set up DB and routes.
	router := routes.SetupRouter()
	// No need to call resetDB here since the health endpoint does not depend on user records.
	// Call the proper health check endpoint instead of /v1/users/1.
	req, _ := http.NewRequest("GET", "/v1/health", nil)
	resp := httptest.NewRecorder()
	router.ServeHTTP(resp, req)

	// Expect a 200 response.
	if resp.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d", resp.Code)
	}
}
