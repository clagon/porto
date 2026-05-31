package server

import (
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"testing/fstest"
)

func TestStatic(t *testing.T) {
	origAssetsFS := assetsFS
	assetsFS = fstest.MapFS{
		"static/index.html": {Data: []byte("<!doctype html><html><head><title>port-mapper</title></head><body></body></html>")},
	}
	t.Cleanup(func() { assetsFS = origAssetsFS })

	tests := []struct {
		name              string
		path              string
		wantStatus        int    // httptest.ResponseRecorder.Code
		wantBodySubstring string // static HTML body substring
	}{
		{name: "root serves index", path: "/", wantStatus: http.StatusOK, wantBodySubstring: "port-mapper"},
		{name: "root injects browser token", path: "/", wantStatus: http.StatusOK, wantBodySubstring: `meta name="porto-browser-token"`},
		{name: "spa fallback serves index", path: "/dashboard", wantStatus: http.StatusOK, wantBodySubstring: "port-mapper"},
	}

	srv := New("127.0.0.1:8080", nil, nil)
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest(http.MethodGet, tt.path, nil)
			rec := httptest.NewRecorder()
			srv.Handler().ServeHTTP(rec, req)
			if rec.Code != tt.wantStatus {
				t.Fatalf("status = %d, want %d", rec.Code, tt.wantStatus)
			}
			body, _ := io.ReadAll(rec.Body)
			if !strings.Contains(string(body), tt.wantBodySubstring) {
				t.Fatalf("body missing %q: %s", tt.wantBodySubstring, string(body))
			}
		})
	}
}
