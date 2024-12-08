package config

import (
	"fmt"
	"log"
	"os"

	"github.com/fnxr21/author-service/internal/model"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/schema"
)

var DB *gorm.DB

func ConnectPostgresGORM() {
	var err error
	host := os.Getenv("DB_HOST")
	port := os.Getenv("DB_PORT")
	user := os.Getenv("DB_USER")
	pass := os.Getenv("DB_PASSWORD")
	name := os.Getenv("DB_NAME")
	var gormConfig = &gorm.Config{
		NamingStrategy: schema.NamingStrategy{
			SingularTable: true,
		}}

	dsn := fmt.Sprintf(
		"host=%s port=%s user=%s password=%s dbname=%s sslmode=disable",
		host, port, user, pass, name,
	)
	DB, err = gorm.Open(postgres.Open(dsn), gormConfig)

	if err != nil {
		log.Fatalf("Error connecting to database: %v", err)
	}
	MigrationModel()
}

func MigrationModel() {
	var err error

	err = DB.AutoMigrate(model.Author{})
	if err != nil {
		fmt.Println(err)
		log.Fatalf("Error migrating database: %v", err)
	}
}
