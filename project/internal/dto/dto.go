package dto

import (
	"time"
)

type Status = string

const (
	StatusTodo         Status = "todo"
	StatusWip          Status = "wip"
	StatusDone         Status = "done"
	StatusNotSpecified Status = "not_specified"
)

type Priority = string

const (
	PriorityLow          Priority = "low"
	PriorityMedium       Priority = "medium"
	PriorityHigh         Priority = "high"
	PriorityNotSpecified Priority = "not_specified"
)

type Task struct {
	ID          int64
	Title       string
	Description string
	DueDate     *time.Time
	Priority    Priority
	Status      Status
	AssigneeID  *int64

	CreatedAt *time.Time
	UpdatedAt *time.Time
}

type User struct {
	ID    int64
	Login string

	CreatedAt *time.Time
	UpdatedAt *time.Time
}

type TaskRequest struct {
	Title       string     `json:"title"`
	Description string     `json:"description"`
	DueDate     *time.Time `json:"due_date,omitempty"`
	Priority    int        `json:"priority"`
	Status      int        `json:"status"`
	AssigneeID  *int64     `json:"assignee_id,omitempty"`
}
