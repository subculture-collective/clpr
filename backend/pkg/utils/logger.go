package utils

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"regexp"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
)

// ANSI color codes for terminal output
const (
	colorReset  = "\033[0m"
	colorRed    = "\033[31m"
	colorGreen  = "\033[32m"
	colorYellow = "\033[33m"
	colorBlue   = "\033[34m"
	colorPurple = "\033[35m"
	colorCyan   = "\033[36m"
	colorWhite  = "\033[37m"
	colorGray   = "\033[90m"
	colorBold   = "\033[1m"
	colorDim    = "\033[2m"
)

// LogLevel represents logging severity
type LogLevel string

const (
	LogLevelDebug LogLevel = "debug"
	LogLevelInfo  LogLevel = "info"
	LogLevelWarn  LogLevel = "warn"
	LogLevelError LogLevel = "error"
	LogLevelFatal LogLevel = "fatal"
)

// LogFormat represents the output format
type LogFormat string

const (
	LogFormatJSON   LogFormat = "json"
	LogFormatPretty LogFormat = "pretty"
)

// StructuredLogger provides JSON structured logging
type StructuredLogger struct {
	writer   io.Writer
	minLevel LogLevel
	format   LogFormat
}

// LogEntry represents a structured log entry
type LogEntry struct {
	Timestamp  string                 `json:"timestamp"`
	Level      string                 `json:"level"`
	Message    string                 `json:"message"`
	Service    string                 `json:"service,omitempty"`
	TraceID    string                 `json:"trace_id,omitempty"`
	SpanID     string                 `json:"span_id,omitempty"`
	UserID     string                 `json:"user_id,omitempty"`
	Method     string                 `json:"method,omitempty"`
	Path       string                 `json:"path,omitempty"`
	StatusCode int                    `json:"status_code,omitempty"`
	Latency    string                 `json:"latency,omitempty"`
	ClientIP   string                 `json:"client_ip,omitempty"`
	UserAgent  string                 `json:"user_agent,omitempty"`
	Error      string                 `json:"error,omitempty"`
	Fields     map[string]interface{} `json:"fields,omitempty"`
}

// NewStructuredLogger creates a new structured logger
func NewStructuredLogger(minLevel LogLevel) *StructuredLogger {
	// Default to pretty format in development (when GIN_MODE != release)
	format := LogFormatPretty
	if os.Getenv("LOG_FORMAT") == "json" || os.Getenv("GIN_MODE") == "release" {
		format = LogFormatJSON
	}
	return &StructuredLogger{
		writer:   os.Stdout,
		minLevel: minLevel,
		format:   format,
	}
}

// NewStructuredLoggerWithFormat creates a new structured logger with specific format
func NewStructuredLoggerWithFormat(minLevel LogLevel, format LogFormat) *StructuredLogger {
	return &StructuredLogger{
		writer:   os.Stdout,
		minLevel: minLevel,
		format:   format,
	}
}

// shouldLog checks if the log level should be logged
func (l *StructuredLogger) shouldLog(level LogLevel) bool {
	levels := map[LogLevel]int{
		LogLevelDebug: 0,
		LogLevelInfo:  1,
		LogLevelWarn:  2,
		LogLevelError: 3,
		LogLevelFatal: 4,
	}
	return levels[level] >= levels[l.minLevel]
}

// log writes a structured log entry
func (l *StructuredLogger) log(entry *LogEntry) {
	if !l.shouldLog(LogLevel(entry.Level)) {
		return
	}

	entry.Timestamp = time.Now().UTC().Format(time.RFC3339)

	// Redact PII from message and error
	entry.Message = RedactPII(entry.Message)
	if entry.Error != "" {
		entry.Error = RedactPII(entry.Error)
	}

	// Redact PII from fields
	if entry.Fields != nil {
		entry.Fields = RedactPIIFromFields(entry.Fields)
	}

	if l.format == LogFormatPretty {
		l.logPretty(entry)
		return
	}

	data, err := json.Marshal(entry)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to marshal log entry: %v\n", err)
		return
	}
	fmt.Fprintln(l.writer, string(data))
}

// logPretty writes a colorized, human-readable log entry
func (l *StructuredLogger) logPretty(entry *LogEntry) {
	// Format timestamp
	ts := time.Now().Format("15:04:05.000")

	// Level colors and icons
	var levelColor, levelIcon string
	switch LogLevel(entry.Level) {
	case LogLevelDebug:
		levelColor = colorGray
		levelIcon = "🐛"
	case LogLevelInfo:
		levelColor = colorCyan
		levelIcon = "ℹ️ "
	case LogLevelWarn:
		levelColor = colorYellow
		levelIcon = "⚠️ "
	case LogLevelError:
		levelColor = colorRed
		levelIcon = "❌"
	case LogLevelFatal:
		levelColor = colorRed + colorBold
		levelIcon = "💀"
	default:
		levelColor = colorWhite
		levelIcon = "  "
	}

	// Build the log line
	levelStr := fmt.Sprintf("%s%s%-5s%s", levelColor, levelIcon, strings.ToUpper(entry.Level), colorReset)
	line := fmt.Sprintf("%s%s%s %s %s", colorGray, ts, colorReset, levelStr, entry.Message)

	// Add HTTP request details if present
	if entry.Method != "" && entry.Path != "" {
		statusColor := colorGreen
		if entry.StatusCode >= 400 && entry.StatusCode < 500 {
			statusColor = colorYellow
		} else if entry.StatusCode >= 500 {
			statusColor = colorRed
		}
		line = fmt.Sprintf("%s%s%s %s %s%s%s %s %s%d%s %s%s%s",
			colorGray, ts, colorReset,
			levelStr,
			colorBold, entry.Method, colorReset,
			entry.Path,
			statusColor, entry.StatusCode, colorReset,
			colorGray, entry.Latency, colorReset,
		)
	}

	// Add error if present
	if entry.Error != "" {
		line += fmt.Sprintf(" %s[error: %s]%s", colorRed, entry.Error, colorReset)
	}

	// Add fields if present
	if len(entry.Fields) > 0 {
		fieldsStr := ""
		for k, v := range entry.Fields {
			if fieldsStr != "" {
				fieldsStr += " "
			}
			fieldsStr += fmt.Sprintf("%s%s%s=%v", colorPurple, k, colorReset, v)
		}
		line += fmt.Sprintf(" %s", fieldsStr)
	}

	// Add trace ID if present (abbreviated)
	if entry.TraceID != "" {
		shortTrace := entry.TraceID
		if len(shortTrace) > 8 {
			shortTrace = shortTrace[:8]
		}
		line += fmt.Sprintf(" %strace=%s%s", colorGray, shortTrace, colorReset)
	}

	fmt.Fprintln(l.writer, line)
}

// Debug logs a debug message
func (l *StructuredLogger) Debug(message string, fields ...map[string]interface{}) {
	entry := LogEntry{
		Level:   string(LogLevelDebug),
		Message: message,
		Service: "clpr-backend",
	}
	if len(fields) > 0 {
		entry.Fields = fields[0]
	}
	l.log(&entry)
}

// Info logs an info message
func (l *StructuredLogger) Info(message string, fields ...map[string]interface{}) {
	entry := LogEntry{
		Level:   string(LogLevelInfo),
		Message: message,
		Service: "clpr-backend",
	}
	if len(fields) > 0 {
		entry.Fields = fields[0]
	}
	l.log(&entry)
}

// Warn logs a warning message
func (l *StructuredLogger) Warn(message string, fields ...map[string]interface{}) {
	entry := LogEntry{
		Level:   string(LogLevelWarn),
		Message: message,
		Service: "clpr-backend",
	}
	if len(fields) > 0 {
		entry.Fields = fields[0]
	}
	l.log(&entry)
}

// Error logs an error message
func (l *StructuredLogger) Error(message string, err error, fields ...map[string]interface{}) {
	entry := LogEntry{
		Level:   string(LogLevelError),
		Message: message,
		Service: "clpr-backend",
	}
	if err != nil {
		entry.Error = err.Error()
	}
	if len(fields) > 0 {
		entry.Fields = fields[0]
	}
	l.log(&entry)
}

// GinLogger returns a Gin middleware for structured logging
func (l *StructuredLogger) GinLogger() gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()
		path := c.Request.URL.Path
		query := c.Request.URL.RawQuery

		// Process request
		c.Next()

		// Calculate latency
		latency := time.Since(start)

		// Get user ID from context if available (hash it for privacy)
		userID := ""
		if uid, exists := c.Get("user_id"); exists {
			userID = hashForLogging(fmt.Sprintf("%v", uid))
		}

		// Get trace ID from context if available (use request ID)
		traceID := ""
		if tid, exists := c.Get("RequestId"); exists {
			traceID = fmt.Sprintf("%v", tid)
		}

		entry := &LogEntry{
			Level:      string(LogLevelInfo),
			Message:    "HTTP Request",
			Service:    "clpr-backend",
			TraceID:    traceID,
			UserID:     userID,
			Method:     c.Request.Method,
			Path:       path,
			StatusCode: c.Writer.Status(),
			Latency:    latency.String(),
			ClientIP:   c.ClientIP(),
			UserAgent:  c.Request.UserAgent(),
		}

		// Add query string if present
		if query != "" {
			if entry.Fields == nil {
				entry.Fields = make(map[string]interface{})
			}
			entry.Fields["query"] = query
		}

		// Add error if present
		if len(c.Errors) > 0 {
			entry.Error = c.Errors.String()
			entry.Level = string(LogLevelError)
		}

		l.log(entry)
	}
}

// Global logger instance
var defaultLogger *StructuredLogger

// InitLogger initializes the global logger
func InitLogger(minLevel LogLevel) {
	defaultLogger = NewStructuredLogger(minLevel)
}

// GetLogger returns the global logger instance
func GetLogger() *StructuredLogger {
	if defaultLogger == nil {
		defaultLogger = NewStructuredLogger(LogLevelInfo)
	}
	return defaultLogger
}

// Debug logs a debug message using the global logger
func Debug(message string, fields ...map[string]interface{}) {
	GetLogger().Debug(message, fields...)
}

// Info logs an info message using the global logger
func Info(message string, fields ...map[string]interface{}) {
	GetLogger().Info(message, fields...)
}

// Warn logs a warning message using the global logger
func Warn(message string, fields ...map[string]interface{}) {
	GetLogger().Warn(message, fields...)
}

// Error logs an error message using the global logger
func Error(message string, err error, fields ...map[string]interface{}) {
	GetLogger().Error(message, err, fields...)
}

// Fatal logs a fatal message
func (l *StructuredLogger) Fatal(message string, err error, fields ...map[string]interface{}) {
	entry := LogEntry{
		Level:   string(LogLevelFatal),
		Message: message,
		Service: "clpr-backend",
	}
	if err != nil {
		entry.Error = err.Error()
	}
	if len(fields) > 0 {
		entry.Fields = fields[0]
	}
	l.log(&entry)
	os.Exit(1)
}

// Fatal logs a fatal message using the global logger
func Fatal(message string, err error, fields ...map[string]interface{}) {
	GetLogger().Fatal(message, err, fields...)
}

// hashForLogging creates a SHA-256 hash prefix for PII protection in logs
func hashForLogging(value string) string {
	hash := sha256.Sum256([]byte(value))
	return hex.EncodeToString(hash[:8]) // Use first 8 bytes for shorter hash
}

// Patterns for PII redaction
var (
	// Email pattern
	emailPattern = regexp.MustCompile(`\b[A-Za-z0-9._%+-]+@[A-Za-z0-9.-]+\.[A-Z|a-z]{2,}\b`)
	// Credit card pattern (with optional separators)
	creditCardPattern = regexp.MustCompile(`\b\d{4}[-\s]?\d{4}[-\s]?\d{4}[-\s]?\d{4}\b`)
	// Phone number patterns (various formats)
	phonePattern = regexp.MustCompile(`\b(\+?1[-.\s]?)?(\(?\d{3}\)?[-.\s]?)?\d{3}[-.\s]?\d{4}\b`)
	// SSN pattern
	ssnPattern = regexp.MustCompile(`\b\d{3}-\d{2}-\d{4}\b`)
	// API key/token pattern (common formats)
	tokenPattern = regexp.MustCompile(`\b[A-Za-z0-9_-]{32,}\b`)
	// Password in query strings or JSON - improved to handle JSON properly
	passwordPattern = regexp.MustCompile(`(?i)(password|passwd|pwd|secret|token|apikey|api_key|access_token|auth_token)["']?\s*[:=]\s*["']?([^"'\s,}&]+)["']?`)
	// Bearer tokens
	bearerPattern = regexp.MustCompile(`(?i)Bearer\s+[A-Za-z0-9\-._~+/]+=*`)
)

// RedactPII redacts personally identifiable information from a string
func RedactPII(text string) string {
	// Redact emails
	text = emailPattern.ReplaceAllString(text, "[REDACTED_EMAIL]")
	// Redact credit cards (must come before phone)
	text = creditCardPattern.ReplaceAllString(text, "[REDACTED_CARD]")
	// Redact phone numbers
	text = phonePattern.ReplaceAllString(text, "[REDACTED_PHONE]")
	// Redact SSNs
	text = ssnPattern.ReplaceAllString(text, "[REDACTED_SSN]")
	// Redact passwords and secrets in key-value pairs
	text = passwordPattern.ReplaceAllString(text, `$1":"[REDACTED]"`)
	// Redact Bearer tokens
	text = bearerPattern.ReplaceAllString(text, "Bearer [REDACTED_TOKEN]")
	return text
}

// RedactPIIFromFields redacts PII from log entry fields
func RedactPIIFromFields(fields map[string]interface{}) map[string]interface{} {
	if fields == nil {
		return nil
	}

	redacted := make(map[string]interface{})
	for key, value := range fields {
		lowerKey := strings.ToLower(key)

		// Redact sensitive field names
		if strings.Contains(lowerKey, "password") ||
			strings.Contains(lowerKey, "secret") ||
			strings.Contains(lowerKey, "token") ||
			strings.Contains(lowerKey, "api_key") ||
			strings.Contains(lowerKey, "apikey") ||
			strings.Contains(lowerKey, "authorization") ||
			strings.Contains(lowerKey, "auth") {
			redacted[key] = "[REDACTED]"
			continue
		}

		// Redact PII from string values
		if str, ok := value.(string); ok {
			redacted[key] = RedactPII(str)
		} else {
			redacted[key] = value
		}
	}
	return redacted
}

// PIIRedactionMiddleware returns a Gin middleware that redacts PII from logs
func PIIRedactionMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		// Store original values for processing
		c.Next()

		// No actual modification needed here as redaction happens at log time
		// This middleware serves as a marker that PII redaction is enabled
	}
}
