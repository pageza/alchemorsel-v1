package integration

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/pageza/alchemorsel-v1/internal/db"
	"github.com/pageza/alchemorsel-v1/internal/routes"
)

func TestHealthCheck(t *testing.T) {
	// Initialize the database
	config := db.NewConfig()
	database, err := db.InitDB(config)
	if err != nil {
		t.Fatalf("Failed to initialize DB: %v", err)
	}

	// Initialize the router with the database
	router := routes.SetupRouter(database)

	// Call the health check endpoint
	req, _ := http.NewRequest("GET", "/v1/health", nil)
	resp := httptest.NewRecorder()
	router.ServeHTTP(resp, req)

	// Expect a 200 response
	if resp.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d", resp.Code)
	}
}

func TestExample(t *testing.T) {
	// Initialize the database
	config := db.NewConfig()
	database, err := db.InitDB(config)
	if err != nil {
		t.Fatalf("Failed to initialize DB: %v", err)
	}

	// Initialize the router with the database
	_ = routes.SetupRouter(database)

	// Use the router in your tests
	// ...
}
