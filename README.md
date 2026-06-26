# 🚀 Coding Challenge Platform

> Platform coding challenge seperti HackerRank dengan compiler Go, real-time execution, progressive hints, dan leaderboard.

![Go](https://img.shields.io/badge/Go-1.25+-00ADD8?style=flat&logo=go)
![React](https://img.shields.io/badge/React-18+-61DAFB?style=flat&logo=react)
![PostgreSQL](https://img.shields.io/badge/PostgreSQL-15+-336791?style=flat&logo=postgresql)
![Redis](https://img.shields.io/badge/Redis-7+-DC382D?style=flat&logo=redis)
![RabbitMQ](https://img.shields.io/badge/RabbitMQ-3.12-FF6600?style=flat&logo=rabbitmq)
![Docker](https://img.shields.io/badge/Docker-24+-2496ED?style=flat&logo=docker)
![Kubernetes](https://img.shields.io/badge/Kubernetes-1.28+-326CE5?style=flat&logo=kubernetes)
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
- [Kontribusi](#kontribusi)
- [Lisensi](#lisensi)

---

## ✨ Fitur Utama

### Untuk User (Peserta)
- ✅ **15+ Soal** — Algoritma, Data Structure, Dynamic Programming
- ✅ **3 Tingkat Kesulitan** — Easy, Medium, Hard
- ✅ **Real-time Code Execution** — Compiler Go dengan feedback instan
- ✅ **Progressive Hint System** — 3 level hints per soal
- ✅ **Leaderboard** — Weekly, Monthly, All-time ranking
- ✅ **Code Editor** — Monaco Editor dengan syntax highlighting
- ✅ **Submission History** — Track semua submission
- ✅ **Profil & Statistik** — Rating, streak, solved problems

### Untuk Admin
- ✅ **Problem Management** — CRUD soal dengan test cases
- ✅ **User Management** — Ban/unban, role management
- ✅ **Monitoring** — Dashboard dengan statistik real-time
- ✅ **Log Viewer** — View dan filter logs

### Teknis
- ✅ **Microservices Architecture** — Scalable, maintainable
- ✅ **Docker Sandbox** — Secure code execution
- ✅ **WebSocket** — Real-time submission updates
- ✅ **Redis Caching** — Fast problem list loading
- ✅ **RabbitMQ** — Async code execution queue
- ✅ **JWT Authentication** — Secure auth dengan refresh tokens
- ✅ **Rate Limiting** — Prevent abuse
- ✅ **Graceful Shutdown** — Zero-downtime deployment

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
| **Container** | Docker 24+ |
| **Orchestration** | Kubernetes 1.28+ |
| **Monitoring** | Prometheus + Grafana |
| **CI/CD** | GitHub Actions |

---

## 🚀 Quick Start

### Opsi 1: Docker Compose (Rekomendasi)

```bash
# Clone repository
git clone https://github.com/codingchallenge/platform.git
cd platform

# Start semua services
docker-compose up -d --build

# Buka di browser
open http://localhost:9100
```

Selengkapnya lihat [docs/HOW_TO_RUN.md](docs/HOW_TO_RUN.md)

### Opsi 2: Manual Development

```bash
# 1. Install dependencies
go mod download

# 2. Setup infrastructure (PostgreSQL, Redis, RabbitMQ)
docker-compose up -d postgres redis rabbitmq

# 3. Run migrations
make migrate

# 4. Start services (di terminal terpisah)
go run services/auth-service/main.go
go run services/problem-service/main.go
go run services/execution-service/main.go
go run services/leaderboard-service/main.go
go run services/hint-service/main.go

# 5. Start frontend (di terminal lain)
cd web/frontend
npm install
npm run dev

# 6. Buka http://localhost:5173
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
| [docs/TEST_PLAN.md](docs/TEST_PLAN.md) | Test plan dan test cases |
| [docs/CODE_REVIEW.md](docs/CODE_REVIEW.md) | Code review findings |
| [docs/API.md](docs/API.md) | API documentation |

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
│                       API Gateway (Nginx)                        │
│  • Rate Limiting • JWT Validation • Load Balancing • CORS       │
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
│                  Code Execution (Docker Sandbox)                  │
│  • --network=none • --read-only • --memory=256m • --cpus=1     │
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
| Method | Endpoint | Deskripsi |
|--------|----------|-----------|
| POST | `/auth/register` | Registrasi |
| POST | `/auth/login` | Login |
| POST | `/auth/refresh` | Refresh token |

#### Problems
| Method | Endpoint | Deskripsi |
|--------|----------|-----------|
| GET | `/api/problems` | List soal |
| GET | `/api/problems/:id` | Detail soal |
| POST | `/api/problems` | Buat soal (Admin) |

#### Submissions
| Method | Endpoint | Deskripsi |
|--------|----------|-----------|
| POST | `/api/submissions` | Submit kode |
| GET | `/api/submissions/:id` | Get status |

#### Leaderboard
| Method | Endpoint | Deskripsi |
|--------|----------|-----------|
| GET | `/api/leaderboard/weekly` | Peringkat mingguan |
| GET | `/api/leaderboard/monthly` | Peringatan bulanan |
| GET | `/api/leaderboard/all-time` | Peringkat keseluruhan |

#### Hints
| Method | Endpoint | Deskripsi |
|--------|----------|-----------|
| GET | `/api/problems/:id/hints` | Get hints |
| POST | `/api/problems/:id/hints/:hintId/use` | Gunakan hint |

---

## 🗂 Struktur Project

```
coding-challange/
├── 📂 docs/                           # Dokumentasi
│   ├── HOW_TO_RUN.md                  # Panduan menjalankan
│   ├── HOW_TO_USE.md                  # Panduan menggunakan
│   ├── ARCHITECTURE_V2.md             # Arsitektur sistem
│   ├── DATABASE_SCHEMA.md             # Database schema
│   ├── API.md                         # API documentation
│   ├── TEST_PLAN.md                   # Test plan
│   └── CODE_REVIEW.md                 # Code review
│
├── 📂 services/                       # Microservices
│   ├── api-gateway/                   # API Gateway (Nginx)
│   ├── auth-service/                  # Authentication service
│   ├── problem-service/               # Problem management
│   ├── execution-service/             # Code execution orchestrator
│   ├── execution-worker/              # Code execution worker
│   ├── websocket-service/             # WebSocket real-time
│   ├── leaderboard-service/           # Ranking & scoring
│   └── hint-service/                  # Hint management
│
├── 📂 pkg/                            # Shared packages
│   ├── database/                      # PostgreSQL connection
│   ├── redis/                         # Redis client
│   ├── rabbitmq/                      # RabbitMQ client
│   ├── middleware/                    # Auth, CORS, rate limiting
│   ├── websocket/                     # WebSocket hub
│   ├── logger/                        # Structured logging
│   ├── config/                        # Configuration
│   ├── security/                      # Security utilities
│   ├── errors/                        # Error handling
│   └── health/                        # Health checks
│
├── 📂 internal/                       # Internal packages
│   ├── handler/                       # HTTP handlers (legacy)
│   ├── service/                       # Business logic (legacy)
│   ├── repository/                    # Data access layer
│   ├── model/                         # Data models
│   └── config/                        # App config
│
├── 📂 web/                            # Frontend
│   └── frontend/                      # React SPA
│       ├── src/
│       │   ├── components/            # UI components
│       │   ├── pages/                 # Page components
│       │   ├── hooks/                 # Custom hooks
│       │   ├── services/              # API client
│       │   ├── store/                 # State management
│       │   ├── types/                 # TypeScript types
│       │   └── styles/                # CSS/Tailwind
│       ├── public/                    # Static assets
│       └── package.json
│
├── 📂 problems/                       # Problem bank (YAML)
│   ├── easy/                          # Easy problems
│   ├── medium/                        # Medium problems
│   └── hard/                          # Hard problems
│
├── 📂 deployments/                    # Deployment configs
│   ├── kubernetes/                    # K8s manifests
│   └── prometheus/                    # Prometheus config
│
├── 📂 tests/                          # Tests
│   ├── integration/                   # Integration tests
│   ├── harness/                       # Test harness tests
│   └── load/                          # Load tests
│
├── 📄 docker-compose.yml              # Docker Compose
├── 📄 Makefile                        # Build commands
├── 📄 go.mod                          # Go module
├── 📄 README.md                       # This file
└── 📄 proto/                          # gRPC proto files
```

---

## 🧪 Testing

```bash
# Run all tests
make test

# Run with coverage
make test-coverage

# Run benchmarks
go test -bench=. ./benchmarks/

# Run load tests (needs K6)
k6 run tests/load/k6-script.js
```

---

## 🚢 Deployment

### Docker Compose (Development)

```bash
# Start all services
make docker-up

# Stop all services
make docker-down

# View logs
make docker-logs
```

### Kubernetes (Production)

```bash
# Apply all manifests
kubectl apply -f deployments/kubernetes/

# Check status
kubectl get pods -n coding-challenge

# View logs
kubectl logs -f deployment/api-gateway -n coding-challenge
```

Lihat [docs/HOW_TO_RUN.md](docs/HOW_TO_RUN.md) untuk detail lengkap.

---

## 🤝 Kontribusi

Kami menerima kontribusi dari siapapun! Lihat [docs/CONTRIBUTING.md](docs/CONTRIBUTING.md) untuk panduan.

1. Fork repository
2. Buat branch baru (`git checkout -b feature/amazing-feature`)
3. Commit changes (`git commit -m 'Add amazing feature'`)
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

---

> **Dibuat dengan ❤️ menggunakan Go, React, PostgreSQL, Redis, RabbitMQ, Docker, dan Kubernetes.**
