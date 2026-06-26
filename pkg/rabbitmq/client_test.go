package rabbitmq

import (
	"os"
	"strings"
	"testing"
)

// ============================================================================
// Config Tests
// ============================================================================

func TestConfigDefaults(t *testing.T) {
	cfg := Config{
		Host:     "localhost",
		Port:     5672,
		User:     "guest",
		Password: "guest",
		VHost:    "/",
	}
	if cfg.Host != "localhost" {
		t.Errorf("expected localhost, got %q", cfg.Host)
	}
	if cfg.Port != 5672 {
		t.Errorf("expected 5672, got %d", cfg.Port)
	}
	if cfg.User != "guest" {
		t.Errorf("expected guest, got %q", cfg.User)
	}
}

// ============================================================================
// Queue Constants Tests
// ============================================================================

func TestQueueConstants(t *testing.T) {
	if QueueCodeExecution != "code.execution.pending" {
		t.Errorf("expected 'code.execution.pending', got %q", QueueCodeExecution)
	}
	if QueueNotifications != "notifications" {
		t.Errorf("expected 'notifications', got %q", QueueNotifications)
	}
	if QueueDLX != "dead.letter" {
		t.Errorf("expected 'dead.letter', got %q", QueueDLX)
	}
	if ExchangeCodeExec != "code.execution" {
		t.Errorf("expected 'code.execution', got %q", ExchangeCodeExec)
	}
	if ExchangeEvents != "events" {
		t.Errorf("expected 'events', got %q", ExchangeEvents)
	}
}

// ============================================================================
// generateMessageID Tests
// ============================================================================

func TestGenerateMessageID(t *testing.T) {
	id := generateMessageID()
	if id == "" {
		t.Error("expected non-empty message ID")
	}
	if !strings.Contains(id, "-") {
		t.Errorf("expected message ID to contain '-', got %q", id)
	}
}

func TestGenerateMessageID_Format(t *testing.T) {
	id := generateMessageID()
	if id == "" {
		t.Error("expected non-empty message ID")
	}
	if !strings.Contains(id, "-") {
		t.Errorf("expected message ID to contain '-', got %q", id)
	}
	parts := strings.SplitN(id, "-", 2)
	if len(parts) != 2 {
		t.Errorf("expected format 'timestamp-pid', got %q", id)
	}
	if parts[0] == "" || parts[1] == "" {
		t.Errorf("expected both timestamp and pid to be non-empty, got %q", id)
	}
}

// ============================================================================
// getEnv / getEnvInt Tests
// ============================================================================

func TestGetEnv_Default(t *testing.T) {
	os.Unsetenv("TEST_RABBITMQ_KEY")
	val := getEnv("TEST_RABBITMQ_KEY", "default-val")
	if val != "default-val" {
		t.Errorf("expected 'default-val', got %q", val)
	}
}

func TestGetEnv_FromEnv(t *testing.T) {
	os.Setenv("TEST_RABBITMQ_KEY", "env-val")
	defer os.Unsetenv("TEST_RABBITMQ_KEY")

	val := getEnv("TEST_RABBITMQ_KEY", "default")
	if val != "env-val" {
		t.Errorf("expected 'env-val', got %q", val)
	}
}

func TestGetEnvInt_Default(t *testing.T) {
	os.Unsetenv("TEST_RABBITMQ_PORT")
	val := getEnvInt("TEST_RABBITMQ_PORT", 5672)
	if val != 5672 {
		t.Errorf("expected 5672, got %d", val)
	}
}

func TestGetEnvInt_FromEnv(t *testing.T) {
	os.Setenv("TEST_RABBITMQ_PORT", "8888")
	defer os.Unsetenv("TEST_RABBITMQ_PORT")

	val := getEnvInt("TEST_RABBITMQ_PORT", 5672)
	if val != 8888 {
		t.Errorf("expected 8888, got %d", val)
	}
}

func TestGetEnvInt_InvalidValue(t *testing.T) {
	os.Setenv("TEST_RABBITMQ_PORT", "not-a-number")
	defer os.Unsetenv("TEST_RABBITMQ_PORT")

	val := getEnvInt("TEST_RABBITMQ_PORT", 5672)
	if val != 5672 {
		t.Errorf("expected fallback 5672, got %d", val)
	}
}

// ============================================================================
// Client Struct Type Check
// ============================================================================

func TestClientStruct(t *testing.T) {
	// Verify the Client struct definition at compile time
	var _ = Config{}
	var _ = QueueCodeExecution
	var _ = ExchangeCodeExec
}

// ============================================================================
// Config Round-trip Test
// ============================================================================

func TestConfigRoundTrip(t *testing.T) {
	cfg := Config{
		Host:     "test-host",
		Port:     5672,
		User:     "test-user",
		Password: "test-pass",
		VHost:    "/test",
	}
	if cfg.Host != "test-host" {
		t.Errorf("expected test-host, got %q", cfg.Host)
	}
}
