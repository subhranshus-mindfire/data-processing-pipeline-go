package middleware

import "net/http"

// AuthMiddleware is a stub for future authentication logic
func AuthMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// TODO: Implement authentication (e.g., check JWT token)
		next.ServeHTTP(w, r)
	})
}
