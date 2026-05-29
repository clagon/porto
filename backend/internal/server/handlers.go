package server

import (
	"net/http"

	"github.com/clagon/port-mapper/backend/internal/config"
	"github.com/clagon/port-mapper/backend/internal/upnp"
	"github.com/labstack/echo/v4"
)

type apiHandlers struct {
	svc ApplicationService
}

func newAPIHandlers(svc ApplicationService) *apiHandlers {
	if svc == nil {
		svc = NewService(ServiceOptions{Config: config.DefaultConfig()})
	}
	return &apiHandlers{svc: svc}
}

func (h *apiHandlers) health(c echo.Context) error {
	return c.JSON(http.StatusOK, HealthResponse{Ok: true})
}

func (h *apiHandlers) status(c echo.Context) error {
	return c.JSON(http.StatusOK, h.svc.Status())
}

func (h *apiHandlers) discover(c echo.Context) error {
	status, err := h.svc.Discover()
	if err != nil {
		return echo.NewHTTPError(http.StatusBadGateway, err.Error())
	}
	return c.JSON(http.StatusAccepted, status)
}

func (h *apiHandlers) portsOpen(c echo.Context) error {
	var req upnp.PortMapping
	if err := c.Bind(&req); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, err.Error())
	}
	status, err := h.svc.OpenPort(req)
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, err.Error())
	}
	return c.JSON(http.StatusAccepted, status)
}

func (h *apiHandlers) portsClose(c echo.Context) error {
	var req upnp.PortMapping
	if err := c.Bind(&req); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, err.Error())
	}
	status, err := h.svc.ClosePort(req)
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, err.Error())
	}
	return c.JSON(http.StatusAccepted, status)
}

func (h *apiHandlers) getSettings(c echo.Context) error {
	return c.JSON(http.StatusOK, h.svc.Settings())
}

func (h *apiHandlers) updateSettings(c echo.Context) error {
	var req config.Config
	if err := c.Bind(&req); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, err.Error())
	}
	if _, err := h.svc.UpdateSettings(req); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, err.Error())
	}
	return c.JSON(http.StatusOK, ActionResponse{Ok: true})
}
