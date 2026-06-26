1|# 🚀 Coding Challenge Platform - Panduan Menjalankan
2|
3|> Platform coding challenge seperti HackerRank dengan compiler Go, real-time execution, hints, dan leaderboard.
4|
5|---
6|
7|## 📋 Daftar Isi
8|
9|1. [Quick Start (5 Menit)](#quick-start-5-menit)
10|2. [Prasyarat](#prasyarat)
11|3. [Instalasi](#instalasi)
12|4. [Menjalankan dengan Docker (Rekomendasi)](#menjalankan-dengan-docker)
13|5. [Menjalankan Manual (Development)](#menjalankan-manual-development)
14|6. [Struktur Project](#struktur-project)
15|7. [Konfigurasi](#konfigurasi)
16|8. [Verifikasi](#verifikasi)
17|9. [Troubleshooting](#troubleshooting)
18|10. [Development Guide](#development-guide)
19|
20|---
21|
22|## ⚡ Quick Start (5 Menit)
23|
24|```bash
25|# 1. Clone / masuk ke direktori project
26|cd C:\Users\PC\go\src\project\coding-challange
27|
28|# 2. Copy file config
29|cp .env.example .env
30|
31|# 3. Jalankan semua services dengan Docker
32|docker-compose up -d --build
33|
34|# 4. Setup database (migrate + seed)
35|make migrate
36|
37|# 5. Buka browser
38|start http://localhost:9100
39|```
40|
41|**Selesai!** Aplikasi berjalan di `http://localhost:9100`
42|
43|---
44|
45|## 📦 Prasyarat
46|
47|### Wajib
48|| Tool | Versi | Download |
49||------|-------|----------|
50|| Go | 1.21+ | [golang.org](https://golang.org/dl/) |
51|| Docker Desktop | 24+ | [docker.com](https://www.docker.com/products/docker-desktop/) |
52|| Docker Compose | v2+ | Sudah include di Docker Desktop |
53|
54|### Opsional (untuk development)
55|| Tool | Versi | Kegunaan |
56||------|-------|----------|
57|| Node.js | 18+ | Frontend development |
58|| npm | 9+ | Frontend package manager |
59|| pgAdmin | 4+ | Database management |
60|| RedisInsight | 2+ | Redis monitoring |
61|
62|### System Requirements
63|- **RAM:** Minimal 8GB (untuk Docker)
64|- **Disk:** Minimal 10GB free
65|- **OS:** Windows 10/11 (WSL2), macOS, Linux
66|
67|---
68|
69|## 💻 Instalasi
70|
71|### 1. Clone Repository
72|
73|```bash
74|cd C:\Users\PC\go\src\project
75|git clone <repository-url> coding-challange
76|cd coding-challange
77|```
78|
79|### 2. Download Dependencies
80|
81|```bash
82|# Backend (Go)
83|go mod download
84|go mod tidy
85|
86|# Frontend (Node.js)
87|cd web/frontend
88|npm install
89|cd ../..
90|```
91|
92|### 3. Setup Environment
93|
94|```bash
95|# Copy file konfigurasi
96|cp .env.example .env
97|
98|# Edit .env jika diperlukan (lihat bagian Konfigurasi)
99|notepad .env
100|```
101|
102|---
103|
104|## 🐳 Menjalankan dengan Docker (Rekomendasi)
105|
106|### Start Semua Services
107|
108|```bash
109|# Build dan start semua container
110|docker-compose up -d --build
111|
112|# Lihat status semua services
113|docker-compose ps
114|
115|# Lihat logs
116|docker-compose logs -f
117|```
118|
119|### Port Mapping
120|
121|| Service | Port | URL |
122||---------|------|-----|
123|| API Gateway | 9100 | http://localhost:9100 |
124|| Frontend (React) | 5173 | http://localhost:5173 |
125|| Auth Service | 8081 | http://localhost:9101 |
126|| Problem Service | 8082 | http://localhost:9102 |
127|| Execution Service | 8083 | http://localhost:9103 |
128|| Leaderboard Service | 8084 | http://localhost:9104 |
129|| Hint Service | 8085 | http://localhost:9105 |
130|| Execution Worker | 8086 | http://localhost:9106 |
131|| WebSocket Service | 8087 | http://localhost:9107 |
132|| PostgreSQL | 5432 | localhost:5432 |
133|| Redis | 6379 | localhost:6379 |
134|| RabbitMQ | 5672/15672 | localhost:15672 (management) |
135|| Prometheus (opsional) | 9090 | http://localhost:9090 |
136|| Grafana (opsional) | 3000 | http://localhost:3000 |
137|
138|### Docker Commands
139|
140|```bash
141|# Start services
142|docker-compose up -d
143|
144|# Stop services
145|docker-compose down
146|
147|# Restart specific service
148|docker-compose restart auth-service
149|
150|# View logs specific service
151|docker-compose logs -f execution-service
152|
153|# Rebuild specific service
154|docker-compose build --no-cache execution-worker
155|
156|# Stop dan hapus volumes (RESET DATABASE)
157|docker-compose down -v
158|
159|# Scale workers
160|docker-compose up -d --scale execution-worker=3
161|
162|# Enable monitoring (Prometheus + Grafana)
163|docker-compose --profile monitoring up -d
164|
165|# Check health
166|docker-compose exec postgres pg_isready -U postgres
167|```
168|
169|### Docker Compose Format (PowerShell / cmd.exe)
170|
171|Jika `**` glob pattern tidak bekerja di cmd.exe, ganti dengan absolute path:
172|```cmd
173|cd C:\Users\PC\go\src\project\coding-challange
174|docker-compose -f "C:\Users\PC\go\src\project\coding-challange\docker-compose.yml" up -d --build
175|```
176|
177|PowerShell juga perlu escape. Gunakan:
178|```powershell
179|Set-Location "C:\Users\PC\go\src\project\coding-challange"
180|docker-compose up -d --build
181|```
182|
183|---
184|
185|## 🔧 Menjalankan Manual (Development)
186|
187|### Terminal 1: Start Infrastructure
188|
189|Jika hanya butuh PostgreSQL, Redis, RabbitMQ di Docker:
190|
191|```bash
192|# Start hanya infrastructure
193|docker-compose up -d postgres redis rabbitmq
194|
195|# Atau install manual di Windows:
196|# PostgreSQL: https://www.postgresql.org/download/windows/
197|# Redis: https://github.com/microsoftarchive/redis/releases
198|# RabbitMQ: https://www.rabbitmq.com/download.html
199|```
200|
201|### Terminal 2: Run API Gateway
202|
203|```bash
204|go run services/api-gateway/main.go
205|# Berjalan di :9100
206|```
207|
208|### Terminal 3: Run Auth Service
209|
210|```bash
211|go run services/auth-service/main.go
212|# Berjalan di :9101
213|```
214|
215|### Terminal 4: Run Problem Service
216|
217|```bash
218|go run services/problem-service/main.go
219|# Berjalan di :9102
220|```
221|
222|### Terminal 5: Run Execution Service
223|
224|```bash
225|go run services/execution-service/main.go
226|# Berjalan di :9103
227|```
228|
229|### Terminal 6: Run Execution Worker
230|
231|```bash
232|go run services/execution-worker/main.go
233|# Berjalan di :9106
234|```
235|
236|### Terminal 7: Run WebSocket Service
237|
238|```bash
239|go run services/websocket-service/main.go
240|# Berjalan di :9107
241|```
242|
243|### Terminal 8: Run Leaderboard Service
244|
245|```bash
246|go run services/leaderboard-service/main.go
247|# Berjalan di :9104
248|```
249|
250|### Terminal 9: Run Hint Service
251|
252|```bash
253|go run services/hint-service/main.go
254|# Berjalan di :9105
255|```
256|
257|### Terminal 10: Run Frontend
258|
259|```bash
260|cd web/frontend
261|npm run dev
262|# Berjalan di :5173
263|```
264|
265|---
266|
267|## 📁 Struktur Project
268|
269|```
270|coding-challange/
271|├── 📄 README.md                    # Project overview
272|├── 📄 Makefile                     # Build & run commands
273|├── 📄 docker-compose.yml           # Docker orchestration
274|├── 📄 .env.example                 # Template environment variables
275|│
276|├── 📁 cmd/
277|│   └── 📁 migrate/                 # Database migration tool
278|│
279|├── 📁 docs/
280|│   ├── 📄 ARCHITECTURE.md           # Arsitektur system v1
281|│   ├── 📄 ARCHITECTURE_V2.md        # Arsitektur system v2 (microservices)
282|│   ├── 📄 API.md                    # REST API documentation
283|│   ├── 📄 DATABASE_SCHEMA.md        # Database schema documentation
284|│   ├── 📄 PROBLEM_SCHEMA.md         # Problem YAML schema
285|│   ├── 📄 RUNBOOK.md                # Operational runbook
286|│   └── 📄 DEPLOYMENT.md             # Deployment guide
287|│
288|├── 📁 internal/
289|│   ├── 📁 config/                   # App configuration
290|│   │   └── 📄 config.go
291|│   ├── 📁 model/                    # Data models
292|│   │   └── 📄 problem.go
293|│   ├── 📁 repository/               # Data access layer
294|│   │   ├── 📄 problem.go
295|│   │   ├── 📄 optimizations.go
296|│   │   ├── 📄 migrations/           # SQL migrations
297|│   │   └── 📄 schema.sql            # Full schema
298|│   ├── 📁 service/                  # Business logic
299|│   │   ├── 📄 problem.go
300|│   │   ├── 📄 runner.go             # Code execution
301|│   │   ├── 📄 hint.go               # Hint system
302|│   │   ├── 📄 runner_test.go
303|│   │   └── 📄 hint_test.go
304|│   ├── 📁 handler/                  # HTTP handlers
305|│   │   ├── 📄 problem.go
306|│   │   └── 📄 problem_test.go
307|│   └── 📁 middleware/               # Shared middleware
308|│
309|├── 📁 pkg/
310|│   ├── 📁 database/                 # DB connection pool
311|│   │   └── 📄 postgres.go
312|│   ├── 📁 redis/                    # Redis client
313|│   │   └── 📄 client.go
314|│   ├── 📁 rabbitmq/                 # RabbitMQ client
315|│   │   └── 📄 client.go
316|│   ├── 📁 middleware/               # Shared middleware
317|│   │   ├── 📄 auth.go
318|│   │   └── 📄 auth_test.go
319|│   ├── 📁 websocket/                # WebSocket hub
320|│   │   └── 📄 hub.go
321|│   ├── 📁 logger/                   # Structured logger
322|│   │   ├── 📄 logger.go
323|│   │   └── 📄 logger_test.go
324|│   ├── 📁 security/                 # Security utilities
325|│   │   └── 📄 security.go
326|│   ├── 📁 health/                   # Health checker
327|│   │   └── 📄 checker.go
328|│   ├── 📁 errors/                   # Error handling
329|│   │   └── 📄 errors.go
330|│   └── 📁 config/                   # Config management
331|│       └── 📄 config.go
332|│
333|├── 📁 services/
334|│   ├── 📁 api-gateway/              # API Gateway
335|│   │   ├── 📄 main.go
336|│   │   └── 📄 nginx.conf
337|│   ├── 📁 auth-service/             # Authentication
338|│   │   ├── 📄 main.go
339|│   │   └── 📄 Dockerfile
340|│   ├── 📁 problem-service/          # Problem management
341|│   │   ├── 📄 main.go
342|│   │   └── 📄 Dockerfile
343|│   ├── 📁 execution-service/        # Code execution
344|│   │   ├── 📄 main.go
345|│   │   └── 📄 Dockerfile
346|│   ├── 📁 execution-worker/         # Execution worker
347|│   │   ├── 📄 main.go
348|│   │   ├── 📄 harness.go
349|│   │   ├── 📄 sandbox.go
350|│   │   ├── 📄 dp-visualizer.go
351|│   │   └── 📄 Dockerfile
352|│   ├── 📁 websocket-service/        # WebSocket service
353|│   │   ├── 📄 main.go
354|│   │   └── 📄 Dockerfile
355|│   ├── 📁 leaderboard-service/      # Leaderboard
356|│   │   ├── 📄 main.go
357|│   │   └── 📄 Dockerfile
358|│   └── 📁 hint-service/             # Hint management
359|│       ├── 📄 main.go
360|│       └── 📄 Dockerfile
361|│
362|├── 📁 problems/                     # Problem bank (YAML)
363|│   ├── 📁 easy/
364|│   │   ├── 📄 two-sum.yaml
365|│   │   ├── 📄 fizz-buzz.yaml
366|│   │   ├── 📄 reverse-string.yaml
367|│   │   ├── 📄 max-subarray.yaml
368|│   │   └── 📄 contains-duplicate.yaml
369|│   ├── 📁 medium/
370|│   │   ├── 📄 valid-parentheses.yaml
371|│   │   └── 📄 merge-sorted-arrays.yaml
372|│   └── 📁 hard/
373|│       ├── 📄 longest-palindromic-substring.yaml
374|│       ├── 📄 trapping-rain-water.yaml
375|│       └── 📄 coin-change.yaml
376|│
377|├── 📁 web/
378|│   └── 📁 frontend/                 # React + TypeScript frontend
379|│       ├── 📁 src/
380|│       │   ├── 📁 components/
381|│       │   ├── 📁 pages/
382|│       │   ├── 📁 store/
383|│       │   ├── 📁 hooks/
384|│       │   ├── 📁 services/
385|│       │   ├── 📁 types/
386|│       │   └── 📁 styles/
387|│       ├── 📄 package.json
388|│       ├── 📄 vite.config.ts
389|│       └── 📄 tailwind.config.js
390|│
391|├── 📁 deployments/
392|│   └── 📁 kubernetes/
393|│       ├── 📄 sandbox-pod.yaml
394|│       ├── 📄 sandbox-network-policy.yaml
395|│       ├── 📄 worker-deployment.yaml
396|│       ├── 📄 rabbitmq-deployment.yaml
397|│       ├── 📄 redis-deployment.yaml
398|│       ├── 📄 postgres-deployment.yaml
399|│       └── 📄 ingress.yaml
400|│
401|├── 📁 tests/
402|│   ├── 📁 integration/
403|│   ├── 📁 load/
404|│   └── 📁 harness/
405|│
406|├── 📁 runner/
407|│   └── 📄 Dockerfile.runner          # Sandbox runner image
408|│
409|└── 📁 hermes/                        # Hermes agent setup
410|```
411|
412|---
413|
414|## ⚙️ Konfigurasi
415|
416|### File `.env`
417|
418|```env
419|# ============================================
420|# Server Configuration
421|# ============================================
422|PORT=8080
423|LOG_LEVEL=debug  # debug, info, warn, error
424|
425|# ============================================
426|# Database (PostgreSQL)
427|# ============================================
428|DB_HOST=localhost
429|DB_PORT=5432
430|DB_USER=postgres
431|DB_PASSWORD=postgres
432|DB_NAME=coding_challange
433|DB_MAX_CONNS=25
434|DB_IDLE_CONNS=5
435|
436|# ============================================
437|# Redis
438|# ============================================
439|REDIS_ADDR=localhost:6379
440|REDIS_PASSWORD=
441|REDIS_DB=0
442|REDIS_POOL_SIZE=10
443|
444|# ============================================
445|# RabbitMQ
446|# ============================================
447|RABBITMQ_HOST=localhost
448|RABBITMQ_PORT=5672
449|RABBITMQ_USER=guest
450|RABBITMQ_PASSWORD=guest
451|RABBITMQ_VHOST=/
452|
453|# ============================================
454|# JWT Authentication
455|# ============================================
456|JWT_ACCESS_SECRET=your-super-secret-access-key-change-in-production
457|JWT_REFRESH_SECRET=your-super-secret-refresh-key-change-in-production
458|JWT_ACCESS_TTL_MIN=15        # Access token expiry (minutes)
459|JWT_REFRESH_TTL_HOURS=168    # Refresh token expiry (hours = 7 days)
460|
461|# ============================================
462|# Code Execution
463|# ============================================
464|SANDBOX_IMAGE=golang:1.21-alpine
465|SANDBOX_MEMORY_MB=256
466|SANDBOX_TIME_SEC=10
467|DISABLE_LOCAL_FALLBACK=false
468|MAX_OUTPUT_SIZE_MB=1
469|MAX_CODE_SIZE_KB=64
470|
471|# ============================================
472|# Problem Loader
473|# ============================================
474|PROBLEMS_DIR=./problems
475|
476|# ============================================
477|# Rate Limiting
478|# ============================================
479|RATE_LIMIT_REQUESTS=10
480|RATE_LIMIT_WINDOW_SECONDS=60
481|SUBMIT_RATE_LIMIT=10
482|SUBMIT_RATE_WINDOW_SECONDS=60
483|
484|# ============================================
485|# Frontend
486|# ============================================
487|VITE_API_URL=http://localhost:9100
488|VITE_WS_URL=ws://localhost:9107
489|```
490|
491|### Konfigurasi per Environment
492|
493|**Development** (`.env.development`)
494|```env
495|LOG_LEVEL=debug
496|DISABLE_LOCAL_FALLBACK=true
497|DB_HOST=localhost
498|```
499|
500|**Production** (`.env.production`)
501|```env
502|LOG_LEVEL=warn
503|DISABLE_LOCAL_FALLBACK=false
504|DB_HOST=postgres-cluster.internal
505|JWT_ACCESS_SECRET=<generate-strong-secret>
506|```
507|
508|**Testing** (`.env.test`)
509|```env
510|DB_NAME=coding_challange_test
511|LOG_LEVEL=error
512|JWT_ACCESS_TTL_MIN=5
513|```
514|
515|---
516|
517|## ✅ Verifikasi
518|
519|### 1. Cek Health Services
520|
521|```bash
522|# Via Makefile
523|make health
524|
525|# Manual (PowerShell / cmd.exe)
526|curl http://localhost:9100/health
527|curl http://localhost:9101/health
528|curl http://localhost:9102/health
529|curl http://localhost:9103/health
530|curl http://localhost:9104/health
531|curl http://localhost:9105/health
532|curl http://localhost:9106/health
533|```
534|
535|### 2. Cek API Gateway
536|
537|```bash
538|# Get problem list
539|curl http://localhost:9100/api/problems
540|
541|# Expected: {"data":[...],"meta":{...}}
542|```
543|
544|=== USER ===
545|### 3. Cek Infrastructure
546|
547|```bash
548|# PostgreSQL
549|docker-compose exec postgres psql -U postgres -d coding_challange -c "\dt"
550|
551|# Redis
552|docker-compose exec redis redis-cli ping
553|
554|# RabbitMQ
555|curl -u guest:guest http://localhost:15672/api/overview
556|```
557|
558|### 4. Run Tests
559|
560|```bash
561|# All tests
562|go test ./... -v
563|
564|# Specific package
565|go test ./internal/service/... -v
566|
567|# With coverage
568|make test-coverage
569|
570|# Benchmarks
571|go test -bench=. ./benchmarks/
572|```
573|
574|---
575|
576|## 🔧 Troubleshooting
577|
578|### Docker Issues
579|
580|**Issue:** `port is already allocated`
581|```bash
582|# Find process using port
583|netstat -ano | findstr :9100
584|# Kill process
585|taskkill /F /PID <PID>
586|
587|# Atau ubah port di .env atau docker-compose.yml
588|```
589|
590|**Issue:** `docker-compose up` gagal
591|```bash
592|# Reset everything
593|docker-compose down -v
594|docker system prune -a
595|docker-compose up -d --build
596|```
597|
598|**Issue:** Container tidak bisa connect ke service lain
599|```bash
600|# Check network
601|docker network ls
602|docker network inspect coding-challange_coding-challange
603|
604|# Restart
605|docker-compose restart
606|EOF
607|```
608|
609|### Build Issues
610|
611|**Issue:** `go build` error
612|```bash
613|# Clean and redo
614|go clean -cache
615|go mod download
616|go mod tidy
617|go build ./...
618|```
619|
620|### Database Issues
621|
622|**Issue:** Migration gagal
623|```bash
624|# Reset database
625|docker-compose down -v
626|docker-compose up -d postgres
627|make migrate
628|```
629|
630|**Issue:** Cannot connect to PostgreSQL
631|```bash
632|# Check if postgres is running
633|docker-compose ps postgres
634|
635|# Check logs
636|docker-compose logs postgres
637|
638|# Connect manually
639|docker-compose exec postgres psql -U postgres
640|```
641|
642|---
643|
644|## 🛠 Development Guide
645|
646|### Menambah Problem Baru
647|
648|1. Buat file YAML di `problems/<difficulty>/`
649|
650|2. Format YAML:
651|```yaml
652|id: "problem-slug"
653|title: "Problem Title"
654|difficulty: "easy"  # easy, medium, hard
655|category: "array"
656|tags: ["hash-map", "two-pointers"]
657|
658|description: |
659|  Problem description here...
660|  Supports **Markdown**.
661|
662|examples:
663|  - input: "input description"
664|    output: "expected output"
665|    explanation: "explanation here"
666|
667|hints:
668|  - level: 1
669|    title: "Approach hint"
670|    content: "Use a hash map..."
671|  - level: 2
672|    title: "Technical hint"
673|    content: "Store seen values..."
674|  - level: 3
675|    title: "Advanced hint"
676|    content: "One-pass: for each x, check target-x..."
677|
678|template: |
679|  function problemFunction(params) {
680|      // Your code
681|      return null;
682|  }
683|
684|test_cases:
685|  - input: "[1,2,3], 5"
686|    expected: "[0,1]"
687|    description: "Basic case"
688|  - input: "[1], 2"
689|    expected: "[]"
690|    description: "Single element"
691|
692|constraints:
693|  - "1 ≤ n ≤ 10⁴"
694|  - "-10⁹ ≤ nums[i] ≤ 10⁹"
695|
696|solution:
697|  code: |
698|    function problemFunction(params) {
699|      // Solution here
700|    }
701|  approach: "Two pointers"
702|  time_complexity: "O(n)"
703|  space_complexity: "O(n)"
704|```
705|
706|3. Restart problem service atau akan auto-reload di development mode.
707|
708|### Menambah Endpoint API
709|
710|1. Tambah handler di `internal/handler/`
711|2. Tambah service method di `internal/service/`
712|3. Tambah route di `cmd/server/main.go`
713|4. Test dengan curl
714|
715|### Menambah Frontend Component
716|
717|1. Buat file di `web/frontend/src/components/`
718|2. Export di `web/frontend/src/components/index.ts`
719|3. Import di page yang membutuhkan
720|
721|### Git Workflow
722|
723|```bash
724|# Buat branch baru
725|git checkout -b feature/your-feature
726|
727|# Commit changes
728|git add .
729|git commit -m "feat: add your feature"
730|
731|# Push
732|git push origin feature/your-feature
733|
734|# Pull request di GitHub
735|```
736|
737|---
738|
739|## 📞 Bantuan
740|
741|Jika mengalami masalah:
742|
743|1. Cek logs: `docker-compose logs -f <service-name>`
744|2. Cek dokumentasi di `docs/`
745|3. Run `make health` untuk cek semua services
746|4. Buka issue di GitHub repository
747|
748|---
749|
750|**Selamat coding! 🚀**
751|