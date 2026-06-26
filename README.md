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
![Loops](https://img.shields.io/badge/Loops-30-orange)
![License](https://img.shields.io/badge/License-MIT-green.svg)

---

## 📋 Daftar Isi

- [Fitur Utama](#fitur-utama)
- [Tech Stack](#tech-stack)
- [Quick Start](#quick-start)
- [Dokumentasi](#dokumentasi)
- [Arsitektur](#arsitektur)
- [API Reference](#api-reference)
- [30 Loop Improvement](#30-loop-improvement)
- [Kontribusi](#kontribusi)

---

## ✨ Fitur Utama

### Untuk User
- ✅ **25 Soal** — Algoritma, Data Structure, Dynamic Programming (3 tingkat kesulitan)
- ✅ **Tag Filter** — Filter soal berdasarkan tags (`?tags=hash-map,dp`)
- ✅ **Real-time Code Execution** — Docker sandbox dengan Go compiler
- ✅ **Progressive Hint System** — 3 level hints, auto-unlock after failed attempts
- ✅ **Badge & Achievement** — 6 badges (Gold/Silver/Bronze/Streak/Grinder/Genius)
- ✅ **Leaderboard** — Weekly, Monthly, All-time ranking dengan ELO rating
- ✅ **Code Editor** — Monaco Editor dengan 5 themes + Go snippets + auto-completion
- ✅ **DP Visualization** — Step-by-step animasi (Fibonacci, Knapsack, LCS, Edit Distance, Coin Change, LIS)
- ✅ **Real-time Updates** — Submission status via WebSocket
- ✅ **Keyboard Shortcuts** — Ctrl+Enter submit, Ctrl+R reset, Ctrl+Shift+P toggle panel

### Untuk Admin
- ✅ **Problem Management** — CRUD dengan test cases, hints, template, solution
- ✅ **User Management** — Ban/unban, role management (user/premium/admin)
- ✅ **Admin Dashboard** — Statistik, charts, server health, activity feed
- ✅ **Swagger UI** — Dokumentasi API interaktif di `/swagger/index.html`
- ✅ **Email Notifications** — SMTP-based notification service
- ✅ **Sentry Error Tracking** — Real-time error monitoring

### Teknis
- ✅ **Microservices** — 9 services (API Gateway, Auth, Problem, Execution, Worker, Leaderboard, Hint, WebSocket, Notification)
- ✅ **API Versioning** — `/v1/` prefix + backward compatibility
- ✅ **Docker Sandbox** — seccomp + AppArmor + network isolation
- ✅ **Rate Limiting** — 3 tiers (Free/Premium/Admin) with sliding window
- ✅ **Redis Caching** — Cache warming, cursor pagination, TTL 5 menit
- ✅ **RabbitMQ** — Priority queue + DLQ + work stealing
- ✅ **Prometheus + Grafana** — Monitoring dashboards auto-provisioned
- ✅ **CORS Security** — Specific origin dari env var + CSRF protection
- ✅ **E2E Tests** — Playwright test suite
- ✅ **Terraform IaC** — AWS: VPC, EKS, RDS, ElastiCache, MQ

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
| **Auth** | JWT (access + refresh tokens) |
| **Container** | Docker 24+ (multi-stage builds) |
| **Orchestration** | Kubernetes 1.28+ |
| **Monitoring** | Prometheus + Grafana |
| **Error Tracking** | Sentry |
| **E2E Tests** | Playwright |
| **CI/CD** | GitHub Actions |
| **IaC** | Terraform (AWS) |
| **API Docs** | Swagger/OpenAPI 2.0 |

---

## 🚀 Quick Start

```bash
# Prasyarat
export JWT_ACCESS_SECRET=*** JWT_REFRESH_SECRET=***

# Docker Compose
docker-compose up -d --build
# Buka http://localhost:9100

# Atau manual
go run services/api-gateway/main.go        # :9100
go run services/auth-service/main.go       # :9101
go run services/problem-service/main.go    # :9102
go run services/execution-service/main.go  # :9103
go run services/leaderboard-service/main.go # :9104
go run services/hint-service/main.go       # :9105
go run services/execution-worker/main.go   # :9106
go run services/websocket-service/main.go  # :9107
go run services/notification-service/main.go # :9108
```

---

## 📖 Dokumentasi

| Dokumen | Deskripsi |
|---------|-----------|
| [docs/HOW_TO_RUN.md](docs/HOW_TO_RUN.md) | Panduan menjalankan aplikasi |
| [docs/HOW_TO_USE.md](docs/HOW_TO_USE.md) | Panduan penggunaan (user & admin) |
| [docs/ARCHITECTURE_V2.md](docs/ARCHITECTURE_V2.md) | Diagram arsitektur sistem |
| [docs/API.md](docs/API.md) | API documentation + versioning |
| [docs/FINAL_REPORT.md](docs/FINAL_REPORT.md) | Laporan final build & test |
| [docs/PERFORMANCE_REPORT.md](docs/PERFORMANCE_REPORT.md) | Benchmark & load test |
| [docs/MONITORING_SETUP.md](docs/MONITORING_SETUP.md) | Setup Prometheus + Grafana |
| [docs/SECURITY_AUDIT.md](docs/SECURITY_AUDIT.md) | Security audit findings |
| [docs/QUERY_OPTIMIZATION.md](docs/QUERY_OPTIMIZATION.md) | Database optimization |
| [docs/SECURITY.md](docs/SECURITY.md) | Security policy |
| [docs/CONTRIBUTING.md](docs/CONTRIBUTING.md) | Panduan kontribusi |
| [docs/ROADMAP.md](docs/ROADMAP.md) | Rencana pengembangan |

---

## 🏗 Arsitektur

```
┌─────────────────────────────────────────────────────────────────┐
│                       Browser (Client)                           │
│  ┌──────────────┐  ┌──────────────┐  ┌───────────────────────┐ │
│  │ Problem List │  │ Code Editor  │  │ Test Results + Hints │ │
│  │   (React)    │  │  (Monaco)    │  │   (Real-time WS)     │ │
│  └──────────────┘  └──────────────┘  └───────────────────────┘ │
└──────────────────────────┬──────────────────────────────────────┘
                           │ HTTP / WebSocket
┌──────────────────────────▼──────────────────────────────────────┐
│                    API Gateway (Go/Gin)                          │
│  Rate Limiting • JWT • Gzip • Timeout • CORS • CSRF • Sentry   │
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
│                    Data Layer + Services                          │
│  ┌──────────────┐  ┌──────────────┐  ┌──────────────────────┐  │
│  │  PostgreSQL  │  │    Redis     │  │      RabbitMQ        │  │
│  │  (Primary)   │  │   (Cache)    │  │  (Job + Notif Queue) │  │
│  └──────────────┘  └──────────────┘  └──────────────────────┘  │
│  ┌──────────────────┐  ┌────────────────────────────────────┐  │
│  │ Execution Worker │  │     Notification Service (:9108)   │  │
│  │  (Docker Sandbox)│  │  (SMTP + RabbitMQ consumer)        │  │
│  └──────────────────┘  └────────────────────────────────────┘  │
└──────────────────────────────────────────────────────────────────┘
```

---

## 🔌 API Reference

### Base URL
```
Development: http://localhost:9100/api
Versioned:   http://localhost:9100/v1
```

### Endpoints

| Method | Endpoint | Deskripsi | Auth | Tier |
|--------|----------|-----------|------|------|
| GET | `/health` | Health check | ❌ | - |
| POST | `/auth/register` | Registrasi | ❌ | Free |
| POST | `/auth/login` | Login | ❌ | Free |
| POST | `/auth/refresh` | Refresh token | ❌ | Free |
| GET | `/api/problems` | List soal (filter: difficulty, category, tags, page) | ❌ | Free |
| GET | `/api/problems/:id` | Detail soal | ❌ | Free |
| GET | `/api/problems/:id/template` | Template code | ❌ | Free |
| POST | `/api/submissions` | Submit kode | ✅ | Free/Premium |
| GET | `/api/submissions/:id` | Submission status | ✅ | Free |
| GET | `/api/leaderboard/*` | Leaderboard | ❌ | Free |
| GET | `/api/problems/:id/hints` | Get hints | ✅ | Free |
| POST | `/api/validate` | Validasi syntax | ✅ | Free |
| GET | `/swagger/index.html` | Swagger UI | ❌ | - |
| WS | `/ws` | WebSocket updates | ✅ | Free |

### Rate Limit Tiers
| Tier | `/run` | GET endpoints | Admin |
|------|--------|---------------|-------|
| **Free** | 10 req/min | 30 req/min | - |
| **Premium** | 100 req/min | 300 req/min | - |
| **Admin** | Unlimited | Unlimited | ✅ |

### API Versioning
- Legacy: `/api/...` — backward compatible
- Versioned: `/v1/...` — recommended
- Header: `X-API-Version: v1`

---

## 🗂 Service Ports

| Service | Port | Health |
|---------|------|--------|
| API Gateway | **9100** | ✅ |
| Auth Service | **9101** | ✅ |
| Problem Service | **9102** | ✅ |
| Execution Service | **9103** | ✅ |
| Leaderboard Service | **9104** | ✅ |
| Hint Service | **9105** | ✅ |
| Execution Worker | **9106** | ✅ |
| WebSocket Service | **9107** | ✅ |
| Notification Service | **9108** | ✅ |

---

## 📊 30 Loop Improvement

| # | Fokus | Hasil |
|---|-------|-------|
| 1 | Foundation | Architecture docs, fix tests, 0 TS errors |
| 2-3 | Core Fixes | Hapus monolith, security, konsolidasi |
| 4 | Problems + UI | 25 problems, tag filter, search, sort |
| 5-6 | Perf + Security | Cursor pagination, Redis cache, Gzip |
| 7-8 | QA + Docs | Testing, SUMMARY.md, .env.example |
| 9-10 | UI + Final | Skeleton loading, mobile tabs, final report |
| 11 | Code Review | 7 findings fixed, CORS hardened |
| 12 | Test Coverage | +3 test packages (security, rabbitmq, worker) |
| 13 | More Problems | 20 → 25 problems |
| 14 | Frontend Tests | Vitest setup, component tests |
| 15 | Swagger | OpenAPI annotations, Swagger UI |
| 16 | Load Test | K6 scripts, benchmark, performance report |
| 17 | Docker | Multi-stage builds, .dockerignore |
| 18 | Monitoring | Grafana dashboards, Prometheus, setup guide |
| 19 | Hardening | Nil pointer checks, recover middleware |
| 20 | Final | All checks PASS, production ready |
| 21 | Problems | +5 problems (climbing-stairs, best-time-to-buy, etc) |
| 22 | E2E Tests | Playwright (home, problem, auth specs) |
| 23 | API Versioning | /v1/ prefix + backward compat |
| 24 | Rate Limiting | Free/Premium/Admin tiers |
| 25 | Notifications | Email service via RabbitMQ + SMTP |
| 26 | Database | Index optimization + QUERY_OPTIMIZATION.md |
| 27 | Error Tracking | Sentry integration (Gin middleware) |
| 28 | Lighthouse | Lazy loading, preload, SEO, OG tags |
| 29 | Security Audit | CSRF, SQL injection, sensitive data check |
| 30 | Final | Build ✅ Tests ✅ Vet ✅ Production ready |

---

## Kontribusi

1. Fork repository
2. Buat branch (`git checkout -b feature/xxx`)
3. Commit (`git commit -m 'feat: ...'`)
4. Push & Pull Request

---

> **30 Loops • Build ✅ Tests ✅ 10 Packages ✅ 25 Problems ✅**
> **Dibuat dengan ❤️ menggunakan Go, React, PostgreSQL, Redis, RabbitMQ, Docker, Kubernetes**
