package main

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"sync"
	"time"

	"github.com/user/data-pipeline/internal/domain"
)

const (
	APIBase           = "http://localhost:8080/api/v1/pipelines"
	NumConcurrentJobs = 5
)

func main() {
	log.Printf("Starting End-to-End Stress Test with %d concurrent jobs...\n", NumConcurrentJobs)

	var wg sync.WaitGroup
	startTime := time.Now()

	for i := 1; i <= NumConcurrentJobs; i++ {
		wg.Add(1)
		go runJob(i, &wg)
	}

	wg.Wait()

	duration := time.Since(startTime)
	log.Println("==================================================")
	log.Println("ALL STRESS TESTS COMPLETE")
	log.Printf("Total Processing Time for %d jobs: %v", NumConcurrentJobs, duration)
	log.Println("==================================================")
}

func runJob(workerID int, wg *sync.WaitGroup) {
	defer wg.Done()

	client := &http.Client{Timeout: 30 * time.Second}
	ctx := context.Background()

	// 1. Create the Job
	jobSpec := domain.JobSpec{
		Sources: []domain.SourceConfig{
			{Type: "csv", URL: "http://localhost:8080/samples/stress-input.csv"},
			{Type: "json", URL: "http://localhost:8080/samples/stress-input.json"},
		},
	}

	reqBody, err := json.Marshal(jobSpec)
	if err != nil {
		log.Printf("[Worker %d] Failed to marshal job spec: %v", workerID, err)
		return
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, APIBase, bytes.NewReader(reqBody))
	if err != nil {
		log.Printf("[Worker %d] Failed to create request: %v", workerID, err)
		return
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := client.Do(req)
	if err != nil {
		log.Printf("[Worker %d] Failed to trigger job: %v", workerID, err)
		return
	}
	defer func() {
		_ = resp.Body.Close()
	}()

	if resp.StatusCode != http.StatusCreated {
		body, _ := io.ReadAll(resp.Body)
		log.Printf("[Worker %d] Expected 201 Created, got %d: %s", workerID, resp.StatusCode, string(body))
		return
	}

	var createResp map[string]string
	if err := json.NewDecoder(resp.Body).Decode(&createResp); err != nil {
		log.Printf("[Worker %d] Failed to decode create response: %v", workerID, err)
		return
	}
	jobID := createResp["id"]
	if jobID == "" {
		jobID = createResp["job_id"]
	}

	log.Printf("[Worker %d] Successfully triggered Job: %s", workerID, jobID)

	// 2. Poll for Progress
	for {
		time.Sleep(1000 * time.Millisecond)

		pollCtx, cancel := context.WithTimeout(ctx, 5*time.Second)
		progReq, err := http.NewRequestWithContext(pollCtx, http.MethodGet, fmt.Sprintf("%s/%s/progress", APIBase, jobID), nil)
		if err != nil {
			cancel()
			log.Printf("[Worker %d] Failed to create progress request: %v", workerID, err)
			return
		}

		progressResp, err := client.Do(progReq)
		if err != nil {
			cancel()
			log.Printf("[Worker %d] Failed to fetch progress: %v", workerID, err)
			return
		}

		var metrics domain.Metrics
		_ = json.NewDecoder(progressResp.Body).Decode(&metrics)
		_ = progressResp.Body.Close()
		cancel()

		log.Printf("[Worker %d - Job %s] Progress: %.2f%% | Processed: %d | Pending: %d | Errors: %d",
			workerID, jobID, metrics.PercentComplete, metrics.RecordsProcessed, metrics.RecordsPending, metrics.ErrorCount)

		// Check job status to see if it's done
		statusCtx, statusCancel := context.WithTimeout(ctx, 5*time.Second)
		statReq, err := http.NewRequestWithContext(statusCtx, http.MethodGet, fmt.Sprintf("%s/%s", APIBase, jobID), nil)
		if err != nil {
			statusCancel()
			log.Printf("[Worker %d] Failed to create status request: %v", workerID, err)
			return
		}

		statusResp, err := client.Do(statReq)
		if err != nil {
			statusCancel()
			log.Printf("[Worker %d] Failed to fetch status: %v", workerID, err)
			return
		}

		var job domain.Job
		_ = json.NewDecoder(statusResp.Body).Decode(&job)
		_ = statusResp.Body.Close()
		statusCancel()

		if job.Status == domain.StatusCompleted {
			break
		} else if job.Status == domain.StatusFailed || job.Status == domain.StatusCancelled {
			log.Printf("[Worker %d] Job %s ended abruptly with status: %s", workerID, jobID, job.Status)
			return
		}
	}

	// 3. Complete
	log.Printf("[Worker %d] STRESS TEST JOB %s COMPLETE", workerID, jobID)

	// Print Results Summary
	resReq, err := http.NewRequestWithContext(ctx, http.MethodGet, fmt.Sprintf("%s/%s/results", APIBase, jobID), nil)
	if err == nil {
		resResp, err := client.Do(resReq)
		if err == nil {
			resBody, _ := io.ReadAll(resResp.Body)
			_ = resResp.Body.Close()
			log.Printf("[Worker %d] Final Results Payload: %s", workerID, string(resBody))
		}
	}
}
