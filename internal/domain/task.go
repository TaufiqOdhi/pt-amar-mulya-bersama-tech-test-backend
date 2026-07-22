package domain

import (
	"context"
	"time"

	"github.com/google/uuid"
)

type TaskStatus string

const (
	TaskStatusPending   TaskStatus = "pending"
	TaskStatusCompleted TaskStatus = "completed"
)

type Task struct {
	ID          uuid.UUID  `json:"id"`
	UserID      uuid.UUID  `json:"user_id,omitempty"`
	Title       string     `json:"title"`
	Description string     `json:"description"`
	Status      TaskStatus `json:"status"`
	DueDate     string     `json:"due_date"`
	CreatedAt   time.Time  `json:"created_at,omitempty"`
	UpdatedAt   time.Time  `json:"updated_at,omitempty"`
	DeletedAt   *time.Time `json:"deleted_at,omitempty"`
}

type CreateTaskRequest struct {
	Title       string `json:"title" binding:"required"`
	Description string `json:"description"`
	Status      string `json:"status" binding:"required,oneof=pending completed"`
	DueDate     string `json:"due_date" binding:"required"`
}

type UpdateTaskRequest struct {
	Title       string `json:"title" binding:"required"`
	Description string `json:"description"`
	Status      string `json:"status" binding:"required,oneof=pending completed"`
	DueDate     string `json:"due_date" binding:"required"`
}

type TaskQueryParams struct {
	Status string `form:"status"`
	Page   int    `form:"page"`
	Limit  int    `form:"limit"`
	Search string `form:"search"`
}

type Pagination struct {
	CurrentPage int `json:"current_page"`
	TotalPages  int `json:"total_pages"`
	TotalTasks  int `json:"total_tasks"`
}

type GetTasksResponse struct {
	Tasks      []*Task    `json:"tasks"`
	Pagination Pagination `json:"pagination"`
}

type TaskResponse struct {
	Message string `json:"message,omitempty"`
	Task    *Task  `json:"task,omitempty"`
}

type DeleteTaskResponse struct {
	Message string `json:"message"`
}

type TaskRepository interface {
	CreateTask(ctx context.Context, task *Task) error
	GetTaskByID(ctx context.Context, id uuid.UUID, userID uuid.UUID) (*Task, error)
	GetTasks(ctx context.Context, userID uuid.UUID, params TaskQueryParams) ([]*Task, int, error)
	UpdateTask(ctx context.Context, task *Task) error
	DeleteTask(ctx context.Context, id uuid.UUID, userID uuid.UUID) error
}

type CacheRepository interface {
	Get(ctx context.Context, key string, dest interface{}) error
	Set(ctx context.Context, key string, value interface{}, expiration time.Duration) error
	DeletePattern(ctx context.Context, pattern string) error
}

type TaskService interface {
	CreateTask(ctx context.Context, userID uuid.UUID, req *CreateTaskRequest) (*TaskResponse, error)
	GetTaskByID(ctx context.Context, userID uuid.UUID, id uuid.UUID) (*Task, error)
	GetTasks(ctx context.Context, userID uuid.UUID, params TaskQueryParams) (*GetTasksResponse, error)
	UpdateTask(ctx context.Context, userID uuid.UUID, id uuid.UUID, req *UpdateTaskRequest) (*TaskResponse, error)
	DeleteTask(ctx context.Context, userID uuid.UUID, id uuid.UUID) (*DeleteTaskResponse, error)
}
