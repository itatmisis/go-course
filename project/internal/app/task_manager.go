package app

import (
	"task-manager/internal/presentations/task_presentation"

	"github.com/labstack/echo/v4"
)

func BuildApp(taskPresentation *task_presentation.Presentation) *echo.Echo {
	e := echo.New()

	g := e.Group("/api")

	g.GET("/task/:id", taskPresentation.GetTask)
	g.POST("/task", taskPresentation.CreateTask)
	g.PUT("/task/:id", taskPresentation.UpdateTask)
	g.DELETE("/task/:id", taskPresentation.DeleteTask)

	return e
}
