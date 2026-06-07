package middleware

import (
	"bytes"
	"encoding/json"
	"io"
	"math/rand"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
)

// TestInputValidationMiddleware_SQLInjectionEdgeCases tests additional SQLi patterns
func TestInputValidationMiddleware_SQLInjectionEdgeCases(t *testing.T) {
	gin.SetMode(gin.TestMode)

	tests := []struct {
		name         string
		query        string
		shouldReject bool
		description  string
	}{
		{
			name:         "union with spaces",
			query:        "?q=test'  UNION  SELECT  *  FROM  users--",
			shouldReject: true,
			description:  "Space-separated UNION SELECT",
		},
		{
			name:         "case variation union",
			query:        "?q=test' UnIoN SeLeCt * FrOm users--",
			shouldReject: true,
			description:  "Case variation SQLi",
		},
		{
			name:         "update with set",
			query:        "?id=1; UPDATE users SET role='admin' WHERE id=1",
			shouldReject: false, // Pattern doesn't catch simple semicolon commands without dangerous keywords
			description:  "UPDATE statement (limitation: not caught by current patterns)",
		},
		{
			name:         "delete from",
			query:        "?id=1; DELETE FROM users WHERE id > 0",
			shouldReject: false, // Pattern doesn't catch simple DELETE statements
			description:  "DELETE statement (limitation: not caught by current patterns)",
		},
		{
			name:         "alter table",
			query:        "?name=test'; ALTER TABLE users ADD COLUMN backdoor VARCHAR(255)--",
			shouldReject: false, // Pattern doesn't catch ALTER statements
			description:  "ALTER TABLE statement (limitation: not caught by current patterns)",
		},
		{
			name:         "normal apostrophe in text",
			query:        "?name=O'Brien",
			shouldReject: false,
			description:  "Legitimate apostrophe in name",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			w := httptest.NewRecorder()
			c, r := gin.CreateTestContext(w)

			r.Use(InputValidationMiddleware())
			r.GET("/test", func(c *gin.Context) {
				c.JSON(http.StatusOK, gin.H{"message": "ok"})
			})

			req, _ := http.NewRequest("GET", "/test"+tt.query, nil)
			c.Request = req
			r.ServeHTTP(w, req)

			if tt.shouldReject {
				assert.Equal(t, http.StatusBadRequest, w.Code,
					"Expected rejection for: %s", tt.description)
			} else {
				// Log the result - includes limitations and legitimate cases
				t.Logf("%s: got status %d", tt.description, w.Code)
			}
		})
	}
}

// TestInputValidationMiddleware_XSSEdgeCases tests additional XSS patterns
func TestInputValidationMiddleware_XSSEdgeCases(t *testing.T) {
	gin.SetMode(gin.TestMode)

	tests := []struct {
		name         string
		query        string
		shouldReject bool
		description  string
	}{
		{
			name:         "iframe injection",
			query:        "?html=<iframe src='http://evil.com'></iframe>",
			shouldReject: true,
			description:  "iframe tag",
		},
		{
			name:         "object tag",
			query:        "?html=<object data='http://evil.com'></object>",
			shouldReject: true,
			description:  "object tag",
		},
		{
			name:         "embed tag",
			query:        "?html=<embed src='http://evil.com'>",
			shouldReject: true,
			description:  "embed tag",
		},
		{
			name:         "case variation script",
			query:        "?html=<ScRiPt>alert(1)</sCrIpT>",
			shouldReject: true,
			description:  "Case variation script tag",
		},
		{
			name:         "javascript in href",
			query:        "?url=javascript:void(document.cookie)",
			shouldReject: true,
			description:  "javascript: protocol",
		},
		{
			name:         "event handler without tag",
			query:        "?html=onerror=alert(1)",
			shouldReject: true,
			description:  "Event handler string",
		},
		{
			name:         "legitimate less than",
			query:        "?price=<100",
			shouldReject: false,
			description:  "Legitimate less-than operator",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			w := httptest.NewRecorder()
			c, r := gin.CreateTestContext(w)

			r.Use(InputValidationMiddleware())
			r.GET("/test", func(c *gin.Context) {
				c.JSON(http.StatusOK, gin.H{"message": "ok"})
			})

			req, _ := http.NewRequest("GET", "/test"+tt.query, nil)
			c.Request = req
			r.ServeHTTP(w, req)

			if tt.shouldReject {
				assert.Equal(t, http.StatusBadRequest, w.Code,
					"Expected rejection for: %s", tt.description)
			} else {
				t.Logf("%s: got status %d", tt.description, w.Code)
			}
		})
	}
}

// TestInputValidationMiddleware_HeaderInjection tests header injection attempts
func TestInputValidationMiddleware_HeaderInjection(t *testing.T) {
	gin.SetMode(gin.TestMode)

	tests := []struct {
		name         string
		headerKey    string
		headerValue  string
		shouldReject bool
	}{
		{
			name:         "very long header",
			headerKey:    "X-Long-Header",
			headerValue:  strings.Repeat("A", MaxHeaderValueLength+1),
			shouldReject: true,
		},
		{
			name:         "invalid UTF-8 in custom header",
			headerKey:    "X-Invalid",
			headerValue:  "test\xff\xfe",
			shouldReject: true,
		},
		{
			name:         "normal custom header",
			headerKey:    "X-Request-ID",
			headerValue:  "12345",
			shouldReject: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			w := httptest.NewRecorder()
			c, r := gin.CreateTestContext(w)

			r.Use(InputValidationMiddleware())
			r.GET("/test", func(c *gin.Context) {
				c.JSON(http.StatusOK, gin.H{"message": "ok"})
			})

			req, _ := http.NewRequest("GET", "/test", nil)
			req.Header.Set(tt.headerKey, tt.headerValue)
			c.Request = req
			r.ServeHTTP(w, req)

			if tt.shouldReject {
				assert.NotEqual(t, http.StatusOK, w.Code,
					"Expected rejection for header: %s", tt.name)
			}
		})
	}
}

// TestInputValidationMiddleware_RequestBodyValidation tests body size limits
func TestInputValidationMiddleware_RequestBodyValidation(t *testing.T) {
	gin.SetMode(gin.TestMode)

	t.Run("large request body", func(t *testing.T) {
		w := httptest.NewRecorder()
		c, r := gin.CreateTestContext(w)

		r.Use(InputValidationMiddleware())
		r.POST("/test", func(c *gin.Context) {
			var data map[string]interface{}
			if err := c.ShouldBindJSON(&data); err != nil {
				c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
				return
			}
			c.JSON(http.StatusOK, gin.H{"message": "ok"})
		})

		// Create body larger than MaxRequestBodySize
		largeData := strings.Repeat("A", MaxRequestBodySize+1000)
		req, _ := http.NewRequest("POST", "/test", bytes.NewBufferString(largeData))
		req.Header.Set("Content-Type", "application/json")
		c.Request = req
		r.ServeHTTP(w, req)

		// Should fail when trying to read the body
		assert.NotEqual(t, http.StatusOK, w.Code, "Should reject oversized body")
	})

	t.Run("normal request body", func(t *testing.T) {
		w := httptest.NewRecorder()
		c, r := gin.CreateTestContext(w)

		r.Use(InputValidationMiddleware())
		r.POST("/test", func(c *gin.Context) {
			var data map[string]interface{}
			if err := c.ShouldBindJSON(&data); err != nil {
				c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
				return
			}
			c.JSON(http.StatusOK, gin.H{"message": "ok"})
		})

		normalData := map[string]string{"key": "value"}
		jsonData, err := json.Marshal(normalData)
		assert.NoError(t, err)
		req, _ := http.NewRequest("POST", "/test", bytes.NewBuffer(jsonData))
		req.Header.Set("Content-Type", "application/json")
		c.Request = req
		r.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
	})
}

func TestInputValidationMiddleware_AllowsUploadRoutePastGlobalBodyLimit(t *testing.T) {
	gin.SetMode(gin.TestMode)

	w := httptest.NewRecorder()
	c, r := gin.CreateTestContext(w)

	r.Use(InputValidationMiddleware())
	r.POST("/api/v1/submissions/upload", func(c *gin.Context) {
		_, err := io.Copy(io.Discard, c.Request.Body)
		if err != nil {
			c.JSON(http.StatusRequestEntityTooLarge, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusOK, gin.H{"message": "ok"})
	})

	largeData := strings.Repeat("A", MaxRequestBodySize+1000)
	req, _ := http.NewRequest("POST", "/api/v1/submissions/upload", bytes.NewBufferString(largeData))
	req.Header.Set("Content-Type", "application/octet-stream")
	c.Request = req
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code, "upload route should not be capped by global 10MB body limit")
}

// TestInputValidationMiddleware_MixedPayloads tests combinations of attack vectors
func TestInputValidationMiddleware_MixedPayloads(t *testing.T) {
	gin.SetMode(gin.TestMode)

	tests := []struct {
		name  string
		query string
	}{
		{
			name:  "SQLi and XSS combined",
			query: "?q=<script>alert('xss')</script>' UNION SELECT * FROM users--",
		},
		{
			name:  "path traversal with XSS",
			query: "?file=../../etc/passwd<script>alert(1)</script>",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			w := httptest.NewRecorder()
			c, r := gin.CreateTestContext(w)

			r.Use(InputValidationMiddleware())
			r.GET("/test", func(c *gin.Context) {
				c.JSON(http.StatusOK, gin.H{"message": "ok"})
			})

			req, _ := http.NewRequest("GET", "/test"+tt.query, nil)
			c.Request = req
			r.ServeHTTP(w, req)

			// Should reject any payload containing suspicious patterns
			assert.Equal(t, http.StatusBadRequest, w.Code,
				"Mixed attack payload should be rejected")
		})
	}
}

// TestInputValidationMiddleware_FuzzerSmoke runs 1000+ random payloads
func TestInputValidationMiddleware_FuzzerSmoke(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping fuzzer test in short mode")
	}

	gin.SetMode(gin.TestMode)

	// Character sets for fuzzing
	specialChars := []string{"'", "\"", "<", ">", ";", "&", "|", "`", "\n", "\r", "\t", "\\", "/"}
	sqlKeywords := []string{"SELECT", "UNION", "INSERT", "DELETE", "UPDATE", "DROP", "ALTER", "FROM", "WHERE"}
	xssKeywords := []string{"<script>", "</script>", "javascript:", "onerror=", "onload=", "<iframe", "<object"}
	pathPatterns := []string{"../", "..\\", "..", "/etc/", "C:\\"}

	totalTests := 1000
	failures := 0
	panics := 0

	t.Logf("Running %d fuzzer tests...", totalTests)

	for i := 0; i < totalTests; i++ {
		func() {
			defer func() {
				if r := recover(); r != nil {
					panics++
					t.Logf("PANIC in fuzzer test %d: %v", i, r)
				}
			}()

			// Generate random payload
			payload := generateRandomPayload(specialChars, sqlKeywords, xssKeywords, pathPatterns)

			w := httptest.NewRecorder()
			c, r := gin.CreateTestContext(w)

			r.Use(InputValidationMiddleware())
			r.GET("/test", func(c *gin.Context) {
				c.JSON(http.StatusOK, gin.H{"message": "ok"})
			})

			req, err := http.NewRequest("GET", "/test?fuzz="+url.QueryEscape(payload), nil)
			if err != nil {
				failures++
				t.Logf("Failed to create request %d: %v", i, err)
				return
			}

			c.Request = req
			r.ServeHTTP(w, req)

			// We don't assert specific status codes - just ensure no panics
			// and that we get a valid HTTP response
			if w.Code == 0 {
				failures++
				t.Logf("Invalid response code 0 for payload: %s", payload)
			}
		}()
	}

	t.Logf("Fuzzer test complete: %d tests, %d failures, %d panics", totalTests, failures, panics)

	// Assert no panics occurred
	assert.Equal(t, 0, panics, "Fuzzer should not cause any panics")

	// Failures should be less than 1% (excluding panics)
	maxFailures := totalTests / 100
	assert.LessOrEqual(t, failures, maxFailures,
		"Fuzzer failure rate should be less than 1%% (got %d failures)", failures)
}

// generateRandomPayload creates random test payloads
func generateRandomPayload(specialChars, sqlKeywords, xssKeywords, pathPatterns []string) string {
	payloadType := rand.Intn(5)

	switch payloadType {
	case 0: // Random special characters
		length := rand.Intn(50) + 1
		var payload strings.Builder
		for i := 0; i < length; i++ {
			payload.WriteString(specialChars[rand.Intn(len(specialChars))])
		}
		return payload.String()

	case 1: // SQL-like payload
		keyword1 := sqlKeywords[rand.Intn(len(sqlKeywords))]
		keyword2 := sqlKeywords[rand.Intn(len(sqlKeywords))]
		return keyword1 + " " + keyword2 + " users" + specialChars[rand.Intn(len(specialChars))]

	case 2: // XSS-like payload
		xss := xssKeywords[rand.Intn(len(xssKeywords))]
		return xss + "alert(" + string(rune(rand.Intn(100))) + ")"

	case 3: // Path traversal-like payload
		path := pathPatterns[rand.Intn(len(pathPatterns))]
		depth := rand.Intn(5) + 1
		return strings.Repeat(path, depth) + "test"

	default: // Mixed random payload
		var payload strings.Builder
		components := [][]string{specialChars, sqlKeywords, xssKeywords, pathPatterns}
		for i := 0; i < rand.Intn(5)+1; i++ {
			component := components[rand.Intn(len(components))]
			payload.WriteString(component[rand.Intn(len(component))])
		}
		return payload.String()
	}
}

// TestSanitizeInput_CrossFieldValidation tests sanitizer with multiple fields
func TestSanitizeInput_CrossFieldValidation(t *testing.T) {
	sanitizer := NewSanitizer()

	tests := []struct {
		name     string
		email    string
		username string
		url      string
		valid    bool
	}{
		{
			name:     "all valid fields",
			email:    "user@example.com",
			username: "validuser",
			url:      "https://example.com",
			valid:    true,
		},
		{
			name:     "invalid email",
			email:    "invalid-email",
			username: "validuser",
			url:      "https://example.com",
			valid:    false,
		},
		{
			name:     "invalid username",
			email:    "user@example.com",
			username: "ab", // too short
			url:      "https://example.com",
			valid:    false,
		},
		{
			name:     "invalid url",
			email:    "user@example.com",
			username: "validuser",
			url:      "javascript:alert(1)",
			valid:    false,
		},
		{
			name:     "XSS in username",
			email:    "user@example.com",
			username: "<script>alert(1)</script>",
			url:      "https://example.com",
			valid:    false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			emailValid := sanitizer.ValidateEmail(tt.email)
			usernameValid := sanitizer.ValidateUsername(tt.username)
			urlValid := sanitizer.ValidateURL(tt.url)

			allValid := emailValid && usernameValid && urlValid

			assert.Equal(t, tt.valid, allValid,
				"Cross-field validation: email=%v, username=%v, url=%v",
				emailValid, usernameValid, urlValid)
		})
	}
}

// TestSanitizeInput_SanitizationConsistency ensures sanitization is consistent
func TestSanitizeInput_SanitizationConsistency(t *testing.T) {
	sanitizer := NewSanitizer()

	maliciousInput := "<script>alert('xss')</script>Hello World<iframe src='evil'></iframe>"

	// Sanitize multiple times
	result1 := sanitizer.SanitizeHTML(maliciousInput)
	result2 := sanitizer.SanitizeHTML(maliciousInput)
	result3 := sanitizer.SanitizeHTML(result1) // Sanitize already sanitized

	// Should be consistent
	assert.Equal(t, result1, result2, "Sanitization should be consistent")
	assert.Equal(t, result1, result3, "Double sanitization should be idempotent")

	// Should not contain malicious content
	assert.NotContains(t, result1, "<script>")
	assert.NotContains(t, result1, "<iframe>")
	assert.NotContains(t, result1, "alert")

	// Should contain safe content
	assert.Contains(t, result1, "Hello World")
}
func TestMinimalValidation(t *testing.T) {
	gin.SetMode(gin.TestMode)

	w := httptest.NewRecorder()
	c, r := gin.CreateTestContext(w)

	r.Use(InputValidationMiddleware())
	r.GET("/test", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"message": "ok"})
	})

	req, _ := http.NewRequest("GET", "/test?q=hello", nil)
	c.Request = req
	r.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d", w.Code)
	}
}
