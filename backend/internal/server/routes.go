package server

import (
	"encoding/json"
	"net/http"
)

// NewMux builds the HTTP routes for the application.
func NewMux() http.Handler {
	mux := http.NewServeMux()
	mux.HandleFunc("/api/health", healthHandler)
	return mux
}

func healthHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		w.Header().Set("Allow", http.MethodGet)
		http.Error(w, http.StatusText(http.StatusMethodNotAllowed), http.StatusMethodNotAllowed)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	_ = json.NewEncoder(w).Encode(HealthResponse{Ok: true})
}
