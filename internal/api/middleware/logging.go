package middleware

import (
	"log"
	"net/http"
	"time"
)

// LoggingMiddleware logs the details of each incoming HTTP request
func LoggingMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		reqID, _ := r.Context().Value(RequestIDKey).(string)

		next.ServeHTTP(w, r)

		log.Printf("[HTTP] req_id=%s %s %s from %s took %v", reqID, r.Method, r.URL.Path, r.RemoteAddr, time.Since(start))
	})
}
