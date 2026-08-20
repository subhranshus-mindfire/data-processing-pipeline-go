.PHONY: all build test fmt lint run

all: fmt lint test build

build:
	docker compose build

run:
	docker compose up

test:
	docker run --rm -v $(PWD):/app -w /app golang:1.22-alpine sh -c "apk add --no-cache build-base && go test ./... -v"

fmt:
	docker run --rm -v $(PWD):/app -w /app golang:1.22-alpine go fmt ./...

lint:
	docker run --rm -v $(PWD):/app -w /app golangci/golangci-lint:v1.59.1 golangci-lint run -v
