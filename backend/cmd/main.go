package main

import (
	"fmt"
	"log"

	"github.com/gofiber/fiber/v2/middleware/cors"
	"github.com/gofiber/fiber/v2"
)

func main() {
	app := fiber.New()

	// Add CORS middleware
	app.Use(cors.New(cors.Config{
		AllowMethods:     "GET,POST,PUT,DELETE,OPTIONS",
		AllowHeaders:     "Origin,Content-Type,Accept,Authorization",
		// AllowCredentials: true,
	}))

	fmt.Println("🚀 Server starting on localhost:3000")
	log.Fatal(app.Listen("localhost:3000"))
}
