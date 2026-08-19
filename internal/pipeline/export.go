package pipeline

import (
	"context"
	"encoding/json"
	"log"
	"sync"

	"github.com/user/data-pipeline/internal/domain"
	"github.com/user/data-pipeline/internal/repository"
)

// StartExport reads the final summary from resultCh and persists it using ResultStore.
func StartExport(ctx context.Context, job *domain.Job, resultCh <-chan SummaryRecord, resultStore repository.ResultStore, wg *sync.WaitGroup) {
	defer wg.Done()

	// Wait for the final result
	var summary SummaryRecord
	select {
	case <-ctx.Done():
		log.Printf("[Job %s] Export cancelled before completion.", job.ID)
		return
	case res, ok := <-resultCh:
		if !ok {
			log.Printf("[Job %s] No results to export.", job.ID)
			return
		}
		summary = res
	}

	// Serialize the payload
	payloadBytes, err := json.Marshal(summary)
	if err != nil {
		log.Printf("[Job %s] Failed to marshal summary payload: %v", job.ID, err)
		return
	}

	// Insert the result via Repository
	if err := resultStore.SaveResult(ctx, job.ID, summary.TotalRecords, payloadBytes); err != nil {
		log.Printf("[Job %s] Failed to export result to DB: %v", job.ID, err)
		return
	}

	log.Printf("[Job %s] Export successful! Results saved to database.", job.ID)
}
