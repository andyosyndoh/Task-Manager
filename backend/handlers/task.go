package handlers

import (
	"strings"
	"time"

	"task/backend/database"
	"task/backend/models"

	"github.com/go-playground/validator/v10"
	"github.com/gofiber/fiber/v2"
	"gorm.io/gorm"
)

func CreateTask(c *fiber.Ctx) error {
	validate := validator.New()

	// Custom validation for future date
	validate.RegisterValidation("future", func(fl validator.FieldLevel) bool {
		if date, ok := fl.Field().Interface().(time.Time); ok {
			return date.After(time.Now())
		}
		return false
	})

	// Custom validation for no spaces
	validate.RegisterValidation("nospaces", func(fl validator.FieldLevel) bool {
		return !strings.Contains(fl.Field().String(), " ")
	})

	taskRequest := new(models.CreateTaskRequest)
	if err := c.BodyParser(taskRequest); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Invalid body"})
	}

	if err := validate.Struct(taskRequest); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": err.Error()})
	}

	task := models.Task{
		Title:       taskRequest.Title,
		Description: taskRequest.Description,
		Status:      models.TaskStatusPending,
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}

	if taskRequest.Status != "" {
		task.Status = taskRequest.Status
	}

	if taskRequest.DueDate != nil {
		task.DueDate = taskRequest.DueDate
	}

	if result := database.DB.Create(&task); result.Error != nil {
		if strings.Contains(result.Error.Error(), "duplicate key value violates unique constraint \"tasks_title_key\"") {
			return c.Status(fiber.StatusConflict).JSON(fiber.Map{"error": "Task with this title already exists"})
		}
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Could not create task"})
	}

	return c.Status(fiber.StatusCreated).JSON(task)
}

func GetTask(c *fiber.Ctx) error {
	taskTitle := c.Params("title")
	if taskTitle == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Task title cannot be empty"})
	}

	var task models.Task
	if result := database.DB.Where("title = ?", taskTitle).First(&task); result.Error != nil {
		if result.Error == gorm.ErrRecordNotFound {
			return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"error": "Task not found"})
		}
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Could not retrieve task"})
	}

	return c.Status(fiber.StatusOK).JSON(task)
}
