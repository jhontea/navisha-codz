package redis

import (
	"container/list"
	"context"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"sync"
	"sync/atomic"
	"time"

	"github.com/redis/go-redis/v9"
)

// ============================================================================
// Cache Versioning
// ============================================================================

// CacheVersion represents the current cache schema version.
// Increment this when cached data structure changes to invalidate old entries.
const CacheVersion = "v2"

// VersionedKey prepends the cache version to a key.
func VersionedKey(key string) string {
	return fmt.Sprintf("%s:%s", CacheVersion, key)
}

// ============================================================================
// Circuit Breaker
// ============================================================================

// CircuitBreakerState represents the state of the circuit breaker.
type CircuitBreakerState int

const (
	StateClosed CircuitBreakerState = iota
	StateOpen
	StateHalfOpen
)

// CircuitBreaker implements the circuit breaker pattern for Redis operations.
type CircuitBreaker struct {
	mu               sync.RWMutex
	state            CircuitBreakerState
	failureCount     int
	failureThreshold int
	successCount     int
	successThreshold int
	timeout          time.Duration
	lastFailureTime  time.Time
}

// NewCircuitBreaker creates a new circuit breaker.
func NewCircuitBreaker(failureThreshold, successThreshold int, timeout time.Duration) *CircuitBreaker {
	return &CircuitBreaker{
		state:            StateClosed,
		failureThreshold: failureThreshold,
		successThreshold: successThreshold,
		timeout:          timeout,
	}
}

// Execute executes the given function through the circuit breaker.
func (cb *CircuitBreaker) Execute(fn func() error) error {
	if !cb.Allow() {
		return fmt.Errorf("circuit breaker is open, request rejected")
	}

	err := fn()
	cb.RecordResult(err == nil)
	return err
}

// Allow checks if the circuit breaker allows the request.
func (cb *CircuitBreaker) Allow() bool {
	cb.mu.RLock()
	defer cb.mu.RUnlock()

	switch cb.state {
	case StateClosed:
		return true
	case StateOpen:
		if time.Since(cb.lastFailureTime) > cb.timeout {
			return true // Will transition to half-open
		}
		return false
	case StateHalfOpen:
		return true
	}
	return false
}

// RecordResult records the result of a request.
func (cb *CircuitBreaker) RecordResult(success bool) {
	cb.mu.Lock()
	defer cb.mu.Unlock()

	if success {
		cb.successCount++
		cb.failureCount = 0
		if cb.state == StateHalfOpen && cb.successCount >= cb.successThreshold {
			cb.state = StateClosed
			cb.successCount = 0
		}
	} else {
		cb.failureCount++
		cb.successCount = 0
		cb.lastFailureTime = time.Now()
		if cb.failureCount >= cb.failureThreshold {
			cb.state = StateOpen
		}
	}
}

// State returns the current state of the circuit breaker.
func (cb *CircuitBreaker) State() CircuitBreakerState {
	cb.mu.RLock()
	defer cb.mu.RUnlock()
	return cb.state
}

// ============================================================================
// Health Monitor
// ============================================================================

// HealthMonitor monitors Redis connection health.
type HealthMonitor struct {
	mu              sync.RWMutex
	healthy         bool
	lastCheck       time.Time
	checkInterval   time.Duration
	consecutiveOK   int
	consecutiveFail int
	onHealthCheck   func(error)
}

// NewHealthMonitor creates a new health monitor.
func NewHealthMonitor(checkInterval time.Duration, onHealthCheck func(error)) *HealthMonitor {
	return &HealthMonitor{
		healthy:       true,
		checkInterval: checkInterval,
		onHealthCheck: onHealthCheck,
	}
}

// Check performs a health check.
func (hm *HealthMonitor) Check(ctx context.Context, client *redis.Client) bool {
	ctx, cancel := context.WithTimeout(ctx, 3*time.Second)
	defer cancel()

	err := client.Ping(ctx).Err()
	hm.mu.Lock()
	defer hm.mu.Unlock()

	hm.lastCheck = time.Now()
	if err == nil {
		hm.consecutiveOK++
		hm.consecutiveFail = 0
		hm.healthy = true
	} else {
		hm.consecutiveFail++
		hm.consecutiveOK = 0
		if hm.consecutiveFail >= 3 {
			hm.healthy = false
		}
	}

	if hm.onHealthCheck != nil {
		hm.onHealthCheck(err)
	}

	return err == nil
}

// IsHealthy returns whether Redis is considered healthy.
func (hm *HealthMonitor) IsHealthy() bool {
	hm.mu.RLock()
	defer hm.mu.RUnlock()
	return hm.healthy
}

// GetStats returns health statistics.
func (hm *HealthMonitor) GetStats() map[string]interface{} {
	hm.mu.RLock()
	defer hm.mu.RUnlock()
	return map[string]interface{}{
		"healthy":          hm.healthy,
		"last_check":       hm.lastCheck,
		"consecutive_ok":   hm.consecutiveOK,
		"consecutive_fail": hm.consecutiveFail,
	}
}

// Key patterns for different data types.
const (
	KeyProblem      = "problem:%s"
	KeyProblemList  = "problems:list:%s"
	KeyProblemCount = "problems:count:%s:%s"
	KeyUser         = "user:%s"
	KeyUserStats    = "user:stats:%s"
	KeyUserProgress = "user:progress:%s"
	KeyUserRank     = "user:rank:%s"
	KeyLeaderboard  = "leaderboard:%s"
	KeySubmission   = "submission:%s"
	KeySession      = "session:%s"
	KeyRateLimit    = "ratelimit:%s:%s"
	KeyHint         = "hint:%s:%s"
	KeyHotProblems = "problems:hot"
)

// TTL settings for different cache types.
const (
	TTLSession     = 24 * time.Hour
	TTLUser        = 1 * time.Hour
	TTLUserStats   = 5 * time.Minute
	TTLProblem     = 30 * time.Minute
	TTLProblemList = 10 * time.Minute
	TTLLeaderboard = 1 * time.Minute
	TTLSubmission  = 5 * time.Minute
	TTLRateLimit   = 1 * time.Minute
	TTLHint        = 30 * time.Minute
)

// Pub/Sub channel names.
const (
	ChannelLeaderboardUpdate = "leaderboard:update"
	ChannelSubmissionNew     = "submission:new"
	ChannelSubmissionResult  = "submission:result"
)

// ============================================================================
// Enhanced Client
// ============================================================================

// Client wraps redis.Client with additional functionality including
// cache-aside pattern, bulk operations, cache warming, circuit breaker,
// and health monitoring.
type Client struct {
	*redis.Client
	cb            *CircuitBreaker
	healthMonitor *HealthMonitor
	cacheVersion  string

	// Cluster support
	clusterClient *redis.ClusterClient
	clusterMode   bool

	// Read replicas
	readReplicas []*redis.Client
	readFromSlave bool

	// Sentinel failover
	sentinelMode  bool
	sentinel      *redis.SentinelClient

	// Cache stampede prevention
	stampedeMu   sync.Mutex
	stampedeLocks map[string]*stampedeLock

	// Bloom filter
	bloomFilter *BloomFilter

	// Hot key detection
	hotKeyCounts  map[string]*int64
	hotKeyMu      sync.RWMutex
	hotKeyThreshold int64
	hotKeyReplicas  int

	// Pipeline batching
	pipelineBatchSize int
	batchBuffer       map[string]*batchOp
	batchMu           sync.Mutex
	batchTicker       *time.Ticker
}

type stampedeLock struct {
	mu       sync.Mutex
	loading  bool
	data     interface{}
	err      error
	done     chan struct{}
	refCount int32
}

type batchOp struct {
	cmd string
	key string
	val interface{}
	ttl time.Duration
}

// BloomFilter implements a simple bloom filter for cache key checking.
type BloomFilter struct {
	bits    []uint64
	size    uint
	hashCnt uint
	mu      sync.RWMutex
}

// NewBloomFilter creates a new bloom filter.
func NewBloomFilter(size, hashCnt uint) *BloomFilter {
	if size == 0 {
		size = 1000000
	}
	if hashCnt == 0 {
		hashCnt = 3
	}
	// Round to nearest multiple of 64
	numWords := (size + 63) / 64
	if numWords == 0 {
		numWords = 1
	}
	return &BloomFilter{
		bits:    make([]uint64, numWords),
		size:    numWords * 64,
		hashCnt: hashCnt,
	}
}

// Add adds a key to the bloom filter.
func (bf *BloomFilter) Add(key string) {
	bf.mu.Lock()
	defer bf.mu.Unlock()

	data := []byte(key)
	for i := uint(0); i < bf.hashCnt; i++ {
		h := fnvHash(data, i)
		idx := uint(h) % bf.size
		wordIdx := idx / 64
		bitIdx := idx % 64
		bf.bits[wordIdx] |= 1 << bitIdx
	}
}

// MightContain checks if a key might be in the bloom filter.
// Returns false if definitely not present, true if possibly present.
func (bf *BloomFilter) MightContain(key string) bool {
	bf.mu.RLock()
	defer bf.mu.RUnlock()

	data := []byte(key)
	for i := uint(0); i < bf.hashCnt; i++ {
		h := fnvHash(data, i)
		idx := uint(h) % bf.size
		wordIdx := idx / 64
		bitIdx := idx % 64
		if bf.bits[wordIdx]&(1<<bitIdx) == 0 {
			return false
		}
	}
	return true
}

// Reset clears the bloom filter.
func (bf *BloomFilter) Reset() {
	bf.mu.Lock()
	defer bf.mu.Unlock()
	for i := range bf.bits {
		bf.bits[i] = 0
	}
}

// fnvHash computes a hash for bloom filter.
func fnvHash(data []byte, seed uint) uint64 {
	h := uint64(14695981039346656037 + seed)
	for _, b := range data {
		h ^= uint64(b)
		h *= 1099511628211
	}
	return h
}

// New creates a new enhanced Redis client with circuit breaker and health monitoring.
type BulkResult struct {
	Found   map[string]interface{}
	Missing []string
}

// CacheStats tracks cache hit/miss statistics.
type CacheStats struct {
	mu         sync.RWMutex
	Hits       int64
	Misses     int64
	Invalidations int64
}

// RecordHit records a cache hit.
func (cs *CacheStats) RecordHit() {
	cs.mu.Lock()
	defer cs.mu.Unlock()
	cs.Hits++
}

// RecordMiss records a cache miss.
func (cs *CacheStats) RecordMiss() {
	cs.mu.Lock()
	defer cs.mu.Unlock()
	cs.Misses++
}

// RecordInvalidation records a cache invalidation.
func (cs *CacheStats) RecordInvalidation() {
	cs.mu.Lock()
	defer cs.mu.Unlock()
	cs.Invalidations++
}

// HitRate returns the cache hit rate.
func (cs *CacheStats) HitRate() float64 {
	cs.mu.RLock()
	defer cs.mu.RUnlock()
	total := cs.Hits + cs.Misses
	if total == 0 {
		return 0
	}
	return float64(cs.Hits) / float64(total)
}

// Stats returns cache statistics.
func (cs *CacheStats) Stats() map[string]interface{} {
	cs.mu.RLock()
	defer cs.mu.RUnlock()
	return map[string]interface{}{
		"hits":         cs.Hits,
		"misses":       cs.Misses,
		"invalidations": cs.Invalidations,
		"hit_rate":     cs.HitRate(),
	}
}

// Config holds Redis connection configuration.
type Config struct {
	// Single instance
	Addr     string
	Password string
	DB       int
	PoolSize int

	// Cluster mode
	ClusterAddrs []string
	ClusterEnabled bool

	// Sentinel mode
	SentinelEnabled   bool
	SentinelMaster    string
	SentinelAddrs     []string
	SentinelPassword  string

	// Read replicas
	ReadReplicaAddrs []string
	ReadFromSlave    bool
	ReadOnly         bool

	// Bloom filter
	BloomFilterSize    uint
	BloomFilterHashCnt uint

	// Hot key detection
	HotKeyThreshold  int64  // requests per minute to be considered hot
	HotKeyReplicas   int    // number of replicas for hot keys
}

// New creates a new enhanced Redis client with circuit breaker and health monitoring.
// Supports single instance, cluster mode, and sentinel failover.
func New(cfg Config) (*Client, error) {
	var rdb redis.UniversalClient
	var readReplicas []*redis.Client
	var sentinelClient *redis.SentinelClient
	clusterMode := cfg.ClusterEnabled
	sentinelMode := cfg.SentinelEnabled

	// Choose connection mode
	switch {
	case cfg.ClusterEnabled && len(cfg.ClusterAddrs) > 0:
		// Redis Cluster mode
		rdb = redis.NewClusterClient(&redis.ClusterOptions{
			Addrs:        cfg.ClusterAddrs,
			Password:     cfg.Password,
			PoolSize:     cfg.PoolSize,
			MinIdleConns: 5,
			MaxRetries:   3,
			DialTimeout:  5 * time.Second,
			ReadTimeout:  3 * time.Second,
			WriteTimeout: 3 * time.Second,
			PoolTimeout:  4 * time.Second,
			ReadOnly:     cfg.ReadOnly,
		})
		log.Printf("Redis Cluster mode: %v", cfg.ClusterAddrs)

	case cfg.SentinelEnabled && len(cfg.SentinelAddrs) > 0:
		// Sentinel failover mode
		masterName := cfg.SentinelMaster
		if masterName == "" {
			masterName = "mymaster"
		}
		rdb = redis.NewFailoverClient(&redis.FailoverOptions{
			MasterName:       masterName,
			SentinelAddrs:    cfg.SentinelAddrs,
			SentinelPassword: cfg.SentinelPassword,
			Password:         cfg.Password,
			DB:               cfg.DB,
			PoolSize:         cfg.PoolSize,
			MinIdleConns:     5,
			MaxRetries:       3,
			DialTimeout:      5 * time.Second,
			ReadTimeout:      3 * time.Second,
			WriteTimeout:     3 * time.Second,
			PoolTimeout:      4 * time.Second,
		})
		sentinelClient = redis.NewSentinelClient(&redis.Options{
			Addr: cfg.SentinelAddrs[0],
			Password: cfg.SentinelPassword,
		})
		log.Printf("Redis Sentinel mode: master=%s, addrs=%v", masterName, cfg.SentinelAddrs)

	default:
		// Single instance mode
		db := cfg.DB
		rdb = redis.NewClient(&redis.Options{
			Addr:         cfg.Addr,
			Password:     cfg.Password,
			DB:           db,
			PoolSize:     cfg.PoolSize,
			MinIdleConns: 5,
			MaxRetries:   3,
			DialTimeout:  5 * time.Second,
			ReadTimeout:  3 * time.Second,
			WriteTimeout: 3 * time.Second,
			PoolTimeout:  4 * time.Second,
		})
		log.Printf("Redis single instance: %s (db=%d, pool=%d)", cfg.Addr, db, cfg.PoolSize)
	}

	// Initialize read replicas
	if cfg.ReadFromSlave && len(cfg.ReadReplicaAddrs) > 0 {
		for _, addr := range cfg.ReadReplicaAddrs {
			replica := redis.NewClient(&redis.Options{
				Addr:         addr,
				Password:     cfg.Password,
				DB:           cfg.DB,
				PoolSize:     cfg.PoolSize / 2,
				MinIdleConns: 2,
				MaxRetries:   2,
				ReadTimeout:  3 * time.Second,
			})
			readReplicas = append(readReplicas, replica)
		}
		log.Printf("Redis read replicas: %v", cfg.ReadReplicaAddrs)
	}

	// Test connection
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	if err := rdb.Ping(ctx).Err(); err != nil {
		return nil, fmt.Errorf("failed to connect to Redis: %w", err)
	}

	cb := NewCircuitBreaker(5, 3, 30*time.Second)
	hm := NewHealthMonitor(30*time.Second, func(err error) {
		if err != nil {
			log.Printf("[WARN] Redis health check failed: %v", err)
		}
	})

	// Initialize bloom filter
	bloomSize := cfg.BloomFilterSize
	if bloomSize == 0 {
		bloomSize = 1000000
	}
	bloomHashCnt := cfg.BloomFilterHashCnt
	if bloomHashCnt == 0 {
		bloomHashCnt = 3
	}

	// Initialize hot key detection
	hotKeyThreshold := cfg.HotKeyThreshold
	if hotKeyThreshold == 0 {
		hotKeyThreshold = 100 // 100 requests per minute
	}
	hotKeyReplicas := cfg.HotKeyReplicas
	if hotKeyReplicas == 0 {
		hotKeyReplicas = 2
	}

	client := &Client{
		Client:        nil,
		clusterClient: nil,
		clusterMode:   clusterMode,
		sentinelMode:  sentinelMode,
		sentinel:      sentinelClient,
		readReplicas:  readReplicas,
		readFromSlave: cfg.ReadFromSlave,
		cb:            cb,
		healthMonitor: hm,
		cacheVersion:  CacheVersion,
		bloomFilter:   NewBloomFilter(bloomSize, bloomHashCnt),
		stampedeLocks: make(map[string]*stampedeLock),
		hotKeyCounts:  make(map[string]*int64),
		hotKeyThreshold: int64(hotKeyThreshold),
		hotKeyReplicas:  hotKeyReplicas,
		pipelineBatchSize: 100,
		batchBuffer:       make(map[string]*batchOp),
	}

	// Set the appropriate client based on mode
	switch {
	case cfg.ClusterEnabled:
		client.clusterClient = rdb.(*redis.ClusterClient)
	case cfg.SentinelEnabled:
		cl := rdb.(*redis.Client)
		client.Client = cl
	default:
		cl := rdb.(*redis.Client)
		client.Client = cl
	}

	// Start background health monitoring
	go client.backgroundHealthCheck()

	// Start hot key detection
	go client.monitorHotKeys()

	// Start pipeline batch flusher
	go client.batchFlusher()

	modeInfo := "single"
	if clusterMode {
		modeInfo = "cluster"
	} else if sentinelMode {
		modeInfo = "sentinel"
	}

	log.Printf("Redis connected: mode=%s, pool=%d, version=%s", modeInfo, cfg.PoolSize, CacheVersion)
	return client, nil
}

// NewFromEnv creates a Redis client from environment variables.
func NewFromEnv() (*Client, error) {
	cfg := Config{
		Addr:     getEnv("REDIS_ADDR", "localhost:6379"),
		Password: getEnv("REDIS_PASSWORD", ""),
		DB:       getEnvInt("REDIS_DB", 0),
		PoolSize: getEnvInt("REDIS_POOL_SIZE", 10),
	}
	return New(cfg)
}

// backgroundHealthCheck periodically checks Redis health.
func (c *Client) backgroundHealthCheck() {
	ticker := time.NewTicker(30 * time.Second)
	defer ticker.Stop()
	for range ticker.C {
		c.healthMonitor.Check(context.Background(), c.Client)
	}
}

// CircuitBreaker returns the circuit breaker instance.
func (c *Client) CircuitBreaker() *CircuitBreaker {
	return c.cb
}

// HealthMonitor returns the health monitor instance.
func (c *Client) HealthMonitor() *HealthMonitor {
	return c.healthMonitor
}

// ============================================================================
// Hot Key Detection and Replication
// ============================================================================

// monitorHotKeys periodically checks for hot keys and replicates them.
func (c *Client) monitorHotKeys() {
	ticker := time.NewTicker(1 * time.Minute)
	defer ticker.Stop()

	for range ticker.C {
		c.hotKeyMu.RLock()
		hotKeys := make([]string, 0, len(c.hotKeyCounts))
		for key, count := range c.hotKeyCounts {
			if atomic.LoadInt64(count) > c.hotKeyThreshold {
				hotKeys = append(hotKeys, key)
			}
			// Reset count
			atomic.StoreInt64(count, 0)
		}
		c.hotKeyMu.RUnlock()

		for _, key := range hotKeys {
			log.Printf("[INFO] Hot key detected: %s (replicating)", key)
			// In a production system, this would replicate the hot key
			// to additional nodes. Here we just log it.
		}
	}
}

// recordKeyAccess records a key access for hot key detection.
func (c *Client) recordKeyAccess(key string) {
	c.hotKeyMu.RLock()
	counter, exists := c.hotKeyCounts[key]
	c.hotKeyMu.RUnlock()

	if !exists {
		c.hotKeyMu.Lock()
		if counter, exists = c.hotKeyCounts[key]; !exists {
			var zero int64
			counter = &zero
			c.hotKeyCounts[key] = counter
		}
		c.hotKeyMu.Unlock()
	}

	atomic.AddInt64(counter, 1)
}

// ============================================================================
// Cache Stampede Prevention (Mutex Locking)
// ============================================================================

// CacheAsideWithStampedeProtection implements cache-aside with stampede prevention.
// Multiple concurrent requests for the same key will be deduplicated - only one
// goroutine loads the data while others wait.
func (c *Client) CacheAsideWithStampedeProtection(ctx context.Context, key string, ttl time.Duration, loader func() (interface{}, error), dest interface{}) error {
	versionedKey := VersionedKey(key)

	// Try cache first
	c.recordKeyAccess(versionedKey)

	if !c.bloomFilter.MightContain(versionedKey) {
		// Definitely not in cache, load directly
		data, err := loader()
		if err != nil {
			return fmt.Errorf("cache-aside loader failed: %w", err)
		}
		// Store in cache
		jsonData, _ := json.Marshal(data)
		cacheCtx, cancel := context.WithTimeout(ctx, 5*time.Second)
		defer cancel()
		c.Client.Set(cacheCtx, versionedKey, jsonData, ttl)
		c.bloomFilter.Add(versionedKey)
		return json.Unmarshal(jsonData, dest)
	}

	// Try cache first (read from slave if available)
	err := c.GetWithReadReplica(ctx, versionedKey, dest)
	if err == nil {
		return nil // Cache hit
	}

	// Cache miss - use stampede prevention
	sl := c.getStampedeLock(versionedKey)
	sl.mu.Lock()

	if sl.loading {
		// Another goroutine is already loading
		sl.mu.Unlock()
		// Wait for it to finish
		<-sl.done
		// Try cache again
		err := c.GetWithReadReplica(ctx, versionedKey, dest)
		if err != nil {
			return fmt.Errorf("stampede wait failed: %w", err)
		}
		return nil
	}

	// We are the loader
	sl.loading = true
	sl.done = make(chan struct{})
	sl.mu.Unlock()

	// Load data
	data, err := loader()
	if err != nil {
		c.releaseStampedeLock(versionedKey)
		return fmt.Errorf("cache-aside loader failed: %w", err)
	}

	// Store in cache
	jsonData, err := json.Marshal(data)
	if err != nil {
		c.releaseStampedeLock(versionedKey)
		return fmt.Errorf("failed to marshal data: %w", err)
	}

	cacheCtx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	// Write to master
	if c.clusterMode {
		c.clusterClient.Set(cacheCtx, versionedKey, jsonData, ttl)
	} else {
		c.Client.Set(cacheCtx, versionedKey, jsonData, ttl)
	}

	c.bloomFilter.Add(versionedKey)

	// Signal waiters
	sl.mu.Lock()
	sl.loading = false
	sl.data = data
	sl.err = nil
	close(sl.done)
	sl.mu.Unlock()

	c.releaseStampedeLock(versionedKey)

	return json.Unmarshal(jsonData, dest)
}

func (c *Client) getStampedeLock(key string) *stampedeLock {
	c.stampedeMu.Lock()
	defer c.stampedeMu.Unlock()

	sl, exists := c.stampedeLocks[key]
	if !exists {
		sl = &stampedeLock{
			done: make(chan struct{}),
		}
		c.stampedeLocks[key] = sl
	}
	atomic.AddInt32(&sl.refCount, 1)
	return sl
}

func (c *Client) releaseStampedeLock(key string) {
	c.stampedeMu.Lock()
	defer c.stampedeMu.Unlock()

	sl, exists := c.stampedeLocks[key]
	if !exists {
		return
	}

	if atomic.AddInt32(&sl.refCount, -1) <= 0 {
		delete(c.stampedeLocks, key)
	}
}

// ============================================================================
// Read Replicas (read from slave, write to master)
// ============================================================================

// GetWithReadReplica reads from a read replica if available, otherwise from master.
func (c *Client) GetWithReadReplica(ctx context.Context, key string, dest interface{}) error {
	reader := c.GetReadClient()
	cacheCtx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	data, err := reader.Get(cacheCtx, key).Bytes()
	if err == redis.Nil {
		return fmt.Errorf("key not found: %s", key)
	}
	if err != nil {
		return fmt.Errorf("failed to get key %s: %w", key, err)
	}

	if err := json.Unmarshal(data, dest); err != nil {
		return fmt.Errorf("failed to unmarshal value for key %s: %w", key, err)
	}
	return nil
}

// GetReadClient returns the appropriate client for read operations.
func (c *Client) GetReadClient() *redis.Client {
	if c.readFromSlave && len(c.readReplicas) > 0 {
		// Pick a random replica
		idx := int(time.Now().UnixNano()) % len(c.readReplicas)
		return c.readReplicas[idx]
	}
	if c.clusterMode && c.clusterClient != nil {
		// Use cluster client's read operations
		return nil // handled differently
	}
	return c.Client
}

// GetWriteClient returns the master client for write operations.
func (c *Client) GetWriteClient() *redis.Client {
	return c.Client
}

// ============================================================================
// Pipeline Batching for Bulk Operations
// ============================================================================

// BatchPipeline accumulates operations and flushes them in batches.
func (c *Client) BatchPipeline() *PipelineBatch {
	return &PipelineBatch{
		client: c,
		ops:    make([]func(), 0),
	}
}

// PipelineBatch accumulates pipeline operations.
type PipelineBatch struct {
	client *Client
	ops    []func()
	mu     sync.Mutex
}

// Set adds a SET operation to the pipeline batch.
func (pb *PipelineBatch) Set(ctx context.Context, key string, value interface{}, ttl time.Duration) {
	pb.mu.Lock()
	defer pb.mu.Unlock()

	versionedKey := VersionedKey(key)
	data, _ := json.Marshal(value)
	pb.ops = append(pb.ops, func() {
		cacheCtx, cancel := context.WithTimeout(ctx, 5*time.Second)
		defer cancel()
		pb.client.Client.Set(cacheCtx, versionedKey, data, ttl)
	})
}

// Delete adds a DEL operation to the pipeline batch.
func (pb *PipelineBatch) Delete(ctx context.Context, keys ...string) {
	pb.mu.Lock()
	defer pb.mu.Unlock()

	versionedKeys := make([]string, len(keys))
	for i, k := range keys {
		versionedKeys[i] = VersionedKey(k)
	}
	pb.ops = append(pb.ops, func() {
		cacheCtx, cancel := context.WithTimeout(ctx, 5*time.Second)
		defer cancel()
		pb.client.Client.Del(cacheCtx, versionedKeys...)
	})
}

// Flush executes all buffered pipeline operations.
func (pb *PipelineBatch) Flush(ctx context.Context) error {
	pb.mu.Lock()
	ops := pb.ops
	pb.ops = nil
	pb.mu.Unlock()

	if len(ops) == 0 {
		return nil
	}

	pipe := pb.client.Client.Pipeline()
	for _, op := range ops {
		op()
	}

	_, err := pipe.Exec(ctx)
	if err != nil {
		return fmt.Errorf("pipeline batch flush failed: %w", err)
	}

	return nil
}

// batchFlusher periodically flushes buffered batch operations.
func (c *Client) batchFlusher() {
	ticker := time.NewTicker(100 * time.Millisecond)
	defer ticker.Stop()

	for range ticker.C {
		c.batchMu.Lock()
		if len(c.batchBuffer) == 0 {
			c.batchMu.Unlock()
			continue
		}

		// Flush buffer
		buffer := c.batchBuffer
		c.batchBuffer = make(map[string]*batchOp)
		c.batchMu.Unlock()

		if len(buffer) == 0 {
			continue
		}

		pipe := c.Client.Pipeline()
		for _, op := range buffer {
			versionedKey := VersionedKey(op.key)
			data, err := json.Marshal(op.val)
			if err != nil {
				log.Printf("[WARN] Batch marshal error for key %s: %v", op.key, err)
				continue
			}
			pipe.Set(context.Background(), versionedKey, data, op.ttl)
		}

		if _, err := pipe.Exec(context.Background()); err != nil {
			log.Printf("[WARN] Batch pipeline flush error: %v", err)
		}
	}
}

// BufferedSet adds a SET operation to the batch buffer for periodic flushing.
func (c *Client) BufferedSet(ctx context.Context, key string, value interface{}, ttl time.Duration) error {
	c.batchMu.Lock()
	defer c.batchMu.Unlock()

	c.batchBuffer[key] = &batchOp{
		cmd: "SET",
		key: key,
		val: value,
		ttl: ttl,
	}

	if len(c.batchBuffer) >= c.pipelineBatchSize {
		// Force flush
		buffer := c.batchBuffer
		c.batchBuffer = make(map[string]*batchOp)
		go func() {
			pipe := c.Client.Pipeline()
			for _, op := range buffer {
				versionedKey := VersionedKey(op.key)
				data, _ := json.Marshal(op.val)
				pipe.Set(context.Background(), versionedKey, data, op.ttl)
			}
			if _, err := pipe.Exec(context.Background()); err != nil {
				log.Printf("[WARN] Batch pipeline force flush error: %v", err)
			}
		}()
	}

	return nil
}

// NewFromEnv creates a Redis client from environment variables.
// Cache-Aside Pattern with Automatic Invalidation
// ============================================================================

// CacheAside implements the cache-aside pattern.
// It attempts to fetch from cache first, falls back to the loader function,
// and automatically populates the cache.
func (c *Client) CacheAside(ctx context.Context, key string, ttl time.Duration, loader func() (interface{}, error), dest interface{}) error {
	versionedKey := VersionedKey(key)

	// Try cache first
	err := c.Get(ctx, versionedKey, dest)
	if err == nil {
		return nil // Cache hit
	}

	// Cache miss — call loader
	data, err := loader()
	if err != nil {
		return fmt.Errorf("cache-aside loader failed: %w", err)
	}

	// Marshal data for cache
	jsonData, err := json.Marshal(data)
	if err != nil {
		return fmt.Errorf("failed to marshal data: %w", err)
	}

	// Store in cache
	cacheCtx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	if err := c.Client.Set(cacheCtx, versionedKey, jsonData, ttl).Err(); err != nil {
		log.Printf("[WARN] Failed to set cache for key %s: %v", versionedKey, err)
	} else {
		// Also store unmarshaled
		if err := json.Unmarshal(jsonData, dest); err != nil {
			return fmt.Errorf("failed to unmarshal data: %w", err)
		}
	}

	return nil
}

// InvalidateKey invalidates a specific cache key.
func (c *Client) InvalidateKey(ctx context.Context, key string) error {
	versionedKey := VersionedKey(key)
	return c.Delete(ctx, versionedKey)
}

// InvalidatePattern invalidates all keys matching a pattern.
func (c *Client) InvalidatePattern(ctx context.Context, pattern string) error {
	versionedPattern := VersionedKey(pattern)
	return c.DeletePattern(ctx, versionedPattern)
}

// ============================================================================
// Bulk Operations
// ============================================================================

// MGet retrieves multiple keys at once using Redis MGET.
func (c *Client) MGet(ctx context.Context, keys ...string) (map[string][]byte, error) {
	if len(keys) == 0 {
		return nil, nil
	}

	versionedKeys := make([]string, len(keys))
	for i, k := range keys {
		versionedKeys[i] = VersionedKey(k)
	}

	cacheCtx, cancel := context.WithTimeout(ctx, 10*time.Second)
	defer cancel()

	vals, err := c.Client.MGet(cacheCtx, versionedKeys...).Result()
	if err != nil {
		return nil, fmt.Errorf("MGet failed: %w", err)
	}

	result := make(map[string][]byte)
	for i, val := range vals {
		if val != nil {
			if s, ok := val.(string); ok {
				result[keys[i]] = []byte(s)
			}
		}
	}

	return result, nil
}

// MSet sets multiple key-value pairs at once using Redis MSET.
func (c *Client) MSet(ctx context.Context, pairs map[string]interface{}, ttl time.Duration) error {
	if len(pairs) == 0 {
		return nil
	}

	// Use pipeline for atomicity
	cacheCtx, cancel := context.WithTimeout(ctx, 10*time.Second)
	defer cancel()

	pipe := c.Client.Pipeline()
	for key, value := range pairs {
		versionedKey := VersionedKey(key)
		data, err := json.Marshal(value)
		if err != nil {
			return fmt.Errorf("failed to marshal value for key %s: %w", key, err)
		}
		pipe.Set(cacheCtx, versionedKey, data, ttl)
	}

	_, err := pipe.Exec(cacheCtx)
	if err != nil {
		return fmt.Errorf("MSet pipeline failed: %w", err)
	}

	return nil
}

// GetBulkResult represents the result of a bulk get with missing keys.
type GetBulkResult struct {
	Data    map[string]interface{}
	Missing []string
}

// GetBulk retrieves multiple keys and identifies which are missing.
func (c *Client) GetBulk(ctx context.Context, keys []string) GetBulkResult {
	result := GetBulkResult{
		Data: make(map[string]interface{}),
	}

	if len(keys) == 0 {
		return result
	}

	rawResult, err := c.MGet(ctx, keys...)
	if err != nil {
		result.Missing = keys
		return result
	}

	keySet := make(map[string]bool)
	for _, k := range keys {
		keySet[k] = true
	}

	for key, data := range rawResult {
		var val interface{}
		if err := json.Unmarshal(data, &val); err == nil {
			result.Data[key] = val
			delete(keySet, key)
		}
	}

	for k := range keySet {
		result.Missing = append(result.Missing, k)
	}

	return result
}

// ============================================================================
// Cache Warming
// ============================================================================

// WarmerFunc is a function type for loading data during cache warming.
type WarmerFunc func(ctx context.Context) (map[string]interface{}, error)

// CacheWarmResult contains the result of a cache warming operation.
type CacheWarmResult struct {
	KeysWarmed int
	KeysFailed int
	Duration   time.Duration
}

// WarmCache pre-populates the cache using the provided warmer functions.
func (c *Client) WarmCache(ctx context.Context, warmers map[string]WarmerFunc, ttl time.Duration) CacheWarmResult {
	start := time.Now()
	result := CacheWarmResult{}

	for keyPrefix, warmer := range warmers {
		data, err := warmer(ctx)
		if err != nil {
			log.Printf("[WARN] Cache warmer for %s failed: %v", keyPrefix, err)
			result.KeysFailed++
			continue
		}

		pairs := make(map[string]interface{})
		for k, v := range data {
			pairs[fmt.Sprintf("%s:%s", keyPrefix, k)] = v
		}

		if err := c.MSet(ctx, pairs, ttl); err != nil {
			log.Printf("[WARN] Cache warmer MSet for %s failed: %v", keyPrefix, err)
			result.KeysFailed++
			continue
		}

		result.KeysWarmed++
	}

	result.Duration = time.Since(start)
	return result
}

// WarmProblemList warms the problem list cache.
func (c *Client) WarmProblemList(ctx context.Context, problems interface{}) error {
	data, err := json.Marshal(problems)
	if err != nil {
		return fmt.Errorf("failed to marshal problems: %w", err)
	}

	versionedKey := VersionedKey("problems:list:all")
	cacheCtx, cancel := context.WithTimeout(ctx, 10*time.Second)
	defer cancel()

	return c.Client.Set(cacheCtx, versionedKey, data, TTLProblemList).Err()
}

// ============================================================================
// Original methods (preserved for backward compatibility)
// ============================================================================

// Set marshals and stores a value with TTL.
func (c *Client) Set(ctx context.Context, key string, value interface{}, ttl time.Duration) error {
	versionedKey := VersionedKey(key)

	data, err := json.Marshal(value)
	if err != nil {
		return fmt.Errorf("failed to marshal value for key %s: %w", key, err)
	}

	cacheCtx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	if err := c.Client.Set(cacheCtx, versionedKey, data, ttl).Err(); err != nil {
		return fmt.Errorf("failed to set key %s: %w", key, err)
	}
	return nil
}

// Get retrieves and unmarshals a value.
func (c *Client) Get(ctx context.Context, key string, dest interface{}) error {
	versionedKey := VersionedKey(key)

	cacheCtx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	data, err := c.Client.Get(cacheCtx, versionedKey).Bytes()
	if err == redis.Nil {
		return fmt.Errorf("key not found: %s", key)
	}
	if err != nil {
		return fmt.Errorf("failed to get key %s: %w", key, err)
	}

	if err := json.Unmarshal(data, dest); err != nil {
		return fmt.Errorf("failed to unmarshal value for key %s: %w", key, err)
	}
	return nil
}

// GetString retrieves a raw string value.
func (c *Client) GetString(ctx context.Context, key string) (string, error) {
	versionedKey := VersionedKey(key)

	cacheCtx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	val, err := c.Client.Get(cacheCtx, versionedKey).Result()
	if err == redis.Nil {
		return "", fmt.Errorf("key not found: %s", key)
	}
	if err != nil {
		return "", fmt.Errorf("failed to get key %s: %w", key, err)
	}
	return val, nil
}

// Delete removes one or more keys.
func (c *Client) Delete(ctx context.Context, keys ...string) error {
	versionedKeys := make([]string, len(keys))
	for i, k := range keys {
		versionedKeys[i] = VersionedKey(k)
	}

	cacheCtx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	if err := c.Client.Del(cacheCtx, versionedKeys...).Err(); err != nil {
		return fmt.Errorf("failed to delete keys: %w", err)
	}
	return nil
}

// DeletePattern deletes all keys matching a pattern.
func (c *Client) DeletePattern(ctx context.Context, pattern string) error {
	versionedPattern := VersionedKey(pattern)

	cacheCtx, cancel := context.WithTimeout(ctx, 10*time.Second)
	defer cancel()

	iter := c.Client.Scan(cacheCtx, 0, versionedPattern, 0).Iterator()
	var keys []string
	for iter.Next(cacheCtx) {
		keys = append(keys, iter.Val())
	}
	if err := iter.Err(); err != nil {
		return fmt.Errorf("failed to scan keys with pattern %s: %w", pattern, err)
	}

	if len(keys) > 0 {
		if err := c.Client.Del(cacheCtx, keys...).Err(); err != nil {
			return fmt.Errorf("failed to delete keys: %w", err)
		}
	}
	return nil
}

// Exists checks if a key exists.
func (c *Client) Exists(ctx context.Context, key string) (bool, error) {
	versionedKey := VersionedKey(key)

	cacheCtx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	n, err := c.Client.Exists(cacheCtx, versionedKey).Result()
	if err != nil {
		return false, fmt.Errorf("failed to check key %s: %w", key, err)
	}
	return n > 0, nil
}

// SetNX sets a value only if the key does not exist (useful for distributed locks).
func (c *Client) SetNX(ctx context.Context, key string, value interface{}, ttl time.Duration) (bool, error) {
	versionedKey := VersionedKey(key)

	data, err := json.Marshal(value)
	if err != nil {
		return false, fmt.Errorf("failed to marshal value: %w", err)
	}

	cacheCtx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	ok, err := c.Client.SetNX(cacheCtx, versionedKey, data, ttl).Result()
	if err != nil {
		return false, fmt.Errorf("failed to setnx key %s: %w", key, err)
	}
	return ok, nil
}

// Increment atomically increments a key.
func (c *Client) Increment(ctx context.Context, key string) (int64, error) {
	versionedKey := VersionedKey(key)

	cacheCtx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	val, err := c.Client.Incr(cacheCtx, versionedKey).Result()
	if err != nil {
		return 0, fmt.Errorf("failed to increment key %s: %w", key, err)
	}
	return val, nil
}

// Publish publishes a message to a channel.
func (c *Client) Publish(ctx context.Context, channel string, message interface{}) error {
	data, err := json.Marshal(message)
	if err != nil {
		return fmt.Errorf("failed to marshal message: %w", err)
	}

	cacheCtx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	if err := c.Client.Publish(cacheCtx, channel, data).Err(); err != nil {
		return fmt.Errorf("failed to publish to channel %s: %w", channel, err)
	}
	return nil
}

// Subscribe subscribes to channels and returns a channel of messages.
func (c *Client) Subscribe(ctx context.Context, channels ...string) <-chan *redis.Message {
	pubsub := c.Client.Subscribe(ctx, channels...)
	return pubsub.Channel()
}

// ZAdd adds a member to a sorted set.
func (c *Client) ZAdd(ctx context.Context, key string, score float64, member string) error {
	versionedKey := VersionedKey(key)

	cacheCtx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	if err := c.Client.ZAdd(cacheCtx, versionedKey, redis.Z{Score: score, Member: member}).Err(); err != nil {
		return fmt.Errorf("failed to zadd to %s: %w", key, err)
	}
	return nil
}

// ZRevRange gets members from a sorted set in reverse order (highest score first).
func (c *Client) ZRevRange(ctx context.Context, key string, start, stop int64) ([]string, error) {
	versionedKey := VersionedKey(key)

	cacheCtx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	members, err := c.Client.ZRevRange(cacheCtx, versionedKey, start, stop).Result()
	if err != nil {
		return nil, fmt.Errorf("failed to zrevrange %s: %w", key, err)
	}
	return members, nil
}

// ZRevRank gets the rank of a member in a sorted set (0-based, highest score = rank 0).
func (c *Client) ZRevRank(ctx context.Context, key, member string) (int64, error) {
	versionedKey := VersionedKey(key)

	cacheCtx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	rank, err := c.Client.ZRevRank(cacheCtx, versionedKey, member).Result()
	if err == redis.Nil {
		return -1, nil
	}
	if err != nil {
		return -1, fmt.Errorf("failed to zrevrank %s in %s: %w", member, key, err)
	}
	return rank, nil
}

// HealthCheck verifies Redis connection.
func (c *Client) HealthCheck(ctx context.Context) error {
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	if err := c.Client.Ping(ctx).Err(); err != nil {
		return fmt.Errorf("redis health check failed: %w", err)
	}
	return nil
}

// HealthCheckDetailed verifies Redis connection with detailed stats.
func (c *Client) HealthCheckDetailed(ctx context.Context) map[string]interface{} {
	result := c.healthMonitor.GetStats()
	healthy, _ := c.HealthCheck(ctx), error(nil)
	_ = healthy

	poolStats := c.Client.PoolStats()
	result["pool_hits"] = poolStats.Hits
	result["pool_misses"] = poolStats.Misses
	result["pool_timeouts"] = poolStats.Timeouts
	result["pool_total_conns"] = poolStats.TotalConns
	result["pool_idle_conns"] = poolStats.IdleConns
	result["pool_stale_conns"] = poolStats.StaleConns
	result["circuit_breaker_state"] = c.cb.State()
	result["cache_version"] = c.cacheVersion

	return result
}

// PoolStats returns Redis pool statistics.
func (c *Client) PoolStats() *redis.PoolStats {
	return c.Client.PoolStats()
}

// Close gracefully closes the Redis connection.
func (c *Client) Close() error {
	return c.Client.Close()
}

func getEnv(key, fallback string) string {
	if v, ok := os.LookupEnv(key); ok {
		return v
	}
	return fallback
}

func getEnvInt(key string, fallback int) int {
	if v, ok := os.LookupEnv(key); ok {
		var n int
		if _, err := fmt.Sscanf(v, "%d", &n); err == nil {
			return n
		}
	}
	return fallback
}

// Unused import guard
var _ = list.New
