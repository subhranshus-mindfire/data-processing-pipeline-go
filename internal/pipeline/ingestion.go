package pipeline

import (
	"context"
	"encoding/csv"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/user/data-pipeline/internal/domain"
)

// ingestCSV fetches a CSV file via HTTP and pushes parsed records to the channel
func ingestCSV(ctx context.Context, url string, recordsCh chan<- *domain.Record) error {
	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return fmt.Errorf("failed to create CSV request: %w", err)
	}

	client := &http.Client{Timeout: 30 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return fmt.Errorf("failed to fetch CSV: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("failed to fetch CSV, status code: %d", resp.StatusCode)
	}

	reader := csv.NewReader(resp.Body)

	// Read header
	headers, err := reader.Read()
	if err != nil {
		return fmt.Errorf("failed to read CSV headers: %w", err)
	}

	rowCount := 0
	for {
		// Check for context cancellation
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
		}

		row, err := reader.Read()
		if err != nil {
			// eof or other error, break out
			break
		}

		rowCount++
		data := make(map[string]interface{})
		for i, val := range row {
			if i < len(headers) {
				data[headers[i]] = val
			}
		}

		record := &domain.Record{
			ID:        fmt.Sprintf("csv-%d", rowCount),
			Source:    url,
			Data:      data,
			CreatedAt: time.Now(),
		}

		recordsCh <- record
	}

	return nil
}

// ingestJSON fetches a JSON array via HTTP and pushes parsed records to the channel
func ingestJSON(ctx context.Context, url string, recordsCh chan<- *domain.Record) error {
	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return fmt.Errorf("failed to create JSON request: %w", err)
	}

	client := &http.Client{Timeout: 30 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return fmt.Errorf("failed to fetch JSON: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("failed to fetch JSON, status code: %d", resp.StatusCode)
	}

	// Assuming the API returns a JSON array of objects
	var results []map[string]interface{}
	if err := json.NewDecoder(resp.Body).Decode(&results); err != nil {
		return fmt.Errorf("failed to decode JSON array: %w", err)
	}

	for i, item := range results {
		// Check for context cancellation
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
		}

		record := &domain.Record{
			ID:        fmt.Sprintf("json-%d", i+1),
			Source:    url,
			Data:      item,
			CreatedAt: time.Now(),
		}

		recordsCh <- record
	}

	return nil
}
