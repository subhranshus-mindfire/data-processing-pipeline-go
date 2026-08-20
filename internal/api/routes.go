package api

import (
	"net/http"

	_ "github.com/user/data-pipeline/docs" // Blank import to register generated docs
	httpSwagger "github.com/swaggo/http-swagger"
	"github.com/user/data-pipeline/internal/api/handler"
	"github.com/user/data-pipeline/internal/api/middleware"
	"github.com/user/data-pipeline/internal/config"
	"github.com/user/data-pipeline/internal/service"
)

// RegisterRoutes sets up all the HTTP routes and applies global middleware
func RegisterRoutes(mux *http.ServeMux, pipelineService service.PipelineService, cfg *config.Config) http.Handler {
	// Initialize the API handlers
	apiHandler := handler.NewAPI(pipelineService, cfg)

	// Delegate route registration
	apiHandler.RegisterPipelineRoutes(mux)

	// Swagger documentation route
	mux.Handle("/swagger/", httpSwagger.WrapHandler)

	// Wrap mux with global middleware
	var h http.Handler = mux
	h = middleware.RecoveryMiddleware(h)
	h = middleware.LoggingMiddleware(h)
	h = middleware.CORSMiddleware(h)

	return h
}
