package pipeline

import (
	"context"
	"log"

	"github.com/user/data-pipeline/internal/domain"
)

// SummaryRecord represents the final aggregated data
type SummaryRecord struct {
	TotalRecords int              `json:"total_records"`
	SourceCounts map[string]int   `json:"source_counts"`
}

// StartAggregation collects all transformed records and generates a summary.
// Once transformedCh is closed, it pushes the final summary to resultCh.
func StartAggregation(ctx context.Context, transformedCh <-chan *domain.Record, resultCh chan<- SummaryRecord) {
	summary := SummaryRecord{
		TotalRecords: 0,
		SourceCounts: make(map[string]int),
	}

	for {
		select {
		case <-ctx.Done():
			log.Println("Aggregation cancelled.")
			return
		case record, ok := <-transformedCh:
			if !ok {
				// transformedCh is closed, meaning all workers have finished.
				// Push the final summary and close resultCh.
				log.Printf("Aggregation complete. Total records: %d", summary.TotalRecords)
				resultCh <- summary
				close(resultCh)
				return
			}

			summary.TotalRecords++
			summary.SourceCounts[record.Source]++
		}
	}
}
