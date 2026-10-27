package storage

import (
	"context"
	"task-manager/internal/dto"
)

type Storage interface {
	CreateTask(context.Context, dto.Task) error
	GetTask(context.Context, int64) (*dto.Task, error)
	UpdateTask(context.Context, dto.Task) error
	DeleteTask(context.Context, int64) error
}
