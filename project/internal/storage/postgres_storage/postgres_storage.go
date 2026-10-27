package postgres_storage

import (
	"context"
	"fmt"
	"task-manager/internal/dto"

	"github.com/jackc/pgx/v5/pgxpool"
)

type Storage struct {
	pool *pgxpool.Pool
}

func NewStorage(pool *pgxpool.Pool) *Storage {
	return &Storage{pool: pool}
}

func (s *Storage) CreateTask(ctx context.Context, task dto.Task) error {
	query := `
INSERT INTO tasks (title, description, status, due_date, priority) VALUES ($1, $2, $3, $4, $5)
`
	if _, err := s.pool.Exec(ctx, query, task.Title, task.Description, task.Status, task.DueDate, task.Priority); err != nil {
		return fmt.Errorf("s.pool.Exec: %w", err)
	}

	return nil
}

func (s *Storage) GetTask(ctx context.Context, id int64) (*dto.Task, error) {
	var task dto.Task

	query := `
SELECT id, title, description, status, assignee_id, due_date, priority, created_at, updated_at
FROM tasks WHERE id = $1
`
	if err := s.pool.QueryRow(ctx, query, id).Scan(
		&task.ID, &task.Title, &task.Description, &task.Status, &task.AssigneeID,
		&task.DueDate, &task.Priority, &task.CreatedAt, &task.UpdatedAt); err != nil {
		return nil, fmt.Errorf("s.pool.QueryRow: %w", err)
	}

	return &task, nil
}

func (s *Storage) UpdateTask(ctx context.Context, task dto.Task) error {
	query := `
UPDATE tasks SET title = $1, description = $2, status = $3, assignee_id = $4, due_date = $5, priority = $6
WHERE id = $7
`
	if _, err := s.pool.Exec(ctx, query, task.Title, task.Description, task.Status,
		task.AssigneeID, task.DueDate, task.Priority, task.ID); err != nil {
		return fmt.Errorf("s.pool.Exec: %w", err)
	}

	return nil
}

func (s *Storage) DeleteTask(ctx context.Context, id int64) error {
	query := `
DELETE FROM tasks WHERE id = $1 
`
	if _, err := s.pool.Exec(ctx, query, id); err != nil {
		return fmt.Errorf("s.pool.Exec: %w", err)
	}

	return nil
}
