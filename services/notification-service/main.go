package main

import (
	"bytes"
	"context"
	"crypto/tls"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"net/smtp"
	"os"
	"os/signal"
	"strconv"
	"strings"
	"sync"
	"syscall"
	"time"

	amqp "github.com/rabbitmq/amqp091-go"

	"coding-challange/pkg/health"
	"coding-challange/pkg/rabbitmq"
)

const (
	ServiceName = "notification-service"
	ServicePort = "9108"
)

// ---------------------------------------------------------------------------
// Notification Event Types
// ---------------------------------------------------------------------------

// NotificationEvent is the generic envelope for all notification messages
// consumed from the RabbitMQ 'notifications' queue.
type NotificationEvent struct {
	Type      string          `json:"type"`
	Timestamp time.Time       `json:"timestamp"`
	Payload   json.RawMessage `json:"payload"`
}

// SubmissionAcceptedPayload is sent when a user's submission is accepted.
type SubmissionAcceptedPayload struct {
	UserID         string `json:"user_id"`
	Email          string `json:"email"`
	Username       string `json:"username"`
	ProblemID      int    `json:"problem_id"`
	ProblemTitle   string `json:"problem_title"`
	SubmissionID   string `json:"submission_id"`
	Score          int    `json:"score"`
	ExecutionTimeMs int   `json:"execution_time_ms"`
	MemoryUsedKb   int    `json:"memory_used_kb"`
}

// NewProblemPayload is sent when a new problem is published.
type NewProblemPayload struct {
	ProblemID   int    `json:"problem_id"`
	Title       string `json:"title"`
	Difficulty  string `json:"difficulty"`
	Category    string `json:"category"`
	Description string `json:"description"`
}

// WeeklyLeaderboardPayload is sent with the weekly leaderboard summary.
type WeeklyLeaderboardPayload struct {
	Recipients []LeaderboardRecipient `json:"recipients"`
	WeekOf     string                 `json:"week_of"`
	TopUsers   []LeaderboardUser      `json:"top_users"`
	TotalUsers int                    `json:"total_users"`
}

// LeaderboardRecipient holds individual user data for personalised emails.
type LeaderboardRecipient struct {
	UserID   string `json:"user_id"`
	Email    string `json:"email"`
	Username string `json:"username"`
	Rank     int    `json:"rank"`
	Score    int    `json:"score"`
}

// LeaderboardUser is a condensed entry for the top-N display.
type LeaderboardUser struct {
	Rank     int    `json:"rank"`
	Username string `json:"username"`
	Score    int    `json:"score"`
}

// ---------------------------------------------------------------------------
// SMTP Configuration
// ---------------------------------------------------------------------------

// SMTPConfig holds the SMTP server configuration, populated from env vars.
type SMTPConfig struct {
	Host     string
	Port     int
	User     string
	Pass     string
	FromAddr string
	FromName string
}

// Addr returns the server address in "host:port" format.
func (c SMTPConfig) Addr() string {
	return fmt.Sprintf("%s:%d", c.Host, c.Port)
}

func loadSMTPConfig() SMTPConfig {
	port, _ := strconv.Atoi(getEnv("SMTP_PORT", "587"))
	return SMTPConfig{
		Host:     getEnv("SMTP_HOST", "smtp.example.com"),
		Port:     port,
		User:     getEnv("SMTP_USER", ""),
		Pass:     getEnv("SMTP_PASS", ""),
		FromAddr: getEnv("SMTP_FROM_ADDR", "noreply@codingchallange.com"),
		FromName: getEnv("SMTP_FROM_NAME", "Coding Challenge"),
	}
}

// ---------------------------------------------------------------------------
// Email Templates
// ---------------------------------------------------------------------------

// buildSubmissionAcceptedEmail renders the HTML email body for an accepted
// submission notification.
func buildSubmissionAcceptedEmail(p SubmissionAcceptedPayload, cfg SMTPConfig) (subject, body string) {
	subject = fmt.Sprintf("✓ Submission Accepted — %s", p.ProblemTitle)

	body = fmt.Sprintf(`<!DOCTYPE html>
<html>
<head><meta charset="utf-8"></head>
<body style="font-family: -apple-system, BlinkMacSystemFont, 'Segoe UI', Roboto, sans-serif; background: #f5f5f5; margin: 0; padding: 24px;">
<div style="max-width: 560px; margin: 0 auto; background: #fff; border-radius: 8px; overflow: hidden; box-shadow: 0 1px 3px rgba(0,0,0,.1);">
<div style="background: #10b981; padding: 24px; text-align: center;">
<span style="font-size: 40px;">✅</span>
<h1 style="color: #fff; margin: 8px 0 0; font-size: 20px;">Submission Accepted!</h1>
</div>
<div style="padding: 24px;">
<p style="color: #374151; font-size: 15px; line-height: 1.6;">Hi <strong>%s</strong>,</p>
<p style="color: #374151; font-size: 15px; line-height: 1.6;">Your solution for <strong>%s</strong> has been accepted.</p>
<table style="width: 100%%; border-collapse: collapse; margin: 16px 0;">
<tr><td style="padding: 8px; color: #6b7280; font-size: 14px;">Score</td><td style="padding: 8px; text-align: right; font-weight: 600; color: #10b981;">%d pts</td></tr>
<tr style="background: #f9fafb;"><td style="padding: 8px; color: #6b7280; font-size: 14px;">Execution Time</td><td style="padding: 8px; text-align: right; font-weight: 600;">%d ms</td></tr>
<tr><td style="padding: 8px; color: #6b7280; font-size: 14px;">Memory Used</td><td style="padding: 8px; text-align: right; font-weight: 600;">%d KB</td></tr>
</table>
<p style="color: #9ca3af; font-size: 13px;">Keep solving to climb the leaderboard!</p>
</div>
</div>
</body>
</html>`,
		p.Username, p.ProblemTitle, p.Score, p.ExecutionTimeMs, p.MemoryUsedKb)
	return
}

// buildNewProblemEmail renders the HTML email body for a new problem announcement.
func buildNewProblemEmail(p NewProblemPayload, cfg SMTPConfig) (subject, body string) {
	subject = fmt.Sprintf("New Problem: %s [%s]", p.Title, strings.ToUpper(p.Difficulty))

	difficultyColor := "#10b981" // easy (green)
	switch strings.ToLower(p.Difficulty) {
	case "medium":
		difficultyColor = "#f59e0b"
	case "hard":
		difficultyColor = "#ef4444"
	}

	body = fmt.Sprintf(`<!DOCTYPE html>
<html>
<head><meta charset="utf-8"></head>
<body style="font-family: -apple-system, BlinkMacSystemFont, 'Segoe UI', Roboto, sans-serif; background: #f5f5f5; margin: 0; padding: 24px;">
<div style="max-width: 560px; margin: 0 auto; background: #fff; border-radius: 8px; overflow: hidden; box-shadow: 0 1px 3px rgba(0,0,0,.1);">
<div style="background: #3b82f6; padding: 24px; text-align: center;">
<span style="font-size: 40px;">🚀</span>
<h1 style="color: #fff; margin: 8px 0 0; font-size: 20px;">New Problem Available</h1>
</div>
<div style="padding: 24px;">
<h2 style="color: #111827; font-size: 18px; margin: 0 0 8px;">%s</h2>
<span style="display: inline-block; background: %s; color: #fff; padding: 2px 10px; border-radius: 12px; font-size: 12px; font-weight: 600;">%s</span>
<span style="display: inline-block; background: #e5e7eb; color: #374151; padding: 2px 10px; border-radius: 12px; font-size: 12px; margin-left: 6px;">%s</span>
<p style="color: #6b7280; font-size: 14px; line-height: 1.6; margin-top: 12px;">%s</p>
</div>
</div>
</body>
</html>`,
		p.Title, difficultyColor, strings.ToUpper(p.Difficulty), p.Category, p.Description)
	return
}

// buildWeeklyLeaderboardEmail renders the HTML email body for the weekly
// leaderboard summary.
func buildWeeklyLeaderboardEmail(r LeaderboardRecipient, week string, top []LeaderboardUser, total int, cfg SMTPConfig) (subject, body string) {
	subject = fmt.Sprintf("Weekly Leaderboard — You're #%d!", r.Rank)

	rows := ""
	for _, u := range top {
		medal := ""
		switch u.Rank {
		case 1:
			medal = "🥇"
		case 2:
			medal = "🥈"
		case 3:
			medal = "🥉"
		default:
			medal = fmt.Sprintf("#%d", u.Rank)
		}
		highlight := ""
		if u.Username == r.Username {
			highlight = " style=\"font-weight: 700; background: #fef3c7;\""
		}
		rows += fmt.Sprintf(`<tr%s><td style="padding: 8px;">%s</td><td style="padding: 8px;">%s</td><td style="padding: 8px; text-align: right;">%d</td></tr>`, highlight, medal, u.Username, u.Score)
	}

	body = fmt.Sprintf(`<!DOCTYPE html>
<html>
<head><meta charset="utf-8"></head>
<body style="font-family: -apple-system, BlinkMacSystemFont, 'Segoe UI', Roboto, sans-serif; background: #f5f5f5; margin: 0; padding: 24px;">
<div style="max-width: 560px; margin: 0 auto; background: #fff; border-radius: 8px; overflow: hidden; box-shadow: 0 1px 3px rgba(0,0,0,.1);">
<div style="background: #8b5cf6; padding: 24px; text-align: center;">
<span style="font-size: 40px;">🏆</span>
<h1 style="color: #fff; margin: 8px 0 0; font-size: 20px;">Weekly Leaderboard</h1>
<p style="color: #c4b5fd; font-size: 14px; margin: 4px 0 0;">Week of %s</p>
</div>
<div style="padding: 24px;">
<p style="color: #374151; font-size: 15px; line-height: 1.6;">Hi <strong>%s</strong>,</p>
<p style="color: #374151; font-size: 15px; line-height: 1.6;">You finished at <strong>#%d</strong> with <strong>%d points</strong> this week!</p>
<table style="width: 100%%; border-collapse: collapse; margin: 16px 0;">
<thead><tr style="background: #f3f0ff;"><th style="padding: 8px; text-align: left; font-size: 13px; color: #6b7280;">Rank</th><th style="padding: 8px; text-align: left; font-size: 13px; color: #6b7280;">User</th><th style="padding: 8px; text-align: right; font-size: 13px; color: #6b7280;">Score</th></tr></thead>
<tbody>%s</tbody>
</table>
<p style="color: #9ca3af; font-size: 13px;">%d users participated this week. See you next round!</p>
</div>
</div>
</body>
</html>`,
		week, r.Username, r.Rank, r.Score, rows, total)
	return
}

// ---------------------------------------------------------------------------
// Email Sender
// ---------------------------------------------------------------------------

// EmailSender handles sending emails via SMTP.
type EmailSender struct {
	cfg     SMTPConfig
	mu      sync.Mutex
	metrics EmailMetrics
}

// EmailMetrics tracks aggregate email sending statistics.
type EmailMetrics struct {
	TotalSent   int            `json:"total_sent"`
	TotalFailed int            `json:"total_failed"`
	ByType      map[string]int `json:"by_type"`
}

// NewEmailSender creates a new EmailSender with the given SMTP config.
func NewEmailSender(cfg SMTPConfig) *EmailSender {
	return &EmailSender{
		cfg: cfg,
		metrics: EmailMetrics{
			ByType: make(map[string]int),
		},
	}
}

// Send sends an email with the given recipient, subject, and HTML body.
func (s *EmailSender) Send(ctx context.Context, to, subject, body string, eventType string) error {
	from := fmt.Sprintf("%s <%s>", s.cfg.FromName, s.cfg.FromAddr)

	var msgBuf bytes.Buffer
	msgBuf.WriteString(fmt.Sprintf("From: %s\r\n", from))
	msgBuf.WriteString(fmt.Sprintf("To: %s\r\n", to))
	msgBuf.WriteString(fmt.Sprintf("Subject: %s\r\n", subject))
	msgBuf.WriteString("MIME-Version: 1.0\r\n")
	msgBuf.WriteString("Content-Type: text/html; charset=\"utf-8\"\r\n")
	msgBuf.WriteString("\r\n")
	msgBuf.WriteString(body)

	addr := s.cfg.Addr()
	auth := smtp.PlainAuth("", s.cfg.User, s.cfg.Pass, s.cfg.Host)

	var err error
	if s.cfg.Port == 465 {
		// SMTPS — use explicit TLS on port 465
		err = s.sendTLS(addr, auth, s.cfg.FromAddr, []string{to}, msgBuf.Bytes())
	} else {
		// STARTTLS on port 587 or 25
		err = smtp.SendMail(addr, auth, s.cfg.FromAddr, []string{to}, msgBuf.Bytes())
	}

	s.mu.Lock()
	if err != nil {
		s.metrics.TotalFailed++
		log.Printf("Failed to send email to %s (type=%s): %v", to, eventType, err)
	} else {
		s.metrics.TotalSent++
		s.metrics.ByType[eventType]++
		log.Printf("Email sent to %s (type=%s, subject=%s)", to, eventType, subject)
	}
	s.mu.Unlock()

	return err
}

// sendTLS sends email over an explicit TLS connection (port 465).
func (s *EmailSender) sendTLS(addr string, auth smtp.Auth, from string, to []string, msg []byte) error {
	tlsCfg := &tls.Config{
		ServerName: s.cfg.Host,
		MinVersion: tls.VersionTLS12,
	}

	conn, err := tls.Dial("tcp", addr, tlsCfg)
	if err != nil {
		return fmt.Errorf("tls dial: %w", err)
	}
	defer conn.Close()

	client, err := smtp.NewClient(conn, s.cfg.Host)
	if err != nil {
		return fmt.Errorf("smtp client: %w", err)
	}
	defer client.Close()

	if err = client.Auth(auth); err != nil {
		return fmt.Errorf("smtp auth: %w", err)
	}
	if err = client.Mail(from); err != nil {
		return fmt.Errorf("smtp mail: %w", err)
	}
	for _, rcpt := range to {
		if err = client.Rcpt(rcpt); err != nil {
			return fmt.Errorf("smtp rcpt %s: %w", rcpt, err)
		}
	}
	w, err := client.Data()
	if err != nil {
		return fmt.Errorf("smtp data: %w", err)
	}
	if _, err = w.Write(msg); err != nil {
		return fmt.Errorf("smtp write: %w", err)
	}
	if err = w.Close(); err != nil {
		return fmt.Errorf("smtp close: %w", err)
	}
	return client.Quit()
}

// Metrics returns a snapshot of the email metrics.
func (s *EmailSender) Metrics() EmailMetrics {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.metrics
}

// ---------------------------------------------------------------------------
// Notification Consumer
// ---------------------------------------------------------------------------

// NotificationConsumer listens on the 'notifications' queue and dispatches
// events to the appropriate email handler.
type NotificationConsumer struct {
	emailSender *EmailSender
	rabbitMQ    *rabbitmq.Client
	metrics     EmailMetrics
}

// NewNotificationConsumer creates a new consumer.
func NewNotificationConsumer(rmq *rabbitmq.Client, sender *EmailSender) *NotificationConsumer {
	return &NotificationConsumer{
		emailSender: sender,
		rabbitMQ:    rmq,
	}
}

// Start begins consuming messages from the 'notifications' queue.
func (nc *NotificationConsumer) Start(ctx context.Context) error {
	log.Println("[NOTIFICATION CONSUMER] Starting...")
	return nc.rabbitMQ.Consume(ctx, rabbitmq.QueueNotifications, nc.handleMessage)
}

// handleMessage processes a single notification message from the queue.
func (nc *NotificationConsumer) handleMessage(msg amqp.Delivery) error {
	var event NotificationEvent
	if err := json.Unmarshal(msg.Body, &event); err != nil {
		log.Printf("Invalid notification message (malformed JSON): %v", err)
		msg.Ack(false) // Discard malformed messages
		return nil
	}
	msg.Ack(false)

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	switch event.Type {
	case "submission_accepted":
		return nc.handleSubmissionAccepted(ctx, event.Payload)
	case "new_problem":
		return nc.handleNewProblem(ctx, event.Payload)
	case "weekly_leaderboard":
		return nc.handleWeeklyLeaderboard(ctx, event.Payload)
	default:
		log.Printf("Unknown notification type: %s", event.Type)
		return nil
	}
}

// handleSubmissionAccepted processes a submission_accepted event.
func (nc *NotificationConsumer) handleSubmissionAccepted(ctx context.Context, raw json.RawMessage) error {
	var p SubmissionAcceptedPayload
	if err := json.Unmarshal(raw, &p); err != nil {
		return fmt.Errorf("unmarshal submission_accepted payload: %w", err)
	}
	if p.Email == "" {
		log.Printf("Skipping submission_accepted for user %s: no email", p.UserID)
		return nil
	}

	subject, body := buildSubmissionAcceptedEmail(p, nc.emailSender.cfg)
	return nc.emailSender.Send(ctx, p.Email, subject, body, "submission_accepted")
}

// handleNewProblem processes a new_problem event.
func (nc *NotificationConsumer) handleNewProblem(ctx context.Context, raw json.RawMessage) error {
	var p NewProblemPayload
	if err := json.Unmarshal(raw, &p); err != nil {
		return fmt.Errorf("unmarshal new_problem payload: %w", err)
	}
	if p.Title == "" {
		log.Printf("Skipping new_problem: missing title")
		return nil
	}

	// For new problem notifications, the payload may include individual
	// recipients or be broadcast-only. If no per-user data, we skip email
	// delivery here — the publisher is expected to send one message per
	// recipient with their email.
	return nil
}

// handleWeeklyLeaderboard processes a weekly_leaderboard event.
func (nc *NotificationConsumer) handleWeeklyLeaderboard(ctx context.Context, raw json.RawMessage) error {
	var p WeeklyLeaderboardPayload
	if err := json.Unmarshal(raw, &p); err != nil {
		return fmt.Errorf("unmarshal weekly_leaderboard payload: %w", err)
	}

	for _, r := range p.Recipients {
		if r.Email == "" {
			continue
		}
		subject, body := buildWeeklyLeaderboardEmail(r, p.WeekOf, p.TopUsers, p.TotalUsers, nc.emailSender.cfg)
		if err := nc.emailSender.Send(ctx, r.Email, subject, body, "weekly_leaderboard"); err != nil {
			log.Printf("Failed to send leaderboard email to %s: %v", r.Email, err)
		}
	}
	return nil
}

// ---------------------------------------------------------------------------
// Server
// ---------------------------------------------------------------------------

// Server holds all dependencies for the notification service.
type Server struct {
	config     SMTPConfig
	rabbitMQ   *rabbitmq.Client
	emailSender *EmailSender
	consumer   *NotificationConsumer
	health     *health.Manager
	httpServer *http.Server
}

// NewServer creates a new Server and wires its dependencies.
func NewServer() (*Server, error) {
	smtpCfg := loadSMTPConfig()
	sender := NewEmailSender(smtpCfg)

	rmq, err := rabbitmq.NewFromEnv()
	if err != nil {
		return nil, fmt.Errorf("rabbitmq connect: %w", err)
	}

	consumer := NewNotificationConsumer(rmq, sender)

	svc := &Server{
		config:      smtpCfg,
		rabbitMQ:    rmq,
		emailSender: sender,
		consumer:    consumer,
		health:      health.NewManager(),
	}

	// Register health checkers.
	svc.health.Register(health.NewRabbitMQChecker("rabbitmq", func(ctx context.Context) error {
		return rmq.HealthCheck(ctx)
	}))

	svc.health.Register(health.NewFunctionChecker("smtp", func(ctx context.Context) error {
		// Liveness: check SMTP config is minimally valid
		if svc.config.Host == "" || svc.config.Port == 0 {
			return fmt.Errorf("smtp not configured (SMTP_HOST/SMTP_PORT)")
		}
		return nil
	}))

	return svc, nil
}

// Start starts the HTTP server and RabbitMQ consumer.
func (s *Server) Start(ctx context.Context) error {
	// Start RabbitMQ consumer in background.
	go func() {
		log.Println("Starting notification consumer...")
		if err := s.consumer.Start(ctx); err != nil {
			log.Printf("Notification consumer exited: %v", err)
		}
	}()

	// HTTP server.
	mux := http.NewServeMux()

	// Health endpoint.
	healthHandler := s.health.HTTPHandler()
	mux.HandleFunc("/health", healthHandler.HealthHandler)
	mux.HandleFunc("/liveness", healthHandler.LivenessHandler)
	mux.HandleFunc("/readiness", healthHandler.ReadinessHandler)

	// Metrics endpoint (internal).
	mux.HandleFunc("/metrics", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(s.emailSender.Metrics())
	})

	s.httpServer = &http.Server{
		Addr:         ":" + ServicePort,
		Handler:      mux,
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 10 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	log.Printf("Notification service starting on :%s", ServicePort)
	if err := s.httpServer.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		return fmt.Errorf("http server: %w", err)
	}
	return nil
}

// Shutdown gracefully stops the server.
func (s *Server) Shutdown(ctx context.Context) error {
	log.Println("Shutting down notification service...")

	if s.httpServer != nil {
		if err := s.httpServer.Shutdown(ctx); err != nil {
			log.Printf("HTTP server shutdown error: %v", err)
		}
	}
	if s.rabbitMQ != nil {
		s.rabbitMQ.Close()
	}
	return nil
}

// ---------------------------------------------------------------------------
// Main
// ---------------------------------------------------------------------------

func main() {
	log.SetFlags(log.LstdFlags | log.Lshortfile)
	log.Printf("Starting %s...", ServiceName)

	server, err := NewServer()
	if err != nil {
		log.Fatalf("Failed to create server: %v", err)
	}

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Trap OS signals for graceful shutdown.
	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM)

	go func() {
		sig := <-sigCh
		log.Printf("Received signal: %v", sig)
		cancel()
	}()

	go func() {
		if err := server.Start(ctx); err != nil && err != http.ErrServerClosed {
			log.Fatalf("Server error: %v", err)
		}
	}()

	<-ctx.Done()

	shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer shutdownCancel()

	if err := server.Shutdown(shutdownCtx); err != nil {
		log.Printf("Shutdown error: %v", err)
	}
	log.Println("Notification service stopped.")
}

// ---------------------------------------------------------------------------
// Helpers
// ---------------------------------------------------------------------------

func getEnv(key, fallback string) string {
	if v, ok := os.LookupEnv(key); ok {
		return v
	}
	return fallback
}
