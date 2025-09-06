package database

import (
	"log"
	"task/backend/models"

	"gorm.io/gorm"
)

func RunMigrations() {
	log.Println("Running database migrations...")

	// Drop the tasks table if it exists
	err := DB.Migrator().DropTable(&models.Task{})
	if err != nil && err != gorm.ErrCantStartTransaction {
		log.Printf("Warning: Could not drop table 'tasks'. It might not exist or there's another issue: %v", err)
	}

	err = DB.AutoMigrate(
		&models.Task{},
	)

	if err != nil {
		log.Fatalf("Failed to run migrations: %v", err)
	}

	log.Println("✅ Database migrations completed successfully")
}
