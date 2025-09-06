package routes

import (
	"github.com/gofiber/fiber/v2"
)

func Setup(app *fiber.App) {
	app.Get("/", func(c *fiber.Ctx) error {
		return c.SendString("Hey Champ, the Task API is now live!")
	})

	// Public wallet routes
	// app.Post("/create", handlers.CreateWallet)
}
