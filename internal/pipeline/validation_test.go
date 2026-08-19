package pipeline

import (
	"context"
	"testing"
	"time"

	"github.com/user/data-pipeline/internal/domain"
)

func TestStartValidationPool(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	recordsCh := make(chan *domain.Record, 10)
	validatedCh := make(chan *domain.Record, 10)
	errCh := make(chan error, 10)

	// Start 2 workers
	StartValidationPool(ctx, 2, recordsCh, validatedCh, errCh)

	// Send a valid record
	recordsCh <- &domain.Record{
		ID:     "valid-1",
		Source: "test",
		Data:   map[string]interface{}{"key": "value"},
	}

	// Send an invalid record (empty payload)
	recordsCh <- &domain.Record{
		ID:     "invalid-1",
		Source: "test",
		Data:   map[string]interface{}{},
	}

	close(recordsCh)

	// We expect 1 valid record and 1 error
	select {
	case validRec := <-validatedCh:
		if validRec.ID != "valid-1" {
			t.Errorf("Expected valid-1, got %s", validRec.ID)
		}
		if !validRec.IsValid {
			t.Error("Expected IsValid to be true")
		}
	case <-time.After(1 * time.Second):
		t.Fatal("Timeout waiting for validated record")
	}

	select {
	case err := <-errCh:
		if err == nil {
			t.Fatal("Expected an error for invalid record, got nil")
		}
	case <-time.After(1 * time.Second):
		t.Fatal("Timeout waiting for error")
	}
}
