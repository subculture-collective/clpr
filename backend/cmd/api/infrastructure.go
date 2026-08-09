package main

import (
	"context"
	"crypto/sha256"
	"fmt"
	"log"
	"strings"
	"time"

	"git.subcult.tv/subculture-collective/clpr/config"
	"git.subcult.tv/subculture-collective/clpr/pkg/database"
	jwtpkg "git.subcult.tv/subculture-collective/clpr/pkg/jwt"
	opensearchpkg "git.subcult.tv/subculture-collective/clpr/pkg/opensearch"
	redispkg "git.subcult.tv/subculture-collective/clpr/pkg/redis"
	"git.subcult.tv/subculture-collective/clpr/pkg/twitch"
)

// Infrastructure holds core infrastructure clients initialized at startup.
type Infrastructure struct {
	DB           *database.DB
	Redis        *redispkg.Client
	OpenSearch   *opensearchpkg.Client // may be nil
	JWTManager   *jwtpkg.Manager
	TwitchClient *twitch.Client // may be nil
	Config       *config.Config
	IsProduction bool
}

func initInfrastructure(cfg *config.Config) *Infrastructure {
	jwtManager, jwtErr := initializeJWTManager(cfg)
	if jwtErr != nil {
		log.Fatalf("Failed to initialize JWT manager: %v", jwtErr)
	}

	// Initialize database connection pool
	db, dbErr := database.NewDBWithTracing(&cfg.Database, cfg.Telemetry.Enabled)
	if dbErr != nil {
		log.Fatalf("Failed to connect to database: %v", dbErr)
	}

	// Initialize Redis client
	redisClient, redisErr := redispkg.NewClientWithTracing(&cfg.Redis, cfg.Telemetry.Enabled)
	if redisErr != nil {
		log.Fatalf("Failed to connect to Redis: %v", redisErr)
	}

	// Initialize OpenSearch client
	var osClient *opensearchpkg.Client
	client, osErr := opensearchpkg.NewClient(&opensearchpkg.Config{
		URL:                cfg.OpenSearch.URL,
		Username:           cfg.OpenSearch.Username,
		Password:           cfg.OpenSearch.Password,
		InsecureSkipVerify: cfg.OpenSearch.InsecureSkipVerify,
	})
	if osErr != nil {
		log.Printf("WARNING: Failed to initialize OpenSearch client: %v", osErr)
		log.Printf("Search features will use PostgreSQL FTS fallback")
	} else {
		// Test connection
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		if pingErr := client.Ping(ctx); pingErr != nil {
			log.Printf("WARNING: OpenSearch ping failed: %v", pingErr)
			log.Printf("Search features will use PostgreSQL FTS fallback")
		} else {
			log.Println("OpenSearch connection established")
			osClient = client
		}
	}

	// Initialize Twitch client
	twitchClient, err := twitch.NewClient(&cfg.Twitch, redisClient)
	if err != nil {
		log.Printf("WARNING: Failed to initialize Twitch client: %v", err)
		log.Printf("Twitch API features will be disabled. Please configure TWITCH_CLIENT_ID and TWITCH_CLIENT_SECRET")
	}

	isProduction := cfg.Server.GinMode == "release"

	return &Infrastructure{
		DB:           db,
		Redis:        redisClient,
		OpenSearch:   osClient,
		JWTManager:   jwtManager,
		TwitchClient: twitchClient,
		Config:       cfg,
		IsProduction: isProduction,
	}
}

func initializeJWTManager(cfg *config.Config) (*jwtpkg.Manager, error) {
	if strings.TrimSpace(cfg.JWT.PrivateKey) != "" {
		manager, err := jwtpkg.NewManager(cfg.JWT.PrivateKey)
		if err != nil {
			return nil, fmt.Errorf("parse configured private key: %w", err)
		}
		return manager, nil
	}

	if !isDevelopmentProfile(cfg) {
		return nil, fmt.Errorf("JWT_PRIVATE_KEY is required outside the development profile")
	}

	privateKey, publicKey, err := jwtpkg.GenerateRSAKeyPair()
	if err != nil {
		return nil, fmt.Errorf("generate development RSA key pair: %w", err)
	}

	fingerprint := sha256.Sum256([]byte(publicKey))
	log.Printf("WARNING: using an ephemeral development JWT key (public fingerprint sha256:%x); sessions will not survive restart", fingerprint[:8])

	manager, err := jwtpkg.NewManager(privateKey)
	if err != nil {
		return nil, fmt.Errorf("initialize generated development key: %w", err)
	}
	return manager, nil
}

func isDevelopmentProfile(cfg *config.Config) bool {
	environment := strings.ToLower(strings.TrimSpace(cfg.Server.Environment))
	if cfg.Server.GinMode == "release" {
		return false
	}

	switch environment {
	case "development", "dev", "local", "test":
		return true
	default:
		return false
	}
}
