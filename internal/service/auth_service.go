package service

import (
	"context"
	"errors"
	"fmt"

	"todo-backend/internal/domain"
	"todo-backend/pkg/jwt"

	"golang.org/x/crypto/bcrypt"
)

type authService struct {
	userRepo           domain.UserRepository
	jwtSecret          string
	jwtExpirationHours int
}

func NewAuthService(userRepo domain.UserRepository, jwtSecret string, jwtExpirationHours int) domain.AuthService {
	return &authService{
		userRepo:           userRepo,
		jwtSecret:          jwtSecret,
		jwtExpirationHours: jwtExpirationHours,
	}
}

func (s *authService) Register(ctx context.Context, req *domain.RegisterRequest) (*domain.UserResponse, error) {
	existingUser, err := s.userRepo.GetUserByEmail(ctx, req.Email)
	if err == nil && existingUser != nil {
		return nil, domain.ErrEmailAlreadyRegistered
	}
	if err != nil && !errors.Is(err, domain.ErrUserNotFound) {
		return nil, fmt.Errorf("failed to check existing user: %w", err)
	}

	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		return nil, fmt.Errorf("failed to hash password: %w", err)
	}

	user := &domain.User{
		Email:        req.Email,
		PasswordHash: string(hashedPassword),
	}

	if err := s.userRepo.CreateUser(ctx, user); err != nil {
		return nil, err
	}

	return &domain.UserResponse{
		ID:    user.ID,
		Email: user.Email,
	}, nil
}

func (s *authService) Login(ctx context.Context, req *domain.LoginRequest) (*domain.LoginResponse, error) {
	user, err := s.userRepo.GetUserByEmail(ctx, req.Email)
	if err != nil {
		return nil, domain.ErrInvalidEmailOrPassword
	}

	if err := bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(req.Password)); err != nil {
		return nil, domain.ErrInvalidEmailOrPassword
	}

	token, err := jwt.GenerateToken(user.ID, user.Email, s.jwtSecret, s.jwtExpirationHours)
	if err != nil {
		return nil, fmt.Errorf("failed to generate token: %w", err)
	}

	return &domain.LoginResponse{
		Message: "Login successful",
		Token:   token,
	}, nil
}
