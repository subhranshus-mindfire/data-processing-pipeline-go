package api

import (
	"fmt"
	"net/http"
	"runtime"

	httpSwagger "github.com/swaggo/http-swagger"
	_ "github.com/user/data-pipeline/docs" // Blank import to register generated docs
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

	// Basic system metrics endpoint
	mux.HandleFunc("GET /metrics", func(w http.ResponseWriter, r *http.Request) {
		var ms runtime.MemStats
		runtime.ReadMemStats(&ms)
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		_, _ = fmt.Fprintf(w, `{"status":"ok","alloc_mb":%.2f,"sys_mb":%.2f,"gc_cycles":%d}`,
			float64(ms.Alloc)/1024/1024,
			float64(ms.Sys)/1024/1024,
			ms.NumGC,
		)
	})

	// Serve the samples directory statically so the engine can pull files from it via HTTP
	mux.Handle("/samples/", http.StripPrefix("/samples/", http.FileServer(http.Dir("./samples"))))

	// Wrap mux with global middleware
	var h http.Handler = mux
	h = middleware.RecoveryMiddleware(h)
	h = middleware.CorrelationMiddleware(h)
	h = middleware.LoggingMiddleware(h)
	h = middleware.CORSMiddleware(h)

	return h
}
