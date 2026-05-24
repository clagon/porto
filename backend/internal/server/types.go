package server

// HealthResponse is the JSON payload returned by GET /api/health.
type HealthResponse struct {
	Ok bool `json:"ok"`
}
