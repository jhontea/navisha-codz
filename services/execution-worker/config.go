package main

import (
	"fmt"
	"os"
	"strconv"
	"strings"
	"time"
)

// Difficulty level constants.
const (
	DifficultyEasy   = "easy"
	DifficultyMedium = "medium"
	DifficultyHard   = "hard"
)

// ResourceLimits defines the resource constraints for a specific difficulty or category.
type ResourceLimits struct {
	// TimeLimit is the maximum execution time per test case.
	TimeLimit time.Duration `json:"time_limit"`
	// MemoryLimitMB is the maximum memory usage in megabytes.
	MemoryLimitMB int `json:"memory_limit_mb"`
	// CPUQuota is the CPU quota (e.g., "50000" for 50% of a CPU core).
	CPUQuota string `json:"cpu_quota"`
	// PIDsLimit is the maximum number of processes.
	PIDsLimit int `json:"pids_limit"`
	// MaxOutputSize is the maximum allowed output size in bytes.
	MaxOutputSize int64 `json:"max_output_size"`
	// DiskLimitMB is the maximum disk usage in megabytes.
	DiskLimitMB int `json:"disk_limit_mb"`
}

// DefaultResourceLimits returns the default resource limits for a given difficulty.
func DefaultResourceLimits(difficulty string) ResourceLimits {
	switch strings.ToLower(difficulty) {
	case DifficultyEasy:
		return ResourceLimits{
			TimeLimit:     1 * time.Second,
			MemoryLimitMB: 256,
			CPUQuota:      "50000", // 50% of a CPU core
			PIDsLimit:     50,
			MaxOutputSize: 256 * 1024, // 256KB
			DiskLimitMB:   50,
		}
	case DifficultyMedium:
		return ResourceLimits{
			TimeLimit:     3 * time.Second,
			MemoryLimitMB: 512,
			CPUQuota:      "100000", // 1 full CPU core
			PIDsLimit:     100,
			MaxOutputSize: 512 * 1024, // 512KB
			DiskLimitMB:   100,
		}
	case DifficultyHard:
		return ResourceLimits{
			TimeLimit:     5 * time.Second,
			MemoryLimitMB: 1024,
			CPUQuota:      "200000", // 2 CPU cores
			PIDsLimit:     200,
			MaxOutputSize: 1 * 1024 * 1024, // 1MB
			DiskLimitMB:   200,
		}
	default:
		// Default to medium limits
		return ResourceLimits{
			TimeLimit:     3 * time.Second,
			MemoryLimitMB: 512,
			CPUQuota:      "100000",
			PIDsLimit:     100,
			MaxOutputSize: 512 * 1024,
			DiskLimitMB:   100,
		}
	}
}

// CategoryOverride holds per-category resource overrides.
type CategoryOverride struct {
	// Category is the problem category (e.g., "algorithms", "data-structures", "dp").
	Category string `json:"category"`
	// ResourceLimits overrides for this category.
	ResourceLimits
}

// WorkerConfig holds the complete execution worker configuration.
type WorkerConfig struct {
	// MaxConcurrentExecutions is the maximum number of concurrent test case executions.
	MaxConcurrentExecutions int `json:"max_concurrent_executions"`
	// QueueCapacity is the maximum number of queued submissions per worker.
	QueueCapacity int `json:"queue_capacity"`
	// WarmPoolSize is the number of warm containers to keep ready.
	WarmPoolSize int `json:"warm_pool_size"`
	// DefaultDifficulty is the default difficulty level.
	DefaultDifficulty string `json:"default_difficulty"`
	// DefaultResourceLimits is the fallback resource limits.
	DefaultResourceLimits ResourceLimits `json:"default_resource_limits"`
	// DifficultyLimits maps difficulty levels to resource limits.
	DifficultyLimits map[string]ResourceLimits `json:"difficulty_limits"`
	// CategoryOverrides maps problem categories to resource overrides.
	CategoryOverrides []CategoryOverride `json:"category_overrides"`
	// SandboxImage is the Docker sandbox image.
	SandboxImage string `json:"sandbox_image"`
	// SandboxWorkDir is the working directory for sandbox operations.
	SandboxWorkDir string `json:"sandbox_work_dir"`
	// EnableLocalFallback enables local execution when Docker is unavailable.
	EnableLocalFallback bool `json:"enable_local_fallback"`
	// GracePeriod is the additional time before SIGKILL after timeout.
	GracePeriod time.Duration `json:"grace_period"`
	// CompiledBinCacheSize is the maximum number of cached compiled binaries.
	CompiledBinCacheSize int `json:"compiled_bin_cache_size"`
}

// DefaultWorkerConfig returns the default worker configuration.
func DefaultWorkerConfig() WorkerConfig {
	return WorkerConfig{
		MaxConcurrentExecutions: 4,
		QueueCapacity:           100,
		WarmPoolSize:            2,
		DefaultDifficulty:       DifficultyMedium,
		DefaultResourceLimits:   DefaultResourceLimits(DifficultyMedium),
		DifficultyLimits: map[string]ResourceLimits{
			DifficultyEasy:   DefaultResourceLimits(DifficultyEasy),
			DifficultyMedium: DefaultResourceLimits(DifficultyMedium),
			DifficultyHard:   DefaultResourceLimits(DifficultyHard),
		},
		CategoryOverrides:  []CategoryOverride{},
		SandboxImage:       getEnv("SANDBOX_IMAGE", "golang:1.21-alpine"),
		SandboxWorkDir:     getEnv("SANDBOX_WORKDIR", "/tmp/sandbox"),
		EnableLocalFallback: getEnvBool("ENABLE_LOCAL_FALLBACK", false),
		GracePeriod:        2 * time.Second,
		CompiledBinCacheSize: 20,
	}
}

// GetLimitsForProblem returns the effective resource limits for a problem based on difficulty and category.
func (c *WorkerConfig) GetLimitsForProblem(difficulty, category string) ResourceLimits {
	// Start with difficulty-based limits
	limits, ok := c.DifficultyLimits[strings.ToLower(difficulty)]
	if !ok {
		limits = c.DefaultResourceLimits
	}

	// Apply category overrides
	for _, override := range c.CategoryOverrides {
		if strings.EqualFold(override.Category, category) {
			if override.TimeLimit > 0 {
				limits.TimeLimit = override.TimeLimit
			}
			if override.MemoryLimitMB > 0 {
				limits.MemoryLimitMB = override.MemoryLimitMB
			}
			if override.CPUQuota != "" {
				limits.CPUQuota = override.CPUQuota
			}
			if override.PIDsLimit > 0 {
				limits.PIDsLimit = override.PIDsLimit
			}
			if override.MaxOutputSize > 0 {
				limits.MaxOutputSize = override.MaxOutputSize
			}
			if override.DiskLimitMB > 0 {
				limits.DiskLimitMB = override.DiskLimitMB
			}
			break
		}
	}

	return limits
}

// IsDockerEnabled checks if Docker sandbox should be used.
func (c *WorkerConfig) IsDockerEnabled() bool {
	return getEnv("DOCKER_ENABLED", "true") == "true"
}

// String returns a human-readable summary of the worker config.
func (c *WorkerConfig) String() string {
	return fmt.Sprintf(
		"WorkerConfig{MaxConcurrent=%d, QueueCapacity=%d, WarmPool=%d, DefaultDifficulty=%s}",
		c.MaxConcurrentExecutions, c.QueueCapacity, c.WarmPoolSize, c.DefaultDifficulty,
	)
}

// getEnvBool returns the boolean value of an environment variable.
func getEnvBool(key string, fallback bool) bool {
	if v, ok := os.LookupEnv(key); ok {
		b, err := strconv.ParseBool(v)
		if err == nil {
			return b
		}
	}
	return fallback
}

// getEnvInt returns the integer value of an environment variable.
func getEnvInt(key string, fallback int) int {
	if v, ok := os.LookupEnv(key); ok {
		n, err := strconv.Atoi(v)
		if err == nil {
			return n
		}
	}
	return fallback
}

// getEnv returns the string value of an environment variable.
func getEnv(key, fallback string) string {
	if v, ok := os.LookupEnv(key); ok {
		return v
	}
	return fallback
}
