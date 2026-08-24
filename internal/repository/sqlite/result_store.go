package sqlite

import (
	"context"
	"database/sql"

	"github.com/user/data-pipeline/internal/repository"
)

// SQLiteResultStore is a SQLite implementation of ResultStore
type SQLiteResultStore struct {
	db *sql.DB
}

// NewSQLiteResultStore creates a new SQLiteResultStore
func NewSQLiteResultStore(db *sql.DB) *SQLiteResultStore {
	return &SQLiteResultStore{
		db: db,
	}
}

// SaveResult inserts a new job result into the database
func (s *SQLiteResultStore) SaveResult(ctx context.Context, jobID string, totalRecords int64, payload []byte) error {
	recordID := "res-" + jobID
	query := `INSERT INTO job_results (id, job_id, total_records, data_payload) VALUES (?, ?, ?, ?)`

	_, err := s.db.ExecContext(ctx, query, recordID, jobID, totalRecords, string(payload))
	return err
}

// GetResult retrieves a job result by jobID from the database
func (s *SQLiteResultStore) GetResult(ctx context.Context, jobID string) ([]byte, error) {
	query := `SELECT data_payload FROM job_results WHERE job_id = ?`

	var payload string
	err := s.db.QueryRowContext(ctx, query, jobID).Scan(&payload)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, repository.ErrResultNotFound
		}
		return nil, err
	}

	return []byte(payload), nil
}
