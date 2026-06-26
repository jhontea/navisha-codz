package security

import (
	"crypto/hmac"
	"crypto/rand"
	"crypto/sha256"
	"crypto/subtle"
	"encoding/base32"
	"encoding/base64"
	"encoding/hex"
	"fmt"
	"regexp"
	"strings"
	"sync"
	"time"
	"unicode"
)

// ============================================================================
// Password Policy Enforcement
// ============================================================================

// PasswordPolicy defines password requirements.
type PasswordPolicy struct {
	MinLength        int
	MaxLength        int
	RequireUppercase bool
	RequireLowercase bool
	RequireDigit     bool
	RequireSpecial   bool
	MinUniqueChars   int
	BlacklistedWords []string
	DefaultBlacklist bool
}

// DefaultPasswordPolicy returns a strong default password policy.
func DefaultPasswordPolicy() *PasswordPolicy {
	return &PasswordPolicy{
		MinLength:        12,
		MaxLength:        128,
		RequireUppercase: true,
		RequireLowercase: true,
		RequireDigit:     true,
		RequireSpecial:   true,
		MinUniqueChars:   6,
		DefaultBlacklist: true,
	}
}

// PasswordValidationResult contains password validation results.
type PasswordValidationResult struct {
	Valid   bool     `json:"valid"`
	Errors  []string `json:"errors,omitempty"`
	Score   int      `json:"score"` // 0-100
	Strength string  `json:"strength"` // "weak", "fair", "good", "strong"
}

// ValidatePassword validates a password against the policy.
func ValidatePassword(policy *PasswordPolicy, password string) PasswordValidationResult {
	result := PasswordValidationResult{
		Valid:  true,
		Errors: make([]string, 0),
		Score:  0,
	}

	// Check length
	if len(password) < policy.MinLength {
		result.Errors = append(result.Errors, fmt.Sprintf("password must be at least %d characters", policy.MinLength))
		result.Valid = false
	}
	if len(password) > policy.MaxLength {
		result.Errors = append(result.Errors, fmt.Sprintf("password must be at most %d characters", policy.MaxLength))
		result.Valid = false
	}

	// Check character classes
	var hasUpper, hasLower, hasDigit, hasSpecial bool
	uniqueChars := make(map[rune]bool)
	for _, r := range password {
		uniqueChars[r] = true
		switch {
		case unicode.IsUpper(r):
			hasUpper = true
		case unicode.IsLower(r):
		hasLower = true
		case unicode.IsDigit(r):
			hasDigit = true
		case unicode.IsSymbol(r) || unicode.IsPunct(r):
			hasSpecial = true
		}
	}

	if policy.RequireUppercase && !hasUpper {
		result.Errors = append(result.Errors, "password must contain at least one uppercase letter")
		result.Valid = false
	}
	if policy.RequireLowercase && !hasLower {
		result.Errors = append(result.Errors, "password must contain at least one lowercase letter")
		result.Valid = false
	}
	if policy.RequireDigit && !hasDigit {
		result.Errors = append(result.Errors, "password must contain at least one digit")
		result.Valid = false
	}
	if policy.RequireSpecial && !hasSpecial {
		result.Errors = append(result.Errors, "password must contain at least one special character")
		result.Valid = false
	}

	if len(uniqueChars) < policy.MinUniqueChars {
		result.Errors = append(result.Errors, fmt.Sprintf("password must contain at least %d unique characters", policy.MinUniqueChars))
		result.Valid = false
	}

	// Check against common passwords
	if policy.DefaultBlacklist {
		if isCommonPassword(password) {
			result.Errors = append(result.Errors, "this password is too common, please choose a stronger password")
			result.Valid = false
		}
	}

	// Check custom blacklist
	for _, word := range policy.BlacklistedWords {
		if strings.Contains(strings.ToLower(password), strings.ToLower(word)) {
			result.Errors = append(result.Errors, fmt.Sprintf("password must not contain the word '%s'", word))
			result.Valid = false
		}
	}

	// Calculate score
	result.Score = calculatePasswordScore(password, hasUpper, hasLower, hasDigit, hasSpecial, len(uniqueChars))
	result.Strength = scoreToStrength(result.Score)

	return result
}

func calculatePasswordScore(password string, hasUpper, hasLower, hasDigit, hasSpecial bool, uniqueCount int) int {
	score := 0
	length := len(password)

	// Length contribution (max 30)
	if length >= 16 {
		score += 30
	} else if length >= 12 {
		score += 25
	} else if length >= 8 {
		score += 15
	} else {
		score += 5
	}

	// Character class contributions (max 40)
	if hasUpper {
		score += 10
	}
	if hasLower {
		score += 10
	}
	if hasDigit {
		score += 10
	}
	if hasSpecial {
		score += 10
	}

	// Unique characters (max 30)
	if uniqueCount >= 10 {
		score += 30
	} else if uniqueCount >= 7 {
		score += 20
	} else if uniqueCount >= 5 {
		score += 10
	} else {
		score += 5
	}

	return min(score, 100)
}

func scoreToStrength(score int) string {
	switch {
	case score >= 80:
		return "strong"
	case score >= 60:
		return "good"
	case score >= 40:
		return "fair"
	default:
		return "weak"
	}
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

// Common passwords list (subset for brevity — extend in production).
var commonPasswords = map[string]bool{
	"password": true, "password123": true, "123456": true, "123456789": true,
	"12345678": true, "admin": true, "qwerty": true, "abc123": true,
	"letmein": true, "welcome": true, "monkey": true, "master": true,
	"dragon": true, "login": true, "princess": true, "football": true,
	"sunshine": true, "iloveyou": true, "trustno1": true, "batman": true,
	"password1": true, "admin123": true, "root": true, "toor": true,
	"pass123": true, "pass@123": true, "123123": true, "654321": true,
}

func isCommonPassword(password string) bool {
	lower := strings.ToLower(password)
	return commonPasswords[lower]
}

// ============================================================================
// SQL Injection Detection
// ============================================================================

// SQLInjectionDetector detects potential SQL injection patterns.
type SQLInjectionDetector struct {
	patterns []*regexp.Regexp
}

// NewSQLInjectionDetector creates a new SQL injection detector.
func NewSQLInjectionDetector() *SQLInjectionDetector {
	patterns := []string{
		`(?i)(?:--|#|/\*)`,                                                            // SQL comments
		`(?i)(?:\bUNION\b\s+(?:\bALL\b\s+)?\bSELECT\b)`,                               // UNION SELECT
		`(?i)(?:\bSELECT\b.*\bFROM\b)`,                                                 // SELECT FROM
		`(?i)(?:\bINSERT\b.*\bINTO\b)`,                                                 // INSERT INTO
		`(?i)(?:\bDELETE\b.*\bFROM\b)`,                                                 // DELETE FROM
		`(?i)(?:\bDROP\b.*\b(?:TABLE|DATABASE)\b)`,                                     // DROP TABLE/DATABASE
		`(?i)(?:';\s*\b(?:DROP|DELETE|UPDATE|INSERT|ALTER)\b)`,                         // Terminator + destructive
		`(?i)(?:\bOR\b\s+'?\d+'?\s*=\s*'?\d+'?)`,                                      // OR 1=1
		`(?i)(?:\bAND\b\s+'?\d+'?\s*=\s*'?\d+'?)`,                                     // AND 1=1
		`(?i)(?:\bWAITFOR\b\s+\bDELAY\b)`,                                             // Time-based blind injection
		`(?i)(?:\bBENCHMARK\b\s*\()`,                                                  // MySQL benchmark
		`(?i)(?:\bSLEEP\b\s*\()`,                                                     // MySQL sleep
		`(?i)(?:\\x27|\\u0027)`,                                                        // Hex-encoded quotes
		`(?i)(?:\bEXEC\b\s*(?:\b\b|\())`,                                             // EXEC
		`(?i)(?:\bINFORMATION_SCHEMA\b)`,                                              // Schema enumeration
		`(?i)(?:\bLOAD_FILE\b)`,                                                        // File access
		`(?i)(?:\bOUTFILE\b)`,                                                          // File write
	}

	compiled := make([]*regexp.Regexp, len(patterns))
	for i, p := range patterns {
		compiled[i] = regexp.MustCompile(p)
	}

	return &SQLInjectionDetector{patterns: compiled}
}

// DetectionResult represents the result of SQL injection detection.
type DetectionResult struct {
	Detected   bool     `json:"detected"`
	Patterns   []string `json:"matched_patterns,omitempty"`
	Risk       string   `json:"risk"` // "low", "medium", "high"
	Original   string   `json:"-"`    // Don't leak in JSON
	Cleaned    string   `json:"cleaned,omitempty"`
}

// Detect checks if input contains SQL injection patterns.
func (d *SQLInjectionDetector) Detect(input string) DetectionResult {
	result := DetectionResult{
		Detected: false,
		Risk:     "low",
	}

	matches := 0
	for _, pattern := range d.patterns {
		if pattern.MatchString(input) {
			matches++
		}
	}

	if matches > 0 {
		result.Detected = true
		switch {
		case matches >= 5:
			result.Risk = "high"
		case matches >= 3:
			result.Risk = "medium"
		default:
			result.Risk = "low"
		}
	}

	result.Cleaned = SanitizeSQLInput(input)
	return result
}

// SanitizeSQLInput sanitizes input to prevent SQL injection.
func SanitizeSQLInput(input string) string {
	// Remove null bytes
	input = strings.ReplaceAll(input, "\x00", "")
	// Remove control characters except newline and tab
	result := strings.Builder{}
	for _, r := range input {
		if r == '\n' || r == '\t' || (r >= 0x20 && r != 0x7f) {
			result.WriteRune(r)
		}
	}
	return result.String()
}

// ============================================================================
// CSRF Token Generation
// ============================================================================

// CSRFToken generates a cryptographically secure CSRF token.
func CSRFToken() (string, error) {
	b := make([]byte, 32)
	if _, err := rand.Read(b); err != nil {
		return "", fmt.Errorf("failed to generate CSRF token: %w", err)
	}
	return base64.URLEncoding.EncodeToString(b), nil
}

// CSRFTokenHex generates a CSRF token as hex.
func CSRFTokenHex() (string, error) {
	b := make([]byte, 32)
	if _, err := rand.Read(b); err != nil {
		return "", fmt.Errorf("failed to generate CSRF token: %w", err)
	}
	return hex.EncodeToString(b), nil
}

// ValidateCSRFToken validates a CSRF token against the expected token.
func ValidateCSRFToken(token, expected string) bool {
	return subtle.ConstantTimeCompare([]byte(token), []byte(expected)) == 1
}

// ============================================================================
// Helper: context key type for storing security context
// ============================================================================

// SuspiciousActivityType represents the type of suspicious activity detected.
type SuspiciousActivityType string

const (
	ActivityBruteForce    SuspiciousActivityType = "brute_force"
	ActivitySQLInjection  SuspiciousActivityType = "sql_injection"
	ActivityXSS           SuspiciousActivityType = "xss"
	ActivityPathTraversal SuspiciousActivityType = "path_traversal"
	ActivityRateLimit     SuspiciousActivityType = "rate_limit"
)

// SuspiciousEvent represents a detected suspicious activity.
type SuspiciousEvent struct {
	Type      SuspiciousActivityType `json:"type"`
	IP        string                 `json:"ip"`
	UserID    string                 `json:"user_id,omitempty"`
	Details   string                 `json:"details"`
	Severity  string                 `json:"severity"`
	Meta      map[string]interface{} `json:"meta,omitempty"`
	Timestamp time.Time              `json:"timestamp"`
}

// SuspiciousActivityDetector monitors and detects suspicious activities.
type SuspiciousActivityDetector struct {
	mu            sync.RWMutex
	events        []SuspiciousEvent
	loginAttempts map[string][]time.Time
	ipActivity    map[string]map[SuspiciousActivityType]int
	onDetection   func(SuspiciousEvent)
}

// NewSuspiciousActivityDetector creates a new activity detector.
func NewSuspiciousActivityDetector(onDetection func(SuspiciousEvent)) *SuspiciousActivityDetector {
	d := &SuspiciousActivityDetector{
		events:        make([]SuspiciousEvent, 0),
		loginAttempts: make(map[string][]time.Time),
		ipActivity:    make(map[string]map[SuspiciousActivityType]int),
		onDetection:   onDetection,
	}

	go d.cleanup()
	return d
}

// RecordLoginFailure records a failed login attempt and detects brute force.
func (d *SuspiciousActivityDetector) RecordLoginFailure(ip, userID string) {
	d.mu.Lock()
	defer d.mu.Unlock()

	now := time.Now()
	d.loginAttempts[ip] = append(d.loginAttempts[ip], now)

	// Keep only recent attempts (last 15 minutes)
	cutoff := now.Add(-15 * time.Minute)
	valid := d.loginAttempts[ip][:0]
	for _, t := range d.loginAttempts[ip] {
		if t.After(cutoff) {
			valid = append(valid, t)
		}
	}
	d.loginAttempts[ip] = valid

	// Detect brute force (> 10 failures in 15 minutes)
	if len(d.loginAttempts[ip]) > 10 {
		d.recordEventLocked(SuspiciousEvent{
			Type:      ActivityBruteForce,
			IP:        ip,
			UserID:    userID,
			Timestamp: now,
			Details:   fmt.Sprintf("%d failed login attempts in 15 minutes", len(d.loginAttempts[ip])),
			Severity:  "high",
		})
	}
}

// RecordSQLInjectionAttempt records a detected SQL injection attempt.
func (d *SuspiciousActivityDetector) RecordSQLInjectionAttempt(ip, input string) {
	now := time.Now()
	d.mu.Lock()

	// Track activity counts
	if d.ipActivity[ip] == nil {
		d.ipActivity[ip] = make(map[SuspiciousActivityType]int)
	}
	d.ipActivity[ip][ActivitySQLInjection]++
	count := d.ipActivity[ip][ActivitySQLInjection]
	d.mu.Unlock()

	severity := "medium"
	if count > 5 {
		severity = "critical"
	} else if count > 2 {
		severity = "high"
	}

	d.recordEventLocked(SuspiciousEvent{
		Type:      ActivitySQLInjection,
		IP:        ip,
		Timestamp: now,
		Details:   fmt.Sprintf("SQL injection attempt detected (attempt #%d)", count),
		Severity:  severity,
		Meta: map[string]interface{}{
			"input_sample": input[:min(len(input), 100)],
		},
	})
}

// RecordXSSEvent records a detected XSS attempt.
func (d *SuspiciousActivityDetector) RecordXSSEvent(ip, input string) {
	d.mu.Lock()
	if d.ipActivity[ip] == nil {
		d.ipActivity[ip] = make(map[SuspiciousActivityType]int)
	}
	d.ipActivity[ip][ActivityXSS]++
	count := d.ipActivity[ip][ActivityXSS]
	d.mu.Unlock()

	severity := "low"
	if count > 10 {
		severity = "high"
	} else if count > 5 {
		severity = "medium"
	}

	d.recordEventLocked(SuspiciousEvent{
		Type:      ActivityXSS,
		IP:        ip,
		Timestamp: time.Now(),
		Details:   fmt.Sprintf("XSS attempt detected (attempt #%d)", count),
		Severity:  severity,
	})
}

// RecordPathTraversal records a path traversal attempt.
func (d *SuspiciousActivityDetector) RecordPathTraversal(ip, path string) {
	d.recordEventLocked(SuspiciousEvent{
		Type:      ActivityPathTraversal,
		IP:        ip,
		Timestamp: time.Now(),
		Details:   fmt.Sprintf("Path traversal attempt: %s", path),
		Severity:  "high",
	})
}

// RecordEvent records a generic suspicious event.
func (d *SuspiciousActivityDetector) RecordEvent(event SuspiciousEvent) {
	d.mu.Lock()
	defer d.mu.Unlock()
	d.recordEventLocked(event)
}

func (d *SuspiciousActivityDetector) recordEventLocked(event SuspiciousEvent) {
	d.events = append(d.events, event)
	if len(d.events) > 10000 {
		d.events = d.events[len(d.events)-10000:]
	}

	if d.onDetection != nil {
		go d.onDetection(event)
	}
}

// GetEvents returns all recorded suspicious events.
func (d *SuspiciousActivityDetector) GetEvents() []SuspiciousEvent {
	d.mu.RLock()
	defer d.mu.RUnlock()
	result := make([]SuspiciousEvent, len(d.events))
	copy(result, d.events)
	return result
}

// GetIPRiskScore calculates a risk score for an IP address.
func (d *SuspiciousActivityDetector) GetIPRiskScore(ip string) int {
	d.mu.RLock()
	defer d.mu.RUnlock()

	score := 0
	activities := d.ipActivity[ip]
	for activityType, count := range activities {
		switch activityType {
		case ActivitySQLInjection:
			score += count * 20
		case ActivityBruteForce:
			score += count * 15
		case ActivityXSS:
			score += count * 10
		case ActivityPathTraversal:
			score += count * 25
		default:
			score += count * 5
		}
	}

	// Factor in recent login failures
	if attempts, ok := d.loginAttempts[ip]; ok {
		score += len(attempts) * 10
	}

	return min(score, 100)
}

// cleanup removes old entries.
func (d *SuspiciousActivityDetector) cleanup() {
	ticker := time.NewTicker(1 * time.Hour)
	defer ticker.Stop()

	for range ticker.C {
		d.mu.Lock()
		cutoff := time.Now().Add(-24 * time.Hour)
		valid := d.events[:0]
		for _, e := range d.events {
			if e.Timestamp.After(cutoff) {
				valid = append(valid, e)
			}
		}
		d.events = valid
		d.mu.Unlock()
	}
}

// ============================================================================
// Audit Logging
// ============================================================================

// AuditEventType represents the type of audit event.
type AuditEventType string

const (
	AuditLogin           AuditEventType = "login"
	AuditLogout          AuditEventType = "logout"
	AuditLoginFailed     AuditEventType = "login_failed"
	AuditPasswordChange  AuditEventType = "password_change"
	AuditTokenRefresh    AuditEventType = "token_refresh"
	AuditTokenRevoke     AuditEventType = "token_revoke"
	AuditAccountCreate   AuditEventType = "account_create"
	AuditAccountUpdate   AuditEventType = "account_update"
	AuditAccountDelete   AuditEventType = "account_delete"
	AuditSubmission      AuditEventType = "submission"
	AuditAdminAction     AuditEventType = "admin_action"
	AuditPermissionChange AuditEventType = "permission_change"
)

// AuditEvent represents an auditable event.
type AuditEvent struct {
	Type      AuditEventType       `json:"type"`
	UserID    string               `json:"user_id,omitempty"`
	Username  string               `json:"username,omitempty"`
	IP        string               `json:"ip,omitempty"`
	Timestamp time.Time            `json:"timestamp"`
	Success   bool                 `json:"success"`
	Details   string               `json:"details,omitempty"`
	Meta      map[string]interface{} `json:"meta,omitempty"`
}

// AuditLogger handles audit logging.
type AuditLogger struct {
	mu     sync.Mutex
	events []AuditEvent
	output chan AuditEvent
}

// NewAuditLogger creates a new audit logger.
func NewAuditLogger(bufferSize int) *AuditLogger {
	al := &AuditLogger{
		events: make([]AuditEvent, 0),
		output: make(chan AuditEvent, bufferSize),
	}

	go al.process()
	return al
}

// Log records an audit event.
func (al *AuditLogger) Log(event AuditEvent) {
	if event.Timestamp.IsZero() {
		event.Timestamp = time.Now()
	}
	select {
	case al.output <- event:
	default:
		// Channel full, drop oldest event
		select {
		case <-al.output:
		default:
		}
		al.output <- event
	}
}

func (al *AuditLogger) process() {
	for event := range al.output {
		al.mu.Lock()
		al.events = append(al.events, event)
		if len(al.events) > 50000 {
			al.events = al.events[len(al.events)-50000:]
		}
		al.mu.Unlock()
	}
}

// GetEvents returns audit events, optionally filtered by type.
func (al *AuditLogger) GetEvents(eventType AuditEventType, since time.Time) []AuditEvent {
	al.mu.Lock()
	defer al.mu.Unlock()

	result := make([]AuditEvent, 0)
	for _, e := range al.events {
		if eventType != "" && e.Type != eventType {
			continue
		}
		if !since.IsZero() && e.Timestamp.Before(since) {
			continue
		}
		result = append(result, e)
	}
	return result
}

// GetRecentEvents returns the most recent N events.
func (al *AuditLogger) GetRecentEvents(n int) []AuditEvent {
	al.mu.Lock()
	defer al.mu.Unlock()

	if n > len(al.events) {
		n = len(al.events)
	}
	result := make([]AuditEvent, n)
	copy(result, al.events[len(al.events)-n:])
	return result
}

// ============================================================================
// TOTP (Time-based One-Time Password) for 2FA
// ============================================================================

// TOTPConfig holds TOTP configuration.
type TOTPConfig struct {
	Secret    string `json:"secret"`
	Digits    int    `json:"digits"`
	Period    int    `json:"period"`
	Algorithm string `json:"algorithm"`
}

// GenerateTOTPSecret generates a new TOTP secret.
func GenerateTOTPSecret() (string, error) {
	b := make([]byte, 20)
	if _, err := rand.Read(b); err != nil {
		return "", fmt.Errorf("failed to generate TOTP secret: %w", err)
	}
	return base32.StdEncoding.EncodeToString(b), nil
}

// GenerateTOTPCode generates a TOTP code for the given secret and time.
func GenerateTOTPCode(secret string, t time.Time) (string, error) {
	key, err := base32.StdEncoding.DecodeString(strings.ToUpper(secret))
	if err != nil {
		return "", fmt.Errorf("invalid TOTP secret: %w", err)
	}

	period := 30
	counter := uint64(t.Unix()) / uint64(period)

	// Pack counter as big-endian 8 bytes
	var buf [8]byte
	for i := 7; i >= 0; i-- {
		buf[i] = byte(counter & 0xff)
		counter >>= 8
	}

	// HMAC-SHA1
	mac := hmac.New(sha256.New, key)
	mac.Write(buf[:])
	hash := mac.Sum(nil)

	// Dynamic truncation
	offset := hash[len(hash)-1] & 0x0f
	binary := ((int(hash[offset] & 0x7f)) << 24) |
		((int(hash[offset+1] & 0xff)) << 16) |
		((int(hash[offset+2] & 0xff)) << 8) |
		(int(hash[offset+3]) & 0xff)

	otp := binary % 1000000
	return fmt.Sprintf("%06d", otp), nil
}

// ValidateTOTP validates a TOTP code with a time window tolerance.
func ValidateTOTP(secret, code string, window int) bool {
	if window < 0 {
		window = 1
	}

	now := time.Now()
	for i := -window; i <= window; i++ {
		t := now.Add(time.Duration(i) * 30 * time.Second)
		expected, err := GenerateTOTPCode(secret, t)
		if err != nil {
			continue
		}
		if subtle.ConstantTimeCompare([]byte(code), []byte(expected)) == 1 {
			return true
		}
	}
	return false
}

// TOTPProvisioningURI generates an otpauth:// URI for QR code generation.
func TOTPProvisioningURI(secret, username, issuer string) string {
	return fmt.Sprintf("otpauth://totp/%s:%s?secret=%s&issuer=%s&algorithm=SHA256&digits=6&period=30",
		issuer, username, secret, issuer)
}

// ============================================================================
// Helper: context key type for storing security context
// ============================================================================


// SecurityInfo holds security-related context information.
type SecurityInfo struct {
	IP        string
	UserAgent string
	RequestID string
	RiskScore int
}

// ============================================================================
// Input Validation Helpers
// ============================================================================

// IsPathTraversal checks if a path contains traversal patterns.
func IsPathTraversal(path string) bool {
	traversalPatterns := []string{
		"../", "..\\", "%2e%2e/", "%2e%2e\\",
		"....//", "..../\\",
	}
	for _, pattern := range traversalPatterns {
		if strings.Contains(path, pattern) {
			return true
		}
	}
	return false
}

// SanitizeInput removes potentially dangerous characters from input.
func SanitizeInput(input string) string {
	// Remove null bytes
	input = strings.ReplaceAll(input, "\x00", "")
	// Remove control characters
	result := strings.Builder{}
	for _, r := range input {
		if r >= 0x20 && r != 0x7f {
			result.WriteRune(r)
		}
	}
	return result.String()
}

// ValidateEmail performs basic email validation.
func ValidateEmail(email string) bool {
	if len(email) > 254 {
		return false
	}
	emailRegex := regexp.MustCompile(`^[a-zA-Z0-9._%+\-]+@[a-zA-Z0-9.\-]+\.[a-zA-Z]{2,}$`)
	return emailRegex.MatchString(email)
}

// ValidateUsername validates a username.
func ValidateUsername(username string) bool {
	if len(username) < 3 || len(username) > 50 {
		return false
	}
	usernameRegex := regexp.MustCompile(`^[a-zA-Z0-9_\-]+$`)
	return usernameRegex.MatchString(username)
}
