package config

import (
	"fmt"
	"os"
	"strconv"
	"strings"
	"sync"
	"time"

	"gopkg.in/yaml.v3"
)

// Config holds the complete application configuration.
type Config struct {
	Server   ServerConfig   `yaml:"server" validate:"required"`
	Database DatabaseConfig `yaml:"database" validate:"required"`
	Redis    RedisConfig    `yaml:"redis"`
	RabbitMQ RabbitMQConfig `yaml:"rabbitmq"`
	JWT      JWTConfig      `yaml:"jwt" validate:"required"`
	Log      LogConfig      `yaml:"log"`
}

// ServerConfig holds HTTP server configuration.
type ServerConfig struct {
	Port            int           `yaml:"port"`
	Mode            string        `yaml:"mode"`
	ReadTimeout     time.Duration `yaml:"read_timeout"`
	WriteTimeout    time.Duration `yaml:"write_timeout"`
	ShutdownTimeout time.Duration `yaml:"shutdown_timeout"`
	AllowedOrigins  []string      `yaml:"allowed_origins"`
}

// DatabaseConfig holds database configuration.
type DatabaseConfig struct {
	Host            string        `yaml:"host"`
	Port            int           `yaml:"port"`
	User            string        `yaml:"user"`
	Password        string        `yaml:"password"`
	Database        string        `yaml:"database"`
	MaxConns        int32         `yaml:"max_conns"`
	MinConns        int32         `yaml:"min_conns"`
	MaxConnIdleTime time.Duration `yaml:"max_conn_idle_time"`
	MaxConnLifetime time.Duration `yaml:"max_conn_lifetime"`
}

// RedisConfig holds Redis configuration.
type RedisConfig struct {
	Addr     string `yaml:"addr"`
	Password string `yaml:"password"`
	DB       int    `yaml:"db"`
	PoolSize int    `yaml:"pool_size"`
}

// RabbitMQConfig holds RabbitMQ configuration.
type RabbitMQConfig struct {
	Host     string `yaml:"host"`
	Port     int    `yaml:"port"`
	User     string `yaml:"user"`
	Password string `yaml:"password"`
	VHost    string `yaml:"vhost"`
}

// JWTConfig holds JWT configuration.
type JWTConfig struct {
	AccessTokenSecret  string        `yaml:"access_token_secret" validate:"required"`
	RefreshTokenSecret string        `yaml:"refresh_token_secret" validate:"required"`
	AccessTokenTTL     time.Duration `yaml:"access_token_ttl"`
	RefreshTokenTTL    time.Duration `yaml:"refresh_token_ttl"`
	Issuer             string        `yaml:"issuer"`
}

// LogConfig holds logging configuration.
type LogConfig struct {
	Level  string `yaml:"level"`
	Format string `yaml:"format"`
}

// Manager manages configuration with hot reload support.
type Manager struct {
	mu       sync.RWMutex
	config   Config
	filePath string
	modTime  time.Time
}

// NewManager creates a new configuration manager.
func NewManager(configPath string) (*Manager, error) {
	m := &Manager{filePath: configPath}

	if err := m.load(); err != nil {
		return nil, fmt.Errorf("failed to load config: %w", err)
	}

	return m, nil
}

// load reads the config file and environment overrides.
func (m *Manager) load() error {
	data, err := os.ReadFile(m.filePath)
	if err != nil {
		return fmt.Errorf("failed to read config file %s: %w", m.filePath, err)
	}

	var cfg Config
	if err := yaml.Unmarshal(data, &cfg); err != nil {
		return fmt.Errorf("failed to parse config YAML: %w", err)
	}

	// Apply environment variable overrides
	applyEnvOverrides(&cfg)

	// Validate
	if err := validate(&cfg); err != nil {
		return fmt.Errorf("config validation failed: %w", err)
	}

	m.mu.Lock()
	m.config = cfg
	m.modTime = time.Now()
	m.mu.Unlock()

	return nil
}

// Get returns the current configuration (thread-safe).
func (m *Manager) Get() Config {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.config
}

// Reload reloads the configuration from file.
func (m *Manager) Reload() error {
	return m.load()
}

// Watch starts watching the config file for changes.
func (m *Manager) Watch(interval time.Duration, onChange func()) {
	go func() {
		ticker := time.NewTicker(interval)
		defer ticker.Stop()

		for range ticker.C {
			info, err := os.Stat(m.filePath)
			if err != nil {
				continue
			}

			m.mu.RLock()
			modTime := m.modTime
			m.mu.RUnlock()

			if info.ModTime().After(modTime) {
				if err := m.Reload(); err == nil && onChange != nil {
					onChange()
				}
			}
		}
	}()
}

// applyEnvOverrides applies environment variable overrides to the config.
func applyEnvOverrides(cfg *Config) {
	// Server overrides
	if v := os.Getenv("PORT"); v != "" {
		if port, err := strconv.Atoi(v); err == nil {
			cfg.Server.Port = port
		}
	}
	if v := os.Getenv("GIN_MODE"); v != "" {
		cfg.Server.Mode = v
	}

	// Database overrides
	if v := os.Getenv("DB_HOST"); v != "" {
		cfg.Database.Host = v
	}
	if v := os.Getenv("DB_PORT"); v != "" {
		if port, err := strconv.Atoi(v); err == nil {
			cfg.Database.Port = port
		}
	}
	if v := os.Getenv("DB_USER"); v != "" {
		cfg.Database.User = v
	}
	if v := os.Getenv("DB_PASSWORD"); v != "" {
		cfg.Database.Password = v
	}
	if v := os.Getenv("DB_NAME"); v != "" {
		cfg.Database.Database = v
	}

	// Redis overrides
	if v := os.Getenv("REDIS_ADDR"); v != "" {
		cfg.Redis.Addr = v
	}
	if v := os.Getenv("REDIS_PASSWORD"); v != "" {
		cfg.Redis.Password = v
	}

	// RabbitMQ overrides
	if v := os.Getenv("RABBITMQ_HOST"); v != "" {
		cfg.RabbitMQ.Host = v
	}
	if v := os.Getenv("RABBITMQ_USER"); v != "" {
		cfg.RabbitMQ.User = v
	}
	if v := os.Getenv("RABBITMQ_PASSWORD"); v != "" {
		cfg.RabbitMQ.Password = v
	}

	// JWT overrides
	if v := os.Getenv("JWT_ACCESS_SECRET"); v != "" {
		cfg.JWT.AccessTokenSecret = v
	}
	if v := os.Getenv("JWT_REFRESH_SECRET"); v != "" {
		cfg.JWT.RefreshTokenSecret = v
	}
	if v := os.Getenv("JWT_ISSUER"); v != "" {
		cfg.JWT.Issuer = v
	}

	// Log overrides
	if v := os.Getenv("LOG_LEVEL"); v != "" {
		cfg.Log.Level = v
	}
}

// validate performs basic validation on the config.
func validate(cfg *Config) error {
	var errs []string

	if cfg.Server.Port < 1 || cfg.Server.Port > 65535 {
		errs = append(errs, "server.port must be between 1 and 65535")
	}
	if cfg.Database.Host == "" {
		errs = append(errs, "database.host is required")
	}
	if cfg.Database.Database == "" {
		errs = append(errs, "database.database is required")
	}
	if cfg.JWT.AccessTokenSecret == "" {
		errs = append(errs, "jwt.access_token_secret is required")
	}
	if cfg.JWT.RefreshTokenSecret == "" {
		errs = append(errs, "jwt.refresh_token_secret is required")
	}

	if len(errs) > 0 {
		return fmt.Errorf("validation errors: %s", strings.Join(errs, "; "))
	}
	return nil
}

// DefaultConfig returns a default configuration for development.
func DefaultConfig() Config {
	return Config{
		Server: ServerConfig{
			Port:            9100,
			Mode:            "debug",
			ReadTimeout:     30 * time.Second,
			WriteTimeout:    30 * time.Second,
			ShutdownTimeout: 15 * time.Second,
			AllowedOrigins:  []string{"*"},
		},
		Database: DatabaseConfig{
			Host:            "localhost",
			Port:            5432,
			User:            "postgres",
			Password:        "postgres",
			Database:        "coding_challange",
			MaxConns:        25,
			MinConns:        5,
			MaxConnIdleTime: 5 * time.Minute,
			MaxConnLifetime: 1 * time.Hour,
		},
		Redis: RedisConfig{
			Addr:     "localhost:6379",
			Password: "",
			DB:       0,
			PoolSize: 10,
		},
		RabbitMQ: RabbitMQConfig{
			Host:     "localhost",
			Port:     5672,
			User:     "guest",
			Password: "guest",
			VHost:    "/",
		},
		JWT: JWTConfig{
			AccessTokenSecret:  "dev-access-secret-key",
			RefreshTokenSecret: "dev-refresh-secret-key",
			AccessTokenTTL:     15 * time.Minute,
			RefreshTokenTTL:    7 * 24 * time.Hour,
			Issuer:             "coding-challange",
		},
		Log: LogConfig{
			Level:  "info",
			Format: "json",
		},
	}
}
