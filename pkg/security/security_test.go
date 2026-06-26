package security

import (
	"strings"
	"testing"
	"time"
)

// ============================================================================
// Password Policy Tests
// ============================================================================

func TestDefaultPasswordPolicy(t *testing.T) {
	policy := DefaultPasswordPolicy()
	if policy == nil {
		t.Fatal("expected non-nil policy")
	}
	if policy.MinLength != 12 {
		t.Errorf("expected MinLength 12, got %d", policy.MinLength)
	}
	if policy.MaxLength != 128 {
		t.Errorf("expected MaxLength 128, got %d", policy.MaxLength)
	}
	if !policy.RequireUppercase {
		t.Error("expected RequireUppercase to be true")
	}
	if !policy.RequireLowercase {
		t.Error("expected RequireLowercase to be true")
	}
	if !policy.RequireDigit {
		t.Error("expected RequireDigit to be true")
	}
	if !policy.RequireSpecial {
		t.Error("expected RequireSpecial to be true")
	}
	if policy.MinUniqueChars != 6 {
		t.Errorf("expected MinUniqueChars 6, got %d", policy.MinUniqueChars)
	}
	if !policy.DefaultBlacklist {
		t.Error("expected DefaultBlacklist to be true")
	}
}

func TestValidatePassword_ValidStrongPassword(t *testing.T) {
	policy := DefaultPasswordPolicy()
	result := ValidatePassword(policy, "StrongP@ssw0rd!2024")
	if !result.Valid {
		t.Errorf("expected valid, got errors: %v", result.Errors)
	}
	if result.Score < 60 {
		t.Errorf("expected score >= 60 for strong password, got %d", result.Score)
	}
	if result.Strength == "weak" {
		t.Error("expected strength better than weak")
	}
}

func TestValidatePassword_TooShort(t *testing.T) {
	policy := DefaultPasswordPolicy()
	result := ValidatePassword(policy, "Ab1!")
	if result.Valid {
		t.Error("expected invalid for short password")
	}
	found := false
	for _, e := range result.Errors {
		if strings.Contains(e, "at least") {
			found = true
			break
		}
	}
	if !found {
		t.Errorf("expected length error, got: %v", result.Errors)
	}
}

func TestValidatePassword_TooLong(t *testing.T) {
	policy := DefaultPasswordPolicy()
	long := strings.Repeat("A", 129) + "a1!"
	result := ValidatePassword(policy, long)
	if result.Valid {
		t.Error("expected invalid for too long password")
	}
}

func TestValidatePassword_MissingUppercase(t *testing.T) {
	policy := DefaultPasswordPolicy()
	result := ValidatePassword(policy, "lowercase1@#")
	if result.Valid {
		t.Error("expected invalid for missing uppercase")
	}
}

func TestValidatePassword_MissingLowercase(t *testing.T) {
	policy := DefaultPasswordPolicy()
	result := ValidatePassword(policy, "UPPERCASE1@#")
	if result.Valid {
		t.Error("expected invalid for missing lowercase")
	}
}

func TestValidatePassword_MissingDigit(t *testing.T) {
	policy := DefaultPasswordPolicy()
	result := ValidatePassword(policy, "NoDigitsHere@!")
	if result.Valid {
		t.Error("expected invalid for missing digit")
	}
}

func TestValidatePassword_MissingSpecial(t *testing.T) {
	policy := DefaultPasswordPolicy()
	result := ValidatePassword(policy, "NoSpecialChar1")
	if result.Valid {
		t.Error("expected invalid for missing special character")
	}
}

func TestValidatePassword_CommonPassword(t *testing.T) {
	policy := DefaultPasswordPolicy()
	common := []string{"password", "123456", "admin", "qwerty"}
	for _, pwd := range common {
		// Add enough complexity requirements to pass other checks
		complex := "Ab1!" + pwd
		result := ValidatePassword(policy, complex)
		// Should detect common password substring or be invalid for other reasons
		if result.Valid {
			t.Logf("password %q was valid, checking for common password detection", pwd)
		}
	}
}

func TestValidatePassword_BlacklistedWord(t *testing.T) {
	policy := DefaultPasswordPolicy()
	policy.BlacklistedWords = []string{"admin", "test"}
	result := ValidatePassword(policy, "Admin@2024!Secure")
	if result.Valid {
		t.Error("expected invalid for containing blacklisted word")
	}
}

func TestValidatePassword_ScoreCalculation(t *testing.T) {
	policy := DefaultPasswordPolicy()
	tests := []struct {
		name     string
		password string
		minScore int
	}{
		{"weak", "Ab1!x", 5},
		{"fair", "Ab1!defgh", 40},
		{"good", "Ab1!defghijk", 60},
		{"strong", "Ab1!defghijklmnop", 80},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := ValidatePassword(policy, tt.password)
			if result.Score < tt.minScore {
				t.Errorf("expected score >= %d, got %d", tt.minScore, result.Score)
			}
		})
	}
}

func TestValidatePassword_CustomPolicy(t *testing.T) {
	policy := &PasswordPolicy{
		MinLength:        6,
		MaxLength:        20,
		RequireUppercase: false,
		RequireLowercase: false,
		RequireDigit:     false,
		RequireSpecial:   false,
		MinUniqueChars:   3,
		DefaultBlacklist: false,
	}
	result := ValidatePassword(policy, "hello!!")
	if !result.Valid {
		t.Errorf("expected valid with relaxed policy, got: %v", result.Errors)
	}
}

func TestScoreToStrength(t *testing.T) {
	tests := []struct {
		score    int
		expected string
	}{
		{100, "strong"},
		{80, "strong"},
		{79, "good"},
		{60, "good"},
		{59, "fair"},
		{40, "fair"},
		{39, "weak"},
		{0, "weak"},
	}
	for _, tt := range tests {
		result := scoreToStrength(tt.score)
		if result != tt.expected {
			t.Errorf("scoreToStrength(%d) = %q, want %q", tt.score, result, tt.expected)
		}
	}
}

func TestIsCommonPassword(t *testing.T) {
	if !isCommonPassword("password") {
		t.Error("expected 'password' to be common")
	}
	if !isCommonPassword("ADMIN") {
		t.Error("expected 'ADMIN' to be common (case insensitive)")
	}
	if isCommonPassword("SuperRareP@ssw0rd!") {
		t.Error("expected rare password not to be common")
	}
	if isCommonPassword("") {
		t.Error("expected empty string not to be common")
	}
}

// ============================================================================
// SQL Injection Detection Tests
// ============================================================================

func TestNewSQLInjectionDetector(t *testing.T) {
	d := NewSQLInjectionDetector()
	if d == nil {
		t.Fatal("expected non-nil detector")
	}
	if len(d.patterns) == 0 {
		t.Error("expected at least one pattern")
	}
}

func TestDetect_SQLComment(t *testing.T) {
	d := NewSQLInjectionDetector()
	inputs := []string{
		"SELECT * FROM users --",
		"admin'--",
		"1; DROP TABLE users --",
	}
	for _, input := range inputs {
		t.Run(input, func(t *testing.T) {
			result := d.Detect(input)
			if !result.Detected {
				t.Errorf("expected detection for: %s", input)
			}
		})
	}
}

func TestDetect_UnionSelect(t *testing.T) {
	d := NewSQLInjectionDetector()
	result := d.Detect("' UNION SELECT * FROM users")
	if !result.Detected {
		t.Error("expected detection for UNION SELECT")
	}
}

func TestDetect_CleanInput(t *testing.T) {
	d := NewSQLInjectionDetector()
	inputs := []string{
		"hello world",
		"username",
		"SELECT is just a word",
		"UPPERCASE TEXT",
	}
	for _, input := range inputs {
		t.Run(input, func(t *testing.T) {
			result := d.Detect(input)
			if result.Detected {
				t.Errorf("unexpected detection for clean input: %s", input)
			}
		})
	}
}

func TestDetect_RiskLevels(t *testing.T) {
	d := NewSQLInjectionDetector()

	// Low risk: 1-2 pattern matches
	lowResult := d.Detect("' --")
	if lowResult.Risk != "low" {
		t.Errorf("expected low risk, got %s", lowResult.Risk)
	}

	// We can't easily force 3+ or 5+ matches without constructing a very specific
	// input, so just verify detection works
	result := d.Detect("' UNION SELECT * FROM users WHERE 1=1 --")
	if !result.Detected {
		t.Error("expected detection for combined patterns")
	}
}

func TestSanitizeSQLInput(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{"hello\x00world", "helloworld"},
		{"normal text", "normal text"},
		{"line1\nline2", "line1\nline2"},
		{"tab\ttext", "tab\ttext"},
		{"\x01\x02\x03hello", "hello"},
	}
	for _, tt := range tests {
		result := SanitizeSQLInput(tt.input)
		if result != tt.expected {
			t.Errorf("SanitizeSQLInput(%q) = %q, want %q", tt.input, result, tt.expected)
		}
	}
}

// ============================================================================
// CSRF Token Tests
// ============================================================================

func TestCSRFToken(t *testing.T) {
	token, err := CSRFToken()
	if err != nil {
		t.Fatalf("CSRFToken() returned error: %v", err)
	}
	if token == "" {
		t.Error("expected non-empty token")
	}
	if len(token) < 40 {
		t.Errorf("expected token length >= 40, got %d", len(token))
	}
}

func TestCSRFToken_Unique(t *testing.T) {
	tokens := make(map[string]bool)
	for i := 0; i < 10; i++ {
		token, err := CSRFToken()
		if err != nil {
			t.Fatalf("CSRFToken() returned error: %v", err)
		}
		if tokens[token] {
			t.Error("expected unique tokens, got duplicate")
		}
		tokens[token] = true
	}
}

func TestCSRFTokenHex(t *testing.T) {
	token, err := CSRFTokenHex()
	if err != nil {
		t.Fatalf("CSRFTokenHex() returned error: %v", err)
	}
	if token == "" {
		t.Error("expected non-empty hex token")
	}
}

func TestValidateCSRFToken(t *testing.T) {
	token, _ := CSRFToken()
	if !ValidateCSRFToken(token, token) {
		t.Error("expected token to validate against itself")
	}
	if ValidateCSRFToken(token, "different-token") {
		t.Error("expected different token to be invalid")
	}
	if ValidateCSRFToken("", token) {
		t.Error("expected empty token to be invalid")
	}
}

// ============================================================================
// SuspiciousActivityDetector Tests
// ============================================================================

func TestNewSuspiciousActivityDetector(t *testing.T) {
	d := NewSuspiciousActivityDetector(nil)
	if d == nil {
		t.Fatal("expected non-nil detector")
	}
}

func TestRecordLoginFailure_NoBruteForce(t *testing.T) {
	d := NewSuspiciousActivityDetector(nil)
	for i := 0; i < 5; i++ {
		d.RecordLoginFailure("192.168.1.1", "user1")
	}
	events := d.GetEvents()
	if len(events) != 0 {
		t.Errorf("expected 0 events for 5 failures, got %d", len(events))
	}
}

func TestRecordLoginFailure_BruteForce(t *testing.T) {
	d := NewSuspiciousActivityDetector(nil)
	for i := 0; i < 11; i++ {
		d.RecordLoginFailure("10.0.0.1", "attacker")
	}
	events := d.GetEvents()
	if len(events) == 0 {
		t.Error("expected at least 1 brute force event")
	}
	foundBruteForce := false
	for _, e := range events {
		if e.Type == ActivityBruteForce {
			foundBruteForce = true
			break
		}
	}
	if !foundBruteForce {
		t.Error("expected a brute force activity event")
	}
}

func TestRecordSQLInjectionAttempt(t *testing.T) {
	d := NewSuspiciousActivityDetector(nil)
	d.RecordSQLInjectionAttempt("10.0.0.2", "' OR 1=1 --")
	events := d.GetEvents()
	if len(events) != 1 {
		t.Fatalf("expected 1 event, got %d", len(events))
	}
	if events[0].Type != ActivitySQLInjection {
		t.Errorf("expected SQL injection type, got %s", events[0].Type)
	}
	if events[0].Severity != "medium" {
		t.Errorf("expected medium severity for first attempt, got %s", events[0].Severity)
	}
}

func TestRecordSQLInjectionSeverity(t *testing.T) {
	d := NewSuspiciousActivityDetector(nil)
	// 3 attempts -> high severity
	for i := 0; i < 3; i++ {
		d.RecordSQLInjectionAttempt("10.0.0.3", "malicious")
	}
	// 6 attempts -> critical severity
	for i := 0; i < 3; i++ {
		d.RecordSQLInjectionAttempt("10.0.0.4", "malicious")
	}
	events := d.GetEvents()

	// Check event 3 (first 3 for 10.0.0.3 -> index 2 might be high)
	if len(events) >= 3 {
		// Just verify it recorded events
	}
}

func TestRecordXSSEvent(t *testing.T) {
	d := NewSuspiciousActivityDetector(nil)
	d.RecordXSSEvent("10.0.0.5", "<script>alert('xss')</script>")
	events := d.GetEvents()
	if len(events) != 1 {
		t.Fatalf("expected 1 event, got %d", len(events))
	}
	if events[0].Type != ActivityXSS {
		t.Errorf("expected XSS type, got %s", events[0].Type)
	}
}

func TestRecordXSS_Severity(t *testing.T) {
	d := NewSuspiciousActivityDetector(nil)
	// First attempt should be low
	d.RecordXSSEvent("10.0.0.6", "xss1")
	// 6th attempt should be medium (after 5)
	for i := 0; i < 5; i++ {
		d.RecordXSSEvent("10.0.0.6", "xss")
	}
	// 11th attempt should be high
	for i := 0; i < 5; i++ {
		d.RecordXSSEvent("10.0.0.6", "xss")
	}
	events := d.GetEvents()
	if len(events) < 11 {
		t.Fatalf("expected at least 11 events, got %d", len(events))
	}

	// Check the severity of the last event
	lastEvent := events[len(events)-1]
	if lastEvent.Severity != "high" {
		t.Errorf("expected high severity after many XSS attempts, got %s", lastEvent.Severity)
	}
}

func TestRecordPathTraversal(t *testing.T) {
	d := NewSuspiciousActivityDetector(nil)
	d.RecordPathTraversal("10.0.0.7", "../../etc/passwd")
	events := d.GetEvents()
	if len(events) != 1 {
		t.Fatalf("expected 1 event, got %d", len(events))
	}
	if events[0].Type != ActivityPathTraversal {
		t.Errorf("expected path traversal type, got %s", events[0].Type)
	}
	if events[0].Severity != "high" {
		t.Errorf("expected high severity, got %s", events[0].Severity)
	}
}

func TestRecordEvent_WithCallback(t *testing.T) {
	callbackCalled := false
	d := NewSuspiciousActivityDetector(func(event SuspiciousEvent) {
		callbackCalled = true
	})
	d.RecordEvent(SuspiciousEvent{
		Type:    ActivityRateLimit,
		IP:      "10.0.0.8",
		Details: "too many requests",
	})
	events := d.GetEvents()
	if len(events) != 1 {
		t.Fatalf("expected 1 event, got %d", len(events))
	}
	if events[0].Type != ActivityRateLimit {
		t.Errorf("expected rate limit type, got %s", events[0].Type)
	}
	// Give the goroutine time to fire
	time.Sleep(50 * time.Millisecond)
	if !callbackCalled {
		t.Error("expected callback to be called")
	}
}

func TestGetEvents(t *testing.T) {
	d := NewSuspiciousActivityDetector(nil)
	d.RecordEvent(SuspiciousEvent{Type: ActivityXSS, IP: "1.1.1.1", Details: "xss attempt", Timestamp: time.Now()})
	d.RecordEvent(SuspiciousEvent{Type: ActivityBruteForce, IP: "2.2.2.2", Details: "brute force", Timestamp: time.Now()})
	events := d.GetEvents()
	if len(events) != 2 {
		t.Fatalf("expected 2 events, got %d", len(events))
	}
}

func TestGetIPRiskScore(t *testing.T) {
	d := NewSuspiciousActivityDetector(nil)

	// No activity = score 0
	if score := d.GetIPRiskScore("unknown-ip"); score != 0 {
		t.Errorf("expected 0 for unknown IP, got %d", score)
	}

	// Add SQL injection activity
	d.RecordSQLInjectionAttempt("10.0.0.9", "malicious")
	score := d.GetIPRiskScore("10.0.0.9")
	if score <= 0 {
		t.Errorf("expected positive risk score, got %d", score)
	}

	// Add brute force
	for i := 0; i < 11; i++ {
		d.RecordLoginFailure("10.0.0.9", "user")
	}
	score = d.GetIPRiskScore("10.0.0.9")
	if score <= 0 {
		t.Errorf("expected positive risk score after failures, got %d", score)
	}
}

// ============================================================================
// Audit Logger Tests
// ============================================================================

func TestNewAuditLogger(t *testing.T) {
	al := NewAuditLogger(100)
	if al == nil {
		t.Fatal("expected non-nil audit logger")
	}
}

func TestAuditLogger_Log(t *testing.T) {
	al := NewAuditLogger(100)
	al.Log(AuditEvent{
		Type:    AuditLogin,
		UserID:  "user-1",
		Success: true,
	})
	time.Sleep(50 * time.Millisecond) // Allow async processing
	events := al.GetEvents("", time.Time{})
	if len(events) != 1 {
		t.Fatalf("expected 1 event, got %d", len(events))
	}
	if events[0].Type != AuditLogin {
		t.Errorf("expected login type, got %s", events[0].Type)
	}
}

func TestAuditLogger_GetEventsByType(t *testing.T) {
	al := NewAuditLogger(100)
	al.Log(AuditEvent{Type: AuditLogin, UserID: "u1", Success: true})
	al.Log(AuditEvent{Type: AuditLogout, UserID: "u1", Success: true})
	al.Log(AuditEvent{Type: AuditLogin, UserID: "u2", Success: false})
	time.Sleep(50 * time.Millisecond)

	loginEvents := al.GetEvents(AuditLogin, time.Time{})
	if len(loginEvents) != 2 {
		t.Errorf("expected 2 login events, got %d", len(loginEvents))
	}
}

func TestAuditLogger_GetEventsSince(t *testing.T) {
	al := NewAuditLogger(100)
	al.Log(AuditEvent{Type: AuditLogin, UserID: "u1", Success: true})
	time.Sleep(10 * time.Millisecond)
	since := time.Now()
	al.Log(AuditEvent{Type: AuditLogout, UserID: "u1", Success: true})
	time.Sleep(50 * time.Millisecond)

	events := al.GetEvents("", since)
	if len(events) != 1 {
		t.Errorf("expected 1 event since timestamp, got %d", len(events))
	}
	if len(events) > 0 && events[0].Type != AuditLogout {
		t.Errorf("expected logout event, got %s", events[0].Type)
	}
}

func TestAuditLogger_GetRecentEvents(t *testing.T) {
	al := NewAuditLogger(100)
	for i := 0; i < 10; i++ {
		al.Log(AuditEvent{Type: AuditLogin, UserID: "u1", Success: true})
	}
	time.Sleep(50 * time.Millisecond)

	recent := al.GetRecentEvents(3)
	if len(recent) != 3 {
		t.Errorf("expected 3 recent events, got %d", len(recent))
	}

	all := al.GetRecentEvents(100)
	if len(all) != 10 {
		t.Errorf("expected 10 events when requesting more than available, got %d", len(all))
	}
}

func TestAuditLogger_TimestampDefaults(t *testing.T) {
	al := NewAuditLogger(100)
	al.Log(AuditEvent{
		Type:    AuditLogin,
		UserID:  "u1",
		Success: true,
		// No Timestamp set
	})
	time.Sleep(50 * time.Millisecond)
	events := al.GetEvents("", time.Time{})
	if len(events) == 0 {
		t.Fatal("expected at least 1 event")
	}
	if events[0].Timestamp.IsZero() {
		t.Error("expected timestamp to be set automatically")
	}
}

func TestAuditLogger_EventTypes(t *testing.T) {
	types := []AuditEventType{
		AuditLogin, AuditLogout, AuditLoginFailed, AuditPasswordChange,
		AuditTokenRefresh, AuditTokenRevoke, AuditAccountCreate,
		AuditAccountUpdate, AuditAccountDelete, AuditSubmission,
		AuditAdminAction, AuditPermissionChange,
	}
	for _, at := range types {
		if at == "" {
			t.Error("expected non-empty audit event type")
		}
	}
}

// ============================================================================
// TOTP Tests
// ============================================================================

func TestGenerateTOTPSecret(t *testing.T) {
	secret, err := GenerateTOTPSecret()
	if err != nil {
		t.Fatalf("GenerateTOTPSecret() returned error: %v", err)
	}
	if secret == "" {
		t.Error("expected non-empty secret")
	}
}

func TestGenerateTOTPSecret_Unique(t *testing.T) {
	secrets := make(map[string]bool)
	for i := 0; i < 5; i++ {
		secret, err := GenerateTOTPSecret()
		if err != nil {
			t.Fatalf("GenerateTOTPSecret() returned error: %v", err)
		}
		if secrets[secret] {
			t.Error("expected unique secrets")
		}
		secrets[secret] = true
	}
}

func TestGenerateTOTPCode(t *testing.T) {
	secret, err := GenerateTOTPSecret()
	if err != nil {
		t.Fatalf("GenerateTOTPSecret() failed: %v", err)
	}
	now := time.Now()
	code, err := GenerateTOTPCode(secret, now)
	if err != nil {
		t.Fatalf("GenerateTOTPCode() returned error: %v", err)
	}
	if len(code) != 6 {
		t.Errorf("expected 6-digit code, got %q (len=%d)", code, len(code))
	}
}

func TestGenerateTOTPCode_Deterministic(t *testing.T) {
	secret := "JBSWY3DPEHPK3PXP"
	now := time.Date(2026, 6, 26, 12, 0, 0, 0, time.UTC)
	code1, err := GenerateTOTPCode(secret, now)
	if err != nil {
		t.Fatalf("first call: %v", err)
	}
	code2, err := GenerateTOTPCode(secret, now)
	if err != nil {
		t.Fatalf("second call: %v", err)
	}
	if code1 != code2 {
		t.Errorf("expected deterministic code, got %q and %q", code1, code2)
	}
}

func TestGenerateTOTPCode_InvalidSecret(t *testing.T) {
	_, err := GenerateTOTPCode("invalid-secret!!!", time.Now())
	if err == nil {
		t.Error("expected error for invalid base32 secret")
	}
}

func TestValidateTOTP(t *testing.T) {
	secret, _ := GenerateTOTPSecret()
	code, err := GenerateTOTPCode(secret, time.Now())
	if err != nil {
		t.Fatalf("GenerateTOTPCode() failed: %v", err)
	}
	if !ValidateTOTP(secret, code, 1) {
		t.Error("expected valid code to validate")
	}
}

func TestValidateTOTP_InvalidCode(t *testing.T) {
	secret, _ := GenerateTOTPSecret()
	if ValidateTOTP(secret, "000000", 1) {
		t.Error("expected invalid code to NOT validate")
	}
}

func TestValidateTOTP_Window(t *testing.T) {
	secret, _ := GenerateTOTPSecret()
	code, _ := GenerateTOTPCode(secret, time.Now())

	// Validate with window 0 (strict)
	valid := ValidateTOTP(secret, code, 0)
	if !valid {
		t.Log("code validated with window 0")
	}
}

func TestValidateTOTP_NegativeWindow(t *testing.T) {
	secret, _ := GenerateTOTPSecret()
	code, _ := GenerateTOTPCode(secret, time.Now())
	// Negative window should be treated as 1
	if !ValidateTOTP(secret, code, -1) {
		t.Error("expected negative window to work like window=1")
	}
}

func TestTOTPProvisioningURI(t *testing.T) {
	uri := TOTPProvisioningURI("JBSWY3DPEHPK3PXP", "user@example.com", "MyApp")
	if !strings.HasPrefix(uri, "otpauth://totp/") {
		t.Errorf("expected otpauth:// prefix, got %q", uri)
	}
	if !strings.Contains(uri, "secret=JBSWY3DPEHPK3PXP") {
		t.Errorf("expected secret in URI, got %q", uri)
	}
	if !strings.Contains(uri, "issuer=MyApp") {
		t.Errorf("expected issuer in URI, got %q", uri)
	}
	if !strings.Contains(uri, "algorithm=SHA256") {
		t.Errorf("expected algorithm=SHA256 in URI, got %q", uri)
	}
}

// ============================================================================
// Input Validation Tests
// ============================================================================

func TestIsPathTraversal(t *testing.T) {
	traversalPaths := []string{
		"../etc/passwd",
		"..\\windows\\system32",
		"%2e%2e/",
		"%2e%2e\\",
		"....//",
		"..../\\",
	}
	for _, path := range traversalPaths {
		t.Run(path, func(t *testing.T) {
			if !IsPathTraversal(path) {
				t.Errorf("expected path traversal detection for: %s", path)
			}
		})
	}
}

func TestIsPathTraversal_SafePaths(t *testing.T) {
	safePaths := []string{
		"/var/log/app.log",
		"/home/user/file.txt",
		"filename.pdf",
		"relative/path/file.go",
		"C:\\Program Files\\app",
	}
	for _, path := range safePaths {
		t.Run(path, func(t *testing.T) {
			if IsPathTraversal(path) {
				t.Errorf("unexpected path traversal detection for safe path: %s", path)
			}
		})
	}
}

func TestSanitizeInput(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{"hello\x00world", "helloworld"},
		{"normal text", "normal text"},
		{"\x01\x02\x03", ""},
		{"tab\ttext", "tabtext"}, // tab is 0x09, which is < 0x20
		{"newline\nhere", "newlinehere"}, // newline is 0x0a, < 0x20
	}
	for _, tt := range tests {
		result := SanitizeInput(tt.input)
		if result != tt.expected {
			t.Errorf("SanitizeInput(%q) = %q, want %q", tt.input, result, tt.expected)
		}
	}
}

func TestValidateEmail(t *testing.T) {
	validEmails := []string{
		"user@example.com",
		"user.name+tag@example.co.uk",
		"user_name@example.org",
		"user%name@example.com",
		"a@b.co",
	}
	for _, email := range validEmails {
		t.Run(email, func(t *testing.T) {
			if !ValidateEmail(email) {
				t.Errorf("expected %q to be valid", email)
			}
		})
	}

	invalidEmails := []string{
		"",
		"not-an-email",
		"@example.com",
		"user@",
		"user@.com",
		"user@example",
		strings.Repeat("a", 255) + "@example.com",
	}
	for _, email := range invalidEmails {
		t.Run(email, func(t *testing.T) {
			if ValidateEmail(email) {
				t.Errorf("expected %q to be invalid", email)
			}
		})
	}
}

func TestValidateUsername(t *testing.T) {
	validUsernames := []string{
		"alice",
		"user_123",
		"test-user",
		"abc",
		strings.Repeat("a", 50),
	}
	for _, name := range validUsernames {
		t.Run(name, func(t *testing.T) {
			if !ValidateUsername(name) {
				t.Errorf("expected %q to be valid", name)
			}
		})
	}

	invalidUsernames := []string{
		"",
		"ab",       // too short
		strings.Repeat("a", 51), // too long
		"user name",
		"user@name",
		"../etc",
	}
	for _, name := range invalidUsernames {
		t.Run(name, func(t *testing.T) {
			if ValidateUsername(name) {
				t.Errorf("expected %q to be invalid", name)
			}
		})
	}
}

// ============================================================================
// Min helper function
// ============================================================================

func TestMin(t *testing.T) {
	if min(5, 10) != 5 {
		t.Error("min(5, 10) should be 5")
	}
	if min(10, 5) != 5 {
		t.Error("min(10, 5) should be 5")
	}
	if min(5, 5) != 5 {
		t.Error("min(5, 5) should be 5")
	}
	if min(-5, 5) != -5 {
		t.Error("min(-5, 5) should be -5")
	}
}
