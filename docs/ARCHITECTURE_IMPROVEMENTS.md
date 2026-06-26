# Architecture V2 Improvement Review

> **Project**: Coding Challenge Platform  
> **Reviewer**: Architect  
> **Date**: 2026-06-26  
> **Reference**: docs/ARCHITECTURE_V2.md vs Actual Implementation  
> **Status**: Gap & Inconsistency Analysis

---

## Daftar Isi

1. [Executive Summary](#executive-summary)
2. [Gap Analysis](#gap-analysis)
3. [Inkronistencies](#inkronistencies)
4. [Dual Codebase Problem](#dual-codebase-problem)
5. [Rekomendasi](#rekomendasi)

---

## Executive Summary

Architecture V2 didesain sebagai arsitektur **microservices** yang sangat ambisius:
- 6 microservices terpisah (Auth, Problem, Execution, Leaderboard, Hint, Notification)
- Kubernetes sandbox untuk code execution
- PostgreSQL + Redis + RabbitMQ stack
- gRPC untuk inter-service communication
- React + Monaco Editor frontend
- Prometheus + Jaeger + ELK untuk observability

Namun implementasi aktual menunjukkan pola **monolith modular** dengan beberapaservice yang berkembang tidak konsisten, dual codebase pada execution layer, dan beberapa komponen V2 yang belum diimplementasikan.

---

## Gap Analysis

### 1. Komponen yang Belum Diimplementasikan

| Komponen V2 | Status Aktual | Keterangan |
|-------------|---------------|------------|
| **Notification Service** | TIDAK ADA | Di V2 direncanakan sebagai WebSocket server (port 8086). Implementasi hanya ada `pkg/websocket/hub.go` yang di-import oleh execution-service. |
| **K8s Sandbox** | TIDAK ADA | V2 merencanakan Kubernetes Pod per submission dengan NetworkPolicy. Implementasi aktual menggunakan Docker sandbox di `services/execution-worker/sandbox.go`. |
| **gRPC Communication** | TIDAK ADA | V2 merencanakan gRPC untuk Execution → Leaderboard dan Problem → Execution. Implementasi saat ini hanya REST + RabbitMQ. |
| **API Gateway (Kong/Nginx Ingress)** | PARTIAL | Ada `services/api-gateway/main.go` dengan circuit breaker + rate limiting, tapi masih dalam bentuk Go binary terpisah (bukan edge proxy). |
| **CDN + Static Assets** | TIDAK ADA | V2 merencanakan CloudFront/Cloudflare untuk React SPA. |
| **ELK Stack / Fluentd** | TIDAK ADA | V2 merencanakan log aggregation ke Elasticsearch. |
| **Distributed Tracing (Jaeger)** | TIDAK ADA | V2 merencanakan span propagation via B3 headers. |
| **React Frontend** | TIDAK ADA | V2 merencanakan React + Monaco Editor. Masih server-rendered HTML. |
| **PostgreSQL Read Replica** | TIDAK ADA | V2 merencanakan read replica untuk leaderboard queries. |
| **Partitioned Submissions Table** | SEBAGIAN | V2 merencanakan RANGE partitioning per bulan. Schema SQL ada dengan partitioning setup (comments), tapi tabel aktual TIDAK di-partisi. |
| **Redis Cluster** | TIDAK ADA | V2 merencankan Redis Cluster. Implementasi saat ini single Redis instance via `pkg/redis/client.go`. |

### 2. Arsitektur Sandbox

| Aspek | V2 Design | Actual Implementation |
|-------|-----------|----------------------|
| Runtime | Kubernetes Pod (namespace: sandbox) | Docker container via `docker exec` |
| Isolation | K8s NetworkPolicy (deny-all egress) | Docker `--network=none` + `--cap-drop=ALL` |
| Resource Limits | Per-difficulty (Easy/Medium/Hard) | Configurable via `WorkerConfig.GetLimitsForProblem()` |
| Scaling | HPA 3-50 replicas | Fixed worker slots via semaphore (`execSem`) |
| Cleanup | Worker deletes pod after execution | Docker `--rm` flag |
| DNS | Allow only DNS (port 53) | `--network=none` (no DNS at all) |

**Gap mayor**: Implementasi sandbox TIDAK menggunakan Kubernetes seperti yang direncanakan V2. Docker-only approach dengan fallback ke local execution membuat beberapa security guarantees lebih lemah.

### 3. Database Schema Inconsistencies

| V2 Schema | Aktual (schema.sql / migrations) |
|-----------|----------------------------------|
| `submissions.status` values: `queued`, `processing`, `running`, `completed`, `failed`, `timeout`, `runtime_error`, `compilation_error`, `oom` | Migration 003: `pending`, `queued`, `running`, `completed`, `failed`, `timeout`, `compilation_error` |
| `problems.difficulty` enum: `easy`, `medium`, `hard` | Migration 002: `easy`, `medium`, `hard`, `expert` |
| `problems.time_limit_ms` | Schema.sql: `time_limit_seconds` |
| `users.total_score`, `users.problems_solved` | Schema.sql: `rating`, `total_solved`, `total_submissions`, `streak_days` |
| `TEST_CASES` separate table | Embed as `test_case_json JSONB` in problems (migration 002) |
| `HINTS` with `user_hints` tracking via UUID | `hints_used` tracking with `user_id` + `problem_id` |
| `LEADERBOARD` with `contest_id` | Separate `weekly_score`, `monthly_score`, `all_time_score` columns |
| `REFRESH_TOKENS` with bcrypt hash | Schema.sql: `token_hash VARCHAR(255)` (plain, no bcrypt) |
| UUID primary key for users | Migration 001: consistent UUID |

---

## Inkronistencies

### A. Dual Codebase: `internal/service/` vs `services/`

Project memiliki DUA implementasi runner service:

1. **`internal/service/runner.go`** — Monolith-style, hardcoded to `golang:1.21-alpine` image, `sync.Pool` for buffers
   - Menggunakan `exec.CommandContext` dengan Docker
   - Fallback ke local execution dengan `DISABLE_LOCAL_FALLBACK` env var
   - `buildTestHarness()` inline

2. **`services/execution-worker/`** — Microservices-style, multi-environment sandbox
   - `sandbox.go`: Docker → containerd → local fallback chain
   - `harness.go`: Separate harness generation
   - `config.go`: `WorkerConfig.GetLimitsForProblem()`
   - Priority queue (3 priority + 2 normal workers)
   - Work stealing goroutine
   - Semaphore-based concurrency control
   - DLQ with exponential backoff

**Problem**: Kedua kode jalan secara paralel. `cmd/server/main.go` (monolith) menggunakan internal service, sementara `services/execution-worker/main.go` standalone service.

### B. Execution Service Model Mismatch

| `execution-service` model | `execution-worker` model |
|---------------------------|--------------------------|
| `ProblemID int` | `ProblemID int` |
| `Language` (go/python/javascript/java/cpp) | `Language` |
| `ExecutionMessage.Priority int` | `ExecutionMessage.Priority (omitempty)` |
| `Submission.Status`: pending → queued → completed | N/A |
| DLQ with 3 retries max | DLQ with 3 retries max |
| `QueueStats` struct | `executionMetrics` struct |
| No per-test-case metrics | `TestCaseMetrics` with CPU/DiskIO/NetworkIO |
| Single consumer per service | Priority + Normal + DLQ workers |

### C. Hint Service Route Mismatch

- **V2 API Design**: `GET /api/problems/:id/hints`, `GET /api/problems/:id/hints` (progressive)
- **Actual hint-service**: `/api/problems/:id/hints` + `/api/problems/:id/hints/:hintId/use` + `/api/problems/:id/hints/analytics` + `/api/problems/:id/hints/recommended`
- **Actual api-gateway routes**: `/api/hints/:problemId` (salah — seharusnya `/api/problems/:id/hints`)

### D. Auth Service Refresh Token Storage

- **V2 Design**: bcrypt hash refresh token + Redis
- **Actual**: bcrypt hash refresh token in PostgreSQL (`refresh_tokens` table), no Redis untuk refresh tokens

### E. Leaderboard Async Update

- **V2 Design**: Execution service publishes event to RabbitMQ → Leaderboard service consumes via gRPC → Updates ELO rating
- **Actual**: `leaderboard-service` has `internalUpdateScore` endpoint (POST /api/internal/update-score), tapi TIDAK ADA consumer yang menghubungkan execution-worker ke leaderboard.

### F. Missing Tables dari V2

V2 merencanakan tabel-tabel berikut yang TIDAK ada di schema:
- `TEST_CASES` separate table (embedded sebagai JSONB)
- `USER_HINTS` (pakai `hints_used`)
- `USER_STREAKS` (partial — ada di leaderboard-service queries tapi tidak di schema initial)
- `REFRESH_TOKENS` (ada tapi tanpa bcrypt)
- `problem_categories` (ada)

### G. Redis Cache Keys Mismatch

V2 defines:
```
problem:{id}    → 1 hour TTL
problems:list:{page}:{filter}  → 5 minutes
leaderboard:{contest_id}:{page} → 30 seconds
session:{token} → 24 hours
rate:{ip}:{endpoint} → 1 minute
hint:{problem_id}:{user_id} → 30 minutes
```

Actual (problem-service):
```
problem:{id}  → redis.TTLProblem
problems:list:{difficulty}:{category_id}:{page}:{search}:{page_size}  → redis.TTLProblemList
hint:{problem_id}:{user_id} → redis.TTLHint
```

Missing: `leaderboard:*`, `session:*`, `rate:*`

---

## Dual Codebase Problem

### Masalah Utama

```
cmd/server/monolith ──→ internal/service/RunnerService (Docker only)
                              ↓
                          model.Problem (YAML-based)

services/execution-worker/   ──→ SandboxExecutor (Docker + containerd + local)
                              ↓
                          WorkerConfig.GetLimitsForProblem()

services/execution-service/   ──→ Consumer + QueueStats + DLQHandler
services/leaderboard-service/ ──→ ELO + Streaks + Badges
services/hint-service/        ──→ Adaptive Hints + Analytics
services/auth-service/        ──→ JWT + PostgreSQL refresh tokens
services/problem-service/     ──→ Problem CRUD + Cache
services/api-gateway/         ──→ Reverse proxy + Circuit breaker
```

### Risiko

1. **Inconsistency bugs**: Fix di satu codebase tidak propag ke yang lain
2. **Maintenance burden**: 2x unit tests, 2x bug surface
3. **Confusing onboarding**: New developer tidak tahu mana yang "source of truth"
4. **Resource waste**: Docker image perlu build 2 versi berbeda

---

## Rekomendasi

### Prioritas Tinggi (Critical)

#### 1. Konsolidasi Dual Codebase
- **Keputusan**: Pilih salah satu approach. Rekomendasi: `services/` (microservices) sebagai primary.
- **Tindakan**: Hapus `internal/service/runner.go` dan `internal/model/problem.go` atau jangkau sebagai legacy adapter.
- **Alasan**: services/ sudah punya fitur lebih lengkap (priority queue, DLQ, work stealing, multi-environment sandbox).

#### 2. API Gateway Route Fix
- `/api/hints/:problemId` → `/api/problems/:id/hints` (sesuai V2 design)
- Atau update V2 design untuk match actual (lebih RESTful).

#### 3. Refresh Token Security
- Implementasi bcrypt hashing untuk refresh tokens (sudah dilakukan — schema.sql pakai `token_hash VARCHAR(255)` tapi actual auth-service stores bcrypt hash).
- Tambahkan Redis untuk session/blacklist sesuai V2.

#### 4. Connection: Execution-Worker → Leaderboard
- Worker menyelesaikan submission → publish completion event ke RabbitMQ
- Leaderboard service consume event → update ELO via `internalUpdateScore`
- Atau implementasi V2 design: direct gRPC call dari execution-service.

### Prioritas Sedang (Important)

#### 5. Standardisasi Database Schema
- Pilih SATU definisi schema: migration-based ATAU schema.sql-based.
- Schema.sql saat ini lebih lengkap tapi tidak dalam migration files.
- Migrations (migration 001-006) tidak match dengan schema.sql final.
- **Rekomendasi**: Buat migration baru yang mereflect schema.sql final, hapus schema.sql sebagai authoritative source.

#### 6. Unifikasi Model Problem
- `internal/model/problem.go`: YAML-based fields (id, title, type, difficulty, category, tags, description, examples, hints, template, test_cases, constraints, time/space_complexity)
- `services/problem-service`: DB-based fields (id, title, slug, description, difficulty, category_id, time_limit_seconds, memory_limit_mb, max_score, function_name, template_code, is_published)
- **Rekomendasi**: services/problem-service sebagai canonical. model.Problem untuk YAML import only.

#### 7. Hint Route Consistency
- Actual: `/api/problems/:id/hints` (lebih RESTful)
- V2: `GET /api/problems/:id/hints` (sudah correct)
- Update api-gateway proxy untuk match.

#### 8. Missing Tables / Features
- Tambahkan `user_streaks` table creation di migration (saat ini hanya di-query di leaderboard-service).
- Pastikan `rating_history` table ada di schema/migration.

### Prioritas Rendah (Nice-to-have)

#### 9. Observability Stack
- Implement structured logging (ada `pkg/logger/logger.go`)
- Tambahkan metrics endpoint `/metrics` di setiap service
- Atau skip entirely untuk MVP (single-machine deployment).

#### 10. Partitioning
- Untuk MVP dengan < 100K submissions, partitioning tidak urgent.
- Implement ketika table size > 10GB.

#### 11. Kubernetes Sandbox
- Docker sandbox sudah cukup untuk 1000 active users target.
- K8s add complexity without proportional benefit at current scale.
- Re-evaluate ketika: > 500 submissions/day consistently.

#### 12. Frontend Migration (React + Monaco)
- Server-rendered HTML + Alpine.js + CodeMirror sudah functional.
- React SPA add build complexity dan membutuhkan CDN.

---

## Summary Metrics

| Category | Implemented | Partial | Missing |
|----------|-------------|---------|---------|
| Microservices | 5/6 | 1 (notification) | 0 |
| Data Stores | 2/3 (PG + Redis) | 0 | RabbitMQ (imported tapi optional) |
| Security | JWT + bcrypt | Rate limiting | WAF, Vault, K8s NetworkPolicy |
| Observability | Logger | Health check | Prometheus, Jaeger, ELK |
| Frontend | Server-rendered | — | React SPA |
| Sandbox | Docker | Containerd fallback | K8s pods |
| Database | Single PG | Partitioning (schema only) | Read replicas, Cluster |

---

## Action Items

1. [ ] Merge atau pilih satu runner codebase (internal vs services/)
2. [ ] Fix api-gateway route: `/api/hints/:problemId` → `/api/problems/:id/hints`
3. [ ] Create migration final yang match schema.sql
4. [ ] Add `user_streaks` + `rating_history` ke migration files
5. [ ] Wire execution-worker completion event → leaderboard update
6. [ ] Redis session store untuk rate limiting
7. [ ] Remove atau deprecate `cmd/server/monolith` duplicate code
8. [ ] Document actual architecture (berdasarkan implementasi, bukan V2 design)

---

*Critical finding: V2 architecture design dan actual implementation sudah drift cukup signifikan. Rekomendasi utama adalah consolidate ke single source of truth dan update documentation untuk merefleksikan actual system.*
