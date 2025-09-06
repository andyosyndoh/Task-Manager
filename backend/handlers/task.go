package handlers

import (
	"time"

	"task/backend/database"
	"task/backend/models"

	"github.com/go-playground/validator/v10"
	"github.com/gofiber/fiber/v2"
)

func CreateTask(c *fiber.Ctx) error {
	// Create a new validator instance
	validate := validator.New()

	taskRequest := new(models.CreateTaskRequest)
	if err := c.BodyParser(taskRequest); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Invalid body"})
	}

	// Validate the request body
	if err := validate.Struct(taskRequest); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": err.Error()})
	}

	// Create a new Task model instance
	task := models.Task{
		Title:       taskRequest.Title,
		Description: taskRequest.Description,
		Status:      models.TaskStatusPending, // Default status
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}

	// Set status from request if provided and valid
	if taskRequest.Status != "" {
		task.Status = taskRequest.Status
	}

	// Save the task to the database
	if result := database.DB.Create(&task); result.Error != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Could not create task"})
	}

	// Return the created task
	return c.Status(fiber.StatusCreated).JSON(task)
}
