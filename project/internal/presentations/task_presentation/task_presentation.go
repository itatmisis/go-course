package task_presentation

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"strconv"
	"task-manager/internal/dto"
	"task-manager/internal/services/task_service"

	"github.com/jackc/pgx/v5"
	"github.com/labstack/echo/v4"
)

type Presentation struct {
	service *task_service.TaskService
}

func NewPresentation(service *task_service.TaskService) *Presentation {
	return &Presentation{service: service}
}

func (p *Presentation) CreateTask(c echo.Context) error {
	var task dto.TaskRequest

	if err := c.Bind(&task); err != nil {
		return c.JSON(http.StatusUnprocessableEntity, map[string]string{"err": fmt.Errorf("c.Bind: %w", err).Error()})
	}

	if err := p.service.CreateTask(context.Background(), task); err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"err": fmt.Errorf("p.service.CreateTask: %w", err).Error()})
	}

	return c.JSON(http.StatusOK, nil)
}

func (p *Presentation) GetTask(c echo.Context) error {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		return c.JSON(http.StatusUnprocessableEntity, map[string]string{"err": fmt.Errorf("invalid task id: err: %w", err).Error()})
	}

	task, err := p.service.GetTask(context.Background(), int64(id))
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return c.JSON(http.StatusNotFound, map[string]string{"err": fmt.Errorf("p.service.GetTask: %w", err).Error()})
		}
		return c.JSON(http.StatusInternalServerError, map[string]string{"err": fmt.Errorf("p.service.GetTask: %w", err).Error()})
	}

	taskJson, err := json.Marshal(&task)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"err": fmt.Errorf("json.Marshal: %w", err).Error()})
	}

	return c.JSON(http.StatusOK, map[string]string{"task": string(taskJson)})
}

func (p *Presentation) UpdateTask(c echo.Context) error {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		return c.JSON(http.StatusUnprocessableEntity, map[string]string{"err": fmt.Errorf("invalid task id: err: %w", err).Error()})
	}

	var task dto.TaskRequest

	if err := c.Bind(&task); err != nil {
		return c.JSON(http.StatusUnprocessableEntity, map[string]string{"err": fmt.Errorf("c.Bind: %w", err).Error()})
	}

	if err := p.service.UpdateTask(context.Background(), task, int64(id)); err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"err": fmt.Errorf("p.service.UpdateTask: %w", err).Error()})
	}

	return c.JSON(http.StatusOK, nil)
}

func (p *Presentation) DeleteTask(c echo.Context) error {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		return c.JSON(http.StatusUnprocessableEntity, map[string]string{"err": fmt.Errorf("invalid task id: err: %w", err).Error()})
	}

	if err := p.service.DeleteTask(context.Background(), int64(id)); err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"err": fmt.Errorf("p.service.DeleteTask: %w", err).Error()})
	}

	return c.JSON(http.StatusOK, nil)
}
