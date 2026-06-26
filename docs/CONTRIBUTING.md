# Contributing to Coding Challenge Platform

Kami senang Anda ingin berkontribusi! Berikut panduan untuk memulai.

## 📋 Prasyarat

- Go 1.25+
- Node.js 18+
- Docker & Docker Compose
- PostgreSQL 15+
- Redis 7+
- RabbitMQ 3.12+

## 🚀 Setup Development

```bash
# Clone repo
git clone https://github.com/codingchallenge/platform.git
cd platform

# Setup environment
make setup

# Install Go dependencies
go mod download

# Install Frontend dependencies
cd web/frontend
npm install
cd ../..

# Build all services
make build

# Run tests
make test
```

## 🔧 Development Workflow

### Branch Strategy
- `main` — Production-ready code
- `develop` — Integration branch
- `feature/*` — New features
- `fix/*` — Bug fixes
- `docs/*` — Documentation

### Commit Convention
```
<type>(<scope>): <description>

Types: feat, fix, docs, style, refactor, perf, test, chore, ci
Scopes: auth, problem, execution, leaderboard, hint, frontend, infra, docs

Examples:
feat(auth): add 2FA TOTP support
fix(execution): timeout handling for large test cases
docs(api): update submission endpoint docs
```

### Code Standards

**Go:**
- `gofmt` before commit
- Error wrapping with `fmt.Errorf("...: %w", err)`
- Context propagation for cancellation
- Structured logging via `pkg/logger`
- Tests with `testing` package

**TypeScript/React:**
- Strict TypeScript mode
- Functional components with hooks
- Tailwind CSS for styling
- Zustand for state management
- React Query for server state

## 🧪 Testing

```bash
# Run all tests
make test

# Run specific package tests
make test-services/auth-service

# Run frontend tests
cd web/frontend && npm test

# Run integration tests
go test ./tests/integration/... -v

# Run benchmarks
go test -bench=. ./benchmarks/

# Load test (requires K6)
k6 run tests/load/k6-script.js
```

## 📦 Build & Deploy

```bash
# Build all services
make build

# Build Docker images
make docker-build

# Local deployment
make docker-up

# Production deployment (requires K8s)
./scripts/deploy.sh production
```

## 📝 Pull Request Process

1. Fork repository
2. Buat branch `feature/your-feature`
3. Commit dengan pesan jelas
4. Push ke branch
5. Buat Pull Request ke `develop`
6. Tunggu review dari maintainer
7. Setelah approve, merge akan dilakukan

## 🔍 Code Review Guidelines

Reviewer akan memeriksa:
- ✅ Kode mengikuti standar Go/TypeScript
- ✅ Error handling proper
- ✅ Tests mencakup edge cases
- ✅ Dokumentasi diupdate
- ✅ Tidak ada security vulnerability
- ✅ Performance OK

## 📚 Additional Resources

- [Architecture Docs](docs/ARCHITECTURE_V2.md)
- [API Docs](docs/API.md)
- [Database Schema](docs/DATABASE_SCHEMA.md)
- [Deployment Guide](docs/HOW_TO_RUN.md)

## ❓ Questions?

Buka issue di GitHub atau chat di Discord: https://discord.gg/codingchallenge
