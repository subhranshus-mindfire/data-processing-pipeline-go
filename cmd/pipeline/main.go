package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/user/data-pipeline/internal/api"
	"github.com/user/data-pipeline/internal/service"
	"github.com/user/data-pipeline/internal/store"
)

func main() {
	mux := http.NewServeMux()

	// Initialize dependencies
	pipelineStore := store.NewInMemoryPipelineStore()
	pipelineService := service.NewPipelineService(pipelineStore)
	
	// Initialize API and register routes
	apiHandler := api.NewAPI(pipelineService)
	apiHandler.RegisterRoutes(mux)

	// Apply middleware
	var handler http.Handler = mux
	handler = api.RecoveryMiddleware(handler)
	handler = api.LoggingMiddleware(handler)

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	server := &http.Server{
		Addr:    ":" + port,
		Handler: handler,
	}

	go func() {
		log.Printf("Starting server on port %s", port)
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("Could not listen on %s: %v\n", port, err)
		}
	}()

	// Graceful shutdown
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, os.Interrupt, syscall.SIGTERM)
	<-quit
	log.Println("Shutting down server...")

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := server.Shutdown(ctx); err != nil {
		log.Fatalf("Server forced to shutdown: %v", err)
	}

	log.Println("Server exiting")
}
