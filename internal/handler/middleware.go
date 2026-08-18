package handler

import (
	"log"
	"net/http"
	"time"
)

// LoggingMiddleware logs the details of each incoming HTTP request
func LoggingMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		
		// In a real application, you might want to capture the status code
		// by wrapping the ResponseWriter. For simplicity, we just log the request here.
		log.Printf("%s %s %s", r.RemoteAddr, r.Method, r.URL.Path)
		
		next.ServeHTTP(w, r)
		
		log.Printf("Completed %s in %v", r.URL.Path, time.Since(start))
	})
}

// RecoveryMiddleware recovers from panics and returns a 500 status code
func RecoveryMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		defer func() {
			if err := recover(); err != nil {
				log.Printf("PANIC RECOVERED: %v", err)
				writeError(w, http.StatusInternalServerError, "Internal Server Error")
			}
		}()
		
		next.ServeHTTP(w, r)
	})
}
