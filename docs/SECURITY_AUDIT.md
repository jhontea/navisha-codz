# Security Audit Report

**Project:** Coding Challenge Platform
**Date:** 2025-06-26
**Auditor:** QA Engineer (Loop 29)
**Scope:** All Go services, middleware, handlers, repositories

---

## Summary

| Area | Severity | Status |
|------|----------|--------|
| SQL Injection | **LOW** | All queries use parameterized bindings |
| XSS | **MODERATE** | Gin auto-escapes JSON; CSP/Sanitization middleware not wired |
| CSRF | **HIGH** | CSRF middleware defined but NEVER used in any route |
| Sensitive Data | **HIGH** | Password format string leak; JWT secret in JSON-serializable struct |
| WebSocket Origin Bypass | **MODERATE** | WebSocket upgrader allows all origins |

**Overall Risk: HIGH** — CSRF protection is entirely absent on all mutation endpoints, and the database connection string format bug leaks passwords.

---

## 1. SQL Injection — LOW

### Finding: All SQL queries use parameterized bindings

Every SQL query in the codebase uses PostgreSQL `$N` placeholders via `pgxpool.QueryRow()` / `pgxpool.Exec()`. No string concatenation or `fmt.Sprintf` is used for query building.

**Affected services (verified safe):**
- `services/auth-service/main.go` — `$1` bindings for username, email, password
- `services/execution-service/main.go` — `$1` bindings
- `services/execution-worker/main.go` — `$1` bindings
- `services/leaderboard-service/main.go` — `$1` bindings
- `services/hint-service/main.go` — `$1` bindings
- `services/problem-service/main.go` — `$1` bindings
- `pkg/database/postgres.go` — `$1` bindings

### Recommendation
- Maintain current parameterized query discipline.
- The `SQLInjectionDetector` in `pkg/security/security.go` is a good defence-in-depth layer — consider wiring it into the gateway input sanitization middleware.

---

## 2. XSS — MODERATE

### Finding 2a: Gin auto-escapes JSON (safe for API responses)

All API responses use `c.JSON()`, which automatically HTML-escapes string values in JSON. This is safe for JSON APIs.

### Finding 2b: CSP and InputSanitization middleware are NOT wired

- `pkg/middleware/security.go` defines `SecurityHeaders()` with a strong CSP policy.
- `internal/middleware/middleware.go` defines `SecurityHeaders()` and `InputSanitization()` middleware.
- **Neither is used in `cmd/server/main.go`** or any service's route setup.

```go
// cmd/server/main.go — NO security middleware:
router := gin.Default()       // No custom SecurityHeaders()
router.GET("/health", ...)    // No CSP headers on responses
```

### Finding 2c: WebSocket allows all origins

```go
// pkg/websocket/hub.go:26-28
var Upgrader = websocket.Upgrader{
    CheckOrigin: func(r *http.Request) bool {
        return true  // ANY website can open a WS connection
    },
}
```

### Finding 2d: `websocket-service` CORS allows all origins

```go
// services/websocket-service/main.go:396
c.Header("Access-Control-Allow-Origin", "*")
```

### Recommendation
1. Wire `SecurityHeaders()` middleware into all Gin routers (especially `cmd/server/main.go`, `services/websocket-service/main.go`, all services).
2. Wire `InputSanitization()` middleware into the problem-service router.
3. Replace `CheckOrigin: return true` with a whitelist of allowed origins from environment variable.
4. Replace websocket-service's `*` CORS with specific allowed origins.

---

## 3. CSRF — HIGH (CRITICAL)

### Finding: CSRF middleware defined but NEVER used

`pkg/middleware/auth.go` defines `CSRFTokenMiddleware()` (lines 707-734) with:
- Token validation via `X-CSRF-Token` header + `csrf_token` cookie
- Constant-time comparison (`subtle.ConstantTimeCompare`)

**However, this middleware is NOT wired into any route in any service.**

All mutation endpoints are vulnerable to CSRF:

| Endpoint | Method | CSRF Protection |
|----------|--------|-----------------|
| `/api/auth/register` | POST | ❌ None |
| `/api/auth/login` | POST | ❌ None |
| `/api/auth/refresh` | POST | ❌ None |
| `/api/auth/logout` | POST | ❌ None |
| `/api/submissions` | POST | ❌ None |
| `/api/admin/problems` | POST | ❌ None |
| `/api/admin/problems/:id` | PUT | ❌ None |
| `/api/admin/problems/:id` | DELETE | ❌ None |
| `/api/validate` | POST | ❌ None |
| `/internal/broadcast` | POST | ❌ None |
| `/internal/notify/user/:userId` | POST | ❌ None |
| `/internal/notify/room/:roomName` | POST | ❌ None |

The `api-gateway` CORS middleware allows credentials:
```go
c.Writer.Header().Set("Access-Control-Allow-Credentials", "true")
```

This makes cookie-based CSRF attacks feasible if a session cookie mechanism were used.

### Recommendation
1. Wire `CSRFTokenMiddleware()` into all protected route groups that process state-changing operations.
2. Add a `GET /csrf-token` endpoint that returns a fresh CSRF token.
3. For the API gateway (`services/api-gateway/main.go`), add CSRF validation on the `protected` and `admin` route groups.
4. For internal service-to-service endpoints (`/internal/*`), use an internal API key header instead of or in addition to CSRF.

---

## 4. Sensitive Data Exposure — HIGH

### Finding 4a: Database connection string leaks password

**File:** `pkg/database/postgres.go` — lines 34-37

```go
connString := fmt.Sprintf(
    "postgres://%s:***@%s:%d/%s?sslmode=disable",
    cfg.User, cfg.Password, cfg.Host, cfg.Port, cfg.Database,
)
```

The `***` is literal text between `%s` format verbs. The actual argument order is:
1. `cfg.User` → first `%s` ✓
2. `cfg.Password` → second `%s` ← **PASSWORD APPEARS IN CONNECTION STRING AFTER `:***@`**
3. `cfg.Host` → `%d` ← type mismatch (string to int)
4. `cfg.Port` → `%s` ← type mismatch (int to string)
5. `cfg.Database` → final part

The intended masking is broken by incorrect format verb order. The resulting string would either panic at runtime or expose the password.

**Note:** `connString` is only used for `pgxpool.ParseConfig()` — it is never logged to stdout. But if it were (e.g., during debugging), the password would be visible.

### Finding 4b: JWT Secret in JSON-marshalable struct

**File:** `services/api-gateway/main.go` — lines 204-211

```go
type GatewayConfig struct {
    Port             int              `json:"port"`
    JWTSecret        string           `json:"jwt_secret"`   // 🚨 Exported, tagged
    RateLimitPerMin  int              `json:"rate_limit_per_min"`
    CircuitThreshold int              `json:"circuit_threshold"`
    CircuitTimeout   int              `json:"circuit_timeout_sec"`
    Services         []ServiceConfig  `json:"services"`
}
```

If this struct is ever JSON-marshaled (for logs, debugging, health endpoints), the `jwt_secret` value is included in the output.

### Finding 4c: Log entries include path + query parameters

- `services/api-gateway/main.go` — loggerMiddleware logs `path` and query string
- `services/websocket-service/main.go` — logs `path + "?" + query`
- `pkg/middleware/auth.go` — `LoggerMiddleware()` logs query parameters

If auth tokens or passwords are passed as query parameters (unlikely but possible), they would appear in logs.

### Recommendation
1. **Fix the connection string format immediately** — use `cfg.Password` only in `pgxpool.ParseConfig()` directly, never in a format string. Or mask it properly:
   ```go
   connString := fmt.Sprintf("postgres://%s:%s@%s:%d/%s?sslmode=disable",
       cfg.User, url.QueryEscape(cfg.Password), cfg.Host, cfg.Port, cfg.Database)
   ```
2. Change `JWTSecret` to `json:"-"` to prevent accidental serialization. Or create a separate config struct for JSON responses without the secret field.
3. Strip sensitive query parameters (tokens, passwords) before logging.
4. Consider using the structured `pkg/logger` instead of `log.Printf` across all services for consistent log-level control.

---

## 5. Additional Findings

### 5a. No HTTPS enforcement in development

- All services listen on HTTP without TLS.
- `Strict-Transport-Security` header is set by `pkg/middleware/security.go:53` but will have no effect over plain HTTP.
- **Risk:** Low for local development. Ensure production deployment uses a reverse proxy (nginx/Caddy) for TLS termination.

### 5b. JWT parsing in api-gateway uses `MapClaims`

**File:** `services/api-gateway/main.go:470`

```go
claims, ok := token.Claims.(jwt.MapClaims)
```

This is an untyped map access. Compare with `pkg/middleware/auth.go:127` which uses the typed `*JWTClaims` struct. Using `MapClaims` means no compile-time type safety on claim fields.

**Risk:** Low — values are still read from the map, but field name typos won't be caught.

### 5c. Rate limiting on problem-service is per-IP only

**File:** `internal/handler/problem.go:114`

```go
ip := c.ClientIP()
if !h.rateLimiter.AllowWithLimit(ip, 60, time.Minute) {
```

Per-IP rate limiting can be bypassed via IP spoofing or botnets. Consider per-user rate limiting for authenticated endpoints.

### 5d. Refresh token reuse detection

**File:** `pkg/middleware/auth.go:876-880`

The `ValidateAndRotate` function correctly detects refresh token reuse and revokes all tokens for that user. **Good security practice.**

### 5e. Password validation

**File:** `pkg/security/security.go:59-136`

Password policy enforcement exists but is **not called** in the auth-service registration handler. The `RegisterRequest` uses Gin's `min=8` binding but no character class or password strength checks.

**Risk:** Moderate — users can set weak passwords (e.g., "aaaaaaaa").

### 5f. No email verification

The registration endpoint immediately returns tokens without email verification. This allows automated account creation.

**Risk:** Low for a coding challenge platform; Medium for production.

---

## Security Score

| Category | Score |
|----------|-------|
| SQL Injection Protection | **A** ✅ |
| XSS Protection | **C** ⚠️ |
| CSRF Protection | **F** ❌ |
| Sensitive Data Handling | **C** ⚠️ |
| Authentication | **B** ✅ |
| Authorization (RBAC) | **B** ✅ |
| Rate Limiting | **B** ✅ |
| Input Validation | **C** ⚠️ |
| Output Encoding | **A** ✅ |
| Security Headers | **D** ⚠️ |

**Overall: C (Needs Improvement)**

---

## Priority Actions

1. **CRITICAL** — Wire `CSRFTokenMiddleware()` into all mutation endpoints
2. **CRITICAL** — Fix `postgres.go` connection string format to not leak passwords
3. **HIGH** — Set `json:"-"` on JWTSecret in GatewayConfig
4. **HIGH** — Wire `SecurityHeaders()` middleware into all service routers
5. **HIGH** — Replace WebSocket `CheckOrigin: return true` with origin whitelist
6. **MEDIUM** — Wire `InputSanitization()` into problem-service router
7. **MEDIUM** — Wire `PasswordPolicy` validation into auth-service register handler
8. **LOW** — Replace `jwt.MapClaims` with typed `*JWTClaims` in api-gateway
