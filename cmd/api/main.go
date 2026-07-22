package main

import (
	"context"
	"errors"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"todo-backend/internal/config"
	"todo-backend/internal/domain"
	"todo-backend/internal/handler"
	"todo-backend/internal/middleware"
	postgresRepo "todo-backend/internal/repository/postgres"
	redisRepo "todo-backend/internal/repository/redis"
	"todo-backend/internal/service"
	"todo-backend/pkg/database"
	"todo-backend/pkg/logger"
	pkgRedis "todo-backend/pkg/redis"

	"github.com/gin-gonic/gin"
	"github.com/redis/go-redis/v9"
)

func main() {
	cfg := config.LoadConfig()
	logger.InitLogger(cfg.AppEnv)

	slog.Info("Starting To-Do List Backend API...", "env", cfg.AppEnv, "port", cfg.Port)

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	// Connect PostgreSQL
	dbPool, err := database.NewPostgresPool(ctx, cfg.DatabaseURL)
	if err != nil {
		slog.Error("Database connection failed", "error", err)
		os.Exit(1)
	}
	defer dbPool.Close()

	// Run Database Migrations
	if err := database.RunMigrations(ctx, dbPool, "migrations"); err != nil {
		slog.Error("Migration failure", "error", err)
		os.Exit(1)
	}

	// Connect Redis (Optional graceful fallback)
	var redisClient *redis.Client
	redisClient, err = pkgRedis.NewRedisClient(ctx, cfg.RedisURL)
	if err != nil {
		slog.Warn("Redis connection failed, continuing without caching", "error", err)
	} else {
		defer redisClient.Close()
	}

	// Setup Repositories
	userRepo := postgresRepo.NewUserRepo(dbPool)
	taskRepo := postgresRepo.NewTaskRepo(dbPool)

	var cacheRepo domain.CacheRepository
	if redisClient != nil {
		cacheRepo = redisRepo.NewCacheRepo(redisClient)
	}

	// Setup Services
	authService := service.NewAuthService(userRepo, cfg.JWTSecret, cfg.JWTExpirationHours)
	taskService := service.NewTaskService(taskRepo, cacheRepo)

	// Setup Handlers
	authHandler := handler.NewAuthHandler(authService)
	taskHandler := handler.NewTaskHandler(taskService)

	// Router Setup
	if cfg.AppEnv == "production" {
		gin.SetMode(gin.ReleaseMode)
	}

	router := gin.New()
	router.Use(gin.Recovery())
	router.Use(middleware.LoggerMiddleware())

	// CORS Middleware
	router.Use(func(c *gin.Context) {
		origin := c.Request.Header.Get("Origin")
		if origin != "" {
			c.Writer.Header().Set("Access-Control-Allow-Origin", origin)
		} else {
			c.Writer.Header().Set("Access-Control-Allow-Origin", "*")
		}
		c.Writer.Header().Set("Access-Control-Allow-Credentials", "true")
		c.Writer.Header().Set("Access-Control-Allow-Headers", "Content-Type, Content-Length, Accept-Encoding, X-CSRF-Token, Authorization, accept, origin, Cache-Control, X-Requested-With")
		c.Writer.Header().Set("Access-Control-Allow-Methods", "POST, OPTIONS, GET, PUT, DELETE")

		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(http.StatusNoContent)
			return
		}
		c.Next()
	})

	// Public Routes
	router.GET("/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"status": "ok", "timestamp": time.Now()})
	})

	authGroup := router.Group("/auth")
	{
		authGroup.POST("/register", authHandler.Register)
		authGroup.POST("/login", authHandler.Login)
	}

	// Protected Task Routes
	taskGroup := router.Group("/tasks")
	taskGroup.Use(middleware.AuthMiddleware(cfg.JWTSecret))
	{
		taskGroup.POST("", taskHandler.CreateTask)
		taskGroup.GET("", taskHandler.GetTasks)
		taskGroup.GET("/:id", taskHandler.GetTaskByID)
		taskGroup.PUT("/:id", taskHandler.UpdateTask)
		taskGroup.DELETE("/:id", taskHandler.DeleteTask)
	}

	srv := &http.Server{
		Addr:    ":" + cfg.Port,
		Handler: router,
	}

	go func() {
		slog.Info("Server listening on port", "port", cfg.Port)
		if err := srv.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			slog.Error("Server listen error", "error", err)
		}
	}()

	// Graceful Shutdown
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit
	slog.Info("Shutting down server gracefully...")

	shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer shutdownCancel()

	if err := srv.Shutdown(shutdownCtx); err != nil {
		slog.Error("Server forced to shutdown", "error", err)
	}

	slog.Info("Server exiting cleanly")
}
