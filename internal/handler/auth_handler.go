package handler

import (
	"errors"
	"log/slog"
	"net/http"

	"todo-backend/internal/domain"

	"github.com/gin-gonic/gin"
)

type AuthHandler struct {
	authService domain.AuthService
}

func NewAuthHandler(authService domain.AuthService) *AuthHandler {
	return &AuthHandler{authService: authService}
}

func (h *AuthHandler) Register(c *gin.Context) {
	var req domain.RegisterRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	res, err := h.authService.Register(c.Request.Context(), &req)
	if err != nil {
		if errors.Is(err, domain.ErrEmailAlreadyRegistered) {
			c.JSON(http.StatusConflict, gin.H{"error": domain.ErrEmailAlreadyRegistered.Error()})
			return
		}
		slog.Error("Failed to register user", "error", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Internal server error"})
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"message": "User registered successfully",
		"user":    res,
	})
}

func (h *AuthHandler) Login(c *gin.Context) {
	var req domain.LoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	res, err := h.authService.Login(c.Request.Context(), &req)
	if err != nil {
		if errors.Is(err, domain.ErrInvalidEmailOrPassword) {
			c.JSON(http.StatusUnauthorized, gin.H{"error": domain.ErrInvalidEmailOrPassword.Error()})
			return
		}
		slog.Error("Failed to login user", "error", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Internal server error"})
		return
	}

	c.JSON(http.StatusOK, res)
}
