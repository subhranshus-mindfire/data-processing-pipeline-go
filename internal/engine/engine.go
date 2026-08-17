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

	// 3. Validation Stage
	validatedCh := make(chan *domain.Record, 100)
	StartValidationPool(ctx, 5, recordsCh, validatedCh, errCh)

	// 4. Transformation Stage
	transformedCh := make(chan *domain.Record, 100)
	StartTransformationPool(ctx, 3, validatedCh, transformedCh, errCh)

	// 5. Error Collection
	go func() {
		for {
			select {
			case <-ctx.Done():
				return
			case err, ok := <-errCh:
				if !ok {
					return
				}
				log.Printf("[Job %s] ERROR: %v", job.ID, err)
			}
		}
	}()

	// 6. Final Drain (Temporary for Day 3)
	// In Day 4, this will be replaced by the Aggregation/Export stage.
	go func() {
		for {
			select {
			case <-ctx.Done():
				log.Printf("[Job %s] Engine shutting down gracefully due to cancellation.", job.ID)
				return
			case record, ok := <-transformedCh:
				if !ok {
					log.Printf("[Job %s] Engine finished processing all records.", job.ID)
					return
				}
				// Log a few fully processed records to verify
				if record.ID == "csv-3" || record.ID == "json-3" {
					log.Printf("[Job %s] Final Transformed Record: %v", job.ID, record.Data)
				}
			}
		}
	}()
}
