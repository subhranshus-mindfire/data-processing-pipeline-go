package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"sync"
	"time"

	"github.com/user/data-pipeline/internal/domain"
)

const APIBase = "http://localhost/api/v1/pipelines"
const NumConcurrentJobs = 5

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
	log.Println("✅ ALL STRESS TESTS COMPLETE")
	log.Printf("Total Processing Time for %d jobs: %v", NumConcurrentJobs, duration)
	log.Println("==================================================")
}

func runJob(workerID int, wg *sync.WaitGroup) {
	defer wg.Done()

	// 1. Create the Job
	jobSpec := domain.JobSpec{
		Sources: []domain.SourceConfig{
			{Type: "csv", URL: "http://localhost:8080/samples/stress-input.csv"},
			{Type: "json", URL: "http://localhost:8080/samples/stress-input.json"},
		},
	}

	reqBody, _ := json.Marshal(jobSpec)

	resp, err := http.Post(APIBase, "application/json", bytes.NewReader(reqBody))
	if err != nil {
		log.Printf("[Worker %d] Failed to trigger job: %v", workerID, err)
		return
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusCreated {
		body, _ := io.ReadAll(resp.Body)
		log.Printf("[Worker %d] Expected 201 Created, got %d: %s", workerID, resp.StatusCode, string(body))
		return
	}

	var createResp map[string]string
	json.NewDecoder(resp.Body).Decode(&createResp)
	jobID := createResp["id"]
	if jobID == "" {
		jobID = createResp["job_id"]
	}

	log.Printf("[Worker %d] Successfully triggered Job: %s", workerID, jobID)

	// 2. Poll for Progress
	for {
		time.Sleep(1000 * time.Millisecond)

		progressResp, err := http.Get(fmt.Sprintf("%s/%s/progress", APIBase, jobID))
		if err != nil {
			log.Printf("[Worker %d] Failed to fetch progress: %v", workerID, err)
			return
		}

		var metrics domain.Metrics
		json.NewDecoder(progressResp.Body).Decode(&metrics)
		progressResp.Body.Close()

		log.Printf("[Worker %d - Job %s] Progress: %.2f%% | Processed: %d | Pending: %d | Errors: %d",
			workerID, jobID, metrics.PercentComplete, metrics.RecordsProcessed, metrics.RecordsPending, metrics.ErrorCount)

		// Check job status to see if it's done
		statusResp, err := http.Get(fmt.Sprintf("%s/%s", APIBase, jobID))
		if err != nil {
			log.Printf("[Worker %d] Failed to fetch status: %v", workerID, err)
			return
		}

		var job domain.Job
		json.NewDecoder(statusResp.Body).Decode(&job)
		statusResp.Body.Close()

		if job.Status == domain.StatusCompleted {
			break
		} else if job.Status == domain.StatusFailed || job.Status == domain.StatusCancelled {
			log.Printf("[Worker %d] Job %s ended abruptly with status: %s", workerID, jobID, job.Status)
			return
		}
	}

	// 3. Complete
	log.Printf("[Worker %d] ✅ STRESS TEST JOB %s COMPLETE", workerID, jobID)

	// Print Results Summary
	resResp, _ := http.Get(fmt.Sprintf("%s/%s/results", APIBase, jobID))
	resBody, _ := io.ReadAll(resResp.Body)
	resResp.Body.Close()
	log.Printf("[Worker %d] Final Results Payload: %s", workerID, string(resBody))
}
