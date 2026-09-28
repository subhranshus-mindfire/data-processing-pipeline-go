.PHONY: all build test fmt lint run cover docs test-concurrent stress-gen stress-test

all: fmt lint test build docs

build:
	docker compose build

run:
	docker compose up

test:
	docker run --rm -v $(PWD):/app -w /app golang:1.22-alpine sh -c "apk add --no-cache build-base && go test ./... -v"

cover:
	docker run --rm -v $(PWD):/app -w /app golang:1.22-alpine sh -c "apk add --no-cache build-base && go test ./... -coverprofile=coverage.out && go tool cover -func=coverage.out"

docs:
	docker run --rm -v $(PWD):/app -w /app golang:1.22-alpine sh -c "go install github.com/swaggo/swag/cmd/swag@v1.16.2 && swag init -g cmd/pipeline/main.go"

fmt:
	docker run --rm -v $(PWD):/app -w /app golang:1.22-alpine go fmt ./...

lint:
	docker run --rm -v $(PWD):/app -w /app golangci/golangci-lint:v1.59.1 golangci-lint run -v

stress-gen:
	docker run --rm -v $(PWD):/app -w /app golang:1.22-alpine sh -c "go run cmd/stressgen/main.go"

stress-test:
	docker run --rm --network host -v $(PWD):/app -w /app golang:1.22-alpine sh -c "go run cmd/stresstest/main.go"
