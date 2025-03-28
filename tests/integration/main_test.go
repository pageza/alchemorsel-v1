package integration

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/docker/go-connections/nat"
	_ "github.com/lib/pq" // cursor-- Added to register the Postgres SQL driver for wait.ForSQL
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/wait"
)

func TestMain(m *testing.M) {
	ctx := context.Background()

	// Create a temporary directory for test files
	tmpDir, err := os.MkdirTemp("", "alchemorsel-integration-test-*")
	if err != nil {
		fmt.Printf("Failed to create temp directory: %v\n", err)
		os.Exit(1)
	}
	defer os.RemoveAll(tmpDir)

	req := testcontainers.ContainerRequest{
		Image:        "postgres:13", // Use a stable Postgres version matching production
		ExposedPorts: []string{"5432/tcp"},
		Env: map[string]string{
			"POSTGRES_USER":     "postgres",
			"POSTGRES_PASSWORD": "testpass",
			"POSTGRES_DB":       "testdb",
		},
		WaitingFor: wait.ForSQL("5432/tcp", "postgres", func(host string, port nat.Port) string {
			return fmt.Sprintf("host=%s port=%s user=postgres password=testpass dbname=testdb sslmode=disable", host, port.Port())
		}).WithStartupTimeout(60 * time.Second),
	}

	postgresC, err := testcontainers.GenericContainer(ctx, testcontainers.GenericContainerRequest{
		ContainerRequest: req,
		Started:          true,
	})
	if err != nil {
		fmt.Printf("Failed to start PostgreSQL container: %v\n", err)
		os.Exit(1)
	}
	// Ensure the container is terminated after all tests complete
	defer func() {
		_ = postgresC.Terminate(ctx)
	}()

	host, err := postgresC.Host(ctx)
	if err != nil {
		fmt.Printf("Failed to get container host: %v\n", err)
		os.Exit(1)
	}
	mappedPort, err := postgresC.MappedPort(ctx, "5432")
	if err != nil {
		fmt.Printf("Failed to get mapped port: %v\n", err)
		os.Exit(1)
	}

	// Create a temporary file for the DSN
	dsnFile := filepath.Join(tmpDir, "testdb.dsn")
	dsn := fmt.Sprintf("host=%s port=%s user=postgres password=testpass dbname=testdb sslmode=disable options='-c search_path=public,pg_catalog'", host, mappedPort.Port())
	if err := os.WriteFile(dsnFile, []byte(dsn), 0644); err != nil {
		fmt.Printf("Failed to write DSN file: %v\n", err)
		os.Exit(1)
	}

	// Set environment variables so that the application uses the ephemeral Postgres instance.
	os.Setenv("DB_DRIVER", "postgres")
	os.Setenv("DB_SOURCE", dsn)

	// Also set the individual Postgres environment variables required by other parts of the code.
	os.Setenv("POSTGRES_HOST", host)
	os.Setenv("POSTGRES_PORT", mappedPort.Port())
	os.Setenv("POSTGRES_USER", "postgres")
	os.Setenv("POSTGRES_PASSWORD", "testpass")
	os.Setenv("POSTGRES_DB", "testdb")
	// Set the JWT secret so that login endpoints can sign tokens successfully.
	os.Setenv("JWT_SECRET", "testsecret")

	code := m.Run()
	os.Exit(code)
}
