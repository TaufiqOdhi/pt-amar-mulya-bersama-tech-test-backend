package handler_test

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"todo-backend/internal/domain"
	"todo-backend/internal/handler"
	"todo-backend/internal/middleware"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

type MockTaskService struct {
	mock.Mock
}

func (m *MockTaskService) CreateTask(ctx context.Context, userID uuid.UUID, req *domain.CreateTaskRequest) (*domain.TaskResponse, error) {
	args := m.Called(ctx, userID, req)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.TaskResponse), args.Error(1)
}

func (m *MockTaskService) GetTaskByID(ctx context.Context, userID uuid.UUID, id uuid.UUID) (*domain.Task, error) {
	args := m.Called(ctx, userID, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.Task), args.Error(1)
}

func (m *MockTaskService) GetTasks(ctx context.Context, userID uuid.UUID, params domain.TaskQueryParams) (*domain.GetTasksResponse, error) {
	args := m.Called(ctx, userID, params)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.GetTasksResponse), args.Error(1)
}

func (m *MockTaskService) UpdateTask(ctx context.Context, userID uuid.UUID, id uuid.UUID, req *domain.UpdateTaskRequest) (*domain.TaskResponse, error) {
	args := m.Called(ctx, userID, id, req)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.TaskResponse), args.Error(1)
}

func (m *MockTaskService) DeleteTask(ctx context.Context, userID uuid.UUID, id uuid.UUID) (*domain.DeleteTaskResponse, error) {
	args := m.Called(ctx, userID, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.DeleteTaskResponse), args.Error(1)
}

func setupTaskRouter(mockService *MockTaskService, userID uuid.UUID) *gin.Engine {
	gin.SetMode(gin.TestMode)
	taskHandler := handler.NewTaskHandler(mockService)

	router := gin.Default()
	taskGroup := router.Group("/tasks")
	taskGroup.Use(func(c *gin.Context) {
		c.Set(middleware.UserIDKey, userID)
		c.Next()
	})

	taskGroup.POST("", taskHandler.CreateTask)
	taskGroup.GET("", taskHandler.GetTasks)
	taskGroup.GET("/:id", taskHandler.GetTaskByID)
	taskGroup.PUT("/:id", taskHandler.UpdateTask)
	taskGroup.DELETE("/:id", taskHandler.DeleteTask)

	return router
}

func TestCreateTask_Success(t *testing.T) {
	mockService := new(MockTaskService)
	userID := uuid.New()
	router := setupTaskRouter(mockService, userID)

	reqBody := domain.CreateTaskRequest{
		Title:       "Test Task",
		Description: "Test Description",
		Status:      "pending",
		DueDate:     "2026-08-01",
	}

	taskID := uuid.New()
	expectedRes := &domain.TaskResponse{
		Message: "Task created successfully",
		Task: &domain.Task{
			ID:          taskID,
			UserID:      userID,
			Title:       "Test Task",
			Description: "Test Description",
			Status:      "pending",
			DueDate:     "2026-08-01",
		},
	}

	mockService.On("CreateTask", mock.Anything, userID, &reqBody).Return(expectedRes, nil)

	body, _ := json.Marshal(reqBody)
	req, _ := http.NewRequest(http.MethodPost, "/tasks", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	resp := httptest.NewRecorder()

	router.ServeHTTP(resp, req)

	assert.Equal(t, http.StatusCreated, resp.Code)
	mockService.AssertExpectations(t)
}

func TestGetTasks_Success(t *testing.T) {
	mockService := new(MockTaskService)
	userID := uuid.New()
	router := setupTaskRouter(mockService, userID)

	queryParams := domain.TaskQueryParams{
		Status: "",
		Page:   0,
		Limit:  0,
		Search: "",
	}

	expectedRes := &domain.GetTasksResponse{
		Tasks: []*domain.Task{
			{
				ID:          uuid.New(),
				Title:       "Task 1",
				Description: "Desc 1",
				Status:      "pending",
				DueDate:     "2026-08-01",
			},
		},
		Pagination: domain.Pagination{
			CurrentPage: 1,
			TotalPages:  1,
			TotalTasks:  1,
		},
	}

	mockService.On("GetTasks", mock.Anything, userID, queryParams).Return(expectedRes, nil)

	req, _ := http.NewRequest(http.MethodGet, "/tasks", nil)
	resp := httptest.NewRecorder()

	router.ServeHTTP(resp, req)

	assert.Equal(t, http.StatusOK, resp.Code)
	mockService.AssertExpectations(t)
}

func TestGetTaskByID_Success(t *testing.T) {
	mockService := new(MockTaskService)
	userID := uuid.New()
	taskID := uuid.New()
	router := setupTaskRouter(mockService, userID)

	expectedTask := &domain.Task{
		ID:          taskID,
		UserID:      userID,
		Title:       "Test Task",
		Description: "Test Description",
		Status:      "pending",
		DueDate:     "2026-08-01",
	}

	mockService.On("GetTaskByID", mock.Anything, userID, taskID).Return(expectedTask, nil)

	req, _ := http.NewRequest(http.MethodGet, "/tasks/"+taskID.String(), nil)
	resp := httptest.NewRecorder()

	router.ServeHTTP(resp, req)

	assert.Equal(t, http.StatusOK, resp.Code)
	mockService.AssertExpectations(t)
}

func TestDeleteTask_Success(t *testing.T) {
	mockService := new(MockTaskService)
	userID := uuid.New()
	taskID := uuid.New()
	router := setupTaskRouter(mockService, userID)

	expectedRes := &domain.DeleteTaskResponse{
		Message: "Task deleted successfully",
	}

	mockService.On("DeleteTask", mock.Anything, userID, taskID).Return(expectedRes, nil)

	req, _ := http.NewRequest(http.MethodDelete, "/tasks/"+taskID.String(), nil)
	resp := httptest.NewRecorder()

	router.ServeHTTP(resp, req)

	assert.Equal(t, http.StatusOK, resp.Code)
	mockService.AssertExpectations(t)
}

func TestUpdateTask_NotFound(t *testing.T) {
	mockService := new(MockTaskService)
	userID := uuid.New()
	taskID := uuid.New()
	router := setupTaskRouter(mockService, userID)

	reqBody := domain.UpdateTaskRequest{
		Title:       "Updated Title",
		Description: "Updated Desc",
		Status:      "completed",
		DueDate:     "2026-08-01",
	}

	mockService.On("UpdateTask", mock.Anything, userID, taskID, &reqBody).Return(nil, domain.ErrTaskNotFound)

	body, _ := json.Marshal(reqBody)
	req, _ := http.NewRequest(http.MethodPut, "/tasks/"+taskID.String(), bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	resp := httptest.NewRecorder()

	router.ServeHTTP(resp, req)

	assert.Equal(t, http.StatusNotFound, resp.Code)
	mockService.AssertExpectations(t)
}
