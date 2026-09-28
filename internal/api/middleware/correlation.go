package middleware

import (
	"context"
	"fmt"
	"net/http"
	"time"
)

type contextKey string

const RequestIDKey contextKey = "request_id"

// CorrelationMiddleware injects a unique X-Request-ID into the request context and response headers
func CorrelationMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		reqID := r.Header.Get("X-Request-ID")
		if reqID == "" {
			reqID = fmt.Sprintf("req-%d", time.Now().UnixNano())
		}

		w.Header().Set("X-Request-ID", reqID)
		ctx := context.WithValue(r.Context(), RequestIDKey, reqID)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}
