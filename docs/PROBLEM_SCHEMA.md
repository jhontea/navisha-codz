# Problem Schema — Coding Challenge Website

## Overview

Semua problem disimpan dalam format YAML di direktori `problems/`. Setiap file YAML mendefinisikan satu problem lengkap dengan metadata, deskripsi, contoh, test cases, hints, template code, dan solusi.

Mendukung **dua tipe problem**:
- **`function`**: User menulis function body, test harness memanggil function tersebut dengan parameter
- **`main`**: User menulis complete main package, test via stdin/stdout

---

## Directory Structure

```
problems/
├── easy/
│   ├── two-sum.yaml
│   ├── fizz-buzz.yaml
│   ├── reverse-string.yaml
│   └── ...
├── medium/
│   ├── valid-parentheses.yaml
│   ├── merge-sorted-arrays.yaml
│   └── ...
└── hard/
    ├── merge-k-sorted-lists.yaml
    └── ...
```

### Naming Rules

| Aturan | Penjelasan | Contoh |
|--------|-----------|--------|
| Direktori | Berdasarkan difficulty (`easy/`, `medium/`, `hard/`) | `easy/`, `medium/` |
| Nama file | lowercase-with-dashes, sesuai judul problem | `two-sum.yaml`, `valid-parentheses.yaml` |
| Ekstensi | `.yaml` (bukan `.yml`) | `two-sum.yaml` |
| ID | Sama dengan nama file tanpa ekstensi | `two-sum` |

---

## YAML Schema Specification

### Top-Level Fields

| Field | Type | Required | Keterangan |
|-------|------|----------|------------|
| `id` | string | ✅ | Unique identifier, sama dengan nama file tanpa ekstensi |
| `title` | string | ✅ | Judul problem (human-readable) |
| `type` | enum: `function`, `main` | ✅ | Tipe problem — lihat [Problem Types](#problem-types) |
| `difficulty` | enum: `easy`, `medium`, `hard` | ✅ | Tingkat kesulitan |
| `category` | string | ✅ | Kategori utama (contoh: `array`, `string`, `tree`, `dp`, `graph`) |
| `tags` | list[string] | ✅ | Tag tambahan untuk filtering |
| `description` | string | ✅ | Deskripsi problem (mendukung Markdown) |
| `examples` | list[Example] | ✅ | Minimal 1 contoh |
| `hints` | list[Hint] | ✅ | Minimal 1 hint |
| `template` | string | ✅ | Go code template yang diisi user |
| `test_cases` | list[TestCase] | ✅ | Minimal 1 test case |
| `function_name` | string | ⚠️ | Required if `type: function`. Nama function yang harus di-implement |
| `parameters` | list[Parameter] | ⚠️ | Required if `type: function`. Daftar parameter function |
| `return_type` | string | ⚠️ | Required if `type: function`. Tipe return function |
| `constraints` | list[string] | ❌ | Batasan input |
| `time_complexity_hint` | string | ❌ | Hint time complexity yang diharapkan |
| `space_complexity_hint` | string | ❌ | Hint space complexity yang diharapkan |
| `solution` | Solution | ❌ | Solusi referensi (HIDDEN dari API response) |

---

## Problem Types

### Type: `function`

User menulis **function body** saja. Test harness akan:
1. Wrap user code dalam `package main`
2. Generate `main()` yang memanggil function dengan test parameters
3. Compare return value dengan expected

**Required additional fields:**
- `function_name`: Nama function (e.g., `twoSum`)
- `parameters`: Daftar parameter dengan nama dan tipe
- `return_type`: Tipe data return

**Template format:**
```go
func <function_name>(<params>) <return_type> {
    // Your code here
}
```

### Type: `main`

User menulis **complete program** dengan `package main` dan `func main()`. Test harness akan:
1. Compile user code apa adanya
2. Feed input via stdin
3. Capture dan compare stdout dengan expected

**Template format:**
```go
package main

import "fmt"

func main() {
    // Your code here
}
```

---

### Parameter Object (function-based only)

```yaml
parameters:
  - name: "nums"
    type: "[]int"
    description: "Array of integers"
  - name: "target"
    type: "int"
    description: "Target sum"
```

| Field | Type | Required | Keterangan |
|-------|------|----------|------------|
| `name` | string | ✅ | Nama parameter (valid Go identifier) |
| `type` | string | ✅ | Tipe data Go (e.g., `[]int`, `string`, `*ListNode`) |
| `description` | string | ❌ | Penjelasan parameter |

---

### Example Object

```yaml
examples:
  - input: "nums = [2,7,11,15], target = 9"
    output: "[0,1]"
    explanation: "Because nums[0] + nums[1] == 9, we return [0, 1]."
```

| Field | Type | Keterangan |
|-------|------|------------|
| `input` | string | Format parameter input (bisa multi-line) |
| `output` | string | Expected output |
| `explanation` | string | Penjelasan mengapa output tersebut (opsional per example) |

---

### Hint Object

```yaml
hints:
  - level: 1
    title: "Think about lookup"
    content: "What data structure gives O(1) lookup? Consider using a map."
```

| Field | Type | Required | Keterangan |
|-------|------|----------|------------|
| `level` | int | ✅ | 1 (vague) → 2 (directional) → 3 (almost giving away) |
| `title` | string | ✅ | Judul singkat hint |
| `content` | string | ✅ | Isi hint (mendukung Markdown) |

---

### TestCase Object

```yaml
test_cases:
  - params: [[2, 7, 11, 15], 9]
    expected: [0, 1]
    description: "Basic case: first two elements sum to target"
```

| Field | Type | Required | Keterangan |
|-------|------|----------|------------|
| `params` | list[any] | ✅ | Array of parameter values (urutan sesuai `parameters`). Bisa任何 JSON type: array, string, int, bool, null |
| `expected` | any | ✅ | Expected return value. Bisa任何 JSON type — tidak hanya string |
| `description` | string | ❌ | Deskripsi test case untuk debugging |

> **Note:** Field `input` dan `expected` yang lama (string-based) tetap didukung untuk backward compatibility dengan `type: main`. Untuk `type: function`, gunakan `params` dan `expected`.

#### Test Case Format Specification

**For `type: function`:**
- `params` adalah array JSON yang berisi nilai setiap parameter
- `expected` adalah nilai return yang diharapkan (any JSON type)
- Tipe data yang didukung: `int`, `float`, `string`, `bool`, `null`, `[]int`, `[]string`, `[][]int`, nested arrays

**For `type: main`:**
- `input` adalah raw string yang akan di-feed ke stdin
- `expected` adalah raw string yang diharapkan di stdout
- Multiple values dipisahkan newline

---

### Solution Object (Hidden)

```yaml
solution:
  code: |
    func twoSum(nums []int, target int) []int {
        seen := make(map[int]int)
        for i, num := range nums {
            if j, ok := seen[target-num]; ok {
                return []int{j, i}
            }
            seen[num] = i
        }
        return nil
    }
  approach: "Hash map for O(n) single-pass"
  time_complexity: "O(n)"
  space_complexity: "O(n)"
```

| Field | Type | Keterangan |
|-------|------|------------|
| `code` | string | Solusi lengkap Go (multi-line) |
| `approach` | string | Penjelasan pendekatan |
| `time_complexity` | string | Big-O time |
| `space_complexity` | string | Big-O space |

---

## Complete Example 1: Easy Problem — Two Sum (function-based)

```yaml
# File: problems/easy/two-sum.yaml

id: "two-sum"
title: "Two Sum"
type: "function"
difficulty: "easy"
category: "array"
tags: ["hash-map", "array"]

description: |
  Given an array of integers `nums` and an integer `target`, return indices
  of the two numbers such that they add up to `target`.

  You may assume that each input would have exactly one solution, and you
  may not use the same element twice.

function_name: "twoSum"
parameters:
  - name: "nums"
    type: "[]int"
    description: "Array of integers"
  - name: "target"
    type: "int"
    description: "Target sum"
return_type: "[]int"

examples:
  - input: "nums = [2,7,11,15], target = 9"
    output: "[0,1]"
    explanation: "nums[0] + nums[1] = 2 + 7 = 9, so we return [0, 1]."
  - input: "nums = [3,2,4], target = 6"
    output: "[1,2]"
    explanation: "nums[1] + nums[2] = 2 + 4 = 6, so we return [1, 2]."

hints:
  - level: 1
    title: "Use extra memory"
    content: "Can you use a data structure to remember what you've seen so far?"
  - level: 2
    title: "Hash map approach"
    content: "For each element x, check if (target - x) exists in a map you've built."
  - level: 3
    title: "One-pass hash map"
    content: |
      Iterate through nums once. For each num:
      1. Compute complement = target - num
      2. If complement in current map, return [map[complement], i]
      3. Otherwise add num → i to map

template: |
  func twoSum(nums []int, target int) []int {
      // Your code here
      return nil
  }

test_cases:
  - params: [[2, 7, 11, 15], 9]
    expected: [0, 1]
    description: "Basic case with positive numbers"
  - params: [[3, 2, 4], 6]
    expected: [1, 2]
    description: "Non-adjacent elements"
  - params: [[3, 3], 6]
    expected: [0, 1]
    description: "Duplicate values"
  - params: [[-1, -2, -3, -4, -5], -8]
    expected: [2, 4]
    description: "Negative numbers"
  - params: [[0, 4, 3, 0], 0]
    expected: [0, 3]
    description: "Zeros in input"

constraints:
  - "2 ≤ len(nums) ≤ 10⁴"
  - "-10⁹ ≤ nums[i] ≤ 10⁹"
  - "-10⁹ ≤ target ≤ 10⁹"
  - "Only one valid answer exists."

time_complexity_hint: "O(n)"
space_complexity_hint: "O(n)"

solution:
  code: |
    func twoSum(nums []int, target int) []int {
        seen := make(map[int]int)
        for i, num := range nums {
            if j, ok := seen[target-num]; ok {
                return []int{j, i}
            }
            seen[num] = i
        }
        return nil
    }
  approach: "Single-pass hash map"
  time_complexity: "O(n)"
  space_complexity: "O(n)"
```

---

## Complete Example 2: Easy Problem — FizzBuzz (function-based)

```yaml
# File: problems/easy/fizz-buzz.yaml

id: "fizz-buzz"
title: "FizzBuzz"
type: "function"
difficulty: "easy"
category: "math"
tags: ["math", "string", "conditional"]

description: |
  Given an integer `n`, return a string array `answer` (1-indexed) where:

  - `answer[i] == "FizzBuzz"` if `i` is divisible by both 3 and 5
  - `answer[i] == "Fizz"` if `i` is divisible by 3
  - `answer[i] == "Buzz"` if `i` is divisible by 5
  - `answer[i] == strconv.Itoa(i)` if none of the above conditions are true

function_name: "fizzBuzz"
parameters:
  - name: "n"
    type: "int"
    description: "Upper bound of range (1 to n)"
return_type: "[]string"

examples:
  - input: "n = 3"
    output: '["1","2","Fizz"]'
  - input: "n = 5"
    output: '["1","2","Fizz","4","Buzz"]'
  - input: "n = 15"
    output: '["1","2","Fizz","4","Buzz","Fizz","7","8","Fizz","Buzz","11","Fizz","13","14","FizzBuzz"]'

hints:
  - level: 1
    title: "Modulo operator"
    content: "Use the `%` operator to check divisibility."
  - level: 2
    title: "Order matters"
    content: "Check divisibility by both 3 and 5 FIRST, before checking individually."
  - level: 3
    title: "String conversion"
    content: "Use `strconv.Itoa(i)` to convert integer to string."

template: |
  func fizzBuzz(n int) []string {
      // Your code here
      return nil
  }

test_cases:
  - params: [3]
    expected: ["1", "2", "Fizz"]
    description: "Simple case: n=3"
  - params: [5]
    expected: ["1", "2", "Fizz", "4", "Buzz"]
    description: "n=5 with Fizz and Buzz"
  - params: [15]
    expected: ["1", "2", "Fizz", "4", "Buzz", "Fizz", "7", "8", "Fizz", "Buzz", "11", "Fizz", "13", "14", "FizzBuzz"]
    description: "Full FizzBuzz case"
  - params: [1]
    expected: ["1"]
    description: "Minimum n"
  - params: [100]
    expected: ["1", "2", "Fizz", "4", "Buzz", "Fizz", "7", "8", "Fizz", "Buzz", "11", "Fizz", "13", "14", "FizzBuzz", "16", "17", "Fizz", "19", "Buzz", "Fizz", "22", "23", "Fizz", "Buzz", "26", "Fizz", "28", "29", "FizzBuzz", "31", "32", "Fizz", "34", "Buzz", "Fizz", "37", "38", "Fizz", "Buzz", "41", "Fizz", "43", "44", "FizzBuzz", "46", "47", "Fizz", "49", "Buzz", "Fizz", "52", "53", "Fizz", "Buzz", "56", "Fizz", "58", "59", "FizzBuzz", "61", "62", "Fizz", "64", "Buzz", "Fizz", "67", "68", "Fizz", "Buzz", "71", "Fizz", "73", "74", "FizzBuzz", "76", "77", "Fizz", "79", "Buzz", "Fizz", "82", "83", "Fizz", "Buzz", "86", "Fizz", "88", "89", "FizzBuzz", "91", "92", "Fizz", "94", "Buzz", "Fizz", "97", "98", "Fizz", "Buzz"]
    description: "Large n=100"

constraints:
  - "1 ≤ n ≤ 10⁴"

time_complexity_hint: "O(n)"
space_complexity_hint: "O(n)"

solution:
  code: |
    func fizzBuzz(n int) []string {
        result := make([]string, n)
        for i := 1; i <= n; i++ {
            switch {
            case i%15 == 0:
                result[i-1] = "FizzBuzz"
            case i%3 == 0:
                result[i-1] = "Fizz"
            case i%5 == 0:
                result[i-1] = "Buzz"
            default:
                result[i-1] = strconv.Itoa(i)
            }
        }
        return result
    }
  approach: "Simple iteration with modulo checks"
  time_complexity: "O(n)"
  space_complexity: "O(n)"
```

---

## Complete Example 3: Easy Problem — Reverse String (function-based)

```yaml
# File: problems/easy/reverse-string.yaml

id: "reverse-string"
title: "Reverse String"
type: "function"
difficulty: "easy"
category: "string"
tags: ["two-pointers", "string", "array"]

description: |
  Write a function that reverses a string. The input string is given as an
  array of characters `s`. You must do this by modifying the input array
  in-place with O(1) extra memory.

function_name: "reverseString"
parameters:
  - name: "s"
    type: "[]byte"
    description: "Array of characters to reverse in-place"
return_type: "void"

examples:
  - input: 's = ["h","e","l","l","o"]'
    output: '["o","l","l","e","h"]'
  - input: 's = ["H","a","n","n","a","h"]'
    output: '["h","a","n","n","a","H"]'

hints:
  - level: 1
    title: "Two pointers"
    content: "Think about swapping elements from both ends moving inward."
  - level: 2
    title: "Swap in place"
    content: "Use two pointers: one at start, one at end. Swap and move inward."
  - level: 3
    title: "Loop condition"
    content: "Loop while left < right. Swap s[left] and s[right], then left++, right--."

template: |
  func reverseString(s []byte) {
      // Your code here
  }

test_cases:
  - params: [[["h", "e", "l", "l", "o"]]]
    expected: [["o", "l", "l", "e", "h"]]
    description: "Odd length string"
  - params: [[["H", "a", "n", "n", "a", "h"]]]
    expected: [["h", "a", "n", "n", "a", "H"]]
    description: "Even length string"
  - params: [[["a"]]]
    expected: [["a"]]
    description: "Single character"
  - params: [[[]]]
    expected: [[]]
    description: "Empty array"

constraints:
  - "1 ≤ len(s) ≤ 10⁵"
  - "s[i] is a printable ASCII character."

time_complexity_hint: "O(n)"
space_complexity_hint: "O(1)"

solution:
  code: |
    func reverseString(s []byte) {
        left, right := 0, len(s)-1
        for left < right {
            s[left], s[right] = s[right], s[left]
            left++
            right--
        }
    }
  approach: "Two-pointer swap"
  time_complexity: "O(n)"
  space_complexity: "O(1)"
```

---

## Complete Example 4: Medium Problem — Valid Parentheses (function-based)

```yaml
# File: problems/medium/valid-parentheses.yaml

id: "valid-parentheses"
title: "Valid Parentheses"
type: "function"
difficulty: "medium"
category: "stack"
tags: ["stack", "string", "matching"]

description: |
  Given a string `s` containing just the characters `(`, `)`, `{`, `}`, `[` and `]`,
  determine if the input string is valid.

  An input string is valid if:
  1. Open brackets must be closed by the same type of brackets.
  2. Open brackets must be closed in the correct order.
  3. Every close bracket has a corresponding open bracket of the same type.

function_name: "isValid"
parameters:
  - name: "s"
    type: "string"
    description: "String containing only parentheses characters"
return_type: "bool"

examples:
  - input: "s = '()'"
    output: "true"
  - input: "s = '()[]{}'"
    output: "true"
  - input: "s = '(]'"
    output: "false"
    explanation: "'(' is closed by ']' which is a different type."

hints:
  - level: 1
    title: "Stack data structure"
    content: "Think about LIFO (Last In, First Out) — what data structure matches this?"
  - level: 2
    title: "Push and pop"
    content: "Push opening brackets onto a stack. When you see a closing bracket, pop and check if it matches."
  - level: 3
    title: "Map for matching pairs"
    content: |
      Use a map: `{')': '(', '}': '{', ']': '['}`
      For each char:
      - If it's an opening bracket → push to stack
      - If it's a closing bracket → check if stack top matches the expected opening bracket

template: |
  func isValid(s string) bool {
      // Your code here
      return false
  }

test_cases:
  - params: ["()"]
    expected: true
    description: "Simple valid pair"
  - params: ["()[]{}"]
    expected: true
    description: "Multiple valid pairs"
  - params: ["(]"]
    expected: false
    description: "Mismatched pair"
  - params: ["([)]"]
    expected: false
    description: "Wrong nesting order"
  - params: ["{[]}"]
    expected: true
    description: "Nested valid brackets"
  - params: ["("]
    expected: false
    description: "Unclosed opening bracket"
  - params: [")"]
    expected: false
    description: "Closing bracket without opening"
  - params: [""]
    expected: true
    description: "Empty string is valid"

constraints:
  - "1 ≤ len(s) ≤ 10⁴"
  - "s consists of parentheses only '()[]{}'."

time_complexity_hint: "O(n)"
space_complexity_hint: "O(n)"

solution:
  code: |
    func isValid(s string) bool {
        stack := []rune{}
        pairs := map[rune]rune{')': '(', '}': '{', ']': '['}

        for _, ch := range s {
            if opening, ok := pairs[ch]; ok {
                if len(stack) == 0 || stack[len(stack)-1] != opening {
                    return false
                }
                stack = stack[:len(stack)-1]
            } else {
                stack = append(stack, ch)
            }
        }
        return len(stack) == 0
    }
  approach: "Stack with matching map"
  time_complexity: "O(n)"
  space_complexity: "O(n)"
```

---

## Complete Example 5: Medium Problem — Merge Sorted Arrays (function-based)

```yaml
# File: problems/medium/merge-sorted-arrays.yaml

id: "merge-sorted-arrays"
title: "Merge Sorted Array"
type: "function"
difficulty: "medium"
category: "array"
tags: ["array", "two-pointers", "sorting"]

description: |
  You are given two integer arrays `nums1` and `nums2`, sorted in non-decreasing order,
  and two integers `m` and `n`, representing the number of elements in `nums1` and `nums2`
  respectively.

  Merge `nums1` and `nums2` into a single array sorted in non-decreasing order.

  The final sorted array should not be returned by the function, but instead be stored
  inside the array `nums1`. To accommodate this, `nums1` has a length of `m + n`, where
  the first `m` elements denote the elements that should be merged, and the last `n` elements
  are set to 0 and should be ignored. `nums2` has a length of `n`.

function_name: "merge"
parameters:
  - name: "nums1"
    type: "[]int"
    description: "First sorted array with length m+n, last n elements are 0"
  - name: "m"
    type: "int"
    description: "Number of valid elements in nums1"
  - name: "nums2"
    type: "[]int"
    description: "Second sorted array"
  - name: "n"
    type: "int"
    description: "Number of elements in nums2"
return_type: "void"

examples:
  - input: "nums1 = [1,2,3,0,0,0], m = 3, nums2 = [2,5,6], n = 3"
    output: "[1,2,2,3,5,6]"
    explanation: "Merging [1,2,3] and [2,5,6] gives [1,2,2,3,5,6]."
  - input: "nums1 = [1], m = 1, nums2 = [], n = 0"
    output: "[1]"
  - input: "nums1 = [0], m = 0, nums2 = [1], n = 1"
    output: "[1]"

hints:
  - level: 1
    title: "Work backwards"
    content: "Instead of merging from the front, try merging from the back to avoid overwriting."
  - level: 2
    title: "Three pointers"
    content: "Use pointers at the end of valid elements in nums1, end of nums2, and end of nums1 array."
  - level: 3
    title: "Fill from the end"
    content: |
      Start from index m+n-1. Compare nums1[m-1] and nums2[n-1], place larger at end.
      Decrement pointers accordingly. If nums2 still has elements, copy them over.

template: |
  func merge(nums1 []int, m int, nums2 []int, n int) {
      // Your code here
  }

test_cases:
  - params: [[1, 2, 3, 0, 0, 0], 3, [2, 5, 6], 3]
    expected: [1, 2, 2, 3, 5, 6]
    description: "Standard merge"
  - params: [[1], 1, [], 0]
    expected: [1]
    description: "nums2 is empty"
  - params: [[0], 0, [1], 1]
    expected: [1]
    description: "nums1 has no valid elements"
  - params: [[2, 0], 1, [1], 1]
    expected: [1, 2]
    description: "Single element each, nums2 smaller"
  - params: [[-1, 0, 0, 3, 0, 0, 0], 3, [-2, 5, 6], 3]
    expected: [-2, -1, 0, 0, 3, 5, 6]
    description: "Mixed positive and negative"

constraints:
  - "len(nums1) == m + n"
  - "len(nums2) == n"
  - "0 ≤ m, n ≤ 200"
  - "1 ≤ m + n ≤ 200"
  - "-10⁹ ≤ nums1[i], nums2[j] ≤ 10⁹"

time_complexity_hint: "O(m + n)"
space_complexity_hint: "O(1)"

solution:
  code: |
    func merge(nums1 []int, m int, nums2 []int, n int) {
        i, j, k := m-1, n-1, m+n-1
        for j >= 0 {
            if i >= 0 && nums1[i] > nums2[j] {
                nums1[k] = nums1[i]
                i--
            } else {
                nums1[k] = nums2[j]
                j--
            }
            k--
        }
    }
  approach: "Three-pointer reverse merge"
  time_complexity: "O(m + n)"
  space_complexity: "O(1)"
```

---

## Validation Rules

1. **ID uniqueness**: Setiap `id` harus unik di seluruh problem bank.
2. **Difficulty-directory match**: `difficulty: easy` harus berada di `easy/`, dst.
3. **Type-specific fields**: Jika `type: function`, maka `function_name`, `parameters`, `return_type` wajib ada.
4. **Minimal examples**: Minimal 1 example per problem.
5. **Minimal test cases**: Minimal 1 test case per problem.
6. **Minimal hints**: Minimal 1 hint per problem.
7. **Template compilable**: `template` harus bisa di-compile tanpa error (setelah user mengisi).
8. **Test cases solvable**: Semua `test_cases` harus solvable oleh `solution.code`.
9. **Solution hidden**: Field `solution` TIDAK boleh di-include dalam API response `GET /api/problems/:id`.
10. **Params match parameters**: Jumlah dan tipe `params` di test case harus cocok dengan `parameters` definition.
11. **Return type match**: Tipe `expected` di test case harus cocok dengan `return_type`.

---

## Adding New Problems

1. Buat file YAML di direktori sesuai difficulty
2. Ikuti schema di atas
3. Pastikan `id` = nama file tanpa `.yaml`
4. Tentukan `type` (`function` atau `main`)
5. Jika `function`, definisikan `function_name`, `parameters`, `return_type`
6. Jalankan validator: `go run cmd/validator/main.go`
7. Test: `go run cmd/server/main.go` lalu akses `GET /api/problems/{id}`

---

## Test Harness Generation

### For `type: function`

Test harness auto-generated dari YAML:

```go
package main

import (
    "encoding/json"
    "fmt"
    "reflect"
)

// === USER CODE (inserted here) ===
func twoSum(nums []int, target int) []int {
    // User's implementation
    return nil
}
// === END USER CODE ===

func main() {
    tests := []struct {
        params   []interface{}
        expected interface{}
    }{
        {params: []interface{}{[]int{2, 7, 11, 15}, 9}, expected: []int{0, 1}},
        // ... more tests
    }

    for i, test := range tests {
        result := twoSum(test.params[0].([]int), test.params[1].(int))
        if reflect.DeepEqual(result, test.expected) {
            fmt.Printf("test_%d: PASS\n", i+1)
        } else {
            expectedJSON, _ := json.Marshal(test.expected)
            resultJSON, _ := json.Marshal(result)
            fmt.Printf("test_%d: FAIL expected=%s got=%s\n", i+1, expectedJSON, resultJSON)
        }
    }
}
```

### For `type: main`

Test harness menggunakan stdin/stdout:

```bash
echo "$INPUT" | go run user_code.go | diff - "$EXPECTED"
```
