package api

import (
	"encoding/json"
	"net/http"

	"github.com/user/data-pipeline/internal/service"
)

// API holds the dependencies for the HTTP handlers
type API struct {
	pipelineService service.PipelineService
}

// NewAPI creates a new API instance with the given service dependencies
func NewAPI(pipelineService service.PipelineService) *API {
	return &API{
		pipelineService: pipelineService,
	}
}

// RegisterRoutes sets up all the HTTP routes and applies global middleware
func (a *API) RegisterRoutes(mux *http.ServeMux) {
	// Handlers are defined in pipeline_handler.go
	mux.HandleFunc("POST /api/v1/pipelines", a.createPipelineJob)
	mux.HandleFunc("GET /api/v1/pipelines", a.listPipelineJobs)
	mux.HandleFunc("GET /api/v1/pipelines/{id}", a.getPipelineJob)
	mux.HandleFunc("GET /api/v1/pipelines/{id}/progress", a.getPipelineProgress)
	mux.HandleFunc("GET /api/v1/pipelines/{id}/results", a.getPipelineResults)
	mux.HandleFunc("GET /api/v1/pipelines/{id}/errors", a.getPipelineErrors)
	mux.HandleFunc("PATCH /api/v1/pipelines/{id}/cancel", a.cancelPipelineJob)
	mux.HandleFunc("DELETE /api/v1/pipelines/{id}", a.deletePipelineJob)
}

// writeJSON is a helper to write JSON responses
func writeJSON(w http.ResponseWriter, status int, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	if data != nil {
		json.NewEncoder(w).Encode(data)
	}
}

// writeError is a helper to write error responses
func writeError(w http.ResponseWriter, status int, message string) {
	writeJSON(w, status, map[string]string{"error": message})
}
