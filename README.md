1|# 🚀 Coding Challenge Platform
2|
3|> Platform coding challenge seperti HackerRank dengan compiler Go, real-time execution, progressive hints, dan leaderboard.
4|
5|![Go](https://img.shields.io/badge/Go-1.25+-00ADD8?style=flat&logo=go)
6|![React](https://img.shields.io/badge/React-18+-61DAFB?style=flat&logo=react)
7|![PostgreSQL](https://img.shields.io/badge/PostgreSQL-15+-336791?style=flat&logo=postgresql)
8|![Redis](https://img.shields.io/badge/Redis-7+-DC382D?style=flat&logo=redis)
9|![RabbitMQ](https://img.shields.io/badge/RabbitMQ-3.12-FF6600?style=flat&logo=rabbitmq)
10|![Docker](https://img.shields.io/badge/Docker-24+-2496ED?style=flat&logo=docker)
11|![Kubernetes](https://img.shields.io/badge/Kubernetes-1.28+-326CE5?style=flat&logo=kubernetes)
12|![License](https://img.shields.io/badge/License-MIT-green.svg)
13|
14|---
15|
16|## 📋 Daftar Isi
17|
18|- [Fitur Utama](#fitur-utama)
19|- [Tech Stack](#tech-stack)
20|- [Quick Start](#quick-start)
21|- [Dokumentasi](#dokumentasi)
22|- [Arsitektur](#arsitektur)
23|- [API Reference](#api-reference)
24|- [Development](#development)
25|- [Deployment](#deployment)
26|- [Kontribusi](#kontribusi)
27|- [Lisensi](#lisensi)
28|
29|---
30|
31|## ✨ Fitur Utama
32|
33|### Untuk User (Peserta)
34|- ✅ **15+ Soal** — Algoritma, Data Structure, Dynamic Programming
35|- ✅ **3 Tingkat Kesulitan** — Easy, Medium, Hard
36|- ✅ **Real-time Code Execution** — Compiler Go dengan feedback instan
37|- ✅ **Progressive Hint System** — 3 level hints per soal
38|- ✅ **Leaderboard** — Weekly, Monthly, All-time ranking
39|- ✅ **Code Editor** — Monaco Editor dengan syntax highlighting
40|- ✅ **Submission History** — Track semua submission
41|- ✅ **Profil & Statistik** — Rating, streak, solved problems
42|
43|### Untuk Admin
44|- ✅ **Problem Management** — CRUD soal dengan test cases
45|- ✅ **User Management** — Ban/unban, role management
46|- ✅ **Monitoring** — Dashboard dengan statistik real-time
47|- ✅ **Log Viewer** — View dan filter logs
48|
49|### Teknis
50|- ✅ **Microservices Architecture** — Scalable, maintainable
51|- ✅ **Docker Sandbox** — Secure code execution
52|- ✅ **WebSocket** — Real-time submission updates
53|- ✅ **Redis Caching** — Fast problem list loading
54|- ✅ **RabbitMQ** — Async code execution queue
55|- ✅ **JWT Authentication** — Secure auth dengan refresh tokens
56|- ✅ **Rate Limiting** — Prevent abuse
57|- ✅ **Graceful Shutdown** — Zero-downtime deployment
58|
59|---
60|
61|## 🛠 Tech Stack
62|
63|| Layer | Technology |
64||-------|-----------|
65|| **Backend** | Go 1.25+ (Gin framework) |
66|| **Frontend** | React 18+, TypeScript, Tailwind CSS |
67|| **Code Editor** | Monaco Editor |
68|| **Database** | PostgreSQL 15+ |
69|| **Cache** | Redis 7+ |
70|| **Message Queue** | RabbitMQ 3.12 |
71|| **Authentication** | JWT (access + refresh tokens) |
72|| **Container** | Docker 24+ |
73|| **Orchestration** | Kubernetes 1.28+ |
74|| **Monitoring** | Prometheus + Grafana |
75|| **CI/CD** | GitHub Actions |
76|
77|---
78|
79|## 🚀 Quick Start
80|
81|### Opsi 1: Docker Compose (Rekomendasi)
82|
83|```bash
84|# Clone repository
85|git clone https://github.com/codingchallenge/platform.git
86|cd platform
87|
88|# Start semua services
89|docker-compose up -d --build
90|
91|# Buka di browser
92|open http://localhost:9100
93|```
94|
95|Selengkapnya lihat [docs/HOW_TO_RUN.md](docs/HOW_TO_RUN.md)
96|
97|### Opsi 2: Manual Development
98|
99|```bash
100|# 1. Install dependencies
101|go mod download
102|
103|# 2. Setup infrastructure (PostgreSQL, Redis, RabbitMQ)
104|docker-compose up -d postgres redis rabbitmq
105|
106|# 3. Run migrations
107|make migrate
108|
109|# 4. Start services (di terminal terpisah)
110|go run services/auth-service/main.go
111|go run services/problem-service/main.go
112|go run services/execution-service/main.go
113|go run services/leaderboard-service/main.go
114|go run services/hint-service/main.go
115|
116|# 5. Start frontend (di terminal lain)
117|cd web/frontend
118|npm install
119|npm run dev
120|
121|# 6. Buka http://localhost:5173
122|```
123|
124|Selengkapnya lihat [docs/HOW_TO_RUN.md](docs/HOW_TO_RUN.md)
125|
126|---
127|
128|## 📖 Dokumentasi
129|
130|| Dokumen | Deskripsi |
131||---------|-----------|
132|| [docs/HOW_TO_RUN.md](docs/HOW_TO_RUN.md) | Panduan lengkap menjalankan aplikasi |
133|| [docs/HOW_TO_USE.md](docs/HOW_TO_USE.md) | Panduan menggunakan aplikasi (user & admin) |
134|| [docs/ARCHITECTURE_V2.md](docs/ARCHITECTURE_V2.md) | Diagram arsitektur dan penjelasan |
135|| [docs/DATABASE_SCHEMA.md](docs/DATABASE_SCHEMA.md) | Schema database dan query examples |
136|| [docs/TEST_PLAN.md](docs/TEST_PLAN.md) | Test plan dan test cases |
137|| [docs/CODE_REVIEW.md](docs/CODE_REVIEW.md) | Code review findings |
138|| [docs/API.md](docs/API.md) | API documentation |
139|
140|---
141|
142|## 🏗 Arsitektur
143|
144|```
145|┌─────────────────────────────────────────────────────────────────┐
146|│                        Browser (Client)                          │
147|│  ┌──────────────┐  ┌──────────────┐  ┌───────────────────────┐ │
148|│  │ Problem List │  │ Code Editor  │  │ Test Results + Hints │ │
149|│  │   (React)    │  │  (Monaco)    │  │   (Real-time WS)     │ │
150|│  └──────────────┘  └──────────────┘  └───────────────────────┘ │
151|└──────────────────────────┬──────────────────────────────────────┘
152|                           │ HTTP / WebSocket
153|┌──────────────────────────▼──────────────────────────────────────┐
154|│                       API Gateway (Nginx)                        │
155|│  • Rate Limiting • JWT Validation • Load Balancing • CORS       │
156|└──────┬────────┬────────┬────────┬────────┬────────┬─────────────┘
157|       │        │        │        │        │        │
158|┌──────▼──┐ ┌───▼────┐ ┌─▼──────┐ ┌─▼────┐ ┌─▼────┐ ┌─────────▼──┐
159|│  Auth   │ │Problem │ │Executi-│ │Leader│ │Hint  │ │  WebSocket  │
160|│ Service │ │Service │ │  on    │ │board │ │Servi-│ │   Service   │
161|│  :9101  │ │ :9102  │ │Service │ │:9104 │ │ce    │ │    :9107    │
162|└────┬────┘ └───┬────┐ │ :9103  │ └──┬───┘ │:9105 │ └─────────────┘
163|     │         │    │ └──┬─────┘    │     └──────┘
164|     │         │    │    │          │
165|     └────┬────┴────┴────┴────┬─────┘
166|          │                   │
167|┌─────────▼───────────────────▼────────────────────────────────────┐
168|│                    Data Layer                                     │
169|│  ┌──────────────┐  ┌──────────────┐  ┌──────────────┐          │
170|│  │  PostgreSQL  │  │    Redis     │  │   RabbitMQ   │          │
171|│  │  (Primary)   │  │   (Cache)    │  │  (Job Queue) │          │
172|│  └──────────────┘  └──────────────┘  └──────────────┘          │
173|└──────────────────────────────────────────────────────────────────┘
174|                           │
175|┌──────────────────────────▼──────────────────────────────────────┐
176|│                  Code Execution (Docker Sandbox)                  │
177|│  • --network=none • --read-only • --memory=256m • --cpus=1     │
178|└─────────────────────────────────────────────────────────────────┘
179|```
180|
181|---
182|
183|## 🔌 API Reference
184|
185|### Base URL
186|```
187|Development: http://localhost:9100/api
188|Production:  https://api.codingchallenge.com/api
189|```
190|
191|### Endpoints
192|
193|#### Auth
194|| Method | Endpoint | Deskripsi |
195||--------|----------|-----------|
196|| POST | `/auth/register` | Registrasi |
197|| POST | `/auth/login` | Login |
198|| POST | `/auth/refresh` | Refresh token |
199|
200|#### Problems
201|| Method | Endpoint | Deskripsi |
202||--------|----------|-----------|
203|| GET | `/api/problems` | List soal |
204|| GET | `/api/problems/:id` | Detail soal |
205|| POST | `/api/problems` | Buat soal (Admin) |
206|
207|#### Submissions
208|| Method | Endpoint | Deskripsi |
209||--------|----------|-----------|
210|| POST | `/api/submissions` | Submit kode |
211|| GET | `/api/submissions/:id` | Get status |
212|
213|#### Leaderboard
214|| Method | Endpoint | Deskripsi |
215||--------|----------|-----------|
216|| GET | `/api/leaderboard/weekly` | Peringkat mingguan |
217|| GET | `/api/leaderboard/monthly` | Peringatan bulanan |
218|| GET | `/api/leaderboard/all-time` | Peringkat keseluruhan |
219|
220|#### Hints
221|| Method | Endpoint | Deskripsi |
222||--------|----------|-----------|
223|| GET | `/api/problems/:id/hints` | Get hints |
224|| POST | `/api/problems/:id/hints/:hintId/use` | Gunakan hint |
225|
226|---
227|
228|## 🗂 Struktur Project
229|
230|```
231|coding-challange/
232|├── 📂 docs/                           # Dokumentasi
233|│   ├── HOW_TO_RUN.md                  # Panduan menjalankan
234|│   ├── HOW_TO_USE.md                  # Panduan menggunakan
235|│   ├── ARCHITECTURE_V2.md             # Arsitektur sistem
236|│   ├── DATABASE_SCHEMA.md             # Database schema
237|│   ├── API.md                         # API documentation
238|│   ├── TEST_PLAN.md                   # Test plan
239|│   └── CODE_REVIEW.md                 # Code review
240|│
241|├── 📂 services/                       # Microservices
242|│   ├── api-gateway/                   # API Gateway (Nginx)
243|│   ├── auth-service/                  # Authentication service
244|│   ├── problem-service/               # Problem management
245|│   ├── execution-service/             # Code execution orchestrator
246|│   ├── execution-worker/              # Code execution worker
247|│   ├── websocket-service/             # WebSocket real-time
248|│   ├── leaderboard-service/           # Ranking & scoring
249|│   └── hint-service/                  # Hint management
250|│
251|├── 📂 pkg/                            # Shared packages
252|│   ├── database/                      # PostgreSQL connection
253|│   ├── redis/                         # Redis client
254|│   ├── rabbitmq/                      # RabbitMQ client
255|│   ├── middleware/                    # Auth, CORS, rate limiting
256|│   ├── websocket/                     # WebSocket hub
257|│   ├── logger/                        # Structured logging
258|│   ├── config/                        # Configuration
259|│   ├── security/                      # Security utilities
260|│   ├── errors/                        # Error handling
261|│   └── health/                        # Health checks
262|│
263|├── 📂 internal/                       # Internal packages
264|│   ├── handler/                       # HTTP handlers (legacy)
265|│   ├── service/                       # Business logic (legacy)
266|│   ├── repository/                    # Data access layer
267|│   ├── model/                         # Data models
268|│   └── config/                        # App config
269|│
270|├── 📂 web/                            # Frontend
271|│   └── frontend/                      # React SPA
272|│       ├── src/
273|│       │   ├── components/            # UI components
274|│       │   ├── pages/                 # Page components
275|│       │   ├── hooks/                 # Custom hooks
276|│       │   ├── services/              # API client
277|│       │   ├── store/                 # State management
278|│       │   ├── types/                 # TypeScript types
279|│       │   └── styles/                # CSS/Tailwind
280|│       ├── public/                    # Static assets
281|│       └── package.json
282|│
283|├── 📂 problems/                       # Problem bank (YAML)
284|│   ├── easy/                          # Easy problems
285|│   ├── medium/                        # Medium problems
286|│   └── hard/                          # Hard problems
287|│
288|├── 📂 deployments/                    # Deployment configs
289|│   ├── kubernetes/                    # K8s manifests
290|│   └── prometheus/                    # Prometheus config
291|│
292|├── 📂 tests/                          # Tests
293|│   ├── integration/                   # Integration tests
294|│   ├── harness/                       # Test harness tests
295|│   └── load/                          # Load tests
296|│
297|├── 📄 docker-compose.yml              # Docker Compose
298|├── 📄 Makefile                        # Build commands
299|├── 📄 go.mod                          # Go module
300|├── 📄 README.md                       # This file
301|└── 📄 proto/                          # gRPC proto files
302|```
303|
304|---
305|
306|## 🧪 Testing
307|
308|```bash
309|# Run all tests
310|make test
311|
312|# Run with coverage
313|make test-coverage
314|
315|# Run benchmarks
316|go test -bench=. ./benchmarks/
317|
318|# Run load tests (needs K6)
319|k6 run tests/load/k6-script.js
320|```
321|
322|---
323|
324|## 🚢 Deployment
325|
326|### Docker Compose (Development)
327|
328|```bash
329|# Start all services
330|make docker-up
331|
332|# Stop all services
333|make docker-down
334|
335|# View logs
336|make docker-logs
337|```
338|
339|### Kubernetes (Production)
340|
341|```bash
342|# Apply all manifests
343|kubectl apply -f deployments/kubernetes/
344|
345|# Check status
346|kubectl get pods -n coding-challenge
347|
348|# View logs
349|kubectl logs -f deployment/api-gateway -n coding-challenge
350|```
351|
352|Lihat [docs/HOW_TO_RUN.md](docs/HOW_TO_RUN.md) untuk detail lengkap.
353|
354|---
355|
356|## 🤝 Kontribusi
357|
358|Kami menerima kontribusi dari siapapun! Lihat [docs/CONTRIBUTING.md](docs/CONTRIBUTING.md) untuk panduan.
359|
360|1. Fork repository
361|2. Buat branch baru (`git checkout -b feature/amazing-feature`)
362|3. Commit changes (`git commit -m 'Add amazing feature'`)
363|4. Push ke branch (`git push origin feature/amazing-feature`)
364|5. Buat Pull Request
365|
366|---
367|
368|## 📝 License
369|
370|Distributed under the MIT License. See [LICENSE](LICENSE) for more.
371|
372|---
373|
374|## 📞 Kontak
375|
376|- **Email**: support@codingchallenge.com
377|- **Discord**: https://discord.gg/codingchallenge
378|- **GitHub**: https://github.com/codingchallenge/platform
379|
380|---
381|
382|> **Dibuat dengan ❤️ menggunakan Go, React, PostgreSQL, Redis, RabbitMQ, Docker, dan Kubernetes.**
383|