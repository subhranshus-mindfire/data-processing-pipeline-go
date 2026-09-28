package pipeline

import (
	"context"
	"encoding/json"
	"sync"
	"testing"

	"github.com/user/data-pipeline/internal/domain"
	"github.com/user/data-pipeline/internal/repository"
)

func TestStartExport(t *testing.T) {
	ctx := context.Background()
	resultStore := repository.NewMockResultStore()

	job := &domain.Job{ID: "job-1"}
	resultCh := make(chan SummaryRecord, 1)

	// Create mock summary
	summary := SummaryRecord{
		TotalRecords: 10,
		SourceCounts: map[string]int{"source1": 10},
	}
	resultCh <- summary
	close(resultCh)

	var wg sync.WaitGroup
	wg.Add(1)
	StartExport(ctx, job, resultCh, resultStore, &wg)

	// Verify result
	payload, err := resultStore.GetResult(ctx, "job-1")
	if err != nil {
		t.Fatalf("Failed to get result: %v", err)
	}

	var savedSummary SummaryRecord
	err = json.Unmarshal(payload, &savedSummary)
	if err != nil {
		t.Fatalf("Failed to unmarshal saved summary: %v", err)
	}

	if savedSummary.TotalRecords != 10 {
		t.Errorf("Expected 10 total records, got %d", savedSummary.TotalRecords)
	}
}

func TestStartExportCancelled(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	resultStore := repository.NewMockResultStore()

	job := &domain.Job{ID: "job-2"}
	resultCh := make(chan SummaryRecord)

	var wg sync.WaitGroup
	wg.Add(1)

	cancel() // Cancel before anything is sent
	StartExport(ctx, job, resultCh, resultStore, &wg)

	// Verify no result saved
	_, err := resultStore.GetResult(ctx, "job-2")
	if err != repository.ErrResultNotFound {
		t.Errorf("Expected ErrResultNotFound, got %v", err)
	}
}
