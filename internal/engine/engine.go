package engine

import (
	"context"
	"log"
	"sync"
	"sync/atomic"
	"time"

	"github.com/user/data-pipeline/internal/config"
	"github.com/user/data-pipeline/internal/domain"
)

// StartJob orchestrates the entire pipeline: Ingestion -> Validation -> Transformation -> Aggregation -> Export
func StartJob(ctx context.Context, job *domain.Job, recordsCh chan *domain.Record, errCh chan error, cfg *config.Config, onUpdate func(*domain.Job)) {
	log.Printf("[Job %s] Starting engine...", job.ID)
	
	// Atomic metrics tracking
	var recordsProcessed atomic.Int64
	var recordsPending atomic.Int64 // Ingested but not yet finished
	var errorCount atomic.Int64

	// Initialize job metrics
	job.Metrics = &domain.Metrics{
		JobID:     job.ID,
		StartTime: time.Now(),
	}
	onUpdate(job)

	// Progress Tracking Goroutine
	go func() {
		ticker := time.NewTicker(1 * time.Second)
		defer ticker.Stop()
		for {
			select {
			case <-ctx.Done():
				return
			case <-ticker.C:
				proc := recordsProcessed.Load()
				pend := recordsPending.Load()
				errs := errorCount.Load()
				
				job.Metrics.RecordsProcessed = proc
				job.Metrics.RecordsPending = pend
				job.Metrics.ErrorCount = errs
				
				// Very basic percent calculation based on some theoretical total.
				// In a real system, you'd know total rows beforehand to compute this.
				if proc+errs > 0 {
					job.Metrics.PercentComplete = float64(proc) / float64(proc+pend+errs) * 100
				}
				
				onUpdate(job)
			}
		}
	}()

	// 1. Ingestion Stage (Fan-out)
	var ingestWg sync.WaitGroup
	for _, source := range job.Spec.Sources {
		ingestWg.Add(1)
		go func(src domain.SourceConfig) {
			proxyCh := make(chan *domain.Record, 100)
			
			var drainWg sync.WaitGroup
			drainWg.Add(1)
			go func() {
				defer drainWg.Done()
				for r := range proxyCh {
					recordsPending.Add(1)
					recordsCh <- r
				}
			}()

			log.Printf("[Job %s] Starting ingestion from %s (%s)", job.ID, src.URL, src.Type)
			if src.Type == "csv" {
				if err := ingestCSV(ctx, src.URL, proxyCh); err != nil {
					errCh <- err
				}
			} else if src.Type == "json" {
				if err := ingestJSON(ctx, src.URL, proxyCh); err != nil {
					errCh <- err
				}
			}
			
			close(proxyCh)
			drainWg.Wait() // Wait for all records to be sent to recordsCh
			ingestWg.Done() // Now it is safe to signal that this ingestor is done
		}(source)
	}

	// 2. Wait for ingestion
	go func() {
		ingestWg.Wait()
		close(recordsCh)
	}()

	// 3. Validation Stage
	validatedCh := make(chan *domain.Record, 100)
	StartValidationPool(ctx, cfg.ValidationWorkers, recordsCh, validatedCh, errCh)

	// 4. Transformation Stage
	transformedCh := make(chan *domain.Record, 100)
	StartTransformationPool(ctx, cfg.TransformationWorkers, validatedCh, transformedCh, errCh)

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
				errorCount.Add(1)
				log.Printf("[Job %s] ERROR: %v", job.ID, err)
			}
		}
	}()

	// 6. Aggregation Stage
	resultCh := make(chan SummaryRecord, 1)
	go func() {
		// Wrap transformedCh to count processed records
		proxyCh := make(chan *domain.Record, 100)
		go func() {
			for r := range transformedCh {
				recordsProcessed.Add(1)
				recordsPending.Add(-1)
				proxyCh <- r
			}
			close(proxyCh)
		}()
		StartAggregation(ctx, proxyCh, resultCh)
	}()

	// 7. Export Stage
	var exportWg sync.WaitGroup
	exportWg.Add(1)
	go StartExport(ctx, job, resultCh, cfg, &exportWg)

	// 8. Wait for export to finish and mark job complete
	go func() {
		exportWg.Wait()
		
		endTime := time.Now()
		job.Metrics.EndTime = &endTime
		job.Metrics.RecordsProcessed = recordsProcessed.Load()
		job.Metrics.ErrorCount = errorCount.Load()
		job.Metrics.RecordsPending = 0
		job.Metrics.PercentComplete = 100
		
		if job.Status != domain.StatusCancelled {
			job.Status = domain.StatusCompleted
		}
		onUpdate(job)
		log.Printf("[Job %s] Engine has successfully finished all operations.", job.ID)
	}()
}
