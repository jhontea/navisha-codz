# Hermes Multi-Agent Setup

Setup untuk **multiple agent Hermes** yang bekerja secara paralel untuk membangun aplikasi **Coding Challenge Website** (bank soal algoritma & data structure berbasis Go).

## Arsitektur Multi-Agent

```
                    ┌─────────────────────────┐
                    │  Engineering Manager     │
                    │   (Orchestrator/Leader)  │
                    └────────────┬────────────┘
                                 │
                    ┌────────────▼────────────┐
                    │    Hermes Kanban Board   │
                    │    (Task Bus / Hub)      │
                    └──┬──────┬──────┬────────┘
                       │      │      │
              ┌────────▼┐  ┌──▼───┐  ┌▼─────────┐
              │Architect│  │Backend│  │Frontend  │
              │         │  │       │  │          │
              └─────────┘  └───┬───┘  └──────────┘
                               │
                          ┌────▼────┐
                          │   QA    │
                          │ Review  │
                          └─────────┘
```

## Agent Roles

| Agent | Profile Name | Role | Tanggung Jawab |
|-------|-------------|------|----------------|
| Engineering Manager | `em` | Orkestrator | Memecah task, dispatch ke specialist, review deliverable, gate iterasi |
| Architect | `architect` | System Designer | Merancang API contract, data model, arsitektur sistem, problem schema |
| Backend Engineer | `backend` | Go Developer | Implementasi Go backend, API endpoint, code runner, problem loading |
| Frontend Engineer | `frontend` | UI Developer | Implementasi template, code editor, problem display, result UI |
| QA Engineer | `qa` | Tester/Reviewer | Menulis test case, review code, validasi fungsi, report bug |

## Struktur Direktori

```
hermes/
├── README.md                    # File ini
├── runbook.md                   # Panduan setup step-by-step
├── board-setup.md               # Setup Kanban board
├── orchestration.yaml           # Definisi task, acceptance criteria, aturan iterasi
├── profiles/                    # Konfigurasi profile per agent
│   ├── em/
│   │   └── config.yaml
│   ├── architect/
│   │   └── config.yaml
│   ├── backend/
│   │   └── config.yaml
│   ├── frontend/
│   │   └── config.yaml
│   └── qa/
│       └── config.yaml
└── skills/                      # SKILL.md per agent
    ├── em/
    │   └── SKILL.md
    ├── architect/
    │   └── SKILL.md
    ├── backend/
    │   └── SKILL.md
    ├── frontend/
    │   └── SKILL.md
    └── qa/
        └── SKILL.md
```

## Quick Start

1. Install Hermes Agent (lihat `runbook.md` Section 1)
2. Buat semua profile: `runbook.md` Section 3
3. Install semua skill: `runbook.md` Section 5
4. Setup Kanban board: `board-setup.md`
5. Start gateway: `runbook.md` Section 9
6. Smoke test: `runbook.md` Section 10

## Workflow Iterasi

```
1. EM menerima goal → buat architecture task card
2. Architect merancang → submit design doc → EM review
3. EM buat backend + frontend task cards (paralel)
4. Backend & Frontend kerja paralel → submit untuk review
5. QA review keduanya → tulis test case → jalankan test
   ├── Pass → QA mark done → EM verify → ship
   └── Fail → QA buat bug card → route ke Backend/Frontend
6. Engineer fix → resubmit → QA re-review (iterasi)
7. EM final gate check → semua acceptance criteria terpenuhi → done
```

## Komunikasi Antar Agent

- **Kanban Board** (`coding-challenge-dev`) adalah task bus sentral
- EM membuat task card dengan role assignment
- Specialist mengambil card, mengerjakan, pindahkan melalui stage: `todo → in-progress → review → done`
- QA review work yang completed; jika ada issue, buat `bug` card yang route kembali ke engineer terkait
- Iterasi berlanjut sampai QA mark semua acceptance criteria terpenuhi
- EM melakukan final gate check

## Prasyarat

- Hermes Agent v0.12.0+ terinstall
- Go 1.21+ terinstall
- Docker terinstall (untuk code runner sandbox)
- WSL2 (jika di Windows — Hermes berjalan di WSL)
- Model provider API key (Anthropic/OpenAI/dll)

## Dokumentasi Terkait

- [Runbook Lengkap](./runbook.md) — Setup step-by-step
- [Board Setup](./board-setup.md) — Konfigurasi Kanban board
- [Orchestration Config](./orchestration.yaml) — Definisi task & workflow
- [Dokumentasi Arsitektur Aplikasi](../docs/ARCHITECTURE.md)
- [Panduan Setup Hermes](../docs/HERMES_SETUP.md)