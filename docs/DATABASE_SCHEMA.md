# Database Schema — Coding Challenge Platform

> **Versi:** 1.0.0  
> **Database:** PostgreSQL 15+  
> **Encoding:** UTF-8  
> **Tanggal:** Juni 2026  
> **Author:** Database Architect  

---

## Daftar Isi

1. [Overview](#overview)
2. [Entity Relationship Diagram](#entity-relationship-diagram)
3. [Tabel-tabel](#tabel-tabel)
   - [Users](#1-users)
   - [Problem Categories](#2-problem-categories)
   - [Problems](#3-problems)
   - [Problem Tags](#4-problem-tags)
   - [Test Cases](#5-test-cases)
   - [Problem Hints](#6-problem-hints)
   - [Submissions](#7-submissions)
   - [Submission Test Results](#8-submission-test-results)
   - [Hints Used](#9-hints-used)
   - [Leaderboard](#10-leaderboard)
   - [User Problem Status](#11-user-problem-status)
   - [Leaderboard History](#12-leaderboard-history)
   - [User Rating History](#13-user-rating-history)
4. [Indeks](#indeks)
5. [Query Penting](#query-penting)
6. [Redis Cache Strategy](#redis-cache-strategy)
7. [Partitioning Strategy](#partitioning-strategy)
8. [Maintenance & Monitoring](#maintenance--monitoring)

---

## Overview

Database schema untuk platform Coding Challenge (HackerRank-like) yang mendukung:
- **1000+ active users**
- **500+ submissions/hari**
- Real-time leaderboard
- Multi-language code execution
- Progressive hint system
- Rating system (Elo-based)

### Tech Stack
- **Database:** PostgreSQL 15+
- **Cache:** Redis 7+
- **Message Queue:** RabbitMQ
- **Search:** PostgreSQL Full-Text Search (opsional: Elasticsearch)

---

## Entity Relationship Diagram

```
┌──────────────┐     ┌─────────────────────┐     ┌──────────────────┐
│    users     │     │  problem_categories  │     │   leaderboard    │
├──────────────┤     ├─────────────────────┤     ├──────────────────┤
│ id (PK, UUID)│◄────┤ id (PK, SERIAL)     │     │ id (PK, SERIAL)  │
│ username     │     │ name                │     │ user_id (FK)     │
│ email        │     │ slug                │     │ weekly_score     │
│ password_hash│     │ description         │     │ monthly_score    │
│ role         │     │ icon                │     │ all_time_score   │
│ rating       │     │ color               │     │ weekly_rank      │
│ total_solved │     └─────────┬───────────┘     │ monthly_rank     │
│ created_at   │               │                  │ all_time_rank    │
└──────┬───────┘               │                  └──────────────────┘
       │                       │
       │         ┌─────────────▼───────────┐        ┌──────────────────┐
       │         │       problems          │        │  problem_tags    │
       │         ├─────────────────────────┤        ├──────────────────┤
       │         │ id (PK, SERIAL)         │◄───────┤ problem_id (FK)  │
       │         │ title                   │        │ tag_name         │
       │         │ slug                    │        └──────────────────┘
       │         │ description             │
       │         │ difficulty (ENUM)       │        ┌──────────────────┐
       │         │ category_id (FK)        │◄───────┤   test_cases     │
       │         │ time_limit_seconds      │        ├──────────────────┤
       │         │ memory_limit_mb         │        │ id (PK, SERIAL)  │
       │         │ max_score               │        │ problem_id (FK)  │
       │         │ function_name           │        │ input            │
       │         │ template_code           │        │ expected_output  │
       │         │ is_published            │        │ is_hidden        │
       │         │ created_by (FK)         │        │ is_sample        │
       │         └────────────┬────────────┘        │ weight           │
       │                      │                     └────────┬─────────┘
       │                      │                              │
       │         ┌────────────▼────────────┐                 │
       │         │    problem_hints        │                 │
       │         ├─────────────────────────┤                 │
       │         │ id (PK, SERIAL)         │                 │
       │         │ problem_id (FK)         │                 │
       │         │ level (1-3)             │                 │
       │         │ title                   │                 │
       │         │ content                 │                 │
       │         │ score_penalty           │                 │
       │         └────────────┬────────────┘                 │
       │                      │                              │
  ┌────▼──────────────────────▼──────────────────────────────▼────┐
  │                        submissions                             │
  ├───────────────────────────────────────────────────────────────┤
  │ id (PK, UUID)                                                 │
  │ user_id (FK)                                                  │
  │ problem_id (FK)                                               │
  │ code                                                          │
  │ language                                                      │
  │ status (ENUM)                                                 │
  │ score                                                         │
  │ execution_time_ms                                             │
  │ memory_used_kb                                                │
  │ test_cases_passed                                             │
  │ test_cases_total                                              │
  │ submitted_at                                                  │
  └───────────────────────────┬───────────────────────────────────┘
                              │
  ┌───────────────────────────▼───────────────────────────────────┐
  │                  submission_test_results                       │
  ├───────────────────────────────────────────────────────────────┤
  │ id (PK, SERIAL)                                               │
  │ submission_id (FK)                                            │
  │ test_case_id (FK)                                             │
  │ status (ENUM)                                                 │
  │ execution_time_ms                                             │
  │ memory_used_kb                                                │
  │ actual_output                                                 │
  └───────────────────────────────────────────────────────────────┘

  ┌─────────────────────────┐    ┌───────────────────────────────┐
  │      hints_used         │    │     user_problem_status       │
  ├─────────────────────────┤    ├───────────────────────────────┤
  │ id (PK, SERIAL)         │    │ user_id (FK, PK)              │
  │ user_id (FK)            │    │ problem_id (FK, PK)           │
  │ problem_id (FK)         │    │ status (ENUM)                 │
  │ hint_id (FK)            │    │ best_score                    │
  │ used_at                 │    │ attempts                      │
  └─────────────────────────┘    │ solved_at                     │
                                 │ hints_used_count              │
                                 └───────────────────────────────┘
```

---

## Tabel-tabel

### 1. Users

Menyimpan data pengguna platform.

| Kolom | Tipe | Constraint | Keterangan |
|-------|------|------------|------------|
| `id` | UUID | PRIMARY KEY, DEFAULT uuid_generate_v4() | Unique identifier |
| `username` | VARCHAR(50) | UNIQUE, NOT NULL | Username untuk login |
| `email` | VARCHAR(255) | UNIQUE, NOT NULL | Email pengguna |
| `password_hash` | VARCHAR(255) | NOT NULL | Bcrypt hash password |
| `full_name` | VARCHAR(100) | NULLABLE | Nama lengkap |
| `avatar_url` | VARCHAR(500) | NULLABLE | URL avatar |
| `role` | user_role | DEFAULT 'user' | ENUM: user, admin, moderator |
| `rating` | INTEGER | DEFAULT 1200, CHECK >= 0 | Elo rating |
| `total_solved` | INTEGER | DEFAULT 0 | Total problem solved |
| `total_submissions` | INTEGER | DEFAULT 0 | Total submissions |
| `streak_days` | INTEGER | DEFAULT 0 | Streak harian |
| `last_active_at` | TIMESTAMPTZ | NULLABLE | Terakhir aktif |
| `created_at` | TIMESTAMPTZ | NOT NULL, DEFAULT NOW() | Waktu dibuat |
| `updated_at` | TIMESTAMPTZ | NOT NULL, DEFAULT NOW() | Waktu diupdate |

---

### 2. Problem Categories

Kategori utama problem (array, string, DP, graph, dll).

| Kolom | Tipe | Constraint | Keterangan |
|-------|------|------------|------------|
| `id` | SERIAL | PRIMARY KEY | Auto-increment |
| `name` | VARCHAR(50) | UNIQUE NOT NULL | "Dynamic Programming" |
| `slug` | VARCHAR(50) | UNIQUE NOT NULL | "dynamic-programming" |
| `description` | TEXT | NULLABLE | Deskripsi kategori |
| `icon` | VARCHAR(100) | NULLABLE | Icon class/emoji |
| `color` | VARCHAR(7) | CHECK hex format | Warna hex (#FF5733) |

---

### 3. Problems

Master data problem/challenge.

| Kolom | Tipe | Constraint | Keterangan |
|-------|------|------------|------------|
| `id` | SERIAL | PRIMARY KEY | Auto-increment |
| `title` | VARCHAR(255) | NOT NULL | Judul problem |
| `slug` | VARCHAR(255) | UNIQUE NOT NULL | URL-friendly slug |
| `description` | TEXT | NOT NULL | Markdown content |
| `difficulty` | problem_difficulty | NOT NULL | ENUM: easy, medium, hard |
| `category_id` | INTEGER | FK → problem_categories | Kategori utama |
| `time_limit_seconds` | INTEGER | DEFAULT 1, CHECK 1-30 | Batas waktu eksekusi |
| `memory_limit_mb` | INTEGER | DEFAULT 256, CHECK 64-1024 | Batas memori |
| `max_score` | INTEGER | DEFAULT 100 | Skor maksimum |
| `function_name` | VARCHAR(100) | NULLABLE | Nama function (function-based) |
| `return_type` | VARCHAR(50) | NULLABLE | Tipe return |
| `template_code` | TEXT | NULLABLE | Starter code |
| `is_published` | BOOLEAN | DEFAULT FALSE | Status publikasi |
| `created_by` | UUID | FK → users | Creator |
| `created_at` | TIMESTAMPTZ | NOT NULL | Waktu dibuat |
| `updated_at` | TIMESTAMPTZ | NOT NULL | Waktu diupdate |

---

### 4. Problem Tags

Relasi many-to-many problem dengan tags.

| Kolom | Tipe | Constraint | Keterangan |
|-------|------|------------|------------|
| `problem_id` | INTEGER | FK → problems, PK | Reference ke problem |
| `tag_name` | VARCHAR(50) | PK | "memoization", "two-pointers" |

---

### 5. Test Cases

Test case untuk setiap problem.

| Kolom | Tipe | Constraint | Keterangan |
|-------|------|------------|------------|
| `id` | SERIAL | PRIMARY KEY | Auto-increment |
| `problem_id` | INTEGER | FK → problems, NOT NULL | Reference ke problem |
| `input` | TEXT | NOT NULL | Input test case |
| `expected_output` | TEXT | NOT NULL | Expected output |
| `is_hidden` | BOOLEAN | FALSE | Hidden dari user |
| `is_sample` | BOOLEAN | FALSE | Shown di deskripsi |
| `description` | TEXT | NULLABLE | Deskripsi test case |
| `weight` | INTEGER | DEFAULT 1, CHECK > 0 | Bobot skor |

---

### 6. Problem Hints

Hint progresif untuk setiap problem.

| Kolom | Tipe | Constraint | Keterangan |
|-------|------|------------|------------|
| `id` | SERIAL | PRIMARY KEY | Auto-increment |
| `problem_id` | INTEGER | FK → problems | Reference ke problem |
| `level` | INTEGER | CHECK 1-3 | 1=umum, 2=teknis, 3=advanced |
| `title` | VARCHAR(255) | NULLABLE | Judul hint |
| `content` | TEXT | NOT NULL | Isi hint |
| `score_penalty` | INTEGER | DEFAULT 0, CHECK 0-100 | Penalti skor (%) |

---

### 7. Submissions

Record setiap submission kode.

| Kolom | Tipe | Constraint | Keterangan |
|-------|------|------------|------------|
| `id` | UUID | PRIMARY KEY | Unique identifier |
| `user_id` | UUID | FK → users, NOT NULL | User yang submit |
| `problem_id` | INTEGER | FK → problems, NOT NULL | Problem yang dikerjakan |
| `code` | TEXT | NOT NULL | Source code |
| `language` | VARCHAR(20) DEFAULT 'go' | NOT NULL | Bahasa pemrograman |
| `status` | submission_status | DEFAULT 'pending' | Status eksekusi |
| `score` | INTEGER | DEFAULT 0 | Skor yang didapat |
| `execution_time_ms` | INTEGER | NULLABLE | Waktu eksekusi (ms) |
| `memory_used_kb` | INTEGER | NULLABLE | Memori terpakai (KB) |
| `test_cases_passed` | INTEGER | DEFAULT 0 | Test case passed |
| `test_cases_total` | INTEGER | DEFAULT 0 | Total test case |
| `error_message` | TEXT | NULLABLE | Pesan error |
| `submitted_at` | TIMESTAMPTZ | NOT NULL | Waktu submit |
| `completed_at` | TIMESTAMPTZ | NULLABLE | Waktu selesai |

---

### 8. Submission Test Results

Detail hasil per test case untuk setiap submission.

| Kolom | Tipe | Constraint | Keterangan |
|-------|------|------------|------------|
| `id` | SERIAL | PRIMARY KEY | Auto-increment |
| `submission_id` | UUID | FK → submissions | Reference ke submission |
| `test_case_id` | INTEGER | FK → test_cases | Reference ke test case |
| `status` | test_result_status | NOT NULL | passed/failed/timeout/error |
| `execution_time_ms` | INTEGER | NULLABLE | Waktu eksekusi |
| `memory_used_kb` | INTEGER | NULLABLE | Memori terpakai |
| `actual_output` | TEXT | NULLABLE | Output aktual |
| `error_message` | TEXT | NULLABLE | Pesan error |

---

### 9. Hints Used

Tracking penggunaan hint oleh user.

| Kolom | Tipe | Constraint | Keterangan |
|-------|------|------------|------------|
| `id` | SERIAL | PRIMARY KEY | Auto-increment |
| `user_id` | UUID | FK → users | User yang pakai hint |
| `problem_id` | INTEGER | FK → problems | Problem terkait |
| `hint_id` | INTEGER | FK → problem_hints | Hint yang dipakai |
| `used_at` | TIMESTAMPTZ | NOT NULL | Waktu digunakan |

---

### 10. Leaderboard

Cache leaderboard untuk ranking cepat.

| Kolom | Tipe | Constraint | Keterangan |
|-------|------|------------|------------|
| `id` | SERIAL | PRIMARY KEY | Auto-increment |
| `user_id` | UUID | FK → users, UNIQUE | Reference ke user |
| `weekly_score` | INTEGER | DEFAULT 0 | Skor mingguan |
| `monthly_score` | INTEGER | DEFAULT 0 | Skor bulanan |
| `all_time_score` | INTEGER | DEFAULT 0 | Skor all-time |
| `weekly_rank` | INTEGER | NULLABLE | Ranking mingguan |
| `monthly_rank` | INTEGER | NULLABLE | Ranking bulanan |
| `all_time_rank` | INTEGER | NULLABLE | Ranking all-time |
| `updated_at` | TIMESTAMPTZ | NOT NULL | Waktu diupdate |

---

### 11. User Problem Status

Cache status problem per user (unsolved/attempted/solved).

| Kolom | Tipe | Constraint | Keterangan |
|-------|------|------------|------------|
| `user_id` | UUID | FK → users, PK | Reference ke user |
| `problem_id` | INTEGER | FK → problems, PK | Reference ke problem |
| `status` | user_problem_status_enum | DEFAULT 'unsolved' | Status pengerjaan |
| `best_score` | INTEGER | DEFAULT 0 | Skor terbaik |
| `attempts` | INTEGER | DEFAULT 0 | Jumlah attempt |
| `solved_at` | TIMESTAMPTZ | NULLABLE | Waktu solved |
| `hints_used_count` | INTEGER | DEFAULT 0 | Jumlah hint dipakai |

---

### 12. Leaderboard History

History leaderboard per periode (di-partisi per kuartal).

| Kolom | Tipe | Constraint | Keterangan |
|-------|------|------------|------------|
| `id` | BIGSERIAL | PK (dengan period_start) | Auto-increment |
| `user_id` | UUID | FK → users | Reference ke user |
| `contest_id` | VARCHAR(50) | NOT NULL | ID kontes/periode |
| `score` | INTEGER | DEFAULT 0 | Skor periode |
| `rank` | INTEGER | NULLABLE | Ranking periode |
| `period_start` | DATE | NOT NULL, PK | Mulai periode |
| `period_end` | DATE | NOT NULL | Akhir periode |
| `recorded_at` | TIMESTAMPTZ | NOT NULL | Waktu record |

---

### 13. User Rating History

History perubahan rating user.

| Kolom | Tipe | Constraint | Keterangan |
|-------|------|------------|------------|
| `id` | BIGSERIAL | PRIMARY KEY | Auto-increment |
| `user_id` | UUID | FK → users | Reference ke user |
| `old_rating` | INTEGER | NOT NULL | Rating sebelumnya |
| `new_rating` | INTEGER | NOT NULL | Rating baru |
| `rating_change` | INTEGER | NOT NULL | Perubahan rating |
| `reason` | VARCHAR(100) | NULLABLE | Alasan perubahan |
| `reference_id` | UUID | NULLABLE | Reference ke submission |
| `created_at` | TIMESTAMPTZ | NOT NULL | Waktu perubahan |

---

## Indeks

### Indeks Foreign Key
```cpp
CREATE INDEX idx_users_username ON users(username);
CREATE INDEX idx_users_email ON users(email);
CREATE INDEX idx_problems_category_id ON problems(category_id);
CREATE INDEX idx_problems_created_by ON problems(created_by);
CREATE INDEX idx_test_cases_problem_id ON test_cases(problem_id);
CREATE INDEX idx_problem_hints_problem_id ON problem_hints(problem_id);
CREATE INDEX idx_submissions_user_id ON submissions(user_id);
CREATE INDEX idx_submissions_problem_id ON submissions(problem_id);
CREATE INDEX idx_submission_test_results_submission_id ON submission_test_results(submission_id);
CREATE INDEX idx_submission_test_results_test_case_id ON submission_test_results(test_case_id);
CREATE INDEX idx_hints_used_user_id ON hints_used(user_id);
CREATE INDEX idx_hints_used_problem_id ON hints_used(problem_id);
CREATE INDEX idx_user_problem_status_user_id ON user_problem_status(user_id);
CREATE INDEX idx_user_problem_status_problem_id ON user_problem_status(problem_id);
CREATE INDEX idx_user_rating_history_user_id ON user_rating_history(user_id);
```

### Indeks Composite untuk Query Sering Dipakai
```cpp
CREATE INDEX idx_problems_difficulty_category ON problems(difficulty, category_id);
CREATE INDEX idx_submissions_user_problem ON submissions(user_id, problem_id);
CREATE INDEX idx_submissions_user_status ON submissions(user_id, status);
CREATE INDEX idx_submissions_problem_status ON submissions(problem_id, status);
CREATE INDEX idx_submissions_accepted ON submissions(user_id, problem_id, submitted_at DESC)
    WHERE status = 'accepted';
CREATE INDEX idx_user_problem_status_user_status ON user_problem_status(user_id, status);
CREATE INDEX idx_problem_hints_level ON problem_hints(problem_id, level);
```

### Indeks Partial
```cpp
-- Hanya untuk pending submissions (queue processing)
CREATE INDEX idx_submissions_pending ON submissions(status, submitted_at)
    WHERE status = 'pending';

-- Hanya untuk running submissions
CREATE INDEX idx_submissions_running ON submissions(status, submitted_at)
    WHERE status = 'running';

-- Hanya problem yang published
CREATE INDEX idx_problems_is_published ON problems(is_published) WHERE is_published = TRUE;

-- Hanya test case yang tidak hidden
CREATE INDEX idx_test_cases_is_hidden ON test_cases(is_hidden) WHERE is_hidden = FALSE;
```

### Indeks Leaderboard Ranking
```cpp
CREATE INDEX idx_leaderboard_weekly_score ON leaderboard(weekly_score DESC);
CREATE INDEX idx_leaderboard_monthly_score ON leaderboard(monthly_score DESC);
CREATE INDEX idx_leaderboard_all_time_score ON leaderboard(all_time_score DESC);
CREATE INDEX idx_leaderboard_weekly_rank ON leaderboard(weekly_rank);
CREATE INDEX idx_leaderboard_monthly_rank ON leaderboard(monthly_rank);
CREATE INDEX idx_leaderboard_all_time_rank ON leaderboard(all_time_rank);
```

---

## Query Penting

### 1. Get Problems by Difficulty + Category dengan Pagination

```sql
-- Get problems filtered by difficulty and category with pagination
SELECT 
    p.id,
    p.title,
    p.slug,
    p.difficulty,
    p.max_score,
    pc.name AS category_name,
    pc.slug AS category_slug,
    ARRAY_AGG(pt.tag_name) FILTER (WHERE pt.tag_name IS NOT NULL) AS tags,
    COUNT(*) OVER() AS total_count
FROM problems p
JOIN problem_categories pc ON p.category_id = pc.id
LEFT JOIN problem_tags pt ON p.id = pt.problem_id
WHERE p.is_published = TRUE
    AND ($1::problem_difficulty IS NULL OR p.difficulty = $1)
    AND ($2::INTEGER IS NULL OR p.category_id = $2)
GROUP BY p.id, pc.name, pc.slug
ORDER BY 
    CASE p.difficulty 
        WHEN 'easy' THEN 1 
        WHEN 'medium' THEN 2 
        WHEN 'hard' THEN 3 
    END,
    p.title
LIMIT $3 OFFSET $4;
```

### 2. Get User Submission History

```sql
-- Get submission history for a specific user
SELECT 
    s.id,
    s.problem_id,
    pr.title AS problem_title,
    pr.slug AS problem_slug,
    pr.difficulty,
    s.language,
    s.status,
    s.score,
    s.execution_time_ms,
    s.memory_used_kb,
    s.test_cases_passed,
    s.test_cases_total,
    s.submitted_at,
    s.completed_at
FROM submissions s
JOIN problems pr ON s.problem_id = pr.id
WHERE s.user_id = $1
ORDER BY s.submitted_at DESC
LIMIT $2 OFFSET $3;
```

### 3. Calculate User Score

```sql
-- Calculate total score for a user (best score per problem)
SELECT 
    u.id,
    u.username,
    u.rating,
    COALESCE(SUM(ups.best_score), 0) AS total_score,
    COUNT(ups.problem_id) FILTER (WHERE ups.status = 'solved') AS problems_solved,
    COUNT(ups.problem_id) AS problems_attempted
FROM users u
LEFT JOIN user_problem_status ups ON u.id = ups.user_id
WHERE u.id = $1
GROUP BY u.id, u.username, u.rating;
```

### 4. Get Leaderboard (Weekly, Monthly, All-Time)

```sql
-- Get weekly leaderboard top N
SELECT 
    lb.weekly_rank AS rank,
    u.id AS user_id,
    u.username,
    u.avatar_url,
    u.rating,
    lb.weekly_score AS score
FROM leaderboard lb
JOIN users u ON lb.user_id = u.id
WHERE lb.weekly_rank IS NOT NULL
ORDER BY lb.weekly_rank ASC
LIMIT $1;

-- Get monthly leaderboard top N
SELECT 
    lb.monthly_rank AS rank,
    u.id AS user_id,
    u.username,
    u.avatar_url,
    u.rating,
    lb.monthly_score AS score
FROM leaderboard lb
JOIN users u ON lb.user_id = u.id
WHERE lb.monthly_rank IS NOT NULL
ORDER BY lb.monthly_rank ASC
LIMIT $1;

-- Get all-time leaderboard top N
SELECT 
    lb.all_time_rank AS rank,
    u.id AS user_id,
    u.username,
    u.avatar_url,
    u.rating,
    lb.all_time_score AS score
FROM leaderboard lb
JOIN users u ON lb.user_id = u.id
WHERE lb.all_time_rank IS NOT NULL
ORDER BY lb.all_time_rank ASC
LIMIT $1;
```

### 5. Get Problem dengan Success Rate

```sql
-- Get problems with success rate
SELECT 
    p.id,
    p.title,
    p.slug,
    p.difficulty,
    pc.name AS category_name,
    COUNT(s.id) AS total_submissions,
    COUNT(s.id) FILTER (WHERE s.status = 'accepted') AS accepted_submissions,
    CASE 
        WHEN COUNT(s.id) > 0 
        THEN ROUND(COUNT(s.id) FILTER (WHERE s.status = 'accepted') * 100.0 / COUNT(s.id), 2)
        ELSE 0 
    END AS success_rate
FROM problems p
JOIN problem_categories pc ON p.category_id = pc.id
LEFT JOIN submissions s ON p.id = s.problem_id
WHERE p.is_published = TRUE
GROUP BY p.id, pc.name
ORDER BY success_rate DESC;
```

### 6. Get User Progress by Category

```sql
-- Get user progress breakdown by category
SELECT 
    pc.id AS category_id,
    pc.name AS category_name,
    pc.slug AS category_slug,
    pc.color,
    COUNT(DISTINCT p.id) AS total_problems,
    COUNT(DISTINCT ups.problem_id) FILTER (WHERE ups.status = 'solved') AS solved,
    COUNT(DISTINCT ups.problem_id) FILTER (WHERE ups.status = 'attempted') AS attempted,
    COALESCE(SUM(ups.best_score), 0) AS total_score
FROM problem_categories pc
JOIN problems p ON pc.id = p.category_id AND p.is_published = TRUE
LEFT JOIN user_problem_status ups ON p.id = ups.problem_id AND ups.user_id = $1
GROUP BY pc.id, pc.name, pc.slug, pc.color
ORDER BY pc.name;
```

### 7. Get Hottest Problems (Most Submitted)

```sql
-- Get most submitted problems (trending/hot)
SELECT 
    p.id,
    p.title,
    p.slug,
    p.difficulty,
    pc.name AS category_name,
    COUNT(s.id) AS submission_count,
    COUNT(DISTINCT s.user_id) AS unique_users,
    COUNT(s.id) FILTER (WHERE s.submitted_at > NOW() - INTERVAL '7 days') AS submissions_last_7_days
FROM problems p
JOIN problem_categories pc ON p.category_id = pc.id
LEFT JOIN submissions s ON p.id = s.problem_id
WHERE p.is_published = TRUE
GROUP BY p.id, pc.name
ORDER BY submissions_last_7_days DESC, submission_count DESC
LIMIT $1;
```

### 8. Get User Rating History

```sql
-- Get user rating history
SELECT 
    id,
    old_rating,
    new_rating,
    rating_change,
    reason,
    reference_id,
    created_at
FROM user_rating_history
WHERE user_id = $1
ORDER BY created_at DESC
LIMIT $2 OFFSET $3;
```

### 9. Get Submission Statistics Per Day

```sql
-- Get daily submission statistics
SELECT 
    DATE(submitted_at) AS date,
    COUNT(*) AS total_submissions,
    COUNT(*) FILTER (WHERE status = 'accepted') AS accepted,
    COUNT(*) FILTER (WHERE status = 'wrong_answer') AS wrong_answer,
    COUNT(*) FILTER (WHERE status = 'time_limit_exceeded') AS tle,
    COUNT(*) FILTER (WHERE status = 'runtime_error') AS runtime_error,
    COUNT(*) FILTER (WHERE status = 'compilation_error') AS compilation_error,
    COUNT(DISTINCT user_id) AS unique_users
FROM submissions
WHERE submitted_at >= $1 AND submitted_at < $2
GROUP BY DATE(submitted_at)
ORDER BY date DESC;
```

### 10. Get Problems Not Solved by User

```sql
-- Get problems not yet solved by user (recommendation)
SELECT 
    p.id,
    p.title,
    p.slug,
    p.difficulty,
    p.max_score,
    pc.name AS category_name,
    pc.slug AS category_slug,
    ARRAY_AGG(pt.tag_name) FILTER (WHERE pt.tag_name IS NOT NULL) AS tags,
    CASE 
        WHEN ups.status = 'attempted' THEN TRUE 
        ELSE FALSE 
    END AS attempted
FROM problems p
JOIN problem_categories pc ON p.category_id = pc.id
LEFT JOIN problem_tags pt ON p.id = pt.problem_id
LEFT JOIN user_problem_status ups ON p.id = ups.problem_id AND ups.user_id = $1
WHERE p.is_published = TRUE
    AND (ups.status IS NULL OR ups.status != 'solved')
GROUP BY p.id, pc.name, pc.slug, ups.status
ORDER BY 
    CASE p.difficulty 
        WHEN 'easy' THEN 1 
        WHEN 'medium' THEN 2 
        WHEN 'hard' THEN 3 
    END,
    p.created_at DESC
LIMIT $2 OFFSET $3;
```

### 11. Get User Dashboard Stats

```sql
-- Get comprehensive user dashboard statistics
SELECT 
    u.id,
    u.username,
    u.rating,
    u.total_solved,
    u.total_submissions,
    u.streak_days,
    u.last_active_at,
    lb.all_time_rank,
    lb.weekly_rank,
    (SELECT COUNT(*) FROM problems WHERE is_published = TRUE) AS total_available_problems,
    (SELECT COUNT(DISTINCT problem_id) FROM user_problem_status WHERE user_id = u.id AND status = 'solved') AS solved_problems,
    (SELECT COUNT(DISTINCT problem_id) FROM user_problem_status WHERE user_id = u.id) AS attempted_problems
FROM users u
LEFT JOIN leaderboard lb ON u.id = lb.user_id
WHERE u.id = $1;
```

### 12. Get Problem Detail with User Status

```sql
-- Get problem detail with user's status and best submission
SELECT 
    p.*,
    pc.name AS category_name,
    pc.slug AS category_slug,
    pc.color AS category_color,
    ARRAY_AGG(DISTINCT pt.tag_name) FILTER (WHERE pt.tag_name IS NOT NULL) AS tags,
    ups.status AS user_status,
    ups.best_score AS user_best_score,
    ups.attempts AS user_attempts,
    ups.solved_at AS user_solved_at,
    ups.hints_used_count,
    (SELECT json_agg(json_build_object(
        'id', tc.id,
        'input', tc.input,
        'expected_output', tc.expected_output,
        'description', tc.description
    ) ORDER BY tc.id)
    FROM test_cases tc 
    WHERE tc.problem_id = p.id AND tc.is_sample = TRUE
    ) AS sample_test_cases
FROM problems p
JOIN problem_categories pc ON p.category_id = pc.id
LEFT JOIN problem_tags pt ON p.id = pt.problem_id
LEFT JOIN user_problem_status ups ON p.id = ups.problem_id AND ups.user_id = $2
WHERE p.slug = $1 AND p.is_published = TRUE
GROUP BY p.id, pc.name, pc.slug, pc.color, ups.status, ups.best_score, ups.attempts, ups.solved_at, ups.hints_used_count;
```

---

## Redis Cache Strategy

### Key Patterns

| Cache Type | Key Pattern | TTL | Description |
|------------|-------------|-----|-------------|
| User Session | `session:{token}` | 24h | JWT/session data |
| User Profile | `user:{user_id}` | 1h | User profile data |
| User Stats | `user:stats:{user_id}` | 5m | Dashboard statistics |
| Problem Detail | `problem:{slug}` | 30m | Problem full detail |
| Problem List | `problems:list:{filters_hash}` | 10m | Filtered problem list |
| Problem Count | `problems:count:{difficulty}:{category}` | 10m | Total problem count |
| Leaderboard Weekly | `leaderboard:weekly` | 1m | Weekly top 100 |
| Leaderboard Monthly | `leaderboard:monthly` | 5m | Monthly top 100 |
| Leaderboard All-Time | `leaderboard:alltime` | 10m | All-time top 100 |
| User Rank | `user:rank:{user_id}` | 1m | User's current rank |
| Submission Result | `submission:{submission_id}` | 5m | Submission result |
| Submission Queue | `submissions:pending` | - | Pending queue length |
| Hot Problems | `problems:hot` | 5m | Trending problems |
| Category Stats | `category:stats:{category_id}` | 10m | Category statistics |
| User Progress | `user:progress:{user_id}` | 5m | User progress cache |
| Rate Limit | `ratelimit:{user_id}:{action}` | 1m | Rate limiting |

### TTL Settings

```bash
# Session & Auth
session:*           → 86400 (24 hours)
refresh_token:*     → 604800 (7 days)

# User Data
user:*              → 3600 (1 hour)
user:stats:*        → 300 (5 minutes)
user:progress:*     → 300 (5 minutes)
user:rank:*         → 60 (1 minute)

# Problem Data
problem:*           → 1800 (30 minutes)
problems:list:*     → 600 (10 minutes)
problems:count:*    → 600 (10 minutes)
problems:hot        → 300 (5 minutes)

# Leaderboard
leaderboard:weekly  → 60 (1 minute)
leaderboard:monthly → 300 (5 minutes)
leaderboard:alltime → 600 (10 minutes)

# Submissions
submission:*        → 300 (5 minutes)
submissions:pending → volatile (no TTL, updated by worker)

# Rate Limiting
ratelimit:*         → 60 (1 minute)
```

### Cache Invalidation Strategy

```bash
# 1. Write-Through Cache
# Saat data di-update, cache langsung di-update juga
# Contoh: Update user profile → DEL user:{user_id}

# 2. Write-Behind Cache (Lazy Invalidation)
# Cache dihapus, query berikutnya akan populate ulang
# Contoh: New submission → DEL user:stats:{user_id}, DEL leaderboard:*

# 3. TTL-Based Expiration
# Cache otomatis expired berdasarkan TTL
# Cocok untuk data yang jarang berubah

# 4. Event-Driven Invalidation
# Menggunakan Redis Pub/Sub untuk broadcast invalidation
```

### Cache Invalidation Events

| Event | Keys to Invalidate |
|-------|-------------------|
| New submission | `user:stats:{user_id}`, `leaderboard:*`, `problems:list:*` |
| Submission completed | `submission:{submission_id}`, `user:progress:{user_id}` |
| Problem created/updated | `problem:{slug}`, `problems:list:*`, `problems:count:*` |
| User profile updated | `user:{user_id}` |
| Hint used | `user:progress:{user_id}`, `problem:{slug}` |
| Weekly reset | `leaderboard:weekly`, `user:rank:*` |
| Monthly reset | `leaderboard:monthly` |

### Pub/Sub for Real-Time Leaderboard

```bash
# Channels
leaderboard:update      → Broadcast saat ada score berubah
leaderboard:weekly      → Weekly leaderboard updates
leaderboard:monthly     → Monthly leaderboard updates
submission:new          → New submission notification
submission:result       → Submission result ready

# Example: Publish leaderboard update
PUBLISH leaderboard:update '{"user_id": "uuid", "score": 150, "rank": 5}'

# Example: Subscribe to leaderboard
SUBSCRIBE leaderboard:update
SUBSCRIBE submission:result
```

### Redis Data Structures

```bash
# Leaderboard - Sorted Set
ZADD leaderboard:weekly {score} {user_id}
ZADD leaderboard:monthly {score} {user_id}
ZADD leaderboard:alltime {score} {user_id}

# Get top 10 weekly
ZREVRANGE leaderboard:weekly 0 9 WITHSCORES

# Get user rank
ZREVRANK leaderboard:weekly {user_id}

# Hot Problems - Sorted Set (by submission count)
ZADD problems:hot {submission_count} {problem_id}

# Rate Limiting - Sliding Window
# Key: ratelimit:{user_id}:{action}
# Value: counter with TTL

# Session - Hash
HSET session:{token} user_id {user_id} username {username} role {role}
EXPIRE session:{token} 86400
```

---

## Partitioning Strategy

### 1. Submissions Table — Partition by Month

**Alasan:** Tabel submissions tumbuh paling cepat (~500/hari = ~15,000/bulan). Partition memudahkan:
- Query berdasarkan waktu (monthly stats)
- Archive data lama
- Performa query yang lebih baik

```sql
-- Membuat tabel submissions yang di-partisi
CREATE TABLE submissions_partitioned (
    id                  UUID NOT NULL DEFAULT uuid_generate_v4(),
    user_id             UUID NOT NULL,
    problem_id          INTEGER NOT NULL,
    code                TEXT NOT NULL,
    language            VARCHAR(20) DEFAULT 'go',
    status              submission_status DEFAULT 'pending',
    score               INTEGER DEFAULT 0,
    execution_time_ms   INTEGER,
    memory_used_kb      INTEGER,
    test_cases_passed   INTEGER DEFAULT 0,
    test_cases_total    INTEGER DEFAULT 0,
    error_message       TEXT,
    submitted_at        TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    completed_at        TIMESTAMPTZ,
    PRIMARY KEY (id, submitted_at)
) PARTITION BY RANGE (submitted_at);

-- Membuat partition per bulan (contoh 2025)
CREATE TABLE submissions_y2025m01 PARTITION OF submissions_partitioned
    FOR VALUES FROM ('2025-01-01') TO ('2025-02-01');
CREATE TABLE submissions_y2025m02 PARTITION OF submissions_partitioned
    FOR VALUES FROM ('2025-02-01') TO ('2025-03-01');
CREATE TABLE submissions_y2025m03 PARTITION OF submissions_partitioned
    FOR VALUES FROM ('2025-03-01') TO ('2025-04-01');
CREATE TABLE submissions_y2025m04 PARTITION OF submissions_partitioned
    FOR VALUES FROM ('2025-04-01') TO ('2025-05-01');
CREATE TABLE submissions_y2025m05 PARTITION OF submissions_partitioned
    FOR VALUES FROM ('2025-05-01') TO ('2025-06-01');
CREATE TABLE submissions_y2025m06 PARTITION OF submissions_partitioned
    FOR VALUES FROM ('2025-06-01') TO ('2025-07-01');
CREATE TABLE submissions_y2025m07 PARTITION OF submissions_partitioned
    FOR VALUES FROM ('2025-07-01') TO ('2025-08-01');
CREATE TABLE submissions_y2025m08 PARTITION OF submissions_partitioned
    FOR VALUES FROM ('2025-08-01') TO ('2025-09-01');
CREATE TABLE submissions_y2025m09 PARTITION OF submissions_partitioned
    FOR VALUES FROM ('2025-09-01') TO ('2025-10-01');
CREATE TABLE submissions_y2025m10 PARTITION OF submissions_partitioned
    FOR VALUES FROM ('2025-10-01') TO ('2025-11-01');
CREATE TABLE submissions_y2025m11 PARTITION OF submissions_partitioned
    FOR VALUES FROM ('2025-11-01') TO ('2025-12-01');
CREATE TABLE submissions_y2025m12 PARTITION OF submissions_partitioned
    FOR VALUES FROM ('2025-12-01') TO ('2026-01-01');

-- Function untuk auto-create partition berikutnya
CREATE OR REPLACE FUNCTION create_monthly_partition()
RETURNS void AS $$
DECLARE
    partition_date DATE;
    partition_name TEXT;
    start_date DATE;
    end_date DATE;
BEGIN
    partition_date := DATE_TRUNC('month', NOW() + INTERVAL '1 month');
    partition_name := 'submissions_y' || TO_CHAR(partition_date, 'YYYY') || 'm' || TO_CHAR(partition_date, 'MM');
    start_date := partition_date;
    end_date := partition_date + INTERVAL '1 month';
    
    EXECUTE format(
        'CREATE TABLE IF NOT EXISTS %I PARTITION OF submissions_partitioned FOR VALUES FROM (%L) TO (%L)',
        partition_name, start_date, end_date
    );
END;
$$ LANGUAGE plpgsql;
```

### 2. Leaderboard History — Partition by Contest Period

**Alasan:** Leaderboard history di-reset per periode (mingguan/bulanan). Partition berdasarkan periode memudahkan:
- Query history per kuartal
- Archive data lama
- Reset score tanpa menghapus data

```sql
-- Partition per kuartal
CREATE TABLE leaderboard_history_2025_q1 PARTITION OF leaderboard_history
    FOR VALUES FROM ('2025-01-01') TO ('2025-04-01');
CREATE TABLE leaderboard_history_2025_q2 PARTITION OF leaderboard_history
    FOR VALUES FROM ('2025-04-01') TO ('2025-07-01');
CREATE TABLE leaderboard_history_2025_q3 PARTITION OF leaderboard_history
    FOR VALUES FROM ('2025-07-01') TO ('2025-10-01');
CREATE TABLE leaderboard_history_2025_q4 PARTITION OF leaderboard_history
    FOR VALUES FROM ('2025-10-01') TO ('2026-01-01');
```

### 3. User Rating History — Partition by Quarter

```sql
CREATE TABLE user_rating_history_partitioned (
    id              BIGSERIAL,
    user_id         UUID NOT NULL,
    old_rating      INTEGER NOT NULL,
    new_rating      INTEGER NOT NULL,
    rating_change   INTEGER NOT NULL,
    reason          VARCHAR(100),
    reference_id    UUID,
    created_at      TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    PRIMARY KEY (id, created_at)
) PARTITION BY RANGE (created_at);
```

### 4. Submission Test Results — Partition by Month

```sql
CREATE TABLE submission_test_results_partitioned (
    id                  SERIAL,
    submission_id       UUID NOT NULL,
    test_case_id        INTEGER NOT NULL,
    status              test_result_status NOT NULL,
    execution_time_ms   INTEGER,
    memory_used_kb      INTEGER,
    actual_output       TEXT,
    error_message       TEXT,
    created_at          TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    PRIMARY KEY (id, created_at)
) PARTITION BY RANGE (created_at);
```

### Partition Management

```sql
-- Detach old partition (untuk archive)
ALTER TABLE submissions_partitioned DETACH PARTITION submissions_y2024m01;

-- Attach existing table as partition
ALTER TABLE submissions_partitioned ATTACH PARTITION submissions_y2026m01
    FOR VALUES FROM ('2026-01-01') TO ('2026-02-01');

-- Drop old partition (setelah archive)
DROP TABLE submissions_y2024m01;
```

---

## Maintenance & Monitoring

### Routine Maintenance

```sql
-- Analyze tables (update statistics)
ANALYZE users;
ANALYZE problems;
ANALYZE submissions;
ANALYZE leaderboard;

-- Reindex (jika ada bloat)
REINDEX TABLE CONCURRENTLY submissions;
REINDEX INDEX CONCURRENTLY idx_submissions_pending;

-- Vacuum (reclaim storage)
VACUUM ANALYZE submissions;
```

### Monitoring Queries

```sql
-- Table sizes
SELECT 
    schemaname,
    tablename,
    pg_size_pretty(pg_total_relation_size(schemaname||'.'||tablename)) AS size,
    n_live_tup AS row_count
FROM pg_stat_user_tables
ORDER BY pg_total_relation_size(schemaname||'.'||tablename) DESC;

-- Index usage
SELECT 
    indexrelname AS index_name,
    idx_scan AS times_used,
    pg_size_pretty(pg_relation_size(indexrelid)) AS index_size
FROM pg_stat_user_indexes
WHERE schemaname = 'public'
ORDER BY idx_scan DESC;

-- Slow queries
SELECT 
    query,
    calls,
    mean_time,
    total_time
FROM pg_stat_statements
ORDER BY mean_time DESC
LIMIT 20;
```

### Backup Strategy

```bash
# Daily full backup
pg_dump -Fc coding_challange > backup_$(date +%Y%m%d).dump

# Continuous archiving (WAL)
# Configure postgresql.conf:
# archive_mode = on
# archive_command = 'cp %p /backup/wal/%f'

# Point-in-time recovery (PITR)
pg_restore -d coding_challange backup_20260626.dump
```

---

## Catatan Implementasi

### Migrasi dari Non-Partitioned ke Partitioned

1. Buat tabel baru dengan partitioning
2. Copy data dari tabel lama ke tabel baru
3. Rename tables secara atomic
4. Recreate indexes dan constraints

```sql
BEGIN;
ALTER TABLE submissions RENAME TO submissions_old;
ALTER TABLE submissions_partitioned RENAME TO submissions;
COMMIT;
```

### Connection Pooling

Gunakan **PgBouncer** untuk connection pooling:
- Max connections: 100
- Default pool size: 25
- Reserve pool size: 5

### Read Replicas

Untuk scale read-heavy queries (leaderboard, problem list):
- 1 primary (write)
- 2 replicas (read)
- Route read queries ke replicas

---

*Dokumen ini merupakan living document dan akan diperbarui seiring perkembangan platform.*
