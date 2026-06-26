# Security Policy

## Supported Versions

| Version | Supported          |
| ------- | ------------------ |
| 1.x     | ✅ Active support  |
| < 1.0   | ❌ Not supported   |

## 🔒 Security Features

### Code Execution Sandbox
- Docker container with `--network=none` (no network access)
- `--read-only` filesystem (code mounted as read-only)
- Seccomp profile limiting syscalls to Go-compatible only
- All capabilities dropped with `--cap-drop=ALL`
- Memory limit (256MB-1GB based on difficulty)
- CPU limit (1 CPU core)
- PIDs limit (50 max processes)
- Output size limit (1MB max)
- Timeout enforcement (1-5 seconds)

### Authentication & Authorization
- JWT with access (15min) and refresh (7 days) tokens
- Token rotation with refresh token invalidation
- Device fingerprinting for session tracking
- Rate limiting: 10 req/min for /run, 30/min for /validate, 60/min for GET
- Password policy: min 8 chars, must contain uppercase, lowercase, number
- 2FA/TOTP support for admin accounts

### Data Protection
- All passwords hashed with bcrypt (cost 12)
- All database connections encrypted (TLS 1.3)
- Redis connections with ACL authentication
- RabbitMQ with TLS and user authentication
- Secrets stored in K8s secrets (encrypted at rest)

### Network Security
- API Gateway validates all requests
- CORS restricted to allowed origins
- CSP headers prevent XSS attacks
- Rate limiting at gateway level
- DDoS protection via CloudFront/WAF

## 🐛 Reporting a Vulnerability

Jika Anda menemukan security vulnerability:

1. **Jangan buka issue publik** — email ke security@codingchallenge.com
2. Sertakan detail lengkap:
   - Type of vulnerability
   - Steps to reproduce
   - Potential impact
   - Suggested fix (optional)
3. Kami akan merespon dalam 48 jam
4. Setelah fix, kami akan mengakui kontribusi Anda

## 🔐 Disclosure Policy

- 90 days disclosure timeline
- CVE assignment after fix
- Credit in security acknowledgments

## ✅ Security Checklist (for developers)

- [ ] Input validation for all user inputs
- [ ] Parameterized queries (no raw SQL)
- [ ] Output encoding for HTML responses
- [ ] Proper error messages (no stack traces)
- [ ] Rate limiting on all endpoints
- [ ] Authentication on all protected routes
- [ ] Authorization checks (RBAC)
- [ ] HTTPS in production
- [ ] CORS configuration
- [ ] CSP headers
- [ ] No hardcoded secrets
- [ ] Dependency scanning (Dependabot)
- [ ] Container scanning (Trivy)
