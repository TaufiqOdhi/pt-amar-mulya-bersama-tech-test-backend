package handler

import (
	"errors"
	"log/slog"
	"net/http"

	"todo-backend/internal/domain"
	"todo-backend/internal/middleware"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

type TaskHandler struct {
	taskService domain.TaskService
}

func NewTaskHandler(taskService domain.TaskService) *TaskHandler {
	return &TaskHandler{taskService: taskService}
}

func (h *TaskHandler) CreateTask(c *gin.Context) {
	userID, ok := middleware.GetUserIDFromContext(c)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
		return
	}

	var req domain.CreateTaskRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	res, err := h.taskService.CreateTask(c.Request.Context(), userID, &req)
	if err != nil {
		if errors.Is(err, domain.ErrInvalidDateFormat) {
			c.JSON(http.StatusBadRequest, gin.H{"error": domain.ErrInvalidDateFormat.Error()})
			return
		}
		slog.Error("Failed to create task", "error", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Internal server error"})
		return
	}

	c.JSON(http.StatusCreated, res)
}

func (h *TaskHandler) GetTasks(c *gin.Context) {
	userID, ok := middleware.GetUserIDFromContext(c)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
		return
	}

	var params domain.TaskQueryParams
	if err := c.ShouldBindQuery(&params); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	res, err := h.taskService.GetTasks(c.Request.Context(), userID, params)
	if err != nil {
		slog.Error("Failed to fetch tasks", "error", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Internal server error"})
		return
	}

	c.JSON(http.StatusOK, res)
}

func (h *TaskHandler) GetTaskByID(c *gin.Context) {
	userID, ok := middleware.GetUserIDFromContext(c)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
		return
	}

	idParam := c.Param("id")
	taskID, err := uuid.Parse(idParam)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid task ID format"})
		return
	}

	res, err := h.taskService.GetTaskByID(c.Request.Context(), userID, taskID)
	if err != nil {
		if errors.Is(err, domain.ErrTaskNotFound) {
			c.JSON(http.StatusNotFound, gin.H{"error": domain.ErrTaskNotFound.Error()})
			return
		}
		slog.Error("Failed to fetch task by ID", "error", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Internal server error"})
		return
	}

	c.JSON(http.StatusOK, res)
}

func (h *TaskHandler) UpdateTask(c *gin.Context) {
	userID, ok := middleware.GetUserIDFromContext(c)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
		return
	}

	idParam := c.Param("id")
	taskID, err := uuid.Parse(idParam)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid task ID format"})
		return
	}

	var req domain.UpdateTaskRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	res, err := h.taskService.UpdateTask(c.Request.Context(), userID, taskID, &req)
	if err != nil {
		if errors.Is(err, domain.ErrTaskNotFound) {
			c.JSON(http.StatusNotFound, gin.H{"error": domain.ErrTaskNotFound.Error()})
			return
		}
		if errors.Is(err, domain.ErrInvalidDateFormat) {
			c.JSON(http.StatusBadRequest, gin.H{"error": domain.ErrInvalidDateFormat.Error()})
			return
		}
		slog.Error("Failed to update task", "error", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Internal server error"})
		return
	}

	c.JSON(http.StatusOK, res)
}

func (h *TaskHandler) DeleteTask(c *gin.Context) {
	userID, ok := middleware.GetUserIDFromContext(c)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
		return
	}

	idParam := c.Param("id")
	taskID, err := uuid.Parse(idParam)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid task ID format"})
		return
	}

	res, err := h.taskService.DeleteTask(c.Request.Context(), userID, taskID)
	if err != nil {
		if errors.Is(err, domain.ErrTaskNotFound) {
			c.JSON(http.StatusNotFound, gin.H{"error": domain.ErrTaskNotFound.Error()})
			return
		}
		slog.Error("Failed to delete task", "error", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Internal server error"})
		return
	}

	c.JSON(http.StatusOK, res)
}
