# Add Comprehensive Test Suite and API Documentation

## 🎯 Objective
This PR introduces a robust automated testing suite for the data pipeline engine and its accompanying services. It also drastically improves developer onboarding by providing concrete sample files and full API documentation.

## 🧪 Testing & Mocking
* **Pipeline Unit Tests**: Added isolated, channel-driven unit tests for the concurrent engine stages (`validation_test.go`, `transformation_test.go`, `aggregation_test.go`). These tests verify that the fan-out/fan-in workers accurately process data and respect context cancellation without needing a database.
* **In-Memory Mocks**: Built `MockPipelineStore` and `MockResultStore` (`mock_store.go`) to safely test business logic without hitting the SQLite disk.
* **Service & API Tests**: 
  * Added `pipeline_service_test.go` to verify job creation and graceful cancellation race-conditions.
  * Added `pipeline_handler_test.go` utilizing Go's `httptest` package to simulate live HTTP traffic and assert proper JSON serialization and status codes.

## 🐛 Bug Fixes
* **Export Type Casting**: Fixed an issue in `export.go` where `summary.TotalRecords` was improperly passed as an `int` instead of an `int64` to the repository.

## 📚 Documentation & Tooling
* **Sample Data**: Added `samples/input.csv` and `samples/input.json` containing realistic payloads for quick manual testing.
* **README Overhaul**: Added a comprehensive `README.md` including a Mermaid architecture diagram, design trade-offs, and exact `cURL` commands for all 5 API endpoints.

## ✅ Checklist
- [x] All unit and integration tests pass via `go test ./...`
- [x] Test suite handles goroutine context cancellation safely
- [x] Documentation is up-to-date and sample payloads are valid
