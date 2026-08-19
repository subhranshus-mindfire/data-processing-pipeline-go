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

	var responseMap map[string]string
	json.NewDecoder(res.Body).Decode(&responseMap)

	if _, exists := responseMap["job_id"]; !exists {
		t.Error("Expected job_id in response")
	}
	if responseMap["status"] != "job started" {
		t.Errorf("Expected status 'job started', got %s", responseMap["status"])
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
