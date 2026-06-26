package repository

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"strings"
	"sync"
	"time"

	"coding-challange/internal/model"
)

// ============================================================================
// Statement Cache — prepared statements for frequently used queries
// ============================================================================

// PreparedStatements holds frequently used prepared statements.
type PreparedStatements struct {
	mu         sync.RWMutex
	statements map[string]*sql.Stmt
	db         *sql.DB
}

// NewPreparedStatements creates a new prepared statement cache.
func NewPreparedStatements(db *sql.DB) *PreparedStatements {
	return &PreparedStatements{
		statements: make(map[string]*sql.Stmt),
		db:         db,
	}
}

// Get retrieves or creates a prepared statement by name.
func (ps *PreparedStatements) Get(name string, query string) (*sql.Stmt, error) {
	ps.mu.RLock()
	if stmt, ok := ps.statements[name]; ok {
		ps.mu.RUnlock()
		return stmt, nil
	}
	ps.mu.RUnlock()

	stmt, err := ps.db.Prepare(query)
	if err != nil {
		return nil, fmt.Errorf("failed to prepare statement %s: %w", name, err)
	}

	ps.mu.Lock()
	ps.statements[name] = stmt
	ps.mu.Unlock()

	return stmt, nil
}

// Close closes all prepared statements.
func (ps *PreparedStatements) Close() {
	ps.mu.Lock()
	defer ps.mu.Unlock()
	for _, stmt := range ps.statements {
		stmt.Close()
	}
}

// ============================================================================
// Slow Query Logger
// ============================================================================

// SlowQueryLogger logs queries that exceed a threshold duration.
type SlowQueryLogger struct {
	threshold time.Duration
	mu        sync.Mutex
	queries   []SlowQueryEntry
}

// SlowQueryEntry represents a logged slow query.
type SlowQueryEntry struct {
	Query    string        `json:"query"`
	Duration time.Duration `json:"duration"`
	Args     []interface{} `json:"args,omitempty"`
	Time     time.Time     `json:"time"`
}

// NewSlowQueryLogger creates a new slow query logger.
func NewSlowQueryLogger(threshold time.Duration) *SlowQueryLogger {
	return &SlowQueryLogger{
		threshold: threshold,
		queries:   make([]SlowQueryEntry, 0),
	}
}

// Log logs a query if it exceeds the threshold.
func (sql *SlowQueryLogger) Log(query string, duration time.Duration, args ...interface{}) {
	if duration < sql.threshold {
		return
	}

	entry := SlowQueryEntry{
		Query:    query,
		Duration: duration,
		Args:     args,
		Time:     time.Now(),
	}

	sql.mu.Lock()
	sql.queries = append(sql.queries, entry)
	if len(sql.queries) > 1000 {
		sql.queries = sql.queries[len(sql.queries)-1000:]
	}
	sql.mu.Unlock()

	log.Printf("[SLOW QUERY] %v — %s (args: %v)", duration, truncate(query, 200), args)
}

// GetQueries returns a copy of slow query entries.
func (sql *SlowQueryLogger) GetQueries() []SlowQueryEntry {
	sql.mu.Lock()
	defer sql.mu.Unlock()
	result := make([]SlowQueryEntry, len(sql.queries))
	copy(result, sql.queries)
	return result
}

func truncate(s string, n int) string {
	if len(s) <= n {
		return s
	}
	return s[:n] + "..."
}

// ============================================================================
// Cursor-Based Pagination
// ============================================================================

// CursorPaginator implements keyset/cursor-based pagination.
// This is significantly faster than OFFSET-based pagination for large datasets.
type CursorPaginator struct {
	db                *sql.DB
	baseQuery         string
	countQuery        string
	defaultPageSize   int
	maxPageSize       int
	orderColumn       string
}

// NewCursorPaginator creates a new cursor-based paginator.
func NewCursorPaginator(db *sql.DB, baseQuery, countQuery, orderColumn string, defaultPageSize, maxPageSize int) *CursorPaginator {
	if defaultPageSize <= 0 {
		defaultPageSize = 20
	}
	if maxPageSize <= 0 {
		maxPageSize = 100
	}
	return &CursorPaginator{
		db:              db,
		baseQuery:       baseQuery,
		countQuery:      countQuery,
		defaultPageSize: defaultPageSize,
		maxPageSize:     maxPageSize,
		orderColumn:      orderColumn,
	}
}

// PageResult contains pagination results.
type PageResult struct {
	Items      []map[string]interface{} `json:"items"`
	NextCursor string                   `json:"next_cursor,omitempty"`
	PrevCursor string                   `json:"prev_cursor,omitempty"`
	HasMore    bool                     `json:"has_more"`
	TotalCount int                      `json:"total_count"`
}

// FetchPage fetches a page using cursor-based pagination.
func (cp *CursorPaginator) FetchPage(ctx context.Context, pageSize int, cursor string, args ...interface{}) (*PageResult, error) {
	if pageSize <= 0 {
		pageSize = cp.defaultPageSize
	}
	if pageSize > cp.maxPageSize {
		pageSize = cp.maxPageSize
	}

	query := cp.baseQuery
	queryArgs := make([]interface{}, 0, len(args)+2)
	queryArgs = append(queryArgs, args...)

	if cursor != "" {
		// Keyset pagination: WHERE id > cursor
		if !strings.Contains(strings.ToUpper(query), "WHERE") {
			query += " WHERE " + cp.orderColumn + " > $1"
		} else {
			query += " AND " + cp.orderColumn + " > $1"
		}
		queryArgs = append(queryArgs, cursor)
	}

	query += fmt.Sprintf(" ORDER BY %s ASC LIMIT $%d", cp.orderColumn, len(queryArgs)+1)
	queryArgs = append(queryArgs, pageSize+1) // Fetch one extra to check if there are more

	startTime := time.Now()
	rows, err := cp.db.QueryContext(ctx, query, queryArgs...)
	if err != nil {
		return nil, fmt.Errorf("cursor pagination query failed: %w", err)
	}
	defer rows.Close()

	columns, _ := rows.Columns()
	items := make([]map[string]interface{}, 0, pageSize)
	var lastCursor string

	for rows.Next() {
		if err := rows.Err(); err != nil {
			return nil, fmt.Errorf("rows iteration error: %w", err)
		}
		values := make([]interface{}, len(columns))
		valuePtrs := make([]interface{}, len(columns))
		for i := range values {
			valuePtrs[i] = &values[i]
		}
		if err := rows.Scan(valuePtrs...); err != nil {
			return nil, fmt.Errorf("row scan error: %w", err)
		}

		row := make(map[string]interface{})
		for i, col := range columns {
			if b, ok := values[i].([]byte); ok {
				row[col] = string(b)
			} else {
				row[col] = values[i]
			}
		}
		items = append(items, row)
		lastCursor = fmt.Sprintf("%v", row[cp.orderColumn])
	}

	result := &PageResult{
		Items:   items[:min(len(items), pageSize)],
		HasMore: len(items) > pageSize,
	}

	if result.HasMore {
		result.NextCursor = lastCursor
	}

	// Get total count
	var totalCount int
	countStart := time.Now()
	if cp.countQuery != "" {
		_ = cp.db.QueryRowContext(ctx, cp.countQuery, args...).Scan(&totalCount)
	}
	_ = countStart
	result.TotalCount = totalCount

	log.Printf("[CURSOR PAGE] fetched %d rows, has_more=%t (%v)", len(result.Items), result.HasMore, time.Since(startTime))
	return result, nil
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

// ============================================================================
// Batch Operations
// ============================================================================

// BatchInserter handles batch inserts for efficiency.
type BatchInserter struct {
	db         *sql.DB
	tableName  string
	columns    []string
	batchSize  int
	buffer     []map[string]interface{}
	mu         sync.Mutex
	totalCount int64
}

// NewBatchInserter creates a new batch inserter.
func NewBatchInserter(db *sql.DB, tableName string, columns []string, batchSize int) *BatchInserter {
	if batchSize <= 0 {
		batchSize = 100
	}
	return &BatchInserter{
		db:        db,
		tableName: tableName,
		columns:   columns,
		batchSize: batchSize,
		buffer:    make([]map[string]interface{}, 0, batchSize),
	}
}

// Add adds a row to the batch buffer.
func (bi *BatchInserter) Add(row map[string]interface{}) {
	bi.mu.Lock()
	bi.buffer = append(bi.buffer, row)
	if len(bi.buffer) >= bi.batchSize {
		bi.flush()
	}
	bi.mu.Unlock()
}

// Flush flushes the remaining rows in the buffer.
func (bi *BatchInserter) Flush(ctx context.Context) error {
	bi.mu.Lock()
	defer bi.mu.Unlock()
	if len(bi.buffer) == 0 {
		return nil
	}
	return bi.flushCtx(ctx)
}

func (bi *BatchInserter) flush() {
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
	bi.flushCtx(ctx)
}

func (bi *BatchInserter) flushCtx(ctx context.Context) error {
	if len(bi.buffer) == 0 {
		return nil
	}

	// Build multi-row INSERT
	valuePlaceholders := make([]string, 0, len(bi.buffer))
	args := make([]interface{}, 0)
	argIdx := 0

	for _, row := range bi.buffer {
		placeholders := make([]string, 0, len(bi.columns))
		for _, col := range bi.columns {
			argIdx++
			placeholders = append(placeholders, fmt.Sprintf("$%d", argIdx))
			args = append(args, row[col])
		}
		valuePlaceholders = append(valuePlaceholders, "("+strings.Join(placeholders, ", ")+")")
	}

	query := fmt.Sprintf("INSERT INTO %s (%s) VALUES %s",
		bi.tableName,
		strings.Join(bi.columns, ", "),
		strings.Join(valuePlaceholders, ", "))

	startTime := time.Now()
	_, err := bi.db.ExecContext(ctx, query, args...)
	if err != nil {
		return fmt.Errorf("batch insert failed: %w", err)
	}

	bi.totalCount += int64(len(bi.buffer))
	bi.buffer = bi.buffer[:0]
	log.Printf("[BATCH INSERT] %d rows into %s (%v, total: %d)", argIdx/len(bi.columns), bi.tableName, time.Since(startTime), bi.totalCount)
	return nil
}

// TotalCount returns the total number of rows inserted.
func (bi *BatchInserter) TotalCount() int64 {
	bi.mu.Lock()
	defer bi.mu.Unlock()
	return bi.totalCount
}

// ============================================================================
// Batch Test Case Inserter (specific use case)
// ============================================================================

// BatchTestCaseInsert inserts multiple test cases for a problem efficiently.
func BatchTestCaseInsert(ctx context.Context, db *sql.DB, problemID int, testCases []model.TestCase) error {
	if len(testCases) == 0 {
		return nil
	}

	valuePlaceholders := make([]string, 0, len(testCases))
	args := make([]interface{}, 0)
	argIdx := 0

	for _, tc := range testCases {
	 placeholders := make([]string, 4)
		for i := 0; i < 4; i++ {
			argIdx++
			placeholders[i] = fmt.Sprintf("$%d", argIdx)
		}
		valuePlaceholders = append(valuePlaceholders, "("+strings.Join(placeholders, ", ")+")")

		inputStr, _ := json.Marshal(tc.Input)
		expectedStr, _ := json.Marshal(tc.Expected)
		args = append(args, problemID, string(inputStr), string(expectedStr), tc.Description)
	}

	query := fmt.Sprintf(
		"INSERT INTO test_cases (problem_id, input, expected_output, description) VALUES %s",
		strings.Join(valuePlaceholders, ", "),
	)

	_, err := db.ExecContext(ctx, query, args...)
	if err != nil {
		return fmt.Errorf("batch test case insert failed: %w", err)
	}

	log.Printf("[BATCH TEST CASES] inserted %d test cases for problem %d", len(testCases), problemID)
	return nil
}

// ============================================================================
// Materialized View for Leaderboard
// ============================================================================

const (
	// CreateLeaderboardMV creates the materialized view for leaderboard rankings.
	CreateLeaderboardMV = `
		CREATE MATERIALIZED VIEW IF NOT EXISTS mv_leaderboard AS
		SELECT
			u.id AS user_id,
			u.username,
			u.full_name,
			u.avatar_url,
			u.rating,
			COUNT(DISTINCT s.problem_id) AS solved_count,
			COALESCE(SUM(s.score), 0) AS total_score,
			u.total_submissions,
			ROW_NUMBER() OVER (ORDER BY u.rating DESC, COUNT(DISTINCT s.problem_id) DESC) AS rank
		FROM users u
		LEFT JOIN submissions s ON u.id = s.user_id AND s.status = 'accepted'
		GROUP BY u.id, u.username, u.full_name, u.avatar_url, u.rating, u.total_submissions
		ORDER BY u.rating DESC, solved_count DESC
	`

	// CreateLeaderboardMVIndex creates indexes on the materialized view.
	CreateLeaderboardMVIndex = `
		CREATE UNIQUE INDEX IF NOT EXISTS idx_mv_leaderboard_user_id ON mv_leaderboard(user_id);
		CREATE INDEX IF NOT EXISTS idx_mv_leaderboard_rank ON mv_leaderboard(rank);
		CREATE INDEX IF NOT EXISTS idx_mv_leaderboard_rating ON mv_leaderboard(rating DESC)
	`

	// RefreshLeaderboardMV refreshes the leaderboard materialized view.
	RefreshLeaderboardMV = `REFRESH MATERIALIZED VIEW CONCURRENTLY mv_leaderboard`

	// QueryLeaderboard queries from the materialized view.
	QueryLeaderboard = `SELECT * FROM mv_leaderboard ORDER BY rank ASC LIMIT $1 OFFSET $2`

	// QueryLeaderboardUser gets a specific user's leaderboard entry.
	QueryLeaderboardUser = `SELECT * FROM mv_leaderboard WHERE user_id = $1`

	// QueryLeaderboardRank gets a user's rank from the materialized view.
	QueryLeaderboardRank = `SELECT rank FROM mv_leaderboard WHERE user_id = $1`
)

// EnsureLeaderboardMV creates and indexes the leaderboard materialized view.
func EnsureLeaderboardMV(ctx context.Context, db *sql.DB) error {
	_, err := db.ExecContext(ctx, CreateLeaderboardMV)
	if err != nil {
		return fmt.Errorf("failed to create leaderboard materialized view: %w", err)
	}

	_, err = db.ExecContext(ctx, CreateLeaderboardMVIndex)
	if err != nil {
		return fmt.Errorf("failed to create leaderboard MV indexes: %w", err)
	}

	log.Println("[LEADERBOARD MV] materialized view ensured")
	return nil
}

// RefreshLeaderboardMaterializedView refreshes the materialized view.
func RefreshLeaderboardMaterializedView(ctx context.Context, db *sql.DB) error {
	start := time.Now()
	_, err := db.ExecContext(ctx, RefreshLeaderboardMV)
	if err != nil {
		return fmt.Errorf("failed to refresh leaderboard MV: %w", err)
	}
	log.Printf("[LEADERBOARD MV] refreshed in %v", time.Since(start))
	return nil
}

// ============================================================================
// Query Timeout Enforcement
// ============================================================================

// ExecuteWithTimeout executes a query with a timeout.
func ExecuteWithTimeout(ctx context.Context, db *sql.DB, timeout time.Duration, query string, args ...interface{}) (sql.Result, error) {
	ctx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()

	start := time.Now()
	result, err := db.ExecContext(ctx, query, args...)
	if err != nil {
		if ctx.Err() == context.DeadlineExceeded {
			return nil, fmt.Errorf("query timeout exceeded (%v): %w", timeout, err)
		}
		return nil, err
	}
	log.Printf("[QUERY] executed in %v: %s", time.Since(start), truncate(query, 100))
	return result, nil
}

// QueryRowWithTimeout executes a query row with a timeout.
func QueryRowWithTimeout(ctx context.Context, db *sql.DB, timeout time.Duration, query string, args ...interface{}) (*sql.Row, error) {
	ctx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()

	return db.QueryRowContext(ctx, query, args...), nil
}

// ============================================================================
// Repository Optimizations — wraps all optimizations
// ============================================================================

// Optimizations holds all database optimization utilities.
type Optimizations struct {
	PreparedStatements *PreparedStatements
	SlowQueryLogger    *SlowQueryLogger
	CursorPaginator    *CursorPaginator
	BatchInserter      *BatchInserter
	db                 *sql.DB
}

// NewOptimizations creates a new optimizations wrapper.
func NewOptimizations(db *sql.DB, slowQueryThreshold time.Duration) *Optimizations {
	return &Optimizations{
		PreparedStatements: NewPreparedStatements(db),
		SlowQueryLogger:    NewSlowQueryLogger(slowQueryThreshold),
		db:                 db,
	}
}

// SetCursorPaginator sets the cursor paginator.
func (o *Optimizations) SetCursorPaginator(baseQuery, countQuery, orderColumn string, defaultPageSize, maxPageSize int) {
	o.CursorPaginator = NewCursorPaginator(o.db, baseQuery, countQuery, orderColumn, defaultPageSize, maxPageSize)
}

// GetSlowQueries returns logged slow queries.
func (o *Optimizations) GetSlowQueries() []SlowQueryEntry {
	return o.SlowQueryLogger.GetQueries()
}

// Close closes all optimization resources.
func (o *Optimizations) Close() {
	o.PreparedStatements.Close()
}
