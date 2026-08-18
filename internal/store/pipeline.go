package store

import (
	"context"
	"errors"

	"github.com/user/data-pipeline/internal/domain"
)

var ErrJobNotFound = errors.New("job not found")

// PipelineStore defines the interface for interacting with job data
type PipelineStore interface {
	Create(ctx context.Context, job *domain.Job) error
	Get(ctx context.Context, id string) (*domain.Job, error)
	List(ctx context.Context) ([]*domain.Job, error)
	Update(ctx context.Context, job *domain.Job) error
	Delete(ctx context.Context, id string) error
}
