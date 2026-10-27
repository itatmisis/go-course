package task_service

import (
	"context"
	"fmt"
	"task-manager/internal/dto"
	"task-manager/internal/storage/postgres_storage"
)

type TaskService struct {
	storage *postgres_storage.Storage
}

func NewService(storage *postgres_storage.Storage) *TaskService {
	return &TaskService{storage: storage}
}

func (ts *TaskService) CreateTask(ctx context.Context, taskReq dto.TaskRequest) error {
	task := convertTask(taskReq)
	if err := ts.storage.CreateTask(ctx, task); err != nil {
		return fmt.Errorf("ts.storage.CreateTask: %w", err)
	}

	return nil
}

func (ts *TaskService) GetTask(ctx context.Context, id int64) (*dto.Task, error) {
	task, err := ts.storage.GetTask(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("ts.storage.GetTask: %w", err)
	}

	return task, nil
}

func (ts *TaskService) UpdateTask(ctx context.Context, taskReq dto.TaskRequest, id int64) error {
	task := convertTask(taskReq)
	task.ID = id
	if err := ts.storage.UpdateTask(ctx, task); err != nil {
		return fmt.Errorf("ts.storage.UpdateTask: %w", err)
	}

	return nil
}

func (ts *TaskService) DeleteTask(ctx context.Context, id int64) error {
	if err := ts.storage.DeleteTask(ctx, id); err != nil {
		return fmt.Errorf("ts.storage.DeleteTask: %w", err)
	}

	return nil
}

func convertTask(taskReq dto.TaskRequest) dto.Task {
	return dto.Task{
		Title:       taskReq.Title,
		Description: taskReq.Description,
		DueDate:     taskReq.DueDate,
		Priority:    toPriority(taskReq.Priority),
		Status:      toStatus(taskReq.Status),
	}
}

func toPriority(priority int) dto.Priority {
	switch priority {
	case 0:
		return dto.PriorityLow
	case 1:
		return dto.PriorityMedium
	case 2:
		return dto.PriorityHigh
	default:
		return dto.PriorityNotSpecified
	}
}

func toStatus(status int) dto.Status {
	switch status {
	case 0:
		return dto.StatusTodo
	case 1:
		return dto.StatusWip
	case 2:
		return dto.StatusDone
	default:
		return dto.StatusNotSpecified
	}
}
