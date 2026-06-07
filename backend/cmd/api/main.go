package main

import (
	"context"
	"log"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"git.subcult.tv/subculture-collective/clpr/config"
	sentrypkg "git.subcult.tv/subculture-collective/clpr/pkg/sentry"
	telemetrypkg "git.subcult.tv/subculture-collective/clpr/pkg/telemetry"
	"git.subcult.tv/subculture-collective/clpr/pkg/utils"
)

func main() {
	// Load configuration
	cfg, cfgErr := config.Load()
	if cfgErr != nil {
		log.Fatalf("Failed to load configuration: %v", cfgErr)
	}

	// Initialize structured logger
	logLevel := utils.LogLevelInfo
	if cfg.Server.GinMode == "debug" {
		logLevel = utils.LogLevelDebug
	}
	utils.InitLogger(logLevel)
	logger := utils.GetLogger()

	// Initialize Sentry
	if initErr := sentrypkg.Init(&cfg.Sentry); initErr != nil {
		log.Printf("WARNING: Failed to initialize Sentry: %v", initErr)
	} else if cfg.Sentry.Enabled {
		log.Printf("Sentry initialized: environment=%s, release=%s", cfg.Sentry.Environment, cfg.Sentry.Release)
		defer sentrypkg.Close()
	}

	// Initialize OpenTelemetry
	if initErr := telemetrypkg.Init(&telemetrypkg.Config{
		Enabled:          cfg.Telemetry.Enabled,
		ServiceName:      cfg.Telemetry.ServiceName,
		ServiceVersion:   cfg.Telemetry.ServiceVersion,
		OTLPEndpoint:     cfg.Telemetry.OTLPEndpoint,
		Insecure:         cfg.Telemetry.Insecure,
		TracesSampleRate: cfg.Telemetry.TracesSampleRate,
		Environment:      cfg.Telemetry.Environment,
	}); initErr != nil {
		log.Printf("WARNING: Failed to initialize telemetry: %v", initErr)
	} else if cfg.Telemetry.Enabled {
		defer func() {
			ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
			defer cancel()
			if err := telemetrypkg.Shutdown(ctx); err != nil {
				log.Printf("WARNING: Failed to shutdown telemetry: %v", err)
			}
		}()
	}

	// Set Gin mode
	gin.SetMode(cfg.Server.GinMode)

	// Initialize infrastructure (DB, Redis, OpenSearch, JWT, Twitch)
	infra := initInfrastructure(cfg)
	defer infra.DB.Close()
	defer infra.Redis.Close()

	// Initialize repositories
	repos := initRepositories(infra.DB.Pool)

	// Initialize services
	svcs := initServices(cfg, repos, infra, logger)

	// Initialize handlers
	h := initHandlers(svcs, repos, infra)

	// Initialize router
	r := gin.New()

	// Apply global middleware (includes template loading and rate limit whitelist)
	applyGlobalMiddleware(r, cfg, infra, svcs, logger)

	// Register routes
	v1 := r.Group("/api/v1")
	registerPublicRoutes(r, v1, h, svcs, infra)
	registerAuthRoutes(v1, h, svcs, infra)
	registerClipRoutes(v1, h, svcs, infra)
	registerContentRoutes(v1, h, svcs, infra)
	registerUserRoutes(v1, h, svcs, infra)
	registerSocialRoutes(v1, h, svcs, infra)
	registerPlatformRoutes(v1, h, svcs, infra)
	registerAdminRoutes(v1, h, svcs, infra)

	// Start background schedulers
	schedulers := startSchedulers(svcs, repos, infra)

	// Create HTTP server
	srv := &http.Server{
		Addr:              ":" + cfg.Server.Port,
		Handler:           r,
		ReadHeaderTimeout: 10 * time.Second, // Prevent Slowloris attacks
	}

	// Start server in goroutine
	go func() {
		log.Printf("Starting server on :%s (mode: %s)", cfg.Server.Port, cfg.Server.GinMode)
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("Failed to start server: %v", err)
		}
	}()

	// Block until shutdown signal, then gracefully stop everything
	gracefulShutdown(srv, svcs, schedulers, infra)
}
