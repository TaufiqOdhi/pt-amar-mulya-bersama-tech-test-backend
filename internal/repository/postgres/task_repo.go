package postgres

import (
	"context"
	"errors"
	"fmt"

	"todo-backend/internal/domain"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"golang.org/x/sync/errgroup"
)

type taskRepo struct {
	db *pgxpool.Pool
}

func NewTaskRepo(db *pgxpool.Pool) domain.TaskRepository {
	return &taskRepo{db: db}
}

func (r *taskRepo) CreateTask(ctx context.Context, task *domain.Task) error {
	query := `
		INSERT INTO tasks (id, user_id, title, description, status, due_date, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, NOW(), NOW())
		RETURNING created_at, updated_at
	`
	if task.ID == uuid.Nil {
		task.ID = uuid.New()
	}

	err := r.db.QueryRow(ctx, query,
		task.ID, task.UserID, task.Title, task.Description, task.Status, task.DueDate,
	).Scan(&task.CreatedAt, &task.UpdatedAt)

	if err != nil {
		return fmt.Errorf("failed to create task: %w", err)
	}
	return nil
}

func (r *taskRepo) GetTaskByID(ctx context.Context, id uuid.UUID, userID uuid.UUID) (*domain.Task, error) {
	query := `
		SELECT id, user_id, title, COALESCE(description, ''), status, TO_CHAR(due_date, 'YYYY-MM-DD'), created_at, updated_at, deleted_at
		FROM tasks
		WHERE id = $1 AND user_id = $2 AND deleted_at IS NULL
	`
	task := &domain.Task{}
	err := r.db.QueryRow(ctx, query, id, userID).Scan(
		&task.ID, &task.UserID, &task.Title, &task.Description, &task.Status, &task.DueDate, &task.CreatedAt, &task.UpdatedAt, &task.DeletedAt,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, errors.New("task not found")
		}
		return nil, fmt.Errorf("failed to get task by id: %w", err)
	}
	return task, nil
}

func (r *taskRepo) GetTasks(ctx context.Context, userID uuid.UUID, params domain.TaskQueryParams) ([]*domain.Task, int, error) {
	if params.Page <= 0 {
		params.Page = 1
	}
	if params.Limit <= 0 {
		params.Limit = 10
	}
	offset := (params.Page - 1) * params.Limit

	baseQuery := "FROM tasks WHERE user_id = $1 AND deleted_at IS NULL"
	args := []interface{}{userID}
	argID := 2

	if params.Status != "" {
		baseQuery += fmt.Sprintf(" AND status = $%d", argID)
		args = append(args, params.Status)
		argID++
	}

	if params.Search != "" {
		baseQuery += fmt.Sprintf(" AND (title ILIKE $%d OR description ILIKE $%d)", argID, argID)
		args = append(args, "%"+params.Search+"%")
		argID++
	}

	var tasks []*domain.Task
	var totalCount int

	g, gCtx := errgroup.WithContext(ctx)

	// Goroutine 1: Fetch tasks list concurrently
	g.Go(func() error {
		selectQuery := fmt.Sprintf(`
			SELECT id, title, COALESCE(description, ''), status, TO_CHAR(due_date, 'YYYY-MM-DD')
			%s
			ORDER BY created_at DESC
			LIMIT $%d OFFSET $%d
		`, baseQuery, argID, argID+1)

		queryArgs := append([]interface{}{}, args...)
		queryArgs = append(queryArgs, params.Limit, offset)

		rows, err := r.db.Query(gCtx, selectQuery, queryArgs...)
		if err != nil {
			return fmt.Errorf("failed to query tasks: %w", err)
		}
		defer rows.Close()

		var list []*domain.Task
		for rows.Next() {
			t := &domain.Task{}
			if err := rows.Scan(&t.ID, &t.Title, &t.Description, &t.Status, &t.DueDate); err != nil {
				return fmt.Errorf("failed to scan task row: %w", err)
			}
			list = append(list, t)
		}
		if err := rows.Err(); err != nil {
			return fmt.Errorf("row iteration error: %w", err)
		}

		if list == nil {
			list = []*domain.Task{}
		}
		tasks = list
		return nil
	})

	// Goroutine 2: Count total tasks concurrently
	g.Go(func() error {
		countQuery := fmt.Sprintf("SELECT COUNT(*) %s", baseQuery)
		err := r.db.QueryRow(gCtx, countQuery, args...).Scan(&totalCount)
		if err != nil {
			return fmt.Errorf("failed to count tasks: %w", err)
		}
		return nil
	})

	// Wait for both goroutines to finish concurrently
	if err := g.Wait(); err != nil {
		return nil, 0, err
	}

	return tasks, totalCount, nil
}

func (r *taskRepo) UpdateTask(ctx context.Context, task *domain.Task) error {
	query := `
		UPDATE tasks
		SET title = $1, description = $2, status = $3, due_date = $4, updated_at = NOW()
		WHERE id = $5 AND user_id = $6 AND deleted_at IS NULL
		RETURNING updated_at
	`
	err := r.db.QueryRow(ctx, query,
		task.Title, task.Description, task.Status, task.DueDate, task.ID, task.UserID,
	).Scan(&task.UpdatedAt)

	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return errors.New("task not found")
		}
		return fmt.Errorf("failed to update task: %w", err)
	}
	return nil
}

func (r *taskRepo) DeleteTask(ctx context.Context, id uuid.UUID, userID uuid.UUID) error {
	query := `
		UPDATE tasks
		SET deleted_at = NOW(), updated_at = NOW()
		WHERE id = $1 AND user_id = $2 AND deleted_at IS NULL
	`
	cmdTag, err := r.db.Exec(ctx, query, id, userID)
	if err != nil {
		return fmt.Errorf("failed to delete task: %w", err)
	}
	if cmdTag.RowsAffected() == 0 {
		return errors.New("task not found")
	}
	return nil
}
