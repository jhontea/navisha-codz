package config

import (
	"os"
	"strconv"
	"time"
)

// Config holds all application configuration.
type Config struct {
	Port            string
	ProblemsDir     string
	DockerTimeout   time.Duration
	DockerHost       string
	MaxMemoryMB     int
	LogLevel        string
}

// Load creates a Config from environment variables with sensible defaults.
func Load() *Config {
	return &Config{
		Port:          getEnv("PORT", "9100"),
		ProblemsDir:   getEnv("PROBLEMS_DIR", "./problems"),
		DockerTimeout: time.Duration(getEnvInt("SANDBOX_TIMEOUT", 10)) * time.Second,
		DockerHost:     getEnv("DOCKER_HOST", "unix:///var/run/docker.sock"),
		MaxMemoryMB:   getEnvInt("MAX_MEMORY_MB", 256),
		LogLevel:      getEnv("LOG_LEVEL", "info"),
	}
}

func getEnv(key, fallback string) string {
	if v, ok := os.LookupEnv(key); ok {
		return v
	}
	return fallback
}

func getEnvInt(key string, fallback int) int {
	if v, ok := os.LookupEnv(key); ok {
		if n, err := strconv.Atoi(v); err == nil {
			return n
		}
	}
	return fallback
}
