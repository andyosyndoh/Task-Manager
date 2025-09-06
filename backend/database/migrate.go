package database

import (
	"log"
	"task/backend/models"
)

func RunMigrations() {
	log.Println("Running database migrations...")

	err := DB.AutoMigrate(
		&models.Task{},
	)

	if err != nil {
		log.Fatalf("Failed to run migrations: %v", err)
	}

	log.Println("✅ Database migrations completed successfully")
}
