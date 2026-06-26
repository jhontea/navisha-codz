-- ============================================================================
-- Coding Challenge Platform — PostgreSQL Schema
-- ============================================================================
-- Database: coding_challange
-- Encoding: UTF-8
-- Target: PostgreSQL 15+
-- Description: Complete database schema for HackerRank-like coding challenge platform
-- ============================================================================

-- Enable UUID extension
CREATE EXTENSION IF NOT EXISTS "uuid-ossp";
CREATE EXTENSION IF NOT EXISTS "pgcrypto";

-- ============================================================================
-- 1. ENUMS
-- ============================================================================

-- Role pengguna
CREATE TYPE user_role AS ENUM ('user', 'admin', 'moderator');

-- Tingkat kesulitan problem
CREATE TYPE problem_difficulty AS ENUM ('easy', 'medium', 'hard');

-- Status submission
CREATE TYPE submission_status AS ENUM (
    'pending',
    'running',
    'accepted',
    'wrong_answer',
    'time_limit_exceeded',
    'memory_limit_exceeded',
    'runtime_error',
    'compilation_error'
);

-- Status test case result
CREATE TYPE test_result_status AS ENUM ('passed', 'failed', 'timeout', 'error');

-- Status problem oleh user
CREATE TYPE user_problem_status_enum AS ENUM ('unsolved', 'attempted', 'solved');

-- ============================================================================
-- 2. TABEL USERS
-- ============================================================================

CREATE TABLE users (
    id              UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    username        VARCHAR(50) UNIQUE NOT NULL,
    email           VARCHAR(255) UNIQUE NOT NULL,
    password_hash   VARCHAR(255) NOT NULL,
    full_name       VARCHAR(100),
    avatar_url      VARCHAR(500),
    role            user_role DEFAULT 'user',
    rating          INTEGER DEFAULT 1200 CHECK (rating >= 0),
    total_solved    INTEGER DEFAULT 0 CHECK (total_solved >= 0),
    total_submissions INTEGER DEFAULT 0 CHECK (total_submissions >= 0),
    streak_days     INTEGER DEFAULT 0 CHECK (streak_days >= 0),
    last_active_at  TIMESTAMPTZ,
    created_at      TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at      TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- ============================================================================
-- 3. TABEL PROBLEM CATEGORIES
-- ============================================================================

CREATE TABLE problem_categories (
    id          SERIAL PRIMARY KEY,
    name        VARCHAR(50) UNIQUE NOT NULL,
    slug        VARCHAR(50) UNIQUE NOT NULL,
    description TEXT,
    icon        VARCHAR(100),
    color       VARCHAR(7) CHECK (color ~ '^#[0-9A-Fa-f]{6}$')
);

-- ============================================================================
-- 4. TABEL PROBLEMS
-- ============================================================================

CREATE TABLE problems (
    id                  SERIAL PRIMARY KEY,
    title               VARCHAR(255) NOT NULL,
    slug                VARCHAR(255) UNIQUE NOT NULL,
    description         TEXT NOT NULL,
    difficulty          problem_difficulty NOT NULL,
    category_id         INTEGER REFERENCES problem_categories(id) ON DELETE SET NULL,
    time_limit_seconds  INTEGER DEFAULT 1 CHECK (time_limit_seconds BETWEEN 1 AND 30),
    memory_limit_mb     INTEGER DEFAULT 256 CHECK (memory_limit_mb BETWEEN 64 AND 1024),
    max_score           INTEGER DEFAULT 100 CHECK (max_score > 0),
    function_name       VARCHAR(100),
    return_type         VARCHAR(50),
    template_code       TEXT,
    is_published        BOOLEAN DEFAULT FALSE,
    created_by          UUID REFERENCES users(id) ON DELETE SET NULL,
    created_at          TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at          TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- ============================================================================
-- 5. TABEL PROBLEM TAGS (Many-to-Many)
-- ============================================================================

CREATE TABLE problem_tags (
    problem_id  INTEGER NOT NULL REFERENCES problems(id) ON DELETE CASCADE,
    tag_name    VARCHAR(50) NOT NULL,
    PRIMARY KEY (problem_id, tag_name)
);

-- ============================================================================
-- 6. TABEL TEST CASES
-- ============================================================================

CREATE TABLE test_cases (
    id              SERIAL PRIMARY KEY,
    problem_id      INTEGER NOT NULL REFERENCES problems(id) ON DELETE CASCADE,
    input           TEXT NOT NULL,
    expected_output TEXT NOT NULL,
    is_hidden       BOOLEAN DEFAULT FALSE,
    is_sample       BOOLEAN DEFAULT FALSE,
    description     TEXT,
    weight          INTEGER DEFAULT 1 CHECK (weight > 0)
);

-- ============================================================================
-- 7. TABEL PROBLEM HINTS
-- ============================================================================

CREATE TABLE problem_hints (
    id              SERIAL PRIMARY KEY,
    problem_id      INTEGER NOT NULL REFERENCES problems(id) ON DELETE CASCADE,
    level           INTEGER NOT NULL CHECK (level BETWEEN 1 AND 3),
    title           VARCHAR(255),
    content         TEXT NOT NULL,
    score_penalty   INTEGER DEFAULT 0 CHECK (score_penalty BETWEEN 0 AND 100)
);

-- ============================================================================
-- 8. TABEL SUBMISSIONS
-- ============================================================================

CREATE TABLE submissions (
    id                  UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    user_id             UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    problem_id          INTEGER NOT NULL REFERENCES problems(id) ON DELETE CASCADE,
    code                TEXT NOT NULL,
    language            VARCHAR(20) DEFAULT 'go',
    status              submission_status DEFAULT 'pending',
    score               INTEGER DEFAULT 0 CHECK (score >= 0),
    execution_time_ms   INTEGER CHECK (execution_time_ms >= 0),
    memory_used_kb      INTEGER CHECK (memory_used_kb >= 0),
    test_cases_passed   INTEGER DEFAULT 0 CHECK (test_cases_passed >= 0),
    test_cases_total    INTEGER DEFAULT 0 CHECK (test_cases_total >= 0),
    error_message       TEXT,
    submitted_at        TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    completed_at        TIMESTAMPTZ
);

-- ============================================================================
-- 9. TABEL SUBMISSION TEST RESULTS
-- ============================================================================

CREATE TABLE submission_test_results (
    id                  SERIAL PRIMARY KEY,
    submission_id       UUID NOT NULL REFERENCES submissions(id) ON DELETE CASCADE,
    test_case_id        INTEGER NOT NULL REFERENCES test_cases(id) ON DELETE CASCADE,
    status              test_result_status NOT NULL,
    execution_time_ms   INTEGER CHECK (execution_time_ms >= 0),
    memory_used_kb      INTEGER CHECK (memory_used_kb >= 0),
    actual_output       TEXT,
    error_message       TEXT
);

-- ============================================================================
-- 10. TABEL HINTS USED
-- ============================================================================

CREATE TABLE hints_used (
    id          SERIAL PRIMARY KEY,
    user_id     UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    problem_id  INTEGER NOT NULL REFERENCES problems(id) ON DELETE CASCADE,
    hint_id     INTEGER NOT NULL REFERENCES problem_hints(id) ON DELETE CASCADE,
    used_at     TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- ============================================================================
-- 11. TABEL LEADERBOARD
-- ============================================================================

CREATE TABLE leaderboard (
    id              SERIAL PRIMARY KEY,
    user_id         UUID NOT NULL UNIQUE REFERENCES users(id) ON DELETE CASCADE,
    weekly_score    INTEGER DEFAULT 0 CHECK (weekly_score >= 0),
    monthly_score   INTEGER DEFAULT 0 CHECK (monthly_score >= 0),
    all_time_score  INTEGER DEFAULT 0 CHECK (all_time_score >= 0),
    weekly_rank     INTEGER,
    monthly_rank    INTEGER,
    all_time_rank   INTEGER,
    updated_at      TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- ============================================================================
-- 12. TABEL USER PROBLEM STATUS (Cache Table)
-- ============================================================================

CREATE TABLE user_problem_status (
    user_id             UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    problem_id          INTEGER NOT NULL REFERENCES problems(id) ON DELETE CASCADE,
    status              user_problem_status_enum DEFAULT 'unsolved',
    best_score          INTEGER DEFAULT 0 CHECK (best_score >= 0),
    attempts            INTEGER DEFAULT 0 CHECK (attempts >= 0),
    solved_at           TIMESTAMPTZ,
    hints_used_count    INTEGER DEFAULT 0 CHECK (hints_used_count >= 0),
    PRIMARY KEY (user_id, problem_id)
);

-- ============================================================================
-- 13. TABEL LEADERBOARD HISTORY (untuk partitioning)
-- ============================================================================

CREATE TABLE leaderboard_history (
    id              BIGSERIAL,
    user_id         UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    contest_id      VARCHAR(50) NOT NULL,
    score           INTEGER DEFAULT 0,
    rank            INTEGER,
    period_start    DATE NOT NULL,
    period_end      DATE NOT NULL,
    recorded_at     TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    PRIMARY KEY (id, period_start)
) PARTITION BY RANGE (period_start);

-- ============================================================================
-- 14. TABEL USER RATING HISTORY
-- ============================================================================

CREATE TABLE user_rating_history (
    id              BIGSERIAL PRIMARY KEY,
    user_id         UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    old_rating      INTEGER NOT NULL,
    new_rating      INTEGER NOT NULL,
    rating_change   INTEGER NOT NULL,
    reason          VARCHAR(100), -- 'submission_accepted', 'streak_bonus', dll
    reference_id    UUID, -- submission_id atau reference lain
    created_at      TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- ============================================================================
-- INDEXES
-- ============================================================================

-- Users indexes
CREATE INDEX idx_users_username ON users(username);
CREATE INDEX idx_users_email ON users(email);
CREATE INDEX idx_users_role ON users(role);
CREATE INDEX idx_users_rating ON users(rating DESC);
CREATE INDEX idx_users_created_at ON users(created_at);

-- Problem categories indexes
CREATE INDEX idx_problem_categories_slug ON problem_categories(slug);

-- Problems indexes
CREATE INDEX idx_problems_category_id ON problems(category_id);
CREATE INDEX idx_problems_difficulty ON problems(difficulty);
CREATE INDEX idx_problems_slug ON problems(slug);
CREATE INDEX idx_problems_is_published ON problems(is_published) WHERE is_published = TRUE;
CREATE INDEX idx_problems_created_by ON problems(created_by);
CREATE INDEX idx_problems_difficulty_category ON problems(difficulty, category_id);
CREATE INDEX idx_problems_created_at ON problems(created_at DESC);

-- Composite index untuk problem listing — filter difficulty+categories, sort by title
CREATE INDEX idx_problems_listing ON problems(difficulty, category_id, title);

-- Problem tags indexes
CREATE INDEX idx_problem_tags_tag_name ON problem_tags(tag_name);
CREATE INDEX idx_problem_tags_problem_id ON problem_tags(problem_id);

-- Test cases indexes
CREATE INDEX idx_test_cases_problem_id ON test_cases(problem_id);
CREATE INDEX idx_test_cases_is_hidden ON test_cases(is_hidden) WHERE is_hidden = FALSE;

-- Composite index untuk test case queries — filter by problem + visibility
CREATE INDEX idx_test_cases_problem_visibility ON test_cases(problem_id, is_hidden);

-- Problem hints indexes
CREATE INDEX idx_problem_hints_problem_id ON problem_hints(problem_id);
CREATE INDEX idx_problem_hints_level ON problem_hints(problem_id, level);

-- Subscriptions indexes
CREATE INDEX idx_submissions_user_id ON submissions(user_id);
CREATE INDEX idx_submissions_problem_id ON submissions(problem_id);
CREATE INDEX idx_submissions_status ON submissions(status);
CREATE INDEX idx_submissions_submitted_at ON submissions(submitted_at DESC);
CREATE INDEX idx_submissions_user_problem ON submissions(user_id, problem_id);
CREATE INDEX idx_submissions_user_status ON submissions(user_id, status);
CREATE INDEX idx_submissions_problem_status ON submissions(problem_id, status);
CREATE INDEX idx_submissions_score ON submissions(score DESC);

-- Composite index untuk submission history — filter user+status, sort by submitted_at
CREATE INDEX idx_submissions_history ON submissions(user_id, status, submitted_at DESC);

-- Partial index untuk pending submissions (queue processing)
CREATE INDEX idx_submissions_pending ON submissions(status, submitted_at)
    WHERE status = 'pending';

-- Partial index untuk running submissions
CREATE INDEX idx_submissions_running ON submissions(status, submitted_at)
    WHERE status = 'running';

-- Composite index untuk accepted submissions
CREATE INDEX idx_submissions_accepted ON submissions(user_id, problem_id, submitted_at DESC)
    WHERE status = 'accepted';

-- Submission test results indexes
CREATE INDEX idx_submission_test_results_submission_id ON submission_test_results(submission_id);
CREATE INDEX idx_submission_test_results_test_case_id ON submission_test_results(test_case_id);
CREATE INDEX idx_submission_test_results_status ON submission_test_results(status);

-- Hints used indexes
CREATE INDEX idx_hints_used_user_id ON hints_used(user_id);
CREATE INDEX idx_hints_used_problem_id ON hints_used(problem_id);
CREATE INDEX idx_hints_used_user_problem ON hints_used(user_id, problem_id);

-- Leaderboard indexes
CREATE INDEX idx_leaderboard_weekly_score ON leaderboard(weekly_score DESC);
CREATE INDEX idx_leaderboard_monthly_score ON leaderboard(monthly_score DESC);
CREATE INDEX idx_leaderboard_all_time_score ON leaderboard(all_time_score DESC);
CREATE INDEX idx_leaderboard_weekly_rank ON leaderboard(weekly_rank);
CREATE INDEX idx_leaderboard_monthly_rank ON leaderboard(monthly_rank);
CREATE INDEX idx_leaderboard_all_time_rank ON leaderboard(all_time_rank);

-- User problem status indexes
CREATE INDEX idx_user_problem_status_user_id ON user_problem_status(user_id);
CREATE INDEX idx_user_problem_status_problem_id ON user_problem_status(problem_id);
CREATE INDEX idx_user_problem_status_status ON user_problem_status(status);
CREATE INDEX idx_user_problem_status_user_status ON user_problem_status(user_id, status);

-- User rating history indexes
CREATE INDEX idx_user_rating_history_user_id ON user_rating_history(user_id);
CREATE INDEX idx_user_rating_history_created_at ON user_rating_history(created_at DESC);

-- ============================================================================
-- PARTITIONING UNTUK SUBMISSIONS TABLE
-- ============================================================================

-- Note: Untuk implementasi partitioning, tabel submissions perlu dibuat ulang
-- dengan partitioning by range (submitted_at). Berikut contoh implementasinya:

/*
-- Membuat tabel submissions yang di-partisi
CREATE TABLE submissions_partitioned (
    id                  UUID NOT NULL DEFAULT uuid_generate_v4(),
    user_id             UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    problem_id          INTEGER NOT NULL REFERENCES problems(id) ON DELETE CASCADE,
    code                TEXT NOT NULL,
    language            VARCHAR(20) DEFAULT 'go',
    status              submission_status DEFAULT 'pending',
    score               INTEGER DEFAULT 0 CHECK (score >= 0),
    execution_time_ms   INTEGER CHECK (execution_time_ms >= 0),
    memory_used_kb      INTEGER CHECK (memory_used_kb >= 0),
    test_cases_passed   INTEGER DEFAULT 0 CHECK (test_cases_passed >= 0),
    test_cases_total    INTEGER DEFAULT 0 CHECK (test_cases_total >= 0),
    error_message       TEXT,
    submitted_at        TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    completed_at        TIMESTAMPTZ,
    PRIMARY KEY (id, submitted_at)
) PARTITION BY RANGE (submitted_at);

-- Membuat partition per bulan
CREATE TABLE submissions_y2025m01 PARTITION OF submissions_partitioned
    FOR VALUES FROM ('2025-01-01') TO ('2025-02-01');
CREATE TABLE submissions_y2025m02 PARTITION OF submissions_partitioned
    FOR VALUES FROM ('2025-02-01') TO ('2025-03-01');
CREATE TABLE submissions_y2025m03 PARTITION OF submissions_partitioned
    FOR VALUES FROM ('2025-03-01') TO ('2025-04-01');
-- ... dst, buat otomatis dengan script atau pg_partman

-- Index pada partitioned table
CREATE INDEX idx_submissions_part_user_id ON submissions_partitioned(user_id);
CREATE INDEX idx_submissions_part_problem_id ON submissions_partitioned(problem_id);
CREATE INDEX idx_submissions_part_status ON submissions_partitioned(status);
CREATE INDEX idx_submissions_part_submitted_at ON submissions_partitioned(submitted_at DESC);
CREATE INDEX idx_submissions_part_pending ON submissions_partitioned(status, submitted_at)
    WHERE status = 'pending';
*/

-- ============================================================================
-- PARTITIONING UNTUK LEADERBOARD HISTORY
-- ============================================================================

-- Membuat partition untuk leaderboard history (per kuartal)
CREATE TABLE leaderboard_history_2025_q1 PARTITION OF leaderboard_history
    FOR VALUES FROM ('2025-01-01') TO ('2025-04-01');
CREATE TABLE leaderboard_history_2025_q2 PARTITION OF leaderboard_history
    FOR VALUES FROM ('2025-04-01') TO ('2025-07-01');
CREATE TABLE leaderboard_history_2025_q3 PARTITION OF leaderboard_history
    FOR VALUES FROM ('2025-07-01') TO ('2025-10-01');
CREATE TABLE leaderboard_history_2025_q4 PARTITION OF leaderboard_history
    FOR VALUES FROM ('2025-10-01') TO ('2026-01-01');

-- ============================================================================
-- FUNCTIONS & TRIGGERS
-- ============================================================================

-- Auto-update updated_at timestamp
CREATE OR REPLACE FUNCTION update_updated_at_column()
RETURNS TRIGGER AS $$
BEGIN
    NEW.updated_at = NOW();
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

-- Trigger untuk users
CREATE TRIGGER trigger_users_updated_at
    BEFORE UPDATE ON users
    FOR EACH ROW
    EXECUTE FUNCTION update_updated_at_column();

-- Trigger untuk problems
CREATE TRIGGER trigger_problems_updated_at
    BEFORE UPDATE ON problems
    FOR EACH ROW
    EXECUTE FUNCTION update_updated_at_column();

-- Trigger untuk leaderboard
CREATE TRIGGER trigger_leaderboard_updated_at
    BEFORE UPDATE ON leaderboard
    FOR EACH ROW
    EXECUTE FUNCTION update_updated_at_column();

-- ============================================================================
-- SEED DATA
-- ============================================================================

-- Insert default categories
INSERT INTO problem_categories (name, slug, description, icon, color) VALUES
    ('Array', 'array', 'Problems involving array manipulation and traversal', '📊', '#3B82F6'),
    ('String', 'string', 'Problems involving string processing and manipulation', '📝', '#10B981'),
    ('Dynamic Programming', 'dynamic-programming', 'Problems solvable with DP approach', '🧩', '#F59E0B'),
    ('Graph', 'graph', 'Problems involving graph traversal and algorithms', '🕸️', '#EF4444'),
    ('Tree', 'tree', 'Problems involving tree data structures', '🌳', '#8B5CF6'),
    ('Math', 'math', 'Mathematical and computational problems', '🔢', '#EC4899'),
    ('Sorting', 'sorting', 'Problems involving sorting algorithms', '📈', '#06B6D4'),
    ('Binary Search', 'binary-search', 'Problems solvable with binary search', '🔍', '#84CC16'),
    ('Stack', 'stack', 'Problems involving stack data structure', '📚', '#F97316'),
    ('Queue', 'queue', 'Problems involving queue data structure', '🎫', '#14B8A6'),
    ('Linked List', 'linked-list', 'Problems involving linked list operations', '🔗', '#6366F1'),
    ('Hash Table', 'hash-table', 'Problems involving hash maps and sets', '🗝️', '#A855F7'),
    ('Greedy', 'greedy', 'Problems solvable with greedy approach', '💰', '#EAB308'),
    ('Backtracking', 'backtracking', 'Problems solvable with backtracking', '🔄', '#DC2626'),
    ('Two Pointers', 'two-pointers', 'Problems solvable with two pointers technique', '👉', '#0EA5E9');

-- Insert default admin user (password: admin123 — hashed with bcrypt)
INSERT INTO users (username, email, password_hash, full_name, role)
VALUES ('admin', 'admin@codingchallange.com', '$2a$10$N9qo8uLOickgx2ZMRZoMyeIjZAgcfl7p92ldGxad68LJZdL17lhWy', 'Administrator', 'admin')
ON CONFLICT DO NOTHING;

-- ============================================================================
-- END OF SCHEMA
-- ============================================================================
