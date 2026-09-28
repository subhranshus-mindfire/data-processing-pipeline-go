package config

import (
	"os"
	"strconv"
)

type Config struct {
	Port                  string
	DBPath                string
	ValidationWorkers     int
	TransformationWorkers int
	APIKey                string
	MaxConcurrentJobs     int
}

func LoadConfig() *Config {
	return &Config{
		Port:                  getEnv("PORT", "8080"),
		DBPath:                getEnv("DB_PATH", "./pipeline.db"),
		ValidationWorkers:     getEnvAsInt("VALIDATION_WORKERS", 5),
		TransformationWorkers: getEnvAsInt("TRANSFORMATION_WORKERS", 3),
		APIKey:                getEnv("API_KEY", "secret123"),
		MaxConcurrentJobs:     getEnvAsInt("MAX_CONCURRENT_JOBS", 5),
	}
}

func getEnv(key, defaultVal string) string {
	if value, exists := os.LookupEnv(key); exists {
		return value
	}
	return defaultVal
}

func getEnvAsInt(key string, defaultVal int) int {
	valueStr := getEnv(key, "")
	if value, err := strconv.Atoi(valueStr); err == nil {
		return value
	}
	return defaultVal
}
