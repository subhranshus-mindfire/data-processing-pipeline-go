package service

import (
	"context"
	"fmt"
	"time"

	"github.com/user/data-pipeline/internal/domain"
	"github.com/user/data-pipeline/internal/store"
)

// PipelineService defines the business logic interface
type PipelineService interface {
	CreateJob(ctx context.Context, id string) (*domain.Job, error)
	GetJob(ctx context.Context, id string) (*domain.Job, error)
	ListJobs(ctx context.Context) ([]*domain.Job, error)
	CancelJob(ctx context.Context, id string) (*domain.Job, error)
	DeleteJob(ctx context.Context, id string) error
}

type pipelineService struct {
	store store.PipelineStore
}

func NewPipelineService(store store.PipelineStore) PipelineService {
	return &pipelineService{store: store}
}

func (s *pipelineService) CreateJob(ctx context.Context, id string) (*domain.Job, error) {
	// If id is empty, we would generate a UUID here. For now we use the provided mock ID.
	if id == "" {
		id = fmt.Sprintf("job-%d", time.Now().UnixNano())
	}
	job := &domain.Job{
		ID:        id,
		Status:    domain.StatusPending,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}
	err := s.store.Create(ctx, job)
	if err != nil {
		return nil, err
	}
	return job, nil
}

func (s *pipelineService) GetJob(ctx context.Context, id string) (*domain.Job, error) {
	return s.store.Get(ctx, id)
}

func (s *pipelineService) ListJobs(ctx context.Context) ([]*domain.Job, error) {
	return s.store.List(ctx)
}

func (s *pipelineService) CancelJob(ctx context.Context, id string) (*domain.Job, error) {
	job, err := s.store.Get(ctx, id)
	if err != nil {
		return nil, err
	}

	if job.Status == domain.StatusCompleted || job.Status == domain.StatusFailed {
		return nil, fmt.Errorf("cannot cancel job in %s state", job.Status)
	}

	job.Status = domain.StatusCancelled
	err = s.store.Update(ctx, job)
	if err != nil {
		return nil, err
	}
	return job, nil
}

func (s *pipelineService) DeleteJob(ctx context.Context, id string) error {
	return s.store.Delete(ctx, id)
}
