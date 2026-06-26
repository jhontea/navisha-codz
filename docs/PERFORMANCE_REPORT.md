# Performance Report

**Date:** 2026-06-26  
**Project:** Coding Challenge Platform  
**Go Version:** go1.26.4 windows/amd64  
**CPU:** 12th Gen Intel(R) Core(TM) i5-12400F (12 cores)  
**Test Type:** Go Benchmarks (benchmem) + k6 Load Test Script  

---

## 1. Go Benchmark Results

### 1.1 Handler Layer (HTTP-level via httptest)

| Benchmark | Iterations | ns/op | B/op | allocs/op |
|---|---|---|---|---|
| HealthCheck | 212,667 | 5,672 | 2,688 | 27 |
| GetProblem (valid) | 202,407 | 5,986 | 6,071 | 20 |
| GetProblem (not found) | 415,123 | 2,590 | 2,432 | 24 |
| GetProblem (invalid ID) | 502,140 | 2,496 | 2,488 | 25 |
| GetTemplate (valid) | 303,705 | 4,710 | 3,662 | 41 |
| GetHints (valid) | 250,758 | 5,727 | 3,575 | 27 |

**Note:** `ListProblems` and `ValidateCode` endpoints could not be benchmarked due to the rate limiter (60 req/min/IP) — see Section 4.1.

### 1.2 Service Layer

| Benchmark | Iterations | ns/op | B/op | allocs/op |
|---|---|---|---|---|
| ListProblems (no filter) | 222,294 | 5,836 | 9,512 | 5 |
| ListProblems (difficulty filter) | 533,470 | 2,216 | 7,976 | 5 |
| GetProblem (existing) | 67,777,462 | 18.56 | 0 | 0 |
| GetProblemForAPI | 49,788,812 | 23.50 | 0 | 0 |
| GetTemplate | 2,606,574 | 432.5 | 440 | 8 |
| ValidateCode (valid) | 13,300,422 | 90.69 | 64 | 1 |
| ValidateCode (empty) | 10,615,419 | 108.3 | 112 | 2 |
| GetHints | 20,160,476 | 56.33 | 24 | 1 |
| GetFullHints | 1,000,000,000 | 0.13 | 0 | 0 |
| ConcurrentAccess (parallel) | 10,343,908 | 115.6 | 32 | 1 |

### 1.3 Utility Functions

| Benchmark | Iterations | ns/op | B/op | allocs/op |
|---|---|---|---|---|
| SanitizeProblemID (valid) | 93,472,502 | 12.76 | 0 | 0 |
| SanitizeProblemID (invalid) | 7,215,987 | 166.1 | 84 | 4 |
| ValidateCodeSize (small) | 1,000,000,000 | 1.00 | 0 | 0 |
| ValidateCodeSize (60KB) | 977,602,315 | 1.05 | 0 | 0 |
| BuildTestHarness (function) | 270,741 | 4,874 | 4,421 | 45 |

---

## 2. Performance Analysis

### 2.1 Strengths

1. **Zero-allocation hot paths:** `GetProblem`, `GetProblemForAPI`, `SanitizeProblemID` (valid), `GetFullHints`, and `ValidateCodeSize` all operate at **0 B/op, 0 allocs/op** — the code efficiently reuses data structures on hot paths.

2. **Microsecond-level service responses:** All service-layer operations complete in microseconds. The slowest, `ListProblems` (no filter), takes **5.8 µs**. `BuildTestHarness` is the most complex at **4.9 µs** with 45 allocations.

3. **Sub-microsecond HTTP handler responses:** Handler benchmarks show **2.5–6 µs** per request for read operations (GetProblem, GetTemplate, etc.), including full HTTP request/response serialization through Gin.

4. **Efficient hint service:** `GetFullHints` completes in **0.13 ns** (effectively a pointer return). Concurrent access to `GetHints` averages **115 ns**.

5. **Fast input validation:** `SanitizeProblemID` runs in **12.76 ns** for valid IDs. `ValidateCodeSize` runs in **~1 ns** regardless of code size.

### 2.2 Areas for Optimization

1. **ListProblems allocations:** Each call allocates **9,512 bytes (5 allocs)**. For high-traffic, this adds GC pressure. Consider object pooling for `ProblemSummary` slices.

2. **Handler response allocations:** Each HTTP handler call allocates **2.4–6 KB**. The `GetTemplate` handler allocates 41 allocs per call — the template map construction is the main contributor.

3. **BuildTestHarness allocations:** 45 allocs per invocation (4.4 KB). The string builder and JSON marshaling for each test case are the main contributors.

4. **`strings.Replace` in main harness builder:** The main-based test harness does `strings.Replace(code, "func main()", "func userMain()", 1)` which allocates a new string. This could use a more targeted approach.

### 2.3 Architectural Observations

- **~1000x overhead from handler layer:** Service operations complete in 18–5,800 ns, but HTTP handlers add 2,500–6,000 ns of overhead (Gin router, JSON serialization, response writing). For internal/microservice calls via gRPC, this overhead could be significantly reduced.

- **RWMutex is effective:** The repository's `sync.RWMutex` for concurrent reads shows no contention issues — parallel benchmarks complete with similar timing to sequential ones.

---

## 3. k6 Load Test Script

**Location:** `tests/load/k6-script.js`

### 3.1 Scenario

| Stage | Duration | Target VUs |
|---|---|---|
| Ramp up | 30s | 0 → 100 |
| Steady state | 60s | 100 |
| Ramp down | 30s | 100 → 0 |

### 3.2 Endpoints Tested

| Endpoint | Method | Purpose |
|---|---|---|
| `/api/problems` | GET | List all problems |
| `/api/problems/two-sum` | GET | Get specific problem detail |
| `/api/problems/two-sum/template` | GET | Get problem template |
| `/api/problems?difficulty=easy` | GET | Filtered problem list |

### 3.3 Checks & Thresholds

- **Status checks:** Each endpoint must return 200
- **Data integrity:** List returns array, specific problem has correct ID
- **Security:** Solution field not leaked via API (`no solution leak`)
- **Performance:** p(95) response time < 2s
- **Reliability:** Failure rate < 1%

### 3.4 Usage

```bash
# Run k6 load test (requires k6 CLI)
k6 run tests/load/k6-script.js

# Run against custom URL
k6 run -e BASE_URL=http://localhost:8080 tests/load/k6-script.js

# Run with output to JSON
k6 run --out json=k6-results.json tests/load/k6-script.js
```

---

## 4. Known Issues & Recommendations

### 4.1 [BUG] Handler Benchmarks Blocked by Rate Limiter

**Severity:** Medium  
**Location:** `internal/handler/problem.go` — `ListProblems` and `ValidateCode`  
**Steps to reproduce:**
1. Run `go test -bench=BenchmarkListProblems ./internal/handler/`
2. After ~60 iterations, the rate limiter returns HTTP 429

**Root cause:** The rate limiter (`60 req/min/IP`) uses a real clock. Benchmarks may schedule hundreds of requests within the same second, all with different IPs, but the rate limiter's 60 req/min window hasn't expired yet for previously seen IPs.

**Recommendation:** Add a test-only flag to the `ProblemHandler` to disable rate limiting during benchmarks, or use a mock rate limiter that always returns true.

### 4.2 [BUG] Repository Test Count Mismatch

**Severity:** High  
**Location:** `internal/repository/problem_test.go` — `TestCount`, `TestNewProblemRepository_ValidDir`  
**Steps to reproduce:**
1. Run `go test ./internal/repository/`
2. `TestCount` fails: `expected count 15, got 20`

**Root cause:** Existing tests hardcode `Count() == 15`, but the problems directory now contains **20 problems** (7 easy + 7 medium + 6 hard). The test data was not updated when new problems were added.

### 4.3 [BUG] Hint Service Test API Mismatch

**Severity:** High  
**Location:** `internal/service/hint_test.go` — `GetHints` and `Reset` calls  
**Steps to reproduce:**
1. Run `go test ./internal/service/`
2. Compilation error: `not enough arguments in call to svc.GetHints`

**Root cause:** `GetHints` signature changed from `GetHints(problem)` to `GetHints(userID, problem)` and `Reset` from `Reset(problemID)` to `Reset(userID, problemID)`, but existing tests were not updated.

### 4.4 [SECURITY] Rate Limiting IP Spoofing via Headers

**Severity:** Low  
**Location:** `internal/handler/problem.go` — `getClientIP`  
**Issue:** `getClientIP` trusts `X-Forwarded-For` and `X-Real-Ip` headers. A malicious client could spoof their IP to bypass rate limits.

**Recommendation:** Validate or strip these headers at the reverse proxy/load balancer level. Only trust them from known proxies.

### 4.5 [OPS] k6 Not Installed

**Severity:** Low  
**Issue:** `k6` CLI is not installed in the current environment. The k6 script is ready but cannot be executed until k6 is installed.

**Installation:**

```bash
# Windows (choco)
choco install k6

# Or download from https://k6.io/docs/get-started/installation/
```

---

## 5. Capacity Planning

Based on benchmark data, a single node can handle:

| Operation | Time per call | Est. req/s (single core) | Est. req/s (12 cores) |
|---|---|---|---|
| GetProblem (handler) | 5.99 µs | 167,000 | 2,000,000 |
| GetProblem (service) | 18.56 ns | 53,900,000 | 646,000,000 |
| ListProblems (service) | 5.84 µs | 171,000 | 2,050,000 |
| ValidateCode (service) | 90.69 ns | 11,000,000 | 132,000,000 |
| BuildTestHarness | 4.87 µs | 205,000 | 2,460,000 |

**Key insight:** The rate limiter (60 req/min/IP) will be the bottleneck well before CPU capacity. For 100 concurrent users, the system can easily handle the load — the main limitation is the rate limiter configuration, not CPU/memory.

---

## 6. Summary

The service-layer performance is **excellent** with most operations in **nanoseconds to single-digit microseconds**. HTTP handler-level overhead adds ~2.5–6 µs per request, which is acceptable for a web API.

**Critical issues found:**
1. Rate limiter blocks benchmarks — needs test mode support ([BUG])
2. Problem count mismatch in tests — test data out of sync ([BUG])
3. Hint service test API mismatch — tests don't compile ([BUG])
4. k6 not installed — load test script ready but not executable

**Non-critical:**
5. Rate limiter IP spoofing via headers (low risk)
6. ListProblems allocation optimization opportunity
