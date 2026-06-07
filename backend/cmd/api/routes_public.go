package main

import (
	"context"
	"log"
	"net/http"
	"net/http/pprof"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/subculture-collective/clipper/internal/middleware"
)

func registerPublicRoutes(r *gin.Engine, v1 *gin.RouterGroup, h *Handlers, svcs *Services, infra *Infrastructure) {
	// SEO endpoints (sitemap, robots.txt)
	r.GET("/sitemap.xml", h.SEO.GetSitemap)
	r.GET("/robots.txt", h.SEO.GetRobotsTxt)

	// Programmatic SEO pages (server-rendered HTML for crawlers)
	r.GET("/clips/streamer/:broadcasterName", h.Pages.GetStreamerPage)
	r.GET("/clips/game/:gameSlug", h.Pages.GetGamePage)
	r.GET("/clips/best/*path", func(c *gin.Context) {
		path := strings.TrimPrefix(c.Param("path"), "/")
		parts := strings.SplitN(path, "/", 2)

		if len(parts) == 2 {
			c.Params = append(c.Params, gin.Param{Key: "year", Value: parts[0]}, gin.Param{Key: "month", Value: parts[1]})
			h.Pages.GetBestOfMonthPage(c)
			return
		}

		c.Params = append(c.Params, gin.Param{Key: "period", Value: path})
		h.Pages.GetBestOfPage(c)
	})
	r.GET("/clips/streamer/:broadcasterName/:gameSlug", h.Pages.GetStreamerGamePage)

	// Health check endpoints (additional checks requiring middleware)

	// Basic health check (used by Docker HEALTHCHECK)
	r.GET("/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"status": "healthy",
		})
	})

	// Readiness check - indicates if the service is ready to serve traffic
	r.GET("/health/ready", func(c *gin.Context) {
		// Check database connection
		ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
		defer cancel()

		dbErr := infra.DB.HealthCheck(ctx)

		if dbErr != nil {
			c.JSON(http.StatusServiceUnavailable, gin.H{
				"status": "not ready",
				"error":  "database unavailable",
			})
			return
		}

		// Check Redis connection
		redisErr := infra.Redis.HealthCheck(ctx)

		if redisErr != nil {
			c.JSON(http.StatusServiceUnavailable, gin.H{
				"status": "not ready",
				"error":  "redis unavailable",
			})
			return
		}

		checks := gin.H{
			"database": "ok",
			"redis":    "ok",
		}

		// Check OpenSearch connection (optional)
		if infra.OpenSearch != nil {
			osErr := infra.OpenSearch.Ping(ctx)

			if osErr != nil {
				checks["opensearch"] = "degraded"
				log.Printf("OpenSearch health check failed (%T): %v", osErr, osErr)
			} else {
				checks["opensearch"] = "ok"
			}
		}

		c.JSON(http.StatusOK, gin.H{
			"status": "ready",
			"checks": checks,
		})
	})

	// Liveness check - indicates if the application is alive
	r.GET("/health/live", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"status": "alive",
		})
	})

	// Database statistics endpoint (for monitoring)
	r.GET("/health/stats", func(c *gin.Context) {
		stats := infra.DB.GetStats()
		c.JSON(http.StatusOK, gin.H{
			"database": gin.H{
				"acquired_conns":      stats.AcquiredConns(),
				"idle_conns":          stats.IdleConns(),
				"total_conns":         stats.TotalConns(),
				"max_conns":           stats.MaxConns(),
				"acquire_count":       stats.AcquireCount(),
				"acquire_duration_ms": stats.AcquireDuration().Milliseconds(),
			},
		})
	})

	// Cache monitoring endpoints
	r.GET("/health/cache", h.Monitoring.GetCacheStats)
	r.GET("/health/cache/check", h.Monitoring.GetCacheHealth)

	// Webhook monitoring endpoint
	r.GET("/health/webhooks", h.WebhookMonitoring.GetWebhookRetryStats)

	// Prometheus metrics endpoint (unauthenticated, for internal scraping)
	r.GET("/metrics", gin.WrapH(promhttp.Handler()))

	// Profiling and metrics endpoints (for debugging and monitoring)
	// These should be protected in production (e.g., firewall rules or internal network only)
	debug := r.Group("/debug")
	debug.Use(middleware.AuthMiddleware(svcs.Auth), middleware.RequireRole("admin"))
	{
		// Prometheus metrics endpoint
		debug.GET("/metrics", gin.WrapH(promhttp.Handler()))

		// Go pprof endpoints for profiling
		debug.GET("/pprof/", gin.WrapF(pprof.Index))
		debug.GET("/pprof/cmdline", gin.WrapF(pprof.Cmdline))
		debug.GET("/pprof/profile", gin.WrapF(pprof.Profile))
		debug.GET("/pprof/symbol", gin.WrapF(pprof.Symbol))
		debug.GET("/pprof/trace", gin.WrapF(pprof.Trace))
		debug.GET("/pprof/allocs", gin.WrapH(pprof.Handler("allocs")))
		debug.GET("/pprof/block", gin.WrapH(pprof.Handler("block")))
		debug.GET("/pprof/goroutine", gin.WrapH(pprof.Handler("goroutine")))
		debug.GET("/pprof/heap", gin.WrapH(pprof.Handler("heap")))
		debug.GET("/pprof/mutex", gin.WrapH(pprof.Handler("mutex")))
		debug.GET("/pprof/threadcreate", gin.WrapH(pprof.Handler("threadcreate")))
	}

	// Basic health check
	v1.GET("/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"status": "healthy",
		})
	})

	v1.GET("/ping", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"message": "pong",
		})
	})

	// Public config endpoint
	v1.GET("/config", h.Config.GetPublicConfig)

	// Application logs endpoint (with rate limiting)
	// Rate limit: 100 requests/minute per endpoint per IP address
	// Authenticated and anonymous users behind the same IP share this limit
	logs := v1.Group("/logs")
	{
		// Public log submission endpoint with rate limiting
		// Uses OptionalAuthMiddleware to allow both authenticated and anonymous logs
		logs.POST("",
			middleware.OptionalAuthMiddleware(svcs.Auth),
			middleware.RateLimitMiddleware(infra.Redis, 100, time.Minute),
			h.ApplicationLog.CreateLog)

		// Admin-only log stats endpoint
		logs.GET("/stats",
			middleware.AuthMiddleware(svcs.Auth),
			middleware.RequireRole("admin"),
			h.ApplicationLog.GetLogStats)
	}
}
