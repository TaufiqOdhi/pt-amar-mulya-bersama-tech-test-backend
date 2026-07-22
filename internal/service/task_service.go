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

	// Invalidate user's tasks cache
	s.invalidateUserCache(ctx, userID)

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
	}

	cacheKey := fmt.Sprintf("user:%s:tasks:p%d:l%d:s%s:q%s",
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

	totalPages := int(math.Ceil(float64(totalCount) / float64(params.Limit)))
	if totalPages == 0 {
		totalPages = 1
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
	task, err := s.taskRepo.GetTaskByID(ctx, id, userID)
	if err != nil {
		return nil, err
	}

	task.Title = req.Title
	task.Description = req.Description
	task.Status = domain.TaskStatus(req.Status)
	task.DueDate = req.DueDate

	if err := s.taskRepo.UpdateTask(ctx, task); err != nil {
		return nil, err
	}

	// Invalidate cache
	s.invalidateUserCache(ctx, userID)

	return &domain.TaskResponse{
		Message: "Task updated successfully",
		Task:    task,
	}, nil
}

func (s *taskService) DeleteTask(ctx context.Context, userID uuid.UUID, id uuid.UUID) (*domain.DeleteTaskResponse, error) {
	if err := s.taskRepo.DeleteTask(ctx, id, userID); err != nil {
		return nil, err
	}

	// Invalidate cache
	s.invalidateUserCache(ctx, userID)

	return &domain.DeleteTaskResponse{
		Message: "Task deleted successfully",
	}, nil
}

func (s *taskService) invalidateUserCache(ctx context.Context, userID uuid.UUID) {
	if s.cacheRepo == nil {
		return
	}
	pattern := fmt.Sprintf("user:%s:*", userID.String())
	if err := s.cacheRepo.DeletePattern(ctx, pattern); err != nil {
		slog.Error("Failed to invalidate user cache pattern", "pattern", pattern, "error", err)
	}
}
