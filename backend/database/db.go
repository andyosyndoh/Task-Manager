// database/db.go
package database

import (
	"database/sql"
	"fmt"
	"log"
	"os"

	"github.com/joho/godotenv"
	"github.com/lib/pq" // PostgreSQL driver for checking specific errors
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

var DB *gorm.DB

func ConnectDB() {
	err := godotenv.Load()
	if err != nil {
		log.Fatal("Error loading .env file")
	}
	// --- Step 1: Connect to the default 'postgres' database to check if our DB exists ---
	const (
		host   = "localhost"
		port   = 5432
		dbname = "task_manager"
	)
	user := os.Getenv("DB_USER")
	password := os.Getenv("DB_PASSWORD")

	// DSN for the initial connection to the 'postgres' database
	initialDSN := fmt.Sprintf("host=%s port=%d user=%s password=%s dbname=postgres sslmode=disable", host, port, user, password)

	// Open a standard SQL connection
	sqlDB, err := sql.Open("postgres", initialDSN)
	if err != nil {
		log.Fatal("failed to connect to postgres database:", err)
	}
	defer sqlDB.Close()

	// Ping the database to ensure the connection is alive
	err = sqlDB.Ping()
	if err != nil {
		log.Fatal("failed to ping postgres database:", err)
	}

	// --- Step 2: Try to create the 'task_manager' database ---
	_, err = sqlDB.Exec(fmt.Sprintf("CREATE DATABASE %s", dbname))
	if err != nil {
		// If the error is that the database already exists, we can ignore it.
		// The pq library is needed to inspect the specific PostgreSQL error code.
		if pgErr, ok := err.(*pq.Error); ok && pgErr.Code == "42P04" { // 42P04 is for "duplicate_database"
			fmt.Printf("Database '%s' already exists. Continuing.\n", dbname)
		} else {
			// A different error occurred
			log.Fatal("failed to create database:", err)
		}
	} else {
		fmt.Printf("Database '%s' created successfully.\n", dbname)
	}

	// --- Step 3: Now, connect to the 'task_manager' database with GORM ---
	dsn := fmt.Sprintf("host=%s port=%d user=%s password=%s dbname=%s sslmode=disable", host, port, user, password, dbname)

	DB, err = gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		log.Fatal("failed to connect to task_manager database:", err)
	}

	fmt.Println("Connected to the database successfully.")
}