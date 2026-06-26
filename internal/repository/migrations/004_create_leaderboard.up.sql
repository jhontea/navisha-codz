-- Migration: 004_create_leaderboard
-- Description: Create leaderboard table and materialized view
-- Created: 2024-01-01

CREATE TABLE IF NOT EXISTS leaderboard (
    id SERIAL PRIMARY KEY,
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    problem_id INTEGER NOT NULL REFERENCES problems(id) ON DELETE CASCADE,
    best_score INTEGER NOT NULL DEFAULT 0,
    best_execution_time_ms INTEGER,
    best_memory_used_kb INTEGER,
    submissions_count INTEGER NOT NULL DEFAULT 0,
    last_submission_at TIMESTAMPTZ,
    UNIQUE(user_id, problem_id)
);

CREATE INDEX idx_leaderboard_user_id ON leaderboard (user_id);
CREATE INDEX idx_leaderboard_problem_id ON leaderboard (problem_id);
CREATE INDEX idx_leaderboard_best_score ON leaderboard (best_score DESC);

-- Global leaderboard materialized view
CREATE MATERIALIZED VIEW IF NOT EXISTS global_leaderboard AS
SELECT 
    u.id AS user_id,
    u.username,
    u.display_name,
    u.avatar_url,
    COUNT(DISTINCT l.problem_id) AS problems_solved,
    SUM(l.best_score) AS total_score,
    SUM(l.submissions_count) AS total_submissions,
    MAX(l.last_submission_at) AS last_active
FROM users u
JOIN leaderboard l ON u.id = l.user_id
WHERE u.is_active = TRUE AND l.best_score > 0
GROUP BY u.id, u.username, u.display_name, u.avatar_url
ORDER BY total_score DESC, problems_solved DESC;

CREATE UNIQUE INDEX idx_global_leaderboard_user ON global_leaderboard (user_id);
CREATE INDEX idx_global_leaderboard_score ON global_leaderboard (total_score DESC);

-- Function to refresh leaderboard
CREATE OR REPLACE FUNCTION refresh_leaderboard()
RETURNS void AS $$
BEGIN
    REFRESH MATERIALIZED VIEW CONCURRENTLY global_leaderboard;
END;
$$ LANGUAGE plpgsql;
