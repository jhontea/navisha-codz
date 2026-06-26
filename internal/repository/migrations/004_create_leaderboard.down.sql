-- Migration: 004_create_leaderboard (rollback)
DROP FUNCTION IF EXISTS refresh_leaderboard();
DROP MATERIALIZED VIEW IF EXISTS global_leaderboard;
DROP INDEX IF EXISTS idx_leaderboard_best_score;
DROP INDEX IF EXISTS idx_leaderboard_problem_id;
DROP INDEX IF EXISTS idx_leaderboard_user_id;
DROP TABLE IF EXISTS leaderboard;
