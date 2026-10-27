package memory_storage

import (
	"context"
	"task-manager/internal/dto"
)

type Storage struct {
	taskStorage map[int64]dto.Task
}

func NewStorage() *Storage {
	return &Storage{taskStorage: make(map[int64]dto.Task)}
}

func (s *Storage) CreateTask(_ context.Context, task dto.Task) error {
	s.taskStorage[task.ID] = task
	return nil
}

func (s *Storage) GetTask(_ context.Context, id int64) (*dto.Task, error) {
	return new(s.taskStorage[id]), nil
}

func (s *Storage) UpdateTask(_ context.Context, task dto.Task) error {
	s.taskStorage[task.ID] = task
	return nil
}

func (s *Storage) DeleteTask(_ context.Context, id int64) error {
	delete(s.taskStorage, id)
	return nil
}
