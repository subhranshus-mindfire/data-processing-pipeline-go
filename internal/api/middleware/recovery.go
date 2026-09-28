package middleware

import (
	"log"
	"net/http"
)

// RecoveryMiddleware recovers from panics and returns a 500 status code
func RecoveryMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		defer func() {
			if err := recover(); err != nil {
				log.Printf("PANIC RECOVERED: %v", err)
				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(http.StatusInternalServerError)
				if _, writeErr := w.Write([]byte(`{"error": "Internal Server Error"}`)); writeErr != nil {
					log.Printf("Failed to write recovery response: %v", writeErr)
				}
			}
		}()

		next.ServeHTTP(w, r)
	})
}
