package main

import (
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

	// 1. Generate CSV
	csvFile, err := os.Create("samples/stress-input.csv")
	if err != nil {
		log.Fatalf("Failed to create CSV file: %v", err)
	}
	defer csvFile.Close()

	// Write header
	_, err = csvFile.WriteString("id,name,email,role,status\n")
	if err != nil {
		log.Fatalf("Failed to write CSV header: %v", err)
	}

	for i := 1; i <= NumRecords; i++ {
		line := fmt.Sprintf("%d,User%d,user%d@example.com,worker,active\n", i, i, i)
		if _, err := csvFile.WriteString(line); err != nil {
			log.Fatalf("Failed to write CSV line: %v", err)
		}
	}
	log.Printf("Generated samples/stress-input.csv with %d rows", NumRecords)

	// 2. Generate JSON
	jsonFile, err := os.Create("samples/stress-input.json")
	if err != nil {
		log.Fatalf("Failed to create JSON file: %v", err)
	}
	defer jsonFile.Close()

	if _, err := jsonFile.WriteString("[\n"); err != nil {
		log.Fatalf("Failed to write JSON open bracket: %v", err)
	}

	encoder := json.NewEncoder(jsonFile)
	for i := 1; i <= NumRecords; i++ {
		record := map[string]interface{}{
			"id":           i,
			"product_name": fmt.Sprintf("Product %d", i),
			"price":        10.50 + float64(i),
			"in_stock":     i%2 == 0,
		}

		if err := encoder.Encode(record); err != nil {
			log.Fatalf("Failed to encode JSON object: %v", err)
		}

		// JSON encoder adds a newline, we need to add a comma if it's not the last element
		if i < NumRecords {
			jsonFile.WriteString(",")
		}
	}

	if _, err := jsonFile.WriteString("]\n"); err != nil {
		log.Fatalf("Failed to write JSON close bracket: %v", err)
	}
	log.Printf("Generated samples/stress-input.json with %d objects", NumRecords)

	log.Println("Data generation complete!")
}
