# 📊 Monitoring Setup — Prometheus + Grafana

> Panduan lengkap untuk mengaktifkan dan mengkonfigurasi monitoring stack pada platform Coding Challenge.

---

## 📋 Daftar Isi

1. [Arsitektur Monitoring](#arsitektur-monitoring)
2. [Quick Start](#quick-start)
3. [Aktifkan Monitoring](#aktifkan-monitoring)
4. [Konfigurasi Prometheus](#konfigurasi-prometheus)
5. [Dashboard Grafana](#dashboard-grafana)
6. [Exporters](#exporters)
7. [Alerting Rules](#alerting-rules)
8. [Verifikasi](#verifikasi)
9. [Custom Metrics](#custom-metrics)
10. [Troubleshooting](#troubleshooting)

---

## 🏗 Arsitektur Monitoring

```
┌─────────────────────────────────────────────────────────────┐
│                    Docker Network                            │
│                                                             │
│  ┌──────────┐   ┌──────────┐   ┌──────────┐                │
│  │  Auth     │   │ Problem  │   │Execution │   ... 7x svc  │
│  │  Service  │   │ Service  │   │ Service  │                │
│  └────┬─────┘   └────┬─────┘   └────┬─────┘                │
│       │ :9101         │ :9102        │ :9103                 │
│       └───────────────┼──────────────┘                      │
│                       │                                     │
│              ┌────────▼────────┐                            │
│              │   Prometheus    │  :9090                      │
│              │  (scrape every  │                            │
│              │   15s)          │                            │
│              └────────┬────────┘                            │
│                       │                                     │
│              ┌────────▼────────┐                            │
│              │    Grafana      │  :3000                      │
│              │   Dashboards    │  admin/admin                │
│              └─────────────────┘                            │
│                                                             │
│  ┌──────────┐   ┌──────────┐   ┌──────────┐                │
│  │PostgreSQL│   │  Redis   │   │ RabbitMQ │                │
│  │Exporter  │   │ Exporter │   │ Exporter  │                │
│  │ :9187    │   │ :9121    │   │ :9419    │                │
│  └──────────┘   └──────────┘   └──────────┘                │
│                                                             │
│  ┌──────────┐                                              │
│  │   Node   │  :9100 (host metrics)                        │
│  │ Exporter │                                              │
│  └──────────┘                                              │
└─────────────────────────────────────────────────────────────┘
```

### Komponen

| Komponen | Port | Deskripsi |
|----------|------|-----------|
| **Prometheus** | `:9090` | Metrics store & query engine |
| **Grafana** | `:3000` | Visualisasi dashboard |
| **PostgreSQL Exporter** | `:9187` | Metrics PostgreSQL |
| **Redis Exporter** | `:9121` | Metrics Redis |
| **RabbitMQ Exporter** | `:9419` | Metrics RabbitMQ |
| **Node Exporter** | `:9100` | Metrics host (CPU, disk, network) |

### Application Scrape Targets

| Target | Port | Container |
|--------|------|-----------|
| auth-service | `:9101` | coding-challange-auth |
| problem-service | `:9102` | coding-challange-problem |
| execution-service | `:9103` | coding-challange-execution |
| execution-worker | `:9106` | coding-challange-worker |
| leaderboard-service | `:9104` | coding-challange-leaderboard |
| hint-service | `:9105` | coding-challange-hint |
| websocket-service | `:9107` | coding-challange-websocket |
| api-gateway | `:9100` (Nginx) | coding-challange-gateway |

---

## ⚡ Quick Start

```bash
# 1. Jalankan semua service + monitoring
docker compose --profile monitoring up -d --build

# 2. Verifikasi Prometheus
curl http://localhost:9090/-/ready

# 3. Buka Grafana
#    URL: http://localhost:3000
#    Login: admin / admin
```

---

## 🔧 Aktifkan Monitoring

### Via Docker Compose Profile

Prometheus dan Grafana sudah didefinisikan di `docker-compose.yml` di bawah profile `monitoring`. Jalankan dengan:

```bash
# Hanya aplikasi (tanpa monitoring)
docker compose up -d --build

# Aplikasi + monitoring stack
docker compose --profile monitoring up -d --build
```

### Via Kubernetes

Deployment Kubernetes sudah mencakup monitoring stack:

```bash
kubectl apply -f deployments/kubernetes/
```

---

## ⚙️ Konfigurasi Prometheus

File konfigurasi: `deployments/prometheus/prometheus.yml`

### Global Settings

```yaml
global:
  scrape_interval: 15s      # Default scrape interval
  evaluation_interval: 15s  # Default rule evaluation
  scrape_timeout: 10s       # Default scrape timeout
```

### Menambahkan Service Baru

Untuk menambahkan service baru ke Prometheus:

```yaml
- job_name: 'nama-service'
  static_configs:
    - targets: ['nama-service:PORT']
  metrics_path: '/metrics'
  relabel_configs:
    - source_labels: [__address__]
      target_label: service
      replacement: 'nama'
```

### Alertmanager

Alertmanager dikonfigurasi di:

```bash
deployments/prometheus/alerts/*.yml
```

Contoh aturan:

```yaml
- alert: ServiceDown
  expr: up{job!="prometheus"} == 0
  for: 1m
  annotations:
    summary: "Service {{ $labels.job }} is down"
```

---

## 📈 Dashboard Grafana

### Available Dashboards

| Dashboard | UID | File |
|-----------|-----|------|
| **Services Overview** | `coding-challange-services` | `deployments/grafana/dashboards/services.json` |
| **PostgreSQL Database** | `coding-challange-postgres` | `deployments/grafana/dashboards/database.json` |

### Services Overview

Dashboard ini mencakup panel untuk semua microservices Go:

| Panel | Metrics Source | Deskripsi |
|-------|---------------|-----------|
| **CPU Usage** | `rate(process_cpu_seconds_total[1m])` | CPU usage per service |
| **Memory Usage** | `go_memstats_alloc_bytes` | Heap memory alokasi |
| **Goroutines** | `go_goroutines` | Concurrency level |
| **Open FDs** | `process_open_fds` | File descriptor usage |
| **Request Rate** | `rate(gin_requests_total[1m])` | HTTP request throughput |
| **Error Rate** | `rate(gin_requests_total{status=~"5.."}[1m])` | HTTP 5xx errors |
| **Latency p95** | `histogram_quantile(0.95, ...)` | P95 response time |
| **Latency p99** | `histogram_quantile(0.99, ...)` | P99 response time |
| **GC Duration** | `rate(go_gc_duration_seconds_sum[1m])` | Garbage collection overhead |
| **Heap In Use** | `go_memstats_heap_inuse_bytes` | Heap in use |
| **Service Health** | `up{}` | Up/down status |
| **Queue Depth** | `rabbitmq_queue_messages{}` | Execution queue depth |

### PostgreSQL Dashboard

| Panel | Metrics Source | Deskripsi |
|-------|---------------|-----------|
| **Active Connections** | `pg_stat_database_numbackends` | Koneksi aktif |
| **Connection Pool %** | Pool usage percentage | Utilisasi pool |
| **Idle vs Active** | `pg_stat_activity_count` | Breakdown idle/active |
| **Cache Hit Ratio** | `blks_hit / (blks_hit + blks_read)` | Cache efficiency |
| **Cache Reads** | `blks_hit` vs `blks_read` | Read pattern |
| **Query Throughput** | `tup_fetched/inserted/updated/deleted` | Tuples per second |
| **Transactions** | `xact_commit / xact_rollback` | TPS + error rate |
| **Query Duration** | `pg_stat_activity_max_tx_duration` | Slowest query |
| **Deadlocks** | `rate(pg_stat_database_deadlocks)` | Deadlock frequency |
| **Database Size** | `pg_database_size_bytes` | Ukuran database |
| **Buffers** | Checkpoint/backend writes | Write patterns |
| **Checkpoints** | Timed vs requested | Checkpoint health |

### Import Dashboard Manual

Jika Grafana tidak auto-provisioning:

1. Buka Grafana → `+` → `Import`
2. Upload file JSON dari `deployments/grafana/dashboards/`
3. Pilih datasource `Prometheus`
4. Klik `Import`

### Auto-Provisioning

Untuk auto-load dashboards saat Grafana start, buat file provisioning:

```yaml
# deployments/grafana/provisioning/dashboards/default.yaml
apiVersion: 1

providers:
  - name: 'Coding Challenge'
    orgId: 1
    folder: ''
    type: file
    disableDeletion: false
    editable: true
    options:
      path: /etc/grafana/dashboards
```

Dan mount volume dashboard JSON:

```yaml
# Di docker-compose.yml → grafana.volumes:
- ./deployments/grafana/dashboards:/etc/grafana/dashboards:ro
- ./deployments/grafana/provisioning:/etc/grafana/provisioning:ro
```

---

## 🔌 Exporters

### PostgreSQL Exporter

Belum termasuk di docker-compose.yml. Tambahkan:

```yaml
postgres-exporter:
  image: prometheuscommunity/postgres-exporter:latest
  container_name: coding-challange-pg-exporter
  restart: unless-stopped
  environment:
    DATA_SOURCE_NAME: "postgresql://postgres:postgres@postgres:5432/coding_challange?sslmode=disable"
  ports:
    - "9187:9187"
  networks:
    - coding-challange
  profiles:
    - monitoring
```

### Redis Exporter

```yaml
redis-exporter:
  image: oliver006/redis_exporter:latest
  container_name: coding-challange-redis-exporter
  restart: unless-stopped
  environment:
    REDIS_ADDR: redis:6379
  ports:
    - "9121:9121"
  networks:
    - coding-challange
  profiles:
    - monitoring
```

### RabbitMQ Exporter

```yaml
rabbitmq-exporter:
  image: kbudde/rabbitmq-exporter:latest
  container_name: coding-challange-rmq-exporter
  restart: unless-stopped
  environment:
    RABBIT_URL: http://rabbitmq:15672
    RABBIT_USER: guest
    RABBIT_PASSWORD: guest
  ports:
    - "9419:9419"
  networks:
    - coding-challange
  profiles:
    - monitoring
```

### Node Exporter

```yaml
node-exporter:
  image: prom/node-exporter:latest
  container_name: coding-challange-node-exporter
  restart: unless-stopped
  ports:
    - "9100:9100"
  networks:
    - coding-challange
  profiles:
    - monitoring
```

---

## 🚨 Alerting Rules

Alerting rules ada di `deployments/prometheus/alerts/service-down.yml`:

| Alert | Condition | Severity |
|-------|-----------|----------|
| **ServiceDown** | `up == 0` for 1m | critical |
| **HighErrorRate** | error rate > 5% for 5m | warning |
| **HighLatency** | p95 > 2s for 5m | warning |
| **HighMemory** | memory > 500MB for 5m | warning |
| **QueueGrowing** | queue depth > 100 for 5m | warning |

---

## ✅ Verifikasi

### 1. Cek Prometheus Target

```bash
curl http://localhost:9090/api/v1/targets | jq '.data.activeTargets[] | {job: .labels.job, health: .health, lastScrape: .lastScrape}'
```

### 2. Query Metrics

```bash
# Cek semua service up
curl 'http://localhost:9090/api/v1/query?query=up'

# Cek CPU usage
curl 'http://localhost:9090/api/v1/query?query=rate(process_cpu_seconds_total[1m])'

# Cek request rate
curl 'http://localhost:9090/api/v1/query?query=rate(gin_requests_total[1m])'
```

### 3. Cek Grafana

```bash
# Buka browser
http://localhost:3000

# Login: admin / admin

# Cek datasource
Configuration → Data Sources → Prometheus (Status: OK)
```

### 4. Cek Health Endpoint Service

```bash
# Cek health endpoint masing-masing service
for port in 8081 8082 8083 8084 8085 9107; do
  curl -s http://localhost:$port/health | jq .status
done
```

---

## 📝 Custom Metrics

### Go Service Metrics (prometheus/client_golang)

Setiap service Go secara otomatis mengexpose default metrics dari `prometheus/client_golang`:

| Metric | Type | Deskripsi |
|--------|------|-----------|
| `go_goroutines` | Gauge | Jumlah goroutines |
| `go_memstats_alloc_bytes` | Gauge | Alokasi heap |
| `go_memstats_heap_inuse_bytes` | Gauge | Heap in use |
| `go_gc_duration_seconds` | Summary | Durasi GC |
| `process_cpu_seconds_total` | Counter | CPU time total |
| `process_resident_memory_bytes` | Gauge | Resident memory |
| `process_open_fds` | Gauge | File descriptors |

### Gin Metrics

Untuk mengaktifkan metrics HTTP di Gin:

```go
import "github.com/prometheus/client_golang/prometheus/promhttp"

// Register metrics endpoint
router.GET("/metrics", gin.WrapH(promhttp.Handler()))

// Middleware untuk request metrics
import "github.com/gin-contrib/pprof"
// atau custom middleware yang mengexpose:
//   gin_requests_total (method, path, status_code)
//   gin_request_duration_seconds (histogram)
```

---

## 🔍 Troubleshooting

### 1. Target Down / No Data

```bash
# Cek service HTTP health
curl http://localhost:8081/health
curl http://localhost:9101/metrics  # Langsung ke port metrics

# Cek container logs
docker logs coding-challange-auth
docker logs coding-challange-prometheus
```

### 2. Grafana Tidak Bisa Connect ke Prometheus

1. Buka Grafana → Configuration → Data Sources
2. Cek URL: harusnya `http://prometheus:9090` (service name, bukan localhost)
3. Test connection

### 3. Postgres Exporter Gagal Scrape

```bash
# Pastikan exporter bisa connect ke postgres
docker exec coding-challange-pg-exporter sh -c "wget -qO- http://localhost:9187/metrics | head -20"
```

### 4. Volume Permission Issues

```bash
# Fix permission di Linux
chmod 755 deployments/prometheus/prometheus.yml

# Untuk Grafana provisioning
chmod 755 deployments/grafana/dashboards/*.json
```

### 5. Prometheus Retention

Retention diatur di docker-compose command:

```yaml
command:
  - '--storage.tsdb.retention.time=7d'  # Simpan 7 hari
```

Untuk mengubah, edit dan restart:

```bash
docker compose --profile monitoring up -d prometheus
```

---

## 📚 Referensi

- [Prometheus Documentation](https://prometheus.io/docs/introduction/overview/)
- [Grafana Documentation](https://grafana.com/docs/)
- [PostgreSQL Exporter](https://github.com/prometheus-community/postgres_exporter)
- [Redis Exporter](https://github.com/oliver006/redis_exporter)
- [Node Exporter](https://github.com/prometheus/node_exporter)
- [Go client for Prometheus](https://pkg.go.dev/github.com/prometheus/client_golang/prometheus)
