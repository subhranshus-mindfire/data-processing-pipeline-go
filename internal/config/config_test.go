package config

import (
	"os"
	"testing"
)

func TestLoadConfig(t *testing.T) {
	// Setup initial state
	os.Clearenv()

	// Test default values
	cfg := LoadConfig()
	if cfg.Port != "8080" {
		t.Errorf("Expected default port 8080, got %s", cfg.Port)
	}
	if cfg.DBPath != "./pipeline.db" {
		t.Errorf("Expected default DB_PATH ./pipeline.db, got %s", cfg.DBPath)
	}
	if cfg.ValidationWorkers != 5 {
		t.Errorf("Expected default VALIDATION_WORKERS 5, got %d", cfg.ValidationWorkers)
	}
	if cfg.TransformationWorkers != 3 {
		t.Errorf("Expected default TRANSFORMATION_WORKERS 3, got %d", cfg.TransformationWorkers)
	}

	// Test custom values
	os.Setenv("PORT", "9090")
	os.Setenv("DB_PATH", "/tmp/test.db")
	os.Setenv("VALIDATION_WORKERS", "10")
	os.Setenv("TRANSFORMATION_WORKERS", "8")

	cfg2 := LoadConfig()
	if cfg2.Port != "9090" {
		t.Errorf("Expected port 9090, got %s", cfg2.Port)
	}
	if cfg2.DBPath != "/tmp/test.db" {
		t.Errorf("Expected DB_PATH /tmp/test.db, got %s", cfg2.DBPath)
	}
	if cfg2.ValidationWorkers != 10 {
		t.Errorf("Expected VALIDATION_WORKERS 10, got %d", cfg2.ValidationWorkers)
	}
	if cfg2.TransformationWorkers != 8 {
		t.Errorf("Expected TRANSFORMATION_WORKERS 8, got %d", cfg2.TransformationWorkers)
	}

	// Test invalid int fallback
	os.Setenv("VALIDATION_WORKERS", "invalid")
	cfg3 := LoadConfig()
	if cfg3.ValidationWorkers != 5 {
		t.Errorf("Expected fallback VALIDATION_WORKERS 5, got %d", cfg3.ValidationWorkers)
	}
}
