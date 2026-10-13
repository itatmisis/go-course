package main

import (
	"net/http"

	"github.com/labstack/echo/v4"
)

func buildApp() *echo.Echo {
	e := echo.New()
	e.GET("/hello", func(c echo.Context) error {
		return c.JSON(http.StatusOK, map[string]string{"message": "Hello, Go!"})
	})

	return e
}
