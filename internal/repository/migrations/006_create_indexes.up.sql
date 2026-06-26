-- Migration: 006_create_indexes
-- Description: Additional performance indexes
-- Created: 2024-01-01

-- Composite indexes for common query patterns
CREATE INDEX IF NOT EXISTS idx_submissions_user_status ON submissions (user_id, status);
CREATE INDEX IF NOT EXISTS idx_submissions_problem_status ON submissions (problem_id, status);
CREATE INDEX IF NOT EXISTS idx_submissions_user_problem_status ON submissions (user_id, problem_id, status);

-- Partial index for active/pending submissions
CREATE INDEX IF NOT EXISTS idx_submissions_pending ON submissions (submitted_at) 
    WHERE status IN ('pending', 'queued', 'running');

-- Index for user search
CREATE INDEX IF NOT EXISTS idx_users_username_trgm ON users USING gin (username gin_trgm_ops);
CREATE INDEX IF NOT EXISTS idx_users_email_trgm ON users USING gin (email gin_trgm_ops);

-- Index for problem search
CREATE INDEX IF NOT EXISTS idx_problems_title_trgm ON problems USING gin (title gin_trgm_ops);

-- Index for leaderboard lookups
CREATE INDEX IF NOT EXISTS idx_leaderboard_user_problem_score ON leaderboard (user_id, problem_id, best_score DESC);

-- Comments on indexes
COMMENT ON INDEX idx_submissions_pending IS 'Optimizes polling for pending submissions';
COMMENT ON INDEX idx_global_leaderboard_score IS 'Optimizes leaderboard ranking queries';
