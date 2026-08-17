package api

import (
	"encoding/json"
	"net/http"
)

type API struct {
	// dependencies will be injected here
}

func NewAPI() *API {
	return &API{}
}

func (a *API) RegisterRoutes(mux *http.ServeMux) {
	mux.HandleFunc("POST /api/v1/pipelines", a.createPipelineJob)
	mux.HandleFunc("GET /api/v1/pipelines", a.listPipelineJobs)
	mux.HandleFunc("GET /api/v1/pipelines/{id}", a.getPipelineJob)
	mux.HandleFunc("GET /api/v1/pipelines/{id}/progress", a.getPipelineProgress)
	mux.HandleFunc("GET /api/v1/pipelines/{id}/results", a.getPipelineResults)
	mux.HandleFunc("GET /api/v1/pipelines/{id}/errors", a.getPipelineErrors)
	mux.HandleFunc("PATCH /api/v1/pipelines/{id}/cancel", a.cancelPipelineJob)
	mux.HandleFunc("DELETE /api/v1/pipelines/{id}", a.deletePipelineJob)
}

// Mock handlers for Day 1
func (a *API) createPipelineJob(w http.ResponseWriter, r *http.Request) {
	writeJSON(w, http.StatusCreated, map[string]string{"message": "job created", "id": "mock-job-id"})
}

func (a *API) listPipelineJobs(w http.ResponseWriter, r *http.Request) {
	writeJSON(w, http.StatusOK, []string{"mock-job-id"})
}

func (a *API) getPipelineJob(w http.ResponseWriter, r *http.Request) {
	writeJSON(w, http.StatusOK, map[string]string{"id": r.PathValue("id"), "status": "PENDING"})
}

func (a *API) getPipelineProgress(w http.ResponseWriter, r *http.Request) {
	writeJSON(w, http.StatusOK, map[string]interface{}{"job_id": r.PathValue("id"), "percent_complete": 0})
}

func (a *API) getPipelineResults(w http.ResponseWriter, r *http.Request) {
	writeJSON(w, http.StatusOK, map[string]string{"message": "no results yet"})
}

func (a *API) getPipelineErrors(w http.ResponseWriter, r *http.Request) {
	writeJSON(w, http.StatusOK, []string{})
}

func (a *API) cancelPipelineJob(w http.ResponseWriter, r *http.Request) {
	writeJSON(w, http.StatusOK, map[string]string{"message": "job cancelled"})
}

func (a *API) deletePipelineJob(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusNoContent)
}

func writeJSON(w http.ResponseWriter, status int, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(data)
}
