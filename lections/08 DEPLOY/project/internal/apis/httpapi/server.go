package httpapi

import (
	"github.com/labstack/echo/v4"

	"linkshortener/internal/services/linkservice"
)

type handlers struct {
	linkService *linkservice.Service
	baseURL     string
}

func New(linkService *linkservice.Service, baseURL string) *echo.Echo {
	server := echo.New()
	server.HideBanner = true

	linkHandlers := handlers{linkService: linkService, baseURL: baseURL}
	server.POST("/link", linkHandlers.createLink)
	server.GET("/:link", linkHandlers.getLink)

	return server
}
