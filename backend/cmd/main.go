package main

import (
	"fmt"
	"log"
	routes "task/backend/config"
	"task/backend/database"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/cors"
)

func main() {
	database.ConnectDB()
	database.RunMigrations()
	app := fiber.New()

	// Add CORS middleware
	app.Use(cors.New(cors.Config{
		AllowMethods: "GET,POST,PUT,DELETE,OPTIONS",
		AllowHeaders: "Origin,Content-Type,Accept,Authorization",
		// AllowCredentials: true,
	}))

	routes.Setup(app)

	fmt.Println("🚀 Server starting on localhost:3000")
	log.Fatal(app.Listen("localhost:3000"))
}
