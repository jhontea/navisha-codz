1|# Deployment Guide — Coding Challenge Website
2|
3|## Table of Contents
4|1. [Prerequisites](#prerequisites)
5|2. [Local Development](#local-development)
6|3. [Docker Deployment](#docker-deployment)
7|4. [Production Deployment](#production-deployment)
8|5. [Configuration](#configuration)
9|6. [Monitoring & Logging](#monitoring--logging)
10|7. [Troubleshooting](#troubleshooting)
11|
12|---
13|
14|## Prerequisites
15|
16|- **Go 1.21+** — [Download](https://go.dev/dl/)
17|- **Docker 24.0+** — [Download](https://docs.docker.com/get-docker/)
18|- **Docker Compose v2+** (optional, for local development)
19|
20|---
21|
22|## Local Development
23|
24|### 1. Clone and Build
25|
26|```bash
27|cd coding-challange
28|go build -o server cmd/server/main.go
29|```
30|
31|### 2. Run
32|
33|```bash
34|# Set environment variables (optional, defaults work)
35|export PORT=8080
36|export LOG_LEVEL=debug
37|export PROBLEMS_DIR=./problems
38|
39|# Start server
40|./server
41|```
42|
43|Server starts at `http://localhost:9100`
44|
45|### 3. Run Tests
46|
47|```bash
48|# All tests
49|go test ./... -count=1
50|
51|# Verbose
52|go test ./... -v -count=1
53|
54|# Specific package
55|go test ./internal/handler/... -v
56|```
57|
58|### 4. Build Sandbox Image (Optional)
59|
60|For production-like sandboxed execution:
61|
62|```bash
63|# Build the sandbox image once
64|docker build -t coding-challenge-sandbox:latest -f Dockerfile.sandbox .
65|
66|# Run server with Docker sandbox
67|DISABLE_LOCAL_FALLBACK=true ./server
68|```
69|
70|---
71|
72|## Docker Deployment
73|
74|### Quick Start (Docker Compose)
75|
76|```yaml
77|# docker-compose.yml
78|version: '3.8'
79|
80|services:
81|  coding-challenge:
82|    build:
83|      context: .
84|      dockerfile: Dockerfile
85|    ports:
86|      - "8080:9100"
87|    environment:
88|      - PORT=8080
89|      - LOG_LEVEL=info
90|      - PROBLEMS_DIR=/app/problems
91|      - SANDBOX_TIMEOUT=10
92|      - MAX_MEMORY_MB=256
93|      - DISABLE_LOCAL_FALLBACK=true
94|    volumes:
95|      - ./problems:/app/problems:ro
96|      - /var/run/docker.sock:/var/run/docker.sock
97|    restart: unless-stopped
98|    deploy:
99|      resources:
100|        limits:
101|          cpus: '2.0'
102|          memory: 1G
103|```
104|
105|```bash
106|docker compose up -d --build
107|```
108|
109|### Manual Docker Build
110|
111|```bash
112|# Build app image
113|docker build -t coding-challenge:latest .
114|
115|# Build sandbox image
116|docker build -t coding-challenge-sandbox:latest -f Dockerfile.sandbox .
117|
118|# Run
119|docker run -d \
120|  -p 8080:9100 \
121|  -v $(pwd)/problems:/app/problems:ro \
122|  -v /var/run/docker.sock:/var/run/docker.sock \
123|  -e DISABLE_LOCAL_FALLBACK=true \
124|  coding-challenge:latest
125|```
126|
127|### Dockerfile
128|
129|```dockerfile
130|FROM golang:1.22-alpine AS builder
131|WORKDIR /app
132|COPY go.mod go.sum ./
133|RUN go mod download
134|COPY . .
135|RUN CGO_ENABLED=0 go build -o server cmd/server/main.go
136|
137|FROM alpine:3.19
138|WORKDIR /app
139|COPY --from=builder /app/server .
140|COPY --from=builder /app/problems ./problems
141|COPY --from=builder /app/web ./web
142|RUN adduser -D -u 1000 appuser
143|USER appuser
144|EXPOSE 8080
145|CMD ["./server"]
146|```
147|
148|---
149|
150|## Production Deployment
151|
152|### Kubernetes
153|
154|```yaml
155|apiVersion: apps/v1
156|kind: Deployment
157|metadata:
158|  name: coding-challenge
159|spec:
160|  replicas: 2
161|  selector:
162|    matchLabels:
163|      app: coding-challenge
164|  template:
165|    metadata:
166|      labels:
167|        app: coding-challenge
168|    spec:
169|      containers:
170|      - name: coding-challenge
171|        image: coding-challenge:latest
172|        ports:
173|        - containerPort: 8080
174|        env:
175|        - name: PORT
176|          value: "8080"
177|        - name: LOG_LEVEL
178|          value: "info"
179|        - name: DISABLE_LOCAL_FALLBACK
180|          value: "true"
181|        resources:
182|          requests:
183|            cpu: 250m
184|            memory: 256Mi
185|          limits:
186|            cpu: 500m
187|            memory:512Mi
188|        livenessProbe:
189|          httpGet:
190|            path: /health
191|            port: 8080
192|          initialDelaySeconds: 10
193|          periodSeconds: 30
194|        readinessProbe:
195|          httpGet:
196|            path: /health
197|            port: 8080
198|          initialDelaySeconds: 5
199|          periodSeconds: 10
200|---
201|apiVersion: v1
202|kind: Service
203|metadata:
204|  name: coding-challenge
205|spec:
206|  selector:
207|    app: coding-challenge
208|  ports:
209|  - port: 80
210|    targetPort: 8080
211|  type: ClusterIP
212|```
213|
214|### Reverse Proxy (nginx)
215|
216|```nginx
217|server {
218|    listen 443 ssl http2;
219|    server_name coding-challenge.example.com;
220|
221|    ssl_certificate /etc/ssl/certs/coding-challenge.crt;
222|    ssl_certificate_key /etc/ssl/private/coding-challenge.key;
223|
224|    # Security headers (in addition to app headers)
225|    add_header X-Content-Type-Options nosniff always;
226|    add_header X-Frame-Options DENY always;
227|
228|    # Rate limiting
229|    limit_req_zone $binary_remote_addr zone=run:10m rate=10r/m;
230|    limit_req_zone $binary_remote_addr zone=api:10m rate=60r/m;
231|
232|    location /api/problems/ {
233|        limit_req zone=api burst=10 nodelay;
234|        proxy_pass http://localhost:9100;
235|        proxy_set_header Host $host;
236|        proxy_set_header X-Real-IP $remote_addr;
237|        proxy_set_header X-Forwarded-For $proxy_add_x_forwarded_for;
238|        proxy_set_header X-Forwarded-Proto $scheme;
239|    }
240|
241|    location ~ /run$ {
242|        limit_req zone=run burst=5 nodelay;
243|        proxy_pass http://localhost:9100;
244|        proxy_set_header Host $host;
245|        proxy_set_header X-Real-IP $remote_addr;
246|    }
247|
248|    location / {
249|        proxy_pass http://localhost:9100;
250|        proxy_set_header Host $host;
251|    }
252|}
253|```
254|
255|---
256|
257|## Configuration
258|
259|### Environment Variables
260|
261|| Variable | Default | Description |
262||----------|---------|-------------|
263|| `PORT` | `8080` | HTTP server port |
264|| `PROBLEMS_DIR` | `./problems` | Path to problem YAML files |
265|| `SANDBOX_TIMEOUT` | `10` | Execution timeout in seconds |
266|| `MAX_MEMORY_MB` | `256` | Sandbox memory limit |
267|| `LOG_LEVEL` | `info` | Log verbosity: debug, info, warn, error |
268|| `DISABLE_LOCAL_FALLBACK` | `` | Set to any value to fail if Docker unavailable |
269|
270|### Configuration File (configs/app.yaml)
271|
272|```yaml
273|server:
274|  port: 8080
275|  read_timeout: 15s
276|  write_timeout: 30s
277|  max_body_size: 65536  # 64KB
278|
279|sandbox:
280|  timeout: 10s
281|  max_memory_mb: 256
282|  disable_local_fallback: true
283|
284|cache:
285|  ttl: 5m
286|  cleanup_interval: 30s
287|
288|rate_limit:
289|  run: 10/min
290|  list: 60/min
291|  validate: 30/min
292|```
293|
294|---
295|
296|## Monitoring & Logging
297|
298|### Health Check
299|
300|```bash
301|curl http://localhost:9100/health
302|```
303|
304|Response:
305|```json
306|{
307|  "data": {
308|    "status": "ok",
309|    "version": "1.1.0"
310|  },
311|  "meta": {
312|    "request_id": "req-1700000000000000000",
313|    "timestamp": "2024-01-01T00:00:00Z"
314|  }
315|}
316|```
317|
318|### Structured Logs
319|
320|```json
321|{"level":"info","msg":"Server starting","addr":":9100","time":"2024-01-01T00:00:00Z"}
322|{"level":"info","msg":"Loaded problems","count":10,"dir":"./problems","time":"2024-01-01T00:00:00Z"}
323|{"level":"warn","msg":"Rate limit exceeded","ip":"192.168.1.1","endpoint":"/api/problems/two-sum/run","time":"2024-01-01T00:00:01Z"}
324|```
325|
326|### Metrics (Optional Prometheus)
327|
328|Add to `cmd/server/main.go`:
329|```go
330|import "github.com/prometheus/client_golang/prometheus/promhttp"
331|
332|router.GET("/metrics", gin.WrapH(promhttp.Handler()))
333|```
334|
335|---
336|
337|## Troubleshooting
338|
339|### Server won't start
340|
341|```bash
342|# Check port availability
343|netstat -an | grep 8080
344|
345|# Check Go version
346|go version
347|
348|# Build with verbose output
349|go build -v ./...
350|```
351|
352|### Docker sandbox not working
353|
354|```bash
355|# Test Docker connectivity
356|docker info
357|
358|# Build sandbox image manually
359|docker build -t coding-challenge-sandbox:latest -f Dockerfile.sandbox .
360|
361|# Test sandbox manually
362|echo 'package main; import "fmt"; func main() { fmt.Println("hello") }' > /tmp/test.go
363|docker run --rm -v /tmp:/app:ro coding-challenge-sandbox:latest sh -c "cd /app && go run test.go"
364|```
365|
366|### Tests failing
367|
368|```bash
369|# Run with race detector
370|go test -race ./... -count=1
371|
372|# Run specific test
373|go test -run TestName ./internal/handler/... -v
374|
375|# Check for data races during load
376|go test -race -run TestConcurrency ./internal/repository/... -v
377|```
378|
379|### High memory usage
380|
381|- Reduce `MAX_MEMORY_MB`
382|- Enable runner concurrency limits
383|- Add request queue with backpressure
384|
385|### Rate limiting issues
386|
387|- Adjust rate limits in handler
388|- Use Redis for distributed rate limiting across instances
389|- Add client-side debouncing
390|
391|---
392|
393|## Security Considerations
394|
395|1. **Always run behind HTTPS in production**
396|2. **Use Docker socket proxy** (Tecnativa/docker-socket-proxy)
397|3. **Enable AppArmor/SELinux** for the server process
398|4. **Keep Docker updated** to patch container escape CVEs
399|5. **Monitor audit logs** for abnormal execution patterns
400|6. **Use network segmentation** for the sandbox subnet
401|7. **Set up alerting** for sandbox failures (502/503 errors)
402|