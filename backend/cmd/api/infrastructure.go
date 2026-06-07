package main

import (
	"context"
	"log"
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

	// Initialize JWT manager
	var jwtManager *jwtpkg.Manager
	if cfg.JWT.PrivateKey != "" {
		manager, jwtErr := jwtpkg.NewManager(cfg.JWT.PrivateKey)
		if jwtErr != nil {
			log.Fatalf("Failed to initialize JWT manager: %v", jwtErr)
		}
		jwtManager = manager
	} else {
		// Generate new RSA key pair for development
		log.Println("WARNING: No JWT private key provided. Generating new key pair (not for production!)")
		privateKey, publicKey, keyErr := jwtpkg.GenerateRSAKeyPair()
		if keyErr != nil {
			log.Fatalf("Failed to generate RSA key pair: %v", keyErr)
		}
		log.Printf("Generated RSA key pair. Add these to your .env file:\n")
		log.Printf("JWT_PRIVATE_KEY:\n%s\n", privateKey)
		log.Printf("JWT_PUBLIC_KEY:\n%s\n", publicKey)
		manager, jwtInitErr := jwtpkg.NewManager(privateKey)
		if jwtInitErr != nil {
			log.Fatalf("Failed to initialize JWT manager: %v", jwtInitErr)
		}
		jwtManager = manager
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
