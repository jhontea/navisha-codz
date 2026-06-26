# Changelog

All notable changes to Coding Challenge Platform will be documented in this file.

## [1.0.0] - 2026-06-26

### Added
- 🎯 **15+ Coding Problems** with 3 difficulty levels (Easy/Medium/Hard)
- 💻 **Monaco Code Editor** with Go syntax highlighting, 5 themes, snippets, auto-completion
- 🔄 **Real-time Code Execution** via WebSocket with Docker sandbox
- 💡 **Progressive Hint System** (3-level hints with score penalty)
- 🏆 **ELO Rating + Leaderboard** (Weekly/Monthly/All-time)
- 🎖️ **Badge & Achievement System** (6 badges, 4 achievements)
- 🧠 **DP Visualization** with step-by-step animation (Fibonacci, Knapsack, LCS, Edit Distance, Coin Change, LIS)
- 👤 **User Profile** with stats, rating history, submission history
- 🔐 **JWT Authentication** with access/refresh tokens, 2FA/TOTP support
- 📱 **Responsive Design** with mobile-first approach
- 🌗 **Dark/Light Theme** with system preference detection
- 📦 **PWA Support** with service worker, offline caching, install prompt

### Backend
- 🔧 **Go Microservices** (6 services: Auth, Problem, Execution, Leaderboard, Hint, WebSocket)
- 🐳 **Docker Sandbox** with --network=none, --read-only, seccomp, AppArmor
- ⚙️ **RabbitMQ** async code execution with priority queues, DLQ, retry
- 🗄️ **PostgreSQL** with optimized schema (13 tables, 40+ indexes)
- ⚡ **Redis** caching with cache-aside, circuit breaker, cluster support
- 🔍 **Prometheus + Grafana** monitoring (RED metrics)
- 📊 **ELK Stack** centralized logging
- 🔄 **GitHub Actions** CI/CD with auto-deploy

### Infrastructure
- 🏗️ **Kubernetes** manifests with HPA, PDB, Network Policies
- 🌐 **Terraform** AWS IaC (VPC, EKS, RDS, ElastiCache, MQ, ECR)
- 🐳 **Docker Compose** for local development
- 📚 **Comprehensive Documentation** (10+ docs)

### Security
- 🛡️ Input sanitization (SQL injection, XSS prevention)
- 🚫 Rate limiting (sliding window per user/IP)
- 🔒 Content Security Policy headers
- 🎫 JWT with refresh rotation and device fingerprinting
- 👾 Sandbox escape detection

## [0.9.0] - 2026-06-20

### Added
- Initial architecture design
- Database schema (PostgreSQL)
- Frontend React + Monaco setup
- Docker sandbox implementation
- Basic problem loader (YAML)
- Hint system (3 levels)

### Changed
- Migrated from monolithic to microservices architecture

## [0.1.0] - 2026-06-10

### Added
- Project initialization
- Go module setup
- Basic API endpoints
- YAML problem format specification
