package models

import (
	"database/sql/driver"
	"time"
)

type TaskStatus string

const (
	TaskStatusPending    TaskStatus = "pending"
	TaskStatusInProgress TaskStatus = "in_progress"
	TaskStatusCompleted  TaskStatus = "completed"
)


type Task struct {
	ID          int          `json:"id" gorm:"primaryKey"`
	Title       string       `json:"title" gorm:"unique;not null"`
	Description string       `json:"description"`
	Status      TaskStatus   `json:"status" gorm:"default:'pending'"`
	DueDate     *time.Time   `json:"due_date"`
	CreatedAt   time.Time    `json:"created_at"`
	UpdatedAt   time.Time    `json:"updated_at"`
}

type CreateTaskRequest struct {
	Title       string       `json:"title" validate:"required,min=1,max=200,nospaces"`
	Description string       `json:"description"`
	Status      TaskStatus   `json:"status"`
	DueDate     *time.Time   `json:"due_date" validate:"required,future"`
}

type UpdateTaskRequest struct {
	Title       *string       `json:"title,omitempty"`
	Description *string       `json:"description,omitempty"`
	Status      *TaskStatus   `json:"status,omitempty"`
	DueDate     *time.Time    `json:"due_date,omitempty"`
}

type TasksResponse struct {
	Tasks []Task `json:"tasks"`
	Total int    `json:"total"`
	Page  int    `json:"page"`
	Size  int    `json:"size"`
}

// Implement driver.Valuer interface for TaskStatus
func (ts TaskStatus) Value() (driver.Value, error) {
	return string(ts), nil
}
