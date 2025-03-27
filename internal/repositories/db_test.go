package repositories

import (
	"context"
	"database/sql"
	"testing"
	"time"

	_ "github.com/lib/pq" // Postgres driver
	"github.com/pageza/alchemorsel-v1/internal/config"
	"github.com/pageza/alchemorsel-v1/internal/errors"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/wait"
)

func setupTestDB(t *testing.T) (*sql.DB, func()) {
	ctx := context.Background()

	// Create PostgreSQL container
	req := testcontainers.ContainerRequest{
		Image:        "postgres:15-alpine",
		ExposedPorts: []string{"5432/tcp"},
		Env: map[string]string{
			"POSTGRES_USER":     "test",
			"POSTGRES_PASSWORD": "test",
			"POSTGRES_DB":       "testdb",
		},
		WaitingFor: wait.ForLog("database system is ready to accept connections"),
	}

	container, err := testcontainers.GenericContainer(ctx, testcontainers.GenericContainerRequest{
		ContainerRequest: req,
		Started:          true,
	})
	require.NoError(t, err)

	// Get container host and port
	host, err := container.Host(ctx)
	require.NoError(t, err)
	port, err := container.MappedPort(ctx, "5432")
	require.NoError(t, err)

	// Create database connection
	connStr := "host=" + host + " port=" + port.Port() + " user=test password=test dbname=testdb sslmode=disable"
	db, err := sql.Open("postgres", connStr)
	require.NoError(t, err)

	// Test connection
	err = db.Ping()
	require.NoError(t, err)

	// Return cleanup function
	cleanup := func() {
		db.Close()
		container.Terminate(ctx)
	}

	return db, cleanup
}

func TestInitDB_EdgeCases(t *testing.T) {
	// Setup test database
	db, cleanup := setupTestDB(t)
	defer cleanup()

	tests := []struct {
		name        string
		config      *config.Config
		expectError bool
		validate    func(*testing.T, *sql.DB)
	}{
		{
			name: "Valid connection",
			config: &config.Config{
				Database: config.DatabaseConfig{
					Driver:   "postgres",
					Host:     "localhost",
					Port:     5432,
					User:     "test",
					Password: "test",
					DBName:   "testdb",
					SSLMode:  "disable",
				},
			},
			expectError: false,
			validate: func(t *testing.T, db *sql.DB) {
				// Verify we can query the database
				var result int
				err := db.QueryRow("SELECT 1").Scan(&result)
				require.NoError(t, err)
				assert.Equal(t, 1, result)
			},
		},
		{
			name: "Invalid connection string",
			config: &config.Config{
				Database: config.DatabaseConfig{
					Host:     "invalid-host",
					Port:     5432,
					User:     "test",
					Password: "test",
					DBName:   "testdb",
				},
			},
			expectError: true,
		},
		{
			name: "Invalid driver",
			config: &config.Config{
				Database: config.DatabaseConfig{
					Driver: "invalid-driver",
				},
			},
			expectError: true,
		},
		{
			name: "Invalid SSL mode",
			config: &config.Config{
				Database: config.DatabaseConfig{
					SSLMode: "invalid-ssl-mode",
				},
			},
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := InitDB(tt.config)
			if tt.expectError {
				assert.Error(t, err)
				return
			}
			assert.NoError(t, err)

			if tt.validate != nil {
				tt.validate(t, db)
			}
		})
	}
}

func TestCircuitBreaker_EdgeCases(t *testing.T) {
	tests := []struct {
		name       string
		failures   int
		expectOpen bool
	}{
		{
			name:       "No failures",
			failures:   0,
			expectOpen: false,
		},
		{
			name:       "Just below threshold",
			failures:   4,
			expectOpen: false,
		},
		{
			name:       "At threshold",
			failures:   5,
			expectOpen: true,
		},
		{
			name:       "Above threshold",
			failures:   10,
			expectOpen: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cb := NewCircuitBreaker(5, 10*time.Second)

			for i := 0; i < tt.failures; i++ {
				cb.RecordFailure()
			}

			assert.Equal(t, tt.expectOpen, cb.IsOpen())
		})
	}
}

func TestRetryWithBackoff_EdgeCases(t *testing.T) {
	tests := []struct {
		name        string
		maxRetries  int
		operation   func() error
		expectError bool
	}{
		{
			name:       "Successful operation",
			maxRetries: 3,
			operation: func() error {
				return nil
			},
			expectError: false,
		},
		{
			name:       "Always failing operation",
			maxRetries: 3,
			operation: func() error {
				return errors.New("OPERATION_FAILED", "operation failed")
			},
			expectError: true,
		},
		{
			name:       "Zero max retries",
			maxRetries: 0,
			operation: func() error {
				return errors.New("OPERATION_FAILED", "operation failed")
			},
			expectError: true,
		},
		{
			name:       "Negative max retries",
			maxRetries: -1,
			operation: func() error {
				return errors.New("OPERATION_FAILED", "operation failed")
			},
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			config := &RetryConfig{
				MaxRetries:      tt.maxRetries,
				InitialInterval: time.Second,
				MaxInterval:     time.Second * 5,
				Multiplier:      2.0,
				MaxElapsedTime:  time.Second * 10,
			}
			err := RetryWithBackoff(tt.operation, config)
			if tt.expectError {
				assert.Error(t, err)
				return
			}
			assert.NoError(t, err)
		})
	}
}
