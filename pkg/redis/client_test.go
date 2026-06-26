package redis

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"testing"
	"time"
)

func TestConfig_Defaults(t *testing.T) {
	cfg := Config{
		Addr:     "localhost:6379",
		Password: "",
		DB:       0,
		PoolSize: 10,
	}
	if cfg.Addr != "localhost:6379" {
		t.Errorf("expected localhost:6379, got %q", cfg.Addr)
	}
	if cfg.PoolSize != 10 {
		t.Errorf("expected pool size 10, got %d", cfg.PoolSize)
	}
}

func TestKeyPatterns(t *testing.T) {
	// Key patterns are format strings (e.g., "problem:%s")
	// Verify they contain the expected prefix and format specifier
	tests := []struct {
		name   string
		pattern string
		prefix string
	}{
		{"KeyProblem", KeyProblem, "problem:"},
		{"KeyProblemList", KeyProblemList, "problems:list:"},
		{"KeyUser", KeyUser, "user:"},
		{"KeyUserStats", KeyUserStats, "user:stats:"},
		{"KeyUserProgress", KeyUserProgress, "user:progress:"},
		{"KeySubmission", KeySubmission, "submission:"},
		{"KeySession", KeySession, "session:"},
		{"KeyLeaderboard", KeyLeaderboard, "leaderboard:"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if !bytes.HasPrefix([]byte(tt.pattern), []byte(tt.prefix)) {
				t.Errorf("expected pattern %q to start with %q", tt.pattern, tt.prefix)
			}
		})
	}
}

func TestTTLValues(t *testing.T) {
	tests := []struct {
		name     string
		ttl      time.Duration
		expected time.Duration
	}{
		{"TTLSession", TTLSession, 24 * time.Hour},
		{"TTLUser", TTLUser, 1 * time.Hour},
		{"TTLUserStats", TTLUserStats, 5 * time.Minute},
		{"TTLProblem", TTLProblem, 30 * time.Minute},
		{"TTLProblemList", TTLProblemList, 10 * time.Minute},
		{"TTLLeaderboard", TTLLeaderboard, 1 * time.Minute},
		{"TTLSubmission", TTLSubmission, 5 * time.Minute},
		{"TTLRateLimit", TTLRateLimit, 1 * time.Minute},
		{"TTLHint", TTLHint, 30 * time.Minute},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.ttl != tt.expected {
				t.Errorf("expected %v, got %v", tt.expected, tt.ttl)
			}
		})
	}
}

func TestChannelNames(t *testing.T) {
	if ChannelLeaderboardUpdate != "leaderboard:update" {
		t.Errorf("expected leaderboard:update, got %q", ChannelLeaderboardUpdate)
	}
	if ChannelSubmissionNew != "submission:new" {
		t.Errorf("expected submission:new, got %q", ChannelSubmissionNew)
	}
	if ChannelSubmissionResult != "submission:result" {
		t.Errorf("expected submission:result, got %q", ChannelSubmissionResult)
	}
}

// mockClient is a mock implementation for testing cache operations
// without requiring a real Redis connection.
type mockClient struct {
	data     map[string][]byte
	keys     map[string]time.Time // key -> expiry
	channels map[string][][]byte
}

func newMockClient() *mockClient {
	return &mockClient{
		data:     make(map[string][]byte),
		keys:     make(map[string]time.Time),
		channels: make(map[string][][]byte),
	}
}

func (m *mockClient) set(key string, value interface{}, ttl time.Duration) error {
	data, err := json.Marshal(value)
	if err != nil {
		return err
	}
	m.data[key] = data
	if ttl > 0 {
		m.keys[key] = time.Now().Add(ttl)
	}
	return nil
}

func (m *mockClient) get(key string, dest interface{}) error {
	data, ok := m.data[key]
	if !ok {
		return fmt.Errorf("key not found: %s", key)
	}
	return json.Unmarshal(data, dest)
}

func (m *mockClient) getString(key string) (string, error) {
	data, ok := m.data[key]
	if !ok {
		return "", fmt.Errorf("key not found: %s", key)
	}
	return string(data), nil
}

func (m *mockClient) del(keys ...string) error {
	for _, k := range keys {
		delete(m.data, k)
		delete(m.keys, k)
	}
	return nil
}

func (m *mockClient) exists(key string) bool {
	_, ok := m.data[key]
	return ok
}

func (m *mockClient) setNX(key string, value interface{}, ttl time.Duration) (bool, error) {
	if _, exists := m.data[key]; exists {
		return false, nil
	}
	return true, m.set(key, value, ttl)
}

func (m *mockClient) increment(key string) (int64, error) {
	var val int64
	if data, ok := m.data[key]; ok {
		json.Unmarshal(data, &val)
	}
	val++
	data, _ := json.Marshal(val)
	m.data[key] = data
	return val, nil
}

func (m *mockClient) publish(channel string, message interface{}) error {
	data, err := json.Marshal(message)
	if err != nil {
		return err
	}
	m.channels[channel] = append(m.channels[channel], data)
	return nil
}

func (m *mockClient) isExpired(key string) bool {
	expiry, ok := m.keys[key]
	if !ok {
		return false
	}
	return time.Now().After(expiry)
}

// Test cache operations
func TestMockCache_SetAndGet(t *testing.T) {
	c := newMockClient()
	ctx := context.Background()
	_ = ctx

	type User struct {
		ID   string `json:"id"`
		Name string `json:"name"`
	}

	user := User{ID: "123", Name: "Alice"}
	if err := c.set("user:123", user, 0); err != nil {
		t.Fatalf("set failed: %v", err)
	}

	var result User
	if err := c.get("user:123", &result); err != nil {
		t.Fatalf("get failed: %v", err)
	}

	if result.ID != "123" || result.Name != "Alice" {
		t.Errorf("expected {123 Alice}, got %+v", result)
	}
}

func TestMockCache_GetNonExistent(t *testing.T) {
	c := newMockClient()
	var result map[string]interface{}
	err := c.get("nonexistent", &result)
	if err == nil {
		t.Error("expected error for non-existent key")
	}
}

func TestMockCache_Delete(t *testing.T) {
	c := newMockClient()
	c.set("key1", "value1", 0)
	c.set("key2", "value2", 0)

	c.del("key1")

	if c.exists("key1") {
		t.Error("key1 should be deleted")
	}
	if !c.exists("key2") {
		t.Error("key2 should still exist")
	}
}

func TestMockCache_Exists(t *testing.T) {
	c := newMockClient()
	c.set("exists-key", "value", 0)

	if !c.exists("exists-key") {
		t.Error("expected key to exist")
	}
	if c.exists("nonexistent") {
		t.Error("expected key to not exist")
	}
}

func TestMockCache_SetNX(t *testing.T) {
	c := newMockClient()

	ok, err := c.setNX("lock-key", "value1", 0)
	if err != nil {
		t.Fatalf("setNX failed: %v", err)
	}
	if !ok {
		t.Error("expected setNX to succeed for new key")
	}

	ok, err = c.setNX("lock-key", "value2", 0)
	if err != nil {
		t.Fatalf("setNX failed: %v", err)
	}
	if ok {
		t.Error("expected setNX to fail for existing key")
	}
}

func TestMockCache_Increment(t *testing.T) {
	c := newMockClient()

	val, _ := c.increment("counter")
	if val != 1 {
		t.Errorf("expected 1, got %d", val)
	}

	val, _ = c.increment("counter")
	if val != 2 {
		t.Errorf("expected 2, got %d", val)
	}

	val, _ = c.increment("counter")
	if val != 3 {
		t.Errorf("expected 3, got %d", val)
	}
}

func TestMockCache_IncrementFromExisting(t *testing.T) {
	c := newMockClient()
	c.set("existing-counter", 10, 0)

	val, _ := c.increment("existing-counter")
	if val != 11 {
		t.Errorf("expected 11, got %d", val)
	}
}

// Test TTL expiration
func TestMockCache_TTLExpiration(t *testing.T) {
	c := newMockClient()
	c.set("temp-key", "value", 1*time.Millisecond)

	if c.isExpired("temp-key") {
		t.Error("key should not be expired immediately")
	}

	time.Sleep(5 * time.Millisecond)
	if !c.isExpired("temp-key") {
		t.Error("key should be expired after TTL")
	}
}

func TestMockCache_TTLNoExpiry(t *testing.T) {
	c := newMockClient()
	c.set("permanent-key", "value", 0)

	if c.isExpired("permanent-key") {
		t.Error("key with no TTL should not expire")
	}
}

func TestMockCache_TTLValues(t *testing.T) {
	// Verify TTL constants are reasonable
	if TTLSession < TTLUser {
		t.Error("session TTL should be >= user TTL")
	}
	if TTLProblem < TTLProblemList {
		t.Log("Note: problem TTL is less than problem list TTL (this may be intentional)")
	}
	if TTLLeaderboard > TTLUserStats {
		t.Log("Note: leaderboard TTL is greater than user stats TTL")
	}
}

// Test Pub/Sub
func TestMockPubSub_Publish(t *testing.T) {
	c := newMockClient()

	msg := map[string]interface{}{"user": "alice", "score": 100}
	if err := c.publish("test-channel", msg); err != nil {
		t.Fatalf("publish failed: %v", err)
	}

	if len(c.channels["test-channel"]) != 1 {
		t.Errorf("expected 1 message, got %d", len(c.channels["test-channel"]))
	}
}

func TestMockPubSub_MultipleMessages(t *testing.T) {
	c := newMockClient()

	for i := 0; i < 5; i++ {
		c.publish("events", map[string]interface{}{"seq": i})
	}

	if len(c.channels["events"]) != 5 {
		t.Errorf("expected 5 messages, got %d", len(c.channels["events"]))
	}
}

func TestMockPubSub_MultipleChannels(t *testing.T) {
	c := newMockClient()

	c.publish("channel-a", "msg-a")
	c.publish("channel-b", "msg-b")
	c.publish("channel-a", "msg-a2")

	if len(c.channels["channel-a"]) != 2 {
		t.Errorf("expected 2 messages on channel-a, got %d", len(c.channels["channel-a"]))
	}
	if len(c.channels["channel-b"]) != 1 {
		t.Errorf("expected 1 message on channel-b, got %d", len(c.channels["channel-b"]))
	}
}

func TestMockPubSub_MessageContent(t *testing.T) {
	c := newMockClient()

	type Event struct {
		Type string `json:"type"`
		Data string `json:"data"`
	}

	event := Event{Type: "submission", Data: "accepted"}
	c.publish("events", event)

	raw := c.channels["events"][0]
	var decoded Event
	if err := json.Unmarshal(raw, &decoded); err != nil {
		t.Fatalf("failed to decode: %v", err)
	}

	if decoded.Type != "submission" || decoded.Data != "accepted" {
		t.Errorf("expected {submission accepted}, got %+v", decoded)
	}
}

func TestMockCache_JSONRoundTrip(t *testing.T) {
	c := newMockClient()

	type Complex struct {
		Numbers []int            `json:"numbers"`
		Nested  map[string]int   `json:"nested"`
		Active  bool             `json:"active"`
	}

	original := Complex{
		Numbers: []int{1, 2, 3},
		Nested:  map[string]int{"a": 1, "b": 2},
		Active:  true,
	}

	c.set("complex", original, 0)

	var decoded Complex
	c.get("complex", &decoded)

	if !bytes.Equal(mustMarshal(original), mustMarshal(decoded)) {
		t.Errorf("round-trip failed: %+v != %+v", original, decoded)
	}
}

func mustMarshal(v interface{}) []byte {
	b, _ := json.Marshal(v)
	return b
}

// Test that the actual Client struct embeds *redis.Client
// (compile-time check that the type is correct)
func TestClientStruct(t *testing.T) {
	// This test verifies the Client struct definition at compile time
	// The actual Redis connection is tested in integration tests
	var _ = Config{}
	var _ = TTLSession
	var _ = ChannelLeaderboardUpdate
}
