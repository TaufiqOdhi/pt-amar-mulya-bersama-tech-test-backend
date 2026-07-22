package service

import (
	"context"
	"testing"

	"todo-backend/internal/domain"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
)

type mockUserRepo struct {
	users map[string]*domain.User
}

func newMockUserRepo() *mockUserRepo {
	return &mockUserRepo{users: make(map[string]*domain.User)}
}

func (m *mockUserRepo) CreateUser(ctx context.Context, user *domain.User) error {
	if _, exists := m.users[user.Email]; exists {
		return domain.ErrEmailAlreadyRegistered
	}
	m.users[user.Email] = user
	return nil
}

func (m *mockUserRepo) GetUserByEmail(ctx context.Context, email string) (*domain.User, error) {
	user, exists := m.users[email]
	if !exists {
		return nil, domain.ErrUserNotFound
	}
	return user, nil
}

func (m *mockUserRepo) GetUserByID(ctx context.Context, id uuid.UUID) (*domain.User, error) {
	for _, u := range m.users {
		if u.ID == id {
			return u, nil
		}
	}
	return nil, domain.ErrUserNotFound
}

func TestAuthService_EmailNormalization(t *testing.T) {
	repo := newMockUserRepo()
	svc := NewAuthService(repo, "secret", 24)

	req := &domain.RegisterRequest{
		Email:    "  Test.User@Example.COM  ",
		Password: "password123",
	}

	res, err := svc.Register(context.Background(), req)
	assert.NoError(t, err)
	assert.Equal(t, "test.user@example.com", res.Email)

	// Login with different casing and whitespace
	loginReq := &domain.LoginRequest{
		Email:    "TEST.user@EXAMPLE.com ",
		Password: "password123",
	}

	loginRes, err := svc.Login(context.Background(), loginReq)
	assert.NoError(t, err)
	assert.NotEmpty(t, loginRes.Token)
}
