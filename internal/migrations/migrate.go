package migrations

import (
	"fmt"
	"os"
	"time"

	"go.uber.org/zap"

	"github.com/golang-migrate/migrate/v4"
	_ "github.com/golang-migrate/migrate/v4/database/postgres" // postgres driver
	_ "github.com/golang-migrate/migrate/v4/source/file"       // file source
)

// RunSQLMigrations builds the DSN from environment variables and applies up migrations.
func RunSQLMigrations() error {
	logger := zap.L()

	host := os.Getenv("POSTGRES_HOST")
	port := os.Getenv("POSTGRES_PORT")
	user := os.Getenv("POSTGRES_USER")
	password := os.Getenv("POSTGRES_PASSWORD")
	dbname := os.Getenv("POSTGRES_DB")

	// DSN format: "postgres://username:password@host:port/dbname?sslmode=disable"
	dsn := fmt.Sprintf("postgres://%s:%s@%s:%s/%s?sslmode=disable", user, password, host, port, dbname)

	var m *migrate.Migrate
	var err error

	maxAttempts := 10
	for i := 1; i <= maxAttempts; i++ {
		m, err = migrate.New("file:///app/internal/migrations", dsn)
		if err != nil {
			logger.Error("failed to create migrate instance",
				zap.Int("attempt", i),
				zap.Error(err))
			time.Sleep(5 * time.Second)
			continue
		}

		err = m.Up()
		if err == nil {
			logger.Info("database migrations ran successfully")
			return nil
		}

		if err == migrate.ErrNoChange {
			logger.Info("no new migrations to apply")
			return nil
		}

		logger.Error("failed to run migrations",
			zap.Int("attempt", i),
			zap.Error(err))
		time.Sleep(5 * time.Second)
	}

	// Just log the error and continue
	logger.Error("migrations failed but continuing anyway",
		zap.Error(err))
	return nil
}
