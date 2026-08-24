package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"time"

	"github.com/user/data-pipeline/internal/domain"
)

const APIBase = "http://localhost:8080/api/v1/pipelines"

func main() {
	log.Println("Starting End-to-End Stress Test...")

	// 1. Create the Job
	jobSpec := domain.JobSpec{
		Sources: []domain.SourceConfig{
			{Type: "csv", URL: "http://localhost:8080/samples/stress-input.csv"},
			{Type: "json", URL: "http://localhost:8080/samples/stress-input.json"},
		},
	}

	reqBody, _ := json.Marshal(jobSpec)

	startTime := time.Now()

	resp, err := http.Post(APIBase, "application/json", bytes.NewReader(reqBody))
	if err != nil {
		log.Fatalf("Failed to trigger job: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusCreated {
		body, _ := io.ReadAll(resp.Body)
		log.Fatalf("Expected 201 Created, got %d: %s", resp.StatusCode, string(body))
	}

	var createResp map[string]string
	json.NewDecoder(resp.Body).Decode(&createResp)
	jobID := createResp["id"]
	// fallback if the api returns job_id instead of id
	if jobID == "" {
		jobID = createResp["job_id"]
	}

	log.Printf("Successfully triggered Job: %s", jobID)

	// 2. Poll for Progress
	for {
		time.Sleep(500 * time.Millisecond)

		progressResp, err := http.Get(fmt.Sprintf("%s/%s/progress", APIBase, jobID))
		if err != nil {
			log.Fatalf("Failed to fetch progress: %v", err)
		}

		var metrics domain.Metrics
		json.NewDecoder(progressResp.Body).Decode(&metrics)
		progressResp.Body.Close()

		log.Printf("Progress: %.2f%% | Processed: %d | Pending: %d | Errors: %d",
			metrics.PercentComplete, metrics.RecordsProcessed, metrics.RecordsPending, metrics.ErrorCount)

		// Check job status to see if it's done
		statusResp, err := http.Get(fmt.Sprintf("%s/%s", APIBase, jobID))
		if err != nil {
			log.Fatalf("Failed to fetch status: %v", err)
		}

		var job domain.Job
		json.NewDecoder(statusResp.Body).Decode(&job)
		statusResp.Body.Close()

		if job.Status == domain.StatusCompleted {
			break
		} else if job.Status == domain.StatusFailed || job.Status == domain.StatusCancelled {
			log.Fatalf("Job ended abruptly with status: %s", job.Status)
		}
	}

	// 3. Complete
	duration := time.Since(startTime)
	log.Println("==================================================")
	log.Println("✅ STRESS TEST COMPLETE")
	log.Printf("Total Processing Time: %v", duration)
	log.Println("==================================================")

	// Print Results Summary
	resResp, _ := http.Get(fmt.Sprintf("%s/%s/results", APIBase, jobID))
	resBody, _ := io.ReadAll(resResp.Body)
	resResp.Body.Close()
	log.Printf("Final Results Payload: %s", string(resBody))
}
