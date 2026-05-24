package app

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestNew(t *testing.T) {
	tests := []struct {
		name     string
		opts     AppOptions
		wantAddr string
	}{
		{
			name:     "default address",
			opts:     AppOptions{},
			wantAddr: "127.0.0.1:8080",
		},
		{
			name:     "custom address",
			opts:     AppOptions{ListenAddr: "127.0.0.1:9090"},
			wantAddr: "127.0.0.1:9090",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			a, err := New(tt.opts)
			if err != nil {
				t.Fatalf("New() error = %v", err)
			}

			if got := a.Addr(); got != tt.wantAddr {
				t.Fatalf("Addr() = %q, want %q", got, tt.wantAddr)
			}
		})
	}
}

func TestHealthHandler(t *testing.T) {
	a, err := New(AppOptions{})
	if err != nil {
		t.Fatalf("New() error = %v", err)
	}

	req := httptest.NewRequest(http.MethodGet, "/api/health", nil)
	rec := httptest.NewRecorder()

	a.Handler().ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d", rec.Code, http.StatusOK)
	}

	var got map[string]bool
	if err := json.Unmarshal(rec.Body.Bytes(), &got); err != nil {
		t.Fatalf("unmarshal body: %v", err)
	}
	if !got["ok"] {
		t.Fatalf("body = %s, want {\"ok\":true}", rec.Body.String())
	}
}
