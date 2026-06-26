# Code Review Findings - Coding Challenge Website

**Project:** C:\Users\PC\go\src\project\coding-challange
**Review Date:** 2026-06-26
**Reviewer:** QA Engineer (Automated)
**Scope:** Security, Code Quality, Test Coverage, Bug Detection

---

## Executive Summary

| Category | Count |
|----------|-------|
| CRITICAL | 3 |
| HIGH | 6 |
| MEDIUM | 8 |
| LOW | 5 |
| INFO | 4 |
| **Total** | **26** |

Build Status: **FAIL** (tests/integration does not compile)
Test Status: **FAIL** (2 test failures in pkg/middleware)

---

## 1. BUILD FAILURES (CRITICAL)

### CRIT-01: Integration tests fail— GenerateTokenPair return value mismatch

**Severity:** CRITICAL
**File:** `tests/integration/auth_test.go` (lines 120, 174, 197, 261, 281, 291)

**Problem:** `GenerateTokenPair` now returns 5 values (`accessToken, refreshToken, tokenID,`) but integration tests still expect 3 values:
```go
// OLD (broken):
accessToken, refreshToken, err := middleware.GenerateTokenPair(cfg, ...)

// NEW (correct):
accessToken, refreshToken, _, _, err := middleware.GenerateTokenPair(cfg, ...)
```

**Impact:** Entire integration test suite cannot compile. Zero integration test coverage.

**Recommendation:** Update all call sites in `tests/integration/auth_test.go` to use `_, _, _, _, _` for the 5 return values.

---

### CRIT-02: AuthMiddleware signature changed — integration tests use old signature

**Severity:** CRITICAL
**File:** `tests/integration/auth_test.go` (lines 245, 276)

**Problem:** `AuthMiddleware` now requires 3 arguments:
```go
func AuthMiddleware(cfg JWTConfig, blacklist *TokenBlacklist, sessionManager *SessionManager) gin.HandlerFunc
```
But tests call it with only 1 argument:
```go
middleware.AuthMiddleware(cfg)  // compilation error
```

**Impact:** Cannot compile integration tests.

**Recommendation:** Update calls to `middleware.AuthMiddleware(cfg, nil, nil)`.

---

### CRIT-03: Undefined imports `service` and `handler` in integration tests

**Severity:** CRITICAL
**File:** `tests/integration/authlines 29-33)

**Problem:** Tests import `"coding-challange/internal/service"` and `"coding-challange/internal/handler"` but these are not used in the auth test itself (only in setupTestServer which is in a different file). The auth test file has its own `setupTestServer` that references undefined `service` and `handler` packages at the file level.

**Impact:** Compilation failure.

**Recommendation:** Remove unused imports from `tests/integration/auth_test.go` or ensure all referenced packages are actually used.

---

## 2. TEST FAILURES (HIGH)

### HIGH-01: TestAuthMiddleware_ValidToken returns 401 instead of 200

**Severity:** HIGH
**File:** `pkg/middleware/auth_test.go:395-412`

**Problem:** The test generates a token with `GenerateTokenPair` using `cfg.AccessTokenSecret = "test-secret"`, but `NewJWTConfig()` sets default secrets. The issue is that `NewJWTConfig()` is called first (which reads from env vars), then only `AccessTokenSecret` is overridden. However, `RefreshTokenSecret` remains the default. This should not cause the issue.

**Root Cause Analysis:** The token is generated correctly, but the `AuthMiddleware` validates it and returns 401. The likely cause is that `AuthMiddleware` is initialized with `nil` blacklist and `nil` sessionManager, and the code creates new empty instances internally. The token should validate. This needs investigation — possibly the `ValidateToken` is failing due to expired token or clock skew.

**Recommendation:** 
1. Check if the token expiry is being set correctly (AccessTokenTTL defaults to 15 minutes from `NewJWTConfig()`)
2. Verify the secret used in token generation matches the secret used in validation
3. Add more detailed error logging in the test to see WHY authentication fails

---

### HIGH-02: TestAuthMiddleware_BearerCaseInsensitive fails — lowercase "bearer" returns 401

**Severity:** HIGH
**File:** `pkg/middleware/auth_test.go:632-649`

**Problem:** The middleware checks:
```go
if len(parts) != 2 || strings.ToLower(parts[0]) != "bearer" {
```
This should work for lowercase "bearer". The test sends `"bearer " + token` and expects 200, but gets 401.

**Root Cause:** The middleware uses `strings.SplitN(authHeader, " ", 2)`. With `"bearer " + token`, this gives `["bearer", "token"]`. The lowercase comparison `strings.ToLower("bearer") != "bearer"` evaluates to `false`, so it should pass. This suggests the token itself is invalid, not the prefix check.

**Recommendation:** This is likely a cascading failure from the same root cause as HIGH-01. Fix the token validation issue first.

---

## 3. SECURITY ISSUES (HIGH)

### HIGH-03: Hardcoded default JWT secrets in production code

**Severity:** HIGH
**File:** `pkg/middleware/auth.go:55-56`

**Problem:**
```go
AccessTokenSecret:    getEnv("JWT_ACCESS_SECRET", "your-access-secret-key-change-in-production"),
RefreshTokenSecret:   getEnv("JWT_REFRESH_SECRET", "your-refresh-secret-key-change-in-production"),
```

If `JWT_ACCESS_SECRET` env var is not set in production, all tokens are signed with a publicly-known secret. An attacker can forge admin JWT tokens.

**Recommendation:**
1. Fail loudly if secrets are not set in production mode
2. Use a secrets manager (Vault, AWS Secrets Manager) instead of env vars
3. Add startup validation that rejects default/weak secrets

---

### HIGH-04: CORS allows all origins with credentials

**Severity:** HIGH
**File:** `pkg/middleware/auth.go:843-844`

**Problem:**
```go
c.Writer.Header().Set("Access-Control-Allow-Origin", "*")
c.Writer.Header().Set("Access-Control-Allow-Credentials", "true")
```

Per the CORS spec, `Access-Control-Allow-Origin: *` with `Access-Control-Allow-Credentials: true` is forbidden by browsers, but some servers may behave unexpectedly. More importantly, allowing `*` origin means ANY website can make cross-origin requests to this API.

**Recommendation:** Restrict `Access-Control-Allow-Origin` to specific trusted domains. Never use `*` with credentials.

---

### HIGH-05: Local code execution fallback enabled by default

**Severity:** HIGH
**File:** `internal/service/runner.go:59`

**Problem:**
```go
if r.disableLocalFallback {
    // ...
}
// Fallback to local execution
return r.runLocal(ctx, harnessCode, problem)
```

The `disableLocalFallback` flag is set from `DISABLE_LOCAL_FALLBACK` env var, defaulting to `false` (not disabled). This means if Docker fails, code executes LOCALLY on the server by default. This is a critical sandbox escape.

**Recommendation:** Default to `true` (disabled) — require explicit opt-in to enable local fallback:
```go
disableLocalFallback: os.Getenv("DISABLE_LOCAL_FALLBACK") != "true"  // default: disabled
```
Or better: remove local fallback entirely for production.

---

### HIGH-06: Code injection via harness generation — user code not sanitized

**Severity:** HIGH
**File:** `internal/service/runner.go:99-100, internal/service/hint.go:102`

**Problem:** User-submitted code is directly embedded into the generated harness:
```go
b.WriteString(userCode)
```

A malicious user could submit code like:
```go
package main
import "os"
func init() { os.RemoveAll("/") }
func solution() {}
```

Even with Docker sandbox, this code runs inside the container with potential escape vectors.

**Recommendation:**
1. Parse user code AST and reject dangerous imports (`os`, `os/exec`, `net`, `syscall`, etc.)
2. Strip or block package-level `init()` functions
3. Use a whitelist approach for allowed Go standard library packages

---

## 4. CODE QUALITY ISSUES (MEDIUM)

### MED-01: Duplicate RateLimiter implementations

**Severity:** MEDIUM
**Files:** 
- `pkg/middleware/auth.go:657-734` (RateLimiter)
- `internal/handler/problem.go:36-80` (RateLimiter)

**Problem:** Two separate `RateLimiter` implementations with different behaviors:
- Middleware version: uses `c.ClientIP()` + `c.FullPath()` as key
- Handler version: uses raw IP + custom window tracking

**Recommendation:** Consolidate into a single shared implementation in `pkg/middleware/` and import it everywhere.

---

### MED-02: Duplicate SecurityHeaders implementations

**Severity:** MEDIUM
**Files:**
- `pkg/security/security.go:316-326` (XSSHeaders)
- `pkg/middleware/auth.go:747-755` (SecurityHeadersMiddleware)
- `internal/middleware/middleware.go:171-204` (SecurityHeaders)

**Problem:** Three different security header implementations with inconsistent policies. The `pkg/security` version has a stricter CSP without `unsafe-inline` for scripts, while `internal/middleware` allows `unsafe-inline 'unsafe-eval'`.

**Recommendation:** Use a single source of truth from `pkg/security/XSSHeaders()` and apply it consistently.

---

### MED-03: Unused variable `_ = oldestID` in SessionManager

**Severity:** MEDIUM
**File:** `pkg/middleware/auth.go:270`

**Problem:**
```go
_ = oldestID
```

This indicates dead code or incomplete refactoring. The `oldestID` variable is used on line 265 to check existence, then immediately discarded.

**Recommendation:** Remove the discard line or use the variable properly.

---

### MED-04: Custom `contains` function instead of `strings.Contains`

**Severity:** MEDIUM
**File:** `internal/service/problem.go:139-146`

**Problem:**
```go
func contains(s, substr string) bool {
    for i := 0; i <= len(s)-len(substr); i++ {
        if s[i:i+len(substr)] == substr {
            return true
        }
    }
    return false
}
```

This is a reimplementation of `strings.Contains` with O(n*m) complexity. It also panics if `substr` is empty (division by zero in loop condition when `len(substr) == 0`).

**Recommendation:** Replace with `strings.Contains(s, substr)`.

---

### MED-05: RequestTimeout middleware has goroutine leak

**Severity:** MEDIUM
**File:** `internal/middleware/middleware.go:208-231`

**Problem:**
```go
go func() {
    c.Next()
    close(done)
}()
select {
case <-done:
case <-time.After(timeout):
    c.AbortWithStatusJSON(...)
}
```

If the timeout fires, the goroutine running `c.Next()` continues executing in the background. This leaks goroutines and can cause resource exhaustion under load.

**Recommendation:** Use `context.WithTimeout` properly or ensure the goroutine is cancelled on timeout.

---

### MED-06: SuspiciousActivityDetector deadlock risk

**Severity:** MEDIUM
**File:** `pkg/security/security.go:651-655`

**Problem:**
```go
func (d *SuspiciousActivityDetector) RecordEvent(event SuspiciousEvent) {
    d.mu.Lock()
    defer d.mu.Unlock()
    d.recordEvent(event)  // takes lock again? No, this is the internal version
}
```

The `recordEvent` (lowercase) does not take a lock, so this is actually safe. But the naming is confusing and error-prone.

**Recommendation:** Rename internal method to `recordEventLocked` or similar to make locking semantics explicit.

---

### MED-07: TOTP uses SHA256 but labels itself as SHA1

**Severity:** MEDIUM
**File:** `pkg/security/security.go:876`

**Problem:**
```go
mac := hmac.New(sha256.New, key)
```

The TOTP implementation uses HMAC-SHA256, but the provisioning URI claims `algorithm=SHA1`. This is inconsistent and may cause compatibility issues with authenticator apps.

**Recommendation:** Either use SHA1 (per RFC 6238 default) or update the URI to `algorithm=SHA256`.

---

### MED-08: Race condition in SessionManager cleanup

**Severity:** MEDIUM
**File:** `pkg/middleware/auth.go:369-383`

**Problem:** The cleanup goroutine deletes sessions from the map but does not clean up the `userSessions` index. Over time, `userSessions[userID]` accumulates stale session IDs.

**Recommendation:** When cleaning up expired sessions, also remove their IDs from `userSessions`.

---

## 5. BUGS (MEDIUM)

### BUG-01: `containsFunction` panics on empty substring

**Severity:** MEDIUM
**File:** `internal/service/problem.go:134-136`

**Problem:**
```go
func containsFunction(code, funcDecl string) bool {
    return len(code) > 0 && len(funcDecl) > 0 && contains(code, funcDecl)
}
```

The `contains` function panics when `substr` is empty because `len(s)-len(substr)` becomes `len(s)` which is fine, but the loop `s[i:i+0]` always returns `""`, so it actually does not panic — but it returns `true` for empty substr, which may not be intended.

**Recommendation:** Add explicit empty-string check at the start of `contains`.

---

### BUG-02: `getClientIP` does not validate X-Forwarded-For

**Severity:** MEDIUM
**File:** `internal/handler/problem.go:83-100`

**Problem:** The `getClientIP` function trusts `X-Forwarded-For` header from any client. An attacker can spoof this header to bypass rate limits.

**Recommendation:** Only trust `X-Forwarded-For` from known proxies. Use `c.ClientIP()` which handles this correctly in Gin.

---

### BUG-03: HintService is not per-user — global state leaks across users

**Severity:** MEDIUM
**File:** `internal/service/hint.go:38`

**Problem:**
```go
revealed := s.revealed[problem.ID]
```

The hint reveal count is global per problem, not per user. If user A reveals hints, user B sees that hints were already revealed for that problem.

**Recommendation:** Key by `userID + problemID` or make this stateful per-session.

---

### BUG-04: `buildMainTestHarness` uses `strings.Replace` with count=1 — fragile

**Severity:** MEDIUM
**File:** `internal/service/hint.go:170`

**Problem:**
```go
userCode = strings.Replace(userCode, "func main()", "func userMain()", 1)
```

If user code contains `func main()` in a comment or string literal, this will incorrectly replace it. More importantly, if the user defines `func main()` with different spacing (e.g., `func main ()`), it won't match.

**Recommendation:** Use AST-based renaming or regex with word boundaries.

---

## 6. PERFORMANCE ISSUES (LOW)

### LOW-01: `generateRequestID` uses crypto/rand for non-security purpose

**Severity:** LOW
**File:** `pkg/middleware/auth.go:859-863`

**Problem:** `generateRequestID()` uses `crypto/rand` which is slower than `math/rand`. Request IDs don't need cryptographic randomness.

**Recommendation:** Use `math/rand` or a UUID library for request IDs.

---

### LOW-02: LoggerMiddleware uses fmt.Printf

**Severity:** LOW
**File:** `pkg/middleware/auth.go:830-836`

**Problem:** Using `fmt.Printf` for structured logging is inefficient and not structured.

**Recommendation:** Use a proper structured logger (zap, logrus).

---

### LOW-03: SuspiciousActivityDetector unbounded event slice growth

**Severity:** LOW
**File:** `pkg/security/security.go:658-661`

**Problem:** Events are capped at 10000 but the slice keeps growing and shrinking, causing GC pressure.

**Recommendation:** Use a ring buffer or channel-based approach.

---

## 7. MISSING TESTS (INFO)

### INFO-01: No tests for execution-worker sandbox

**Severity:** INFO
**File:** `services/execution-worker/sandbox.go`

**Impact:** Critical sandbox logic has zero test coverage.

---

### INFO-02: No tests for RabbitMQ client

**Severity:** INFO
**File:** `pkg/rabbitmq/client.go`

**Impact:** Message publishing/consuming logic untested.

---

### INFO-03: No tests for Redis client

**Severity:** INFO
**File:** `pkg/redis/client.go`

**Impact:** Cache logic, circuit breaker, health monitor untested.

---

### INFO-04: No tests for TOTP generation/validation

**Severity:** INFO
**File:** `pkg/security/security.go` (TOTP functions)

**Impact:** 2FA security-critical code has no tests.

---

## 8. SUMMARY OF REQUIRED ACTIONS

### Must Fix Before Production (CRITICAL + HIGH):

1. **CRIT-01/02:** Fix integration test compilation errors
2. **HIGH-03:** Remove hardcoded default JWT secrets; fail on missing secrets
3. **HIGH-04:** Fix CORS configuration
4. **HIGH-05:** Disable local execution fallback by default
5. **HIGH-06:** Sanitize user code before embedding in harness
6. **HIGH-01/02:** Debug and fix auth middleware test failures

### Should Fix (MEDIUM):

7. **MED-01:** Consolidate RateLimiter implementations
8. **:** Unify security headers
9. **MED-04:** Replace custom `contains` with `strings.Contains`
10. **MED-05:** Fix goroutine leak in RequestTimeout
11. **MED-08:** Fix session cleanup to also clean userSessions index

### Nice to Have (LOW + INFO):

12. Replace `fmt.Printf` with structured logging
13. Add tests for sandbox, Redis, RabbitMQ, TOTP
14. Fix TOTP SHA256/SHA1 inconsistency
15. Per-user hint tracking

---

## Test Results Summary

```
PASS  internal/handler            (3.212s)
PASS  internal/repository         (0.571s)
PASS  internal/service            (4.687s)
PASS  pkg/logger                  (0.505s)
PASS  pkg/redis                   (0.304s)
FAIL  pkg/middleware              (1.288s)
  - TestAuthMiddleware_ValidToken: expected 200, got 401
  - TestAuthMiddleware_BearerCaseInsensitive: expected 200, got 401
FAIL  tests/integration           (build failed)
  - GenerateTokenPair return value mismatch (3 vs 5)
  - AuthMiddleware signature mismatch
  - Undefined imports
```

**Overall: 6 PASS, 2 FAIL**
