package app

import (
	"net/http"

	"github.com/clagon/port-mapper/backend/internal/config"
	"github.com/clagon/port-mapper/backend/internal/server"
)

const defaultListenAddr = "127.0.0.1:8080"

// AppOptions configures a new App.
type AppOptions struct {
	ListenAddr string
}

// App is the top-level application container.
type App struct {
	cfg    config.Config
	server *server.Server
}

// New constructs a new App using the provided options.
func New(opts AppOptions) (*App, error) {
	cfg := config.Config{ListenAddr: opts.ListenAddr}.WithDefaults()
	if cfg.ListenAddr == "" {
		cfg.ListenAddr = defaultListenAddr
	}

	return &App{
		cfg:    cfg,
		server: server.New(cfg.ListenAddr),
	}, nil
}

// Addr returns the configured listen address.
func (a *App) Addr() string {
	if a == nil || a.server == nil {
		return ""
	}
	return a.server.Addr()
}

// Handler returns the application's HTTP handler.
func (a *App) Handler() http.Handler {
	if a == nil || a.server == nil {
		return http.NewServeMux()
	}
	return a.server.Handler()
}

// Run starts the HTTP server.
func (a *App) Run() error {
	if a == nil || a.server == nil {
		return nil
	}
	return a.server.ListenAndServe()
}
