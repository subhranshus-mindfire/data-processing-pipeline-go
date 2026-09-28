package middleware

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestCORSMiddleware(t *testing.T) {
	handler := CORSMiddleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))

	// Test Preflight OPTIONS request
	req := httptest.NewRequest("OPTIONS", "/", nil)
	w := httptest.NewRecorder()
	handler.ServeHTTP(w, req)

	res := w.Result()
	defer func() {
		_ = res.Body.Close()
	}()

	if res.StatusCode != http.StatusOK {
		t.Errorf("Expected 200 OK for OPTIONS, got %d", res.StatusCode)
	}
	if w.Header().Get("Access-Control-Allow-Origin") != "*" {
		t.Errorf("Expected CORS origin *, got %s", w.Header().Get("Access-Control-Allow-Origin"))
	}

	// Test GET request
	req2 := httptest.NewRequest("GET", "/", nil)
	w2 := httptest.NewRecorder()
	handler.ServeHTTP(w2, req2)

	res2 := w2.Result()
	defer func() {
		_ = res2.Body.Close()
	}()

	if res2.StatusCode != http.StatusOK {
		t.Errorf("Expected 200 OK for GET, got %d", res2.StatusCode)
	}
	if w2.Header().Get("Access-Control-Allow-Methods") != "GET, POST, PATCH, DELETE, OPTIONS" {
		t.Errorf("Expected CORS methods, got %s", w2.Header().Get("Access-Control-Allow-Methods"))
	}
}

func TestLoggingMiddleware(t *testing.T) {
	handler := LoggingMiddleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusAccepted)
	}))

	req := httptest.NewRequest("GET", "/test", nil)
	w := httptest.NewRecorder()

	// Should not panic, just log
	handler.ServeHTTP(w, req)

	res := w.Result()
	defer func() {
		_ = res.Body.Close()
	}()

	if res.StatusCode != http.StatusAccepted {
		t.Errorf("Expected 202 Accepted, got %d", res.StatusCode)
	}
}

func TestRecoveryMiddleware(t *testing.T) {
	handler := RecoveryMiddleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		panic("test panic")
	}))

	req := httptest.NewRequest("GET", "/", nil)
	w := httptest.NewRecorder()

	handler.ServeHTTP(w, req)

	res := w.Result()
	if res.StatusCode != http.StatusInternalServerError {
		t.Errorf("Expected 500 Internal Server Error, got %d", res.StatusCode)
	}

	defer func() {
		_ = res.Body.Close()
	}()

	var response map[string]string
	if err := json.NewDecoder(res.Body).Decode(&response); err != nil {
		t.Fatalf("Failed to decode response: %v", err)
	}
	if response["error"] != "Internal Server Error" {
		t.Errorf("Expected generic error message, got %s", response["error"])
	}
}

func TestRecoveryMiddlewareNoPanic(t *testing.T) {
	handler := RecoveryMiddleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))

	req := httptest.NewRequest("GET", "/", nil)
	w := httptest.NewRecorder()

	handler.ServeHTTP(w, req)

	res := w.Result()
	defer func() {
		_ = res.Body.Close()
	}()

	if res.StatusCode != http.StatusOK {
		t.Errorf("Expected 200 OK, got %d", res.StatusCode)
	}
}
