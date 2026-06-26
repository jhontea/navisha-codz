-- Migration: 003_create_submissions (rollback)
DROP INDEX IF EXISTS idx_submissions_score;
DROP INDEX IF EXISTS idx_submissions_user_problem;
DROP INDEX IF EXISTS idx_submissions_submitted_at;
DROP INDEX IF EXISTS idx_submissions_status;
DROP INDEX IF EXISTS idx_submissions_problem_id;
DROP INDEX IF EXISTS idx_submissions_user_id;
DROP TABLE IF EXISTS submissions;
DROP TYPE IF EXISTS programming_language;
DROP TYPE IF EXISTS submission_status;
