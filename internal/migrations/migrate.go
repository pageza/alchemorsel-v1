package migrations

import (
	"fmt"
	"os"
	"strings"
	"time"

	"go.uber.org/zap"

	"github.com/golang-migrate/migrate/v4"
	_ "github.com/golang-migrate/migrate/v4/database/postgres" // postgres driver
	_ "github.com/golang-migrate/migrate/v4/source/file"       // file source
)

// RunSQLiteMigrations builds the DSN from environment variables and applies up migrations.
func RunSQLiteMigrations() error {
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
		m, err = migrate.New("file://migrations", dsn)
		if err != nil {
			zap.L().Error("failed to create migrate instance", zap.Int("attempt", i), zap.Error(err))
		} else {
			err = m.Up()
			if err == nil || err == migrate.ErrNoChange {
				zap.L().Info("database migrations ran successfully")
				return nil
			}
			if strings.Contains(err.Error(), "uni_users_email") {
				zap.L().Warn("Ignoring legacy drop error during migrations", zap.Error(err))
				return nil
			}
			zap.L().Error("failed to run migrations", zap.Int("attempt", i), zap.Error(err))
		}
		time.Sleep(5 * time.Second)
	}
	return fmt.Errorf("failed to run migrations after %d attempts: %v", maxAttempts, err)
}
