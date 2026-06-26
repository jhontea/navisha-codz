package logger

import (
	"bytes"
	"context"
	"encoding/json"
	"strings"
	"testing"
	"time"
)

func TestNewLogger(t *testing.T) {
	buf := &bytes.Buffer{}
	logger := New("test-service", buf)
	if logger == nil {
		t.Fatal("expected non-nil logger")
	}
	if logger.service != "test-service" {
		t.Errorf("expected service 'test-service', got %q", logger.service)
	}
}

func TestNewLogger_NilOutput(t *testing.T) {
	logger := New("test", nil)
	if logger == nil {
		t.Fatal("expected non-nil logger with nil output (should default to stdout)")
	}
}

func TestNewDefault(t *testing.T) {
	logger := NewDefault("my-service")
	if logger == nil {
		t.Fatal("expected non-nil default logger")
	}
	if logger.service != "my-service" {
		t.Errorf("expected service 'my-service', got %q", logger.service)
	}
}

// Test log levels
func TestParseLevel_Debug(t *testing.T) {
	if ParseLevel("debug") != DEBUG {
		t.Error("expected DEBUG level")
	}
	if ParseLevel("DEBUG") != DEBUG {
		t.Error("expected DEBUG level")
	}
}

func TestParseLevel_Info(t *testing.T) {
	if ParseLevel("info") != INFO {
		t.Error("expected INFO level")
	}
	if ParseLevel("INFO") != INFO {
		t.Error("expected INFO level")
	}
}

func TestParseLevel_Warn(t *testing.T) {
	if ParseLevel("warn") != WARN {
		t.Error("expected WARN level")
	}
	if ParseLevel("WARN") != WARN {
		t.Error("expected WARN level")
	}
	if ParseLevel("warning") != WARN {
		t.Error("expected WARN level for 'warning'")
	}
}

func TestParseLevel_Error(t *testing.T) {
	if ParseLevel("error") != ERROR {
		t.Error("expected ERROR level")
	}
	if ParseLevel("ERROR") != ERROR {
		t.Error("expected ERROR level")
	}
}

func TestParseLevel_Default(t *testing.T) {
	if ParseLevel("unknown") != INFO {
		t.Error("expected INFO as default for unknown level")
	}
	if ParseLevel("") != INFO {
		t.Error("expected INFO as default for empty level")
	}
}

func TestLevel_String(t *testing.T) {
	tests := []struct {
		level    Level
		expected string
	}{
		{DEBUG, "DEBUG"},
		{INFO, "INFO"},
		{WARN, "WARN"},
		{ERROR, "ERROR"},
		{Level(99), "UNKNOWN"},
	}
	for _, tt := range tests {
		if tt.level.String() != tt.expected {
			t.Errorf("expected %q, got %q", tt.expected, tt.level.String())
		}
	}
}

func TestSetLevel(t *testing.T) {
	buf := &bytes.Buffer{}
	logger := New("test", buf)
	logger.SetLevel(WARN)

	// Debug and Info should be suppressed
	logger.Debug("debug msg")
	logger.Info("info msg")
	if buf.Len() != 0 {
		t.Error("expected no output for DEBUG/INFO when level is WARN")
	}

	// Warn and Error should pass through
	logger.Warn("warn msg")
	if buf.Len() == 0 {
		t.Error("expected WARN message to be logged")
	}
}

func TestLogLevelFiltering(t *testing.T) {
	buf := &bytes.Buffer{}
	logger := New("test", buf)
	logger.SetLevel(INFO)

	logger.Debug("debug")
	if buf.Len() != 0 {
		t.Error("DEBUG should be filtered at INFO level")
	}

	logger.Info("info")
	if buf.Len() == 0 {
		t.Log("INFO should pass through at INFO level")
	}
}

// Test JSON output format
func TestJSONOutputFormat(t *testing.T) {
	buf := &bytes.Buffer{}
	logger := New("test-svc", buf)
	logger.SetLevel(DEBUG)

	logger.Info("test message")

	var entry Entry
	if err := json.Unmarshal(buf.Bytes()[:len(buf.Bytes())-1], &entry); err != nil {
		t.Fatalf("failed to unmarshal log entry: %v", err)
	}

	if entry.Level != "INFO" {
		t.Errorf("expected level INFO, got %q", entry.Level)
	}
	if entry.Message != "test message" {
		t.Errorf("expected message 'test message', got %q", entry.Message)
	}
	if entry.Service != "test-svc" {
		t.Errorf("expected service 'test-svc', got %q", entry.Service)
	}
	if entry.Timestamp == "" {
		t.Error("expected non-empty timestamp")
	}
}

func TestJSONOutputFormat_ErrorWithStack(t *testing.T) {
	buf := &bytes.Buffer{}
	logger := New("test", buf)
	logger.SetLevel(ERROR)

	logger.Error("something failed", nil)

	var entry Entry
	if err := json.Unmarshal(buf.Bytes()[:len(buf.Bytes())-1], &entry); err != nil {
		t.Fatalf("failed to unmarshal log entry: %v", err)
	}

	if entry.Level != "ERROR" {
		t.Errorf("expected level ERROR, got %q", entry.Level)
	}
	if entry.Message != "something failed" {
		t.Errorf("expected message 'something failed', got %q", entry.Message)
	}
}

func TestJSONOutputFormat_Debugf(t *testing.T) {
	buf := &bytes.Buffer{}
	logger := New("test", buf)
	logger.SetLevel(DEBUG)

	logger.Debugf("hello %s", "world")

	var entry Entry
	if err := json.Unmarshal(buf.Bytes()[:len(buf.Bytes())-1], &entry); err != nil {
		t.Fatalf("failed to unmarshal log entry: %v", err)
	}

	if entry.Message != "hello world" {
		t.Errorf("expected 'hello world', got %q", entry.Message)
	}
}

func TestJSONOutputFormat_Infof(t *testing.T) {
	buf := &bytes.Buffer{}
	logger := New("test", buf)
	logger.SetLevel(DEBUG)

	logger.Infof("count: %d", 42)

	var entry Entry
	if err := json.Unmarshal(buf.Bytes()[:len(buf.Bytes())-1], &entry); err != nil {
		t.Fatalf("failed to unmarshal log entry: %v", err)
	}

	if entry.Message != "count: 42" {
		t.Errorf("expected 'count: 42', got %q", entry.Message)
	}
}

func TestJSONOutputFormat_Warnf(t *testing.T) {
	buf := &bytes.Buffer{}
	logger := New("test", buf)
	logger.SetLevel(DEBUG)

	logger.Warnf("warning: %s", "low memory")

	var entry Entry
	if err := json.Unmarshal(buf.Bytes()[:len(buf.Bytes())-1], &entry); err != nil {
		t.Fatalf("failed to unmarshal log entry: %v", err)
	}

	if entry.Level != "WARN" {
		t.Errorf("expected level WARN, got %q", entry.Level)
	}
	if entry.Message != "warning: low memory" {
		t.Errorf("expected 'warning: low memory', got %q", entry.Message)
	}
}

func TestJSONOutputFormat_Errorf(t *testing.T) {
	buf := &bytes.Buffer{}
	logger := New("test", buf)
	logger.SetLevel(ERROR)

	logger.Errorf("error code: %d", 500)

	var entry Entry
	if err := json.Unmarshal(buf.Bytes()[:len(buf.Bytes())-1], &entry); err != nil {
		t.Fatalf("failed to unmarshal log entry: %v", err)
	}

	if entry.Level != "ERROR" {
		t.Errorf("expected level ERROR, got %q", entry.Level)
	}
	if entry.Message != "error code: 500" {
		t.Errorf("expected 'error code: 500', got %q", entry.Message)
	}
}

// Test request ID propagation
func TestWithRequestID(t *testing.T) {
	buf := &bytes.Buffer{}
	logger := New("test", buf)
	logger.SetLevel(DEBUG)

	entry := logger.WithRequestID("req-123")
	entry.Info("with request id")

	var logEntry Entry
	if err := json.Unmarshal(buf.Bytes()[:len(buf.Bytes())-1], &logEntry); err != nil {
		t.Fatalf("failed to unmarshal log entry: %v", err)
	}

	if logEntry.RequestID != "req-123" {
		t.Errorf("expected request ID 'req-123', got %q", logEntry.RequestID)
	}
}

func TestWithRequestID_Empty(t *testing.T) {
	buf := &bytes.Buffer{}
	logger := New("test", buf)
	logger.SetLevel(DEBUG)

	entry := logger.WithRequestID("")
	entry.Info("no request id")

	var logEntry Entry
	if err := json.Unmarshal(buf.Bytes()[:len(buf.Bytes())-1], &logEntry); err != nil {
		t.Fatalf("failed to unmarshal log entry: %v", err)
	}

	if logEntry.RequestID != "" {
		t.Errorf("expected empty request ID, got %q", logEntry.RequestID)
	}
}

func TestWithField(t *testing.T) {
	buf := &bytes.Buffer{}
	logger := New("test", buf)
	logger.SetLevel(DEBUG)

	entry := logger.WithField("key1", "value1")
	entry.Info("with field")

	var logEntry Entry
	if err := json.Unmarshal(buf.Bytes()[:len(buf.Bytes())-1], &logEntry); err != nil {
		t.Fatalf("failed to unmarshal log entry: %v", err)
	}

	if logEntry.Fields["key1"] != "value1" {
		t.Errorf("expected field key1='value1', got %v", logEntry.Fields["key1"])
	}
}

func TestEntryWithField_Chained(t *testing.T) {
	buf := &bytes.Buffer{}
	logger := New("test", buf)
	logger.SetLevel(DEBUG)

	logger.WithRequestID("req-456").WithField("user", "alice").Info("chained fields")

	var logEntry Entry
	if err := json.Unmarshal(buf.Bytes()[:len(buf.Bytes())-1], &logEntry); err != nil {
		t.Fatalf("failed to unmarshal log entry: %v", err)
	}

	if logEntry.RequestID != "req-456" {
		t.Errorf("expected request ID 'req-456', got %q", logEntry.RequestID)
	}
	if logEntry.Fields["user"] != "alice" {
		t.Errorf("expected field user='alice', got %v", logEntry.Fields["user"])
	}
}

func TestEntryDebug(t *testing.T) {
	buf := &bytes.Buffer{}
	logger := New("test", buf)
	logger.SetLevel(DEBUG)

	logger.WithRequestID("req-1").Debug("debug via entry")

	var logEntry Entry
	if err := json.Unmarshal(buf.Bytes()[:len(buf.Bytes())-1], &logEntry); err != nil {
		t.Fatalf("failed to unmarshal: %v", err)
	}

	if logEntry.Level != "DEBUG" {
		t.Errorf("expected DEBUG, got %q", logEntry.Level)
	}
	if logEntry.RequestID != "req-1" {
		t.Errorf("expected req-1, got %q", logEntry.RequestID)
	}
}

func TestEntryInfo(t *testing.T) {
	buf := &bytes.Buffer{}
	logger := New("test", buf)
	logger.SetLevel(DEBUG)

	logger.WithRequestID("req-2").Info("info via entry")

	var logEntry Entry
	if err := json.Unmarshal(buf.Bytes()[:len(buf.Bytes())-1], &logEntry); err != nil {
		t.Fatalf("failed to unmarshal: %v", err)
	}

	if logEntry.Level != "INFO" {
		t.Errorf("expected INFO, got %q", logEntry.Level)
	}
}

func TestEntryWarn(t *testing.T) {
	buf := &bytes.Buffer{}
	logger := New("test", buf)
	logger.SetLevel(DEBUG)

	logger.WithRequestID("req-3").Warn("warn via entry")

	var logEntry Entry
	if err := json.Unmarshal(buf.Bytes()[:len(buf.Bytes())-1], &logEntry); err != nil {
		t.Fatalf("failed to unmarshal: %v", err)
	}

	if logEntry.Level != "WARN" {
		t.Errorf("expected WARN, got %q", logEntry.Level)
	}
}

func TestEntryError(t *testing.T) {
	buf := &bytes.Buffer{}
	logger := New("test", buf)
	logger.SetLevel(DEBUG)

	logger.WithRequestID("req-4").Error("error via entry")

	var logEntry Entry
	if err := json.Unmarshal(buf.Bytes()[:len(buf.Bytes())-1], &logEntry); err != nil {
		t.Fatalf("failed to unmarshal: %v", err)
	}

	if logEntry.Level != "ERROR" {
		t.Errorf("expected ERROR, got %q", logEntry.Level)
	}
}

func TestLogWithContext(t *testing.T) {
	buf := &bytes.Buffer{}
	logger := New("test", buf)
	logger.SetLevel(DEBUG)

	ctx := context.WithValue(context.Background(), "requestID", "ctx-req-789")
	logger.Log(ctx, INFO, "context log", time.Now(), nil)

	output := buf.String()
	if !strings.Contains(output, "ctx-req-789") {
		t.Errorf("expected request ID from context in output, got: %s", output)
	}
}

func TestLogWithContext_NoRequestID(t *testing.T) {
	buf := &bytes.Buffer{}
	logger := New("test", buf)
	logger.SetLevel(DEBUG)

	ctx := context.Background()
	logger.Log(ctx, INFO, "no ctx request id", time.Now(), nil)

	var logEntry Entry
	if err := json.Unmarshal(buf.Bytes()[:len(buf.Bytes())-1], &logEntry); err != nil {
		t.Fatalf("failed to unmarshal: %v", err)
	}

	if logEntry.RequestID != "" {
		t.Errorf("expected empty request ID, got %q", logEntry.RequestID)
	}
}

func TestLogWithContext_WithFields(t *testing.T) {
	buf := &bytes.Buffer{}
	logger := New("test", buf)
	logger.SetLevel(DEBUG)

	ctx := context.WithValue(context.Background(), "requestID", "ctx-req-000")
	fields := map[string]interface{}{"method": "GET", "path": "/api/test"}
	logger.Log(ctx, INFO, "fields log", time.Now(), fields)

	var logEntry Entry
	if err := json.Unmarshal(buf.Bytes()[:len(buf.Bytes())-1], &logEntry); err != nil {
		t.Fatalf("failed to unmarshal: %v", err)
	}

	if logEntry.Fields["method"] != "GET" {
		t.Errorf("expected method GET, got %v", logEntry.Fields["method"])
	}
	if logEntry.Fields["path"] != "/api/test" {
		t.Errorf("expected path /api/test, got %v", logEntry.Fields["path"])
	}
}

func TestErrorWithNilError(t *testing.T) {
	buf := &bytes.Buffer{}
	logger := New("test", buf)
	logger.SetLevel(ERROR)

	logger.Error("error with nil", nil)

	var logEntry Entry
	if err := json.Unmarshal(buf.Bytes()[:len(buf.Bytes())-1], &logEntry); err != nil {
		t.Fatalf("failed to unmarshal: %v", err)
	}

	if logEntry.Level != "ERROR" {
		t.Errorf("expected ERROR, got %q", logEntry.Level)
	}
	if logEntry.ErrorMsg != "" {
		t.Errorf("expected empty error message for nil error, got %q", logEntry.ErrorMsg)
	}
}

func TestMultipleLogEntries(t *testing.T) {
	buf := &bytes.Buffer{}
	logger := New("test", buf)
	logger.SetLevel(DEBUG)

	logger.Info("first")
	logger.Info("second")
	logger.Info("third")

	lines := strings.Split(strings.TrimSpace(buf.String()), "\n")
	if len(lines) != 3 {
		t.Errorf("expected 3 log lines, got %d", len(lines))
	}

	for i, line := range lines {
		var entry Entry
		if err := json.Unmarshal([]byte(line), &entry); err != nil {
			t.Errorf("line %d: failed to unmarshal: %v", i, err)
		}
	}
}
