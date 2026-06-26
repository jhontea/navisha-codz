# 🚀 Coding Challenge Platform

> Platform coding challenge seperti HackerRank dengan compiler Go, real-time execution, progressive hints, leaderboard, dan badge system.

![Go](https://img.shields.io/badge/Go-1.25+-00ADD8?style=flat&logo=go)
![React](https://img.shields.io/badge/React-18+-61DAFB?style=flat&logo=react)
![PostgreSQL](https://img.shields.io/badge/PostgreSQL-15+-336791?style=flat&logo=postgresql)
![Redis](https://img.shields.io/badge/Redis-7+-DC382D?style=flat&logo=redis)
![RabbitMQ](https://img.shields.io/badge/RabbitMQ-3.12-FF6600?style=flat&logo=rabbitmq)
![Docker](https://img.shields.io/badge/Docker-24+-2496ED?style=flat&logo=docker)
![Kubernetes](https://img.shields.io/badge/Kubernetes-1.28+-326CE5?style=flat&logo=kubernetes)
![Tests](https://img.shields.io/badge/Tests-10_packages_passing-brightgreen)
![Problems](https://img.shields.io/badge/Problems-25-blue)
![License](https://img.shields.io/badge/License-MIT-green.svg)

---

## 📋 Daftar Isi

- [Fitur Utama](#fitur-utama)
- [Tech Stack](#tech-stack)
- [Quick Start](#quick-start)
- [Dokumentasi](#dokumentasi)
- [Arsitektur](#arsitektur)
- [API Reference](#api-reference)
- [Development](#development)
- [Deployment](#deployment)
- [Monitoring](#monitoring)
- [20 Loop Improvement](#20-loop-improvement)
- [Kontribusi](#kontribusi)
- [Lisensi](#lisensi)

---

## ✨ Fitur Utama

### Untuk User (Peserta)
- ✅ **25 Soal** — Algoritma, Data Structure, Dynamic Programming
- ✅ **3 Tingkat Kesulitan** — Easy, Medium, Hard
- ✅ **Tag Filter** — Filter soal berdasarkan tags (`?tags=hash-map,dp`)
- ✅ **Real-time Code Execution** — Compiler Go dengan feedback instan
- ✅ **Progressive Hint System** — 3 level hints per soal
- ✅ **Leaderboard** — Weekly, Monthly, All-time ranking dengan ELO
- ✅ **Badge & Achievement System** — 6 badges, 4 achievements
- ✅ **Code Editor** — Monaco Editor dengan 5 themes + auto-completion
- ✅ **DP Visualization** — Step-by-step animasi algoritma DP
- ✅ **Submission History** — Track semua submission
- ✅ **Profil & Statistik** — Rating, streak, solved problems

### Untuk Admin
- ✅ **Problem Management** — CRUD soal dengan test cases
- ✅ **User Management** — Ban/unban, role management
- ✅ **Admin Dashboard** — Statistik real-time, charts, activity feed
- ✅ **Log Viewer** — View dan filter logs
- ✅ **Swagger/OpenAPI** — Dokumentasi API otomatis

### Teknis
- ✅ **Microservices Architecture** — 8 services (API Gateway, Auth, Problem, Execution, Worker, Leaderboard, Hint, WebSocket)
- ✅ **Docker Sandbox** — Secure code execution with seccomp + AppArmor
- ✅ **WebSocket** — Real-time submission updates
- ✅ **Redis Caching** — Cache warming, cursor pagination, TTL 5 menit
- ✅ **RabbitMQ** — Async code execution queue + priority queue + DLQ
- ✅ **JWT Authentication** — Secure auth dengan refresh tokens + device fingerprint
- ✅ **Rate Limiting** — Sliding window, X-RateLimit headers
- ✅ **Graceful Shutdown** — Zero-downtime deployment
- ✅ **CORS Security** — Specific origin dari env var (bukan wildcard)
- ✅ **Prometheus + Grafana** — Monitoring metrics dan dashboards
- ✅ **Gzip Compression** — Kompresi JSON responses
- ✅ **Cursor-based Pagination** — Keyset pagination (bukan OFFSET)
- ✅ **OpenAPI/Swagger** — Dokumentasi API interaktif di `/swagger/index.html`

---

## 🛠 Tech Stack

| Layer | Technology |
|-------|-----------|
| **Backend** | Go 1.25+ (Gin framework) |
| **Frontend** | React 18+, TypeScript, Tailwind CSS |
| **Code Editor** | Monaco Editor |
| **Database** | PostgreSQL 15+ |
| **Cache** | Redis 7+ |
| **Message Queue** | RabbitMQ 3.12 |
| **Authentication** | JWT (access + refresh tokens) |
| **Container** | Docker 24+ (multi-stage builds) |
| **Orchestration** | Kubernetes 1.28+ |
| **Monitoring** | Prometheus + Grafana (auto-provisioned) |
| **CI/CD** | GitHub Actions |
| **IaC** | Terraform (AWS: VPC, EKS, RDS, ElastiCache, MQ) |
| **API Docs** | Swagger/OpenAPI 2.0 |

---

## 🚀 Quick Start

### Prasyarat
- Go 1.25+
- Docker & Docker Compose
- Node.js 18+

### Opsi 1: Docker Compose (Rekomendasi)

```bash
# Clone repository
git clone https://github.com/codingchallenge/platform.git
cd platform

# Setup environment
cp .env.example .env
# Edit .env — set JWT_ACCESS_SECRET dan JWT_REFRESH_SECRET

# Start semua services
docker-compose up -d --build

# Buka di browser
open http://localhost:9100
```

### Opsi 2: Manual Development

```bash
# 1. Install dependencies
go mod download

# 2. Setup infrastructure
docker-compose up -d postgres redis rabbitmq

# 3. Run migrations
make migrate

# 4. Export JWT secrets
export JWT_ACCESS_SECRET=***
export JWT_REFRESH_SECRET=***

# 5. Start services (di terminal terpisah)
go run services/auth-service/main.go      # :9101
go run services/problem-service/main.go    # :9102
go run services/execution-service/main.go  # :9103
go run services/leaderboard-service/main.go # :9104
go run services/hint-service/main.go       # :9105
go run services/execution-worker/main.go   # :9106
go run services/websocket-service/main.go  # :9107
go run services/api-gateway/main.go        # :9100

# 6. Start frontend (di terminal lain)
cd web/frontend
npm install
npm run dev

# 7. Buka http://localhost:5173
```

Selengkapnya lihat [docs/HOW_TO_RUN.md](docs/HOW_TO_RUN.md)

---

## 📖 Dokumentasi

| Dokumen | Deskripsi |
|---------|-----------|
| [docs/HOW_TO_RUN.md](docs/HOW_TO_RUN.md) | Panduan lengkap menjalankan aplikasi |
| [docs/HOW_TO_USE.md](docs/HOW_TO_USE.md) | Panduan menggunakan aplikasi (user & admin) |
| [docs/ARCHITECTURE_V2.md](docs/ARCHITECTURE_V2.md) | Diagram arsitektur dan penjelasan |
| [docs/DATABASE_SCHEMA.md](docs/DATABASE_SCHEMA.md) | Schema database dan query examples |
| [docs/API.md](docs/API.md) | API documentation |
| [docs/SUMMARY.md](docs/SUMMARY.md) | Ringkasan lengkap proyek |
| [docs/FINAL_REPORT.md](docs/FINAL_REPORT.md) | Laporan final build & test |
| [docs/PERFORMANCE_REPORT.md](docs/PERFORMANCE_REPORT.md) | Benchmark dan load test |
| [docs/MONITORING_SETUP.md](docs/MONITORING_SETUP.md) | Setup Prometheus + Grafana |
| [docs/CODE_REVIEW_FINDINGS.md](docs/CODE_REVIEW_FINDINGS.md) | Code review findings |
| [docs/SECURITY.md](docs/SECURITY.md) | Security policy |
| [docs/CONTRIBUTING.md](docs/CONTRIBUTING.md) | Panduan kontribusi |
| [docs/CHANGELOG.md](docs/CHANGELOG.md) | Riwayat perubahan |
| [docs/ROADMAP.md](docs/ROADMAP.md) | Rencana pengembangan |

---

## 🏗 Arsitektur

```
┌─────────────────────────────────────────────────────────────────┐
│                        Browser (Client)                          │
│  ┌──────────────┐  ┌──────────────┐  ┌───────────────────────┐ │
│  │ Problem List │  │ Code Editor  │  │ Test Results + Hints │ │
│  │   (React)    │  │  (Monaco)    │  │   (Real-time WS)     │ │
│  └──────────────┘  └──────────────┘  └───────────────────────┘ │
└──────────────────────────┬──────────────────────────────────────┘
                           │ HTTP / WebSocket
┌──────────────────────────▼──────────────────────────────────────┐
│                    API Gateway (Go/Gin)                          │
│  • Rate Limiting • JWT Validation • Gzip • Timeout • CORS      │
└──────┬────────┬────────┬────────┬────────┬────────┬─────────────┘
       │        │        │        │        │        │
┌──────▼──┐ ┌───▼────┐ ┌─▼──────┐ ┌─▼────┐ ┌─▼────┐ ┌─────────▼──┐
│  Auth   │ │Problem │ │Executi-│ │Leader│ │Hint  │ │  WebSocket  │
│ Service │ │Service │ │  on    │ │board │ │Servi-│ │   Service   │
│  :9101  │ │ :9102  │ │Service │ │:9104 │ │ce    │ │    :9107    │
└────┬────┘ └───┬────┐ │ :9103  │ └──┬───┘ │:9105 │ └─────────────┘
     │         │    │ └──┬─────┘    │     └──────┘
     │         │    │    │          │
     └────┬────┴────┴────┴────┬─────┘
          │                   │
┌─────────▼───────────────────▼────────────────────────────────────┐
│                    Data Layer                                     │
│  ┌──────────────┐  ┌──────────────┐  ┌──────────────┐          │
│  │  PostgreSQL  │  │    Redis     │  │   RabbitMQ   │          │
│  │  (Primary)   │  │   (Cache)    │  │  (Job Queue) │          │
│  └──────────────┘  └──────────────┘  └──────────────┘          │
└──────────────────────────────────────────────────────────────────┘
                           │
┌──────────────────────────▼──────────────────────────────────────┐
│               Code Execution (Docker Sandbox)                    │
│  • --network=none • --read-only • seccomp • AppArmor           │
│  • Memory: 256MB-1GB • CPU: 1 • Timeout: 1-5s                 │
└─────────────────────────────────────────────────────────────────┘
```

---

## 🔌 API Reference

### Base URL
```
Development: http://localhost:9100/api
Production:  https://api.codingchallenge.com/api
```

### Endpoints

#### Auth
| Method | Endpoint | Deskripsi | Auth |
|--------|----------|-----------|------|
| POST | `/auth/register` | Registrasi | ❌ |
| POST | `/auth/login` | Login | ❌ |
| POST | `/auth/refresh` | Refresh token | ❌ |
| POST | `/auth/logout` | Logout | ✅ |

#### Problems
| Method | Endpoint | Deskripsi | Auth |
|--------|----------|-----------|------|
| GET | `/api/problems` | List soal (filter: difficulty, category, tags) | ❌ |
| GET | `/api/problems/:id` | Detail soal | ❌ |
| GET | `/api/problems/:id/template` | Get template code | ❌ |
| POST | `/api/problems` | Buat soal (Admin) | ✅ Admin |

#### Submissions
| Method | Endpoint | Deskripsi | Auth |
|--------|----------|-----------|------|
| POST | `/api/submissions` | Submit kode | ✅ |
| GET | `/api/submissions/:id` | Get submission status | ✅ |
| GET | `/api/submissions/user/:userId` | Riwayat submission | ✅ |
| POST | `/api/validate` | Validasi syntax | ✅ |

#### Leaderboard
| Method | Endpoint | Deskripsi | Auth |
|--------|----------|-----------|------|
| GET | `/api/leaderboard/weekly` | Peringkat mingguan | ❌ |
| GET | `/api/leaderboard/monthly` | Peringkat bulanan | ❌ |
| GET | `/api/leaderboard/all-time` | Peringkat keseluruhan | ❌ |

#### Hints
| Method | Endpoint | Deskripsi | Auth |
|--------|----------|-----------|------|
| GET | `/api/problems/:id/hints` | Get hints | ✅ |
| POST | `/api/problems/:id/hints/:hintId/use` | Gunakan hint | ✅ |

#### WebSocket
| Endpoint | Deskripsi |
|----------|-----------|
| `ws://localhost:9100/ws` | Real-time submission updates |

#### Swagger
| Endpoint | Deskripsi |
|----------|-----------|
| `GET /swagger/index.html` | Dokumentasi API interaktif |

---

## 🗂 Struktur Project

```
coding-challange/
├── 📂 docs/                           # Dokumentasi (15+ files)
├── 📂 services/                       # 8 Microservices
│   ├── api-gateway/                   # API Gateway (Go + Gin)
│   ├── auth-service/                  # Authentication service
│   ├── problem-service/               # Problem management
│   ├── execution-service/             # Code execution orchestrator
│   ├── execution-worker/              # Code execution worker
│   ├── websocket-service/             # WebSocket real-time
│   ├── leaderboard-service/           # Ranking & scoring ELO
│   └── hint-service/                  # Hint management
├── 📂 pkg/                            # Shared packages (10 pkg)
├── 📂 internal/                       # Internal packages (7 pkg)
├── 📂 web/frontend/                   # React SPA (40+ components)
├── 📂 problems/                       # 25 Problem bank (YAML)
├── 📂 deployments/                    # K8s, Terraform, Prometheus
├── 📂 tests/                          # Integration + load tests
├── 📂 benchmarks/                     # Performance benchmarks
├── 📂 proto/                          # gRPC proto files
├── 📂 .github/workflows/              # CI/CD pipelines
├── 📄 docker-compose.yml              # Docker Compose
└── 📄 Makefile                        # Build commands
```

---

## 🧪 Testing

```bash
# Run all tests
make test

# Run with coverage
make test-coverage

# Run benchmarks
go test -bench=. ./benchmarks/ -benchmem

# Run load tests (needs K6)
k6 run tests/load/k6-script.js

# Run specific package
go test ./internal/handler/... -v
```

### Test Status (10 packages)
- ✅ `internal/handler` — HTTP handlers
- ✅ `internal/repository` — Data access
- ✅ `internal/service` — Business logic
- ✅ `pkg/logger` — Logging
- ✅ `pkg/middleware` — Auth, CORS, rate limit
- ✅ `pkg/rabbitmq` — Message queue
- ✅ `pkg/redis` — Caching
- ✅ `pkg/security` — Security utilities
- ✅ `services/execution-worker` — Code execution
- ✅ `tests/integration` — Integration tests

---

## 📊 20 Loop Improvement

| Loop | Fokus | Hasil |
|------|-------|-------|
| 1 | Foundation | Architecture docs, fix tests, 0 TS errors |
| 2-3 | Core Fixes | Hapus monolith, security, konsolidasi |
| 4 | Problems + UI | 25 problems, tag filter, search, sort |
| 5-6 | Perf + Security | Cursor pagination, Redis cache, Gzip, rate limit headers |
| 7-8 | QA + Docs | Testing, SUMMARY.md, .env.example |
| 9-10 | UI + Final | Skeleton loading, mobile tabs, final report |
| 11 | Code Review | 7 remaining findings fixed, CORS hardened |
| 12 | Test Coverage | +3 test packages (security, rabbitmq, worker) |
| 13 | More Problems | 20 → 25 problems |
| 14 | Frontend Tests | Vitest setup, component tests |
| 15 | Swagger | OpenAPI annotations, Swagger UI |
| 16 | Load Test | K6 scripts, benchmark, performance report |
| 17 | Docker | Multi-stage builds, .dockerignore |
| 18 | Monitoring | Grafana dashboards, Prometheus, setup guide |
| 19 | Hardening | Nil pointer checks, recover middleware, edge cases |
| 20 | Final | All checks PASS, production ready |

---

## 🚢 Deployment

### Docker Compose (Development)
```bash
make docker-up    # Start all services
make docker-down  # Stop all services
make docker-logs  # View logs
```

### Kubernetes (Production)
```bash
kubectl apply -f deployments/kubernetes/
kubectl get pods -n coding-challenge
```

### Monitoring
```bash
# Start with monitoring stack
docker-compose --profile monitoring up -d

# Access Grafana: http://localhost:3000 (admin/admin)
# Access Prometheus: http://localhost:9090
```

Lihat [docs/HOW_TO_RUN.md](docs/HOW_TO_RUN.md) untuk detail lengkap.

---

## 🤝 Kontribusi

Kami menerima kontribusi dari siapapun! Lihat [docs/CONTRIBUTING.md](docs/CONTRIBUTING.md) untuk panduan.

1. Fork repository
2. Buat branch baru (`git checkout -b feature/amazing-feature`)
3. Commit changes (`git commit -m 'feat: add amazing feature'`)
4. Push ke branch (`git push origin feature/amazing-feature`)
5. Buat Pull Request

---

## 📝 License

Distributed under the MIT License. See [LICENSE](LICENSE) for more.

---

## 📞 Kontak

- **Email**: support@codingchallenge.com
- **Discord**: https://discord.gg/codingchallenge
- **GitHub**: https://github.com/codingchallenge/platform
- **Swagger Docs**: http://localhost:9100/swagger/index.html

---

> **Dibuat dengan ❤️ menggunakan Go, React, PostgreSQL, Redis, RabbitMQ, Docker, dan Kubernetes.**
> **20 Loops Improvement — Build ✅ Tests ✅ 25 Problems ✅**
