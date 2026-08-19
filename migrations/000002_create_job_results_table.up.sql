CREATE TABLE IF NOT EXISTS job_results (
	id TEXT PRIMARY KEY,
	job_id TEXT NOT NULL,
	total_records INTEGER,
	data_payload TEXT,
	created_at DATETIME DEFAULT CURRENT_TIMESTAMP
);
