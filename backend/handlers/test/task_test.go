package handlers

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"
	"time"

	"task/backend/database"
	"task/backend/handlers"
	"task/backend/models"

	"github.com/gofiber/fiber/v2"
	"github.com/joho/godotenv"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

type HandlerTestSuite struct {
	suite.Suite
	app        *fiber.App
	db         *gorm.DB
	originalDB *gorm.DB
}

func (suite *HandlerTestSuite) SetupSuite() {
	// Load .env file
	err := godotenv.Load("../../../.env")
	suite.Require().NoError(err, "Error loading .env file")

	// Setup PostgreSQL database for testing
	dsn := fmt.Sprintf("host=%s port=%d user=%s password=%s dbname=%s sslmode=disable",
		"localhost",
		5432,
		os.Getenv("DB_USER"),
		os.Getenv("DB_PASSWORD"),
		"task_manager_test",
	)

	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Silent),
	})

	if err != nil {
		// Fallback to main database
		dsn = fmt.Sprintf("host=%s port=%d user=%s password=%s dbname=%s sslmode=disable",
			"localhost", 5432, os.Getenv("DB_USER"), os.Getenv("DB_PASSWORD"), "task_manager",
		)
		db, err = gorm.Open(postgres.Open(dsn), &gorm.Config{
			Logger: logger.Default.LogMode(logger.Silent),
		})
		suite.Require().NoError(err, "Failed to connect to PostgreSQL database")
	}

	// Store the original connection
	suite.originalDB = db
	database.DB = db

	// Migrate the schema
	err = db.AutoMigrate(&models.Task{})
	suite.Require().NoError(err, "Failed to migrate database schema")

	// Setup Fiber app
	suite.app = fiber.New()
	suite.setupRoutes()
}

func (suite *HandlerTestSuite) SetupTest() {
	// Clean the database before each test
	suite.originalDB.Exec("DELETE FROM tasks")

	// Start a new transaction for each test
	suite.db = suite.originalDB.Begin()
	suite.Require().NoError(suite.db.Error)
	database.DB = suite.db
}

func (suite *HandlerTestSuite) TearDownTest() {
	// Rollback transaction after each test
	if suite.db != nil {
		result := suite.db.Rollback()
		// Check for rollback errors but don't fail the test
		if result.Error != nil && !errors.Is(result.Error, gorm.ErrInvalidTransaction) {
			suite.T().Logf("Rollback error: %v", result.Error)
		}
	}
	// Reset to original connection
	database.DB = suite.originalDB

	// Clean up any remaining data
	suite.originalDB.Exec("DELETE FROM tasks")
}

func (suite *HandlerTestSuite) TearDownSuite() {
	// Final cleanup
	if suite.originalDB != nil {
		suite.originalDB.Exec("DELETE FROM tasks")
	}
}

func (suite *HandlerTestSuite) setupRoutes() {
	suite.app.Post("/tasks", handlers.CreateTask)
	suite.app.Get("/tasks", handlers.GetAllTasks)
	suite.app.Get("/tasks/:title", handlers.GetTask)
	suite.app.Put("/tasks/:title", handlers.UpdateTask)
	suite.app.Delete("/tasks/:title", handlers.DeleteTask)
}

func (suite *HandlerTestSuite) createTestTask(title, description string, status models.TaskStatus, dueDate *time.Time) models.Task {
	task := models.Task{
		Title:       title,
		Description: description,
		Status:      status,
		DueDate:     dueDate,
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}
	result := suite.db.Create(&task)
	suite.Require().NoError(result.Error)
	return task
}

func (suite *HandlerTestSuite) makeRequest(method, url string, body interface{}) (*http.Response, []byte) {
	var reqBody io.Reader
	if body != nil {
		jsonBody, err := json.Marshal(body)
		suite.Require().NoError(err)
		reqBody = bytes.NewBuffer(jsonBody)
	}

	req := httptest.NewRequest(method, url, reqBody)
	req.Header.Set("Content-Type", "application/json")

	resp, err := suite.app.Test(req, -1)
	suite.Require().NoError(err)

	respBody, err := io.ReadAll(resp.Body)
	suite.Require().NoError(err)
	resp.Body.Close()

	return resp, respBody
}

// ============================================================================
// CREATE TASK TESTS
// ============================================================================

func (suite *HandlerTestSuite) TestCreateTask_Success() {
	futureDate := time.Now().Add(24 * time.Hour)
	taskReq := models.CreateTaskRequest{
		Title:       "test-task",
		Description: "Test description",
		Status:      models.TaskStatusPending,
		DueDate:     &futureDate,
	}

	resp, body := suite.makeRequest("POST", "/tasks", taskReq)

	assert.Equal(suite.T(), http.StatusCreated, resp.StatusCode)

	var task models.Task
	err := json.Unmarshal(body, &task)
	suite.Require().NoError(err)

	assert.Equal(suite.T(), "test-task", task.Title)
	assert.Equal(suite.T(), "Test description", task.Description)
	assert.Equal(suite.T(), models.TaskStatusPending, task.Status)
	assert.NotNil(suite.T(), task.DueDate)
	assert.True(suite.T(), task.DueDate.After(time.Now()))
	assert.NotZero(suite.T(), task.ID)
	assert.NotZero(suite.T(), task.CreatedAt)
	assert.NotZero(suite.T(), task.UpdatedAt)
}

func (suite *HandlerTestSuite) TestCreateTask_WithCustomStatus() {
	futureDate := time.Now().Add(24 * time.Hour)
	taskReq := models.CreateTaskRequest{
		Title:       "test-task-custom",
		Description: "Test description",
		Status:      models.TaskStatusInProgress,
		DueDate:     &futureDate,
	}

	resp, body := suite.makeRequest("POST", "/tasks", taskReq)

	assert.Equal(suite.T(), http.StatusCreated, resp.StatusCode)

	var task models.Task
	err := json.Unmarshal(body, &task)
	suite.Require().NoError(err)
	assert.Equal(suite.T(), models.TaskStatusInProgress, task.Status)
}

func (suite *HandlerTestSuite) TestCreateTask_WithoutOptionalFields() {
	futureDate := time.Now().Add(24 * time.Hour)
	taskReq := models.CreateTaskRequest{
		Title:   "minimal-task",
		DueDate: &futureDate,
	}

	resp, body := suite.makeRequest("POST", "/tasks", taskReq)

	assert.Equal(suite.T(), http.StatusCreated, resp.StatusCode)

	var task models.Task
	err := json.Unmarshal(body, &task)
	suite.Require().NoError(err)

	assert.Equal(suite.T(), "minimal-task", task.Title)
	assert.Empty(suite.T(), task.Description)
	assert.Equal(suite.T(), models.TaskStatusPending, task.Status) // Default status
}

func (suite *HandlerTestSuite) TestCreateTask_DuplicateTitle() {
	// Create first task
	futureDate := time.Now().Add(24 * time.Hour)
	suite.createTestTask("duplicate-title", "First task", models.TaskStatusPending, &futureDate)

	// Try to create second task with same title
	taskReq := models.CreateTaskRequest{
		Title:   "duplicate-title",
		DueDate: &futureDate,
	}

	resp, body := suite.makeRequest("POST", "/tasks", taskReq)

	assert.Equal(suite.T(), http.StatusConflict, resp.StatusCode)

	var errorResp map[string]interface{}
	err := json.Unmarshal(body, &errorResp)
	suite.Require().NoError(err)
	assert.Contains(suite.T(), errorResp["error"], "Task with this title already exists")
}

func (suite *HandlerTestSuite) TestCreateTask_ValidationErrors() {
	testCases := []struct {
		name        string
		request     models.CreateTaskRequest
		expectedErr string
	}{
		{
			name: "Empty title",
			request: models.CreateTaskRequest{
				Title:   "",
				DueDate: func() *time.Time { t := time.Now().Add(24 * time.Hour); return &t }(),
			},
			expectedErr: "required",
		},
		{
			name: "Title with spaces",
			request: models.CreateTaskRequest{
				Title:   "title with spaces",
				DueDate: func() *time.Time { t := time.Now().Add(24 * time.Hour); return &t }(),
			},
			expectedErr: "nospaces",
		},
		{
			name: "Title too long",
			request: models.CreateTaskRequest{
				Title:   strings.Repeat("a", 201),
				DueDate: func() *time.Time { t := time.Now().Add(24 * time.Hour); return &t }(),
			},
			expectedErr: "max",
		},
		{
			name: "Missing due date",
			request: models.CreateTaskRequest{
				Title: "valid-title",
			},
			expectedErr: "required",
		},
		{
			name: "Past due date",
			request: models.CreateTaskRequest{
				Title:   "past-date-task",
				DueDate: func() *time.Time { t := time.Now().Add(-24 * time.Hour); return &t }(),
			},
			expectedErr: "future",
		},
		{
			name: "Current time (not future)",
			request: models.CreateTaskRequest{
				Title:   "current-time-task",
				DueDate: func() *time.Time { t := time.Now(); return &t }(),
			},
			expectedErr: "future",
		},
	}

	for _, tc := range testCases {
		suite.Run(tc.name, func() {
			resp, body := suite.makeRequest("POST", "/tasks", tc.request)

			assert.Equal(suite.T(), http.StatusBadRequest, resp.StatusCode)

			var errorResp map[string]interface{}
			err := json.Unmarshal(body, &errorResp)
			suite.Require().NoError(err)
			assert.Contains(suite.T(), errorResp["error"], tc.expectedErr)
		})
	}
}

func (suite *HandlerTestSuite) TestCreateTask_InvalidJSON() {
	req := httptest.NewRequest("POST", "/tasks", strings.NewReader("invalid json"))
	req.Header.Set("Content-Type", "application/json")

	resp, err := suite.app.Test(req)
	suite.Require().NoError(err)

	assert.Equal(suite.T(), http.StatusBadRequest, resp.StatusCode)

	body, err := io.ReadAll(resp.Body)
	suite.Require().NoError(err)

	var errorResp map[string]interface{}
	err = json.Unmarshal(body, &errorResp)
	suite.Require().NoError(err)
	assert.Equal(suite.T(), "Invalid body", errorResp["error"])
}

// ============================================================================
// GET TASK TESTS
// ============================================================================

func (suite *HandlerTestSuite) TestGetTask_Success() {
	futureDate := time.Now().Add(24 * time.Hour)
	task := suite.createTestTask("test-get-task", "Test description", models.TaskStatusInProgress, &futureDate)

	resp, body := suite.makeRequest("GET", "/tasks/test-get-task", nil)

	assert.Equal(suite.T(), http.StatusOK, resp.StatusCode)

	var retrievedTask models.Task
	err := json.Unmarshal(body, &retrievedTask)
	suite.Require().NoError(err)

	assert.Equal(suite.T(), task.ID, retrievedTask.ID)
	assert.Equal(suite.T(), task.Title, retrievedTask.Title)
	assert.Equal(suite.T(), task.Description, retrievedTask.Description)
	assert.Equal(suite.T(), task.Status, retrievedTask.Status)
	assert.NotNil(suite.T(), retrievedTask.DueDate)
}

func (suite *HandlerTestSuite) TestGetTask_NotFound() {
	resp, body := suite.makeRequest("GET", "/tasks/non-existent-task", nil)

	assert.Equal(suite.T(), http.StatusNotFound, resp.StatusCode)

	var errorResp map[string]interface{}
	err := json.Unmarshal(body, &errorResp)
	suite.Require().NoError(err)
	assert.Equal(suite.T(), "Task not found", errorResp["error"])
}

// ============================================================================
// GET ALL TASKS TESTS
// ============================================================================

func (suite *HandlerTestSuite) TestGetAllTasks_Success() {
	// Create test tasks
	futureDate1 := time.Now().Add(24 * time.Hour)
	futureDate2 := time.Now().Add(48 * time.Hour)

	suite.createTestTask("task1", "Description 1", models.TaskStatusPending, &futureDate1)
	suite.createTestTask("task2", "Description 2", models.TaskStatusInProgress, &futureDate2)
	suite.createTestTask("task3", "Description 3", models.TaskStatusCompleted, &futureDate1)

	resp, body := suite.makeRequest("GET", "/tasks", nil)

	assert.Equal(suite.T(), http.StatusOK, resp.StatusCode)

	var tasksResp models.TasksResponse
	err := json.Unmarshal(body, &tasksResp)
	suite.Require().NoError(err)

	assert.Equal(suite.T(), int64(3), tasksResp.Total)
	assert.Equal(suite.T(), 1, tasksResp.Page)
	assert.Equal(suite.T(), 10, tasksResp.Size)
	assert.Len(suite.T(), tasksResp.Tasks, 3)
}

func (suite *HandlerTestSuite) TestGetAllTasks_EmptyDatabase() {
	resp, body := suite.makeRequest("GET", "/tasks", nil)

	assert.Equal(suite.T(), http.StatusOK, resp.StatusCode)

	var tasksResp models.TasksResponse
	err := json.Unmarshal(body, &tasksResp)
	suite.Require().NoError(err)

	assert.Equal(suite.T(), int64(0), tasksResp.Total)
	assert.Equal(suite.T(), 1, tasksResp.Page)
	assert.Equal(suite.T(), 10, tasksResp.Size)
	assert.Len(suite.T(), tasksResp.Tasks, 0)
}

func (suite *HandlerTestSuite) TestGetAllTasks_WithStatusFilter() {
	futureDate := time.Now().Add(24 * time.Hour)
	suite.createTestTask("task1", "Description 1", models.TaskStatusPending, &futureDate)
	suite.createTestTask("task2", "Description 2", models.TaskStatusInProgress, &futureDate)
	suite.createTestTask("task3", "Description 3", models.TaskStatusCompleted, &futureDate)

	resp, body := suite.makeRequest("GET", "/tasks?status=pending", nil)

	assert.Equal(suite.T(), http.StatusOK, resp.StatusCode)

	var tasksResp models.TasksResponse
	err := json.Unmarshal(body, &tasksResp)
	suite.Require().NoError(err)

	assert.Equal(suite.T(), int64(1), tasksResp.Total)
	assert.Len(suite.T(), tasksResp.Tasks, 1)
	assert.Equal(suite.T(), models.TaskStatusPending, tasksResp.Tasks[0].Status)
}

func (suite *HandlerTestSuite) TestGetAllTasks_WithSearchFilter() {
	futureDate := time.Now().Add(24 * time.Hour)
	suite.createTestTask("important-task", "Important description", models.TaskStatusPending, &futureDate)
	suite.createTestTask("regular-task", "Regular description", models.TaskStatusPending, &futureDate)
	suite.createTestTask("urgent-task", "Urgent description", models.TaskStatusPending, &futureDate)

	resp, body := suite.makeRequest("GET", "/tasks?search=important", nil)

	assert.Equal(suite.T(), http.StatusOK, resp.StatusCode)

	var tasksResp models.TasksResponse
	err := json.Unmarshal(body, &tasksResp)
	suite.Require().NoError(err)

	assert.Equal(suite.T(), int64(1), tasksResp.Total)
	assert.Len(suite.T(), tasksResp.Tasks, 1)
	assert.Equal(suite.T(), "important-task", tasksResp.Tasks[0].Title)
}

func (suite *HandlerTestSuite) TestGetAllTasks_WithPagination() {
	futureDate := time.Now().Add(24 * time.Hour)
	// Create 15 tasks
	for i := 1; i <= 15; i++ {
		suite.createTestTask(fmt.Sprintf("task%d", i), fmt.Sprintf("Description %d", i), models.TaskStatusPending, &futureDate)
	}

	// Test first page with size 5
	resp, body := suite.makeRequest("GET", "/tasks?page=1&size=5", nil)

	assert.Equal(suite.T(), http.StatusOK, resp.StatusCode)

	var tasksResp models.TasksResponse
	err := json.Unmarshal(body, &tasksResp)
	suite.Require().NoError(err)

	assert.Equal(suite.T(), int64(15), tasksResp.Total)
	assert.Equal(suite.T(), 1, tasksResp.Page)
	assert.Equal(suite.T(), 5, tasksResp.Size)
	assert.Len(suite.T(), tasksResp.Tasks, 5)
}

func (suite *HandlerTestSuite) TestGetAllTasks_InvalidPaginationParams() {
	futureDate := time.Now().Add(24 * time.Hour)
	suite.createTestTask("test-task", "Description", models.TaskStatusPending, &futureDate)

	testCases := []struct {
		name string
		url  string
	}{
		{"Negative page", "/tasks?page=-1&size=10"},
		{"Zero page", "/tasks?page=0&size=10"},
		{"Negative size", "/tasks?page=1&size=-1"},
		{"Zero size", "/tasks?page=1&size=0"},
		{"Non-numeric page", "/tasks?page=abc&size=10"},
		{"Non-numeric size", "/tasks?page=1&size=xyz"},
	}

	for _, tc := range testCases {
		suite.Run(tc.name, func() {
			resp, body := suite.makeRequest("GET", tc.url, nil)

			// The handler should handle invalid params gracefully
			assert.Equal(suite.T(), http.StatusOK, resp.StatusCode)

			var tasksResp models.TasksResponse
			err := json.Unmarshal(body, &tasksResp)
			suite.Require().NoError(err)

			// Should return default values (page=1, size=10) for invalid inputs
			assert.Equal(suite.T(), int64(1), tasksResp.Total)
		})
	}
}

// ============================================================================
// UPDATE TASK TESTS
// ============================================================================

func (suite *HandlerTestSuite) TestUpdateTask_Success() {
	futureDate := time.Now().Add(24 * time.Hour)
	task := suite.createTestTask("original-title", "Original description", models.TaskStatusPending, &futureDate)

	newTitle := "updated-title"
	newDescription := "Updated description"
	newStatus := models.TaskStatusInProgress
	newDueDate := time.Now().Add(48 * time.Hour)

	updateReq := models.UpdateTaskRequest{
		Title:       &newTitle,
		Description: &newDescription,
		Status:      &newStatus,
		DueDate:     &newDueDate,
	}

	resp, body := suite.makeRequest("PUT", "/tasks/original-title", updateReq)

	assert.Equal(suite.T(), http.StatusOK, resp.StatusCode)

	var updatedTask models.Task
	err := json.Unmarshal(body, &updatedTask)
	suite.Require().NoError(err)

	assert.Equal(suite.T(), task.ID, updatedTask.ID)
	assert.Equal(suite.T(), "updated-title", updatedTask.Title)
	assert.Equal(suite.T(), "Updated description", updatedTask.Description)
	assert.Equal(suite.T(), models.TaskStatusInProgress, updatedTask.Status)
	assert.NotNil(suite.T(), updatedTask.DueDate)
	assert.True(suite.T(), updatedTask.UpdatedAt.After(task.UpdatedAt))
}

func (suite *HandlerTestSuite) TestUpdateTask_PartialUpdate() {
	futureDate := time.Now().Add(24 * time.Hour)
	task := suite.createTestTask("partial-update-task", "Original description", models.TaskStatusPending, &futureDate)

	newStatus := models.TaskStatusCompleted
	updateReq := models.UpdateTaskRequest{
		Status: &newStatus,
	}

	resp, body := suite.makeRequest("PUT", "/tasks/partial-update-task", updateReq)

	assert.Equal(suite.T(), http.StatusOK, resp.StatusCode)

	var updatedTask models.Task
	err := json.Unmarshal(body, &updatedTask)
	suite.Require().NoError(err)

	// Only status should be updated
	assert.Equal(suite.T(), task.Title, updatedTask.Title)                             // Unchanged
	assert.Equal(suite.T(), task.Description, updatedTask.Description)                 // Unchanged
	assert.Equal(suite.T(), models.TaskStatusCompleted, updatedTask.Status)            // Changed
	assert.WithinDuration(suite.T(), *task.DueDate, *updatedTask.DueDate, time.Second) // Unchanged
	assert.True(suite.T(), updatedTask.UpdatedAt.After(task.UpdatedAt))                // Should be updated
}

func (suite *HandlerTestSuite) TestUpdateTask_NotFound() {
	newTitle := "new-title"
	updateReq := models.UpdateTaskRequest{
		Title: &newTitle,
	}

	resp, body := suite.makeRequest("PUT", "/tasks/non-existent-task", updateReq)

	assert.Equal(suite.T(), http.StatusNotFound, resp.StatusCode)

	var errorResp map[string]interface{}
	err := json.Unmarshal(body, &errorResp)
	suite.Require().NoError(err)
	assert.Equal(suite.T(), "Task not found", errorResp["error"])
}

func (suite *HandlerTestSuite) TestUpdateTask_DuplicateTitle() {
	futureDate := time.Now().Add(24 * time.Hour)
	suite.createTestTask("existing-title", "Description 1", models.TaskStatusPending, &futureDate)
	suite.createTestTask("task-to-update", "Description 2", models.TaskStatusPending, &futureDate)

	// Try to update second task's title to match the first
	existingTitle := "existing-title"
	updateReq := models.UpdateTaskRequest{
		Title: &existingTitle,
	}

	resp, body := suite.makeRequest("PUT", "/tasks/task-to-update", updateReq)

	assert.Equal(suite.T(), http.StatusConflict, resp.StatusCode)

	var errorResp map[string]interface{}
	err := json.Unmarshal(body, &errorResp)
	suite.Require().NoError(err)
	assert.Contains(suite.T(), errorResp["error"], "Task with this new title already exists")
}

func (suite *HandlerTestSuite) TestUpdateTask_SameTitleUpdate() {
	futureDate := time.Now().Add(24 * time.Hour)
	suite.createTestTask("same-title-task", "Original description", models.TaskStatusPending, &futureDate)

	// Update with same title (should be allowed)
	sameTitle := "same-title-task"
	newDescription := "Updated description"
	updateReq := models.UpdateTaskRequest{
		Title:       &sameTitle,
		Description: &newDescription,
	}

	resp, body := suite.makeRequest("PUT", "/tasks/same-title-task", updateReq)

	assert.Equal(suite.T(), http.StatusOK, resp.StatusCode)

	var updatedTask models.Task
	err := json.Unmarshal(body, &updatedTask)
	suite.Require().NoError(err)

	assert.Equal(suite.T(), "same-title-task", updatedTask.Title)
	assert.Equal(suite.T(), "Updated description", updatedTask.Description)
}


// ============================================================================
// DELETE TASK TESTS
// ============================================================================

func (suite *HandlerTestSuite) TestDeleteTask_Success() {
	futureDate := time.Now().Add(24 * time.Hour)
	suite.createTestTask("task-to-delete", "Description", models.TaskStatusPending, &futureDate)

	resp, body := suite.makeRequest("DELETE", "/tasks/task-to-delete", nil)

	assert.Equal(suite.T(), http.StatusOK, resp.StatusCode)

	var response map[string]interface{}
	err := json.Unmarshal(body, &response)
	suite.Require().NoError(err)
	assert.Equal(suite.T(), "Task deleted successfully", response["message"])

	// Verify task is actually deleted
	var count int64
	suite.db.Model(&models.Task{}).Where("title = ?", "task-to-delete").Count(&count)
	assert.Equal(suite.T(), int64(0), count)
}

func (suite *HandlerTestSuite) TestDeleteTask_NotFound() {
	resp, body := suite.makeRequest("DELETE", "/tasks/non-existent-task", nil)

	assert.Equal(suite.T(), http.StatusNotFound, resp.StatusCode)

	var errorResp map[string]interface{}
	err := json.Unmarshal(body, &errorResp)
	suite.Require().NoError(err)
	assert.Equal(suite.T(), "Task not found", errorResp["error"])
}

// ============================================================================
// EDGE CASES AND SPECIAL TESTS
// ============================================================================


func (suite *HandlerTestSuite) TestHandlers_WithUnicodeCharacters() {
	futureDate := time.Now().Add(24 * time.Hour)
	taskReq := models.CreateTaskRequest{
		Title:       "测试任务",
		Description: "Unicode test: 🚀 💻 ✅",
		Status:      models.TaskStatusPending,
		DueDate:     &futureDate,
	}

	resp, body := suite.makeRequest("POST", "/tasks", taskReq)
	assert.Equal(suite.T(), http.StatusCreated, resp.StatusCode)

	var task models.Task
	err := json.Unmarshal(body, &task)
	suite.Require().NoError(err)
	assert.Equal(suite.T(), "测试任务", task.Title)
	assert.Equal(suite.T(), "Unicode test: 🚀 💻 ✅", task.Description)
}

func (suite *HandlerTestSuite) TestUpdateTask_EmptyUpdate() {
	futureDate := time.Now().Add(24 * time.Hour)
	originalTask := suite.createTestTask("empty-update-task", "Original description", models.TaskStatusPending, &futureDate)

	// Send empty update request
	updateReq := models.UpdateTaskRequest{}

	resp, body := suite.makeRequest("PUT", "/tasks/empty-update-task", updateReq)

	assert.Equal(suite.T(), http.StatusOK, resp.StatusCode)

	var updatedTask models.Task
	err := json.Unmarshal(body, &updatedTask)
	suite.Require().NoError(err)

	// All fields should remain the same except UpdatedAt
	assert.Equal(suite.T(), originalTask.Title, updatedTask.Title)
	assert.Equal(suite.T(), originalTask.Description, updatedTask.Description)
	assert.Equal(suite.T(), originalTask.Status, updatedTask.Status)
	assert.True(suite.T(), updatedTask.UpdatedAt.After(originalTask.UpdatedAt))
}

// ============================================================================
// TEST SUITE RUNNER
// ============================================================================

func TestHandlerTestSuite(t *testing.T) {
	// Skip database tests in CI if no database is available
	if os.Getenv("SKIP_DB_TESTS") == "true" {
		t.Skip("Skipping database tests")
	}

	suite.Run(t, new(HandlerTestSuite))
}
