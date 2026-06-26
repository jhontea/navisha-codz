package migrations

import (
	"context"
	"database/sql"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"
)

// Runner handles database migration execution with versioning.
type Runner struct {
	db         *sql.DB
	migrationsDir string
	trackingTable string
}

// NewRunner creates a new migration runner.
func NewRunner(db *sql.DB, migrationsDir string) *Runner {
	return &Runner{
		db:            db,
		migrationsDir: migrationsDir,
		trackingTable: "schema_migrations",
	}
}

// NewRunnerWithTable creates a new migration runner with a custom tracking table.
func NewRunnerWithTable(db *sql.DB, migrationsDir, trackingTable string) *Runner {
	return &Runner{
		db:            db,
		migrationsDir: migrationsDir,
		trackingTable: trackingTable,
	}
}

// Init creates the migrations tracking table if it doesn't exist.
func (r *Runner) Init(ctx context.Context) error {
	query := fmt.Sprintf(`
		CREATE TABLE IF NOT EXISTS %s (
			id SERIAL PRIMARY KEY,
			name VARCHAR(255) UNIQUE NOT NULL,
			applied_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
		)
	`, r.trackingTable)

	_, err := r.db.ExecContext(ctx, query)
	if err != nil {
		return fmt.Errorf("failed to create migrations tracking table: %w", err)
	}
	return nil
}

// Up applies all pending migrations.
func (r *Runner) Up(ctx context.Context) error {
	if err := r.Init(ctx); err != nil {
		return err
	}

	pending, err := r.getPendingMigrations(ctx)
	if err != nil {
		return err
	}

	if len(pending) == 0 {
		return nil
	}

	for _, migration := range pending {
		if err := r.applyMigration(ctx, migration); err != nil {
			return fmt.Errorf("failed to apply migration %s: %w", migration.name, err)
		}
	}

	return nil
}

// UpN applies up to n pending migrations.
func (r *Runner) UpN(ctx context.Context, n int) error {
	if err := r.Init(ctx); err != nil {
		return err
	}

	pending, err := r.getPendingMigrations(ctx)
	if err != nil {
		return err
	}

	if n > len(pending) {
		n = len(pending)
	}

	for i := 0; i < n; i++ {
		if err := r.applyMigration(ctx, pending[i]); err != nil {
			return fmt.Errorf("failed to apply migration %s: %w", pending[i].name, err)
		}
	}

	return nil
}

// Down rolls back the last n migrations.
func (r *Runner) Down(ctx context.Context, n int) error {
	applied, err := r.getAppliedMigrations(ctx)
	if err != nil {
		return err
	}

	if n > len(applied) {
		n = len(applied)
	}

	for i := len(applied) - 1; i >= len(applied)-n; i-- {
		if err := r.rollbackMigration(ctx, applied[i]); err != nil {
			return fmt.Errorf("failed to rollback migration %s: %w", applied[i].name, err)
		}
	}

	return nil
}

// Status returns the current migration status.
func (r *Runner) Status(ctx context.Context) ([]MigrationStatus, error) {
	if err := r.Init(ctx); err != nil {
		return nil, err
	}

	entries, err := os.ReadDir(r.migrationsDir)
	if err != nil {
		return nil, fmt.Errorf("failed to read migrations directory: %w", err)
	}

	applied, err := r.getAppliedMap(ctx)
	if err != nil {
		return nil, err
	}

	var statuses []MigrationStatus
	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}
		name := entry.Name()
		if !strings.HasSuffix(name, ".up.sql") {
			continue
		}

		baseName := strings.TrimSuffix(name, ".up.sql")
		status := MigrationStatus{
			Name:      baseName,
			Pending:   true,
		}

		if appliedTime, ok := applied[name]; ok {
			status.Pending = false
			status.AppliedAt = &appliedTime
		}

		statuses = append(statuses, status)
	}

	sort.Slice(statuses, func(i, j int) bool {
		return statuses[i].Name < statuses[j].Name
	})

	return statuses, nil
}

// Reset rolls back all migrations and re-applies them.
func (r *Runner) Reset(ctx context.Context) error {
	applied, err := r.getAppliedMigrations(ctx)
	if err != nil {
		return err
	}

	// Rollback all in reverse order
	for i := len(applied) - 1; i >= 0; i-- {
		if err := r.rollbackMigration(ctx, applied[i]); err != nil {
			return fmt.Errorf("failed to rollback migration %s: %w", applied[i].name, err)
		}
	}

	return r.Up(ctx)
}

// MigrationInfo holds information about a migration file.
type MigrationInfo struct {
	name     string
	filename string
	path     string
}

// MigrationStatus represents the status of a migration.
type MigrationStatus struct {
	Name      string
	Pending   bool
	AppliedAt *time.Time
}

func (r *Runner) getPendingMigrations(ctx context.Context) ([]MigrationInfo, error) {
	entries, err := os.ReadDir(r.migrationsDir)
	if err != nil {
		return nil, fmt.Errorf("failed to read migrations directory: %w", err)
	}

	applied, err := r.getAppliedMap(ctx)
	if err != nil {
		return nil, err
	}

	var pending []MigrationInfo
	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}
		name := entry.Name()
		if !strings.HasSuffix(name, ".up.sql") {
			continue
		}

		if _, ok := applied[name]; ok {
			continue
		}

		pending = append(pending, MigrationInfo{
			name:     strings.TrimSuffix(name, ".up.sql"),
			filename: name,
			path:     filepath.Join(r.migrationsDir, name),
		})
	}

	sort.Slice(pending, func(i, j int) bool {
		return pending[i].name < pending[j].name
	})

	return pending, nil
}

func (r *Runner) getAppliedMigrations(ctx context.Context) ([]MigrationInfo, error) {
	query := fmt.Sprintf("SELECT name FROM %s ORDER BY name DESC", r.trackingTable)
	rows, err := r.db.QueryContext(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("failed to query applied migrations: %w", err)
	}
	defer rows.Close()

	var applied []MigrationInfo
	for rows.Next() {
		var name string
		if err := rows.Scan(&name); err != nil {
			return nil, fmt.Errorf("failed to scan migration name: %w", err)
		}
		applied = append(applied, MigrationInfo{name: strings.TrimSuffix(name, ".up.sql")})
	}

	return applied, rows.Err()
}

func (r *Runner) getAppliedMap(ctx context.Context) (map[string]time.Time, error) {
	query := fmt.Sprintf("SELECT name, applied_at FROM %s", r.trackingTable)
	rows, err := r.db.QueryContext(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("failed to query applied migrations: %w", err)
	}
	defer rows.Close()

	result := make(map[string]time.Time)
	for rows.Next() {
		var name string
		var appliedAt time.Time
		if err := rows.Scan(&name, &appliedAt); err != nil {
			return nil, fmt.Errorf("failed to scan migration record: %w", err)
		}
		result[name] = appliedAt
	}

	return result, rows.Err()
}

func (r *Runner) applyMigration(ctx context.Context, migration MigrationInfo) error {
	content, err := os.ReadFile(migration.path)
	if err != nil {
		return fmt.Errorf("failed to read migration file: %w", err)
	}

	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback()

	if _, err := tx.ExecContext(ctx, string(content)); err != nil {
		return fmt.Errorf("failed to execute migration: %w", err)
	}

	query := fmt.Sprintf("INSERT INTO %s (name) VALUES ($1)", r.trackingTable)
	if _, err := tx.ExecContext(ctx, query, migration.filename); err != nil {
		return fmt.Errorf("failed to record migration: %w", err)
	}

	if err := tx.Commit(); err != nil {
		return fmt.Errorf("failed to commit migration: %w", err)
	}

	return nil
}

func (r *Runner) rollbackMigration(ctx context.Context, migration MigrationInfo) error {
	downFile := strings.TrimSuffix(migration.name, ".up.sql") + ".down.sql"
	downPath := filepath.Join(r.migrationsDir, downFile)

	content, err := os.ReadFile(downPath)
	if err != nil {
		// If no down migration, just remove the tracking record
		query := fmt.Sprintf("DELETE FROM %s WHERE name = $1", r.trackingTable)
		_, err = r.db.ExecContext(ctx, query, migration.name+".up.sql")
		if err != nil {
			return fmt.Errorf("failed to remove migration record: %w", err)
		}
		return nil
	}

	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback()

	if _, err := tx.ExecContext(ctx, string(content)); err != nil {
		return fmt.Errorf("failed to execute rollback: %w", err)
	}

	query := fmt.Sprintf("DELETE FROM %s WHERE name = $1", r.trackingTable)
	if _, err := tx.ExecContext(ctx, query, migration.name+".up.sql"); err != nil {
		return fmt.Errorf("failed to remove migration record: %w", err)
	}

	if err := tx.Commit(); err != nil {
		return fmt.Errorf("failed to commit rollback: %w", err)
	}

	return nil
}

// FromFS loads migrations from an fs.FS (useful for embedded migrations).
func (r *Runner) FromFS(ctx context.Context, fsys fs.FS) error {
	if err := r.Init(ctx); err != nil {
		return err
	}

	entries, err := fs.ReadDir(fsys, ".")
	if err != nil {
		return fmt.Errorf("failed to read embedded migrations: %w", err)
	}

	applied, err := r.getAppliedMap(ctx)
	if err != nil {
		return err
	}

	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}
		name := entry.Name()
		if !strings.HasSuffix(name, ".up.sql") {
			continue
		}

		if _, ok := applied[name]; ok {
			continue
		}

		content, err := fs.ReadFile(fsys, name)
		if err != nil {
			return fmt.Errorf("failed to read migration %s: %w", name, err)
		}

		tx, err := r.db.BeginTx(ctx, nil)
		if err != nil {
			return fmt.Errorf("failed to begin transaction: %w", err)
		}

		if _, err := tx.ExecContext(ctx, string(content)); err != nil {
			tx.Rollback()
			return fmt.Errorf("failed to execute migration %s: %w", name, err)
		}

		query := fmt.Sprintf("INSERT INTO %s (name) VALUES ($1)", r.trackingTable)
		if _, err := tx.ExecContext(ctx, query, name); err != nil {
			tx.Rollback()
			return fmt.Errorf("failed to record migration %s: %w", name, err)
		}

		if err := tx.Commit(); err != nil {
			return fmt.Errorf("failed to commit migration %s: %w", name, err)
		}
	}

	return nil
}
