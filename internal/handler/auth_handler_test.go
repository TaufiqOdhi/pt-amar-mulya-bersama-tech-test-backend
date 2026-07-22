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

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

type MockAuthService struct {
	mock.Mock
}

func (m *MockAuthService) Register(ctx context.Context, req *domain.RegisterRequest) (*domain.UserResponse, error) {
	args := m.Called(ctx, req)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.UserResponse), args.Error(1)
}

func (m *MockAuthService) Login(ctx context.Context, req *domain.LoginRequest) (*domain.LoginResponse, error) {
	args := m.Called(ctx, req)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.LoginResponse), args.Error(1)
}

func TestRegister_Success(t *testing.T) {
	gin.SetMode(gin.TestMode)
	mockService := new(MockAuthService)
	authHandler := handler.NewAuthHandler(mockService)

	router := gin.Default()
	router.POST("/auth/register", authHandler.Register)

	userID := uuid.New()
	reqBody := domain.RegisterRequest{
		Email:    "test@example.com",
		Password: "password123",
	}

	expectedRes := &domain.UserResponse{
		ID:    userID,
		Email: "test@example.com",
	}

	mockService.On("Register", mock.Anything, &reqBody).Return(expectedRes, nil)

	body, _ := json.Marshal(reqBody)
	req, _ := http.NewRequest(http.MethodPost, "/auth/register", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	resp := httptest.NewRecorder()

	router.ServeHTTP(resp, req)

	assert.Equal(t, http.StatusCreated, resp.Code)
	mockService.AssertExpectations(t)
}

func TestLogin_Success(t *testing.T) {
	gin.SetMode(gin.TestMode)
	mockService := new(MockAuthService)
	authHandler := handler.NewAuthHandler(mockService)

	router := gin.Default()
	router.POST("/auth/login", authHandler.Login)

	reqBody := domain.LoginRequest{
		Email:    "test@example.com",
		Password: "password123",
	}

	expectedRes := &domain.LoginResponse{
		Message: "Login successful",
		Token:   "mock-jwt-token",
	}

	mockService.On("Login", mock.Anything, &reqBody).Return(expectedRes, nil)

	body, _ := json.Marshal(reqBody)
	req, _ := http.NewRequest(http.MethodPost, "/auth/login", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	resp := httptest.NewRecorder()

	router.ServeHTTP(resp, req)

	assert.Equal(t, http.StatusOK, resp.Code)
	mockService.AssertExpectations(t)
}

func TestRegister_InvalidEmail(t *testing.T) {
	gin.SetMode(gin.TestMode)
	mockService := new(MockAuthService)
	authHandler := handler.NewAuthHandler(mockService)

	router := gin.Default()
	router.POST("/auth/register", authHandler.Register)

	reqBody := map[string]string{
		"email":    "not-an-email",
		"password": "123",
	}

	body, _ := json.Marshal(reqBody)
	req, _ := http.NewRequest(http.MethodPost, "/auth/register", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	resp := httptest.NewRecorder()

	router.ServeHTTP(resp, req)

	assert.Equal(t, http.StatusBadRequest, resp.Code)
}

func TestLogin_Failure(t *testing.T) {
	gin.SetMode(gin.TestMode)
	mockService := new(MockAuthService)
	authHandler := handler.NewAuthHandler(mockService)

	router := gin.Default()
	router.POST("/auth/login", authHandler.Login)

	reqBody := domain.LoginRequest{
		Email:    "test@example.com",
		Password: "wrongpassword",
	}

	mockService.On("Login", mock.Anything, &reqBody).Return(nil, domain.ErrInvalidEmailOrPassword)

	body, _ := json.Marshal(reqBody)
	req, _ := http.NewRequest(http.MethodPost, "/auth/login", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	resp := httptest.NewRecorder()

	router.ServeHTTP(resp, req)

	assert.Equal(t, http.StatusUnauthorized, resp.Code)
	mockService.AssertExpectations(t)
}

func TestRegister_EmailAlreadyExists(t *testing.T) {
	gin.SetMode(gin.TestMode)
	mockService := new(MockAuthService)
	authHandler := handler.NewAuthHandler(mockService)

	router := gin.Default()
	router.POST("/auth/register", authHandler.Register)

	reqBody := domain.RegisterRequest{
		Email:    "existing@example.com",
		Password: "password123",
	}

	mockService.On("Register", mock.Anything, &reqBody).Return(nil, domain.ErrEmailAlreadyRegistered)

	body, _ := json.Marshal(reqBody)
	req, _ := http.NewRequest(http.MethodPost, "/auth/register", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	resp := httptest.NewRecorder()

	router.ServeHTTP(resp, req)

	assert.Equal(t, http.StatusConflict, resp.Code)
	mockService.AssertExpectations(t)
}
