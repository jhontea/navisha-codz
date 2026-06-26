1|# API Documentation — Coding Challenge Website
2|
3|## Base URL
4|
5|```
6|http://localhost:9100
7|```
8|
9|---
10|
11|## Response Format
12|
13|Semua response menggunakan JSON dengan `Content-Type: application/json`.
14|
15|### Success Response
16|
17|```json
18|{
19|  "data": { ... },
20|  "meta": {
21|    "request_id": "abc-123-def",
22|    "timestamp": "2026-06-26T10:30:00Z"
23|  }
24|}
25|```
26|
27|| Field | Type | Keterangan |
28||-------|------|------------|
29|| `data` | any | Response payload |
30|| `meta.request_id` | string | Unique request identifier for debugging |
31|| `meta.timestamp` | string | ISO 8601 timestamp |
32|
33|### Error Response
34|
35|```json
36|{
37|  "error": {
38|    "message": "Human-readable error message",
39|    "code": 400,
40|    "type": "validation_error",
41|    "details": [
42|      {
43|        "field": "code",
44|        "issue": "field is required"
45|      }
46|    ]
47|  },
48|  "meta": {
49|    "request_id": "abc-123-def",
50|    "timestamp": "2026-06-26T10:30:00Z"
51|  }
52|}
53|```
54|
55|| Field | Type | Keterangan |
56||-------|------|------------|
57|| `error.message` | string | Human-readable error description |
58|| `error.code` | int | HTTP status code |
59|| `error.type` | string | Error category: `validation_error`, `not_found`, `compilation_error`, `execution_error`, `rate_limit_error`, `internal_error` |
60|| `error.details` | list[ErrorDetail] | Optional detailed field-level errors |
61|| `error.details[].field` | string | Field name that caused the error |
62|| `error.details[].issue` | string | Description of the issue |
63|
64|---
65|
66|## HTTP Status Codes
67|
68|| Code | Meaning | When |
69||------|---------|------|
70|| `200` | OK | Request berhasil |
71|| `400` | Bad Request | Invalid input, malformed code, validation error |
72|| `404` | Not Found | Problem ID tidak ditemukan |
73|| `429` | Too Many Requests | Rate limit exceeded |
74|| `500` | Internal Server Error | Server-side error tak terduga |
75|| `502` | Bad Gateway | Docker sandbox error / execution failure |
76|| `503` | Service Unavailable | Sandbox unavailable (Docker not running) |
77|
78|---
79|
80|## Endpoints
81|
82|### 1. GET /api/problems
83|
84|Mendaftar semua problem (summary view).
85|
86|#### Query Parameters
87|
88|| Parameter | Type | Required | Keterangan |
89||-----------|------|----------|------------|
90|| `difficulty` | string | ❌ | Filter by difficulty: `easy`, `medium`, `hard` |
91|| `category` | string | ❌ | Filter by category: `array`, `string`, `tree`, `dp`, `graph`, `stack`, etc. |
92|| `type` | string | ❌ | Filter by type: `function`, `main` |
93|
94|#### Example Request
95|
96|```
97|GET /api/problems?difficulty=easy&category=array
98|```
99|
100|#### Response (200 OK)
101|
102|```json
103|{
104|  "data": [
105|    {
106|      "id": "two-sum",
107|      "title": "Two Sum",
108|      "type": "function",
109|      "difficulty": "easy",
110|      "category": "array",
111|      "tags": ["hash-map", "array"]
112|    },
113|    {
114|      "id": "reverse-string",
115|      "title": "Reverse String",
116|      "type": "function",
117|      "difficulty": "easy",
118|      "category": "string",
119|      "tags": ["two-pointers", "string"]
120|    }
121|  ],
122|  "meta": {
123|    "request_id": "req-001",
124|    "timestamp": "2026-06-26T10:30:00Z"
125|  }
126|}
127|```
128|
129|#### Response Fields
130|
131|| Field | Type | Keterangan |
132||-------|------|------------|
133|| `data[].id` | string | Problem identifier |
134|| `data[].title` | string | Human-readable title |
135|| `data[].type` | string | `function` atau `main` |
136|| `data[].difficulty` | string | `easy`, `medium`, atau `hard` |
137|| `data[].category` | string | Problem category |
138|| `data[].tags` | list[string] | Filterable tags |
139|
140|#### Error Responses
141|
142|- `400` — Invalid difficulty or type value
143|  ```json
144|  {
145|    "error": {
146|      "message": "invalid difficulty: 'expert'. Must be one of: easy, medium, hard",
147|      "code": 400,
148|      "type": "validation_error"
149|    }
150|  }
151|  ```
152|- `500` — Server error
153|
154|---
155|
156|### 2. GET /api/problems/:id
157|
158|Detail problem lengkap (tanpa solution).
159|
160|#### Path Parameters
161|
162|| Parameter | Type | Keterangan |
163||-----------|------|------------|
164|| `id` | string | Problem ID (e.g., `two-sum`) |
165|
166|#### Example Request
167|
168|```
169|GET /api/problems/two-sum
170|```
171|
172|#### Response (200 OK) — function-based
173|
174|```json
175|{
176|  "data": {
177|    "id": "two-sum",
178|    "title": "Two Sum",
179|    "type": "function",
180|    "difficulty": "easy",
181|    "category": "array",
182|    "tags": ["hash-map", "array"],
183|    "description": "Given an array of integers `nums` and an integer `target`...",
184|    "function_name": "twoSum",
185|    "parameters": [
186|      {
187|        "name": "nums",
188|        "type": "[]int",
189|        "description": "Array of integers"
190|      },
191|      {
192|        "name": "target",
193|        "type": "int",
194|        "description": "Target sum"
195|      }
196|    ],
197|    "return_type": "[]int",
198|    "examples": [
199|      {
200|        "input": "nums = [2,7,11,15], target = 9",
201|        "output": "[0,1]",
202|        "explanation": "nums[0] + nums[1] = 2 + 7 = 9"
203|      }
204|    ],
205|    "hints": [
206|      {
207|        "level": 1,
208|        "title": "Use extra memory",
209|        "content": "Can you use a data structure to remember what you've seen?"
210|      }
211|    ],
212|    "template": "func twoSum(nums []int, target int) []int {\n    // Your code here\n    return nil\n}",
213|    "test_cases": [
214|      {
215|        "params": [[2, 7, 11, 15], 9],
216|        "expected": [0, 1],
217|        "description": "Basic case"
218|      }
219|    ],
220|    "constraints": [
221|      "2 ≤ len(nums) ≤ 10⁴",
222|      "-10⁹ ≤ nums[i] ≤ 10⁹"
223|    ],
224|    "time_complexity_hint": "O(n)",
225|    "space_complexity_hint": "O(n)"
226|  },
227|  "meta": {
228|    "request_id": "req-002",
229|    "timestamp": "2026-06-26T10:30:00Z"
230|  }
231|}
232|```
233|
234|#### Response (200 OK) — main-based
235|
236|```json
237|{
238|  "data": {
239|    "id": "hello-world",
240|    "title": "Hello World",
241|    "type": "main",
242|    "difficulty": "easy",
243|    "category": "basics",
244|    "tags": ["io", "basics"],
245|    "description": "Read a name from stdin and print a greeting.",
246|    "template": "package main\n\nimport \"fmt\"\n\nfunc main() {\n    // Your code here\n}\n",
247|    "test_cases": [
248|      {
249|        "input": "Alice",
250|        "expected": "Hello, Alice!",
251|        "description": "Basic greeting"
252|      }
253|    ],
254|    "constraints": ["Name is a single word, 1-50 characters"]
255|  },
256|  "meta": {
257|    "request_id": "req-003",
258|    "timestamp": "2026-06-26T10:30:00Z"
259|  }
260|}
261|```
262|
263|> **Note**: Field `solution` TIDAK disertakan dalam response ini untuk mencegah kebocoran solusi.
264|
265|#### Error Responses
266|
267|- `404` — Problem not found
268|  ```json
269|  {
270|    "error": {
271|      "message": "problem 'two-sum' not found",
272|      "code": 404,
273|      "type": "not_found"
274|    }
275|  }
276|  ```
277|
278|---
279|
280|### 3. GET /api/problems/:id/template
281|
282|Get only the code template for a problem (lightweight endpoint for editor).
283|
284|#### Path Parameters
285|
286|| Parameter | Type | Keterangan |
287||-----------|------|------------|
288|| `id` | string | Problem ID |
289|
290|#### Example Request
291|
292|```
293|GET /api/problems/two-sum/template
294|```
295|
296|#### Response (200 OK)
297|
298|```json
299|{
300|  "data": {
301|    "id": "two-sum",
302|    "type": "function",
303|    "template": "func twoSum(nums []int, target int) []int {\n    // Your code here\n    return nil\n}",
304|    "function_name": "twoSum",
305|    "parameters": [
306|      {"name": "nums", "type": "[]int"},
307|      {"name": "target", "type": "int"}
308|    ],
309|    "return_type": "[]int"
310|  },
311|  "meta": {
312|    "request_id": "req-004",
313|    "timestamp": "2026-06-26T10:30:00Z"
314|  }
315|}
316|```
317|
318|#### Response Fields
319|
320|| Field | Type | Keterangan |
321||-------|------|------------|
322|| `data.id` | string | Problem identifier |
323|| `data.type` | string | `function` atau `main` |
324|| `data.template` | string | Code template |
325|| `data.function_name` | string | Function name (function-based only) |
326|| `data.parameters` | list[Parameter] | Parameters (function-based only) |
327|| `data.return_type` | string | Return type (function-based only) |
328|
329|#### Error Responses
330|
331|- `404` — Problem not found
332|
333|---
334|
335|### 4. POST /api/problems/:id/run
336|
337|Execute user code terhadap test cases di sandbox.
338|
339|#### Path Parameters
340|
341|| Parameter | Type | Keterangan |
342||-----------|------|------------|
343|| `id` | string | Problem ID |
344|
345|#### Request Body
346|
347|```json
348|{
349|  "code": "func twoSum(nums []int, target int) []int {\n    seen := make(map[int]int)\n    for i, num := range nums {\n        if j, ok := seen[target-num]; ok {\n            return []int{j, i}\n        }\n        seen[num] = i\n    }\n    return nil\n}"
350|}
351|```
352|
353|| Field | Type | Required | Keterangan |
354||-------|------|----------|------------|
355|| `code` | string | ✅ | Go code (function body for function-based, full program for main-based) |
356|
357|#### Response (200 OK) — All tests passed
358|
359|```json
360|{
361|  "data": {
362|    "success": true,
363|    "compilation_error": null,
364|    "test_results": [
365|      {
366|        "name": "test_1",
367|        "passed": true,
368|        "expected": [0, 1],
369|        "actual": [0, 1],
370|        "error": null,
371|        "execution_time_ms": 12
372|      },
373|      {
374|        "name": "test_2",
375|        "passed": true,
376|        "expected": [1, 2],
377|        "actual": [1, 2],
378|        "error": null,
379|        "execution_time_ms": 8
380|      }
381|    ],
382|    "passed_count": 2,
383|    "total_count": 2,
384|    "execution_time_ms": 145
385|  },
386|  "meta": {
387|    "request_id": "req-005",
388|    "timestamp": "2026-06-26T10:30:00Z"
389|  }
390|}
391|```
392|
393|#### Response (200 OK) — Some tests failed
394|
395|```json
396|{
397|  "data": {
398|    "success": false,
399|    "compilation_error": null,
400|    "test_results": [
401|      {
402|        "name": "test_1",
403|        "passed": true,
404|        "expected": [0, 1],
405|        "actual": [0, 1],
406|        "error": null,
407|        "execution_time_ms": 10
408|      },
409|      {
410|        "name": "test_2",
411|        "passed": false,
412|        "expected": [1, 2],
413|        "actual": [2, 1],
414|        "error": "output mismatch: expected [1,2] got [2,1]",
415|        "execution_time_ms": 11
416|      }
417|    ],
418|    "passed_count": 1,
419|    "total_count": 2,
420|    "execution_time_ms": 132
421|  },
422|  "meta": {
423|    "request_id": "req-006",
424|    "timestamp": "2026-06-26T10:30:00Z"
425|  }
426|}
427|```
428|
429|#### Response (200 OK) — Compilation error
430|
431|```json
432|{
433|  "data": {
434|    "success": false,
435|    "compilation_error": "./main.go:12:5: undefined: x",
436|    "test_results": [],
437|    "passed_count": 0,
438|    "total_count": 0,
439|    "execution_time_ms": 0
440|  },
441|  "meta": {
442|    "request_id": "req-007",
443|    "timestamp": "2026-06-26T10:30:00Z"
444|  }
445|}
446|```
447|
448|#### Response (200 OK) — Runtime error
449|
450|```json
451|{
452|  "data": {
453|    "success": false,
454|    "compilation_error": null,
455|    "test_results": [
456|      {
457|        "name": "test_1",
458|        "passed": false,
459|        "expected": [0, 1],
460|        "actual": null,
461|        "error": "runtime error: index out of range [5] with length 5",
462|        "execution_time_ms": 5
463|      }
464|    ],
465|    "passed_count": 0,
466|    "total_count": 1,
467|    "execution_time_ms": 5
468|  },
469|  "meta": {
470|    "request_id": "req-008",
471|    "timestamp": "2026-06-26T10:30:00Z"
472|  }
473|}
474|```
475|
476|#### Response Fields
477|
478|| Field | Type | Keterangan |
479||-------|------|------------|
480|| `data.success` | boolean | `true` jika semua test passed |
481|| `data.compilation_error` | string \| null | Error message jika gagal compile, `null` jika sukses |
482|| `data.test_results` | list[TestResult] | Hasil per test case |
483|| `data.test_results[].name` | string | Test case identifier |
484|| `data.test_results[].passed` | boolean | Pass atau fail |
485|| `data.test_results[].expected` | any | Expected value (any JSON type) |
486|| `data.test_results[].actual` | any | Actual value dari user code (any JSON type) |
487|| `data.test_results[].error` | string \| null | Error message (runtime error, wrong output, timeout) |
488|| `data.test_results[].execution_time_ms` | int | Execution time for this specific test |
489|| `data.passed_count` | int | Jumlah test yang passed |
490|| `data.total_count` | int | Total test cases |
491|| `data.execution_time_ms` | int | Total execution time dalam milliseconds |
492|
493|#### Error Responses
494|
495|- `400` — Missing or empty code field
496|  ```json
497|  {
498|    "error": {
499|      "message": "code field is required",
500|      "code": 400,
501|      "type": "validation_error",
502|      "details": [
503|        {"field": "code", "issue": "field is required and cannot be empty"}
504|      ]
505|    }
506|  }
507|  ```
508|- `400` — Code exceeds size limit
509|  ```json
510|  {
511|    "error": {
512|      "message": "code exceeds maximum size of 64KB",
513|      "code": 400,
514|      "type": "validation_error"
515|    }
516|  }
517|  ```
518|- `404` — Problem not found
519|  ```json
520|  {
521|    "error": {
522|      "message": "problem 'two-sum' not found",
523|      "code": 404,
524|      "type": "not_found"
525|    }
526|  }
527|  ```
528|- `429` — Rate limit exceeded
529|  ```json
530|  {
531|    "error": {
532|      "message": "rate limit exceeded, try again in 30 seconds",
533|      "code": 429,
534|      "type": "rate_limit_error"
535|    }
536|  }
537|  ```
538|- `502` — Sandbox execution error
539|  ```json
540|  {
541|    "error": {
542|      "message": "sandbox execution failed: timeout exceeded (10s)",
543|      "code": 502,
544|      "type": "execution_error"
545|    }
546|  }
547|  ```
548|- `503` — Sandbox unavailable
549|  ```json
550|  {
551|    "error": {
552|      "message": "sandbox unavailable: Docker daemon not reachable",
553|      "code": 503,
554|      "type": "execution_error"
555|    }
556|  }
557|  ```
558|
559|---
560|
561|### 5. POST /api/validate
562|
563|Validate Go code syntax without executing. Useful for real-time syntax checking in the editor.
564|
565|#### Request Body
566|
567|```json
568|{
569|  "code": "func twoSum(nums []int, target int) []int {\n    // Your code here\n    return nil\n}",
570|  "problem_id": "two-sum"
571|}
572|```
573|
574|| Field | Type | Required | Keterangan |
575||-------|------|----------|------------|
576|| `code` | string | ✅ | Go code to validate |
577|| `problem_id` | string | ❌ | If provided, validates against the problem's function signature |
578|
579|#### Response (200 OK) — Valid
580|
581|```json
582|{
583|  "data": {
584|    "valid": true,
585|    "errors": [],
586|    "warnings": []
587|  },
588|  "meta": {
589|    "request_id": "req-009",
590|    "timestamp": "2026-06-26T10:30:00Z"
591|  }
592|}
593|```
594|
595|#### Response (200 OK) — Invalid
596|
597|```json
598|{
599|  "data": {
600|    "valid": false,
601|    "errors": [
602|      {
603|        "line": 3,
604|        "column": 5,
605|        "message": "undefined: x",
606|        "severity": "error"
607|      }
608|    ],
609|    "warnings": [
610|      {
611|        "line": 1,
612|        "column": 1,
613|        "message": "imported and not used: 'fmt'",
614|        "severity": "warning"
615|      }
616|    ]
617|  },
618|  "meta": {
619|    "request_id": "req-010",
620|    "timestamp": "2026-06-26T10:30:00Z"
621|  }
622|}
623|```
624|
625|#### Response Fields
626|
627|| Field | Type | Keterangan |
628||-------|------|------------|
629|| `data.valid` | boolean | `true` jika code valid |
630|| `data.errors` | list[ValidationError] | Compilation/syntax errors |
631|| `data.errors[].line` | int | Line number (1-indexed) |
632|| `data.errors[].column` | int | Column number (1-indexed) |
633|| `data.errors[].message` | string | Error description |
634|| `data.errors[].severity` | string | `error` atau `warning` |
635|| `data.warnings` | list[ValidationError] | Non-fatal warnings |
636|
637|#### Error Responses
638|
639|- `400` — Missing code field
640|  ```json
641|  {
642|    "error": {
643|      "message": "code field is required",
644|      "code": 400,
645|      "type": "validation_error"
646|    }
647|  }
648|  ```
649|
650|---
651|
652|### 6. GET /api/problems/:id/hints
653|
654|Mendapatkan hints untuk problem tertentu.
655|
656|#### Path Parameters
657|
658|| Parameter | Type | Keterangan |
659||-----------|------|------------|
660|| `id` | string | Problem ID |
661|
662|#### Query Parameters
663|
664|| Parameter | Type | Required | Keterangan |
665||-----------|------|----------|------------|
666|| `level` | int | ❌ | Filter by hint level: 1, 2, atau 3 |
667|
668|#### Example Request
669|
670|```
671|GET /api/problems/two-sum/hints
672|```
673|
674|#### Response (200 OK)
675|
676|```json
677|{
678|  "data": {
679|    "hints": [
680|      {
681|        "level": 1,
682|        "title": "Use extra memory",
683|        "content": "Can you use a data structure to remember what you've seen so far?"
684|      },
685|      {
686|        "level": 2,
687|        "title": "Hash map approach",
688|        "content": "For each element x, check if (target - x) exists in a map."
689|      },
690|      {
691|        "level": 3,
692|        "title": "One-pass hash map",
693|        "content": "Iterate through nums once. For each num: compute complement, check map."
694|      }
695|    ]
696|  },
697|  "meta": {
698|    "request_id": "req-011",
699|    "timestamp": "2026-06-26T10:30:00Z"
700|  }
701|}
702|```
703|
704|#### Response Fields
705|
706|| Field | Type | Keterangan |
707||-------|------|------------|
708|| `data.hints` | list[Hint] | Array of hints, sorted by level ascending |
709|| `data.hints[].level` | int | Hint level (1=vague, 2=directional, 3=detailed) |
710|| `data.hints[].title` | string | Hint title |
711|| `data.hints[].content` | string | Hint content (Markdown supported) |
712|
713|#### Error Responses
714|
715|- `404` — Problem not found
716|  ```json
717|  {
718|    "error": {
719|      "message": "problem 'two-sum' not found",
720|      "code": 404,
721|      "type": "not_found"
722|    }
723|  }
724|  ```
725|
726|---
727|
728|### 7. GET /health
729|
730|Health check endpoint.
731|
732|#### Example Request
733|
734|```
735|GET /health
736|```
737|
738|#### Response (200 OK)
739|
740|```json
741|{
742|  "data": {
743|    "status": "ok",
744|    "uptime_seconds": 342,
745|    "version": "1.0.0",
746|    "sandbox": "available"
747|  },
748|  "meta": {
749|    "request_id": "req-012",
750|    "timestamp": "2026-06-26T10:30:00Z"
751|  }
752|}
753|```
754|
755|#### Response (503 Service Unavailable)
756|
757|```json
758|{
759|  "data": {
760|    "status": "degraded",
761|    "uptime_seconds": 342,
762|    "version": "1.0.0",
763|    "sandbox": "unavailable"
764|  },
765|  "meta": {
766|    "request_id": "req-013",
767|    "timestamp": "2026-06-26T10:30:00Z"
768|  }
769|}
770|```
771|
772|#### Response Fields
773|
774|| Field | Type | Keterangan |
775||-------|------|------------|
776|| `data.status` | string | `"ok"` atau `"degraded"` |
777|| `data.uptime_seconds` | int | Server uptime dalam detik |
778|| `data.version` | string | API version |
779|| `data.sandbox` | string | `"available"` atau `"unavailable"` |
780|
781|---
782|
783|## Rate Limiting
784|
785|| Endpoint | Limit |
786||----------|-------|
787|| `POST /api/problems/:id/run` | 10 requests per minute per IP |
788|| `POST /api/validate` | 30 requests per minute per IP |
789|| All other endpoints | 60 requests per minute per IP |
790|
791|Rate limit exceeded response:
792|
793|```json
794|{
795|  "error": {
796|    "message": "rate limit exceeded, try again in 30 seconds",
797|    "code": 429,
798|    "type": "rate_limit_error"
799|  }
800|}
801|```
802|
803|Rate limit headers included in response:
804|```
805|X-RateLimit-Limit: 10
806|X-RateLimit-Remaining: 7
807|X-RateLimit-Reset: 1719396600
808|```
809|
810|---
811|
812|## Authentication
813|
814|Currently no authentication required (local development). Future versions may implement API key auth via header:
815|
816|```
817|Authorization: Bearer ***
818|```
819|
820|---
821|
822|## Error Type Reference
823|
824|| Type | HTTP Code | Description |
825||------|-----------|-------------|
826|| `validation_error` | 400 | Input validation failed |
827|| `not_found` | 404 | Resource not found |
828|| `compilation_error` | 200 | Code failed to compile (in response body) |
829|| `execution_error` | 502 | Sandbox execution failed |
830|| `rate_limit_error` | 429 | Too many requests |
831|| `internal_error` | 500 | Unexpected server error |
832|