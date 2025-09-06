package main

import (
	"fmt"
	"log"
	"task/backend/database"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/cors"
)

func main() {
	database.ConnectDB()
	app := fiber.New()

	// Add CORS middleware
	app.Use(cors.New(cors.Config{
		AllowMethods: "GET,POST,PUT,DELETE,OPTIONS",
		AllowHeaders: "Origin,Content-Type,Accept,Authorization",
		// AllowCredentials: true,
	}))

	fmt.Println("🚀 Server starting on localhost:3000")
	log.Fatal(app.Listen("localhost:3000"))
}
