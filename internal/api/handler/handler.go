package handler

import (
	"encoding/json"
	"net/http"

	"github.com/user/data-pipeline/internal/config"
	"github.com/user/data-pipeline/internal/service"
)

// API holds the dependencies for the HTTP handlers
type API struct {
	pipelineService service.PipelineService
	cfg             *config.Config
}

// NewAPI creates a new API instance with the given service dependencies
func NewAPI(pipelineService service.PipelineService, cfg *config.Config) *API {
	return &API{
		pipelineService: pipelineService,
		cfg:             cfg,
	}
}

// writeJSON is a helper to write JSON responses
func writeJSON(w http.ResponseWriter, status int, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	if data != nil {
		_ = json.NewEncoder(w).Encode(data)
	}
}

// ErrorResponse represents a standardized API error
type ErrorResponse struct {
	Error string `json:"error"`
}

// writeError is a helper to write error responses
func writeError(w http.ResponseWriter, status int, message string) {
	writeJSON(w, status, ErrorResponse{Error: message})
}
