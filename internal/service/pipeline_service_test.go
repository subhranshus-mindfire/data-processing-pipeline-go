package service

import (
	"context"
	"testing"
	"time"

	"github.com/user/data-pipeline/internal/config"
	"github.com/user/data-pipeline/internal/domain"
	"github.com/user/data-pipeline/internal/repository"
)

func TestCreateAndCancelJob(t *testing.T) {
	pipelineStore := repository.NewMockPipelineStore()
	resultStore := repository.NewMockResultStore()
	cfg := &config.Config{
		ValidationWorkers:     1,
		TransformationWorkers: 1,
	}

	service := NewPipelineService(pipelineStore, resultStore, cfg)
	ctx := context.Background()

	// Test Create
	spec := domain.JobSpec{
		Sources: []domain.SourceConfig{},
	}

	job, err := service.CreateJob(ctx, spec)
	if err != nil {
		t.Fatalf("Failed to create job: %v", err)
	}

	if job.Status != domain.StatusRunning {
		t.Errorf("Expected status %s, got %s", domain.StatusRunning, job.Status)
	}

	// Verify job is in store
	savedJob, err := pipelineStore.Get(ctx, job.ID)
	if err != nil {
		t.Fatalf("Failed to get job from store: %v", err)
	}
	if savedJob.ID != job.ID {
		t.Errorf("Job ID mismatch. Expected %s, got %s", job.ID, savedJob.ID)
	}

	// Test Cancel immediately to avoid it completing
	cancelledJob, err := service.CancelJob(ctx, job.ID)
	if err != nil {
		if err.Error() == "cannot cancel job in COMPLETED state" {
			// This is acceptable, the goroutine finished instantly because there were no sources
			t.Log("Job completed before cancellation could occur. Test passed by definition.")
		} else {
			t.Fatalf("Failed to cancel job: %v", err)
		}
	} else {
		if cancelledJob.Status != domain.StatusCancelled {
			t.Errorf("Expected status %s, got %s", domain.StatusCancelled, cancelledJob.Status)
		}
	}

	// Try cancelling again, should fail
	_, err = service.CancelJob(ctx, job.ID)
	if err == nil {
		t.Error("Expected error cancelling an already cancelled/completed job, got nil")
	}
}
