package main

import (
	"bufio"
	"encoding/json"
	"fmt"
	"log"
	"os"
)

const (
	NumRecords = 100000
)

func main() {
	log.Println("Starting data generation for stress testing...")

	if err := generateCSV(); err != nil {
		log.Fatalf("CSV generation failed: %v", err)
	}

	if err := generateJSON(); err != nil {
		log.Fatalf("JSON generation failed: %v", err)
	}

	log.Println("Data generation complete!")
}

func generateCSV() error {
	csvFile, err := os.Create("samples/stress-input.csv")
	if err != nil {
		return fmt.Errorf("failed to create CSV file: %w", err)
	}
	defer func() {
		_ = csvFile.Close()
	}()

	writer := bufio.NewWriter(csvFile)

	if _, err := writer.WriteString("id,name,email,role,status\n"); err != nil {
		return fmt.Errorf("failed to write CSV header: %w", err)
	}

	for i := 1; i <= NumRecords; i++ {
		line := fmt.Sprintf("%d,User%d,user%d@example.com,worker,active\n", i, i, i)
		if _, err := writer.WriteString(line); err != nil {
			return fmt.Errorf("failed to write CSV line: %w", err)
		}
	}

	if err := writer.Flush(); err != nil {
		return fmt.Errorf("failed to flush CSV writer: %w", err)
	}

	log.Printf("Generated samples/stress-input.csv with %d rows", NumRecords)
	return nil
}

func generateJSON() error {
	jsonFile, err := os.Create("samples/stress-input.json")
	if err != nil {
		return fmt.Errorf("failed to create JSON file: %w", err)
	}
	defer func() {
		_ = jsonFile.Close()
	}()

	writer := bufio.NewWriter(jsonFile)

	if _, err := writer.WriteString("[\n"); err != nil {
		return fmt.Errorf("failed to write JSON open bracket: %w", err)
	}

	encoder := json.NewEncoder(writer)
	for i := 1; i <= NumRecords; i++ {
		record := map[string]interface{}{
			"id":           i,
			"product_name": fmt.Sprintf("Product %d", i),
			"price":        10.50 + float64(i),
			"in_stock":     i%2 == 0,
		}

		if err := encoder.Encode(record); err != nil {
			return fmt.Errorf("failed to encode JSON object: %w", err)
		}

		if i < NumRecords {
			if _, err := writer.WriteString(","); err != nil {
				return fmt.Errorf("failed to write comma separator: %w", err)
			}
		}
	}

	if _, err := writer.WriteString("]\n"); err != nil {
		return fmt.Errorf("failed to write JSON close bracket: %w", err)
	}

	if err := writer.Flush(); err != nil {
		return fmt.Errorf("failed to flush JSON writer: %w", err)
	}

	log.Printf("Generated samples/stress-input.json with %d objects", NumRecords)
	return nil
}
