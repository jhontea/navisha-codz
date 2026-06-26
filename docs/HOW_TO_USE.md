# 📖 Coding Challenge Platform - Panduan Penggunaan

> Panduan lengkap menggunakan platform coding challenge — 25 soal, 9 microservices, 30 loops improvement.

---

## 📋 Daftar Isi

1. [Untuk User](#untuk-user-peserta)
2. [Untuk Admin](#untuk-admin)
3. [API Reference](#api-reference)
4. [Keyboard Shortcuts](#keyboard-shortcuts)
5. [Badge & Achievement](#badge--achievement)
6. [Fitur Lanjutan](#fitur-lanjutan)
7. [FAQ](#faq)

---

## 🧑‍💻 Untuk User

### Registrasi & Login

Buka `http://localhost:9100` → Register/Login di pojok kanan atas.

### Dashboard
- **Sidebar kiri** — Filter soal by category & difficulty
- **Search bar** — Cari realtime by title/tags
- **Sort dropdown** — Sort by difficulty, title
- **Problem cards** — Status indicator (solved/attempted/unsolved)

### Mengerjakan Soal

1. Klik card soal → detail page (2 panel)
2. **Kiri**: Deskripsi, contoh, constraints, hints
3. **Kanan**: Monaco Editor + test results
4. Tulis kode di function yang disediakan
5. `Ctrl+Enter` → Submit (real-time via WebSocket)

### Hint System (3 Level)
| Level | Isi | Penalti | Auto-unlock |
|-------|-----|---------|-------------|
| Hint 1 | Approach/pattern | -10% | After 2 failed attempts |
| Hint 2 | Subproblem breakdown | -20% | After 5 failed attempts |
| Hint 3 | Pseudocode | -30% | After 10 failed attempts |

### Leaderboard
| Tab | Periode | Fitur |
|-----|---------|-------|
| Weekly | Minggu ini | Medals, reward |
| Monthly | Bulan ini | ELO rating |
| All Time | Sepanjang masa | Badge display |

---

## 👨‍💼 Untuk Admin

### Admin Dashboard
- **Stats**: Total users (2,847), submissions (18,432), problems (156), acceptance rate (67.3%)
- **Chart**: Submissions over 7 days
- **Activity Feed**: Recent registrations & submissions
- **Server Health**: Status 9 services (API, Auth, Problem, Execution, Worker, Leaderboard, Hint, WebSocket, Notification)

### Problem Management
1. Admin Panel → Problems → Create/Edit
2. Isi: title, slug, description (Markdown), difficulty, category, tags
3. Test Cases: input, expected output, hidden/sample toggle
4. Hints: 3 levels with score penalty
5. Solution: Hidden code + approach explanation
6. Publish/Draft toggle

### Monitoring
```bash
# Start monitoring
docker-compose --profile monitoring up -d

# Grafana: http://localhost:3000 (admin/admin)
# Prometheus: http://localhost:9090
```

### Error Tracking (Sentry)
Set `SENTRY_DSN` di `.env` untuk aktifkan real-time error monitoring.

---

## 🔌 API Reference

### Base URL
```
Legacy:  http://localhost:9100/api
v1:      http://localhost:9100/v1
```

### Auth
| Method | Endpoint | Deskripsi |
|--------|----------|-----------|
| POST | `/auth/register` | Registrasi |
| POST | `/auth/login` | Login |
| POST | `/auth/refresh` | Refresh token |
| POST | `/auth/logout` | Logout |

### Problems
| Method | Endpoint | Deskripsi |
|--------|----------|-----------|
| GET | `/api/problems` | List (filter: difficulty, category, tags) |
| GET | `/api/problems/:id` | Detail |
| GET | `/api/problems/:id/template` | Template code |
| POST | `/api/problems` | Create (Admin) |

### Submissions
| Method | Endpoint | Deskripsi |
|--------|----------|-----------|
| POST | `/api/submissions` | Submit kode |
| GET | `/api/submissions/:id` | Status |
| POST | `/api/validate` | Validasi syntax |

### Leaderboard
| Method | Endpoint | Deskripsi |
|--------|----------|-----------|
| GET | `/api/leaderboard/weekly` | Peringkat mingguan |
| GET | `/api/leaderboard/monthly` | Peringkat bulanan |
| GET | `/api/leaderboard/all-time` | Peringkat keseluruhan |

### Hints
| Method | Endpoint | Deskripsi |
|--------|----------|-----------|
| GET | `/api/problems/:id/hints` | Get hints |
| POST | `/api/problems/:id/hints/:hintId/use` | Use hint |

---

## ⌨️ Keyboard Shortcuts

| Shortcut | Aksi |
|----------|------|
| `Ctrl+Enter` | Submit kode |
| `Ctrl+R` | Reset kode ke template |
| `Ctrl+S` | Simpan draft (localStorage) |
| `Ctrl+D` | Toggle dark/light theme |
| `Ctrl+/` | Toggle keyboard shortcuts help |
| `Ctrl+Shift+P` | Toggle problem list panel |
| `Esc` | Close modal/panel |

---

## 🏆 Badge & Achievement

### Badge
| Badge | Kriteria |
|-------|----------|
| 🥇 Gold | Top 1% leaderboard |
| 🥈 Silver | Top 5% leaderboard |
| 🥉 Bronze | Top 10% leaderboard |
| 🔥 Streak Master | 7+ hari berturut-turut |
| 💪 Grinder | 50+ problems solved |
| 🧠 Genius | 5 hard problems solved |

### Achievement
| Achievement | Cara |
|-------------|------|
| 🎯 First Solve | Problem pertama selesai |
| ⚡ Speed Demon | Solve dalam <2 menit |
| 🏃 Marathon | 10 problems dalam sehari |
| 💎 Perfectionist | Semua test case pass first try |

---

## 🔧 Fitur Lanjutan

### API Versioning
```
GET /v1/problems      # Recommended (versioned)
GET /api/problems      # Legacy (backward compatible)
Response header: X-API-Version: v1
```

### Rate Limit Tiers
| Tier | `/run` | GET | Header |
|------|--------|-----|--------|
| Free | 10/min | 30/min | `X-RateLimit-Tier: free` |
| Premium | 100/min | 300/min | `X-RateLimit-Tier: premium` |
| Admin | Unlimited | Unlimited | `X-RateLimit-Tier: admin` |

### Swagger UI
```
http://localhost:9100/swagger/index.html
```

### WebSocket
```
ws://localhost:9100/ws
— Real-time submission updates
```

---

## ❓ FAQ

**Q: Bahasa yang didukung?**
A: Golang.

**Q: Limit waktu?**
A: Easy 1s, Medium 3s, Hard 5s.

**Q: Limit memori?**
A: Easy 256MB, Medium 512MB, Hard 1GB.

**Q: Error 401?**
A: Set `JWT_ACCESS_SECRET` dan `JWT_REFRESH_SECRET` di environment.

**Q: Error 404?**
A: Cek port — pakai **9100-9108** (bukan 8080).

**Q: Cara reset database?**
```bash
make migrate-down && make migrate
```

**Q: Cara start monitoring?**
```bash
docker-compose --profile monitoring up -d
# Grafana: http://localhost:3000 (admin/admin)
```

---

> **25 Soal • 9 Services • 30 Loops Improvement • Build ✅ Tests ✅**
