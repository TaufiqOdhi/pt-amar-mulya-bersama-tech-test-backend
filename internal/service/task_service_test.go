package service

import (
	"context"
	"testing"

	"todo-backend/internal/domain"

	"github.com/google/uuid"
)

type mockTaskRepo struct {
	tasks      []*domain.Task
	totalCount int
	err        error
}

func (m *mockTaskRepo) CreateTask(ctx context.Context, task *domain.Task) error {
	return m.err
}

func (m *mockTaskRepo) GetTaskByID(ctx context.Context, id uuid.UUID, userID uuid.UUID) (*domain.Task, error) {
	if len(m.tasks) > 0 {
		return m.tasks[0], m.err
	}
	return nil, m.err
}

func (m *mockTaskRepo) GetTasks(ctx context.Context, userID uuid.UUID, params domain.TaskQueryParams) ([]*domain.Task, int, error) {
	return m.tasks, m.totalCount, m.err
}

func (m *mockTaskRepo) UpdateTask(ctx context.Context, task *domain.Task) error {
	return m.err
}

func (m *mockTaskRepo) DeleteTask(ctx context.Context, id uuid.UUID, userID uuid.UUID) error {
	return m.err
}

func TestGetTasks_Pagination(t *testing.T) {
	userID := uuid.New()

	tests := []struct {
		name               string
		totalCount         int
		limit              int
		expectedTotalPages int
	}{
		{
			name:               "Zero tasks should result in 0 total pages",
			totalCount:         0,
			limit:              10,
			expectedTotalPages: 0,
		},
		{
			name:               "5 tasks with limit 10 should result in 1 total page",
			totalCount:         5,
			limit:              10,
			expectedTotalPages: 1,
		},
		{
			name:               "25 tasks with limit 10 should result in 3 total pages",
			totalCount:         25,
			limit:              10,
			expectedTotalPages: 3,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo := &mockTaskRepo{
				tasks:      []*domain.Task{},
				totalCount: tt.totalCount,
			}
			svc := NewTaskService(repo, nil)

			res, err := svc.GetTasks(context.Background(), userID, domain.TaskQueryParams{
				Page:  1,
				Limit: tt.limit,
			})

			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}

			if res.Pagination.TotalPages != tt.expectedTotalPages {
				t.Errorf("expected total pages %d, got %d", tt.expectedTotalPages, res.Pagination.TotalPages)
			}
			if res.Pagination.TotalTasks != tt.totalCount {
				t.Errorf("expected total tasks %d, got %d", tt.totalCount, res.Pagination.TotalTasks)
			}
		})
	}
}
