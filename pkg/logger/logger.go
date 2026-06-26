package logger

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"runtime"
	"sync"
	"time"
)

// Level represents a log level.
type Level int

const (
	DEBUG Level = iota
	INFO
	WARN
	ERROR
)

func (l Level) String() string {
	switch l {
	case DEBUG:
		return "DEBUG"
	case INFO:
		return "INFO"
	case WARN:
		return "WARN"
	case ERROR:
		return "ERROR"
	default:
		return "UNKNOWN"
	}
}

// ParseLevel parses a string log level.
func ParseLevel(s string) Level {
	switch s {
	case "debug", "DEBUG":
		return DEBUG
	case "info", "INFO":
		return INFO
	case "warn", "WARN", "warning", "WARNING":
		return WARN
	case "error", "ERROR":
		return ERROR
	default:
		return INFO
	}
}

// Logger provides structured JSON logging.
type Logger struct {
	mu      sync.Mutex
	output  io.Writer
	level   Level
	service string
}

// New creates a new Logger.
func New(service string, output io.Writer) *Logger {
	if output == nil {
		output = os.Stdout
	}
	return &Logger{
		output:  output,
		level:   ParseLevel(getEnv("LOG_LEVEL", "info")),
		service: service,
	}
}

// NewDefault creates a default logger writing to stdout.
func NewDefault(service string) *Logger {
	return New(service, os.Stdout)
}

// SetLevel sets the minimum log level.
func (l *Logger) SetLevel(level Level) {
	l.level = level
}

// WithRequestID returns a logger entry with request ID set.
func (l *Logger) WithRequestID(requestID string) *Entry {
	return &Entry{
		parent:    l,
		RequestID: requestID,
		Service:   l.service,
		Fields:    make(map[string]interface{}),
	}
}

// WithField returns a logger entry with a field set.
func (l *Logger) WithField(key string, value interface{}) *Entry {
	return &Entry{
		parent:  l,
		Service: l.service,
		Fields:  map[string]interface{}{key: value},
	}
}

// Debug logs a debug message.
func (l *Logger) Debug(msg string) {
	l.log(DEBUG, msg)
}

// Info logs an info message.
func (l *Logger) Info(msg string) {
	l.log(INFO, msg)
}

// Warn logs a warning message.
func (l *Logger) Warn(msg string) {
	l.log(WARN, msg)
}

// Error logs an error message.
func (l *Logger) Error(msg string, err error) {
	entry := &Entry{
		parent:  l,
		Service: l.service,
		Fields:  make(map[string]interface{}),
	}
	if err != nil {
		entry.ErrorMsg = err.Error()
		entry.Stack = captureStack()
	}
	entry.log(ERROR, msg)
}

// Debugf logs a formatted debug message.
func (l *Logger) Debugf(msg string, args ...interface{}) {
	l.log(DEBUG, fmt.Sprintf(msg, args...))
}

// Infof logs a formatted info message.
func (l *Logger) Infof(msg string, args ...interface{}) {
	l.log(INFO, fmt.Sprintf(msg, args...))
}

// Warnf logs a formatted warning message.
func (l *Logger) Warnf(msg string, args ...interface{}) {
	l.log(WARN, fmt.Sprintf(msg, args...))
}

// Errorf logs a formatted error message.
func (l *Logger) Errorf(msg string, args ...interface{}) {
	l.log(ERROR, fmt.Sprintf(msg, args...))
}

func (l *Logger) log(level Level, msg string) {
	if level < l.level {
		return
	}
	entry := Entry{
		Timestamp: time.Now().UTC().Format(time.RFC3339Nano),
		Level:     level.String(),
		Message:   msg,
		Service:   l.service,
	}
	writeEntry(l.output, entry)
}

// Log logs a performance metric.
func (l *Logger) Log(ctx context.Context, level Level, msg string, startTime time.Time, fields map[string]interface{}) {
	if level < l.level {
		return
	}
	entry := Entry{
		Timestamp: time.Now().UTC().Format(time.RFC3339Nano),
		Level:     level.String(),
		Message:   msg,
		Service:   l.service,
		Duration:  time.Since(startTime).String(),
	}
	if fields != nil {
		entry.Fields = fields
	}
	if reqID, ok := ctx.Value("requestID").(string); ok {
		entry.RequestID = reqID
	}
	writeEntry(l.output, entry)
}

// Entry represents a single log entry.
type Entry struct {
	parent    *Logger
	Timestamp string                 `json:"timestamp"`
	Level     string                 `json:"level"`
	Message   string                 `json:"message"`
	RequestID string                 `json:"request_id,omitempty"`
	Service   string                 `json:"service,omitempty"`
	ErrorMsg  string                 `json:"error,omitempty"`
	Stack     string                 `json:"stack,omitempty"`
	Duration  string                 `json:"duration,omitempty"`
	Fields    map[string]interface{} `json:"fields,omitempty"`
}

func (e *Entry) log(level Level, msg string) {
	if level < e.parent.level {
		return
	}
	e.Timestamp = time.Now().UTC().Format(time.RFC3339Nano)
	e.Level = level.String()
	e.Message = msg
	writeEntry(e.parent.output, *e)
}

// WithField adds a field to the log entry.
func (e *Entry) WithField(key string, value interface{}) *Entry {
	if e.Fields == nil {
		e.Fields = make(map[string]interface{})
	}
	e.Fields[key] = value
	return e
}

// Debug logs at debug level.
func (e *Entry) Debug(msg string) {
	e.log(DEBUG, msg)
}

// Info logs at info level.
func (e *Entry) Info(msg string) {
	e.log(INFO, msg)
}

// Warn logs at warn level.
func (e *Entry) Warn(msg string) {
	e.log(WARN, msg)
}

// Error logs at error level.
func (e *Entry) Error(msg string) {
	e.log(ERROR, msg)
}

func writeEntry(w io.Writer, entry Entry) {
	data, err := json.Marshal(entry)
	if err != nil {
		fmt.Fprintf(os.Stderr, "logger marshal error: %v\n", err)
		return
	}
	w.Write(data)
	w.Write([]byte("\n"))
}

func captureStack() string {
	buf := make([]byte, 4096)
	n := runtime.Stack(buf, false)
	return string(buf[:n])
}

func getEnv(key, fallback string) string {
	if v, ok := os.LookupEnv(key); ok {
		return v
	}
	return fallback
}
