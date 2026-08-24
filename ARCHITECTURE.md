# Pipeline Architecture

This document provides a detailed breakdown of the internal workings of the Go Data Processing Pipeline, focusing specifically on the **concurrency model** and how data flows through the application.

## Overview

The pipeline is designed using a **Fan-Out / Fan-In** concurrency pattern. It leverages Go's powerful `goroutines` and `channels` to process massive amounts of data efficiently without blocking.

The system is orchestrated by the `StartJob` function located in `internal/pipeline/pipeline.go`. When an API request is received to create a job, this orchestrator wires together a series of buffered channels and spawns worker pools for the various processing stages.

## Stage 1: Ingestion (Fan-Out)
**File:** `internal/pipeline/ingestion.go`

The ingestion phase is responsible for reading raw data from external sources (like CSV files or JSON APIs) and converting them into uniform `domain.Record` structures.

- **Source-Level Concurrency:** For every source specified in the `JobSpec`, the orchestrator spins up a dedicated goroutine. 
- **Sequential Reading:** Within that goroutine, the source (e.g., a massive CSV file) is read sequentially row-by-row.
- **Emission:** As soon as a row is parsed, it is immediately pushed onto the shared `recordsCh` channel. This ensures downstream workers can begin processing row 1 even while row 1,000,000 is still being downloaded.

## Stage 2: Validation (Worker Pool)
**File:** `internal/pipeline/validation.go`

Once data is in the `recordsCh`, it enters the validation phase. 
- **Concurrent Processing:** The `StartValidationPool` function spawns a configurable number of worker goroutines (e.g., 5 validators). 
- **Shared Queue:** All these workers constantly pull from the same `recordsCh`. Because channel reads in Go are thread-safe and atomic, the workers naturally distribute the load among themselves without complex locking mechanisms.
- **Output:** Valid records are pushed to the `validatedCh`. Invalid records are dropped, and an error is pushed to the `errCh`.

## Stage 3: Transformation (Worker Pool)
**File:** `internal/pipeline/transformation.go`

Similar to Validation, the Transformation stage utilizes a worker pool.
- The `StartTransformationPool` function spins up multiple workers that pull from the `validatedCh`.
- Here, data enrichment, data scrubbing, and formatting occur (e.g., adding `_processed_at` timestamps).
- Transformed records are pushed to the `transformedCh`.

## Stage 4: Aggregation (Fan-In)
**File:** `internal/pipeline/aggregation.go`

Unlike Validation and Transformation, Aggregation must maintain a stateful summary (e.g., total row count, sums, averages). 
- **Single Worker (Fan-In):** To avoid race conditions and the performance overhead of using `sync.Mutex` on a shared map, Aggregation is performed by a **single** goroutine.
- All transformed records from the multiple Transformation workers "fan-in" to the single `StartAggregation` consumer via the `transformedCh`. 
- Once the `transformedCh` is closed (indicating all sources have been fully ingested, validated, and transformed), this single worker computes the final `SummaryRecord` and pushes it to `resultCh`.

## Stage 5: Export
**File:** `internal/pipeline/export.go`

The final stage waits for the `SummaryRecord` to arrive on the `resultCh`. 
- It then writes this summary to the SQLite database using the `ResultStore`. 
- Finally, it marks the Job as `COMPLETED` and calculates the total duration.

---

## Observability & Graceful Cancellation

### Progress Tracking
A dedicated ticker goroutine (in `pipeline.go`) wakes up every second to read atomic counters (`atomic.Int64`) for processed records, pending records, and errors. It updates the Job's `Metrics` struct, which powers the `/api/v1/pipelines/{id}/progress` endpoint for real-time dashboard updates.

### Context Cancellation
The entire pipeline is deeply integrated with `context.Context`. If a user cancels a job via the API, the context is cancelled. Every stage (Ingestion, Validation, Transformation, Aggregation) contains `select` statements that listen for `ctx.Done()`. Upon receiving this signal, all goroutines immediately exit, closing channels and preventing zombie processes or memory leaks.
