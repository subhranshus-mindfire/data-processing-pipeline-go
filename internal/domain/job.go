package domain

import "time"

// JobStatus represents the current state of a job
type JobStatus string

const (
	StatusRunning   JobStatus = "RUNNING"
	StatusCompleted JobStatus = "COMPLETED"
	StatusFailed    JobStatus = "FAILED"
	StatusCancelled JobStatus = "CANCELLED"
)

// SourceConfig represents a data source configuration
type SourceConfig struct {
	Type string `json:"type"`
	URL  string `json:"url"`
}

// JobSpec represents the configuration for a pipeline job
type JobSpec struct {
	Sources       []SourceConfig `json:"sources"`
	ExportTargets []string       `json:"export_targets"`
}

// Metrics tracks the real-time progress of a job
type Metrics struct {
	JobID            string     `json:"job_id"`
	RecordsProcessed int64      `json:"records_processed"`
	RecordsPending   int64      `json:"records_pending"`
	ErrorCount       int64      `json:"error_count"`
	LastError        string     `json:"last_error,omitempty"`
	PercentComplete  float64    `json:"percent_complete"`
	StartTime        time.Time  `json:"start_time"`
	EndTime          *time.Time `json:"end_time"`
}

// Job represents a data processing job
type Job struct {
	ID        string    `json:"id"`
	Status    JobStatus `json:"status"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
	Spec      JobSpec   `json:"spec"`
	Metrics   *Metrics  `json:"metrics,omitempty"`
}
