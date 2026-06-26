# Kanban Board Setup — `coding-challenge-dev`

Kanban board adalah **task bus sentral** untuk komunikasi antar agent. Semua task, bug, dan review mengalir melalui board ini.

---

## 1. Buat Board

```bash
hermes kanban boards create coding-challenge-dev
```

## 2. Kolom (Columns)

Board memiliki kolom berikut yang merepresentasikan stage workflow:

| Kolom | Deskripsi | Siapa yang Bisa Move |
|-------|-----------|---------------------|
| `backlog` | Task baru yang belum di-assign | EM |
| `todo` | Task sudah di-assign, belum dikerjakan | EM → assign ke agent |
| `in-progress` | Agent sedang mengerjakan | Agent yang ditugaskan |
| `review` | Selesai dikerjakan, menunggu review | Agent → submit untuk review |
| `testing` | QA sedang testing | QA |
| `done` | Selesai dan verified | EM (final gate) |
| `blocked` | Terblokir, butuh input/human decision | Agent mana saja |

## 3. Setup Kolom

```bash
# Set custom columns untuk board
hermes kanban --board coding-challenge-dev columns set \
  backlog \
  todo \
  in-progress \
  review \
  testing \
  done \
  blocked
```

## 4. Task Card Format

Setiap card di board menggunakan format berikut:

### 4.1 Task Card

```yaml
title: "Implement problem loader service"
type: task
role: backend
priority: high
acceptance_criteria:
  - "Problem loader dapat membaca YAML files dari direktori problems/"
  - "Mendukung sub-direktori (easy/, medium/, hard/)"
  - "Cache problems di memory setelah load pertama"
  - "Return error yang jelas jika file tidak valid"
  - "Unit test coverage > 80%"
parents: []
workspace:
  type: dir
  path: /mnt/c/Users/PC/go/src/project/coding-challange
```

### 4.2 Bug Card

```yaml
title: "BUG: Code runner timeout tidak ter-handle"
type: bug
role: backend
priority: critical
acceptance_criteria:
  - "Timeout mengembalikan error message yang user-friendly"
  - "Container Docker di-kill setelah timeout"
  - "Temp files di-clean up setelah timeout"
parents:
  - <card-id-task-yang-related>
repro_steps:
  - "Submit code dengan infinite loop"
  - "Tunggu sampai timeout (10 detik)"
  - "Lihat response: error message tidak jelas"
workspace:
  type: dir
  path: /mnt/c/Users/PC/go/src/project/coding-challange
```

## 5. Dispatch Rules

EM (Engineering Manager) bertanggung jawab untuk dispatch card:

### 5.1 Dispatch Flow

```
backlog → todo         (EM assign card ke agent)
todo → in-progress     (Agent pick up card)
in-progress → review   (Agent submit untuk review)
review → testing       (EM approve, route ke QA)
testing → done         (QA pass)
testing → in-progress  (QA fail, route balik ke engineer dengan bug card)
done                   (Final, no more moves)
```

### 5.2 Parallel Dispatch

Backend dan Frontend task bisa di-dispatch secara paralel:

```bash
# EM dispatch backend task
hermes kanban --board coding-challenge-dev create \
  --title "Implement backend API" \
  --role backend \
  --priority high \
  --status todo

# EM dispatch frontend task (paralel)
hermes kanban --board coding-challenge-dev create \
  --title "Implement frontend templates" \
  --role frontend \
  --priority high \
  --status todo
```

## 6. Workspace Configuration

### 6.1 Persistent Workspace (untuk implementasi)

Task implementasi menggunakan `dir` workspace (persistent):

```yaml
workspace:
  type: dir
  path: /mnt/c/Users/PC/go/src/project/coding-challange
```

> **Penting:** Jangan gunakan `scratch` workspace untuk task implementasi. Scratch di-wipe antar task, menyebabkan final delivery step kehilangan artifacts.

### 6.2 Scratch Workspace (untuk research/exploration)

Task research/exploration bisa menggunakan `scratch`:

```yaml
workspace:
  type: scratch
```

## 7. Card Dependencies

### 7.1 Parent-Child Relationship

```bash
# Backend task depends on architecture task
hermes kanban --board coding-challenge-dev create \
  --title "Implement backend API" \
  --role backend \
  --parent <architecture-card-id> \
  --status todo
```

> **Gotcha:** Card dengan parent yang belum `done` akan stuck di `todo`. Pastikan parent selesai dulu, atau jangan set parent untuk post-gate tasks.

### 7.2 First Post-Gate Task

> **Penting:** First task dalam post-gate chain harus `ready` (tidak ada blocking parent). Child dari triage card yang masih open akan stuck di `todo` forever.

```bash
# WRONG: parent ke architecture card yang masih open
hermes kanban create --title "Backend" --parent <arch-card-open> --status todo

# CORRECT: no parent, status ready
hermes kanban create --title "Backend" --status ready
```

## 8. Board Monitoring

### 8.1 List All Cards

```bash
hermes kanban --board coding-challenge-dev list
```

### 8.2 Filter by Status

```bash
hermes kanban --board coding-challenge-dev list --status in-progress
hermes kanban --board coding-challenge-dev list --status review
hermes kanban --board coding-challenge-dev list --status blocked
```

### 8.3 Filter by Role

```bash
hermes kanban --board coding-challenge-dev list --role backend
hermes kanban --board coding-challenge-dev list --role qa
```

### 8.4 Show Card Detail

```bash
hermes kanban --board coding-challenge-dev show <card-id>
```

## 9. Dispatch Configuration

Set `dispatch_in_gateway: false` di Hermes config agar board dispatch bekerja dengan multi-profile:

```yaml
# ~/.hermes/config.yaml
kanban:
  dispatch_in_gateway: false
  poll_interval: 2
```

## 10. Auto-Dispatch

EM bisa enable auto-dispatch untuk otomatis route card ke agent yang sesuai:

```bash
# Enable auto-dispatch
hermes kanban --board coding-challenge-dev dispatch enable

# Set dispatch rules
hermes kanban --board coding-challenge-dev dispatch rule \
  --role architect --profile architect
hermes kanban --board coding-challenge-dev dispatch rule \
  --role backend --profile backend
hermes kanban --board coding-challenge-dev dispatch rule \
  --role frontend --profile frontend
hermes kanban --board coding-challenge-dev dispatch rule \
  --role qa --profile qa
```

Dengan auto-dispatch enabled, ketika EM membuat card dengan `role: backend`, card otomatis di-route ke profile `backend`.