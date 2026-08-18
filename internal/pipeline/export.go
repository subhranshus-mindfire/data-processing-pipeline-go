package pipeline

import (
	"context"
	"database/sql"
	"encoding/json"
	"log"
	"sync"

	_ "github.com/mattn/go-sqlite3"
	"github.com/user/data-pipeline/internal/config"
	"github.com/user/data-pipeline/internal/domain"
)

// StartExport reads the final summary from resultCh and persists it to a SQLite database.
func StartExport(ctx context.Context, job *domain.Job, resultCh <-chan SummaryRecord, cfg *config.Config, wg *sync.WaitGroup) {
	defer wg.Done()

	// Wait for the final result
	var summary SummaryRecord
	select {
	case <-ctx.Done():
		log.Printf("[Job %s] Export cancelled before completion.", job.ID)
		return
	case res, ok := <-resultCh:
		if !ok {
			log.Printf("[Job %s] No results to export.", job.ID)
			return
		}
		summary = res
	}

	// Open SQLite database
	db, err := sql.Open("sqlite3", cfg.ExportsDBPath)
	if err != nil {
		log.Printf("[Job %s] Failed to open SQLite database: %v", job.ID, err)
		return
	}
	defer db.Close()

	// Auto-Migrate: Create table if it doesn't exist
	createTableSQL := `CREATE TABLE IF NOT EXISTS job_results (
		id TEXT PRIMARY KEY,
		job_id TEXT NOT NULL,
		total_records INTEGER,
		data_payload TEXT,
		created_at DATETIME DEFAULT CURRENT_TIMESTAMP
	);`
	if _, err := db.Exec(createTableSQL); err != nil {
		log.Printf("[Job %s] Failed to create job_results table: %v", job.ID, err)
		return
	}

	// Serialize the payload
	payloadBytes, err := json.Marshal(summary)
	if err != nil {
		log.Printf("[Job %s] Failed to marshal summary payload: %v", job.ID, err)
		return
	}

	// Insert the result
	insertSQL := `INSERT INTO job_results (id, job_id, total_records, data_payload) VALUES (?, ?, ?, ?)`
	recordID := "res-" + job.ID
	
	if _, err := db.Exec(insertSQL, recordID, job.ID, summary.TotalRecords, string(payloadBytes)); err != nil {
		log.Printf("[Job %s] Failed to export result to DB: %v", job.ID, err)
		return
	}

	log.Printf("[Job %s] Export successful! Results saved to exports.db", job.ID)
}
