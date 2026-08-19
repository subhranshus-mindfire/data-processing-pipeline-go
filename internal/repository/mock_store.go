package repository

import (
	"context"
	"sync"

	"github.com/user/data-pipeline/internal/domain"
)

// MockPipelineStore is an in-memory implementation for testing
type MockPipelineStore struct {
	mu   sync.RWMutex
	jobs map[string]*domain.Job
}

func NewMockPipelineStore() *MockPipelineStore {
	return &MockPipelineStore{
		jobs: make(map[string]*domain.Job),
	}
}

func (m *MockPipelineStore) Create(ctx context.Context, job *domain.Job) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.jobs[job.ID] = job
	return nil
}

func (m *MockPipelineStore) Get(ctx context.Context, id string) (*domain.Job, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	job, exists := m.jobs[id]
	if !exists {
		return nil, ErrJobNotFound
	}
	return job, nil
}

func (m *MockPipelineStore) List(ctx context.Context) ([]*domain.Job, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	list := make([]*domain.Job, 0, len(m.jobs))
	for _, job := range m.jobs {
		list = append(list, job)
	}
	return list, nil
}

func (m *MockPipelineStore) Update(ctx context.Context, job *domain.Job) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	if _, exists := m.jobs[job.ID]; !exists {
		return ErrJobNotFound
	}
	m.jobs[job.ID] = job
	return nil
}

func (m *MockPipelineStore) Delete(ctx context.Context, id string) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	if _, exists := m.jobs[id]; !exists {
		return ErrJobNotFound
	}
	delete(m.jobs, id)
	return nil
}

// MockResultStore is an in-memory implementation for testing
type MockResultStore struct {
	mu      sync.RWMutex
	results map[string][]byte
}

func NewMockResultStore() *MockResultStore {
	return &MockResultStore{
		results: make(map[string][]byte),
	}
}

func (m *MockResultStore) SaveResult(ctx context.Context, jobID string, totalRecords int64, payload []byte) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.results[jobID] = payload
	return nil
}

func (m *MockResultStore) GetResult(ctx context.Context, jobID string) ([]byte, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	res, exists := m.results[jobID]
	if !exists {
		return nil, ErrResultNotFound
	}
	return res, nil
}
