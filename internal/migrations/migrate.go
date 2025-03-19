package migrations

import (
	"fmt"
	"os"

	"go.uber.org/zap"

	"github.com/golang-migrate/migrate/v4"
	_ "github.com/golang-migrate/migrate/v4/database/postgres" // postgres driver
	_ "github.com/golang-migrate/migrate/v4/source/file"       // file source
)

// RunMigrations builds the DSN from environment variables and applies up migrations.
func RunMigrations() error {
	host := os.Getenv("POSTGRES_HOST")
	port := os.Getenv("POSTGRES_PORT")
	user := os.Getenv("POSTGRES_USER")
	password := os.Getenv("POSTGRES_PASSWORD")
	dbname := os.Getenv("POSTGRES_DB")

	// DSN format: "postgres://username:password@host:port/dbname?sslmode=disable"
	dsn := fmt.Sprintf("postgres://%s:%s@%s:%s/%s?sslmode=disable", user, password, host, port, dbname)

	m, err := migrate.New("file://migrations", dsn)
	if err != nil {
		zap.L().Error("failed to create migrate instance", zap.Error(err))
		return err
	}

	// Run up migrations. ErrNoChange means there's nothing new to apply.
	err = m.Up()
	if err != nil && err != migrate.ErrNoChange {
		zap.L().Error("failed to run migrations", zap.Error(err))
		return err
	}

	zap.L().Info("database migrations ran successfully")
	return nil
}
