package routes

import (
	"task/backend/handlers"

	"github.com/gofiber/fiber/v2"
)

func Setup(app *fiber.App) {
	app.Get("/", func(c *fiber.Ctx) error {
		return c.SendString("Hey Champ, the Task API is now live!\nThanks to Andy\n")
	})

	// Public wallet routes
	app.Post("/tasks", handlers.CreateTask)
	app.Get("/tasks", handlers.GetAllTasks)
	app.Get("/tasks/:title", handlers.GetTask)
	app.Put("/tasks/:title", handlers.UpdateTask)
	app.Delete("/tasks/:title", handlers.DeleteTask)
}
