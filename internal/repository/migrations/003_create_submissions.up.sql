-- Migration: 003_create_submissions
-- Description: Create submissions table
-- Created: 2024-01-01

CREATE TYPE submission_status AS ENUM ('pending', 'queued', 'running', 'completed', 'failed', 'timeout', 'compilation_error');
CREATE TYPE programming_language AS ENUM ('go', 'python', 'javascript', 'java', 'cpp', 'rust', 'typescript');

CREATE TABLE IF NOT EXISTS submissions (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    problem_id INTEGER NOT NULL REFERENCES problems(id) ON DELETE CASCADE,
    code TEXT NOT NULL,
    language programming_language NOT NULL,
    status submission_status NOT NULL DEFAULT 'pending',
    score INTEGER DEFAULT 0,
    execution_time_ms INTEGER,
    memory_used_kb INTEGER,
    test_cases_passed INTEGER DEFAULT 0,
    test_cases_total INTEGER DEFAULT 0,
    error_message TEXT,
    submitted_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    completed_at TIMESTAMPTZ,
    execution_metadata JSONB DEFAULT '{}'
);

CREATE INDEX idx_submissions_user_id ON submissions (user_id);
CREATE INDEX idx_submissions_problem_id ON submissions (problem_id);
CREATE INDEX idx_submissions_status ON submissions (status);
CREATE INDEX idx_submissions_submitted_at ON submissions (submitted_at DESC);
CREATE INDEX idx_submissions_user_problem ON submissions (user_id, problem_id);
CREATE INDEX idx_submissions_score ON submissions (score DESC);
