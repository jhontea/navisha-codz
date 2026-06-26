-- Migration: 002_create_problems (rollback)
DROP INDEX IF EXISTS idx_problems_created_by;
DROP INDEX IF EXISTS idx_problems_tags;
DROP INDEX IF EXISTS idx_problems_is_published;
DROP INDEX IF EXISTS idx_problems_difficulty;
DROP INDEX IF EXISTS idx_problems_slug;
DROP TABLE IF EXISTS problems;
DROP TYPE IF EXISTS difficulty_level;
