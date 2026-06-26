package database

import (
	"context"
	"fmt"
	"log"
	"os"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
)

// Config holds PostgreSQL connection configuration.
type Config struct {
	Host            string
	Port            int
	User            string
	Password        string
	Database        string
	MaxConns        int32
	MinConns        int32
	MaxConnIdleTime time.Duration
	MaxConnLifetime time.Duration
}

// Pool wraps pgxpool.Pool with additional functionality.
type Pool struct {
	*pgxpool.Pool
	config Config
}

// New creates a new PostgreSQL connection pool.
func New(cfg Config) (*Pool, error) {
	connString := fmt.Sprintf(
		"postgres://%s:%s@%s:%d/%s?sslmode=disable",
		cfg.User, cfg.Password, cfg.Host, cfg.Port, cfg.Database,
	)

	poolConfig, err := pgxpool.ParseConfig(connString)
	if err != nil {
		return nil, fmt.Errorf("failed to parse database config: %w", err)
	}

	poolConfig.MaxConns = cfg.MaxConns
	poolConfig.MinConns = cfg.MinConns
	poolConfig.MaxConnIdleTime = cfg.MaxConnIdleTime
	poolConfig.MaxConnLifetime = cfg.MaxConnLifetime
	poolConfig.HealthCheckPeriod = 5 * time.Minute

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	pgxPool, err := pgxpool.NewWithConfig(ctx, poolConfig)
	if err != nil {
		return nil, fmt.Errorf("failed to create connection pool: %w", err)
	}

	// Verify connection
	if err := pgxPool.Ping(ctx); err != nil {
		pgxPool.Close()
		return nil, fmt.Errorf("failed to ping database: %w", err)
	}

	log.Printf("Database pool connected: %s@%s:%d/%s (max=%d, min=%d)",
		cfg.User, cfg.Host, cfg.Port, cfg.Database, cfg.MaxConns, cfg.MinConns)

	return &Pool{
		Pool:   pgxPool,
		config: cfg,
	}, nil
}

// NewFromEnv creates a pool from environment variables.
func NewFromEnv() (*Pool, error) {
	cfg := Config{
		Host:            getEnv("DB_HOST", "localhost"),
		Port:            getEnvInt("DB_PORT", 5432),
		User:            getEnv("DB_USER", "postgres"),
		Password:        getEnv("DB_PASSWORD", "postgres"),
		Database:        getEnv("DB_NAME", "coding_challange"),
		MaxConns:        int32(getEnvInt("DB_MAX_CONNS", 25)),
		MinConns:        int32(getEnvInt("DB_MIN_CONNS", 5)),
		MaxConnIdleTime: time.Duration(getEnvInt("DB_MAX_IDLE_TIME_SEC", 300)) * time.Second,
		MaxConnLifetime: time.Duration(getEnvInt("DB_MAX_LIFETIME_SEC", 3600)) * time.Second,
	}
	return New(cfg)
}

// HealthCheck verifies the database connection is alive.
func (p *Pool) HealthCheck(ctx context.Context) error {
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	if err := p.Ping(ctx); err != nil {
		return fmt.Errorf("database health check failed: %w", err)
	}

	// Check pool stats
	stat := p.Stat()
	log.Printf("DB pool stats: total=%d, idle=%d, in_use=%d, constructing=%d",
		stat.TotalConns(), stat.IdleConns(), stat.AcquiredConns(), stat.ConstructingConns())

	return nil
}

// RunMigrations executes SQL migration files from the given directory.
func (p *Pool) RunMigrations(ctx context.Context, migrationsDir string) error {
	// Create migrations tracking table
	_, err := p.Exec(ctx, `
		CREATE TABLE IF NOT EXISTS schema_migrations (
			id SERIAL PRIMARY KEY,
			name VARCHAR(255) UNIQUE NOT NULL,
			applied_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
		)
	`)
	if err != nil {
		return fmt.Errorf("failed to create migrations table: %w", err)
	}

	// Read migration files
	entries, err := os.ReadDir(migrationsDir)
	if err != nil {
		return fmt.Errorf("failed to read migrations directory: %w", err)
	}

	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}

		name := entry.Name()
		if !endsWith(name, ".sql") {
			continue
		}

		// Check if already applied
		var exists bool
		err := p.QueryRow(ctx,
			"SELECT EXISTS(SELECT 1 FROM schema_migrations WHERE name = $1)", name,
		).Scan(&exists)
		if err != nil {
			return fmt.Errorf("failed to check migration %s: %w", name, err)
		}
		if exists {
			continue
		}

		// Read and execute migration
		content, err := os.ReadFile(fmt.Sprintf("%s/%s", migrationsDir, name))
		if err != nil {
			return fmt.Errorf("failed to read migration %s: %w", name, err)
		}

		tx, err := p.Begin(ctx)
		if err != nil {
			return fmt.Errorf("failed to begin transaction for migration %s: %w", name, err)
		}

		_, err = tx.Exec(ctx, string(content))
		if err != nil {
			tx.Rollback(ctx)
			return fmt.Errorf("failed to execute migration %s: %w", name, err)
		}

		_, err = tx.Exec(ctx, "INSERT INTO schema_migrations (name) VALUES ($1)", name)
		if err != nil {
			tx.Rollback(ctx)
			return fmt.Errorf("failed to record migration %s: %w", name, err)
		}

		if err := tx.Commit(ctx); err != nil {
			return fmt.Errorf("failed to commit migration %s: %w", name, err)
		}

		log.Printf("Applied migration: %s", name)
	}

	return nil
}

// Stats returns current pool statistics as a map.
func (p *Pool) Stats() map[string]interface{} {
	stat := p.Stat()
	return map[string]interface{}{
		"total_conns":      stat.TotalConns(),
		"idle_conns":       stat.IdleConns(),
		"acquired_conns":   stat.AcquiredConns(),
		"constructing_conns": stat.ConstructingConns(),
		"waiting_conns":    0, // pgxpool.Stat v5 tidak punya WaitingConns
		"max_conns":        p.config.MaxConns,
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
		var n int
		if _, err := fmt.Sscanf(v, "%d", &n); err == nil {
			return n
		}
	}
	return fallback
}

func endsWith(s, suffix string) bool {
	return len(s) >= len(suffix) && s[len(s)-len(suffix):] == suffix
}


