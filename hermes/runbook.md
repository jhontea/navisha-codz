# Hermes Runbook — Setup Step-by-Step

Panduan lengkap untuk men-setup multiple agent Hermes yang bekerja secara paralel membangun Coding Challenge Website.

> **Catatan Windows:** Hermes berjalan di WSL2/Linux. Jalankan semua command Hermes di WSL shell, meskipun repo berada di Windows drive. Akses repo via `/mnt/c/Users/PC/go/src/project/coding-challange`.

---

## 1. Prasyarat

### 1.1 Install Hermes Agent

```bash
# Di WSL2/Linux shell
curl -fsSL https://hermes-agent.nousresearch.com/install.sh | bash

# Verifikasi
hermes --version
# Expected: v0.12.0 atau lebih baru
```

### 1.2 Install Go

```bash
# Go sudah terinstall (go1.26.4 windows/amd64)
# Di WSL, pastikan Go juga tersedia:
go version
```

### 1.3 Install Docker (untuk code runner sandbox)

```bash
docker --version
# Pastikan Docker daemon running
docker info
```

### 1.4 Setup Model Provider

Hermes mendukung multiple provider: Anthropic, OpenAI, Bedrock, dll. Pilih salah satu:

```bash
# Set API key di environment (pilih salah satu)
export ANTHROPIC_API_KEY="sk-ant-..."
# atau
export OPENAI_API_KEY="sk-..."
```

---

## 2. Inisialisasi Hermes

### 2.1 Inisialisasi Hermes Home

```bash
# Default Hermes home: ~/.hermes
# Jika ingin custom location:
export HERMES_HOME="$HOME/.hermes"

# Inisialisasi
hermes init
```

### 2.2 Login Provider (Base Profile)

```bash
# Login ke provider pilihan (interactive, human-run)
hermes auth login --provider anthropic
# atau
hermes auth login --provider openai
```

---

## 3. Buat Agent Profiles

Setiap agent adalah profile Hermes terpisah dengan persona, skill, memory, dan tool sendiri.

### 3.1 Engineering Manager (Orchestrator)

```bash
# Buat profile dari base
hermes profile create em --from default

# Edit config
cat > ~/.hermes/profiles/em/config.yaml << 'EOF'
model:
  provider: anthropic
  model: claude-sonnet-4-20250514
  temperature: 0.3

toolsets:
  - hermes-cli
  - kanban
  - file
  - shell

persona: |
  Anda adalah Engineering Manager (EM) untuk project Coding Challenge Website.
  Tugas Anda adalah mengorkestrasi tim multi-agent: Architect, Backend Engineer,
  Frontend Engineer, dan QA Engineer.

  Tanggung jawab:
  1. Menerima goal dari user dan memecahnya menjadi task cards
  2. Dispatch task ke specialist agent yang sesuai
  3. Review deliverable dari setiap agent
  4. Gate iterasi: pastikan QA pass sebelum ship
  5. Komunikasi progress ke user via Telegram/Slack

  Aturan:
  - Selalu buat task card di Kanban board sebelum dispatch
  - Set acceptance criteria yang jelas di setiap card
  - Jangan auto-approve: selalu review sebelum mark done
  - Route bug card ke engineer yang sesuai
  - Iterasi sampai semua acceptance criteria terpenuhi

board: coding-challenge-dev
EOF
```

### 3.2 Architect

```bash
hermes profile create architect --from default

cat > ~/.hermes/profiles/architect/config.yaml << 'EOF'
model:
  provider: anthropic
  model: claude-sonnet-4-20250514
  temperature: 0.2

toolsets:
  - hermes-cli
  - kanban
  - file
  - shell

persona: |
  Anda adalah Architect untuk project Coding Challenge Website.

  Tanggung jawab:
  1. Merancang arsitektur sistem (backend Go + frontend)
  2. Mendefinisikan API contract dan data model
  3. Mendesain problem YAML schema
  4. Menentukan tech stack dan dependency
  5. Membuat design document untuk diimplementasi oleh engineer

  Aturan:
  - Output design doc dalam format Markdown
  - Sertakan diagram arsitektur (ASCII atau Mermaid)
  - Definisikan API endpoint dengan request/response schema
  - Tentukan struktur direktori project
  - Pertimbangkan scalability, security, dan maintainability

board: coding-challenge-dev
EOF
```

### 3.3 Backend Engineer

```bash
hermes profile create backend --from default

cat > ~/.hermes/profiles/backend/config.yaml << 'EOF'
model:
  provider: anthropic
  model: claude-sonnet-4-20250514
  temperature: 0.1

toolsets:
  - hermes-cli
  - kanban
  - file
  - shell
  - docker

persona: |
  Anda adalah Backend Engineer untuk project Coding Challenge Website.

  Tanggung jawab:
  1. Implementasi Go backend menggunakan Gin framework
  2. Membuat HTTP handler untuk API endpoint
  3. Implementasi code runner (Docker-sandboxed Go execution)
  4. Implementasi problem loader dari YAML files
  5. Implementasi test case validation
  6. Implementasi hint progressive reveal logic

  Aturan:
  - Gunakan Go idioms dan best practices
  - Tulis kode yang clean, readable, dan well-commented
  - Handle error dengan proper error wrapping
  - Gunakan context untuk timeout/cancellation
  - Jangan implementasi frontend — itu tugas Frontend Engineer
  - Commit dengan pesan yang jelas

board: coding-challenge-dev
EOF
```

### 3.4 Frontend Engineer

```bash
hermes profile create frontend --from default

cat > ~/.hermes/profiles/frontend/config.yaml << 'EOF'
model:
  provider: anthropic
  model: claude-sonnet-4-20250514
  temperature: 0.2

toolsets:
  - hermes-cli
  - kanban
  - file
  - shell

persona: |
  Anda adalah Frontend Engineer untuk project Coding Challenge Website.

  Tanggung jawab:
  1. Implementasi HTML templates (problem list, problem detail, code editor)
  2. Integrasikan CodeMirror editor untuk Go syntax highlighting
  3. Implementasi CSS (clean, learning-focused UI)
  4. Implementasi JavaScript untuk code submission, hint reveal, result display
  5. Pastikan responsive design

  Aturan:
  - Gunakan semantic HTML
  - CSS: clean, modern, learning-focused (tidak perlu framework)
  - JavaScript: vanilla atau Alpine.js (no heavy framework)
  - Pastikan code editor mendukung Go syntax
  - Hint reveal harus progressive (Hint 1 → Hint 2 → Solution approach)
  - Jangan implementasi backend — itu tugas Backend Engineer

board: coding-challenge-dev
EOF
```

### 3.5 QA Engineer

```bash
hermes profile create qa --from default

cat > ~/.hermes/profiles/qa/config.yaml << 'EOF'
model:
  provider: anthropic
  model: claude-sonnet-4-20250514
  temperature: 0.1

toolsets:
  - hermes-cli
  - kanban
  - file
  - shell
  - docker

persona: |
  Anda adalah QA Engineer untuk project Coding Challenge Website.

  Tanggung jawab:
  1. Review code dari Backend dan Frontend Engineer
  2. Tulis test cases (unit test, integration test)
  3. Validasi functionality terhadap acceptance criteria
  4. Test API endpoint dengan curl/HTTPie
  5. Test code runner dengan berbagai input
  6. Report bug dengan reproducible steps
  7. Verifikasi fix dari engineer

  Aturan:
  - Tulis test case document sebelum testing
  - Setiap bug harus ada: steps to reproduce, expected, actual
  - Jangan mark done sebelum semua test pass
  - Test edge case (empty input, invalid code, timeout)
  - Verifikasi security (sandbox isolation, no code injection)

board: coding-challenge-dev
EOF
```

---

## 4. Verifikasi Profiles

```bash
# List semua profile
hermes profile list

# Expected output:
# em          (anthropic/claude-sonnet-4-20250514)
# architect   (anthropic/claude-sonnet-4-20250514)
# backend     (anthropic/claude-sonnet-4-20250514)
# frontend    (anthropic/claude-sonnet-4-20250514)
# qa          (anthropic/claude-sonnet-4-20250514)
```

---

## 5. Install Skills

Setiap agent memiliki SKILL.md yang mendefinisikan behavior, constraints, dan deliverables.

### 5.1 Copy Skills ke Profile Directories

```bash
# Dari root project
PROJECT_DIR="/mnt/c/Users/PC/go/src/project/coding-challange"

# EM skill
cp -r "$PROJECT_DIR/hermes/skills/em/" ~/.hermes/profiles/em/skills/coding-challenge-em/

# Architect skill
cp -r "$PROJECT_DIR/hermes/skills/architect/" ~/.hermes/profiles/architect/skills/coding-challenge-architect/

# Backend skill
cp -r "$PROJECT_DIR/hermes/skills/backend/" ~/.hermes/profiles/backend/skills/coding-challenge-backend/

# Frontend skill
cp -r "$PROJECT_DIR/hermes/skills/frontend/" ~/.hermes/profiles/frontend/skills/coding-challenge-frontend/

# QA skill
cp -r "$PROJECT_DIR/hermes/skills/qa/" ~/.hermes/profiles/qa/skills/coding-challenge-qa/
```

### 5.2 Atau Install via Hermes CLI

```bash
# Jika skills sudah di-publish ke skill registry
hermes skills install coding-challenge-em --profile em
hermes skills install coding-challenge-architect --profile architect
hermes skills install coding-challenge-backend --profile backend
hermes skills install coding-challenge-frontend --profile frontend
hermes skills install coding-challenge-qa --profile qa
```

---

## 6. Auth — Login Setiap Profile

> **Penting:** OAuth token stores adalah per-profile. Login satu profile tidak cover profile lain.

```bash
# Login setiap profile ke provider
hermes auth login --profile em --provider anthropic
hermes auth login --profile architect --provider anthropic
hermes auth login --profile backend --provider anthropic
hermes auth login --profile frontend --provider anthropic
hermes auth login --profile qa --provider anthropic
```

---

## 7. Setup Gate Channel (Telegram/Slack)

EM (orchestrator) perlu mengirim notifikasi ke human. Setup Telegram:

### 7.1 Buat Telegram Bot

1. Buka Telegram, cari `@BotFather`
2. `/newbot` → ikuti instruksi → dapatkan `BOT_TOKEN`
3. Dapatkan chat ID Anda: kirim pesan ke bot, lalu akses `https://api.telegram.org/bot<TOKEN>/getUpdates`

### 7.2 Konfigurasi EM Profile

```bash
# Set environment variables di EM profile
cat >> ~/.hermes/profiles/em/.env << 'EOF'
TELEGRAM_BOT_TOKEN=your-bot-token-here
TELEGRAM_ALLOWED_USERS=your-telegram-user-id
TELEGRAM_CHAT_ID=your-chat-id
EOF
```

### 7.3 Test Delivery

```bash
# Test dari EM profile
hermes --profile em send --to telegram "Coding Challenge: setup complete, ready to start"
```

---

## 8. Setup Kanban Board

Lihat [board-setup.md](./board-setup.md) untuk detail lengkap.

```bash
# Buat board
hermes kanban boards create coding-challenge-dev

# Verifikasi
hermes kanban --board coding-challenge-dev list
```

---

## 9. Start Gateway Runtime

Gateway menjalankan dispatcher + cron. EM profile bertindak sebagai gateway.

> **WSL:** Gunakan `gateway run` (foreground), bukan `gateway start` (butuh systemd yang WSL sering tidak punya).

```bash
# Start gateway di foreground (gunakan tmux/screen)
tmux new-session -d -s hermes-gateway
tmux send-keys -t hermes-gateway "hermes --profile em gateway run" Enter

# Cek status (dari shell lain)
hermes --profile em gateway status
# Expected: running
```

---

## 10. Smoke Test — Satu Siklus Penuh

Sebelum menjalankan autonomous, test satu siklus manual:

### 10.1 Trigger EM dengan Goal

```bash
# Kirim goal ke EM
hermes --profile em chat -q "Mulai project Coding Challenge Website. Buat architecture task card untuk Architect. Set acceptance criteria: design doc dengan API contract, data model, problem schema, dan struktur direktori."
```

### 10.2 Monitor Kanban Board

```bash
# Watch board
hermes kanban --board coding-challenge-dev list
```

**Expected flow:**
1. EM buat `architecture` card → dispatch ke Architect
2. Architect ambil card → kerjakan → submit design doc → move ke `review`
3. EM review → approve → buat `backend` + `frontend` cards (paralel)
4. Backend & Frontend kerja paralel → submit ke `review`
5. QA review → tulis test case → jalankan test
6. Jika pass → mark `done` → EM verify
7. Jika fail → QA buat `bug` card → route ke engineer → fix → re-review

### 10.3 Verifikasi Post-Gate

```bash
# Pastikan card pertama post-gate berstatus `ready` (bukan `todo`)
hermes kanban --board coding-challenge-dev list --status ready
```

---

## 11. Go Live (Autonomous)

```bash
# Setelah smoke test berhasil, enable autonomous mode
hermes --profile em config set autonomous true

# EM akan otomatis:
# - Dispatch task ke specialist
# - Monitor progress
# - Gate iterasi
# - Notify human via Telegram untuk approval
```

---

## 12. Day-to-Day Operations

### Monitor Progress

```bash
# Lihat board
hermes kanban --board coding-challenge-dev list

# Lihat detail card
hermes kanban --board coding-challenge-dev show <card-id>
```

### Approve/Reject via Telegram

EM akan mengirim proposal via Telegram. Balas tanpa slash:

```
approve <slug>          # Approve task
shelve <slug>: reason   # Shelve dengan alasan
modify <slug>: change   # Minta modifikasi
reject the rest         # Reject semua sisanya
```

### Stop System

```bash
# Stop gateway
tmux kill-session -t hermes-gateway

# Atau
hermes --profile em gateway stop
```

---

## Troubleshooting

| Symptom | Cause / Fix |
|---------|-------------|
| Agent run, no card appears | Profile missing `kanban` toolset. Edit `config.yaml`, tambahkan `kanban` ke `toolsets`. |
| Card stuck in `todo` | Punya parent yang belum selesai. Jangan parent post-gate task ke triage card. |
| Proposal status set but no DM | EM tidak run `hermes send`. Status ≠ delivery. |
| `/approve` "unknown command" | Telegram reserves `/`. Balas tanpa slash: `approve <slug>`. |
| `gateway start` fails on WSL | Gunakan `gateway run` (foreground). |
| Agent can't access repo | Pastikan workspace dir benar. Set `dir` ke path repo di WSL. |
| Code runner timeout | Tambah `timeout` di `configs/app.yaml`. Default 10 detik. |
| Docker permission denied | Tambah user ke docker group: `sudo usermod -aG docker $USER` |

---

## Checklist Setup

- [ ] Hermes Agent v0.12.0+ terinstall
- [ ] Go 1.21+ terinstall
- [ ] Docker terinstall dan running
- [ ] Model provider API key diset
- [ ] 5 profiles dibuat (em, architect, backend, frontend, qa)
- [ ] Setiap profile login ke provider
- [ ] Skills terinstall untuk setiap profile
- [ ] Kanban board `coding-challenge-dev` dibuat
- [ ] Telegram bot dikonfigurasi untuk EM
- [ ] Gateway running
- [ ] Smoke test berhasil