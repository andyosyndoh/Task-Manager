package routes

import (
	"task/backend/handlers"

	"github.com/gofiber/fiber/v2"
)

func Setup(app *fiber.App) {
	app.Get("/", func(c *fiber.Ctx) error {
		return c.SendString("Hey Champ, the Task API is now live!")
	})

	app.Post("/create", handlers.CreateTask)
	// app.Get("/tasks", handlers.GetTasks)
	app.Get("/tasks/:title", handlers.GetTask)
	// app.Put("/tasks/:title", handlers.UpdateTask)
	// app.Delete("/tasks/:title", handlers.DeleteTask)
}
