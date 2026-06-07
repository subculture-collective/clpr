package main

import (
	"encoding/json"
	"fmt"
	"html/template"
	"log"
	"os"
	"path/filepath"
	"strings"

	"github.com/gin-contrib/requestid"
	"github.com/gin-gonic/gin"
	"git.subcult.tv/subculture-collective/clpr/config"
	"git.subcult.tv/subculture-collective/clpr/internal/middleware"
	"git.subcult.tv/subculture-collective/clpr/pkg/utils"
)

func applyGlobalMiddleware(r *gin.Engine, cfg *config.Config, infra *Infrastructure, svcs *Services, logger *utils.StructuredLogger) {
	// Load HTML templates for pSEO pages
	templatesDir := "templates"
	if _, err := os.Stat(templatesDir); os.IsNotExist(err) {
		// Try relative to binary location
		execPath, _ := os.Executable()
		templatesDir = filepath.Join(filepath.Dir(execPath), "templates")
	}
	funcMap := template.FuncMap{
		"slugify": utils.Slugify,
		"derefStr": func(s *string) string {
			if s == nil {
				return ""
			}
			return *s
		},
		"formatViews": func(v any) string {
			var n int64
			switch val := v.(type) {
			case int:
				n = int64(val)
			case int64:
				n = val
			default:
				return "0"
			}
			if n >= 1_000_000 {
				return fmt.Sprintf("%.1fM", float64(n)/1_000_000)
			}
			if n >= 1_000 {
				return fmt.Sprintf("%.1fK", float64(n)/1_000)
			}
			return fmt.Sprintf("%d", n)
		},
		"safeJSON": func(v any) template.JS {
			if s, ok := v.(string); ok {
				return template.JS(s)
			}
			b, err := json.Marshal(v)
			if err != nil {
				return template.JS("null")
			}
			return template.JS(b)
		},
	}
	if tmpl, err := template.New("").Funcs(funcMap).ParseGlob(filepath.Join(templatesDir, "*.html")); err != nil {
		log.Printf("Warning: could not load pSEO templates: %v", err)
	} else {
		r.SetHTMLTemplate(tmpl)
	}

	// Add custom middleware
	// Request ID must come first to be available in other middleware
	r.Use(requestid.New())

	// Add OpenTelemetry middleware (if enabled)
	if cfg.Telemetry.Enabled {
		r.Use(middleware.TracingMiddleware(cfg.Telemetry.ServiceName))
	}

	// Add Sentry middleware for error tracking (if enabled)
	if cfg.Sentry.Enabled {
		r.Use(middleware.SentryMiddleware())
		r.Use(middleware.RecoverWithSentry())
	} else {
		r.Use(middleware.JSONRecoveryMiddleware())
	}

	// Use structured logger
	r.Use(logger.GinLogger())

	// Apply metrics middleware for Prometheus
	r.Use(middleware.MetricsMiddleware())

	// Apply CORS middleware
	r.Use(middleware.CORSMiddleware(cfg))

	// Apply security headers middleware
	r.Use(middleware.SecurityHeadersMiddleware(cfg))

	// Apply input validation middleware
	r.Use(middleware.InputValidationMiddleware())

	// Apply abuse detection middleware
	r.Use(middleware.AbuseDetectionMiddleware(infra.Redis))

	// Apply CSRF protection middleware (secure in production)
	r.Use(middleware.CSRFMiddleware(infra.Redis, infra.IsProduction))

	// Add middleware to inject base URL and environment into context
	r.Use(func(c *gin.Context) {
		c.Set("base_url", cfg.Server.BaseURL)
		c.Set("environment", cfg.Server.Environment)
		c.Next()
	})

	// Note: Subscription enrichment is now handled on-demand within rate limit middleware
	// to avoid unnecessary database calls for routes that don't use rate limiting

	// Initialize rate limit whitelist from configuration
	middleware.InitRateLimitWhitelist(cfg.RateLimit.WhitelistIPs)
	if cfg.RateLimit.WhitelistIPs != "" {
		// Count IPs in whitelist (split by comma, filter empty strings)
		ips := strings.Split(cfg.RateLimit.WhitelistIPs, ",")
		ipCount := 0
		for _, ip := range ips {
			if strings.TrimSpace(ip) != "" {
				ipCount++
			}
		}
		if ipCount > 0 {
			log.Printf("Rate limit whitelist configured with %d additional IP(s) (plus localhost)", ipCount)
		}
	}
}
