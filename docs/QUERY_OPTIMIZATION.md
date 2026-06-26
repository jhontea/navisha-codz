# Query Optimization Report — Coding Challenge Platform

> **Versi:** 1.0.0  
> **Database:** PostgreSQL 15+  
> **Tanggal:** Juni 2026  
> **Tujuan:** EXPLAIN ANALYZE untuk query berat dan strategi optimasi

---

## Daftar Isi

1. [Metodologi](#metodologi)
2. [Query 1: Problem Listing dengan Filter](#query-1-problem-listing-dengan-filter)
3. [Query 2: Submission History User](#query-2-submission-history-user)
4. [Query 3: Test Case Retrieval by Problem](#query-3-test-case-retrieval-by-problem)
5. [Query 4: User Dashboard Stats](#query-4-user-dashboard-stats)
6. [Query 5: Leaderboard Ranking](#query-5-leaderboard-ranking)
7. [Query 6: Success Rate per Problem](#query-6-success-rate-per-problem)
8. [Query 7: Submissions per Day (Aggregasi)](#query-7-submissions-per-day-aggregasi)
9. [Query 8: Hot Problems (Trending)](#query-8-hot-problems-trending)
10. [Rekomendasi Index Baru](#rekomendasi-index-baru)
11. [Monitoring Plan](#monitoring-plan)

---

## Metodologi

Semua EXPLAIN ANALYZE disimulasikan dengan:
- **Dataset:** 500 problems, 50,000 submissions, 10,000 users, 2,500 test cases
- **Settings:** `seq_page_cost = 1.0`, `random_page_cost = 4.0`, `work_mem = 4MB`
- **Cache:** Cold cache (buffer pool kosong) — `track_io_timing = ON`
- **Format:** JSON EXPLAIN (BUFFERS, ANALYZE, SETTINGS)

### Simbol EXPLAIN

| Simbol | Arti | Dampak |
|--------|------|--------|
| `Seq Scan` | Full table scan | Buruk untuk tabel besar |
| `Index Scan` | Index lookup by value | Baik untuk selectivity tinggi |
| `Index Only Scan` | Semua data di index | Optimal — tanpa heap lookup |
| `Bitmap Heap Scan` | Index → bitmap → heap | Baik untuk selectivity sedang |
| `Sort (Mem)` | Sortir di memory | OK jika memory cukup |
| `Sort (Disk)` | Sortir di disk | **Perlu optimasi** |
| `Nested Loop` | Loop join | Baik untuk small inner |
| `Hash Join` | Hash-based join | Baik untuk medium-large |
| `Merge Join` | Sorted merge join | Baik untuk sorted inputs |

---

## Query 1: Problem Listing dengan Filter

### Query
```sql
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
LIMIT 20 OFFSET 0;
```

### EXPLAIN ANALYZE (Sebelum Index)

```
 Aggregate  (cost=450.32..480.15 rows=20 width=120) (actual time=45.2..48.1 rows=20)
   ->  Gather Merge  (cost=420.10..450.00 rows=20 width=120) (actual time=40.5..45.0 rows=20)
         Workers Planned: 2
         Workers Launched: 2
         ->  Sort  (cost=420.10..420.15 rows=20 width=120) (actual time=38.2..38.3 rows=20)
               Sort Key: p.difficulty, p.title
               Sort Method: quicksort  Memory: 32kB
               ->  Parallel Hash Left Join  (cost=180.20..420.00 rows=20 width=120)
                     ->  Hash Join  (cost=150.10..380.00 rows=50 width=80)
                           ->  Parallel Seq Scan on problems p
                                 Filter: (is_published = TRUE AND ...)
                                 Rows Removed by Filter: 450
                           ->  Hash  (cost=15.00..15.00 rows=500 width=20)
                                 ->  Seq Scan on problem_categories pc
                     ->  Hash  (cost=25.00..25.00 rows=1500 width=16)
                           ->  Seq Scan on problem_tags pt
 Planning Time: 2.1 ms
 Execution Time: 50.3 ms
```

**Masalah:**
1. `Parallel Seq Scan` pada problems — full scan 500 baris, filter menghapus 450 baris (90% discard)
2. `Hash Join` build untuk problem_tags — full scan 1500 baris
3. No index digunakan untuk filter `difficulty` + `category_id`

### EXPLAIN ANALYZE (Sesudah Index: `idx_problems_listing`)

```
 Limit  (cost=2.45..12.50 rows=20 width=120) (actual time=0.35..1.20 rows=20)
   ->  GroupAggregate  (cost=2.45..250.00 rows=500 width=120)
         Group Key: p.difficulty, p.title
         ->  Nested Loop Left Join  (cost=2.45..240.00 rows=500 width=120)
               ->  Nested Loop  (cost=2.00..120.00 rows=50 width=80)
                     ->  Index Scan using idx_problems_listing on problems p
                           Index Cond: (difficulty = 'easy'::problem_difficulty)
                           Filter: (is_published = TRUE)
                           Rows Removed by Filter: 0
                     ->  Index Scan using idx_problem_categories_slug on problem_categories pc
                           Index Cond: (id = p.category_id)
               ->  Index Scan using idx_problem_tags_problem_id on problem_tags pt
                     Index Cond: (problem_id = p.id)
 Planning Time: 1.5 ms
 Execution Time: 1.8 ms
```

**Perbaikan:**
- `Index Scan` → hanya membaca problem dengan difficulty yang sesuai
- Filter `is_published = TRUE` hanya 0 baris discard (semua published sudah di-partial index)
- Execution time: **50.3 ms → 1.8 ms** (≈28× lebih cepat)

### Partial Index Complement

Query juga terbantu oleh `idx_problems_is_published ON problems(is_published) WHERE is_published = TRUE`:

```sql
EXPLAIN ANALYZE
SELECT COUNT(*) FROM problems WHERE is_published = TRUE AND difficulty = 'hard';
```

```
 Aggregate  (cost=4.50..4.51 rows=1 width=8) (actual time=0.08..0.09 rows=1)
   ->  Index Only Scan using idx_problems_is_published on problems p
         Index Cond: (difficulty = 'hard'::problem_difficulty)
         Heap Fetches: 0
 Planning Time: 0.3 ms
 Execution Time: 0.1 ms
```

**Catatan:** Index `idx_problems_is_published` adalah partial index yang hanya mencakup published problems. Saat dikombinasikan dengan query listing, PostgreSQL bisa melakukan BitmapAnd untuk menggabungkan partial index dan `idx_problems_listing`.

---

## Query 2: Submission History User

### Query
```sql
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
LIMIT 20 OFFSET 0;
```

### EXPLAIN ANALYZE (Sebelum Index `idx_submissions_history`)

```
 Limit  (cost=0.56..10.50 rows=20 width=200) (actual time=0.80..3.20 rows=20)
   ->  Nested Loop  (cost=0.56..25000.00 rows=50000 width=200)
         ->  Index Scan Backward using idx_submissions_submitted_at on submissions s
               ->  Index Scan using idx_submissions_user_id
                     Index Cond: (user_id = $1)
                     Filter: (user_id = $1)
         ->  Index Scan using problems_pkey on problems pr
               Index Cond: (id = s.problem_id)
 Planning Time: 1.2 ms
 Execution Time: 4.5 ms
```

**Masalah:**
- Index `idx_submissions_user_id` hanya filter `user_id`, tapi ORDER BY `submitted_at DESC` memerlukan sort terpisah.
- Backward Index Scan pada `idx_submissions_submitted_at` lalu filter `user_id` — tidak efisien.
- Jika user memiliki ribuan submission, perlu sort manual.

### EXPLAIN ANALYZE (Sesudah Index: `idx_submissions_history`)

```
 Limit  (cost=0.43..4.50 rows=20 width=200) (actual time=0.15..0.80 rows=20)
   ->  Nested Loop  (cost=0.43..500.00 rows=5000 width=200)
         ->  Index Scan Backward using idx_submissions_history on submissions s
               Index Cond: (user_id = $1)
               ->  Index Scan using problems_pkey on problems pr
                     Index Cond: (id = s.problem_id)
 Planning Time: 0.8 ms
 Execution Time: 1.1 ms
```

**Perbaikan:**
- `Index Scan using idx_submissions_history` langsung membaca submission dalam urutan `(user_id, status, submitted_at DESC)`.
- ORDER BY tidak perlu sort — index sudah dalam urutan DESC.
- Execution time: **4.5 ms → 1.1 ms** (≈4× lebih cepat).

### Filter by Status

Jika query juga menyertakan filter status:

```sql
WHERE s.user_id = $1 AND s.status = 'accepted'
ORDER BY s.submitted_at DESC
```

```
 Limit  (cost=0.43..2.50 rows=10 width=200) (actual time=0.10..0.40 rows=10)
   ->  Index Scan Backward using idx_submissions_history on submissions s
         Index Cond: (user_id = $1 AND status = 'accepted'::submission_status)
         ->  Index Scan using problems_pkey on problems pr
               Index Cond: (id = s.problem_id)
 Planning Time: 0.6 ms
 Execution Time: 0.6 ms
```

Index 3-column `(user_id, status, submitted_at DESC)` mencakup filter lengkap + sorting tanpa sort tambahan.

---

## Query 3: Test Case Retrieval by Problem

### Query
```sql
-- Get non-hidden test cases for a problem
SELECT id, input, expected_output, description, weight
FROM test_cases
WHERE problem_id = $1 AND is_hidden = FALSE
ORDER BY id;

-- Get all test cases for a problem (hidden + non-hidden)
SELECT id, input, expected_output, description, weight
FROM test_cases
WHERE problem_id = $1
ORDER BY id;
```

### EXPLAIN ANALYZE (Sebelum Index `idx_test_cases_problem_visibility`)

```
 Seq Scan on test_cases  (cost=0.00..45.00 rows=5 width=200) (actual time=2.5..5.8 rows=5)
   Filter: (problem_id = $1 AND is_hidden = FALSE)
   Rows Removed by Filter: 495
 Planning Time: 0.5 ms
 Execution Time: 6.0 ms
```

**Masalah:** Full scan 500 baris, hanya 5 baris yang lolos filter (99% discard).

### EXPLAIN ANALYZE (Sesudah Index: `idx_test_cases_problem_visibility`)

```
 Index Scan using idx_test_cases_problem_visibility on test_cases
   (cost=0.15..8.22 rows=5 width=200) (actual time=0.05..0.12 rows=5)
   Index Cond: (problem_id = $1 AND is_hidden = FALSE)
 Planning Time: 0.2 ms
 Execution Time: 0.18 ms
```

**Perbaikan:** **6.0 ms → 0.18 ms** (≈33× lebih cepat) — index memfilter langsung ke 5 baris yang relevan tanpa scan.

### Perbandingan dengan Partial Index

Partial index `idx_test_cases_is_hidden WHERE is_hidden = FALSE`:

```
 Index Scan using idx_test_cases_is_hidden on test_cases
   (cost=0.15..4.50 rows=5 width=200) (actual time=0.04..0.10 rows=5)
   Index Cond: (problem_id = $1)
```

Partial index sedikit lebih ringkas (1 kolom) untuk query `is_hidden = FALSE` saja. Tapi **tidak bisa** melayani query yang butuh hidden test cases (`is_hidden = TRUE` atau tanpa filter). Composite index `(problem_id, is_hidden)` mencakup **semua** varian query.

---

## Query 4: User Dashboard Stats

### Query
```sql
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
    (SELECT COUNT(DISTINCT problem_id) FROM user_problem_status 
     WHERE user_id = u.id AND status = 'solved') AS solved_problems,
    (SELECT COUNT(DISTINCT problem_id) FROM user_problem_status 
     WHERE user_id = u.id) AS attempted_problems
FROM users u
LEFT JOIN leaderboard lb ON u.id = lb.user_id
WHERE u.id = $1;
```

### EXPLAIN ANALYZE

```
 Nested Loop Left Join  (cost=1.50..25.00 rows=1 width=200) (actual time=0.20..0.50 rows=1)
   ->  Index Scan using users_pkey on users u
         Index Cond: (id = $1)
   ->  Index Scan using leaderboard_user_id_key on leaderboard lb
         Index Cond: (user_id = $1)
   SubPlan 1
     ->  Index Only Scan using idx_problems_is_published on problems
           (cost=0.25..4.50 rows=1 width=8) (actual time=0.05..0.05 rows=1)
   SubPlan 2
     ->  Index Only Scan using idx_user_problem_status_user_status on user_problem_status
           Index Cond: (user_id = $1 AND status = 'solved')
           (cost=0.25..4.00 rows=1 width=8) (actual time=0.05..0.08 rows=1)
   SubPlan 3
     ->  Index Only Scan using idx_user_problem_status_user_id on user_problem_status
           Index Cond: (user_id = $1)
           (cost=0.25..3.50 rows=10 width=8) (actual time=0.03..0.06 rows=1)
 Planning Time: 1.0 ms
 Execution Time: 0.8 ms
```

**Status:** ✅ Optimal. Semua subquery menggunakan Index Scan / Index Only Scan. Execution time < 1 ms.

**Index yang terlibat:**
- `users_pkey` — PK lookup
- `leaderboard_user_id_key` — UNIQUE constraint
- `idx_problems_is_published` — partial index untuk count
- `idx_user_problem_status_user_status` — composite index untuk user+status
- `idx_user_problem_status_user_id` — FK index

---

## Query 5: Leaderboard Ranking

### Query
```sql
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
LIMIT 100;
```

### EXPLAIN ANALYZE

```
 Limit  (cost=0.28..8.50 rows=100 width=150) (actual time=0.10..0.80 rows=100)
   ->  Nested Loop  (cost=0.28..85.00 rows=1000 width=150)
         ->  Index Scan using idx_leaderboard_weekly_rank on leaderboard lb
               Index Cond: (weekly_rank IS NOT NULL)
               (cost=0.28..25.00 rows=1000 width=16) (actual time=0.05..0.30 rows=100)
         ->  Index Scan using users_pkey on users u
               Index Cond: (id = lb.user_id)
               (cost=0.15..0.22 rows=1 width=150) (actual time=0.01..0.01 rows=1)
 Planning Time: 0.6 ms
 Execution Time: 1.0 ms
```

**Status:** ✅ Optimal. Leaderboard adalah cache table — join cepat via PK dan index rank.

**Catatan:** Leaderboard sudah di-cache di Redis via sorted set (`ZREVRANGE leaderboard:weekly 0 99 WITHSCORES`). Query ini hanya fallback untuk cache miss.

---

## Query 6: Success Rate per Problem

### Query
```sql
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

### EXPLAIN ANALYZE

```
 Sort  (cost=2500.00..2600.00 rows=40000 width=120) (actual time=120.0..125.0 rows=500)
   Sort Key: (ROUND(...) DESC)
   Sort Method: quicksort  Memory: 64kB
   ->  HashAggregate  (cost=500.00..1000.00 rows=40000 width=120)
         Group Key: p.id
         ->  Hash Right Join  (cost=50.00..400.00 rows=50000 width=80)
               ->  Seq Scan on submissions s
                     (cost=0.00..300.00 rows=50000 width=16)
               ->  Hash  (cost=30.00..30.00 rows=500 width=64)
                     ->  Hash Join  (cost=5.00..30.00 rows=500 width=64)
                           ->  Index Scan using idx_problems_is_published on problems p
                                 (cost=0.25..15.00 rows=500 width=56)
                           ->  Hash  (cost=3.00..3.00 rows=100 width=8)
                                 ->  Seq Scan on problem_categories pc
                                       (cost=0.00..3.00 rows=100 width=8)
 Planning Time: 1.5 ms
 Execution Time: 128.0 ms
```

**Masalah:**
- Full scan pada `submissions` (50,000 baris) — LEFT JOIN tanpa filter menyebabkan Hash Right Join
- HashAggregate perlu memproses 40,000+ baris

**Optimasi:**
1. **Materialized view** — success rate jarang berubah, bisa di-refresh periodik
2. **Partial aggregation** — gunakan subquery untuk menghitung stats per problem

### Optimized Query (dengan subquery)

```sql
SELECT 
    p.id,
    p.title,
    p.slug,
    p.difficulty,
    pc.name AS category_name,
    COALESCE(s.submission_count, 0) AS total_submissions,
    COALESCE(s.accepted_count, 0) AS accepted_submissions,
    CASE 
        WHEN COALESCE(s.submission_count, 0) > 0 
        THEN ROUND(COALESCE(s.accepted_count, 0) * 100.0 / COALESCE(s.submission_count, 0), 2)
        ELSE 0 
    END AS success_rate
FROM problems p
JOIN problem_categories pc ON p.category_id = pc.id
LEFT JOIN (
    SELECT 
        problem_id,
        COUNT(*) AS submission_count,
        COUNT(*) FILTER (WHERE status = 'accepted') AS accepted_count
    FROM submissions
    GROUP BY problem_id
) s ON p.id = s.problem_id
WHERE p.is_published = TRUE
ORDER BY success_rate DESC;
```

```
 Sort  (cost=350.00..360.00 rows=500 width=120) (actual time=15.0..16.0 rows=500)
   Sort Key: (ROUND(...) DESC)
   Sort Method: quicksort  Memory: 64kB
   ->  Hash Right Join  (cost=50.00..320.00 rows=500 width=120)
         ->  HashAggregate  (cost=50.00..80.00 rows=500 width=16)
               Group Key: s.problem_id
               ->  Index Scan using idx_submissions_problem_id on submissions s
                     (cost=0.25..40.00 rows=50000 width=8)
         ->  Hash  (cost=30.00..30.00 rows=500 width=64)
               ->  Hash Join (p + pc)
 Planning Time: 1.2 ms
 Execution Time: 18.0 ms
```

**Perbaikan:** **128 ms → 18 ms** (≈7× lebih cepat). Subquery agregasi hanya perlu scan submissions sekali.

**Rekomendasi tambahan:** Materialized view untuk production:

```sql
CREATE MATERIALIZED VIEW mv_problem_success_rate AS
SELECT 
    p.id AS problem_id,
    p.title,
    p.slug,
    p.difficulty,
    pc.name AS category_name,
    COUNT(s.id) AS total_submissions,
    COUNT(s.id) FILTER (WHERE s.status = 'accepted') AS accepted_submissions
FROM problems p
JOIN problem_categories pc ON p.category_id = pc.id
LEFT JOIN submissions s ON p.id = s.problem_id
WHERE p.is_published = TRUE
GROUP BY p.id, pc.name;

CREATE UNIQUE INDEX idx_mv_problem_success_rate ON mv_problem_success_rate(problem_id);

-- Refresh every 5 minutes
REFRESH MATERIALIZED VIEW CONCURRENTLY mv_problem_success_rate;
```

---

## Query 7: Submissions per Day (Aggregasi)

### Query
```sql
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

### EXPLAIN ANALYZE

```
 GroupAggregate  (cost=0.56..2000.00 rows=30 width=64) (actual time=10.0..45.0 rows=30)
   Group Key: (date(submitted_at))
   ->  Index Scan using idx_submissions_submitted_at on submissions
         Index Cond: (submitted_at >= $1 AND submitted_at < $2)
         (cost=0.56..1500.00 rows=50000 width=16) (actual time=0.10..8.0 rows=15000)
 Planning Time: 0.8 ms
 Execution Time: 48.0 ms
```

**Status:** ✅ Cukup baik. Index range scan pada `submitted_at` membatasi baris yang diproses. Untuk 15,000 baris dalam range, GroupAggregate 48 ms acceptable.

**Optimasi tambahan:**
- **Partition pruning** — jika submissions di-partisi per bulan, query hanya scan partition yang relevan.
- **Index tambahan** — `(submitted_at, status, user_id)` untuk index-only scan.

---

## Query 8: Hot Problems (Trending)

### Query
```sql
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
LIMIT 20;
```

### EXPLAIN ANALYZE

```
 Limit  (cost=500.00..520.00 rows=20 width=120) (actual time=300.0..305.0 rows=20)
   Sort  (cost=500.00..510.00 rows=500 width=120)
     Sort Key: (... FILTER ...) DESC, COUNT(s.id) DESC
     ->  HashAggregate  (cost=200.00..450.00 rows=500 width=120)
           Group Key: p.id
           ->  Hash Right Join  (cost=50.00..400.00 rows=50000 width=80)
                 ->  Seq Scan on submissions s
                       (cost=0.00..300.00 rows=50000 width=24)
                 ->  Hash (p + pc join)
 Planning Time: 1.0 ms
 Execution Time: 310.0 ms
```

**Masalah:** Full scan submissions (50,000 baris), HashAggregate berat.

**Optimasi:** Partial aggregation dengan window function atau subquery:

```sql
WITH recent_submissions AS (
    SELECT problem_id, COUNT(*) AS cnt
    FROM submissions
    WHERE submitted_at > NOW() - INTERVAL '7 days'
    GROUP BY problem_id
),
all_submissions AS (
    SELECT problem_id, COUNT(*) AS total, COUNT(DISTINCT user_id) AS unique_users
    FROM submissions
    GROUP BY problem_id
)
SELECT 
    p.id,
    p.title,
    p.slug,
    p.difficulty,
    pc.name AS category_name,
    COALESCE(a.total, 0) AS submission_count,
    COALESCE(a.unique_users, 0) AS unique_users,
    COALESCE(r.cnt, 0) AS submissions_last_7_days
FROM problems p
JOIN problem_categories pc ON p.category_id = pc.id
LEFT JOIN all_submissions a ON p.id = a.problem_id
LEFT JOIN recent_submissions r ON p.id = r.problem_id
WHERE p.is_published = TRUE
ORDER BY submissions_last_7_days DESC, submission_count DESC
LIMIT 20;
```

```
 Limit  (cost=120.00..140.00 rows=20 width=120) (actual time=15.0..18.0 rows=20)
   ->  Hash Right Join  (cost=120.00..250.00 rows=500 width=120)
         ->  HashAggregate  (cost=30.00..50.00 rows=500 width=24)
               ->  Index Scan using idx_submissions_submitted_at on submissions
                     Index Cond: (submitted_at > NOW() - '7 days'::interval)
         ->  Hash  (cost=80.00..80.00 rows=500 width=100)
               ->  Hash Right Join  (cost=30.00..80.00 rows=500 width=100)
                     ->  HashAggregate (all submissions)
                           ->  Index Scan using idx_submissions_problem_id
 Planning Time: 1.0 ms
 Execution Time: 19.0 ms
```

**Perbaikan:** **310 ms → 19 ms** (≈16× lebih cepat). CTE memisahkan agregasi 7-hari (range kecil) dari agregasi total.

---

## Rekomendasi Index Baru

### Ringkasan

| Nama Index | Tabel | Kolom | Type | Execution Time Before | After | Speedup |
|------------|-------|-------|------|----------------------|-------|---------|
| `idx_problems_listing` | problems | (difficulty, category_id, title) | Composite B-tree | 50.3 ms | 1.8 ms | **28×** |
| `idx_submissions_history` | submissions | (user_id, status, submitted_at DESC) | Composite DESC | 4.5 ms | 1.1 ms | **4×** |
| `idx_test_cases_problem_visibility` | test_cases | (problem_id, is_hidden) | Composite B-tree | 6.0 ms | 0.18 ms | **33×** |

### DDL Statements

```sql
-- Problem Listing
CREATE INDEX CONCURRENTLY idx_problems_listing 
ON problems(difficulty, category_id, title);

-- Submission History
CREATE INDEX CONCURRENTLY idx_submissions_history 
ON submissions(user_id, status, submitted_at DESC);

-- Test Case Visibility
CREATE INDEX CONCURRENTLY idx_test_cases_problem_visibility 
ON test_cases(problem_id, is_hidden);
```

> **Catatan:** Gunakan `CONCURRENTLY` di production untuk menghindari lock table selama index creation.

### Drop Rekomendasi (Jika Index Redundan)

```sql
-- Jika composite index sudah mencakup workload:
-- DROP INDEX CONCURRENTLY idx_test_cases_is_hidden;
-- Index ini sudah dicakup oleh idx_test_cases_problem_visibility
```

---

## Monitoring Plan

### 1. Index Usage Monitoring

```sql
-- Cek index usage per table
SELECT
    schemaname,
    tablename,
    indexname,
    idx_scan AS index_scans,
    idx_tup_read,
    idx_tup_fetch,
    pg_size_pretty(pg_relation_size(indexrelid)) AS index_size
FROM pg_stat_user_indexes
WHERE tablename IN ('problems', 'submissions', 'test_cases')
ORDER BY idx_scan ASC;
```

### 2. Query Performance Baseline

```sql
-- Track slow queries (enable pg_stat_statements)
SELECT 
    queryid,
    calls,
    ROUND(total_exec_time::numeric, 2) AS total_ms,
    ROUND(mean_exec_time::numeric, 2) AS avg_ms,
    ROUND(stddev_exec_time::numeric, 2) AS stddev_ms,
    rows,
    shared_blks_hit,
    shared_blks_read
FROM pg_stat_statements
WHERE query LIKE '%problems%' 
   OR query LIKE '%submissions%' 
   OR query LIKE '%test_cases%'
ORDER BY total_exec_time DESC
LIMIT 20;
```

### 3. Bloat Check

```sql
-- Check index bloat
SELECT
    indexrelid::regclass AS index_name,
    pg_size_pretty(pg_relation_size(indexrelid)) AS index_size,
    pg_size_pretty(pg_relation_size(indrelid)) AS table_size,
    ROUND(100 * pg_relation_size(indexrelid)::numeric / NULLIF(pg_relation_size(indrelid), 0), 2) AS ratio_pct
FROM pg_stat_user_indexes
WHERE tablename IN ('problems', 'submissions', 'test_cases')
ORDER BY pg_relation_size(indexrelid) DESC;
```

### 4. Vacuum & Analyze Schedule

```sql
-- Rekomendasi autovacuum settings untuk tabel heavy write
ALTER TABLE submissions SET (
    autovacuum_vacuum_scale_factor = 0.01,
    autovacuum_analyze_scale_factor = 0.005,
    autovacuum_vacuum_threshold = 1000
);

ALTER TABLE test_cases SET (
    autovacuum_vacuum_scale_factor = 0.05,
    autovacuum_analyze_scale_factor = 0.02
);
```

### SLA Targets

| Query | Target Time | Priority |
|-------|------------|----------|
| Problem listing (filtered) | < 10 ms | High |
| Submission history | < 5 ms | High |
| Test case retrieval | < 1 ms | High |
| User dashboard stats | < 5 ms | Medium |
| Leaderboard | < 5 ms | Medium |
| Success rate aggregation | < 50 ms | Low |
| Daily submission stats | < 100 ms | Low |
| Hot problems | < 50 ms | Low |

---

> **Referensi:**  
> - [PostgreSQL Index Types](https://www.postgresql.org/docs/15/indexes-types.html)  
> - [Index Scan vs Bitmap Scan](https://www.postgresql.org/docs/15/indexes-bitmap-scans.html)  
> - [pg_stat_statements](https://www.postgresql.org/docs/15/pgstatstatements.html)  
> - [Performance Tuning Guide](https://wiki.postgresql.org/wiki/Performance_Optimization)
