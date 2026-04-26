package database

import (
	"fmt"
	"log"
	"os"

	"cert-system/server/internal/models"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

var DB *gorm.DB

// Initialize connects to PostgreSQL
func Initialize() {
	// Read configuration
	dbHost := os.Getenv("DB_HOST")
	dbUser := os.Getenv("DB_USER")
	dbPassword := os.Getenv("DB_PASSWORD")
	dbName := os.Getenv("DB_NAME")
	dbPort := os.Getenv("DB_PORT")

	// Build connection string
	dsn := fmt.Sprintf(
		"host=%s user=%s password=%s dbname=%s port=%s sslmode=disable",
		dbHost, dbUser, dbPassword, dbName, dbPort,
	)

	// Connect
	var err error
	DB, err = gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		log.Fatal("❌ Failed to connect to database:", err)
	}

	log.Println("✅ Database connected")

	// Auto-create tables
	err = DB.AutoMigrate(
		&models.University{},
		&models.Certificate{},
		&models.User{},
		&models.BatchAnchor{},
	)
	if err != nil {
		log.Fatal("❌ Failed to create tables:", err)
	}

	log.Println("✅ Database tables created")
}
