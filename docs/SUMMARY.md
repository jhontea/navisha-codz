# 📋 Coding Challenge Platform — Ringkasan Proyek

> **Versi**: 1.0.0 | **Status**: ✅ Rilis | **Loop**: 7 (Testing & Quality)

---

## 📑 Daftar Isi

1. [Tentang Proyek](#tentang-proyek)
2. [Fitur Lengkap](#fitur-lengkap)
3. [Tech Stack](#tech-stack)
4. [Struktur Proyek](#struktur-proyek)
5. [Cara Menjalankan](#cara-menjalankan)
6. [API Endpoints](#api-endpoints)
7. [Environment Variables](#environment-variables)
8. [Testing](#testing)
9. [Deployment](#deployment)
10. [Troubleshooting](#troubleshooting)
11. [Dokumentasi Terkait](#dokumentasi-terkait)

---

## 1. Tentang Proyek

Platform coding challenge *self-hosted* seperti HackerRank, dibangun dengan **Go microservices** dan **React frontend**. Mendukung real-time code execution, progressive hints system, leaderboard ELO ranking, dan badge/achievement system.

**Target**: 1,000 active users, 500 submissions/day
**Arsitektur**: Microservices (8 services) + Docker Sandbox + Kubernetes

---

## 2. Fitur Lengkap

### 🔐 Authentication & Authorization
- Registrasi & Login (email/password)
- JWT Access Token (15 menit) + Refresh Token (7 hari)
- Token rotation dengan reuse detection
- Device fingerprinting
- Rate limiting per endpoint
- Admin role management

### 💻 Code Editor
- **Monaco Editor** — Sama dengan VS Code
- Go syntax highlighting, auto-completion
- 5 themes (VS Dark, VS Light, High Contrast, Monokai, GitHub)
- Snippets untuk template code
- Code folding, minimap

### 📝 Problem Bank
- **15 soal** — 5 Easy / 5 Medium / 5 Hard
- YAML-based problem definitions
- 3-level progressive hints (dengan score penalty)
- Tag filter (`?tags=hash-map,dp`)
- Difficulty & category filter
- Template code per problem

### ⚡ Code Execution
- **Docker Sandbox** — Isolated execution
- `--network=none`, `--read-only`, `--cap-drop=ALL`
- Seccomp profile, memory limits, CPU limits
- RabbitMQ async queue dengan priority per difficulty
- Execution Worker dengan HPA scaling (3-50 replicas)
- DP Visualization (Fibonacci, Knapsack, LCS, Edit Distance, Coin Change, LIS)

### 🏆 Leaderboard & Rating
- **ELO Rating System**
- Weekly / Monthly / All-time leaderboard
- Badge system (6 badges): Rising Star, Code Master, Bug Squasher, Speed Demon, Problem Solver, Top Contributor
- Achievement system (4 achievements)
- Score history & rating chart

### 🎨 Frontend
- React 18 + TypeScript + Tailwind CSS
- Responsive design (mobile-first)
- Dark/Light theme (system preference detection)
- PWA support (service worker, offline cache, install prompt)
- Real-time submission updates via WebSocket

### 🔧 Admin Features
- Problem CRUD management
- User management (ban/unban, roles)
- Monitoring dashboard
- Log viewer & filter

---

## 3. Tech Stack

### Backend
| Layer | Teknologi | Versi |
|-------|-----------|-------|
| Language | Go | 1.25+ |
| HTTP Framework | Gin | v1.9.1 |
| JWT | golang-jwt | v5.3.1 |
| UUID | google/uuid | v1.6.0 |
| WebSocket | gorilla/websocket | v1.5.3 |
| PostgreSQL Driver | jackc/pgx | v5.10.0 |
| Redis | go-redis | v9.21.0 |
| RabbitMQ | amqp091-go | v1.12.0 |

### Frontend
| Layer | Teknologi |
|-------|-----------|
| Framework | React 18 |
| Language | TypeScript |
| Styling | Tailwind CSS |
| Code Editor | Monaco Editor |
| State | React Context + Hooks |
| Build | Vite |
| HTTP | Axios / Fetch |

### Infrastructure
| Layer | Teknologi |
|-------|-----------|
| Container | Docker 24+ |
| Orchestration | Kubernetes 1.28+ |
| Message Queue | RabbitMQ 3.12 |
| Cache | Redis 7+ |
| Database | PostgreSQL 15+ |
| Monitoring | Prometheus + Grafana |
| Logging | ELK Stack (Fluentd, Elasticsearch, Kibana) |
| Tracing | Jaeger |
| CI/CD | GitHub Actions |

---

## 4. Struktur Proyek

```
coding-challange/
├── docs/                          # Dokumentasi (11 dokumen)
│   ├── ARCHITECTURE_V2.md         # Arsitektur microservices
│   ├── SUMMARY.md                 # Ringkasan proyek (ini)
│   ├── HOW_TO_RUN.md              # Panduan menjalankan
│   ├── HOW_TO_USE.md              # Panduan penggunaan
│   ├── API.md                     # Dokumentasi API
│   ├── CHANGELOG.md               # Riwayat perubahan
│   ├── ROADMAP.md                 # Roadmap pengembangan
│   ├── SECURITY.md                # Kebijakan keamanan
│   ├── DEPLOYMENT.md              # Panduan deployment
│   ├── CONTRIBUTING.md            # Panduan kontribusi
│   ├── LOOP7_REPORT.md            # Laporan testing Loop 7
│   └── CODE_REVIEW_FINDINGS.md    # Temuan code review
│
├── services/                      # Microservices (8 services)
│   ├── api-gateway/               # Gateway utama (port 9100)
│   ├── auth-service/              # Auth & JWT (port 9101)
│   ├── problem-service/           # Problem CRUD (port 9102)
│   ├── execution-service/         # Execution orchestrator (port 9103)
│   ├── leaderboard-service/       # ELO ranking (port 9104)
│   ├── hint-service/              # Progressive hints (port 9105)
│   ├── execution-worker/          # Sandbox worker (port 9106)
│   └── websocket-service/         # Real-time push (port 9107)
│
├── pkg/                           # Shared packages
│   ├── database/                  # PostgreSQL connection pool
│   ├── redis/                     # Redis client + caching
│   ├── rabbitmq/                  # RabbitMQ producer/consumer
│   ├── middleware/                # Auth, CORS, rate limiting, compression
│   ├── websocket/                 # WebSocket hub
│   ├── logger/                    # Structured logging (JSON)
│   ├── config/                    # Configuration management
│   ├── security/                  # Security utilities
│   ├── errors/                    # Error handling & types
│   └── health/                    # Health check endpoints
│
├── internal/                      # Legacy internal packages
│   ├── handler/                   # HTTP handlers
│   ├── service/                   # Business logic
│   ├── repository/                # Data access layer
│   ├── model/                     # Data models
│   ├── middleware/                # Internal middleware
│   └── config/                    # App configuration
│
├── problems/                      # Problem bank (YAML)
│   ├── easy/                      # 5 problems
│   ├── medium/                    # 5 problems
│   └── hard/                      # 5 problems
│
├── web/frontend/                  # React SPA
│   ├── src/
│   │   ├── components/            # UI components
│   │   ├── pages/                 # Page routes
│   │   ├── hooks/                 # Custom React hooks
│   │   ├── services/              # API client
│   │   ├── store/                 # State management
│   │   ├── types/                 # TypeScript types
│   │   └── styles/                # CSS & Tailwind
│   ├── public/                    # Static assets
│   └── vite.config.ts             # Vite configuration
│
├── deployments/                   # Deployment configs
│   ├── kubernetes/                # K8s manifests
│   └── prometheus/                # Prometheus config
│
├── tests/                         # Test suites
│   ├── integration/               # Integration tests
│   ├── harness/                   # Test harness tests
│   └── load/                      # Load tests (K6)
│
├── runner/                        # Sandbox runner image
└── proto/                         # gRPC proto files
```

---

## 5. Cara Menjalankan

### Prasyarat
- Go 1.21+
- Docker Desktop 24+ (dengan WSL2 di Windows)
- Node.js 18+ (untuk frontend development)

### Opsi 1: Docker Compose (Rekomendasi)

```bash
# 1. Clone & masuk direktori
cd C:\Users\PC\go\src\project\coding-challange

# 2. Setup environment
cp .env.example .env
# Edit .env: isi JWT_ACCESS_SECRET & JWT_REFRESH_SECRET

# 3. Build & start semua services
docker-compose up -d --build

# 4. Setup database
make migrate

# 5. Buka browser
start http://localhost:9100
```

### Opsi 2: Manual (Development)

```bash
# Terminal 1: Infrastructure
docker-compose up -d postgres redis rabbitmq

# Terminal 2-9: Jalankan setiap service
go run services/api-gateway/main.go         # :9100
go run services/auth-service/main.go        # :9101
go run services/problem-service/main.go     # :9102
go run services/execution-service/main.go   # :9103
go run services/leaderboard-service/main.go # :9104
go run services/hint-service/main.go        # :9105
go run services/execution-worker/main.go    # :9106
go run services/websocket-service/main.go   # :9107

# Terminal 10: Frontend
cd web/frontend
npm install
npm run dev  # :5173
```

### Port Mapping

| Service | Port | URL |
|---------|------|-----|
| API Gateway | 9100 | http://localhost:9100 |
| Auth Service | 9101 | http://localhost:9101 |
| Problem Service | 9102 | http://localhost:9102 |
| Execution Service | 9103 | http://localhost:9103 |
| Leaderboard Service | 9104 | http://localhost:9104 |
| Hint Service | 9105 | http://localhost:9105 |
| Execution Worker | 9106 | http://localhost:9106 |
| WebSocket Service | 9107 | http://localhost:9107 |
| Frontend (Dev) | 5173 | http://localhost:5173 |
| PostgreSQL | 5432 | localhost:5432 |
| Redis | 6379 | localhost:6379 |
| RabbitMQ | 5672 / 15672 | localhost:15672 (management UI) |

---

## 6. API Endpoints

### Public Endpoints (No Auth)

| Method | Endpoint | Deskripsi |
|--------|----------|-----------|
| GET | `/health` | Health check server |
| GET | `/api/problems` | List semua problem |
| GET | `/api/problems/:slug` | Detail problem |
| GET | `/api/problems/:slug/template` | Template code |
| POST | `/auth/login` | Login user |
| POST | `/auth/register` | Register user |
| POST | `/auth/refresh` | Refresh JWT token |
| POST | `/auth/logout` | Logout (invalidate token) |

### Protected Endpoints (Auth Required)

| Method | Endpoint | Deskripsi |
|--------|----------|-----------|
| POST | `/api/submissions` | Submit code execution |
| GET | `/api/submissions/:id` | Detail submission |
| GET | `/api/leaderboard/weekly` | Peringkat mingguan |
| GET | `/api/leaderboard/monthly` | Peringkat bulanan |
| GET | `/api/leaderboard/all-time` | Peringkat sepanjang masa |
| GET | `/api/problems/:id/hints` | Get unlocked hints |
| POST | `/api/problems/:id/hints/:hintId/use` | Unlock hint |
| GET | `/api/user/profile` | User profile |
| GET | `/api/user/stats` | User statistics |
| GET | `/api/user/submissions` | Riwayat submission |
| GET | `/ws/submissions/:id` | WebSocket real-time status |

### Filter Parameters untuk Problem List

| Parameter | Tipe | Contoh | Deskripsi |
|-----------|------|--------|-----------|
| `difficulty` | string | `easy`, `medium`, `hard` | Filter tingkat kesulitan |
| `category` | string | `array`, `string`, `dp` | Filter kategori |
| `tags` | string | `hash-map,dp` | Filter tag (AND logic, comma-separated) |
| `type` | string | `function`, `main` | Filter tipe problem |

### Response Format

```json
{
  "data": { ... },
  "meta": {
    "request_id": "abc-123",
    "timestamp": "2026-06-26T10:30:00Z"
  }
}
```

### Error Response

```json
{
  "error": {
    "message": "Human-readable message",
    "code": 400,
    "type": "validation_error",
    "details": [{ "field": "code", "issue": "required" }]
  },
  "meta": { "request_id": "abc-123", "timestamp": "..." }
}
```

### HTTP Status Codes

| Code | Meaning |
|------|---------|
| 200 | Success |
| 400 | Bad Request / Validation Error |
| 401 | Unauthorized (invalid/expired token) |
| 404 | Not Found |
| 429 | Rate Limit Exceeded |
| 500 | Internal Server Error |
| 502 | Bad Gateway (Docker sandbox error) |
| 503 | Service Unavailable |

---

## 7. Environment Variables

### Wajib Diisi

```env
# JWT — HARUS di-generate sendiri! Jangan pake default!
JWT_ACCESS_SECRET=<generate-random-64-char-string>
JWT_REFRESH_SECRET=<generate-random-64-char-string>
```

### Semua Variable

| Variable | Default | Deskripsi |
|----------|---------|-----------|
| `PORT` | `8080` | Server port |
| `LOG_LEVEL` | `debug` | debug, info, warn, error |
| **Database** | | |
| `DB_HOST` | `localhost` | PostgreSQL host |
| `DB_PORT` | `5432` | PostgreSQL port |
| `DB_USER` | `postgres` | Database user |
| `DB_PASSWORD` | `postgres` | Database password |
| `DB_NAME` | `coding_challange` | Database name |
| `DB_MAX_CONNS` | `25` | Max pool connections |
| `DB_IDLE_CONNS` | `5` | Idle connections |
| **Redis** | | |
| `REDIS_ADDR` | `localhost:6379` | Redis address |
| `REDIS_PASSWORD` | (empty) | Redis password |
| `REDIS_DB` | `0` | Redis DB number |
| `REDIS_POOL_SIZE` | `10` | Connection pool size |
| **RabbitMQ** | | |
| `RABBITMQ_HOST` | `localhost` | RabbitMQ host |
| `RABBITMQ_PORT` | `5672` | RabbitMQ port |
| `RABBITMQ_USER` | `guest` | AMQP user |
| `RABBITMQ_PASSWORD` | `guest` | AMQP password |
| **JWT** | | |
| `JWT_ACCESS_SECRET` | **(required)** | Access token signing key |
| `JWT_REFRESH_SECRET` | **(required)** | Refresh token signing key |
| `JWT_ACCESS_TTL_MIN` | `15` | Access token expiry (menit) |
| `JWT_REFRESH_TTL_HOURS` | `168` | Refresh token expiry (jam = 7 hari) |
| **Sandbox** | | |
| `SANDBOX_IMAGE` | `golang:1.21-alpine` | Docker sandbox image |
| `SANDBOX_MEMORY_MB` | `256` | Memory limit |
| `SANDBOX_TIME_SEC` | `10` | Execution timeout |
| `DISABLE_LOCAL_FALLBACK` | `false` | Force Docker sandbox |
| `MAX_OUTPUT_SIZE_MB` | `1` | Max output size |
| `MAX_CODE_SIZE_KB` | `64` | Max code size |
| **Rate Limit** | | |
| `RATE_LIMIT_REQUESTS` | `10` | Requests per window |
| `RATE_LIMIT_WINDOW_SECONDS` | `60` | Window duration |
| `SUBMIT_RATE_LIMIT` | `10` | Submit limit per window |
| `SUBMIT_RATE_WINDOW_SECONDS` | `60` | Submit window duration |

---

## 8. Testing

### Jalankan Semua Test

```bash
# All tests
go test ./... -count=1 -v

# Dengan coverage
go test ./... -coverprofile=coverage.out
go tool cover -html=coverage.out -o coverage.html

# Benchmark
go test -bench=. ./benchmarks/

# Load test (butuh K6)
k6 run tests/load/k6-script.js
```

### Package Tests

| Package | Jumlah Test | Status |
|---------|-------------|--------|
| `internal/handler` | 8 | ✅ PASS |
| `internal/repository` | 5 | ✅ PASS |
| `internal/service` | 6 | ✅ PASS |
| `pkg/logger` | 3 | ✅ PASS |
| `pkg/middleware` | 6 | ✅ PASS |
| `pkg/redis` | 3 | ✅ PASS |
| `tests/integration` | 6 | ✅ PASS |
| **Total** | **37** | **✅ ALL PASS** |

### Makefile Commands

```bash
make test           # Run all tests
make test-coverage  # Tests + coverage report
make lint           # golangci-lint
make build          # Build all service binaries
make docker-up      # Docker Compose start
make docker-down    # Docker Compose stop
make health         # Check all service health
make migrate        # Run DB migrations
```

---

## 9. Deployment

### Docker Compose (Development)

```bash
make docker-up      # Start semua service
make docker-down    # Stop semua
make docker-logs    # View logs
docker-compose up -d --scale execution-worker=3   # Scale workers
```

### Kubernetes (Production)

```bash
kubectl apply -f deployments/kubernetes/
kubectl get pods -n coding-challenge
kubectl logs -f deployment/api-gateway -n coding-challenge
```

### Estimated Infrastructure Cost: ~$1,405/month

| Component | Spec | Cost |
|-----------|------|------|
| API Servers (3x) | 2 vCPU, 4GB | $150 |
| Execution Workers (5x) | 4 vCPU, 8GB | $400 |
| PostgreSQL (Primary) | 4 vCPU, 16GB, 100GB SSD | $200 |
| Read Replica | 2 vCPU, 8GB | $100 |
| Redis Cluster (3 nodes) | 2 vCPU, 4GB | $150 |
| RabbitMQ (3 nodes) | 2 vCPU, 4GB | $120 |
| Monitoring | Prometheus + Grafana | $100 |
| **Total** | | **~$1,405/month** |

---

## 10. Troubleshooting

### Docker Issues

**Port already allocated**
```bash
netstat -ano | findstr :9100
taskkill /F /PID <PID>
```

**Docker Compose fails**
```bash
docker-compose down -v
docker system prune -a
docker-compose up -d --build
```

**Container can't connect to other services**
```bash
docker network ls
docker network inspect coding-challange_coding-challange
docker-compose restart
```

### Build Issues

**go build error**
```bash
go clean -cache
go mod download
go mod tidy
go build ./...
```

### Database Issues

**Migration fails**
```bash
docker-compose down -v
docker-compose up -d postgres
make migrate
```

### Authentication Issues

**Token expired**
```bash
# Login ulang untuk dapat token baru
curl -X POST http://localhost:9100/auth/login \
  -H "Content-Type: application/json" \
  -d '{"email":"user@example.com","password":"yourpassword"}'
```

**JWT_SECRET not set**
```bash
# Pastikan .env memiliki:
JWT_ACCESS_SECRET=<minimal-32-karakter-random>
JWT_REFRESH_SECRET=<minimal-32-karakter-random>
```

### Test Issues

**Integration test timeout**
```bash
# Pastikan infrastructure berjalan
docker-compose up -d postgres redis rabbitmq
go test ./tests/integration/... -v -timeout 120s
```

---

## 11. Dokumentasi Terkait

| Dokumen | Deskripsi |
|---------|-----------|
| [ARCHITECTURE_V2.md](ARCHITECTURE_V2.md) | Arsitektur microservices lengkap dengan diagram |
| [HOW_TO_RUN.md](HOW_TO_RUN.md) | Panduan detail menjalankan aplikasi |
| [HOW_TO_USE.md](HOW_TO_USE.md) | Panduan penggunaan untuk user & admin |
| [API.md](API.md) | Dokumentasi REST API lengkap |
| [CHANGELOG.md](CHANGELOG.md) | Riwayat perubahan versi |
| [ROADMAP.md](ROADMAP.md) | Roadmap pengembangan (Q3 2026 - Q2 2027) |
| [SECURITY.md](SECURITY.md) | Kebijakan dan fitur keamanan |
| [DEPLOYMENT.md](DEPLOYMENT.md) | Panduan deployment ke production |
| [CONTRIBUTING.md](CONTRIBUTING.md) | Panduan kontribusi open source |
| [LOOP7_REPORT.md](LOOP7_REPORT.md) | Laporan testing & quality (37 tests PASS) |
| [CODE_REVIEW_FINDINGS.md](CODE_REVIEW_FINDINGS.md) | Temuan code review & perbaikannya |

---

> **Dibuat dengan ❤️ menggunakan Go, React, PostgreSQL, Redis, RabbitMQ, Docker & Kubernetes**
>
> *Last updated: June 2026 | Coding Challenge Platform v1.0.0*
