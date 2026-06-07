package middleware

import (
	"net/http"
	"regexp"
	"strings"
	"unicode/utf8"

	"github.com/gin-gonic/gin"
	"github.com/microcosm-cc/bluemonday"
)

const (
	// MaxRequestBodySize limits the size of request bodies (10MB)
	MaxRequestBodySize = 10 * 1024 * 1024
	// MaxURLLength limits the length of URLs
	MaxURLLength = 2048
	// MaxHeaderValueLength limits the length of header values
	MaxHeaderValueLength = 8192
)

var (
	// Common injection patterns to detect
	sqlInjectionPattern  = regexp.MustCompile(`(?i)(union[\s(]+select|select[\s(]+\*[\s(]+from|insert[\s(]+into|delete[\s(]+from|drop[\s(]+table|alter[\s(]+table|update[\s(]+\w+[\s(]+set|;[\s]*(drop|insert|delete|update)[\s(]+)`)
	tauteologyPattern    = regexp.MustCompile(`(?i)(\bor\b\s*1=1|\bor\b\s*'1'='1)`)
	xssPattern           = regexp.MustCompile(`(?i)(<script|javascript:|onerror=|onload=|<iframe|<object|<embed)`)
	pathTraversalPattern = regexp.MustCompile(`\.\.[/\\]`)
	// Detect shell command injection attempts (e.g., "; cat /etc/passwd" or "&& rm -rf /")
	commandInjectionPattern = regexp.MustCompile(`(?i)(;|&&|\|\|)\s*[a-z]{2,}`)

	// Strict sanitizer for user-generated content
	strictPolicy = bluemonday.StrictPolicy()

	// UGC sanitizer for markdown/rich content
	ugcPolicy = bluemonday.UGCPolicy()
)

// InputValidationMiddleware validates and sanitizes all incoming requests
func InputValidationMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		// Validate request method
		if !isValidHTTPMethod(c.Request.Method) {
			c.JSON(http.StatusMethodNotAllowed, gin.H{
				"error": "Invalid HTTP method",
			})
			c.Abort()
			return
		}

		// Validate URL length
		if len(c.Request.URL.String()) > MaxURLLength {
			c.JSON(http.StatusRequestURITooLong, gin.H{
				"error": "URL too long",
			})
			c.Abort()
			return
		}

		// Validate path for traversal attacks
		if pathTraversalPattern.MatchString(c.Request.URL.Path) {
			c.JSON(http.StatusBadRequest, gin.H{
				"error": "Invalid path",
			})
			c.Abort()
			return
		}

		// Quick raw query scan for obvious command injection separators before parsing parameters
		if raw := c.Request.URL.RawQuery; raw != "" {
			if commandInjectionPattern.MatchString(raw) {
				c.JSON(http.StatusBadRequest, gin.H{
					"error": "Potentially malicious input detected",
				})
				c.Abort()
				return
			}
		}

		// Validate query parameters
		for key, values := range c.Request.URL.Query() {
			if !isValidParamName(key) {
				c.JSON(http.StatusBadRequest, gin.H{
					"error": "Invalid query parameter name",
				})
				c.Abort()
				return
			}

			for _, value := range values {
				if !utf8.ValidString(value) {
					c.JSON(http.StatusBadRequest, gin.H{
						"error": "Invalid UTF-8 in query parameter",
					})
					c.Abort()
					return
				}

				// Check for common injection patterns
				if containsSuspiciousPatterns(value) {
					c.JSON(http.StatusBadRequest, gin.H{
						"error": "Potentially malicious input detected",
					})
					c.Abort()
					return
				}
			}
		}

		// Validate headers
		for key, values := range c.Request.Header {
			// Skip standard headers
			if isStandardHeader(key) {
				continue
			}

			for _, value := range values {
				if len(value) > MaxHeaderValueLength {
					c.JSON(http.StatusRequestHeaderFieldsTooLarge, gin.H{
						"error": "Header value too long",
					})
					c.Abort()
					return
				}

				if !utf8.ValidString(value) {
					c.JSON(http.StatusBadRequest, gin.H{
						"error": "Invalid UTF-8 in header",
					})
					c.Abort()
					return
				}
			}
		}

		// Limit request body size
		if !isHostedUploadRoute(c) {
			c.Request.Body = http.MaxBytesReader(c.Writer, c.Request.Body, MaxRequestBodySize)
		}

		c.Next()
	}
}

// SanitizeInput provides utility functions for sanitizing user input
type SanitizeInput struct{}

// SanitizeHTML sanitizes HTML content using strict policy
func (s *SanitizeInput) SanitizeHTML(input string) string {
	return strictPolicy.Sanitize(input)
}

// SanitizeMarkdown sanitizes markdown/UGC content
func (s *SanitizeInput) SanitizeMarkdown(input string) string {
	return ugcPolicy.Sanitize(input)
}

// SanitizePlainText strips all HTML tags and returns plain text
func (s *SanitizeInput) SanitizePlainText(input string) string {
	return strictPolicy.Sanitize(input)
}

// ValidateEmail validates email format
func (s *SanitizeInput) ValidateEmail(email string) bool {
	if len(email) > 254 {
		return false
	}

	emailPattern := regexp.MustCompile(`^[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}$`)
	return emailPattern.MatchString(email)
}

// ValidateUsername validates username format
func (s *SanitizeInput) ValidateUsername(username string) bool {
	if len(username) < 3 || len(username) > 30 {
		return false
	}

	usernamePattern := regexp.MustCompile(`^[a-zA-Z0-9_-]+$`)
	return usernamePattern.MatchString(username)
}

// ValidateURL validates URL format and scheme
func (s *SanitizeInput) ValidateURL(url string) bool {
	if len(url) > MaxURLLength {
		return false
	}

	// Only allow http and https schemes
	urlPattern := regexp.MustCompile(`^https?://[^\s<>]+$`)
	return urlPattern.MatchString(url)
}

// isValidHTTPMethod checks if the HTTP method is valid
func isValidHTTPMethod(method string) bool {
	validMethods := []string{"GET", "POST", "PUT", "DELETE", "PATCH", "HEAD", "OPTIONS"}
	for _, m := range validMethods {
		if method == m {
			return true
		}
	}
	return false
}

// isValidParamName checks if a parameter name is valid
func isValidParamName(name string) bool {
	if len(name) == 0 || len(name) > 100 {
		return false
	}

	// Allow alphanumeric, underscore, hyphen, and dot
	paramPattern := regexp.MustCompile(`^[a-zA-Z0-9_.-]+$`)
	return paramPattern.MatchString(name)
}

// containsSuspiciousPatterns checks for common injection patterns
func containsSuspiciousPatterns(input string) bool {
	// Skip checks for very short inputs
	if len(input) < 3 {
		return false
	}

	// Check for SQL injection patterns
	if sqlInjectionPattern.MatchString(input) || tauteologyPattern.MatchString(input) {
		return true
	}

	// Check for XSS patterns
	if xssPattern.MatchString(input) {
		return true
	}

	// Check for command injection patterns (e.g., shell metacharacters)
	if commandInjectionPattern.MatchString(input) {
		return true
	}

	return false
}

// isStandardHeader checks if a header is a standard HTTP header
func isStandardHeader(name string) bool {
	standardHeaders := []string{
		"Accept", "Accept-Encoding", "Accept-Language", "Authorization",
		"Cache-Control", "Connection", "Content-Length", "Content-Type",
		"Cookie", "Host", "Origin", "Referer", "User-Agent",
		"X-Csrf-Token", "X-Requested-With",
	}

	for _, h := range standardHeaders {
		if strings.EqualFold(name, h) {
			return true
		}
	}

	return false
}

func isHostedUploadRoute(c *gin.Context) bool {
	if c == nil || c.Request == nil {
		return false
	}

	if fullPath := c.FullPath(); fullPath == "/api/v1/submissions/upload" || fullPath == "/submissions/upload" {
		return true
	}

	return strings.HasSuffix(c.Request.URL.Path, "/submissions/upload")
}

// NewSanitizer creates a new input sanitizer
func NewSanitizer() *SanitizeInput {
	return &SanitizeInput{}
}
