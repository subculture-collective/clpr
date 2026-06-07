package config

import (
	"fmt"
	"os"
	"strconv"
	"strings"

	"github.com/joho/godotenv"
)

// Config holds all application configuration
type Config struct {
	Server          ServerConfig
	Database        DatabaseConfig
	Redis           RedisConfig
	JWT             JWTConfig
	Twitch          TwitchConfig
	CORS            CORSConfig
	WebSocket       WebSocketConfig
	OpenSearch      OpenSearchConfig
	Stripe          StripeConfig
	Sentry          SentryConfig
	Email           EmailConfig
	Embedding       EmbeddingConfig
	FeatureFlags    FeatureFlagsConfig
	Karma           KarmaConfig
	Jobs            JobsConfig
	RateLimit       RateLimitConfig
	Security        SecurityConfig
	QueryLimits     QueryLimitsConfig
	SearchLimits    SearchLimitsConfig
	HybridSearch    HybridSearchConfig
	CDN             CDNConfig
	Mirror          MirrorConfig
	Recommendations RecommendationsConfig
	Toxicity        ToxicityConfig
	NSFW            NSFWConfig
	Telemetry       TelemetryConfig
}

// ServerConfig holds server-specific configuration
type ServerConfig struct {
	Port        string
	GinMode     string
	BaseURL     string
	Environment string
	ExportDir   string
	DocsPath    string
}

// DatabaseConfig holds database connection configuration
type DatabaseConfig struct {
	Host     string
	Port     string
	User     string
	Password string
	Name     string
	SSLMode  string
}

// RedisConfig holds Redis connection configuration
type RedisConfig struct {
	Host     string
	Port     string
	Password string
	DB       int
}

// JWTConfig holds JWT authentication configuration
type JWTConfig struct {
	PrivateKey string
	PublicKey  string
}

// TwitchConfig holds Twitch API configuration
type TwitchConfig struct {
	ClientID     string
	ClientSecret string
	RedirectURI  string
}

// CORSConfig holds CORS configuration
type CORSConfig struct {
	AllowedOrigins string
}

// WebSocketConfig holds WebSocket server configuration
type WebSocketConfig struct {
	AllowedOrigins []string
}

// OpenSearchConfig holds OpenSearch configuration
type OpenSearchConfig struct {
	URL                string
	Username           string
	Password           string
	InsecureSkipVerify bool
}

// StripeConfig holds Stripe payment configuration
type StripeConfig struct {
	SecretKey            string
	WebhookSecrets       []string
	ProMonthlyPriceID    string
	ProYearlyPriceID     string
	ProMonthlyPriceCents int // Monthly price in cents (e.g., 999 for $9.99)
	ProYearlyPriceCents  int // Full yearly price in cents (e.g., 9999 for $99.99/year) - service layer converts to monthly equivalent
	SuccessURL           string
	CancelURL            string
	TaxEnabled           bool // Enable automatic tax calculation via Stripe Tax
	InvoicePDFEnabled    bool // Enable sending invoice PDFs via email
}

// SentryConfig holds Sentry error tracking configuration
type SentryConfig struct {
	DSN              string
	Environment      string
	Release          string
	TracesSampleRate float64
	Enabled          bool
}

// EmailConfig holds email notification configuration
type EmailConfig struct {
	SendGridAPIKey           string
	SendGridWebhookPublicKey string // ECDSA public key for webhook signature verification
	FromEmail                string
	FromName                 string
	Enabled                  bool
	SandboxMode              bool // Enable sandbox mode for testing (logs emails without sending)
	MaxEmailsPerHour         int
}

// EmbeddingConfig holds embedding service configuration
type EmbeddingConfig struct {
	OpenAIAPIKey             string
	APIBaseURL               string
	Model                    string
	RequestsPerMinute        int
	SchedulerIntervalMinutes int
	Enabled                  bool
}

// FeatureFlagsConfig holds feature flag configuration
type FeatureFlagsConfig struct {
	SemanticSearch       bool
	PremiumSubscriptions bool
	EmailNotifications   bool
	PushNotifications    bool
	Analytics            bool
	Moderation           bool
	DiscoveryLists       bool
}

// KarmaConfig holds karma system configuration
type KarmaConfig struct {
	InitialKarmaPoints        int  // Karma points granted to new users on signup
	SubmissionKarmaRequired   int  // Minimum karma required to submit clips
	RequireKarmaForSubmission bool // Whether to enforce karma requirement for submissions
}

// JobsConfig holds background job interval configuration
type JobsConfig struct {
	HotClipsRefreshIntervalMinutes int
	WebhookRetryIntervalMinutes    int
	WebhookRetryBatchSize          int
}

// RateLimitConfig holds rate limiting configuration
type RateLimitConfig struct {
	// Unauthenticated user limits
	UnauthenticatedRequests int // requests per window
	UnauthenticatedWindow   int // window in minutes

	// Authenticated user limits by tier
	BasicUserRequests   int // requests per hour
	PremiumUserRequests int // requests per hour

	// Endpoint-specific limits (per minute)
	ClipsListLimit       int // GET /api/v1/clips
	ClipsCreateLimit     int // POST /api/v1/clips
	FeedLimit            int // GET /api/v1/feed
	UserProfileLimit     int // GET /api/v1/users/{id}
	CommentCreateLimit   int // POST /api/v1/clips/:id/comments
	VoteLimit            int // POST votes
	FollowLimit          int // POST follow actions
	SubmissionLimit      int // POST /api/v1/submissions
	ReportLimit          int // POST /api/v1/reports
	ExportLimit          int // GET export endpoints
	AccountDeletionLimit int // POST account deletion

	// IP whitelist for bypassing rate limits (comma-separated, for development/testing)
	WhitelistIPs string
}

// SecurityConfig holds security-related configuration
type SecurityConfig struct {
	MFAEncryptionKey string // 32-byte key for AES-256 encryption of MFA secrets
}

// QueryLimitsConfig holds database query limits
type QueryLimitsConfig struct {
	MaxResultSize   int // Maximum rows per query (default: 1000)
	MaxOffset       int // Maximum pagination offset (default: 1000)
	MaxJoinDepth    int // Maximum number of joins (default: 3)
	MaxQueryTimeSec int // Maximum query execution time in seconds (default: 10)
}

// SearchLimitsConfig holds OpenSearch query limits
type SearchLimitsConfig struct {
	MaxResultSize      int // Maximum results per search (default: 100)
	MaxAggregationSize int // Maximum aggregation buckets (default: 100)
	MaxAggregationNest int // Maximum aggregation depth (default: 2)
	MaxQueryClauses    int // Maximum query clauses (default: 20)
	MaxSearchTimeSec   int // Maximum search timeout in seconds (default: 5)
	MaxOffset          int // Maximum search pagination offset (default: 1000)
}

// CDNConfig holds CDN provider configuration
type CDNConfig struct {
	Enabled          bool
	Provider         string // cloudflare, bunny, aws-cloudfront
	CloudflareZoneID string
	CloudflareAPIKey string
	BunnyAPIKey      string
	BunnyStorageZone string
	AWSAccessKeyID   string
	AWSSecretKey     string
	AWSRegion        string
	CacheTTL         int     // Cache TTL in seconds (default: 3600)
	MaxCostPerGB     float64 // Maximum cost per GB in USD (default: 0.10)
}

// MirrorConfig holds mirror hosting configuration
type MirrorConfig struct {
	Enabled                bool
	Regions                []string // List of regions to mirror clips to (e.g., us-east-1, eu-west-1)
	ReplicationThreshold   int      // Minimum view count to trigger mirroring (default: 1000)
	TTLDays                int      // Mirror TTL in days (default: 7)
	MaxMirrorsPerClip      int      // Maximum mirrors per clip (default: 3)
	SyncIntervalMinutes    int      // Interval to check for popular clips (default: 60)
	CleanupIntervalMinutes int      // Interval to cleanup expired mirrors (default: 1440, 24 hours)
	MinMirrorHitRate       float64  // Minimum mirror hit rate percentage (default: 60.0)
}

// RecommendationsConfig holds recommendation algorithm configuration
type RecommendationsConfig struct {
	// Hybrid algorithm weights (should sum to ~1.0)
	ContentWeight       float64 // Weight for content-based filtering (default: 0.5)
	CollaborativeWeight float64 // Weight for collaborative filtering (default: 0.3)
	TrendingWeight      float64 // Weight for trending signal (default: 0.2)

	// Collaborative filtering parameters
	CFFactors        int     // Number of latent factors for matrix factorization (default: 50)
	CFRegularization float64 // L2 regularization parameter (default: 0.01)
	CFLearningRate   float64 // Learning rate for SGD (default: 0.01)
	CFIterations     int     // Number of training iterations (default: 20)

	// Cold start and trending parameters
	TrendingWindowDays   int     // Number of days to look back for trending clips (default: 7)
	TrendingMinScore     float64 // Minimum trending score threshold (default: 0.0)
	PopularityWindowDays int     // Number of days for popularity calculation (default: 30)
	PopularityMinViews   int     // Minimum views for popularity ranking (default: 100)

	// General settings
	EnableHybrid  bool // Enable hybrid recommendations (default: true)
	CacheTTLHours int  // Cache TTL in hours (default: 24)
}

// HybridSearchConfig holds hybrid search weight configuration
type HybridSearchConfig struct {
	// Ranking algorithm weights (should sum to 1.0)
	BM25Weight   float64 // Weight for BM25 text matching (default: 0.7)
	VectorWeight float64 // Weight for semantic vector search (default: 0.3)

	// Field boost parameters for BM25
	TitleBoost   float64 // Field boost for title (default: 3.0)
	CreatorBoost float64 // Field boost for creator name (default: 2.0)
	GameBoost    float64 // Field boost for game name (default: 1.0)

	// Scoring boost parameters
	EngagementBoost float64 // Boost factor for engagement score (default: 0.1)
	RecencyBoost    float64 // Boost factor for recency (default: 0.5)
}

// ToxicityConfig holds toxicity detection configuration
type ToxicityConfig struct {
	Enabled   bool    // Enable toxicity detection (default: false)
	APIKey    string  // API key for toxicity detection service (e.g., Perspective API)
	APIURL    string  // API URL for toxicity detection service
	Threshold float64 // Confidence threshold for flagging content (default: 0.85)
}

// NSFWConfig holds NSFW image detection configuration
type NSFWConfig struct {
	Enabled        bool    // Enable NSFW detection (default: false)
	APIKey         string  // API key for NSFW detection service
	APIURL         string  // API URL for NSFW detection service (e.g., Sightengine, AWS Rekognition)
	Threshold      float64 // Confidence threshold for flagging content (default: 0.80)
	ScanThumbnails bool    // Enable scanning of thumbnails at upload (default: true)
	AutoFlag       bool    // Automatically flag content to moderation queue (default: true)
	MaxLatencyMs   int     // Maximum acceptable latency in milliseconds (default: 200)
	TimeoutSeconds int     // Request timeout in seconds (default: 5)
}

// TelemetryConfig holds distributed tracing configuration
type TelemetryConfig struct {
	Enabled          bool    // Enable OpenTelemetry tracing (default: false)
	ServiceName      string  // Service name for traces (default: "clpr-backend")
	ServiceVersion   string  // Service version for traces
	OTLPEndpoint     string  // OTLP endpoint for trace export (default: "localhost:4317")
	Insecure         bool    // Use insecure connection to OTLP endpoint (default: true for development)
	TracesSampleRate float64 // Sampling rate for traces (0.0 to 1.0, default: 0.1 for 10%)
	Environment      string  // Environment name (development, staging, production)
}

// getEnvBool gets a boolean environment variable with a fallback default value
func getEnvBool(key string, defaultValue bool) bool {
	if value := os.Getenv(key); value != "" {
		if boolVal, err := strconv.ParseBool(value); err == nil {
			return boolVal
		}
	}
	return defaultValue
}

// parseCommaSeparatedList parses a comma-separated string into a slice of trimmed strings
func parseCommaSeparatedList(value string) []string {
	if value == "" {
		return []string{}
	}
	items := strings.Split(value, ",")
	result := make([]string, 0, len(items))
	for _, item := range items {
		trimmed := strings.TrimSpace(item)
		if trimmed != "" {
			result = append(result, trimmed)
		}
	}
	return result
}

// Load loads configuration from environment variables
func Load() (*Config, error) {
	// Load .env file if it exists
	_ = godotenv.Load()

	redisDB, err := strconv.Atoi(getEnv("REDIS_DB", "0"))
	if err != nil {
		redisDB = 0
	}

	config := &Config{
		Server: ServerConfig{
			Port:        getEnv("PORT", "8080"),
			GinMode:     getEnv("GIN_MODE", "debug"),
			BaseURL:     getEnv("BASE_URL", "http://localhost:5173"),
			Environment: getEnv("ENVIRONMENT", "development"),
			ExportDir:   getEnv("EXPORT_DIR", "./exports"),
			DocsPath:    getEnv("DOCS_PATH", "../docs"),
		},
		Database: DatabaseConfig{
			Host:     getEnv("DB_HOST", "localhost"),
			Port:     getEnv("DB_PORT", "5432"),
			User:     getEnv("DB_USER", "clpr"),
			Password: getEnv("DB_PASSWORD", "CHANGEME_SECURE_PASSWORD_HERE"),
			Name:     getEnv("DB_NAME", "clpr_db"),
			SSLMode:  getEnv("DB_SSLMODE", "disable"),
		},
		Redis: RedisConfig{
			Host:     getEnv("REDIS_HOST", "localhost"),
			Port:     getEnv("REDIS_PORT", "6379"),
			Password: getEnv("REDIS_PASSWORD", ""),
			DB:       redisDB,
		},
		JWT: JWTConfig{
			PrivateKey: getEnv("JWT_PRIVATE_KEY", ""),
			PublicKey:  getEnv("JWT_PUBLIC_KEY", ""),
		},
		Twitch: TwitchConfig{
			ClientID:     getEnv("TWITCH_CLIENT_ID", ""),
			ClientSecret: getEnv("TWITCH_CLIENT_SECRET", ""),
			RedirectURI:  getEnv("TWITCH_REDIRECT_URI", "http://localhost:8080/api/v1/auth/twitch/callback"),
		},
		CORS: CORSConfig{
			AllowedOrigins: getEnv("CORS_ALLOWED_ORIGINS", "http://localhost:5173,http://localhost:3000"),
		},
		WebSocket: WebSocketConfig{
			AllowedOrigins: parseCommaSeparatedList(getEnv("WEBSOCKET_ALLOWED_ORIGINS", "http://localhost:5173,http://localhost:3000")),
		},
		OpenSearch: OpenSearchConfig{
			URL:                getEnv("OPENSEARCH_URL", "http://localhost:9200"),
			Username:           getEnv("OPENSEARCH_USERNAME", ""),
			Password:           getEnv("OPENSEARCH_PASSWORD", ""),
			InsecureSkipVerify: getEnv("OPENSEARCH_INSECURE_SKIP_VERIFY", "true") == "true",
		},
		Stripe: StripeConfig{
			SecretKey:            getEnv("STRIPE_SECRET_KEY", ""),
			WebhookSecrets:       collectStripeWebhookSecrets(),
			ProMonthlyPriceID:    getEnv("STRIPE_PRO_MONTHLY_PRICE_ID", ""),
			ProYearlyPriceID:     getEnv("STRIPE_PRO_YEARLY_PRICE_ID", ""),
			ProMonthlyPriceCents: getEnvInt("STRIPE_PRO_MONTHLY_PRICE_CENTS", 999), // Default: $9.99/month
			ProYearlyPriceCents:  getEnvInt("STRIPE_PRO_YEARLY_PRICE_CENTS", 9999), // Default: $99.99/year (full yearly price)
			SuccessURL:           getEnv("STRIPE_SUCCESS_URL", "http://localhost:5173/subscription/success"),
			CancelURL:            getEnv("STRIPE_CANCEL_URL", "http://localhost:5173/subscription/cancel"),
			TaxEnabled:           getEnv("STRIPE_TAX_ENABLED", "false") == "true",
			InvoicePDFEnabled:    getEnv("STRIPE_INVOICE_PDF_ENABLED", "false") == "true",
		},
		Sentry: SentryConfig{
			DSN:              getEnv("SENTRY_DSN", ""),
			Environment:      getEnv("SENTRY_ENVIRONMENT", "development"),
			Release:          getEnv("SENTRY_RELEASE", ""),
			TracesSampleRate: getEnvFloat("SENTRY_TRACES_SAMPLE_RATE", 1.0),
			Enabled:          getEnv("SENTRY_ENABLED", "false") == "true",
		},
		Email: EmailConfig{
			SendGridAPIKey:           getEnv("SENDGRID_API_KEY", ""),
			SendGridWebhookPublicKey: getEnv("SENDGRID_WEBHOOK_PUBLIC_KEY", ""),
			FromEmail:                getEnv("EMAIL_FROM_ADDRESS", "noreply@clpr.gg"),
			FromName:                 getEnv("EMAIL_FROM_NAME", "clpr"),
			Enabled:                  getEnv("EMAIL_ENABLED", "false") == "true",
			SandboxMode:              getEnv("EMAIL_SANDBOX_MODE", "false") == "true",
			MaxEmailsPerHour:         getEnvInt("EMAIL_MAX_PER_HOUR", 10),
		},
		Embedding: EmbeddingConfig{
			OpenAIAPIKey:             getEnv("OPENAI_API_KEY", ""),
			APIBaseURL:               getEnv("EMBEDDING_API_BASE_URL", ""),
			Model:                    getEnv("EMBEDDING_MODEL", "text-embedding-3-small"),
			RequestsPerMinute:        getEnvInt("EMBEDDING_REQUESTS_PER_MINUTE", 500),
			SchedulerIntervalMinutes: getEnvInt("EMBEDDING_SCHEDULER_INTERVAL_MINUTES", 360),
			Enabled:                  getEnv("EMBEDDING_ENABLED", "false") == "true",
		},
		FeatureFlags: FeatureFlagsConfig{
			SemanticSearch:       getEnv("FEATURE_SEMANTIC_SEARCH", "false") == "true",
			PremiumSubscriptions: getEnv("FEATURE_PREMIUM_SUBSCRIPTIONS", "false") == "true",
			EmailNotifications:   getEnv("FEATURE_EMAIL_NOTIFICATIONS", "false") == "true",
			PushNotifications:    getEnv("FEATURE_PUSH_NOTIFICATIONS", "false") == "true",
			Analytics:            getEnv("FEATURE_ANALYTICS", "true") == "true",
			Moderation:           getEnv("FEATURE_MODERATION", "true") == "true",
			DiscoveryLists:       getEnv("FEATURE_DISCOVERY_LISTS", "false") == "true",
		},
		Karma: KarmaConfig{
			InitialKarmaPoints:        getEnvInt("KARMA_INITIAL_POINTS", 100),
			SubmissionKarmaRequired:   getEnvInt("KARMA_SUBMISSION_REQUIRED", 100),
			RequireKarmaForSubmission: getEnv("KARMA_REQUIRE_FOR_SUBMISSION", "true") == "true",
		},
		Jobs: JobsConfig{
			HotClipsRefreshIntervalMinutes: getEnvInt("HOT_CLIPS_REFRESH_INTERVAL_MINUTES", 5),
			WebhookRetryIntervalMinutes:    getEnvInt("WEBHOOK_RETRY_INTERVAL_MINUTES", 1),
			WebhookRetryBatchSize:          getEnvInt("WEBHOOK_RETRY_BATCH_SIZE", 100),
		},
		RateLimit: RateLimitConfig{
			// Unauthenticated: 100 requests per 15 minutes per IP
			UnauthenticatedRequests: getEnvInt("RATE_LIMIT_UNAUTH_REQUESTS", 100),
			UnauthenticatedWindow:   getEnvInt("RATE_LIMIT_UNAUTH_WINDOW_MINUTES", 15),

			// Authenticated: Basic user: 1,000 requests per hour
			BasicUserRequests: getEnvInt("RATE_LIMIT_BASIC_REQUESTS", 1000),
			// Authenticated: Premium user: 5,000 requests per hour
			PremiumUserRequests: getEnvInt("RATE_LIMIT_PREMIUM_REQUESTS", 5000),

			// Endpoint-specific limits (per minute)
			ClipsListLimit:       getEnvInt("RATE_LIMIT_CLIPS_LIST", 100),
			ClipsCreateLimit:     getEnvInt("RATE_LIMIT_CLIPS_CREATE", 10),
			FeedLimit:            getEnvInt("RATE_LIMIT_FEED", 30),
			UserProfileLimit:     getEnvInt("RATE_LIMIT_USER_PROFILE", 200),
			CommentCreateLimit:   getEnvInt("RATE_LIMIT_COMMENT_CREATE", 10),
			VoteLimit:            getEnvInt("RATE_LIMIT_VOTE", 20),
			FollowLimit:          getEnvInt("RATE_LIMIT_FOLLOW", 20),
			SubmissionLimit:      getEnvInt("RATE_LIMIT_SUBMISSION", 5),
			ReportLimit:          getEnvInt("RATE_LIMIT_REPORT", 10),
			ExportLimit:          getEnvInt("RATE_LIMIT_EXPORT", 1),
			AccountDeletionLimit: getEnvInt("RATE_LIMIT_ACCOUNT_DELETION", 1),

			// IP whitelist for development/testing (localhost always included)
			WhitelistIPs: getEnv("RATE_LIMIT_WHITELIST_IPS", ""),
		},
		Security: SecurityConfig{
			MFAEncryptionKey: getEnv("MFA_ENCRYPTION_KEY", ""),
		},
		QueryLimits: QueryLimitsConfig{
			MaxResultSize:   getEnvInt("QUERY_MAX_RESULT_SIZE", 1000),
			MaxOffset:       getEnvInt("QUERY_MAX_OFFSET", 1000),
			MaxJoinDepth:    getEnvInt("QUERY_MAX_JOIN_DEPTH", 3),
			MaxQueryTimeSec: getEnvInt("QUERY_MAX_TIME_SEC", 10),
		},
		SearchLimits: SearchLimitsConfig{
			MaxResultSize:      getEnvInt("SEARCH_MAX_RESULT_SIZE", 100),
			MaxAggregationSize: getEnvInt("SEARCH_MAX_AGGREGATION_SIZE", 100),
			MaxAggregationNest: getEnvInt("SEARCH_MAX_AGGREGATION_NEST", 2),
			MaxQueryClauses:    getEnvInt("SEARCH_MAX_QUERY_CLAUSES", 20),
			MaxSearchTimeSec:   getEnvInt("SEARCH_MAX_TIME_SEC", 5),
			MaxOffset:          getEnvInt("SEARCH_MAX_OFFSET", 1000),
		},
		CDN: CDNConfig{
			Enabled:          getEnvBool("CDN_ENABLED", false),
			Provider:         getEnv("CDN_PROVIDER", "cloudflare"),
			CloudflareZoneID: getEnv("CDN_CLOUDFLARE_ZONE_ID", ""),
			CloudflareAPIKey: getEnv("CDN_CLOUDFLARE_API_KEY", ""),
			BunnyAPIKey:      getEnv("CDN_BUNNY_API_KEY", ""),
			BunnyStorageZone: getEnv("CDN_BUNNY_STORAGE_ZONE", ""),
			AWSAccessKeyID:   getEnv("CDN_AWS_ACCESS_KEY_ID", ""),
			AWSSecretKey:     getEnv("CDN_AWS_SECRET_KEY", ""),
			AWSRegion:        getEnv("CDN_AWS_REGION", "us-east-1"),
			CacheTTL:         getEnvInt("CDN_CACHE_TTL", 3600),
			MaxCostPerGB:     getEnvFloat("CDN_MAX_COST_PER_GB", 0.10),
		},
		Mirror: MirrorConfig{
			Enabled:                getEnvBool("MIRROR_ENABLED", false),
			Regions:                parseCommaSeparatedList(getEnv("MIRROR_REGIONS", "us-east-1,eu-west-1,ap-southeast-1")),
			ReplicationThreshold:   getEnvInt("MIRROR_REPLICATION_THRESHOLD", 1000),
			TTLDays:                getEnvInt("MIRROR_TTL_DAYS", 7),
			MaxMirrorsPerClip:      getEnvInt("MIRROR_MAX_PER_CLIP", 3),
			SyncIntervalMinutes:    getEnvInt("MIRROR_SYNC_INTERVAL_MINUTES", 60),
			CleanupIntervalMinutes: getEnvInt("MIRROR_CLEANUP_INTERVAL_MINUTES", 1440),
			MinMirrorHitRate:       getEnvFloat("MIRROR_MIN_HIT_RATE", 60.0),
		},
		Recommendations: RecommendationsConfig{
			// Hybrid algorithm weights
			ContentWeight:       getEnvFloat("REC_CONTENT_WEIGHT", 0.5),
			CollaborativeWeight: getEnvFloat("REC_COLLABORATIVE_WEIGHT", 0.3),
			TrendingWeight:      getEnvFloat("REC_TRENDING_WEIGHT", 0.2),

			// Collaborative filtering parameters
			CFFactors:        getEnvInt("REC_CF_FACTORS", 50),
			CFRegularization: getEnvFloat("REC_CF_REGULARIZATION", 0.01),
			CFLearningRate:   getEnvFloat("REC_CF_LEARNING_RATE", 0.01),
			CFIterations:     getEnvInt("REC_CF_ITERATIONS", 20),

			// Cold start and trending parameters
			TrendingWindowDays:   getEnvInt("REC_TRENDING_WINDOW_DAYS", 7),
			TrendingMinScore:     getEnvFloat("REC_TRENDING_MIN_SCORE", 0.0),
			PopularityWindowDays: getEnvInt("REC_POPULARITY_WINDOW_DAYS", 30),
			PopularityMinViews:   getEnvInt("REC_POPULARITY_MIN_VIEWS", 100),

			// General settings
			EnableHybrid:  getEnvBool("REC_ENABLE_HYBRID", true),
			CacheTTLHours: getEnvInt("REC_CACHE_TTL_HOURS", 24),
		},
		HybridSearch: HybridSearchConfig{
			// Ranking algorithm weights
			BM25Weight:   getEnvFloat("HYBRID_SEARCH_BM25_WEIGHT", 0.7),
			VectorWeight: getEnvFloat("HYBRID_SEARCH_VECTOR_WEIGHT", 0.3),

			// Field boost parameters
			TitleBoost:   getEnvFloat("HYBRID_SEARCH_TITLE_BOOST", 3.0),
			CreatorBoost: getEnvFloat("HYBRID_SEARCH_CREATOR_BOOST", 2.0),
			GameBoost:    getEnvFloat("HYBRID_SEARCH_GAME_BOOST", 1.0),

			// Scoring boost parameters
			EngagementBoost: getEnvFloat("HYBRID_SEARCH_ENGAGEMENT_BOOST", 0.1),
			RecencyBoost:    getEnvFloat("HYBRID_SEARCH_RECENCY_BOOST", 0.5),
		},
		Toxicity: ToxicityConfig{
			Enabled:   getEnvBool("TOXICITY_ENABLED", false),
			APIKey:    getEnv("TOXICITY_API_KEY", ""),
			APIURL:    getEnv("TOXICITY_API_URL", "https://commentanalyzer.googleapis.com/v1alpha1/comments:analyze"),
			Threshold: getEnvFloat("TOXICITY_THRESHOLD", 0.85),
		},
		NSFW: NSFWConfig{
			Enabled:        getEnvBool("NSFW_ENABLED", false),
			APIKey:         getEnv("NSFW_API_KEY", ""),
			APIURL:         getEnv("NSFW_API_URL", ""),
			Threshold:      getEnvFloat("NSFW_THRESHOLD", 0.80),
			ScanThumbnails: getEnvBool("NSFW_SCAN_THUMBNAILS", true),
			AutoFlag:       getEnvBool("NSFW_AUTO_FLAG", true),
			MaxLatencyMs:   getEnvInt("NSFW_MAX_LATENCY_MS", 200),
			TimeoutSeconds: getEnvInt("NSFW_TIMEOUT_SECONDS", 5),
		},
		Telemetry: TelemetryConfig{
			Enabled:          getEnvBool("TELEMETRY_ENABLED", false),
			ServiceName:      getEnv("TELEMETRY_SERVICE_NAME", "clpr-backend"),
			ServiceVersion:   getEnv("TELEMETRY_SERVICE_VERSION", ""),
			OTLPEndpoint:     getEnv("TELEMETRY_OTLP_ENDPOINT", "localhost:4317"),
			Insecure:         getEnvBool("TELEMETRY_INSECURE", true),
			TracesSampleRate: clampFloat(getEnvFloat("TELEMETRY_TRACES_SAMPLE_RATE", 0.1), 0.0, 1.0),
			Environment:      getEnv("TELEMETRY_ENVIRONMENT", getEnv("ENVIRONMENT", "development")),
		},
	}

	return config, nil
}

// GetDatabaseURL returns a PostgreSQL connection string
func (c *DatabaseConfig) GetDatabaseURL() string {
	return fmt.Sprintf(
		"postgres://%s:%s@%s:%s/%s?sslmode=%s",
		c.User,
		c.Password,
		c.Host,
		c.Port,
		c.Name,
		c.SSLMode,
	)
}

// getEnv gets an environment variable with a fallback default value
func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

// getEnvFloat gets a float environment variable with a fallback default value
func getEnvFloat(key string, defaultValue float64) float64 {
	if value := os.Getenv(key); value != "" {
		if floatVal, err := strconv.ParseFloat(value, 64); err == nil {
			return floatVal
		}
	}
	return defaultValue
}

// getEnvInt gets an int environment variable with a fallback default value
func getEnvInt(key string, defaultValue int) int {
	if value := os.Getenv(key); value != "" {
		if intVal, err := strconv.Atoi(value); err == nil {
			return intVal
		}
	}
	return defaultValue
}

// clampFloat clamps a float64 value between min and max
func clampFloat(value, min, max float64) float64 {
	if value < min {
		return min
	}
	if value > max {
		return max
	}
	return value
}

// collectStripeWebhookSecrets gathers the configured Stripe webhook secrets, supporting
// one primary secret plus optional alternates without requiring multiple endpoints.
func collectStripeWebhookSecrets() []string {
	secrets := make([]string, 0, 3)
	add := func(raw string) {
		if v := strings.TrimSpace(raw); v != "" {
			secrets = append(secrets, v)
		}
	}
	add(getEnv("STRIPE_WEBHOOK_SECRET", ""))
	add(getEnv("STRIPE_WEBHOOK_SECRET_ALT", ""))
	for _, part := range strings.Split(getEnv("STRIPE_WEBHOOK_SECRETS", ""), ",") {
		add(part)
	}
	return secrets
}
