# Test Plan — Coding Challenge Website

## Overview

This document describes the comprehensive test plan for the Coding Challenge Website, covering unit tests, integration tests, edge case tests, and frontend testing.

---

## 1. Unit Test Plan — Problem Loader (`internal/repository/problem.go`)

### 1.1 NewProblemRepository

| Test Case | Input | Expected Output |
|-----------|-------|-----------------|
| Valid directory with YAML files | `problems/` | Repository with 3 problems loaded |
| Non-existent directory | `/nonexistent` | Error: "cannot read problems directory" |
| Empty directory | `empty_dir/` | Repository with 0 problems |
| Directory with invalid YAML | `invalid_yaml/` | Error: "cannot parse YAML" |
| Directory with missing ID field | `missing_id/` | Error: "missing required field 'id'" |
| Directory with duplicate IDs | `duplicate_ids/` | Error: "duplicate problem ID" |

### 1.2 Load

| Test Case | Input | Expected Output |
|-----------|-------|-----------------|
| Load valid directory | `problems/` | 3 problems loaded, nil error |
| Reload after modification | Updated directory | Updated problems in memory |
| Concurrent load | Multiple goroutines | Thread-safe, no data race |

### 1.3 GetAll

| Test Case | Input | Expected Output |
|-----------|-------|-----------------|
| GetAll with no filters | `("", "")` | All problems sorted by difficulty then title |
| Filter by difficulty "easy" | `("easy", "")` | Only easy problems |
| Filter by difficulty "medium" | `("medium", "")` | Only medium problems |
| Filter by category "array" | `("", "array")` | Only array problems |
| Filter by non-existent difficulty | `("impossible", "")` | Empty slice |
| Case-insensitive filter | `("EASY", "")` | Easy problems (case insensitive) |

### 1.4 GetByID

| Test Case | Input | Expected Output |
|-----------|-------|-----------------|
| Existing ID | `"two-sum"` | Problem pointer |
| Non-existing ID | `"nonexistent"` | nil |
| Empty ID | `""` | nil |

### 1.5 Count

| Test Case | Input | Expected Output |
|-----------|-------|-----------------|
| After loading 3 problems | — | 3 |
| After loading empty dir | — | 0 |

---

## 2. Unit Test Plan — Hint Service (`internal/service/hint.go`)

### 2.1 GetHints (Progressive Reveal)

| Test Case | Input | Expected Output |
|-----------|-------|-----------------|
| First request (0 revealed) | Problem with 3 hints | First 2 hints |
| Second request (2 revealed) | Same problem | Third hint only |
| Third request (3 revealed) | Same problem | All hints (no more to reveal) |
| Nil problem | nil | nil |
| Problem with 1 hint | 1 hint | 1 hint on first request |
| Problem with 4 hints | 4 hints | 2, then 2 |

### 2.2 GetFullHints

| Test Case | Input | Expected Output |
|-----------|-------|-----------------|
| Valid problem | Problem with hints | All hints |
| Nil problem | nil | nil |

### 2.3 Reset

| Test Case | Input | Expected Output |
|-----------|-------|-----------------|
| Reset existing problem | problemID | Hints can be re-revealed from start |
| Reset non-existing problem | problemID | No error, no-op |

### 2.4 SanitizeProblemID

| Test Case | Input | Expected Output |
|-----------|-------|-----------------|
| Valid ID | `"two-sum"` | nil error |
| Valid ID with numbers | `"problem123"` | nil error |
| Valid ID with underscore | `"my_problem"` | nil error |
| Empty ID | `""` | Error: "problem ID is empty" |
| ID with spaces | `"two sum"` | Error: "invalid character" |
| ID with special chars | `"two$sum"` | Error: "invalid character" |
| ID with path traversal | `"../../etc/passwd"` | Error: "invalid character" |

### 2.5 ValidateCodeSize

| Test Case | Input | Expected Output |
|-----------|-------|-----------------|
| Small code | `"func main(){}"` | nil error |
| Exactly 64KB | 65536 bytes | nil error |
| Over 64KB | 65537 bytes | Error: "code exceeds maximum size" |
| Empty code | `""` | nil error |

### 2.6 BuildTestHarness

| Test Case | Input | Expected Output |
|-----------|-------|-----------------|
| Valid code + test cases | code, 2 test cases | Valid Go source string |
| Empty code | "", 1 test cases | Go source with empty function |
| No test cases | code, [] | Go source with empty test slice |

---

## 3. Unit Test Plan — Code Runner (`internal/service/runner.go`)

### 3.1 RunCode

| Test Case | Input | Expected Output |
|-----------|-------|-----------------|
| Valid code, all tests pass | Correct solution | Success=true, all passed |
| Valid code, some tests fail | Partial solution | Success=false, partial pass |
| Compilation error | Invalid Go code | CompilationError set, Success=false |
| Timeout code | Infinite loop | Timeout error, all failed |
| Empty code | `""` | Compilation error |

### 3.2 parseOutput

| Test Case | Input | Expected Output |
|-----------|-------|-----------------|
| All lines match expected | "line1\nline2", [exp1, exp2] | All passed |
| Partial match | "line1\nwrong", [exp1, exp2] | 1 passed, 1 failed |
| More lines than tests | "a\nb\nc", [a, b] | 2 tested, extra ignored |
| Fewer lines than tests | "a", [a, b, c] | 1 passed, 2 failed (empty actual) |
| Empty output | "", [exp1] | 1 failed with empty actual |

### 3.3 allFailed

| Test Case | Input | Expected Output |
|-----------|-------|-----------------|
| Multiple test cases | 3 test cases, "timeout" | 3 failed results with "timeout" error |
| Single test case | 1 test case, "error" | 1 failed result |
| No test cases | [] | Empty slice |

---

## 4. Integration Test Plan — API Endpoints (`internal/handler/problem.go`)

### 4.1 GET /health

| Test Case | Expected Output |
|-----------|-----------------|
| GET /health | 200, `{"status":"ok"}` |

### 4.2 GET /api/problems

| Test Case | Expected Output |
|-----------|-----------------|
| No query params | 200, array of all problem summaries |
| ?difficulty=easy | 200, only easy problems |
| ?difficulty=invalid | 400, error message |
| ?category=array | 200, only array problems |
| ?difficulty=easy&category=array | 200, filtered by both |

### 4.3 GET /api/problems/:id

| Test Case | Expected Output |
|-----------|-----------------|
| Valid ID "two-sum" | 200, problem detail without solution |
| Invalid ID "nonexistent" | 404, error message |
| ID with special chars "../../etc" | 400, validation error |

### 4.4 POST /api/problems/:id/run

| Test Case | Expected Output |
|-----------|-----------------|
| Valid code that passes | 200, success=true |
| Valid code that fails some tests | 200, success=false |
| Invalid Go code | 200, compilation_error set |
| Missing code field | 400, error |
| Empty code | 400, error |
| Code exceeding 64KB | 400, error |
| Non-existent problem ID | 404, error |
| Invalid problem ID format | 400, error |

### 4.5 GET /api/problems/:id/hints

| Test Case | Expected Output |
|-----------|-----------------|
| Valid ID | 200, hints array |
| Non-existent ID | 404, error |
| Invalid ID format | 400, error |

---

## 5. Edge Case Tests

| Category | Test Case | Expected Behavior |
|----------|-----------|-------------------|
| **Empty input** | Empty code submission | 400 error |
| **Empty input** | Empty problem directory | 0 problems loaded |
| **Invalid YAML** | Malformed YAML syntax | Parse error with context |
| **Missing fields** | YAML without `id` | Error: missing required field |
| **Missing fields** | YAML without `title` | Loaded with empty title |
| **Duplicate IDs** | Two files with same ID | Error on load |
| **Large code** | 100KB code submission | 400 error (size limit) |
| **Special chars in ID** | ID with `/`, `.`, `..` | 400 validation error |
| **Concurrent access** | Multiple simultaneous reads | Thread-safe, no race conditions |
| **Unicode in code** | Code with Unicode characters | Handled correctly |
| **Very long output** | Code producing MBs of output | Truncated/limited |
| **Nil problem** | GetHints(nil) | Returns nil, no panic |
| **Race condition** | Concurrent hint reveals | Thread-safe counter |

---

## 6. Frontend Test Plan

### 6.1 Problem List Page

| Test Case | Expected Behavior |
|-----------|-------------------|
| Page loads | Problem grid populated with cards |
| Filter by difficulty | Only matching problems shown |
| Filter by category | Only matching problems shown |
| Search by title | Filtered results update in real-time |
| No results | Empty state message displayed |
| Click problem card | Navigates to problem detail |

### 6.2 Problem Detail Page

| Test Case | Expected Behavior |
|-----------|-------------------|
| Page loads | Problem description, examples, constraints rendered |
| Code editor initialized | CodeMirror with Go mode, template loaded |
| Submit code | Loading state, then results displayed |
| Reset code | Editor reset to template |
| Reveal hint | Confirmation dialog, then hint shown |
| All hints revealed | "All hints revealed" message |

### 6.3 Code Editor

| Test Case | Expected Behavior |
|-----------|-------------------|
| CodeMirror loads | Editor visible with line numbers |
| Go syntax highlighting | Keywords highlighted |
| Tab key | Inserts 4 spaces |
| Auto-indent on Enter | Maintains indentation level |
| Ctrl+Enter | Submits code |
| Ctrl+/ | Toggles comment |

### 6.4 Results Panel

| Test Case | Expected Behavior |
|-----------|-------------------|
| All tests passed | Green status, progress bar 100% |
| Some tests failed | Red status, progress bar partial |
| Compilation error | Error message displayed |
| Test case expand/collapse | Toggle visibility of details |

### 6.5 Accessibility

| Test Case | Expected Behavior |
|-----------|-------------------|
| Keyboard navigation | All interactive elements reachable via keyboard |
| ARIA labels | Sidebar toggle has aria-label |
| Color contrast | Text meets WCAG AA contrast ratios |
| Screen reader | Semantic HTML structure |
| Focus indicators | Visible focus on interactive elements |

---

## 7. Test Execution Strategy

1. **Unit tests**: Run with `go test ./... -v -race`
2. **Integration tests**: Start server, use `curl` to test endpoints
3. **Frontend tests**: Manual verification in browser (no automated test framework currently)
4. **Edge case tests**: Covered in unit tests with specific edge case inputs

## 8. Coverage Goals

- Problem repository: >90% coverage
- Hint service: >90% coverage
- Runner service: >80% coverage (Docker-dependent paths may be skipped)
- HTTP handlers: >85% coverage
