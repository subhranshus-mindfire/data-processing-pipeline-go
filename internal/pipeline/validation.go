package pipeline

import (
	"context"
	"fmt"
	"log"
	"sync"

	"github.com/user/data-pipeline/internal/domain"
)

// StartValidationPool creates a pool of workers to validate incoming records.
// Valid records are sent to validatedCh, invalid ones generate an error sent to errCh.
func StartValidationPool(ctx context.Context, numWorkers int, recordsCh <-chan *domain.Record, validatedCh chan<- *domain.Record, errCh chan<- error) {
	var wg sync.WaitGroup

	for i := 1; i <= numWorkers; i++ {
		wg.Add(1)
		go func(workerID int) {
			defer wg.Done()
			defer func() {
				if r := recover(); r != nil {
					errCh <- fmt.Errorf("[Validator-%d] Panic recovered: %v", workerID, r)
				}
			}()
			for {
				select {
				case <-ctx.Done():
					return
				case record, ok := <-recordsCh:
					if !ok {
						return
					}

					// Basic Validation Rules
					if len(record.Data) == 0 {
						errCh <- fmt.Errorf("[Validator-%d] Record %s from %s failed: empty payload", workerID, record.ID, record.Source)
						continue
					}

					record.IsValid = true

					// Optional: Log periodically to show which worker processed it
					if record.ID == "csv-1" || record.ID == "json-1" {
						log.Printf("[Validator-%d] Validated record: %s", workerID, record.ID)
					}

					select {
					case <-ctx.Done():
						return
					case validatedCh <- record:
					}
				}
			}
		}(i)
	}

	// Wait for all workers to finish, then close the validated channel
	go func() {
		wg.Wait()
		log.Println("All validation workers finished. Closing validated channel.")
		close(validatedCh)
	}()
}
