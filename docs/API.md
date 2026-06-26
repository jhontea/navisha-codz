# API Documentation — Coding Challenge Website

## Base URL

```
http://localhost:9100
```

---

## API Versioning

API ini mendukung versioning melalui **URL prefix** dan **response header**.

### Base URL per Version

| Version | Base URL Path | Status |
|---------|---------------|--------|
| `v1`    | `/v1/...`     | ✅ Aktif |
| Legacy  | `/api/...`    | ✅ Backward compatibility (akan di-*deprecate*) |

### Version Header

Setiap response menyertakan header:

| Header | Example | Keterangan |
|--------|---------|------------|
| `X-API-Version` | `v1` | Menunjukkan versi API yang melayani request |

### Contoh Request

```bash
# Via versioned prefix (recommended)
curl http://localhost:9100/v1/problems

# Via legacy prefix (backward compatible)
curl http://localhost:9100/api/problems
```

Kedua endpoint di atas mengembalikan response yang identik, termasuk header `X-API-Version: v1`.

> **Catatan**: Gunakan `/v1/` untuk semua integrasi baru. Prefix `/api/` tetap berfungsi untuk backward compatibility tetapi akan dihapus di rilis mendatang.

---

## Response Format

Semua response menggunakan JSON dengan `Content-Type: application/json`.

### Success Response

```json
{
  "data": { ... },
  "meta": {
    "request_id": "abc-123-def",
    "timestamp": "2026-06-26T10:30:00Z"
  }
}
```

| Field | Type | Keterangan |
|-------|------|------------|
| `data` | any | Response payload |
| `meta.request_id` | string | Unique request identifier for debugging |
| `meta.timestamp` | string | ISO 8601 timestamp |

### Error Response

```json
{
  "error": {
    "message": "Human-readable error message",
    "code": 400,
    "type": "validation_error",
    "details": [
      {
        "field": "code",
        "issue": "field is required"
      }
    ]
  },
  "meta": {
    "request_id": "abc-123-def",
    "timestamp": "2026-06-26T10:30:00Z"
  }
}
```

| Field | Type | Keterangan |
|-------|------|------------|
| `error.message` | string | Human-readable error description |
| `error.code` | int | HTTP status code |
| `error.type` | string | Error category: `validation_error`, `not_found`, `compilation_error`, `execution_error`, `rate_limit_error`, `internal_error` |
| `error.details` | list[ErrorDetail] | Optional detailed field-level errors |
| `error.details[].field` | string | Field name that caused the error |
| `error.details[].issue` | string | Description of the issue |

---

## HTTP Status Codes

| Code | Meaning | When |
|------|---------|------|
| `200` | OK | Request berhasil |
| `400` | Bad Request | Invalid input, malformed code, validation error |
| `404` | Not Found | Problem ID tidak ditemukan |
| `429` | Too Many Requests | Rate limit exceeded |
| `500` | Internal Server Error | Server-side error tak terduga |
| `502` | Bad Gateway | Docker sandbox error / execution failure |
| `503` | Service Unavailable | Sandbox unavailable (Docker not running) |

---

## Endpoints

### 1. GET /api/problems

Mendaftar semua problem (summary view).

#### Query Parameters

| Parameter | Type | Required | Keterangan |
|-----------|------|----------|------------|
| `difficulty` | string | ❌ | Filter by difficulty: `easy`, `medium`, `hard` |
| `category` | string | ❌ | Filter by category: `array`, `string`, `tree`, `dp`, `graph`, `stack`, etc. |
| `type` | string | ❌ | Filter by type: `function`, `main` |
| `tags` | string | ❌ | Filter by tags (AND logic). Comma-separated: `&tags=dp,array` akan mengembalikan problem yang memiliki SEMUA tag tsb |

#### Example Requests

```
GET /api/problems?difficulty=easy&category=array
```

```
GET /api/problems?tags=backtracking
```

```
GET /api/problems?tags=dp,string&difficulty=medium
```

#### Response (200 OK)

```json
{
  "data": [
    {
      "id": "two-sum",
      "title": "Two Sum",
      "type": "function",
      "difficulty": "easy",
      "category": "array",
      "tags": ["hash-map", "array"]
    },
    {
      "id": "reverse-string",
      "title": "Reverse String",
      "type": "function",
      "difficulty": "easy",
      "category": "string",
      "tags": ["two-pointers", "string"]
    }
  ],
  "meta": {
    "request_id": "req-001",
    "timestamp": "2026-06-26T10:30:00Z"
  }
}
```

#### Response Fields

| Field | Type | Keterangan |
|-------|------|------------|
| `data[].id` | string | Problem identifier |
| `data[].title` | string | Human-readable title |
| `data[].type` | string | `function` atau `main` |
| `data[].difficulty` | string | `easy`, `medium`, atau `hard` |
| `data[].category` | string | Problem category |
| `data[].tags` | list[string] | Filterable tags |

#### Error Responses

- `400` — Invalid difficulty or type value
  ```json
  {
    "error": {
      "message": "invalid difficulty: 'expert'. Must be one of: easy, medium, hard",
      "code": 400,
      "type": "validation_error"
    }
  }
  ```
- `500` — Server error

---

### 2. GET /api/problems/:id

Detail problem lengkap (tanpa solution).

#### Path Parameters

| Parameter | Type | Keterangan |
|-----------|------|------------|
| `id` | string | Problem ID (e.g., `two-sum`) |

#### Example Request

```
GET /api/problems/two-sum
```

#### Response (200 OK) — function-based

```json
{
  "data": {
    "id": "two-sum",
    "title": "Two Sum",
    "type": "function",
    "difficulty": "easy",
    "category": "array",
    "tags": ["hash-map", "array"],
    "description": "Given an array of integers `nums` and an integer `target`...",
    "function_name": "twoSum",
    "parameters": [
      {
        "name": "nums",
        "type": "[]int",
        "description": "Array of integers"
      },
      {
        "name": "target",
        "type": "int",
        "description": "Target sum"
      }
    ],
    "return_type": "[]int",
    "examples": [
      {
        "input": "nums = [2,7,11,15], target = 9",
        "output": "[0,1]",
        "explanation": "nums[0] + nums[1] = 2 + 7 = 9"
      }
    ],
    "hints": [
      {
        "level": 1,
        "title": "Use extra memory",
        "content": "Can you use a data structure to remember what you've seen?"
      }
    ],
    "template": "func twoSum(nums []int, target int) []int {\n    // Your code here\n    return nil\n}",
    "test_cases": [
      {
        "params": [[2, 7, 11, 15], 9],
        "expected": [0, 1],
        "description": "Basic case"
      }
    ],
    "constraints": [
      "2 ≤ len(nums) ≤ 10⁴",
      "-10⁹ ≤ nums[i] ≤ 10⁹"
    ],
    "time_complexity_hint": "O(n)",
    "space_complexity_hint": "O(n)"
  },
  "meta": {
    "request_id": "req-002",
    "timestamp": "2026-06-26T10:30:00Z"
  }
}
```

#### Response (200 OK) — main-based

```json
{
  "data": {
    "id": "hello-world",
    "title": "Hello World",
    "type": "main",
    "difficulty": "easy",
    "category": "basics",
    "tags": ["io", "basics"],
    "description": "Read a name from stdin and print a greeting.",
    "template": "package main\n\nimport \"fmt\"\n\nfunc main() {\n    // Your code here\n}\n",
    "test_cases": [
      {
        "input": "Alice",
        "expected": "Hello, Alice!",
        "description": "Basic greeting"
      }
    ],
    "constraints": ["Name is a single word, 1-50 characters"]
  },
  "meta": {
    "request_id": "req-003",
    "timestamp": "2026-06-26T10:30:00Z"
  }
}
```

> **Note**: Field `solution` TIDAK disertakan dalam response ini untuk mencegah kebocoran solusi.

#### Error Responses

- `404` — Problem not found
  ```json
  {
    "error": {
      "message": "problem 'two-sum' not found",
      "code": 404,
      "type": "not_found"
    }
  }
  ```

---

### 3. GET /api/problems/:id/template

Get only the code template for a problem (lightweight endpoint for editor).

#### Path Parameters

| Parameter | Type | Keterangan |
|-----------|------|------------|
| `id` | string | Problem ID |

#### Example Request

```
GET /api/problems/two-sum/template
```

#### Response (200 OK)

```json
{
  "data": {
    "id": "two-sum",
    "type": "function",
    "template": "func twoSum(nums []int, target int) []int {\n    // Your code here\n    return nil\n}",
    "function_name": "twoSum",
    "parameters": [
      {"name": "nums", "type": "[]int"},
      {"name": "target", "type": "int"}
    ],
    "return_type": "[]int"
  },
  "meta": {
    "request_id": "req-004",
    "timestamp": "2026-06-26T10:30:00Z"
  }
}
```

#### Response Fields

| Field | Type | Keterangan |
|-------|------|------------|
| `data.id` | string | Problem identifier |
| `data.type` | string | `function` atau `main` |
| `data.template` | string | Code template |
| `data.function_name` | string | Function name (function-based only) |
| `data.parameters` | list[Parameter] | Parameters (function-based only) |
| `data.return_type` | string | Return type (function-based only) |

#### Error Responses

- `404` — Problem not found

---

### 4. POST /api/problems/:id/run

Execute user code terhadap test cases di sandbox.

#### Path Parameters

| Parameter | Type | Keterangan |
|-----------|------|------------|
| `id` | string | Problem ID |

#### Request Body

```json
{
  "code": "func twoSum(nums []int, target int) []int {\n    seen := make(map[int]int)\n    for i, num := range nums {\n        if j, ok := seen[target-num]; ok {\n            return []int{j, i}\n        }\n        seen[num] = i\n    }\n    return nil\n}"
}
```

| Field | Type | Required | Keterangan |
|-------|------|----------|------------|
| `code` | string | ✅ | Go code (function body for function-based, full program for main-based) |

#### Response (200 OK) — All tests passed

```json
{
  "data": {
    "success": true,
    "compilation_error": null,
    "test_results": [
      {
        "name": "test_1",
        "passed": true,
        "expected": [0, 1],
        "actual": [0, 1],
        "error": null,
        "execution_time_ms": 12
      },
      {
        "name": "test_2",
        "passed": true,
        "expected": [1, 2],
        "actual": [1, 2],
        "error": null,
        "execution_time_ms": 8
      }
    ],
    "passed_count": 2,
    "total_count": 2,
    "execution_time_ms": 145
  },
  "meta": {
    "request_id": "req-005",
    "timestamp": "2026-06-26T10:30:00Z"
  }
}
```

#### Response (200 OK) — Some tests failed

```json
{
  "data": {
    "success": false,
    "compilation_error": null,
    "test_results": [
      {
        "name": "test_1",
        "passed": true,
        "expected": [0, 1],
        "actual": [0, 1],
        "error": null,
        "execution_time_ms": 10
      },
      {
        "name": "test_2",
        "passed": false,
        "expected": [1, 2],
        "actual": [2, 1],
        "error": "output mismatch: expected [1,2] got [2,1]",
        "execution_time_ms": 11
      }
    ],
    "passed_count": 1,
    "total_count": 2,
    "execution_time_ms": 132
  },
  "meta": {
    "request_id": "req-006",
    "timestamp": "2026-06-26T10:30:00Z"
  }
}
```

#### Response (200 OK) — Compilation error

```json
{
  "data": {
    "success": false,
    "compilation_error": "./main.go:12:5: undefined: x",
    "test_results": [],
    "passed_count": 0,
    "total_count": 0,
    "execution_time_ms": 0
  },
  "meta": {
    "request_id": "req-007",
    "timestamp": "2026-06-26T10:30:00Z"
  }
}
```

#### Response (200 OK) — Runtime error

```json
{
  "data": {
    "success": false,
    "compilation_error": null,
    "test_results": [
      {
        "name": "test_1",
        "passed": false,
        "expected": [0, 1],
        "actual": null,
        "error": "runtime error: index out of range [5] with length 5",
        "execution_time_ms": 5
      }
    ],
    "passed_count": 0,
    "total_count": 1,
    "execution_time_ms": 5
  },
  "meta": {
    "request_id": "req-008",
    "timestamp": "2026-06-26T10:30:00Z"
  }
}
```

#### Response Fields

| Field | Type | Keterangan |
|-------|------|------------|
| `data.success` | boolean | `true` jika semua test passed |
| `data.compilation_error` | string \| null | Error message jika gagal compile, `null` jika sukses |
| `data.test_results` | list[TestResult] | Hasil per test case |
| `data.test_results[].name` | string | Test case identifier |
| `data.test_results[].passed` | boolean | Pass atau fail |
| `data.test_results[].expected` | any | Expected value (any JSON type) |
| `data.test_results[].actual` | any | Actual value dari user code (any JSON type) |
| `data.test_results[].error` | string \| null | Error message (runtime error, wrong output, timeout) |
| `data.test_results[].execution_time_ms` | int | Execution time for this specific test |
| `data.passed_count` | int | Jumlah test yang passed |
| `data.total_count` | int | Total test cases |
| `data.execution_time_ms` | int | Total execution time dalam milliseconds |

#### Error Responses

- `400` — Missing or empty code field
  ```json
  {
    "error": {
      "message": "code field is required",
      "code": 400,
      "type": "validation_error"
    }
  }
  ```
- `404` — Problem not found
  ```json
  {
    "error": {
      "message": "problem 'two-sum' not found",
      "code": 404,
      "type": "not_found"
    }
  }
  ```
- `502` — Docker sandbox error
  ```json
  {
    "error": {
      "message": "Docker execution failed: container exited with code 137",
      "code": 502,
      "type": "execution_error"
    }
  }
  ```
- `503` — Service unavailable
  ```json
  {
    "error": {
      "message": "Sandbox unavailable: Docker is not running",
      "code": 503,
      "type": "execution_error"
    }
  }
  ```

---

### 5. GET /api/problems/:id/hints

Mengembalikan hints progresif untuk problem.

#### Path Parameters

| Parameter | Type | Keterangan |
|-----------|------|------------|
| `id` | string | Problem ID |

#### Query Parameters

| Parameter | Type | Required | Keterangan |
|-----------|------|----------|------------|
| `level` | int | ❌ | Maximum hint level to return (1-3). Jika diabaikan, semua hint dikembalikan |

#### Example Request

```
GET /api/problems/two-sum/hints?level=2
```

#### Response (200 OK)

```json
{
  "data": {
    "hints": [
      {
        "level": 1,
        "title": "Use extra memory",
        "content": "Can you use a data structure to remember what you've seen so far?"
      },
      {
        "level": 2,
        "title": "Hash map approach",
        "content": "For each element x, check if (target - x) exists in a map you've built."
      }
    ]
  },
  "meta": {
    "request_id": "req-009",
    "timestamp": "2026-06-26T10:30:00Z"
  }
}
```

#### Response Fields

| Field | Type | Keterangan |
|-------|------|------------|
| `data.hints[].level` | int | Hint level (1 = basic, 2 = intermediate, 3 = advanced) |
| `data.hints[].title` | string | Short hint title |
| `data.hints[].content` | string | Detailed hint |

#### Error Responses

- `404` — Problem not found

---

### 6. GET /health

Health check endpoint.

#### Example Request

```
GET /health
```

#### Response (200 OK)

```json
{
  "data": {
    "status": "ok",
    "version": "1.1.0"
  },
  "meta": {
    "request_id": "req-010",
    "timestamp": "2026-06-26T10:30:00Z"
  }
}
```

---

## Rate Limiting

Rate limiting diterapkan secara per-user tier berdasarkan role dari JWT claims (field `role`).

### Tiers

| Tier | Role di JWT | GET /api/* | POST /run & POST /validate |
|------|-------------|------------|----------------------------|
| **Free** | `user` (default) | 30 requests/min | 10 requests/min |
| **Premium** | `premium` | 300 requests/min | 100 requests/min |
| **Admin** | `admin` | Unlimited | Unlimited |

### Response Headers

Setiap response dari endpoint yang di-rate-limit menyertakan header berikut:

| Header | Contoh | Keterangan |
|--------|--------|------------|
| `X-RateLimit-Tier` | `free` / `premium` / `admin` | Tier user saat ini |
| `X-RateLimit-Limit` | `30` | Batas maksimum requests dalam window |
| `X-RateLimit-Remaining` | `25` | Sisa requests yang tersisa |
| `X-RateLimit-Reset` | `1719360000` | Unix timestamp saat window reset |

### Ketika Rate Limit Tercapai

API mengembalikan status `429 Too Many Requests`:

```json
{
  "error": {
    "message": "rate limit exceeded, try again later",
    "code": 429,
    "type": "rate_limit_error"
  }
}
```

---

## Problem Status

Daftar problem yang tersedia berdasarkan level:

### Easy (5 problems)

| ID | Title | Category | Tags |
|----|-------|----------|------|
| `two-sum` | Two Sum | array | hash-map, array |
| `reverse-string` | Reverse String | string | two-pointers, string |
| `fizz-buzz` | FizzBuzz | math | math, string |
| `contains-duplicate` | Contains Duplicate | array | array, hash-set |
| `max-subarray` | Maximum Subarray | array | array, divide-and-conquer, dp |

### Medium (5 problems)

| ID | Title | Category | Tags |
|----|-------|----------|------|
| `valid-parentheses` | Valid Parentheses | stack | stack, string |
| `merge-sorted-arrays` | Merge Sorted Array | array | array, two-pointers |
| `group-anagrams` | Group Anagrams | string | hash-map, sorting, string, array |
| `word-break` | Word Break | string | dp, string, hash-map |
| `permutations` | Permutations | array | backtracking, recursion, array |

### Hard (5 problems)

| ID | Title | Category | Tags |
|----|-------|----------|------|
| `longest-palindromic-substring` | Longest Palindromic Substring | string | string, dp, two-pointers |
| `coin-change` | Coin Change | array | dynamic-programming, array |
| `trapping-rain-water` | Trapping Rain Water | array | array, two-pointers, dp |
| `sudoku-solver` | Sudoku Solver | matrix | backtracking, matrix, recursion |
| `n-queens` | N-Queens | matrix | backtracking, recursion, matrix |
