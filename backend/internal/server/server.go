package server

import (
	"log/slog"
	"net"
	"net/http"
	"time"

	"github.com/labstack/echo/v4"
)

// Server wraps the HTTP handler used by the application.
type Server struct {
	addr   string
	echo   *echo.Echo
	logger *slog.Logger
	svc    ApplicationService
}

// New constructs a server bound to the provided listen address.
func New(addr string, logger *slog.Logger, svc ApplicationService) *Server {
	if logger == nil {
		logger = slog.Default()
	}
	if svc == nil {
		svc = NewService(ServiceOptions{Logger: logger})
	}

	e := echo.New()
	e.HideBanner = true
	e.HidePort = true
	e.Use(loggingMiddleware(logger))
	registerRoutes(e, svc)

	return &Server{addr: addr, echo: e, logger: logger, svc: svc}
}

// Addr returns the configured listen address.
func (s *Server) Addr() string {
	if s == nil {
		return ""
	}
	return s.addr
}

// Handler returns the server's HTTP handler.
func (s *Server) Handler() http.Handler {
	if s == nil || s.echo == nil {
		return http.NewServeMux()
	}
	return s.echo
}

// Discover runs router discovery and updates the in-memory state.
func (s *Server) Discover() error {
	if s == nil || s.svc == nil {
		return nil
	}
	_, err := s.svc.Discover()
	return err
}

// ListenAndServe runs the server on its configured address.
func (s *Server) ListenAndServe() error {
	if s == nil || s.echo == nil {
		return nil
	}
	return s.echo.Start(s.addr)
}

// Serve runs the server on the provided listener.
func (s *Server) Serve(ln net.Listener) error {
	if s == nil || s.echo == nil {
		return nil
	}
	return (&http.Server{Handler: s.echo, ReadHeaderTimeout: 5 * time.Second}).Serve(ln)
}

func loggingMiddleware(logger *slog.Logger) echo.MiddlewareFunc {
	if logger == nil {
		logger = slog.Default()
	}
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			started := time.Now()
			err := next(c)
			status := c.Response().Status
			if status == 0 {
				status = http.StatusOK
			}
			if err != nil && status < http.StatusBadRequest {
				status = http.StatusInternalServerError
			}
			logger.Info("request",
				"method", c.Request().Method,
				"path", c.Request().URL.Path,
				"status", status,
				"bytes", c.Response().Size,
				"duration", time.Since(started).String(),
			)
			return err
		}
	}
}
