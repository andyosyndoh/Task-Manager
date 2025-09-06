package handlers_test

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"
	"time"

	routes "task/backend/config"
	"task/backend/database"
	"task/backend/models"

	"github.com/gofiber/fiber/v2"
	"github.com/joho/godotenv"
	"github.com/stretchr/testify/assert"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

var app *fiber.App
var testDB *gorm.DB

func TestMain(m *testing.M) {
	// Load .env file
	err := godotenv.Load("../../../.env")
	if err != nil {
		log.Fatalf("Error loading .env file: %v", err)
	}

	// Connect to test database
	testDB = connectTestDB()
	database.DB = testDB // Set the global DB to the test DB

	// Run migrations for the test database
	runTestMigrations(testDB)

	// Setup Fiber app
	app = fiber.New()
	routes.Setup(app) // Use the SetupRoutes from your config package

	// Run tests
	code := m.Run()

	// Teardown
	clearTestDB(testDB)
	os.Exit(code)
}

func connectTestDB() *gorm.DB {
	host := "localhost"
	port := 5432
	user := os.Getenv("DB_USER")
	password := os.Getenv("DB_PASSWORD")
	dbname := "task_manager"

	dsn := fmt.Sprintf("host=%s port=%d user=%s password=%s dbname=%s sslmode=disable",
		host, port, user, password, dbname)

	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		log.Fatalf("Failed to connect to test database: %v", err)
	}
	return db
}

func runTestMigrations(db *gorm.DB) {
	log.Println("Running test database migrations...")
	// Drop existing tables to ensure a clean state for tests
	db.Migrator().DropTable(&models.Task{})

	err := db.AutoMigrate(&models.Task{})
	if err != nil {
		log.Fatalf("Failed to run test migrations: %v", err)
	}
	log.Println("✅ Test database migrations completed successfully")
}

func clearTestDB(db *gorm.DB) {
	db.Migrator().DropTable(&models.Task{})
}

// Helper function to make HTTP requests
func makeRequest(method, url string, body interface{}) (*http.Response, error) {
	var req *http.Request
	var err error

	if body != nil {
		jsonBody, _ := json.Marshal(body)
		req = httptest.NewRequest(method, url, bytes.NewBuffer(jsonBody))
		req.Header.Set("Content-Type", "application/json")
	} else {
		req = httptest.NewRequest(method, url, nil)
	}

	resp, err := app.Test(req, -1) // -1 for no timeout
	if err != nil {
		return nil, err
	}
	return resp, nil
}

func TestCreateTask(t *testing.T) {
	// Test case 1: Valid task creation
	taskRequest := models.CreateTaskRequest{
		Title:       "TestTask1",
		Description: "Description for TestTask1",
		Status:      models.TaskStatusPending,
		DueDate:     func() *time.Time { t := time.Now().Add(24 * time.Hour); return &t }(),
	}
	resp, err := makeRequest("POST", "/tasks", taskRequest)
	assert.NoError(t, err)
	assert.Equal(t, http.StatusCreated, resp.StatusCode)

	var createdTask models.Task
	err = json.NewDecoder(resp.Body).Decode(&createdTask)
	assert.NoError(t, err)
	assert.Equal(t, taskRequest.Title, createdTask.Title)
	assert.Equal(t, taskRequest.Description, createdTask.Description)
	assert.Equal(t, models.TaskStatusPending, createdTask.Status) // Should default to pending

	// Test case 2: Duplicate title
	resp, err = makeRequest("POST", "/tasks", taskRequest)
	assert.NoError(t, err)
	assert.Equal(t, http.StatusConflict, resp.StatusCode)

	// Test case 3: Invalid due_date (past date)
	invalidTaskRequest := models.CreateTaskRequest{
		Title:       "InvalidTask1",
		Description: "Description for InvalidTask1",
		DueDate:     func() *time.Time { t := time.Now().Add(-24 * time.Hour); return &t }(),
	}
	resp, err = makeRequest("POST", "/tasks", invalidTaskRequest)
	assert.NoError(t, err)
	assert.Equal(t, http.StatusBadRequest, resp.StatusCode)

	// Test case 4: Title with spaces
	taskWithSpaces := models.CreateTaskRequest{
		Title:       "Task With Spaces",
		Description: "Description",
		DueDate:     func() *time.Time { t := time.Now().Add(24 * time.Hour); return &t }(),
	}
	resp, err = makeRequest("POST", "/tasks", taskWithSpaces)
	assert.NoError(t, err)
	assert.Equal(t, http.StatusBadRequest, resp.StatusCode)
	// Read the response body
	bodyBytes, err := io.ReadAll(resp.Body)
	assert.NoError(t, err)  // Make sure there was no error reading the body
	defer resp.Body.Close() // IMPORTANT: Always close the body

	// Now you can assert against the body content as a string
	assert.Contains(t, string(bodyBytes), "nospaces")
}

func TestGetTask(t *testing.T) {
	// Create a task first
	taskRequest := models.CreateTaskRequest{
		Title:       "TaskToGet",
		Description: "Description for TaskToGet",
		DueDate:     func() *time.Time { t := time.Now().Add(24 * time.Hour); return &t }(),
	}
	resp, err := makeRequest("POST", "/tasks", taskRequest)
	assert.NoError(t, err)
	assert.Equal(t, http.StatusCreated, resp.StatusCode)

	// Test case 1: Get existing task
	resp, err = makeRequest("GET", "/tasks/TaskToGet", nil)
	assert.NoError(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)

	var retrievedTask models.Task
	err = json.NewDecoder(resp.Body).Decode(&retrievedTask)
	assert.NoError(t, err)
	assert.Equal(t, taskRequest.Title, retrievedTask.Title)

	// Test case 2: Get non-existent task
	resp, err = makeRequest("GET", "/tasks/NonExistentTask", nil)
	assert.NoError(t, err)
	assert.Equal(t, http.StatusNotFound, resp.StatusCode)
}

func TestGetAllTasks(t *testing.T) {
	// Clear existing tasks
	clearTestDB(testDB)
	runTestMigrations(testDB) // Re-run migrations to ensure table exists

	// Create multiple tasks
	tasksToCreate := []models.CreateTaskRequest{
		{Title: "TaskA", Status: models.TaskStatusPending, DueDate: func() *time.Time { t := time.Now().Add(24 * time.Hour); return &t }()},
		{Title: "TaskB", Status: models.TaskStatusInProgress, DueDate: func() *time.Time { t := time.Now().Add(48 * time.Hour); return &t }()},
		{Title: "TaskC", Status: models.TaskStatusCompleted, DueDate: func() *time.Time { t := time.Now().Add(72 * time.Hour); return &t }()},
	}
	for _, tr := range tasksToCreate {
		resp, err := makeRequest("POST", "/tasks", tr)
		assert.NoError(t, err)
		assert.Equal(t, http.StatusCreated, resp.StatusCode)
	}

	// Test case 1: Get all tasks (no filters, no pagination)
	resp, err := makeRequest("GET", "/tasks", nil)
	assert.NoError(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
	var tasksResponse models.TasksResponse
	err = json.NewDecoder(resp.Body).Decode(&tasksResponse)
	assert.NoError(t, err)
	assert.Equal(t, 3, len(tasksResponse.Tasks))
	assert.Equal(t, int64(3), tasksResponse.Total)
	assert.Equal(t, 1, tasksResponse.Page)
	assert.Equal(t, 10, tasksResponse.Size)

	// Test case 2: Filter by status
	resp, err = makeRequest("GET", "/tasks?status=pending", nil)
	assert.NoError(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
	err = json.NewDecoder(resp.Body).Decode(&tasksResponse)
	assert.NoError(t, err)
	assert.Equal(t, 1, len(tasksResponse.Tasks))
	assert.Equal(t, int64(1), tasksResponse.Total)
	assert.Equal(t, "TaskA", tasksResponse.Tasks[0].Title)

	// Test case 3: Filter by due_date (tasks due on or before TaskB's due date)
	resp, err = makeRequest("GET", fmt.Sprintf("/tasks?due_date=%s", time.Now().Add(48*time.Hour).Format("2006-01-02")), nil)
	assert.NoError(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
	err = json.NewDecoder(resp.Body).Decode(&tasksResponse)
	assert.NoError(t, err)
	assert.Equal(t, 2, len(tasksResponse.Tasks)) // TaskA and TaskB
	assert.Equal(t, int64(2), tasksResponse.Total)

	// Test case 4: Pagination - page 1, size 2
	resp, err = makeRequest("GET", "/tasks?page=1&size=2", nil)
	assert.NoError(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
	err = json.NewDecoder(resp.Body).Decode(&tasksResponse)
	assert.NoError(t, err)
	assert.Equal(t, 2, len(tasksResponse.Tasks))
	assert.Equal(t, int64(3), tasksResponse.Total)
	assert.Equal(t, 1, tasksResponse.Page)
	assert.Equal(t, 2, tasksResponse.Size)

	// Test case 5: Pagination - page 2, size 2
	resp, err = makeRequest("GET", "/tasks?page=2&size=2", nil)
	assert.NoError(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
	err = json.NewDecoder(resp.Body).Decode(&tasksResponse)
	assert.NoError(t, err)
	assert.Equal(t, 1, len(tasksResponse.Tasks))
	assert.Equal(t, int64(3), tasksResponse.Total)
	assert.Equal(t, 2, tasksResponse.Page)
	assert.Equal(t, 2, tasksResponse.Size)

	// Test case 6: Search by title
	resp, err = makeRequest("GET", "/tasks?search=TaskA", nil)
	assert.NoError(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
	err = json.NewDecoder(resp.Body).Decode(&tasksResponse)
	assert.NoError(t, err)
	assert.Equal(t, 1, len(tasksResponse.Tasks))
	assert.Equal(t, int64(1), tasksResponse.Total)
	assert.Equal(t, "TaskA", tasksResponse.Tasks[0].Title)

	// Test case 7: Search by partial title (case-insensitive)
	resp, err = makeRequest("GET", "/tasks?search=task", nil)
	assert.NoError(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
	err = json.NewDecoder(resp.Body).Decode(&tasksResponse)
	assert.NoError(t, err)
	assert.Equal(t, 3, len(tasksResponse.Tasks)) // All tasks contain "Task"
	assert.Equal(t, int64(3), tasksResponse.Total)
}

func TestUpdateTask(t *testing.T) {
	// Create a task first
	taskRequest := models.CreateTaskRequest{
		Title:       "TaskToUpdate",
		Description: "Original description",
		Status:      models.TaskStatusPending,
		DueDate:     func() *time.Time { t := time.Now().Add(24 * time.Hour); return &t }(),
	}
	resp, err := makeRequest("POST", "/tasks", taskRequest)
	assert.NoError(t, err)
	assert.Equal(t, http.StatusCreated, resp.StatusCode)

	// Test case 1: Update description and status
	newDescription := "Updated description"
	newStatus := models.TaskStatusCompleted
	updateRequest := models.UpdateTaskRequest{
		Description: &newDescription,
		Status:      &newStatus,
	}
	resp, err = makeRequest("PUT", "/tasks/TaskToUpdate", updateRequest)
	assert.NoError(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)

	var updatedTask models.Task
	err = json.NewDecoder(resp.Body).Decode(&updatedTask)
	assert.NoError(t, err)
	assert.Equal(t, newDescription, updatedTask.Description)
	assert.Equal(t, newStatus, updatedTask.Status)
	assert.True(t, updatedTask.UpdatedAt.After(updatedTask.CreatedAt))

	// Test case 2: Update title to an existing title (should conflict)
	existingTaskTitle := "TaskToGet" // Assuming this task exists from previous tests
	updateTitleRequest := models.UpdateTaskRequest{
		Title: &existingTaskTitle,
	}
	resp, err = makeRequest("PUT", "/tasks/TaskToUpdate", updateTitleRequest)
	assert.NoError(t, err)
	assert.Equal(t, http.StatusConflict, resp.StatusCode)

	// Test case 3: Update non-existent task
	resp, err = makeRequest("PUT", "/tasks/NonExistentTask", updateRequest)
	assert.NoError(t, err)
	assert.Equal(t, http.StatusNotFound, resp.StatusCode)

	// Test case 4: Update with invalid due_date (past date)
	pastDate := time.Now().Add(-24 * time.Hour)
	updateInvalidDueDateRequest := models.UpdateTaskRequest{
		DueDate: &pastDate,
	}
	resp, err = makeRequest("PUT", "/tasks/TaskToUpdate", updateInvalidDueDateRequest)
	assert.NoError(t, err)
	assert.Equal(t, http.StatusBadRequest, resp.StatusCode)
	// Read the response body
	bodyBytes, err := io.ReadAll(resp.Body)
	assert.NoError(t, err)  // Make sure there was no error reading the body
	defer resp.Body.Close() // IMPORTANT: Always close the body

	// Now you can assert against the body content as a string
	assert.Contains(t, string(bodyBytes), "nospaces")
}

func TestDeleteTask(t *testing.T) {
	// Create a task first
	taskRequest := models.CreateTaskRequest{
		Title:       "TaskToDelete",
		Description: "Description for TaskToDelete",
		Status:      models.TaskStatusPending,
		DueDate:     func() *time.Time { t := time.Now().Add(24 * time.Hour); return &t }(),
	}
	resp, err := makeRequest("POST", "/tasks", taskRequest)
	assert.NoError(t, err)
	assert.Equal(t, http.StatusCreated, resp.StatusCode)

	// Test case 1: Delete existing task
	resp, err = makeRequest("DELETE", "/tasks/TaskToDelete", nil)
	assert.NoError(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)

	// Verify task is deleted
	resp, err = makeRequest("GET", "/tasks/TaskToDelete", nil)
	assert.NoError(t, err)
	assert.Equal(t, http.StatusNotFound, resp.StatusCode)

	// Test case 2: Delete non-existent task
	resp, err = makeRequest("DELETE", "/tasks/NonExistentTask", nil)
	assert.NoError(t, err)
	assert.Equal(t, http.StatusNotFound, resp.StatusCode)
}
