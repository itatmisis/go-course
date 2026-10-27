package httpapi

import (
	"errors"
	"fmt"
	"net/http"
	"net/url"

	"github.com/labstack/echo/v4"

	"linkshortener/internal/services/linkservice"
)

type linkBody struct {
	Link string `json:"link"`
}

func (h *handlers) createLink(c echo.Context) error {
	var request linkBody

	err := c.Bind(&request)
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid body")
	}

	if request.Link == "" {
		return echo.NewHTTPError(http.StatusBadRequest, "link is required")
	}

	shortLink, err := h.linkService.CreateLink(c.Request().Context(), request.Link)
	if err != nil {
		return fmt.Errorf("create link: %w", err)
	}

	fullShortLink, err := url.JoinPath(h.baseURL, shortLink)
	if err != nil {
		return fmt.Errorf("build short link: %w", err)
	}

	return c.JSON(http.StatusOK, linkBody{Link: fullShortLink})
}

func (h *handlers) getLink(c echo.Context) error {
	realLink, err := h.linkService.GetRealLink(c.Request().Context(), c.Param("link"))
	if errors.Is(err, linkservice.ErrLinkNotFound) {
		return echo.NewHTTPError(http.StatusNotFound, "Short link not found:(")
	}

	if err != nil {
		return fmt.Errorf("get real link: %w", err)
	}

	return c.Redirect(http.StatusMovedPermanently, realLink)
}
