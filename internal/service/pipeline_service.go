package service

import (
	"context"
	"fmt"
	"log"
	"sync"
	"time"

	"github.com/user/data-pipeline/internal/config"
	"github.com/user/data-pipeline/internal/domain"
	"github.com/user/data-pipeline/internal/pipeline"
	"github.com/user/data-pipeline/internal/repository"
)

// PipelineService defines the business logic interface
type PipelineService interface {
	CreateJob(ctx context.Context, spec domain.JobSpec) (*domain.Job, error)
	GetJob(ctx context.Context, id string) (*domain.Job, error)
	ListJobs(ctx context.Context) ([]*domain.Job, error)
	CancelJob(ctx context.Context, id string) (*domain.Job, error)
	DeleteJob(ctx context.Context, id string) error
	GetJobResult(ctx context.Context, id string) ([]byte, error)
}

type pipelineService struct {
	store       repository.PipelineStore
	resultStore repository.ResultStore
	activeJobs  map[string]context.CancelFunc
	cfg         *config.Config
	mu          sync.Mutex
}

func NewPipelineService(store repository.PipelineStore, resultStore repository.ResultStore, cfg *config.Config) PipelineService {
	return &pipelineService{
		store:       store,
		resultStore: resultStore,
		activeJobs:  make(map[string]context.CancelFunc),
		cfg:         cfg,
	}
}

func (s *pipelineService) CreateJob(ctx context.Context, spec domain.JobSpec) (*domain.Job, error) {
	s.mu.Lock()
	// MaxConcurrentJobs <= 0 means unlimited
	if s.cfg.MaxConcurrentJobs > 0 && len(s.activeJobs) >= s.cfg.MaxConcurrentJobs {
		s.mu.Unlock()
		return nil, fmt.Errorf("server is at capacity, try again later")
	}
	s.mu.Unlock()

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

	s.mu.Lock()
	s.activeJobs[job.ID] = cancelFunc
	s.mu.Unlock()

	// Set up channels
	recordsCh := make(chan *domain.Record, 100)
	errCh := make(chan error, 100)

	jobCopy := *job
	if job.Metrics != nil {
		mCopy := *job.Metrics
		jobCopy.Metrics = &mCopy
	}

	// Launch the engine orchestrator in a goroutine
	go pipeline.StartJob(jobCtx, job, recordsCh, errCh, s.cfg, s.resultStore, func(j *domain.Job) {
		// This callback is invoked by the engine to update metrics/status
		if updateErr := s.store.Update(context.Background(), j); updateErr != nil {
			log.Printf("[Job %s] Failed to update job in store: %v", j.ID, updateErr)
		}

		if j.Status == domain.StatusCompleted || j.Status == domain.StatusFailed || j.Status == domain.StatusCancelled {
			s.mu.Lock()
			delete(s.activeJobs, j.ID)
			s.mu.Unlock()
		}
	})

	return &jobCopy, nil
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

func (s *pipelineService) GetJobResult(ctx context.Context, id string) ([]byte, error) {
	return s.resultStore.GetResult(ctx, id)
}
