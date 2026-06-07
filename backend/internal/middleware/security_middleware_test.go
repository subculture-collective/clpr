package middleware

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"git.subcult.tv/subculture-collective/clpr/config"
)

func TestSecurityHeadersMiddleware(t *testing.T) {
	tests := []struct {
		name       string
		ginMode    string
		wantHSTS   bool
		wantCSP    bool
		wantXFrame bool
	}{
		{
			name:       "production mode with all security headers",
			ginMode:    "release",
			wantHSTS:   true,
			wantCSP:    true,
			wantXFrame: true,
		},
		{
			name:       "development mode without HSTS",
			ginMode:    "debug",
			wantHSTS:   false,
			wantCSP:    true,
			wantXFrame: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Setup
			gin.SetMode(tt.ginMode)
			w := httptest.NewRecorder()
			c, _ := gin.CreateTestContext(w)

			cfg := &config.Config{
				Server: config.ServerConfig{
					GinMode: tt.ginMode,
				},
			}

			// Create middleware
			middleware := SecurityHeadersMiddleware(cfg)

			// Create a test handler
			c.Request, _ = http.NewRequest("GET", "/test", nil)
			middleware(c)

			// Check HSTS header
			hsts := w.Header().Get("Strict-Transport-Security")
			if tt.wantHSTS && hsts == "" {
				t.Error("Expected HSTS header in production mode, got none")
			}
			if !tt.wantHSTS && hsts != "" {
				t.Errorf("Did not expect HSTS header in %s mode, got: %s", tt.ginMode, hsts)
			}

			// Check X-Content-Type-Options
			if got := w.Header().Get("X-Content-Type-Options"); got != "nosniff" {
				t.Errorf("X-Content-Type-Options = %q, want %q", got, "nosniff")
			}

			// Check X-Frame-Options
			if got := w.Header().Get("X-Frame-Options"); got != "DENY" {
				t.Errorf("X-Frame-Options = %q, want %q", got, "DENY")
			}

			// Check X-XSS-Protection
			if got := w.Header().Get("X-XSS-Protection"); got != "1; mode=block" {
				t.Errorf("X-XSS-Protection = %q, want %q", got, "1; mode=block")
			}

			// Check Referrer-Policy
			if got := w.Header().Get("Referrer-Policy"); got != "strict-origin-when-cross-origin" {
				t.Errorf("Referrer-Policy = %q, want %q", got, "strict-origin-when-cross-origin")
			}

			// Check Content-Security-Policy
			csp := w.Header().Get("Content-Security-Policy")
			if csp == "" {
				t.Error("Expected Content-Security-Policy header, got none")
			}

			// Verify CSP contains critical directives
			if tt.wantCSP {
				expectedDirectives := []string{
					"default-src 'self'",
					"frame-ancestors 'none'",
					"upgrade-insecure-requests",
				}
				for _, directive := range expectedDirectives {
					if !contains(csp, directive) {
						t.Errorf("CSP missing directive: %q", directive)
					}
				}
			}

			// Check Permissions-Policy
			pp := w.Header().Get("Permissions-Policy")
			if pp == "" {
				t.Error("Expected Permissions-Policy header, got none")
			}
		})
	}
}

func TestGetSecureCookieOptions(t *testing.T) {
	tests := []struct {
		name       string
		ginMode    string
		wantSecure bool
	}{
		{
			name:       "production mode sets Secure flag",
			ginMode:    "release",
			wantSecure: true,
		},
		{
			name:       "development mode does not set Secure flag",
			ginMode:    "debug",
			wantSecure: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cfg := &config.Config{
				Server: config.ServerConfig{
					GinMode: tt.ginMode,
				},
			}

			options := GetSecureCookieOptions(cfg)

			// Check HTTPOnly is always true
			if !options.HTTPOnly {
				t.Error("HTTPOnly should always be true")
			}

			// Check Secure flag based on mode
			if options.Secure != tt.wantSecure {
				t.Errorf("Secure = %v, want %v", options.Secure, tt.wantSecure)
			}

			// Check SameSite is set
			if options.SameSite != "lax" {
				t.Errorf("SameSite = %q, want %q", options.SameSite, "lax")
			}

			// Check MaxAge is set
			if options.MaxAge != 86400 {
				t.Errorf("MaxAge = %d, want %d", options.MaxAge, 86400)
			}

			// Check Path is set
			if options.Path != "/" {
				t.Errorf("Path = %q, want %q", options.Path, "/")
			}
		})
	}
}

func TestSetSecureCookie(t *testing.T) {
	tests := []struct {
		name         string
		cookieName   string
		cookieValue  string
		sameSite     string
		wantSameSite http.SameSite
	}{
		{
			name:         "strict SameSite",
			cookieName:   "test_strict",
			cookieValue:  "value1",
			sameSite:     "strict",
			wantSameSite: http.SameSiteStrictMode,
		},
		{
			name:         "lax SameSite",
			cookieName:   "test_lax",
			cookieValue:  "value2",
			sameSite:     "lax",
			wantSameSite: http.SameSiteLaxMode,
		},
		{
			name:         "none SameSite",
			cookieName:   "test_none",
			cookieValue:  "value3",
			sameSite:     "none",
			wantSameSite: http.SameSiteNoneMode,
		},
		{
			name:         "default SameSite",
			cookieName:   "test_default",
			cookieValue:  "value4",
			sameSite:     "invalid",
			wantSameSite: http.SameSiteLaxMode, // Default to Lax
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Setup
			w := httptest.NewRecorder()
			c, _ := gin.CreateTestContext(w)
			c.Request, _ = http.NewRequest("GET", "/test", nil)

			options := SecureCookieOptions{
				HTTPOnly: true,
				Secure:   true,
				SameSite: tt.sameSite,
				MaxAge:   3600,
				Path:     "/",
				Domain:   "",
			}

			// Set cookie
			SetSecureCookie(c, tt.cookieName, tt.cookieValue, options)

			// Get cookie from response
			cookies := w.Result().Cookies()
			if len(cookies) == 0 {
				t.Fatal("No cookies set in response")
			}

			cookie := cookies[0]

			// Verify cookie properties
			if cookie.Name != tt.cookieName {
				t.Errorf("Cookie name = %q, want %q", cookie.Name, tt.cookieName)
			}

			if cookie.Value != tt.cookieValue {
				t.Errorf("Cookie value = %q, want %q", cookie.Value, tt.cookieValue)
			}

			if cookie.HttpOnly != options.HTTPOnly {
				t.Errorf("Cookie HttpOnly = %v, want %v", cookie.HttpOnly, options.HTTPOnly)
			}

			if cookie.Secure != options.Secure {
				t.Errorf("Cookie Secure = %v, want %v", cookie.Secure, options.Secure)
			}

			if cookie.MaxAge != options.MaxAge {
				t.Errorf("Cookie MaxAge = %d, want %d", cookie.MaxAge, options.MaxAge)
			}

			if cookie.Path != options.Path {
				t.Errorf("Cookie Path = %q, want %q", cookie.Path, options.Path)
			}

			if cookie.SameSite != tt.wantSameSite {
				t.Errorf("Cookie SameSite = %v, want %v", cookie.SameSite, tt.wantSameSite)
			}
		})
	}
}

func TestSecurityHeadersMiddlewareIntegration(t *testing.T) {
	// Setup router with security middleware
	gin.SetMode(gin.ReleaseMode)
	router := gin.New()

	cfg := &config.Config{
		Server: config.ServerConfig{
			GinMode: "release",
		},
	}

	router.Use(SecurityHeadersMiddleware(cfg))
	router.GET("/test", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"status": "ok"})
	})

	// Make request
	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/test", nil)
	router.ServeHTTP(w, req)

	// Verify response code
	if w.Code != http.StatusOK {
		t.Errorf("Response code = %d, want %d", w.Code, http.StatusOK)
	}

	// Verify all security headers are present
	headers := []string{
		"Strict-Transport-Security",
		"X-Content-Type-Options",
		"X-Frame-Options",
		"X-XSS-Protection",
		"Referrer-Policy",
		"Content-Security-Policy",
		"Permissions-Policy",
	}

	for _, header := range headers {
		if w.Header().Get(header) == "" {
			t.Errorf("Missing security header: %s", header)
		}
	}
}

// Helper function
func contains(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || len(s) > len(substr) && (s[:len(substr)] == substr || s[len(s)-len(substr):] == substr || containsSubstring(s, substr)))
}

func containsSubstring(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}
