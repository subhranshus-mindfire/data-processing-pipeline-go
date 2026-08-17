package api

import (
	"net/http"
	"github.com/user/data-pipeline/internal/store"
)

func (a *API) registerPipelineRoutes(mux *http.ServeMux) {
	mux.HandleFunc("POST /api/v1/pipelines", a.createPipelineJob)
	mux.HandleFunc("GET /api/v1/pipelines", a.listPipelineJobs)
	mux.HandleFunc("GET /api/v1/pipelines/{id}", a.getPipelineJob)
	mux.HandleFunc("GET /api/v1/pipelines/{id}/progress", a.getPipelineProgress)
	mux.HandleFunc("GET /api/v1/pipelines/{id}/results", a.getPipelineResults)
	mux.HandleFunc("GET /api/v1/pipelines/{id}/errors", a.getPipelineErrors)
	mux.HandleFunc("PATCH /api/v1/pipelines/{id}/cancel", a.cancelPipelineJob)
	mux.HandleFunc("DELETE /api/v1/pipelines/{id}", a.deletePipelineJob)
}

func (a *API) createPipelineJob(w http.ResponseWriter, r *http.Request) {
	// For now, we mock the ID or generate one in the service
	job, err := a.pipelineService.CreateJob(r.Context(), "")
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	writeJSON(w, http.StatusCreated, job)
}

func (a *API) listPipelineJobs(w http.ResponseWriter, r *http.Request) {
	jobs, err := a.pipelineService.ListJobs(r.Context())
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, jobs)
}

func (a *API) getPipelineJob(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	job, err := a.pipelineService.GetJob(r.Context(), id)
	if err != nil {
		if err == store.ErrJobNotFound {
			writeError(w, http.StatusNotFound, "job not found")
			return
		}
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, job)
}

func (a *API) getPipelineProgress(w http.ResponseWriter, r *http.Request) {
	// Placeholder: In a real app we'd fetch metrics from a metrics service/store
	writeJSON(w, http.StatusOK, map[string]interface{}{"job_id": r.PathValue("id"), "percent_complete": 0})
}

func (a *API) getPipelineResults(w http.ResponseWriter, r *http.Request) {
	writeJSON(w, http.StatusOK, map[string]string{"message": "no results yet"})
}

func (a *API) getPipelineErrors(w http.ResponseWriter, r *http.Request) {
	writeJSON(w, http.StatusOK, []string{})
}

func (a *API) cancelPipelineJob(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	job, err := a.pipelineService.CancelJob(r.Context(), id)
	if err != nil {
		if err == store.ErrJobNotFound {
			writeError(w, http.StatusNotFound, "job not found")
			return
		}
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, job)
}

func (a *API) deletePipelineJob(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	err := a.pipelineService.DeleteJob(r.Context(), id)
	if err != nil {
		if err == store.ErrJobNotFound {
			writeError(w, http.StatusNotFound, "job not found")
			return
		}
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	w.WriteHeader(http.StatusNoContent)
}
