package engine

import (
	"context"
	"log"
	"sync"
	"time"

	"github.com/user/data-pipeline/internal/domain"
)

// StartTransformationPool creates a pool of workers to transform validated records.
// Transformed records are sent to transformedCh.
func StartTransformationPool(ctx context.Context, numWorkers int, validatedCh <-chan *domain.Record, transformedCh chan<- *domain.Record, errCh chan<- error) {
	var wg sync.WaitGroup

	for i := 1; i <= numWorkers; i++ {
		wg.Add(1)
		go func(workerID int) {
			defer wg.Done()
			for {
				select {
				case <-ctx.Done():
					return
				case record, ok := <-validatedCh:
					if !ok {
						return
					}

					// Transformation: Add processed timestamp
					record.Data["_processed_at"] = time.Now().Format(time.RFC3339)
					
					// Transformation: Add worker ID for tracing
					record.Data["_processed_by"] = workerID

					// Optional: Log periodically
					if record.ID == "csv-2" || record.ID == "json-2" {
						log.Printf("[Transformer-%d] Transformed record: %s", workerID, record.ID)
					}

					select {
					case <-ctx.Done():
						return
					case transformedCh <- record:
					}
				}
			}
		}(i)
	}

	// Wait for all workers to finish, then close the transformed channel
	go func() {
		wg.Wait()
		log.Println("All transformation workers finished. Closing transformed channel.")
		close(transformedCh)
	}()
}
