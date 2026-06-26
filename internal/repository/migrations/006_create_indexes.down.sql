-- Migration: 006_create_indexes (rollback)
DROP INDEX IF EXISTS idx_leaderboard_user_problem_score;
DROP INDEX IF EXISTS idx_problems_title_trgm;
DROP INDEX IF EXISTS idx_users_email_trgm;
DROP INDEX IF EXISTS idx_users_username_trgm;
DROP INDEX IF EXISTS idx_submissions_pending;
DROP INDEX IF EXISTS idx_submissions_user_problem_status;
DROP INDEX IF EXISTS idx_submissions_problem_status;
DROP INDEX IF EXISTS idx_submissions_user_status;
