package sqlite

import (
	"context"
	"database/sql"
	"encoding/json"
	"time"

	"github.com/user/data-pipeline/internal/domain"
	"github.com/user/data-pipeline/internal/repository"
)

// SQLitePipelineStore is a SQLite implementation of PipelineStore
type SQLitePipelineStore struct {
	db *sql.DB
}

// NewSQLitePipelineStore creates a new SQLitePipelineStore
func NewSQLitePipelineStore(db *sql.DB) *SQLitePipelineStore {
	return &SQLitePipelineStore{
		db: db,
	}
}

// Create inserts a new job into the database
func (s *SQLitePipelineStore) Create(ctx context.Context, job *domain.Job) error {
	specJSON, err := json.Marshal(job.Spec)
	if err != nil {
		return err
	}

	var metricsJSON []byte
	if job.Metrics != nil {
		metricsJSON, err = json.Marshal(job.Metrics)
		if err != nil {
			return err
		}
	}

	query := `
		INSERT INTO jobs (id, status, spec, metrics, created_at, updated_at) 
		VALUES (?, ?, ?, ?, ?, ?)
	`
	_, err = s.db.ExecContext(ctx, query, job.ID, job.Status, string(specJSON), string(metricsJSON), job.CreatedAt, job.UpdatedAt)
	return err
}

// Get retrieves a job by ID from the database
func (s *SQLitePipelineStore) Get(ctx context.Context, id string) (*domain.Job, error) {
	query := `SELECT id, status, spec, metrics, created_at, updated_at FROM jobs WHERE id = ?`

	row := s.db.QueryRowContext(ctx, query, id)

	var job domain.Job
	var specJSON string
	var metricsJSON sql.NullString
	var createdAt, updatedAt time.Time

	err := row.Scan(&job.ID, &job.Status, &specJSON, &metricsJSON, &createdAt, &updatedAt)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, repository.ErrJobNotFound
		}
		return nil, err
	}

	job.CreatedAt = createdAt
	job.UpdatedAt = updatedAt

	// Unmarshal JSON fields
	if err := json.Unmarshal([]byte(specJSON), &job.Spec); err != nil {
		return nil, err
	}

	if metricsJSON.Valid && metricsJSON.String != "" {
		var metrics domain.Metrics
		if err := json.Unmarshal([]byte(metricsJSON.String), &metrics); err != nil {
			return nil, err
		}
		job.Metrics = &metrics
	}

	return &job, nil
}

// List retrieves all jobs from the database
func (s *SQLitePipelineStore) List(ctx context.Context) ([]*domain.Job, error) {
	query := `SELECT id, status, spec, metrics, created_at, updated_at FROM jobs ORDER BY created_at DESC`

	rows, err := s.db.QueryContext(ctx, query)
	if err != nil {
		return nil, err
	}
	defer func() {
		_ = rows.Close()
	}()

	var jobs []*domain.Job
	for rows.Next() {
		var job domain.Job
		var specJSON string
		var metricsJSON sql.NullString
		var createdAt, updatedAt time.Time

		err := rows.Scan(&job.ID, &job.Status, &specJSON, &metricsJSON, &createdAt, &updatedAt)
		if err != nil {
			return nil, err
		}

		job.CreatedAt = createdAt
		job.UpdatedAt = updatedAt

		if err := json.Unmarshal([]byte(specJSON), &job.Spec); err != nil {
			return nil, err
		}

		if metricsJSON.Valid && metricsJSON.String != "" {
			var metrics domain.Metrics
			if err := json.Unmarshal([]byte(metricsJSON.String), &metrics); err != nil {
				return nil, err
			}
			job.Metrics = &metrics
		}

		jobs = append(jobs, &job)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return jobs, nil
}

// Update modifies an existing job in the database
func (s *SQLitePipelineStore) Update(ctx context.Context, job *domain.Job) error {
	specJSON, err := json.Marshal(job.Spec)
	if err != nil {
		return err
	}

	var metricsJSON []byte
	if job.Metrics != nil {
		metricsJSON, err = json.Marshal(job.Metrics)
		if err != nil {
			return err
		}
	}

	job.UpdatedAt = time.Now()

	query := `
		UPDATE jobs 
		SET status = ?, spec = ?, metrics = ?, updated_at = ?
		WHERE id = ? AND status NOT IN ('COMPLETED', 'FAILED', 'CANCELLED')
	`
	res, err := s.db.ExecContext(ctx, query, job.Status, string(specJSON), string(metricsJSON), job.UpdatedAt, job.ID)
	if err != nil {
		return err
	}

	rowsAffected, err := res.RowsAffected()
	if err != nil {
		return err
	}
	if rowsAffected == 0 {
		// If rows affected is 0, it might be because the job was already in a terminal state
		// or it actually doesn't exist. We can verify if it exists.
		var currentStatus string
		err := s.db.QueryRowContext(ctx, "SELECT status FROM jobs WHERE id = ?", job.ID).Scan(&currentStatus)
		if err == sql.ErrNoRows {
			return repository.ErrJobNotFound
		}
		if err != nil {
			return err
		}
		// It exists but was in a terminal state, we just ignore the update.
		return nil
	}

	return nil
}

// Delete removes a job from the database
func (s *SQLitePipelineStore) Delete(ctx context.Context, id string) error {
	query := `DELETE FROM jobs WHERE id = ?`
	res, err := s.db.ExecContext(ctx, query, id)
	if err != nil {
		return err
	}

	rowsAffected, err := res.RowsAffected()
	if err != nil {
		return err
	}
	if rowsAffected == 0 {
		return repository.ErrJobNotFound
	}

	return nil
}
