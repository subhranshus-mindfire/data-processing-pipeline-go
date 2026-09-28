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

	jobID, err := triggerJob(ctx, client, workerID)
	if err != nil {
		log.Printf("[Worker %d] Failed to trigger job: %v", workerID, err)
		return
	}
	log.Printf("[Worker %d] Successfully triggered Job: %s", workerID, jobID)

	if err = pollJob(ctx, client, workerID, jobID); err != nil {
		log.Printf("[Worker %d] Polling failed: %v", workerID, err)
		return
	}

	log.Printf("[Worker %d] STRESS TEST JOB %s COMPLETE", workerID, jobID)
	fetchJobResults(ctx, client, workerID, jobID)
}

func triggerJob(ctx context.Context, client *http.Client, workerID int) (string, error) {
	jobSpec := domain.JobSpec{
		Sources: []domain.SourceConfig{
			{Type: "csv", URL: "http://localhost:8080/samples/stress-input.csv"},
			{Type: "json", URL: "http://localhost:8080/samples/stress-input.json"},
		},
	}

	reqBody, err := json.Marshal(jobSpec)
	if err != nil {
		return "", fmt.Errorf("failed to marshal job spec: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, APIBase, bytes.NewReader(reqBody))
	if err != nil {
		return "", fmt.Errorf("failed to create request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := client.Do(req)
	if err != nil {
		return "", fmt.Errorf("failed to execute request: %w", err)
	}
	defer func() {
		_ = resp.Body.Close()
	}()

	if resp.StatusCode != http.StatusCreated {
		body, _ := io.ReadAll(resp.Body)
		return "", fmt.Errorf("expected 201 Created, got %d: %s", resp.StatusCode, string(body))
	}

	var createResp map[string]string
	if err = json.NewDecoder(resp.Body).Decode(&createResp); err != nil {
		return "", fmt.Errorf("failed to decode create response: %w", err)
	}
	jobID := createResp["id"]
	if jobID == "" {
		jobID = createResp["job_id"]
	}
	return jobID, nil
}

func pollJob(ctx context.Context, client *http.Client, workerID int, jobID string) error {
	for {
		time.Sleep(1000 * time.Millisecond)

		pollCtx, cancel := context.WithTimeout(ctx, 5*time.Second)
		progReq, err := http.NewRequestWithContext(pollCtx, http.MethodGet, fmt.Sprintf("%s/%s/progress", APIBase, jobID), nil)
		if err != nil {
			cancel()
			return fmt.Errorf("failed to create progress request: %w", err)
		}

		progressResp, err := client.Do(progReq)
		if err != nil {
			cancel()
			return fmt.Errorf("failed to fetch progress: %w", err)
		}

		var metrics domain.Metrics
		_ = json.NewDecoder(progressResp.Body).Decode(&metrics)
		_ = progressResp.Body.Close()
		cancel()

		log.Printf("[Worker %d - Job %s] Progress: %.2f%% | Processed: %d | Pending: %d | Errors: %d",
			workerID, jobID, metrics.PercentComplete, metrics.RecordsProcessed, metrics.RecordsPending, metrics.ErrorCount)

		done, err := checkJobStatus(ctx, client, jobID)
		if err != nil {
			return err
		}
		if done {
			return nil
		}
	}
}

func checkJobStatus(ctx context.Context, client *http.Client, jobID string) (bool, error) {
	statusCtx, statusCancel := context.WithTimeout(ctx, 5*time.Second)
	defer statusCancel()

	statReq, err := http.NewRequestWithContext(statusCtx, http.MethodGet, fmt.Sprintf("%s/%s", APIBase, jobID), nil)
	if err != nil {
		return false, fmt.Errorf("failed to create status request: %w", err)
	}

	statusResp, err := client.Do(statReq)
	if err != nil {
		return false, fmt.Errorf("failed to fetch status: %w", err)
	}
	defer func() {
		_ = statusResp.Body.Close()
	}()

	var job domain.Job
	_ = json.NewDecoder(statusResp.Body).Decode(&job)

	if job.Status == domain.StatusCompleted {
		return true, nil
	}
	if job.Status == domain.StatusFailed || job.Status == domain.StatusCancelled {
		return false, fmt.Errorf("job ended with status: %s", job.Status)
	}
	return false, nil
}

func fetchJobResults(ctx context.Context, client *http.Client, workerID int, jobID string) {
	resReq, err := http.NewRequestWithContext(ctx, http.MethodGet, fmt.Sprintf("%s/%s/results", APIBase, jobID), nil)
	if err != nil {
		return
	}
	resResp, err := client.Do(resReq)
	if err != nil {
		return
	}
	defer func() {
		_ = resResp.Body.Close()
	}()
	resBody, _ := io.ReadAll(resResp.Body)
	log.Printf("[Worker %d] Final Results Payload: %s", workerID, string(resBody))
}
