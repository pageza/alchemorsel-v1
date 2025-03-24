package integration

import (
	"context"
	"fmt"
	"os"
	"testing"
	"time"

	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/wait"
)

func TestMain(m *testing.M) {
	ctx := context.Background()

	req := testcontainers.ContainerRequest{
		Image:        "postgres:13", // Use a stable Postgres version matching production
		ExposedPorts: []string{"5432/tcp"},
		Env: map[string]string{
			"POSTGRES_USER":     "postgres",
			"POSTGRES_PASSWORD": "testpass",
			"POSTGRES_DB":       "testdb",
		},
		WaitingFor: wait.ForListeningPort("5432/tcp").WithStartupTimeout(60 * time.Second),
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

	// Set environment variables so that the application uses the ephemeral Postgres instance.
	os.Setenv("DB_DRIVER", "postgres")
	dsn := fmt.Sprintf("host=%s port=%s user=testuser password=testpass dbname=testdb sslmode=disable", host, mappedPort.Port())
	os.Setenv("DB_SOURCE", dsn)

	// Also set the individual Postgres environment variables required by other parts of the code.
	os.Setenv("POSTGRES_HOST", host)
	os.Setenv("POSTGRES_PORT", mappedPort.Port())
	os.Setenv("POSTGRES_USER", "testuser")
	os.Setenv("POSTGRES_PASSWORD", "testpass")
	os.Setenv("POSTGRES_DB", "testdb")

	code := m.Run()
	os.Exit(code)
}
