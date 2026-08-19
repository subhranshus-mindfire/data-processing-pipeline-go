package domain

import "time"

// Record represents a normalized piece of data flowing through the pipeline
type Record struct {
	ID        string
	Source    string
	Data      map[string]interface{}
	IsValid   bool
	ErrorMsgs []string
	CreatedAt time.Time
}
