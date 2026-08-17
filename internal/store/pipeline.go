package store

import (
	"context"
	"errors"
	"sync"
	"time"

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

// InMemoryPipelineStore is an in-memory implementation of PipelineStore
type InMemoryPipelineStore struct {
	mu   sync.RWMutex
	jobs map[string]*domain.Job
}

func NewInMemoryPipelineStore() *InMemoryPipelineStore {
	return &InMemoryPipelineStore{
		jobs: make(map[string]*domain.Job),
	}
}

func (s *InMemoryPipelineStore) Create(ctx context.Context, job *domain.Job) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.jobs[job.ID] = job
	return nil
}

func (s *InMemoryPipelineStore) Get(ctx context.Context, id string) (*domain.Job, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	job, exists := s.jobs[id]
	if !exists {
		return nil, ErrJobNotFound
	}
	// Return a copy to avoid race conditions on the pointer outside the lock
	jobCopy := *job
	return &jobCopy, nil
}

func (s *InMemoryPipelineStore) List(ctx context.Context) ([]*domain.Job, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	var jobList []*domain.Job
	for _, job := range s.jobs {
		jobCopy := *job
		jobList = append(jobList, &jobCopy)
	}
	return jobList, nil
}

func (s *InMemoryPipelineStore) Update(ctx context.Context, job *domain.Job) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	if _, exists := s.jobs[job.ID]; !exists {
		return ErrJobNotFound
	}
	job.UpdatedAt = time.Now()
	s.jobs[job.ID] = job
	return nil
}

func (s *InMemoryPipelineStore) Delete(ctx context.Context, id string) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	if _, exists := s.jobs[id]; !exists {
		return ErrJobNotFound
	}
	delete(s.jobs, id)
	return nil
}
