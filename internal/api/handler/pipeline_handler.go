package handler

import (
	"encoding/json"
	"net/http"

	"github.com/user/data-pipeline/internal/domain"
	"github.com/user/data-pipeline/internal/pipeline"
	"github.com/user/data-pipeline/internal/repository"
)

func (a *API) RegisterPipelineRoutes(mux *http.ServeMux) {
	mux.HandleFunc("POST /api/v1/pipelines", a.createPipelineJob)
	mux.HandleFunc("GET /api/v1/pipelines", a.listPipelineJobs)
	mux.HandleFunc("GET /api/v1/pipelines/{id}", a.getPipelineJob)
	mux.HandleFunc("GET /api/v1/pipelines/{id}/progress", a.getPipelineProgress)
	mux.HandleFunc("GET /api/v1/pipelines/{id}/results", a.getPipelineResults)
	mux.HandleFunc("GET /api/v1/pipelines/{id}/errors", a.getPipelineErrors)
	mux.HandleFunc("PATCH /api/v1/pipelines/{id}/cancel", a.cancelPipelineJob)
	mux.HandleFunc("DELETE /api/v1/pipelines/{id}", a.deletePipelineJob)
}

// @Summary Create a pipeline job
// @Description Starts a new pipeline execution in the background based on the provided sources
// @Tags pipeline
// @Accept json
// @Produce json
// @Param spec body domain.JobSpec true "Job Specification"
// @Success 201 {object} domain.Job
// @Failure 400 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Router /api/v1/pipelines [post]
func (a *API) createPipelineJob(w http.ResponseWriter, r *http.Request) {
	var spec domain.JobSpec
	if err := json.NewDecoder(r.Body).Decode(&spec); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	job, err := a.pipelineService.CreateJob(r.Context(), spec)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	writeJSON(w, http.StatusCreated, job)
}

// @Summary List pipeline jobs
// @Description Retrieves a list of all pipeline jobs
// @Tags pipeline
// @Produce json
// @Success 200 {array} domain.Job
// @Failure 500 {object} ErrorResponse
// @Router /api/v1/pipelines [get]
func (a *API) listPipelineJobs(w http.ResponseWriter, r *http.Request) {
	jobs, err := a.pipelineService.ListJobs(r.Context())
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, jobs)
}

// @Summary Get a pipeline job
// @Description Retrieves the details of a specific pipeline job by ID
// @Tags pipeline
// @Produce json
// @Param id path string true "Job ID"
// @Success 200 {object} domain.Job
// @Failure 404 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Router /api/v1/pipelines/{id} [get]
func (a *API) getPipelineJob(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	job, err := a.pipelineService.GetJob(r.Context(), id)
	if err != nil {
		if err == repository.ErrJobNotFound {
			writeError(w, http.StatusNotFound, "job not found")
			return
		}
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, job)
}

// @Summary Get job progress
// @Description Fetches real-time metrics and progress of a running or completed job
// @Tags pipeline
// @Produce json
// @Param id path string true "Job ID"
// @Success 200 {object} domain.Metrics
// @Failure 404 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Router /api/v1/pipelines/{id}/progress [get]
func (a *API) getPipelineProgress(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	job, err := a.pipelineService.GetJob(r.Context(), id)
	if err != nil {
		if err == repository.ErrJobNotFound {
			writeError(w, http.StatusNotFound, "job not found")
			return
		}
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}

	if job.Metrics == nil {
		writeJSON(w, http.StatusOK, map[string]interface{}{"job_id": id, "percent_complete": 0})
		return
	}

	writeJSON(w, http.StatusOK, job.Metrics)
}

// @Summary Get job results
// @Description Retrieves the final aggregated summary once the job reaches COMPLETED status
// @Tags pipeline
// @Produce json
// @Param id path string true "Job ID"
// @Success 200 {object} pipeline.SummaryRecord
// @Failure 404 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Router /api/v1/pipelines/{id}/results [get]
func (a *API) getPipelineResults(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	
	payload, err := a.pipelineService.GetJobResult(r.Context(), id)
	if err != nil {
		if err == repository.ErrResultNotFound {
			writeError(w, http.StatusNotFound, "results not ready or job not found")
			return
		}
		writeError(w, http.StatusInternalServerError, "failed to fetch results")
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	w.Write(payload)
}

// @Summary Get job errors
// @Description Retrieves the errors associated with a job (Stub)
// @Tags pipeline
// @Produce json
// @Param id path string true "Job ID"
// @Success 200 {array} string
// @Router /api/v1/pipelines/{id}/errors [get]
func (a *API) getPipelineErrors(w http.ResponseWriter, r *http.Request) {
	writeJSON(w, http.StatusOK, []string{})
}

// @Summary Cancel a pipeline job
// @Description Gracefully aborts all running goroutines for a specific job
// @Tags pipeline
// @Produce json
// @Param id path string true "Job ID"
// @Success 200 {object} domain.Job
// @Failure 400 {object} ErrorResponse
// @Failure 404 {object} ErrorResponse
// @Router /api/v1/pipelines/{id}/cancel [patch]
func (a *API) cancelPipelineJob(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	job, err := a.pipelineService.CancelJob(r.Context(), id)
	if err != nil {
		if err == repository.ErrJobNotFound {
			writeError(w, http.StatusNotFound, "job not found")
			return
		}
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, job)
}

// @Summary Delete a pipeline job
// @Description Deletes a specific pipeline job by ID
// @Tags pipeline
// @Produce json
// @Param id path string true "Job ID"
// @Success 204 "No Content"
// @Failure 404 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Router /api/v1/pipelines/{id} [delete]
func (a *API) deletePipelineJob(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	err := a.pipelineService.DeleteJob(r.Context(), id)
	if err != nil {
		if err == repository.ErrJobNotFound {
			writeError(w, http.StatusNotFound, "job not found")
			return
		}
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	w.WriteHeader(http.StatusNoContent)
}
