package repository

import (
	"context"
	"errors"
)

var ErrResultNotFound = errors.New("result not found")

// ResultStore defines the interface for interacting with job result data
type ResultStore interface {
	SaveResult(ctx context.Context, jobID string, totalRecords int64, payload []byte) error
	GetResult(ctx context.Context, jobID string) ([]byte, error)
}
