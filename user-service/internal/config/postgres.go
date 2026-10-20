// Package config opens the database connection.
package config

import (
	"fmt"
	"log"
	"os"
	"time"

	"github.com/fnxr21/user-service/internal/model"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
	"gorm.io/gorm/schema"
)

// ConnectDatabase connects to PostgreSQL using the DB_* environment variables
// and creates or updates the tables. It stops the program if anything fails,
// because the service cannot work without its database.
func ConnectDatabase() *gorm.DB {
	dsn := fmt.Sprintf(
		"host=%s port=%s user=%s password=%s dbname=%s sslmode=disable",
		os.Getenv("DB_HOST"), os.Getenv("DB_PORT"), os.Getenv("DB_USER"),
		os.Getenv("DB_PASSWORD"), os.Getenv("DB_NAME"),
	)

	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{
		// Only log slow queries and real errors. "Record not found" is a normal
		// 404, and logging it would also print the query values.
		Logger: logger.New(log.Default(), logger.Config{
			SlowThreshold:             200 * time.Millisecond,
			LogLevel:                  logger.Warn,
			IgnoreRecordNotFoundError: true,
		}),
		// TranslateError turns database-specific errors into GORM errors, e.g.
		// a unique-index violation becomes gorm.ErrDuplicatedKey.
		TranslateError: true,
		// Table names are singular: model.User -> table "user".
		NamingStrategy: schema.NamingStrategy{SingularTable: true},
	})
	if err != nil {
		log.Fatalf("Error connecting to database: %v", err)
	}

	migrate(db)
	return db
}

// migrate creates missing tables and columns. AutoMigrate never drops
// columns, so it is safe to run on every start (fine for a lab; production
// systems usually use versioned migration files instead).
func migrate(db *gorm.DB) {
	if err := db.AutoMigrate(&model.User{}); err != nil {
		log.Fatalf("Error migrating database: %v", err)
	}
}
