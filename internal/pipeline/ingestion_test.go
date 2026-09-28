package pipeline

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/user/data-pipeline/internal/domain"
)

func TestIngestCSV(t *testing.T) {
	// 1. Setup mock CSV server
	csvData := "id,name,age\n1,Alice,30\n2,Bob,25\n"
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/csv")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(csvData))
	}))
	defer server.Close()

	// 2. Setup channels and context
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()
	recordsCh := make(chan *domain.Record, 10)

	// 3. Run ingestCSV
	err := ingestCSV(ctx, server.URL, recordsCh)
	if err != nil {
		t.Fatalf("ingestCSV failed: %v", err)
	}
	close(recordsCh)

	// 4. Verify results
	var records []*domain.Record
	for r := range recordsCh {
		records = append(records, r)
	}

	if len(records) != 2 {
		t.Errorf("Expected 2 records, got %d", len(records))
	}

	if records[0].Data["name"] != "Alice" {
		t.Errorf("Expected Alice, got %v", records[0].Data["name"])
	}
	if records[1].Data["age"] != "25" {
		t.Errorf("Expected 25, got %v", records[1].Data["age"])
	}
}

func TestIngestJSON(t *testing.T) {
	// 1. Setup mock JSON server
	jsonData := `[{"id":1,"name":"Alice"},{"id":2,"name":"Bob"}]`
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(jsonData))
	}))
	defer server.Close()

	// 2. Setup channels and context
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()
	recordsCh := make(chan *domain.Record, 10)

	// 3. Run ingestJSON
	err := ingestJSON(ctx, server.URL, recordsCh)
	if err != nil {
		t.Fatalf("ingestJSON failed: %v", err)
	}
	close(recordsCh)

	// 4. Verify results
	var records []*domain.Record
	for r := range recordsCh {
		records = append(records, r)
	}

	if len(records) != 2 {
		t.Errorf("Expected 2 records, got %d", len(records))
	}

	// Note: encoding/json decodes numbers to float64
	if records[0].Data["name"] != "Alice" {
		t.Errorf("Expected Alice, got %v", records[0].Data["name"])
	}
}
