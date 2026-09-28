package main

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/golang-migrate/migrate/v4"
	"github.com/golang-migrate/migrate/v4/database/sqlite3"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	"github.com/joho/godotenv"
	_ "github.com/mattn/go-sqlite3"

	"github.com/user/data-pipeline/internal/api"
	"github.com/user/data-pipeline/internal/config"
	"github.com/user/data-pipeline/internal/repository/sqlite"
	"github.com/user/data-pipeline/internal/service"
)

// @title Data Processing Pipeline API
// @version 1.0
// @description A robust, concurrent data processing pipeline API.

// @BasePath /
func main() {
	if err := run(); err != nil {
		log.Fatalf("Application error: %v", err)
	}
}

func run() error {
	// Load environment variables
	if err := godotenv.Load(); err != nil {
		log.Println("No .env file found, falling back to system environment variables.")
	}

	cfg := config.LoadConfig()

	mux := http.NewServeMux()

	// Initialize SQLite Database
	db, err := sql.Open("sqlite3", cfg.DBPath)
	if err != nil {
		return fmt.Errorf("failed to open database: %w", err)
	}
	defer func() {
		if closeErr := db.Close(); closeErr != nil {
			log.Printf("Error closing database: %v", closeErr)
		}
	}()

	// Run Migrations
	driver, err := sqlite3.WithInstance(db, &sqlite3.Config{})
	if err != nil {
		return fmt.Errorf("failed to create migration driver: %w", err)
	}
	m, err := migrate.NewWithDatabaseInstance(
		"file://migrations",
		"sqlite3", driver)
	if err != nil {
		return fmt.Errorf("failed to create migration instance: %w", err)
	}

	log.Println("Running database migrations...")
	if err := m.Up(); err != nil && err != migrate.ErrNoChange {
		return fmt.Errorf("failed to run migrations: %w", err)
	}
	log.Println("Migrations complete.")

	// Initialize dependencies
	pipelineStore := sqlite.NewSQLitePipelineStore(db)
	resultStore := sqlite.NewSQLiteResultStore(db)
	pipelineService := service.NewPipelineService(pipelineStore, resultStore, cfg)

	// Initialize API and register routes
	httpHandler := api.RegisterRoutes(mux, pipelineService, cfg)

	server := &http.Server{
		Addr:    ":" + cfg.Port,
		Handler: httpHandler,
	}

	serverErrCh := make(chan error, 1)
	go func() {
		log.Printf("Starting server on port %s", cfg.Port)
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			serverErrCh <- fmt.Errorf("could not listen on %s: %w", cfg.Port, err)
		}
	}()

	// Graceful shutdown
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, os.Interrupt, syscall.SIGTERM)

	select {
	case err := <-serverErrCh:
		return err
	case <-quit:
		log.Println("Shutting down server...")
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := server.Shutdown(ctx); err != nil {
		return fmt.Errorf("server forced to shutdown: %w", err)
	}

	log.Println("Server exiting")
	return nil
}
