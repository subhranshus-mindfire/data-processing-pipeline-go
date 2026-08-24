package handler

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/user/data-pipeline/internal/config"
	"github.com/user/data-pipeline/internal/domain"
	"github.com/user/data-pipeline/internal/repository"
	"github.com/user/data-pipeline/internal/service"
)

func TestCreateJobEndpoint(t *testing.T) {
	pipelineStore := repository.NewMockPipelineStore()
	resultStore := repository.NewMockResultStore()
	cfg := &config.Config{ValidationWorkers: 1, TransformationWorkers: 1}

	svc := service.NewPipelineService(pipelineStore, resultStore, cfg)
	apiHandler := NewAPI(svc, cfg)

	mux := http.NewServeMux()
	apiHandler.RegisterPipelineRoutes(mux)

	// Request Body
	spec := domain.JobSpec{
		Sources: []domain.SourceConfig{
			{Type: "csv", URL: "http://example.com/data.csv"},
		},
	}
	body, _ := json.Marshal(spec)
	req := httptest.NewRequest("POST", "/api/v1/pipelines", bytes.NewReader(body))
	w := httptest.NewRecorder()

	mux.ServeHTTP(w, req)

	res := w.Result()
	if res.StatusCode != http.StatusCreated {
		t.Errorf("Expected status %d, got %d", http.StatusCreated, res.StatusCode)
	}

	var responseMap map[string]interface{}
	json.NewDecoder(res.Body).Decode(&responseMap)

	if _, exists := responseMap["id"]; !exists {
		t.Error("Expected id in response")
	}
	if responseMap["status"] != string(domain.StatusRunning) {
		t.Errorf("Expected status '%s', got %v", domain.StatusRunning, responseMap["status"])
	}
}

func TestGetJobResultEndpoint(t *testing.T) {
	pipelineStore := repository.NewMockPipelineStore()
	resultStore := repository.NewMockResultStore()
	cfg := &config.Config{}

	svc := service.NewPipelineService(pipelineStore, resultStore, cfg)
	apiHandler := NewAPI(svc, cfg)

	mux := http.NewServeMux()
	apiHandler.RegisterPipelineRoutes(mux)

	// Pre-populate result
	resultStore.SaveResult(context.Background(), "job-123", 5, []byte(`{"total_records":5}`))

	req := httptest.NewRequest("GET", "/api/v1/pipelines/job-123/results", nil)
	w := httptest.NewRecorder()

	mux.ServeHTTP(w, req)

	res := w.Result()
	if res.StatusCode != http.StatusOK {
		t.Errorf("Expected status %d, got %d", http.StatusOK, res.StatusCode)
	}

	buf := new(bytes.Buffer)
	buf.ReadFrom(res.Body)
	if buf.String() != `{"total_records":5}` {
		t.Errorf("Expected payload, got %s", buf.String())
	}
}

func TestListJobsEndpoint(t *testing.T) {
	pipelineStore := repository.NewMockPipelineStore()
	resultStore := repository.NewMockResultStore()
	cfg := &config.Config{}
	svc := service.NewPipelineService(pipelineStore, resultStore, cfg)
	apiHandler := NewAPI(svc, cfg)

	mux := http.NewServeMux()
	apiHandler.RegisterPipelineRoutes(mux)

	pipelineStore.Create(context.Background(), &domain.Job{ID: "job-1"})
	pipelineStore.Create(context.Background(), &domain.Job{ID: "job-2"})

	req := httptest.NewRequest("GET", "/api/v1/pipelines", nil)
	w := httptest.NewRecorder()
	mux.ServeHTTP(w, req)

	res := w.Result()
	if res.StatusCode != http.StatusOK {
		t.Errorf("Expected status %d, got %d", http.StatusOK, res.StatusCode)
	}

	var jobs []domain.Job
	json.NewDecoder(res.Body).Decode(&jobs)
	if len(jobs) != 2 {
		t.Errorf("Expected 2 jobs, got %d", len(jobs))
	}
}

func TestGetJobEndpoint(t *testing.T) {
	pipelineStore := repository.NewMockPipelineStore()
	resultStore := repository.NewMockResultStore()
	cfg := &config.Config{}
	svc := service.NewPipelineService(pipelineStore, resultStore, cfg)
	apiHandler := NewAPI(svc, cfg)

	mux := http.NewServeMux()
	apiHandler.RegisterPipelineRoutes(mux)

	pipelineStore.Create(context.Background(), &domain.Job{ID: "job-1"})

	// Test valid job
	req := httptest.NewRequest("GET", "/api/v1/pipelines/job-1", nil)
	w := httptest.NewRecorder()
	mux.ServeHTTP(w, req)

	res := w.Result()
	if res.StatusCode != http.StatusOK {
		t.Errorf("Expected status %d, got %d", http.StatusOK, res.StatusCode)
	}

	// Test not found
	reqNotFound := httptest.NewRequest("GET", "/api/v1/pipelines/job-missing", nil)
	wNotFound := httptest.NewRecorder()
	mux.ServeHTTP(wNotFound, reqNotFound)

	if wNotFound.Result().StatusCode != http.StatusNotFound {
		t.Errorf("Expected status %d for missing job, got %d", http.StatusNotFound, wNotFound.Result().StatusCode)
	}
}

func TestGetPipelineProgressEndpoint(t *testing.T) {
	pipelineStore := repository.NewMockPipelineStore()
	resultStore := repository.NewMockResultStore()
	cfg := &config.Config{}
	svc := service.NewPipelineService(pipelineStore, resultStore, cfg)
	apiHandler := NewAPI(svc, cfg)

	mux := http.NewServeMux()
	apiHandler.RegisterPipelineRoutes(mux)

	job := &domain.Job{
		ID:      "job-1",
		Metrics: &domain.Metrics{PercentComplete: 50.5},
	}
	pipelineStore.Create(context.Background(), job)

	req := httptest.NewRequest("GET", "/api/v1/pipelines/job-1/progress", nil)
	w := httptest.NewRecorder()
	mux.ServeHTTP(w, req)

	res := w.Result()
	if res.StatusCode != http.StatusOK {
		t.Errorf("Expected status %d, got %d", http.StatusOK, res.StatusCode)
	}

	var metrics domain.Metrics
	json.NewDecoder(res.Body).Decode(&metrics)
	if metrics.PercentComplete != 50.5 {
		t.Errorf("Expected percent complete 50.5, got %v", metrics.PercentComplete)
	}
}

func TestCancelPipelineJobEndpoint(t *testing.T) {
	pipelineStore := repository.NewMockPipelineStore()
	resultStore := repository.NewMockResultStore()
	cfg := &config.Config{}
	svc := service.NewPipelineService(pipelineStore, resultStore, cfg)
	apiHandler := NewAPI(svc, cfg)

	mux := http.NewServeMux()
	apiHandler.RegisterPipelineRoutes(mux)

	pipelineStore.Create(context.Background(), &domain.Job{ID: "job-1", Status: domain.StatusRunning})

	req := httptest.NewRequest("PATCH", "/api/v1/pipelines/job-1/cancel", nil)
	w := httptest.NewRecorder()
	mux.ServeHTTP(w, req)

	res := w.Result()
	if res.StatusCode != http.StatusOK {
		t.Errorf("Expected status %d, got %d", http.StatusOK, res.StatusCode)
	}

	job, _ := pipelineStore.Get(context.Background(), "job-1")
	if job.Status != domain.StatusCancelled {
		t.Errorf("Expected job status to be cancelled, got %s", job.Status)
	}
}

func TestDeletePipelineJobEndpoint(t *testing.T) {
	pipelineStore := repository.NewMockPipelineStore()
	resultStore := repository.NewMockResultStore()
	cfg := &config.Config{}
	svc := service.NewPipelineService(pipelineStore, resultStore, cfg)
	apiHandler := NewAPI(svc, cfg)

	mux := http.NewServeMux()
	apiHandler.RegisterPipelineRoutes(mux)

	pipelineStore.Create(context.Background(), &domain.Job{ID: "job-1"})

	req := httptest.NewRequest("DELETE", "/api/v1/pipelines/job-1", nil)
	w := httptest.NewRecorder()
	mux.ServeHTTP(w, req)

	res := w.Result()
	if res.StatusCode != http.StatusNoContent {
		t.Errorf("Expected status %d, got %d", http.StatusNoContent, res.StatusCode)
	}

	_, err := pipelineStore.Get(context.Background(), "job-1")
	if err != repository.ErrJobNotFound {
		t.Errorf("Expected ErrJobNotFound, got %v", err)
	}
}
