-- Migration: 002_create_problems
-- Description: Create problems table and related types
-- Created: 2024-01-01

CREATE TYPE difficulty_level AS ENUM ('easy', 'medium', 'hard', 'expert');

CREATE TABLE IF NOT EXISTS problems (
    id SERIAL PRIMARY KEY,
    title VARCHAR(255) NOT NULL,
    slug VARCHAR(255) UNIQUE NOT NULL,
    description TEXT NOT NULL,
    difficulty difficulty_level NOT NULL DEFAULT 'easy',
    max_score INTEGER NOT NULL DEFAULT 100 CHECK (max_score > 0),
    time_limit_ms INTEGER NOT NULL DEFAULT 1000,
    memory_limit_mb INTEGER NOT NULL DEFAULT 256,
    tags TEXT[] DEFAULT '{}',
    starter_code TEXT,
    solution_code TEXT,
    test_case_json JSONB NOT NULL DEFAULT '[]',
    is_published BOOLEAN NOT NULL DEFAULT FALSE,
    created_by UUID REFERENCES users(id),
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_problems_slug ON problems (slug);
CREATE INDEX idx_problems_difficulty ON problems (difficulty);
CREATE INDEX idx_problems_is_published ON problems (is_published);
CREATE INDEX idx_problems_tags ON problems USING GIN (tags);
CREATE INDEX idx_problems_created_by ON problems (created_by);
