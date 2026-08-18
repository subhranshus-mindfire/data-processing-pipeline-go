package service

import (
	"context"
	"fmt"
	"time"

	"github.com/user/data-pipeline/internal/domain"
	"github.com/user/data-pipeline/internal/engine"
	"github.com/user/data-pipeline/internal/store"
)

// PipelineService defines the business logic interface
type PipelineService interface {
	CreateJob(ctx context.Context, spec domain.JobSpec) (*domain.Job, error)
	GetJob(ctx context.Context, id string) (*domain.Job, error)
	ListJobs(ctx context.Context) ([]*domain.Job, error)
	CancelJob(ctx context.Context, id string) (*domain.Job, error)
	DeleteJob(ctx context.Context, id string) error
}

type pipelineService struct {
	store      store.PipelineStore
	activeJobs map[string]context.CancelFunc
}

func NewPipelineService(store store.PipelineStore) PipelineService {
	return &pipelineService{
		store:      store,
		activeJobs: make(map[string]context.CancelFunc),
	}
}

func (s *pipelineService) CreateJob(ctx context.Context, spec domain.JobSpec) (*domain.Job, error) {
	id := fmt.Sprintf("job-%d", time.Now().UnixNano())
	
	job := &domain.Job{
		ID:        id,
		Status:    domain.StatusRunning,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
		Spec:      spec,
	}
	
	err := s.store.Create(ctx, job)
	if err != nil {
		return nil, err
	}

	// Create a cancellable context for the engine
	jobCtx, cancelFunc := context.WithCancel(context.Background())
	s.activeJobs[job.ID] = cancelFunc

	// Set up channels
	recordsCh := make(chan *domain.Record, 100)
	errCh := make(chan error, 100)

	// Launch the engine orchestrator in a goroutine
	go engine.StartJob(jobCtx, job, recordsCh, errCh, func(j *domain.Job) {
		// This callback is invoked by the engine to update metrics/status
		_ = s.store.Update(context.Background(), j)
	})

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

	if job.Status == domain.StatusCompleted || job.Status == domain.StatusFailed || job.Status == domain.StatusCancelled {
		return nil, fmt.Errorf("cannot cancel job in %s state", job.Status)
	}

	job.Status = domain.StatusCancelled
	err = s.store.Update(ctx, job)
	if err != nil {
		return nil, err
	}

	// Trigger cancellation
	if cancelFunc, exists := s.activeJobs[id]; exists {
		cancelFunc()
		delete(s.activeJobs, id)
	}

	return job, nil
}

func (s *pipelineService) DeleteJob(ctx context.Context, id string) error {
	return s.store.Delete(ctx, id)
}
