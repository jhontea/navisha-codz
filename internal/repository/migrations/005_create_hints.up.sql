-- Migration: 005_create_hints
-- Description: Create hints table
-- Created: 2024-01-01

CREATE TABLE IF NOT EXISTS hints (
    id SERIAL PRIMARY KEY,
    problem_id INTEGER NOT NULL REFERENCES problems(id) ON DELETE CASCADE,
    content TEXT NOT NULL,
    penalty_percent INTEGER NOT NULL DEFAULT 10 CHECK (penalty_percent >= 0 AND penalty_percent <= 100),
    display_order INTEGER NOT NULL DEFAULT 0,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_hints_problem_id ON hints (problem_id);
CREATE INDEX idx_hints_display_order ON hints (problem_id, display_order);

-- Track which hints a user has unlocked
CREATE TABLE IF NOT EXISTS user_hints (
    id SERIAL PRIMARY KEY,
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    hint_id INTEGER NOT NULL REFERENCES hints(id) ON DELETE CASCADE,
    unlocked_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    UNIQUE(user_id, hint_id)
);

CREATE INDEX idx_user_hints_user_id ON user_hints (user_id);
CREATE INDEX idx_user_hints_hint_id ON user_hints (hint_id);
