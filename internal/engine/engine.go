package engine

import (
	"context"
	"log"
	"sync"

	"github.com/user/data-pipeline/internal/domain"
)

// StartJob is the main orchestrator for the pipeline engine.
// It initializes channels, spawns ingestors, and manages the data flow.
func StartJob(ctx context.Context, job *domain.Job, recordsCh chan *domain.Record, errCh chan error) {
	log.Printf("[Job %s] Starting engine...", job.ID)
	
	// Create a wait group for the ingestion stage
	var ingestWg sync.WaitGroup

	// 1. Ingestion Stage (Fan-out)
	for _, source := range job.Spec.Sources {
		ingestWg.Add(1)
		go func(src domain.SourceConfig) {
			defer ingestWg.Done()
			
			log.Printf("[Job %s] Starting ingestion from %s (%s)", job.ID, src.URL, src.Type)
			
			if src.Type == "csv" {
				if err := ingestCSV(ctx, src.URL, recordsCh); err != nil {
					errCh <- err
				}
			} else if src.Type == "json" {
				if err := ingestJSON(ctx, src.URL, recordsCh); err != nil {
					errCh <- err
				}
			} else {
				log.Printf("[Job %s] Unknown source type: %s", job.ID, src.Type)
			}
		}(source)
	}

	// 2. Wait for all ingestion to finish, then close the records channel
	go func() {
		ingestWg.Wait()
		log.Printf("[Job %s] All ingestion complete. Closing records channel.", job.ID)
		close(recordsCh)
	}()

	// 3. For Day 2, we just drain the channel so it doesn't block.
	// In Day 3, this will be replaced by Validation/Transformation workers.
	go func() {
		for {
			select {
			case <-ctx.Done():
				log.Printf("[Job %s] Engine shutting down gracefully due to cancellation.", job.ID)
				return
			case record, ok := <-recordsCh:
				if !ok {
					// recordsCh closed, pipeline complete
					log.Printf("[Job %s] Engine finished draining records.", job.ID)
					return
				}
				// Simulate some tiny work and log periodically
				if record.ID == "" || record.ID == "1" || record.ID == "100" {
					log.Printf("[Job %s] Ingested record: %v", job.ID, record.Data)
				}
			}
		}
	}()
}
