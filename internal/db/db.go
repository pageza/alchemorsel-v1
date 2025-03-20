package db

import (
	"fmt"
	"os"

	"gorm.io/driver/postgres"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

var DB *gorm.DB

// Init initializes the database connection.
func Init() error {
	driver := os.Getenv("DB_DRIVER")
	if driver == "sqlite" {
		source := os.Getenv("DB_SOURCE")
		if source == "" {
			return fmt.Errorf("DB_SOURCE must be set when DB_DRIVER is sqlite")
		}
		var err error
		DB, err = gorm.Open(sqlite.Open(source), &gorm.Config{})
		return err
	}

	host := os.Getenv("POSTGRES_HOST")
	port := os.Getenv("POSTGRES_PORT")
	user := os.Getenv("POSTGRES_USER")
	password := os.Getenv("POSTGRES_PASSWORD")
	dbname := os.Getenv("POSTGRES_DB")

	dsn := fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=disable", host, port, user, password, dbname)
	var err error
	DB, err = gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		return err
	}
	return nil
}
