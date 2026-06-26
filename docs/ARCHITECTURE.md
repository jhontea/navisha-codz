# Architecture — Coding Challenge Website

## Overview

Website bank soal coding challenge algoritma & data structure berbasis Go dengan code editor dan sandboxed execution.

## Component Diagram

```
┌─────────────────────────────────────────────────────┐
│                    Browser (Client)                   │
│  ┌───────────┐  ┌──────────┐  ┌───────────────────┐  │
│  │ Problem    │  │ CodeMirror│  │ Test Results +    │  │
│  │ List       │  │ Editor    │  │ Hints Panel       │  │
│  └───────────┘  └──────────┘  └───────────────────┘  │
└──────────────────────┬──────────────────────────────┘
                       │ HTTP
┌──────────────────────▼──────────────────────────────┐
│              Go Backend (Gin)                         │
│  ┌──────────┐  ┌──────────┐  ┌──────────────────┐   │
│  │ Handler   │  │ Service   │  │ Repository       │   │
│  │ (HTTP)    │→ │ (Logic)   │→ │ (Problem Loader) │   │
│  └──────────┘  └────┬─────┘  └────────┬─────────┘   │
│                     │                  │              │
│                ┌────▼─────┐    ┌──────▼──────┐       │
│                │ Runner    │    │ YAML Files  │       │
│                │ Service   │    │ (problems/) │       │
│                └────┬─────┘    └─────────────┘       │
└─────────────────────┼───────────────────────────────┘
                      │
              ┌───────▼───────┐
              │ Docker Sandbox │
              │ (Go execution) │
              └───────────────┘
```

## Data Flow

1. User membuka problem list → GET /api/problems
2. User klik problem → GET /api/problems/:id
3. User menulis code di CodeMirror editor
4. User klik Submit → POST /api/problems/:id/run
5. Backend build Go program (user code + test harness)
6. Runner execute di Docker sandbox (timeout 10s)
7. Parse test output → return pass/fail per test case
8. Frontend display results
9. User bisa request hints → GET /api/problems/:id/hints

## Tech Stack

| Layer | Technology | Alasan |
|-------|-----------|--------|
| Backend | Go + Gin | Cepat, lightweight, user request Go |
| Frontend | HTML + Alpine.js + CodeMirror | Simple, no build step |
| Database | YAML files | Version-controllable, easy to add problems |
| Code Runner | Docker sandbox | Isolated execution, security |

## Project Structure

```
coding-challange/
├── cmd/server/main.go           # Entry point
├── internal/
│   ├── config/config.go         # App config
│   ├── model/problem.go         # Data models
│   ├── repository/problem.go    # Problem loader
│   ├── service/
│   │   ├── problem.go           # Problem logic
│   │   ├── runner.go            # Code execution
│   │   └── hint.go              # Hint logic
│   └── handler/                 # HTTP handlers
├── problems/                    # Problem bank (YAML)
│   ├── easy/
│   ├── medium/
│   └── hard/
├── web/
│   ├── templates/               # HTML templates
│   └── static/                  # CSS, JS
├── configs/app.yaml             # Config file
├── hermes/                      # Hermes agent setup
└── docs/                        # Documentation
```

---

## Data Models

### Problem

```go
type Problem struct {
    ID                    string      `yaml:"id" json:"id"`
    Title                 string      `yaml:"title" json:"title"`
    Difficulty            string      `yaml:"difficulty" json:"difficulty"`
    Category              string      `yaml:"category" json:"category"`
    Tags                  []string    `yaml:"tags" json:"tags"`
    Description           string      `yaml:"description" json:"description"`
    Examples              []Example   `yaml:"examples" json:"examples"`
    Hints                 []Hint      `yaml:"hints" json:"hints"`
    Template              string      `yaml:"template" json:"template"`
    TestCases             []TestCase  `yaml:"test_cases" json:"test_cases"`
    Constraints           []string    `yaml:"constraints" json:"constraints"`
    TimeComplexityHint    string      `yaml:"time_complexity_hint" json:"time_complexity_hint"`
    SpaceComplexityHint   string      `yaml:"space_complexity_hint" json:"space_complexity_hint"`
    Solution              *Solution   `yaml:"solution" json:"-"` // Hidden from API
}
```

### Example

```go
type Example struct {
    Input       string `yaml:"input" json:"input"`
    Output      string `yaml:"output" json:"output"`
    Explanation string `yaml:"explanation" json:"explanation"`
}
```

### Hint

```go
type Hint struct {
    Level   int    `yaml:"level" json:"level"`
    Title   string `yaml:"title" json:"title"`
    Content string `yaml:"content" json:"content"`
}
```

### TestCase

```go
type TestCase struct {
    Input       string `yaml:"input" json:"input"`
    Expected    string `yaml:"expected" json:"expected"`
    Description string `yaml:"description" json:"description"`
}
```

### Solution (Hidden)

```go
type Solution struct {
    Code             string `yaml:"code"`
    Approach         string `yaml:"approach"`
    TimeComplexity   string `yaml:"time_complexity"`
    SpaceComplexity  string `yaml:"space_complexity"`
}
```

### TestResult

```go
type TestResult struct {
    Name     string `json:"name"`
    Passed   bool   `json:"passed"`
    Expected string `json:"expected"`
    Actual   string `json:"actual"`
    Error    string `json:"error"`
}
```

### RunResponse

```go
type RunResponse struct {
    Success          bool          `json:"success"`
    CompilationError *string       `json:"compilation_error"`
    TestResults      []TestResult  `json:"test_results"`
    PassedCount      int           `json:"passed_count"`
    TotalCount       int           `json:"total_count"`
    ExecutionTimeMs  int64         `json:"execution_time_ms"`
}
```

---

## API Endpoint Summary

| Method | Endpoint | Description |
|--------|----------|-------------|
| `GET` | `/api/problems` | List all problems (summary). Query: `?difficulty=&category=` |
| `GET` | `/api/problems/:id` | Get problem detail (without solution) |
| `POST` | `/api/problems/:id/run` | Execute user code in sandbox. Body: `{"code": "..."}` |
| `GET` | `/api/problems/:id/hints` | Get hints for a problem |
| `GET` | `/health` | Health check |

> Full API documentation: [API.md](API.md)

---

## Problem YAML Schema Reference

Problems are stored as YAML files in `problems/{difficulty}/` directory.

> Full schema specification: [PROBLEM_SCHEMA.md](PROBLEM_SCHEMA.md)

Key conventions:
- File naming: `lowercase-with-dashes.yaml`
- Directory by difficulty: `easy/`, `medium/`, `hard/`
- `id` field must match filename (without `.yaml`)
- `solution` field is hidden from all API responses (`json:"-"`)

---

## Security Considerations

### Sandbox Isolation

| Measure | Implementation |
|---------|---------------|
| **Container isolation** | Each code execution runs in a fresh Docker container |
| **Network disabled** | `--network=none` flag prevents any outbound connections |
| **Resource limits** | CPU: 0.5 cores, Memory: 256MB, Timeout: 10 seconds |
| **Read-only filesystem** | Container root filesystem is read-only (tmpfs for `/tmp`) |
| **No privileged mode** | Container runs without `--privileged` flag |
| **Drop all capabilities** | `--cap-drop=ALL` removes all Linux capabilities |
| **User namespace** | Remap container UID to non-root host user |

### Input Validation

| Layer | Validation |
|-------|-----------|
| **Code size** | Max 64KB per submission |
| **Execution timeout** | Hard limit 10 seconds (SIGKILL after) |
| **Output size** | Max 1MB stdout/stderr captured |
| **Problem ID** | Validated against whitelist (alphanumeric + hyphens only) |
| **YAML parsing** | Strict unmarshaling with unknown field rejection |
| **Rate limiting** | 10 req/min per IP for `/run` endpoint |

### Code Restrictions (Sandbox-Level)

- No filesystem writes (read-only root)
- No network access
- No syscalls (seccomp profile)
- No fork/exec (limited `clone` flags)
- Memory limit prevents OOM attacks

### Potential Risks & Mitigations

| Risk | Mitigation |
|------|-----------|
| Fork bomb | `pids.limit=50` in Docker |
| Infinite loop | 10s timeout + SIGKILL |
| Memory bomb | 256MB memory limit |
| Storage exhaustion | Read-only fs + tmpfs limit |
| Privilege escalation | Non-root user + cap-drop |

---

## Deployment Considerations

### Docker Image

```dockerfile
FROM golang:1.22-alpine
WORKDIR /app
COPY . .
RUN go build -o server cmd/server/main.go
EXPOSE 8080
CMD ["./server"]
```

### Environment Variables

| Variable | Default | Description |
|----------|---------|-------------|
| `PORT` | `8080` | HTTP server port |
| `PROBLEMS_DIR` | `./problems` | Path to problem YAML files |
| `SANDBOX_TIMEOUT` | `10` | Execution timeout in seconds |
| `DOCKER_HOST` | `unix:///var/run/docker.sock` | Docker daemon address |
| `MAX_MEMORY_MB` | `256` | Sandbox memory limit |
| `LOG_LEVEL` | `info` | Log verbosity: debug, info, warn, error |

### Production Checklist

- [ ] Use reverse proxy (nginx/caddy) with TLS termination
- [ ] Set up Docker socket protection (rootless Docker or Docker socket proxy)
- [ ] Configure proper logging (structured JSON logs)
- [ ] Set up monitoring (Prometheus metrics at `/metrics`)
- [ ] Configure rate limiting at reverse proxy level
- [ ] Use read-only host filesystem where possible
- [ ] Run server as non-root user
- [ ] Set up health check orchestration (Kubernetes liveness/readiness)
- [ ] Backup `problems/` directory (Git-based version control)
- [ ] Configure CORS origins for production domain
- [ ] Set up alerting for sandbox failures (502 errors)

### Scaling Considerations

| Concern | Strategy |
|---------|----------|
| **Concurrent executions** | Worker pool with configurable concurrency limit |
| **Problem loading** | In-memory cache with file watcher for hot reload |
| **Static files** | Serve via CDN or nginx (not Go server) |
| **Session state** | Stateless design — no server-side session storage |
| **Problem bank growth** | Lazy loading with pagination for `/api/problems` |

---

## Data Models (Quick Reference)

### ProblemSummary (API list response)
```json
{
  "id": "two-sum",
  "title": "Two Sum",
  "difficulty": "easy",
  "category": "array",
  "tags": ["hash-map", "array"]
}
```

### RunResponse (Execution result)
```json
{
  "success": true,
  "compilation_error": null,
  "test_results": [
    {"name": "test_1", "passed": true, "expected": "[0,1]", "actual": "[0,1]", "error": ""},
    {"name": "test_2", "passed": false, "expected": "[1,2]", "actual": "[2,1]", "error": "output mismatch"}
  ],
  "passed_count": 1,
  "total_count": 2,
  "execution_time_ms": 145
}
```

### Error Response (Uniform format)
```json
{
  "error": "Human-readable message",
  "code": 400
}
```

---

## API Endpoint Summary (Quick Reference)

| Method | Endpoint | Status Codes | Description |
|--------|----------|-------------|-------------|
| `GET` | `/health` | 200 | Health check |
| `GET` | `/api/problems` | 200, 400 | List problems. Query: `?difficulty={easy\|medium\|hard}&category={name}` |
| `GET` | `/api/problems/:id` | 200, 400, 404 | Get problem detail (no solution exposed) |
| `POST` | `/api/problems/:id/run` | 200, 400, 404 | Execute code. Body: `{"code": "..."}` (max 64KB) |
| `GET` | `/api/problems/:id/hints` | 200, 400, 404 | Get all hints for a problem |

---

## Security Considerations (Summary)

### Implemented
- ✅ Docker sandbox with `--network=none`, `--read-only`, `--cap-drop=ALL`, `--pids-limit=50`
- ✅ Problem ID sanitization (alphanumeric + `-_` only)
- ✅ Code size limit (64KB)
- ✅ Solution field hidden from API (`json:"-"`)
- ✅ Execution timeout (10s)
- ✅ Memory limit (256MB)
- ✅ HTML escaping in frontend (XSS prevention)

### Needs Improvement
- ⚠️ Docker `--read-only` conflicts with `go build` — needs `--tmpfs /tmp` (HIGH, new)
- ⚠️ Progressive reveal not wired in handler — uses `GetFullHints` instead of `GetHints` (MEDIUM, new)
- ⚠️ Test cases (with expected outputs) exposed in problem detail API (MEDIUM)
- ⚠️ No rate limiting implemented (MEDIUM)
- ⚠️ CORS allows all origins (MEDIUM)
- ⚠️ No graceful shutdown (MEDIUM)

### Fixed in Iteration 2
- ✅ Local execution fallback now disabled by default via `DISABLE_LOCAL_FALLBACK` env var
- ✅ Race condition in HintService fixed with single lock pattern
- ✅ `go build` moved inside Docker container
- ✅ Unknown difficulty values now sort correctly (value 99)
- ✅ Output size limited to 1MB via `limitWriter`
- ✅ `BuildTestHarness` now generates functional test code
```

---

## Updated Component Diagram (v2 — Function & Main Support)

```
┌─────────────────────────────────────────────────────────────────┐
│                      Browser (Client)                             │
│  ┌───────────┐  ┌──────────────┐  ┌────────────────────────┐    │
│  │ Problem   │  │ CodeMirror   │  │ Test Results +         │    │
│  │ List      │  │ Editor       │  │ Hints Panel            │    │
│  │           │  │ (Go mode)    │  │                        │    │
│  └───────────┘  └──────────────┘  └────────────────────────┘    │
└──────────────────────────┬──────────────────────────────────────┘
                           │ HTTP / JSON
┌──────────────────────────▼──────────────────────────────────────┐
│                    Go Backend (Gin)                               │
│                                                                  │
│  ┌──────────────┐   ┌──────────────┐   ┌──────────────────┐     │
│  │ Handler      │   │ Service      │   │ Repository       │     │
│  │ (HTTP)       │──▶│ (Business)   │──▶│ (YAML Loader)    │     │
│  │              │   │              │   │                  │     │
│  │ - List       │   │ - Problem    │   │ - LoadAll()      │     │
│  │ - Detail     │   │ - Runner     │   │ - LoadByID()     │     │
│  │ - Template   │   │ - Validator  │   │ - Watch()        │     │
│  │ - Run        │   │ - Hints      │   └────────┬─────────┘     │
│  │ - Validate   │   │              │            │               │
│  │ - Hints      │   └──────┬───────┘   ┌────────▼─────────┐     │
│  └──────────────┘          │           │ YAML Files       │     │
│                            │           │ (problems/)      │     │
│                     ┌──────▼───────┐   │                  │     │
│                     │ Test Harness │   │ easy/            │     │
│                     │ Generator    │   │ medium/          │     │
│                     │              │   │ hard/            │     │
│                     │ - Function   │   └──────────────────┘     │
│                     │ - Main       │                             │
│                     └──────┬───────┘                             │
│                            │                                     │
└────────────────────────────┼─────────────────────────────────────┘
                             │
                     ┌───────▼────────┐
                     │ Docker Sandbox │
                     │ (Go execution) │
                     │                │
                     │ --network=none │
                     │ --read-only    │
                     │ --cap-drop=ALL │
                     │ --pids-limit=50│
                     └────────────────┘
```

---

## Test Execution Flow (Detailed)

### Function-Based Execution Flow

```
1. User clicks "Submit"
   │
   ▼
2. Frontend sends POST /api/problems/:id/run
   Body: { "code": "func twoSum(...) []int { ... }" }
   │
   ▼
3. Handler validates request
   - Check problem exists
   - Check code size ≤ 64KB
   - Check rate limit
   │
   ▼
4. Service loads problem YAML
   - Parse function_name, parameters, return_type
   - Parse test_cases (params + expected)
   │
   ▼
5. Test Harness Generator creates complete Go program:
   ┌─────────────────────────────────────────┐
   │ package main                             │
   │                                          │
   │ import (                                 │
   │     "encoding/json"                      │
   │     "fmt"                                │
   │     "reflect"                            │
   │ )                                        │
   │                                          │
   │ // === USER CODE ===                     │
   │ func twoSum(nums []int, target int) []int {│
   │     // User's implementation             │
   │ }                                        │
   │ // === END USER CODE ===                 │
   │                                          │
   │ func main() {                            │
   │     tests := []struct {                  │
   │         name     string                  │
   │         params   []interface{}           │
   │         expected interface{}             │
   │     }{                                   │
   │         {                                │
   │             name: "test_1",              │
   │             params: []interface{}{       │
   │                 []int{2,7,11,15}, 9,     │
   │             },                           │
   │             expected: []int{0, 1},       │
   │         },                               │
   │         // ... more tests                │
   │     }                                    │
   │                                          │
   │     for _, test := range tests {         │
   │         result := twoSum(                │
   │             test.params[0].([]int),      │
   │             test.params[1].(int),        │
   │         )                                │
   │         if reflect.DeepEqual(            │
   │             result, test.expected,       │
   │         ) {                              │
   │             fmt.Printf(                  │
   │                 "%s: PASS\n", test.name, │
   │             )                            │
   │         } else {                         │
   │             exp, _ := json.Marshal(      │
   │                 test.expected,           │
   │             )                            │
   │             act, _ := json.Marshal(      │
   │                 result,                  │
   │             )                            │
   │             fmt.Printf(                  │
   │                 "%s: FAIL exp=%s got=%s\n",│
   │                 test.name, exp, act,     │
   │             )                            │
   │         }                                │
   │     }                                    │
   │ }                                        │
   └─────────────────────────────────────────┘
   │
   ▼
6. Runner writes harness to temp dir
   │
   ▼
7. Docker sandbox executes:
   docker run \
     --network=none \
     --read-only \
     --tmpfs /tmp:rw,noexec,nosuid,size=50m \
     --cap-drop=ALL \
     --pids-limit=50 \
     --memory=256m \
     --cpus=0.5 \
     --timeout=10s \
     -v /tmp/sandbox:/app:ro \
     golang:1.22-alpine \
     go run /app/main.go
   │
   ▼
8. Parse stdout for test results
   - "test_1: PASS" → passed
   - "test_2: FAIL exp=... got=..." → failed
   │
   ▼
9. Return RunResponse JSON
```

### Main-Based Execution Flow

```
1. User clicks "Submit"
   │
   ▼
2. Frontend sends POST /api/problems/:id/run
   Body: { "code": "package main\n\nfunc main() { ... }" }
   │
   ▼
3. Handler validates request
   │
   ▼
4. Service loads problem YAML (type: main)
   │
   ▼
5. Runner compiles user code:
   docker run ... go build -o /tmp/binary /app/main.go
   │
   ▼
6. For each test case:
   docker run ... echo "$INPUT" | /tmp/binary
   │
   ▼
7. Compare stdout with expected
   │
   ▼
8. Return RunResponse JSON
```

---

## Data Models (Updated for Function Support)

### Problem (Updated)

```go
type Problem struct {
    ID                    string       `yaml:"id" json:"id"`
    Title                 string       `yaml:"title" json:"title"`
    Type                  ProblemType  `yaml:"type" json:"type"`        // NEW: "function" | "main"
    Difficulty            string       `yaml:"difficulty" json:"difficulty"`
    Category              string       `yaml:"category" json:"category"`
    Tags                  []string     `yaml:"tags" json:"tags"`
    Description           string       `yaml:"description" json:"description"`
    Examples              []Example    `yaml:"examples" json:"examples"`
    Hints                 []Hint       `yaml:"hints" json:"hints"`
    Template              string       `yaml:"template" json:"template"`
    TestCases             []TestCase   `yaml:"test_cases" json:"test_cases"`
    FunctionName          string       `yaml:"function_name" json:"function_name,omitempty"`  // NEW
    Parameters            []Parameter  `yaml:"parameters" json:"parameters,omitempty"`        // NEW
    ReturnType            string       `yaml:"return_type" json:"return_type,omitempty"`       // NEW
    Constraints           []string     `yaml:"constraints" json:"constraints"`
    TimeComplexityHint    string       `yaml:"time_complexity_hint" json:"time_complexity_hint"`
    SpaceComplexityHint   string       `yaml:"space_complexity_hint" json:"space_complexity_hint"`
    Solution              *Solution    `yaml:"solution" json:"-"`  // Hidden from API
}

type ProblemType string

const (
    ProblemTypeFunction ProblemType = "function"
    ProblemTypeMain     ProblemType = "main"
)
```

### Parameter (New)

```go
type Parameter struct {
    Name        string `yaml:"name" json:"name"`
    Type        string `yaml:"type" json:"type"`
    Description string `yaml:"description" json:"description,omitempty"`
}
```

### TestCase (Updated)

```go
type TestCase struct {
    Input       string      `yaml:"input" json:"input,omitempty"`           // For main-based
    Params      []any       `yaml:"params" json:"params,omitempty"`         // For function-based
    Expected    any         `yaml:"expected" json:"expected"`               // any type, not just string
    Description string      `yaml:"description" json:"description"`
}
```

### ValidationResult (New)

```go
type ValidationResult struct {
    Valid    bool              `json:"valid"`
    Errors   []ValidationError `json:"errors"`
    Warnings []ValidationError `json:"warnings"`
}

type ValidationError struct {
    Line     int    `json:"line"`
    Column   int    `json:"column"`
    Message  string `json:"message"`
    Severity string `json:"error"` // "error" or "warning"
}
```

---

## Security Considerations (Updated)

### Sandbox Isolation

| Measure | Implementation |
|---------|---------------|
| **Container isolation** | Each code execution runs in a fresh Docker container |
| **Network disabled** | `--network=none` flag prevents any outbound connections |
| **Resource limits** | CPU: 0.5 cores, Memory: 256MB, Timeout: 10 seconds |
| **Read-only filesystem** | Container root filesystem is read-only (tmpfs for `/tmp`) |
| **No privileged mode** | Container runs without `--privileged` flag |
| **Drop all capabilities** | `--cap-drop=ALL` removes all Linux capabilities |
| **User namespace** | Remap container UID to non-root host user |
| **PID limit** | `--pids-limit=50` prevents fork bombs |
| **Seccomp profile** | Default Docker seccomp blocks dangerous syscalls |

### Input Validation

| Layer | Validation |
|-------|-----------|
| **Code size** | Max 64KB per submission |
| **Execution timeout** | Hard limit 10 seconds (SIGKILL after) |
| **Output size** | Max 1MB stdout/stderr captured |
| **Problem ID** | Validated against whitelist (alphanumeric + hyphens only) |
| **YAML parsing** | Strict unmarshaling with unknown field rejection |
| **Rate limiting** | 10 req/min per IP for `/run`, 30 req/min for `/validate` |
| **Code injection** | Test harness is generated server-side, never from user input |
| **Template validation** | User code checked against expected function signature |

### Code Restrictions (Sandbox-Level)

- No filesystem writes (read-only root)
- No network access
- No syscalls (seccomp profile)
- No fork/exec (limited `clone` flags)
- Memory limit prevents OOM attacks
- No access to Docker socket

### Test Harness Security

| Risk | Mitigation |
|------|-----------|
| User code escapes harness | Harness uses `reflect.DeepEqual` — no eval/exec |
| Infinite loop in user code | 10s timeout + SIGKILL |
| User overrides test params | Params are hardcoded in generated harness |
| Output injection | Results parsed line-by-line, max 1MB |
| Import abuse | Only stdlib imports allowed in harness |

### Potential Risks & Mitigations

| Risk | Mitigation |
|------|-----------|
| Fork bomb | `pids.limit=50` in Docker |
| Infinite loop | 10s timeout + SIGKILL |
| Memory bomb | 256MB memory limit |
| Storage exhaustion | Read-only fs + tmpfs limit |
| Privilege escalation | Non-root user + cap-drop |
| Code exfiltration | No network, no filesystem writes |
| Docker escape | Keep Docker updated, use rootless mode |

### Security Checklist for Production

- [ ] Run Docker in rootless mode
- [ ] Use Docker socket proxy (e.g., Tecnativa/docker-socket-proxy)
- [ ] Enable AppArmor/SELinux profiles
- [ ] Set up seccomp custom profile
- [ ] Monitor container escape CVEs
- [ ] Regular Docker image updates
- [ ] Network segmentation for sandbox subnet
- [ ] Audit logging for all code executions
- [ ] Alert on abnormal execution patterns

---

## API Endpoint Summary (Updated)

| Method | Endpoint | Status Codes | Description |
|--------|----------|-------------|-------------|
| `GET` | `/health` | 200, 503 | Health check (includes sandbox status) |
| `GET` | `/api/problems` | 200, 400 | List problems. Query: `?difficulty={easy\|medium\|hard}&category={name}&type={function\|main}` |
| `GET` | `/api/problems/:id` | 200, 400, 404 | Get problem detail (no solution exposed) |
| `GET` | `/api/problems/:id/template` | 200, 404 | Get code template only (lightweight) |
| `POST` | `/api/problems/:id/run` | 200, 400, 404, 429, 502, 503 | Execute code. Body: `{"code": "..."}` (max 64KB) |
| `POST` | `/api/validate` | 200, 400 | Validate syntax without execution |
| `GET` | `/api/problems/:id/hints` | 200, 400, 404 | Get hints. Query: `?level={1\|2\|3}` |

---

## Performance Optimizations (Iteration 7)

### In-Memory Cache

| Feature | Implementation |
|---------|---------------|
| **Cache store** | `sync.RWMutex` protected map with TTL-based expiration |
| **TTL** | 5 minutes for problem list, configurable |
| **ETag support** | SHA-256 based ETags for conditional requests (304 Not Modified) |
| **Cache headers** | `X-Cache: HIT/MISS` header for debugging |
| **Auto-cleanup** | Background goroutine removes expired entries every 30s |

### Docker Image Reuse

| Optimization | Description |
|-------------|-------------|
| **Pre-built image** | `coding-challenge-sandbox:latest` built once via `Dockerfile.sandbox` |
| **No Go download** | Image includes pre-warmed Go module cache |
| **Non-root user** | Container runs as UID 1000 (runner) |
| **Minimal attack surface** | Removed curl, wget, nc from image |

### Connection Pooling

| Resource | Strategy |
|----------|----------|
| **HTTP server** | `ReadTimeout: 15s`, `WriteTimeout: 30s`, `IdleTimeout: 60s` |
| **Docker client** | Reuses Docker daemon connection via `exec.CommandContext` |
| **Buffer pooling** | `sync.Pool` for reusable stdout/stderr buffers |

---

## Security Hardening (Iteration 8)

### Security Headers

| Header | Value | Purpose |
|--------|-------|---------|
| `Content-Security-Policy` | `default-src 'self'; script-src 'self' 'unsafe-inline' https://cdn.jsdelivr.net; ...` | Prevents XSS, restricts resource loading |
| `X-Content-Type-Options` | `nosniff` | Prevents MIME type sniffing |
| `X-Frame-Options` | `DENY` | Prevents clickjacking |
| `X-XSS-Protection` | `1; mode=block` | Legacy XSS filter |
| `Strict-Transport-Security` | `max-age=31536000; includeSubDomains` | Enforces HTTPS |
| `Referrer-Policy` | `strict-origin-when-cross-origin` | Limits referrer leakage |
| `Permissions-Policy` | `camera=(), microphone=(), geolocation=()` | Restricts browser features |

### Request Protections

| Protection | Implementation |
|-----------|----------------|
| **Max body size** | 64KB limit via `http.MaxBytesReader` |
| **Input sanitization** | Rejects SQL injection, XSS, path traversal patterns |
| **Problem ID validation** | Only alphanumeric + `-_` allowed |
| **Rate limiting** | 10 req/min for `/run`, 60 req/min for list, 30 req/min for validate |
| **Path traversal prevention** | `SanitizeProblemID` rejects `../`, `..\\`, absolute paths |

### Request Timeout

| Layer | Timeout |
|-------|---------|
| **HTTP read** | 15 seconds |
| **HTTP write** | 30 seconds |
| **Sandbox execution** | 10 seconds (configurable) |
| **Graceful shutdown** | 30 seconds |

---

## UI Polish (Iteration 9)

### Features

| Feature | Description |
|---------|-------------|
| **Confetti animation** | Celebrates when all tests pass |
| **Syntax highlighting** | Go error messages highlighted with file:line:col |
| **Keyboard shortcuts** | `Ctrl+Enter` submit, `Ctrl+R` reset, `Ctrl+/` help |
| **Help panel** | Modal showing all keyboard shortcuts |
| **Touch-friendly** | 44px minimum touch targets on mobile |
| **Smooth animations** | Fade-in, slide-in, pulse effects |

### Responsive Breakpoints

| Breakpoint | Layout |
|-----------|--------|
| `> 1024px` | Two-column problem detail (info + editor) |
| `768-1024px` | Single column, stacked |
| `480-768px` | Mobile sidebar (hamburger), stacked filters |
| `< 480px` | Compact layout, smaller fonts |

---

## Final State Summary (Iteration 10)

### Test Coverage

| Package | Tests |
|---------|-------|
| `internal/repository` | 15 tests (loading, filtering, validation) |
| `internal/service` | 20+ tests (runner, hints, harness) |
| `internal/handler` | 20+ tests (HTTP endpoints, validation) |
| **Total** | **55+ tests** |

### Problem Bank

| Difficulty | Count | Examples |
|-----------|-------|---------|
| Easy | 4 | Two Sum, Fizz Buzz, Reverse String, Contains Duplicate, Max Subarray |
| Medium | 2 | Valid Parentheses, Merge Sorted Arrays |
| Hard | 3 | Coin Change, Trapping Rain Water, Longest Palindromic Substring |
| **Total** | **10** | |

### Build Verification

```bash
go build ./...   # ✅ Clean build
go vet ./...     # ✅ No issues
go test ./...    # ✅ All tests pass
```