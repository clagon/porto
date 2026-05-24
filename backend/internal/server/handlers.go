package server

import (
	"net/http"

	"github.com/clagon/port-mapper/backend/internal/config"
	"github.com/labstack/echo/v4"
)

type apiHandlers struct {
	cfg config.Config
}

func newAPIHandlers() *apiHandlers {
	return &apiHandlers{cfg: config.DefaultConfig()}
}

func (h *apiHandlers) health(c echo.Context) error {
	return c.JSON(http.StatusOK, map[string]bool{"ok": true})
}

func (h *apiHandlers) status(c echo.Context) error {
	return c.JSON(http.StatusOK, map[string]any{
		"discovered": false,
		"ports":      []any{},
	})
}

func (h *apiHandlers) discover(c echo.Context) error {
	return c.JSON(http.StatusAccepted, map[string]bool{"ok": true})
}

func (h *apiHandlers) portsOpen(c echo.Context) error {
	return c.JSON(http.StatusAccepted, map[string]bool{"ok": true})
}

func (h *apiHandlers) portsClose(c echo.Context) error {
	return c.JSON(http.StatusAccepted, map[string]bool{"ok": true})
}

func (h *apiHandlers) getSettings(c echo.Context) error {
	return c.JSON(http.StatusOK, h.cfg.WithDefaults())
}

func (h *apiHandlers) updateSettings(c echo.Context) error {
	return c.JSON(http.StatusOK, map[string]bool{"ok": true})
}
