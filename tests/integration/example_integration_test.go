package integration

import (
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	"github.com/pageza/alchemorsel-v1/internal/db"
	"github.com/pageza/alchemorsel-v1/internal/repositories"
	"github.com/pageza/alchemorsel-v1/internal/routes"
)

func TestHealthCheck(t *testing.T) {
	// Set environment variables for test
	os.Setenv("DB_DRIVER", "postgres")
	os.Setenv("TEST_MODE", "true")

	// Initialize the database
	config := db.NewConfig()
	database, err := db.InitDB(config)
	if err != nil {
		t.Fatalf("Failed to initialize DB: %v", err)
	}

	// Run migrations
	if err := repositories.RunMigrations(database); err != nil {
		t.Fatalf("Failed to run migrations: %v", err)
	}

	// Create test logger
	logger := createTestLogger()

	// Initialize the router with the database and logger
	router := routes.SetupRouter(database, logger)

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
	// Set environment variables for test
	os.Setenv("DB_DRIVER", "postgres")
	os.Setenv("TEST_MODE", "true")

	// Initialize the database
	config := db.NewConfig()
	database, err := db.InitDB(config)
	if err != nil {
		t.Fatalf("Failed to initialize DB: %v", err)
	}

	// Run migrations
	if err := repositories.RunMigrations(database); err != nil {
		t.Fatalf("Failed to run migrations: %v", err)
	}

	// Create test logger
	logger := createTestLogger()

	// Initialize the router with the database and logger
	_ = routes.SetupRouter(database, logger)

	// Use the router in your tests
	// ...
}
