package pipeline

import (
	"context"
	"testing"
	"time"

	"github.com/user/data-pipeline/internal/domain"
)

func TestStartAggregation(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	transformedCh := make(chan *domain.Record, 10)
	resultCh := make(chan SummaryRecord, 1)

	// Send some test records
	transformedCh <- &domain.Record{Source: "csv"}
	transformedCh <- &domain.Record{Source: "csv"}
	transformedCh <- &domain.Record{Source: "json"}

	// Close to signal end of stream
	close(transformedCh)

	// Run aggregation synchronously
	StartAggregation(ctx, transformedCh, resultCh)

	select {
	case result := <-resultCh:
		if result.TotalRecords != 3 {
			t.Errorf("Expected 3 total records, got %d", result.TotalRecords)
		}
		if result.SourceCounts["csv"] != 2 {
			t.Errorf("Expected 2 csv records, got %d", result.SourceCounts["csv"])
		}
		if result.SourceCounts["json"] != 1 {
			t.Errorf("Expected 1 json record, got %d", result.SourceCounts["json"])
		}
	case <-time.After(1 * time.Second):
		t.Fatal("Timeout waiting for aggregation result")
	}
}
