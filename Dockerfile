# Build stage
FROM golang:1.22-alpine AS builder

# Install build dependencies for CGO (required for go-sqlite3)
RUN apk add --no-cache build-base

WORKDIR /app

# Copy go mod and sum files
COPY go.mod ./
# COPY go.sum ./

# Download all dependencies.
RUN go mod download

# Install swag CLI
RUN go install github.com/swaggo/swag/cmd/swag@latest

# Copy the source code
COPY . .

# Generate swagger docs
RUN swag init -g cmd/pipeline/main.go

# Build the application (Enable CGO)
RUN CGO_ENABLED=1 GOOS=linux go build -a -installsuffix cgo -o main ./cmd/pipeline

# Final stage
FROM alpine:latest

RUN apk --no-cache add ca-certificates sqlite-libs

WORKDIR /root/

# Copy the Pre-built binary file from the previous stage
COPY --from=builder /app/main .

# Copy the migrations folder so the binary can find it
COPY --from=builder /app/migrations ./migrations

# Copy the samples folder to serve static files
COPY --from=builder /app/samples ./samples

EXPOSE 8080

CMD ["./main"]
