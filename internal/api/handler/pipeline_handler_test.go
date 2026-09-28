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

const testJobID = "job-1"

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
	defer func() {
		_ = res.Body.Close()
	}()
	if res.StatusCode != http.StatusCreated {
		t.Errorf("Expected status %d, got %d", http.StatusCreated, res.StatusCode)
	}

	var responseMap map[string]interface{}
	if err := json.NewDecoder(res.Body).Decode(&responseMap); err != nil {
		t.Fatalf("Failed to decode response: %v", err)
	}

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
	if err := resultStore.SaveResult(context.Background(), "job-123", 5, []byte(`{"total_records":5}`)); err != nil {
		t.Fatalf("Failed to save result: %v", err)
	}

	req := httptest.NewRequest("GET", "/api/v1/pipelines/job-123/results", nil)
	w := httptest.NewRecorder()

	mux.ServeHTTP(w, req)

	res := w.Result()
	defer func() {
		_ = res.Body.Close()
	}()
	if res.StatusCode != http.StatusOK {
		t.Errorf("Expected status %d, got %d", http.StatusOK, res.StatusCode)
	}

	buf := new(bytes.Buffer)
	if _, err := buf.ReadFrom(res.Body); err != nil {
		t.Fatalf("Failed to read body: %v", err)
	}
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

	if err := pipelineStore.Create(context.Background(), &domain.Job{ID: testJobID}); err != nil {
		t.Fatalf("Failed to create job-1: %v", err)
	}
	if err := pipelineStore.Create(context.Background(), &domain.Job{ID: "job-2"}); err != nil {
		t.Fatalf("Failed to create job-2: %v", err)
	}

	req := httptest.NewRequest("GET", "/api/v1/pipelines", nil)
	w := httptest.NewRecorder()
	mux.ServeHTTP(w, req)

	res := w.Result()
	defer func() {
		_ = res.Body.Close()
	}()
	if res.StatusCode != http.StatusOK {
		t.Errorf("Expected status %d, got %d", http.StatusOK, res.StatusCode)
	}

	var jobs []domain.Job
	if err := json.NewDecoder(res.Body).Decode(&jobs); err != nil {
		t.Fatalf("Failed to decode jobs: %v", err)
	}
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

	if err := pipelineStore.Create(context.Background(), &domain.Job{ID: testJobID}); err != nil {
		t.Fatalf("Failed to create job-1: %v", err)
	}

	// Test valid job
	req := httptest.NewRequest("GET", "/api/v1/pipelines/job-1", nil)
	w := httptest.NewRecorder()
	mux.ServeHTTP(w, req)

	res := w.Result()
	defer func() {
		_ = res.Body.Close()
	}()
	if res.StatusCode != http.StatusOK {
		t.Errorf("Expected status %d, got %d", http.StatusOK, res.StatusCode)
	}

	// Test not found
	reqNotFound := httptest.NewRequest("GET", "/api/v1/pipelines/job-missing", nil)
	wNotFound := httptest.NewRecorder()
	mux.ServeHTTP(wNotFound, reqNotFound)

	resNotFound := wNotFound.Result()
	defer func() {
		_ = resNotFound.Body.Close()
	}()
	if resNotFound.StatusCode != http.StatusNotFound {
		t.Errorf("Expected status %d for missing job, got %d", http.StatusNotFound, resNotFound.StatusCode)
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
		ID:      testJobID,
		Metrics: &domain.Metrics{PercentComplete: 50.5},
	}
	if err := pipelineStore.Create(context.Background(), job); err != nil {
		t.Fatalf("Failed to create job: %v", err)
	}

	req := httptest.NewRequest("GET", "/api/v1/pipelines/job-1/progress", nil)
	w := httptest.NewRecorder()
	mux.ServeHTTP(w, req)

	res := w.Result()
	defer func() {
		_ = res.Body.Close()
	}()
	if res.StatusCode != http.StatusOK {
		t.Errorf("Expected status %d, got %d", http.StatusOK, res.StatusCode)
	}

	var metrics domain.Metrics
	if err := json.NewDecoder(res.Body).Decode(&metrics); err != nil {
		t.Fatalf("Failed to decode metrics: %v", err)
	}
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

	if err := pipelineStore.Create(context.Background(), &domain.Job{ID: testJobID, Status: domain.StatusRunning}); err != nil {
		t.Fatalf("Failed to create job: %v", err)
	}

	req := httptest.NewRequest("PATCH", "/api/v1/pipelines/job-1/cancel", nil)
	w := httptest.NewRecorder()
	mux.ServeHTTP(w, req)

	res := w.Result()
	defer func() {
		_ = res.Body.Close()
	}()
	if res.StatusCode != http.StatusOK {
		t.Errorf("Expected status %d, got %d", http.StatusOK, res.StatusCode)
	}

	job, err := pipelineStore.Get(context.Background(), testJobID)
	if err != nil {
		t.Fatalf("Failed to get job: %v", err)
	}
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

	if err := pipelineStore.Create(context.Background(), &domain.Job{ID: testJobID}); err != nil {
		t.Fatalf("Failed to create job: %v", err)
	}

	req := httptest.NewRequest("DELETE", "/api/v1/pipelines/job-1", nil)
	w := httptest.NewRecorder()
	mux.ServeHTTP(w, req)

	res := w.Result()
	defer func() {
		_ = res.Body.Close()
	}()
	if res.StatusCode != http.StatusNoContent {
		t.Errorf("Expected status %d, got %d", http.StatusNoContent, res.StatusCode)
	}

	_, err := pipelineStore.Get(context.Background(), testJobID)
	if err != repository.ErrJobNotFound {
		t.Errorf("Expected ErrJobNotFound, got %v", err)
	}
}
