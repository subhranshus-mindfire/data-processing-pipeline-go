package domain

import (
	"time"
)

// JobStatus represents the current state of a pipeline job
type JobStatus string

const (
	StatusPending   JobStatus = "PENDING"
	StatusRunning   JobStatus = "RUNNING"
	StatusCompleted JobStatus = "COMPLETED"
	StatusFailed    JobStatus = "FAILED"
	StatusCancelled JobStatus = "CANCELLED"
)

// Job represents a data processing pipeline job
type Job struct {
	ID        string    `json:"id"`
	Status    JobStatus `json:"status"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
	// Additional configuration like sources and exports will be added in Day 2
}

// Metrics tracks the progress of a job
type Metrics struct {
	JobID           string    `json:"job_id"`
	RecordsProcessed int64    `json:"records_processed"`
	RecordsPending   int64    `json:"records_pending"`
	ErrorCount       int64    `json:"error_count"`
	PercentComplete  float64  `json:"percent_complete"`
	StartTime        time.Time `json:"start_time"`
	EndTime          *time.Time `json:"end_time,omitempty"`
}

// Record is the generic data structure passing through the pipeline
type Record struct {
	ID        string                 `json:"id,omitempty"`
	Source    string                 `json:"source"`
	Data      map[string]interface{} `json:"data"`
	Errors    []string               `json:"errors,omitempty"`
	IsValid   bool                   `json:"is_valid"`
	CreatedAt time.Time              `json:"created_at"`
}
