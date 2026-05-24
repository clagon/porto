package server

import (
	"log/slog"
	"net"
	"net/http"
	"time"
)

// Server wraps the HTTP handler used by the application.
type Server struct {
	addr    string
	handler http.Handler
	logger  *slog.Logger
}

// New constructs a server bound to the provided listen address.
func New(addr string, logger *slog.Logger) *Server {
	s := &Server{addr: addr, logger: logger}
	s.handler = NewMux()
	return s
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
	if s == nil {
		return http.NewServeMux()
	}
	if s.handler == nil {
		s.handler = NewMux()
	}
	return loggingMiddleware(s.handler, s.logger)
}

// ListenAndServe runs the server on its configured address.
func (s *Server) ListenAndServe() error {
	ln, err := net.Listen("tcp", s.addr)
	if err != nil {
		return err
	}
	return s.Serve(ln)
}

// Serve runs the server on the provided listener.
func (s *Server) Serve(ln net.Listener) error {
	return http.Serve(ln, s.Handler())
}

type loggingResponseWriter struct {
	http.ResponseWriter
	status int
	bytes  int
}

func (w *loggingResponseWriter) WriteHeader(status int) {
	w.status = status
	w.ResponseWriter.WriteHeader(status)
}

func (w *loggingResponseWriter) Write(p []byte) (int, error) {
	if w.status == 0 {
		w.status = http.StatusOK
	}
	n, err := w.ResponseWriter.Write(p)
	w.bytes += n
	return n, err
}

func loggingMiddleware(next http.Handler, logger *slog.Logger) http.Handler {
	if logger == nil {
		logger = slog.Default()
	}
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		started := time.Now()
		rw := &loggingResponseWriter{ResponseWriter: w}
		next.ServeHTTP(rw, r)
		status := rw.status
		if status == 0 {
			status = http.StatusOK
		}
		logger.Info("request",
			"method", r.Method,
			"path", r.URL.Path,
			"status", status,
			"bytes", rw.bytes,
			"duration", time.Since(started).String(),
		)
	})
}
