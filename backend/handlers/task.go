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

func GetAllTasks(c *fiber.Ctx) error {
	var tasks []models.Task
	var total int64

	// Get query parameters
	status := c.Query("status")
	dueDateStr := c.Query("due_date")
	page := c.QueryInt("page", 1)
	size := c.QueryInt("size", 10)
	search := c.Query("search")

	// Build base query
	query := database.DB.Model(&models.Task{})

	// Apply filters
	if status != "" {
		query = query.Where("status = ?", status)
	}

	if dueDateStr != "" {
		// Parse due_date string to time.Time
		dueDate, err := time.Parse("2006-01-02", dueDateStr)
		if err != nil {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Invalid due_date format. Use YYYY-MM-DD"})
		}
		// Filter tasks due on or before the specified date
		query = query.Where("due_date <= ?", dueDate)
	}

	// Apply search by title
	if search != "" {
		query = query.Where("title ILIKE ?", "%"+search+"%")
	}

	// Count total records
	query.Count(&total)

	// Apply pagination
	offset := (page - 1) * size
	query = query.Offset(offset).Limit(size)

	// Execute query
	if result := query.Find(&tasks); result.Error != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Could not retrieve tasks"})
	}

	return c.Status(fiber.StatusOK).JSON(models.TasksResponse{
		Tasks: tasks,
		Total: total,
		Page:  page,
		Size:  size,
	})
}

func UpdateTask(c *fiber.Ctx) error {
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

	taskTitle := c.Params("title")
	if taskTitle == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Task title cannot be empty"})
	}

	var existingTask models.Task
	if result := database.DB.Where("title = ?", taskTitle).First(&existingTask); result.Error != nil {
		if result.Error == gorm.ErrRecordNotFound {
			return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"error": "Task not found"})
		}
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Could not retrieve task"})
	}

	updateRequest := new(models.UpdateTaskRequest)
	if err := c.BodyParser(updateRequest); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Invalid body"})
	}

	// Validate the update request
	if err := validate.Struct(updateRequest); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": err.Error()})
	}

	// Apply updates
	if updateRequest.Title != nil {
		// Check for unique title if title is being updated
		if *updateRequest.Title != existingTask.Title {
			var count int64
			database.DB.Model(&models.Task{}).Where("title = ?", *updateRequest.Title).Count(&count)
			if count > 0 {
				return c.Status(fiber.StatusConflict).JSON(fiber.Map{"error": "Task with this new title already exists"})
			}
		}
		existingTask.Title = *updateRequest.Title
	}

	if updateRequest.Description != nil {
		existingTask.Description = *updateRequest.Description
	}

	if updateRequest.Status != nil {
		existingTask.Status = *updateRequest.Status
	}

	if updateRequest.DueDate != nil {
		existingTask.DueDate = updateRequest.DueDate
	}

	existingTask.UpdatedAt = time.Now()

	if result := database.DB.Save(&existingTask); result.Error != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Could not update task"})
	}

	return c.Status(fiber.StatusOK).JSON(existingTask)
}

func DeleteTask(c *fiber.Ctx) error {
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

	if result := database.DB.Delete(&task); result.Error != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Could not delete task"})
	}

	return c.Status(fiber.StatusOK).JSON(fiber.Map{"message": "Task deleted successfully"})
}
