# 📖 Coding Challenge Platform - Panduan Penggunaan

> Panduan lengkap menggunakan platform coding challenge untuk user, admin, dan developer.

---

## 📋 Daftar Isi

1. [Untuk User (Peserta)](#untuk-user-peserta)
2. [Untuk Admin](#untuk-admin)
3. [API Reference](#api-reference)
4. [Swagger UI](#swagger-ui)
5. [Keyboard Shortcuts](#keyboard-shortcuts)
6. [Badge & Achievement](#badge--achievement)
7. [DP Visualization](#dp-visualization)
8. [Monitoring](#monitoring)
9. [FAQ](#faq)

---

## 🧑‍💻 Untuk User (Peserta)

### 1. Registrasi & Login

#### Registrasi Akun Baru
1. Buka `http://localhost:9100` di browser
2. Klik tombol **"Register"** di pojok kanan atas
3. Isi form registrasi
4. Klik **"Create Account"`

#### Login
1. Klik **"Login"** di pojok kanan atas
2. Masukkan **email** dan **password**
3. Klik **"Sign In"**

---

### 2. Dashboard & Navigasi

Setelah login, Anda akan melihat halaman utama dengan:
- **Sidebar kiri** — Filter by category & difficulty
- **Problem Grid** — Cards untuk setiap problem
- **Search Bar** — Cari problem by title/tags realtime
- **Sort Dropdown** — Sort by difficulty, title

#### Navigasi
| Menu | Deskripsi |
|------|-----------|
| **🏠 Home** | Dashboard dengan statistik dan soal unggulan |
| **📝 Problems** | Daftar semua soal dengan filter & search |
| **🏆 Leaderboard** | Peringkat pengguna (Weekly/Monthly/All-time) |
| **👤 Profile** | Profil, statistik, badges, riwayat |
| **⚙️ Admin** | Admin panel (khusus admin) |

---

### 3. Mengerjakan Soal

#### Memilih Soal
1. Buka halaman **"Problems"**
2. Gunakan **filter** untuk mencari soal:
   - **Difficulty**: Easy / Medium / Hard
   - **Category**: Array / String / DP / Stack / dll
   - **Status**: Solved / Attempted / Unsolved
   - **Tags**: Cari soal dengan tag spesifik
   - **Search**: Cari berdasarkan judul
3. Klik **card soal** untuk membuka detail

#### Halaman Detail Soal
Halaman terbagi menjadi 2 panel:
- **Kiri**: Deskripsi, contoh, constraints, hints
- **Kanan**: Monaco Code Editor + test results

Di mobile, panel ditampilkan sebagai tab yang bisa di-switch.

#### Menulis Kode
1. **Template code** sudah tersedia
2. Fitur editor:
   - Syntax highlighting Go
   - Auto-completion (snippets: for, if, func)
   - 5 themes (Monokai, Dracula, GitHub Light, VS Code Dark/Light)
   - Font size control (12-24px)
   - Line numbers, bracket matching

#### Submit Kode
1. Klik **"▶ Submit"** atau tekan `Ctrl+Enter`
2. Status akan muncul real-time via WebSocket:
   - ⏳ Pending → 🔄 Running → ✅ Accepted / ❌ Wrong Answer
3. Hasil per test case ditampilkan dengan animasi fade-in

---

### 4. Sistem Hint (3 Level)

| Level | Isi | Penalti Score |
|-------|-----|---------------|
| **Hint 1** | Petunjuk umum (approach/pattern) | -10% |
| **Hint 2** | Petunjuk teknis (subproblem breakdown) | -20% |
| **Hint 3** | Petunjuk lanjutan (pseudocode/solusi hampir lengkap) | -30% |

Hint juga bisa terbuka otomatis setelah beberapa kali gagal:
- Hint 1: setelah 2 failed attempts
- Hint 2: setelah 5 failed attempts
- Hint 3: setelah 10 failed attempts

---

### 5. Leaderboard

| Tab | Periode | Score |
|-----|---------|-------|
| **Weekly** | Minggu ini | Reset tiap minggu |
| **Monthly** | Bulan ini | Reset tiap bulan |
| **All Time** | Sepanjang masa | Akumulasi |

#### Badge
| Badge | Kriteria |
|-------|----------|
| 🥇 Gold | Top 1% |
| 🥈 Silver | Top 5% |
| 🥉 Bronze | Top 10% |
| 🔥 Streak Master | 7+ hari berturut-turut |
| 💪 Grinder | 50+ problems solved |
| 🧠 Genius | 5 hard problems solved |

#### Achievement
| Achievement | Cara Mendapatkan |
|-------------|------------------|
| 🎯 First Solve | Problem pertama selesai |
| ⚡ Speed Demon | Solve dalam <2 menit |
| 🏃 Marathon | 10 problems dalam sehari |
| 💎 Perfectionist | Semua test case pass percobaan pertama |

---

### 6. DP Visualization

Untuk soal Dynamic Programming, tersedia visualisasi interaktif:
- **Step-by-step animation** — Lihat DP table terisi perlahan
- **Slider** — Navigasi steps maju/mundur
- **Play/Pause** — Auto-play dengan speed control
- **Tabulation vs Memoization** — Side-by-side comparison
- **Backtracking path** — Lihat jalur optimal
- **Export** — Simpan visualisasi sebagai gambar

Didukung untuk algoritma: Fibonacci, Knapsack, LCS, Edit Distance, Coin Change, LIS.

---

### 7. Keyboard Shortcuts

| Shortcut | Aksi |
|----------|------|
| `Ctrl + Enter` | Submit kode |
| `Ctrl + R` | Reset kode ke template |
| `Ctrl + S` | Simpan draft (localStorage) |
| `Ctrl + /` | Toggle keyboard shortcuts help |
| `Ctrl + D` | Toggle dark/light theme |
| `Ctrl + Shift + P` | Toggle problem list panel |
| `Esc` | Close modal/panel |
| `Tab` | Indent (4 spaces) |

---

## 👨‍💼 Untuk Admin

### 1. Login Admin
Login dengan akun yang memiliki role `admin`. Menu tambahan akan muncul.

### 2. Admin Dashboard
- **Stats Cards**: Total users, submissions, problems, acceptance rate
- **Chart**: Submissions over time (7 days)
- **Activity Feed**: Recent submissions, registrations
- **Server Health**: Status semua service

### 3. Manajemen Soal
Buka Admin Panel → Problems:
- **Create**: Title, slug, description (Markdown), difficulty, category, tags
- **Test Cases**: Input, expected output, hidden/sample toggle
- **Hints**: 3 levels dengan score penalty
- **Template Code**: Monaco Editor untuk starter code
- **Solution**: Hidden solution dengan approach explanation
- **Publish/Draft**: Toggle status

### 4. Swagger UI
Dokumentasi API interaktif tersedia di:
```
http://localhost:9100/swagger/index.html
```

Dari sini Anda bisa:
- Lihat semua endpoint
- Test request langsung dari browser
- Lihat schema request/response
- Download OpenAPI spec

---

## 🔌 API Reference

### Base URL
```
Development: http://localhost:9100/api
Production:  https://api.codingchallenge.com/api
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
| POST | `/api/problems` | Buat soal (Admin) |

### Submissions
| Method | Endpoint | Deskripsi |
|--------|----------|-----------|
| POST | `/api/submissions` | Submit kode |
| GET | `/api/submissions/:id` | Get status |
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
| POST | `/api/problems/:id/hints/:hintId/use` | Gunakan hint |

---

## 📊 Monitoring

### Grafana Dashboards
Setelah monitoring stack jalan (`docker-compose --profile monitoring up -d`):
- **Services Dashboard**: CPU, Memory, Request rate, Error rate, Latency p95 — `http://localhost:3000/d/services`
- **Database Dashboard**: Connections, query time, cache hit ratio — `http://localhost:3000/d/database`
- Login: `admin / admin`

### Prometheus Metrics
```
http://localhost:9090
```

### Service Health
```bash
curl http://localhost:9100/health
```

---

## ❓ FAQ

**Q: Bahasa pemrograman apa yang didukung?**
A: Saat ini **Golang**.

**Q: Berapa limit waktu?**
A: Easy: 1s, Medium: 3s, Hard: 5s

**Q: Berapa limit memori?**
A: Easy: 256MB, Medium: 512MB, Hard: 1GB

**Q: Kok error 401?**
A: Pastikan environment variables diset:
```bash
export JWT_ACCESS_SECRET=*** JWT_REFRESH_SECRET=***
```

**Q: Kok error 404?**
A: Cek port — sekarang pakai **9100-9107** (bukan 8080).

**Q: Cara reset database?**
```bash
make migrate-down && make migrate
```

**Q: Cara lihat logs?**
```bash
docker-compose logs -f
make docker-logs
```

---

> **25 soal • 20 loops improvement • Build ✅ Tests ✅**
> Selamat belajar coding! 🚀
