package service

import (
	"context"
	"fmt"
	"log/slog"
	"math"
	"time"

	"todo-backend/internal/domain"

	"github.com/google/uuid"
)

const (
	TaskCacheTTL = 5 * time.Minute
)

type taskService struct {
	taskRepo  domain.TaskRepository
	cacheRepo domain.CacheRepository
}

func NewTaskService(taskRepo domain.TaskRepository, cacheRepo domain.CacheRepository) domain.TaskService {
	return &taskService{
		taskRepo:  taskRepo,
		cacheRepo: cacheRepo,
	}
}

func (s *taskService) CreateTask(ctx context.Context, userID uuid.UUID, req *domain.CreateTaskRequest) (*domain.TaskResponse, error) {
	if _, err := time.Parse("2006-01-02", req.DueDate); err != nil {
		return nil, fmt.Errorf("%w: %v", domain.ErrInvalidDateFormat, err)
	}

	// Invalidate cache before and after mutation to prevent stale cache race conditions
	s.invalidateUserCache(ctx, userID)
	defer s.invalidateUserCache(ctx, userID)

	task := &domain.Task{
		UserID:      userID,
		Title:       req.Title,
		Description: req.Description,
		Status:      domain.TaskStatus(req.Status),
		DueDate:     req.DueDate,
	}

	if err := s.taskRepo.CreateTask(ctx, task); err != nil {
		return nil, err
	}

	return &domain.TaskResponse{
		Message: "Task created successfully",
		Task:    task,
	}, nil
}

func (s *taskService) GetTaskByID(ctx context.Context, userID uuid.UUID, id uuid.UUID) (*domain.Task, error) {
	cacheKey := fmt.Sprintf("user:%s:task:%s", userID.String(), id.String())

	// Try reading from Redis cache
	if s.cacheRepo != nil {
		var cachedTask domain.Task
		if err := s.cacheRepo.Get(ctx, cacheKey, &cachedTask); err == nil {
			slog.Debug("Hit Redis cache for task by id", "task_id", id.String())
			return &cachedTask, nil
		}
	}

	task, err := s.taskRepo.GetTaskByID(ctx, id, userID)
	if err != nil {
		return nil, err
	}

	// Store in Redis cache
	if s.cacheRepo != nil {
		_ = s.cacheRepo.Set(ctx, cacheKey, task, TaskCacheTTL)
	}

	return task, nil
}

func (s *taskService) GetTasks(ctx context.Context, userID uuid.UUID, params domain.TaskQueryParams) (*domain.GetTasksResponse, error) {
	if params.Page <= 0 {
		params.Page = 1
	}
	if params.Limit <= 0 {
		params.Limit = 10
	} else if params.Limit > 100 {
		params.Limit = 100
	}

	cacheKey := fmt.Sprintf("user:%s:tasks?page=%d&limit=%d&status=%s&search=%s",
		userID.String(), params.Page, params.Limit, params.Status, params.Search)

	// Try reading from Redis cache
	if s.cacheRepo != nil {
		var cachedResponse domain.GetTasksResponse
		if err := s.cacheRepo.Get(ctx, cacheKey, &cachedResponse); err == nil {
			slog.Debug("Hit Redis cache for task list", "user_id", userID.String())
			return &cachedResponse, nil
		}
	}

	tasks, totalCount, err := s.taskRepo.GetTasks(ctx, userID, params)
	if err != nil {
		return nil, err
	}

	totalPages := 0
	if totalCount > 0 {
		totalPages = int(math.Ceil(float64(totalCount) / float64(params.Limit)))
	}

	response := &domain.GetTasksResponse{
		Tasks: tasks,
		Pagination: domain.Pagination{
			CurrentPage: params.Page,
			TotalPages:  totalPages,
			TotalTasks:  totalCount,
		},
	}

	// Store in Redis cache
	if s.cacheRepo != nil {
		_ = s.cacheRepo.Set(ctx, cacheKey, response, TaskCacheTTL)
	}

	return response, nil
}

func (s *taskService) UpdateTask(ctx context.Context, userID uuid.UUID, id uuid.UUID, req *domain.UpdateTaskRequest) (*domain.TaskResponse, error) {
	if _, err := time.Parse("2006-01-02", req.DueDate); err != nil {
		return nil, fmt.Errorf("%w: %v", domain.ErrInvalidDateFormat, err)
	}

	// Invalidate cache before and after mutation to prevent stale cache race conditions
	s.invalidateUserCache(ctx, userID)
	defer s.invalidateUserCache(ctx, userID)

	task := &domain.Task{
		ID:          id,
		UserID:      userID,
		Title:       req.Title,
		Description: req.Description,
		Status:      domain.TaskStatus(req.Status),
		DueDate:     req.DueDate,
	}

	if err := s.taskRepo.UpdateTask(ctx, task); err != nil {
		return nil, err
	}

	return &domain.TaskResponse{
		Message: "Task updated successfully",
		Task:    task,
	}, nil
}

func (s *taskService) DeleteTask(ctx context.Context, userID uuid.UUID, id uuid.UUID) (*domain.DeleteTaskResponse, error) {
	// Invalidate cache before and after mutation to prevent stale cache race conditions
	s.invalidateUserCache(ctx, userID)
	defer s.invalidateUserCache(ctx, userID)

	if err := s.taskRepo.DeleteTask(ctx, id, userID); err != nil {
		return nil, err
	}

	return &domain.DeleteTaskResponse{
		Message: "Task deleted successfully",
	}, nil
}

func (s *taskService) invalidateUserCache(ctx context.Context, userID uuid.UUID) {
	if s.cacheRepo == nil {
		return
	}
	// Use background context with a 5s timeout to ensure cache invalidation completes
	// even if the HTTP request context was cancelled.
	bgCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	pattern := fmt.Sprintf("user:%s:*", userID.String())
	if err := s.cacheRepo.DeletePattern(bgCtx, pattern); err != nil {
		slog.Error("Failed to invalidate user cache pattern", "pattern", pattern, "error", err)
	}
}
