package sqlite

import (
	"context"
	"database/sql"
	"testing"
	"time"

	_ "github.com/mattn/go-sqlite3"
	"github.com/user/data-pipeline/internal/domain"
	"github.com/user/data-pipeline/internal/repository"
)

func setupTestDB(t *testing.T) *sql.DB {
	db, err := sql.Open("sqlite3", "file::memory:?cache=shared")
	if err != nil {
		t.Fatalf("Failed to open test database: %v", err)
	}

	// Create tables
	queries := []string{
		`CREATE TABLE IF NOT EXISTS jobs (
			id TEXT PRIMARY KEY,
			status TEXT NOT NULL,
			spec TEXT NOT NULL,
			metrics TEXT,
			created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
			updated_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP
		);`,
		`CREATE TABLE IF NOT EXISTS job_results (
			id TEXT PRIMARY KEY,
			job_id TEXT NOT NULL,
			total_records INTEGER,
			data_payload TEXT,
			created_at DATETIME DEFAULT CURRENT_TIMESTAMP
		);`,
	}

	for _, q := range queries {
		_, err := db.ExecContext(context.Background(), q)
		if err != nil {
			t.Fatalf("Failed to create table: %v", err)
		}
	}

	return db
}

func TestSQLitePipelineStore(t *testing.T) {
	db := setupTestDB(t)
	defer func() {
		_ = db.Close()
	}()

	store := NewSQLitePipelineStore(db)
	ctx := context.Background()

	// 1. Create a job
	job := &domain.Job{
		ID:        "job-1",
		Status:    domain.StatusRunning,
		Spec:      domain.JobSpec{Sources: []domain.SourceConfig{{Type: "csv", URL: "http://example.com"}}},
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	err := store.Create(ctx, job)
	if err != nil {
		t.Fatalf("Failed to create job: %v", err)
	}

	// 2. Get the job
	var fetched *domain.Job
	fetched, err = store.Get(ctx, "job-1")
	if err != nil {
		t.Fatalf("Failed to get job: %v", err)
	}
	if fetched.ID != "job-1" {
		t.Errorf("Expected job-1, got %s", fetched.ID)
	}
	if fetched.Status != domain.StatusRunning {
		t.Errorf("Expected StatusRunning, got %s", fetched.Status)
	}

	// 3. Update the job
	fetched.Status = domain.StatusCompleted
	fetched.Metrics = &domain.Metrics{PercentComplete: 100}
	err = store.Update(ctx, fetched)
	if err != nil {
		t.Fatalf("Failed to update job: %v", err)
	}

	updated, err := store.Get(ctx, "job-1")
	if err != nil {
		t.Fatalf("Failed to get updated job: %v", err)
	}
	if updated.Status != domain.StatusCompleted {
		t.Errorf("Expected StatusCompleted, got %s", updated.Status)
	}
	if updated.Metrics.PercentComplete != 100 {
		t.Errorf("Expected 100%% complete, got %v", updated.Metrics.PercentComplete)
	}

	// 4. List jobs
	job2 := &domain.Job{
		ID:        "job-2",
		Status:    domain.StatusRunning,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}
	err = store.Create(ctx, job2)
	if err != nil {
		t.Fatalf("Failed to create job2: %v", err)
	}

	list, err := store.List(ctx)
	if err != nil {
		t.Fatalf("Failed to list jobs: %v", err)
	}
	if len(list) != 2 {
		t.Errorf("Expected 2 jobs in list, got %d", len(list))
	}

	// 5. Delete job
	err = store.Delete(ctx, "job-1")
	if err != nil {
		t.Fatalf("Failed to delete job: %v", err)
	}

	_, err = store.Get(ctx, "job-1")
	if err != repository.ErrJobNotFound {
		t.Errorf("Expected ErrJobNotFound, got %v", err)
	}
}

func TestSQLiteResultStore(t *testing.T) {
	db := setupTestDB(t)
	defer func() {
		_ = db.Close()
	}()

	store := NewSQLiteResultStore(db)
	ctx := context.Background()

	// 1. Get not found
	_, err := store.GetResult(ctx, "job-1")
	if err != repository.ErrResultNotFound {
		t.Errorf("Expected ErrResultNotFound, got %v", err)
	}

	// 2. Save result
	payload := []byte(`{"data":"success"}`)
	err = store.SaveResult(ctx, "job-1", 100, payload)
	if err != nil {
		t.Fatalf("Failed to save result: %v", err)
	}

	// 3. Get result
	var fetched []byte
	fetched, err = store.GetResult(ctx, "job-1")
	if err != nil {
		t.Fatalf("Failed to get result: %v", err)
	}
	if string(fetched) != string(payload) {
		t.Errorf("Expected payload %s, got %s", string(payload), string(fetched))
	}
}
