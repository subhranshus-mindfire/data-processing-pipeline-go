package pipeline

import (
	"context"
	"testing"
	"time"

	"github.com/user/data-pipeline/internal/domain"
)

func TestStartTransformationPool(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	validatedCh := make(chan *domain.Record, 10)
	transformedCh := make(chan *domain.Record, 10)
	errCh := make(chan error, 10)

	StartTransformationPool(ctx, 2, validatedCh, transformedCh, errCh)

	// Send a record
	validatedCh <- &domain.Record{
		ID:      "rec-1",
		Source:  "test",
		Data:    map[string]interface{}{"original": "data"},
		IsValid: true,
	}

	close(validatedCh)

	select {
	case transformed := <-transformedCh:
		if transformed.ID != "rec-1" {
			t.Errorf("Expected rec-1, got %s", transformed.ID)
		}
		
		if _, exists := transformed.Data["_processed_at"]; !exists {
			t.Error("Expected _processed_at to be added")
		}
		
		if _, exists := transformed.Data["_processed_by"]; !exists {
			t.Error("Expected _processed_by to be added")
		}
	case <-time.After(1 * time.Second):
		t.Fatal("Timeout waiting for transformed record")
	}
}
