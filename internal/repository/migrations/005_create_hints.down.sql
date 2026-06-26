-- Migration: 005_create_hints (rollback)
DROP INDEX IF EXISTS idx_user_hints_hint_id;
DROP INDEX IF EXISTS idx_user_hints_user_id;
DROP TABLE IF EXISTS user_hints;
DROP INDEX IF EXISTS idx_hints_display_order;
DROP INDEX IF EXISTS idx_hints_problem_id;
DROP TABLE IF EXISTS hints;
