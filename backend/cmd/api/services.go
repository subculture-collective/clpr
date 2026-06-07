package main

import (
	"context"
	"log"
	"time"

	"git.subcult.tv/subculture-collective/clpr/config"
	"git.subcult.tv/subculture-collective/clpr/internal/services"
	"git.subcult.tv/subculture-collective/clpr/internal/websocket"
	"git.subcult.tv/subculture-collective/clpr/pkg/utils"
)

// Services holds all application service instances.
type Services struct {
	Auth                  *services.AuthService
	Email                 *services.EmailService
	MFA                   *services.MFAService
	Notification          *services.NotificationService
	ToxicityClassifier    *services.ToxicityClassifier
	NSFWDetector          *services.NSFWDetector
	Comment               *services.CommentService
	Clip                  *services.ClipService
	AutoTag               *services.AutoTagService
	Reputation            *services.ReputationService
	Analytics             *services.AnalyticsService
	Engagement            *services.EngagementService
	AuditLog              *services.AuditLogService
	AccountMerge          *services.AccountMergeService
	Dunning               *services.DunningService
	Subscription          *services.SubscriptionService
	WebhookRetry          *services.WebhookRetryService
	UserSettings          *services.UserSettingsService
	Revenue               *services.RevenueService
	Ad                    *services.AdService
	EmailMetrics          *services.EmailMetricsService
	Cache                 *services.CacheService
	Feed                  *services.FeedService
	FilterPreset          *services.FilterPresetService
	Community             *services.CommunityService
	Moderation            *services.ModerationService
	BanReasonTemplate     *services.BanReasonTemplateService
	AccountType           *services.AccountTypeService
	Recommendation        *services.RecommendationService
	Playlist              *services.PlaylistService
	PlaylistScript        *services.PlaylistScriptService
	Queue                 *services.QueueService
	ClipExtractionJob     *services.ClipExtractionJobService
	WatchParty            *services.WatchPartyService
	WatchPartyHubManager  *services.WatchPartyHubManager
	EventTracker          *services.EventTracker
	Export                *services.ExportService
	SearchIndexer         *services.SearchIndexerService   // may be nil
	OpenSearch            *services.OpenSearchService      // may be nil
	HybridSearch          *services.HybridSearchService    // may be nil
	Embedding             *services.EmbeddingService       // may be nil
	ClipSync              *services.ClipSyncService        // may be nil
	Submission            *services.SubmissionService      // may be nil
	LiveStatus            *services.LiveStatusService      // may be nil
	OutboundWebhook       *services.OutboundWebhookService
	TwitchBanSync         *services.TwitchBanSyncService         // may be nil
	TwitchModeration      *services.TwitchModerationService      // may be nil
	WSServer              *websocket.Server
	CancelEventTracker    context.CancelFunc
	Logger                *utils.StructuredLogger
}

func initServices(cfg *config.Config, repos *Repositories, infra *Infrastructure, logger *utils.StructuredLogger) *Services {
	pool := infra.DB.Pool

	authService := services.NewAuthService(cfg, repos.User, repos.RefreshToken, infra.Redis, infra.JWTManager)

	// Initialize email service with notification repo for preference checking
	emailService := services.NewEmailService(&services.EmailConfig{
		SendGridAPIKey:   cfg.Email.SendGridAPIKey,
		FromEmail:        cfg.Email.FromEmail,
		FromName:         cfg.Email.FromName,
		BaseURL:          cfg.Server.BaseURL,
		Enabled:          cfg.Email.Enabled,
		SandboxMode:      cfg.Email.SandboxMode,
		MaxEmailsPerHour: cfg.Email.MaxEmailsPerHour,
	}, repos.EmailNotification, repos.Notification)

	// Initialize MFA service
	mfaService, mfaErr := services.NewMFAService(cfg, repos.MFA, repos.User, emailService)
	if mfaErr != nil {
		log.Fatalf("Failed to initialize MFA service: %v", mfaErr)
	}

	notificationService := services.NewNotificationService(repos.Notification, repos.User, repos.Comment, repos.Clip, repos.Favorite, emailService)

	// Initialize toxicity classifier
	toxicityClassifier := services.NewToxicityClassifier(
		cfg.Toxicity.APIKey,
		cfg.Toxicity.APIURL,
		cfg.Toxicity.Enabled,
		cfg.Toxicity.Threshold,
		pool,
	)

	// Initialize NSFW detector
	nsfwDetector := services.NewNSFWDetector(
		cfg.NSFW.APIKey,
		cfg.NSFW.APIURL,
		cfg.NSFW.Enabled,
		cfg.NSFW.Threshold,
		cfg.NSFW.ScanThumbnails,
		cfg.NSFW.AutoFlag,
		cfg.NSFW.MaxLatencyMs,
		cfg.NSFW.TimeoutSeconds,
		pool,
	)

	commentService := services.NewCommentService(repos.Comment, repos.Clip, repos.User, notificationService, toxicityClassifier)
	clipService := services.NewClipService(repos.Clip, repos.DiscoveryClip, repos.Vote, repos.Favorite, repos.User, repos.WatchHistory, infra.Redis, repos.AuditLog, notificationService)
	autoTagService := services.NewAutoTagService(repos.Tag)
	reputationService := services.NewReputationService(repos.Reputation, repos.User)
	analyticsService := services.NewAnalyticsService(repos.Analytics, repos.Clip)
	engagementService := services.NewEngagementService(repos.Analytics, repos.User, repos.Clip)
	auditLogService := services.NewAuditLogService(repos.AuditLog)

	// Initialize account merge service
	accountMergeService := services.NewAccountMergeService(
		pool,
		repos.User,
		repos.AuditLog,
		repos.Vote,
		repos.Favorite,
		repos.Comment,
		repos.Clip,
		repos.WatchHistory,
	)

	// Initialize dunning service before subscription service
	dunningService := services.NewDunningService(repos.Dunning, repos.Subscription, repos.User, emailService, auditLogService)

	subscriptionService := services.NewSubscriptionService(repos.Subscription, repos.User, repos.Webhook, cfg, auditLogService, dunningService, emailService)
	webhookRetryService := services.NewWebhookRetryService(repos.Webhook, subscriptionService)
	userSettingsService := services.NewUserSettingsService(repos.User, repos.UserSettings, repos.AccountDeletion, repos.Clip, repos.Vote, repos.Favorite, repos.Comment, repos.Submission, repos.Subscription, repos.Consent, auditLogService)
	revenueService := services.NewRevenueService(repos.Revenue, cfg)
	adService := services.NewAdService(repos.Ad, infra.Redis)

	// Initialize email monitoring and metrics service
	emailMetricsService := services.NewEmailMetricsService(repos.EmailLog)

	// Initialize cache service
	cacheService := services.NewCacheService(infra.Redis)

	// Initialize feed service
	feedService := services.NewFeedService(repos.Feed, repos.Clip, repos.User, repos.Broadcaster, repos.Vote, repos.Favorite)

	// Initialize filter preset service
	filterPresetService := services.NewFilterPresetService(repos.FilterPreset)

	// Initialize community service
	communityService := services.NewCommunityService(repos.Community, repos.Clip, repos.User, notificationService)

	// Initialize moderation service for ban management
	moderationService := services.NewModerationService(pool, repos.Community, repos.User, repos.AuditLog)

	// Initialize ban reason template service
	banReasonTemplateService := services.NewBanReasonTemplateService(repos.BanReasonTemplate, repos.Community, logger)

	// Initialize account type service
	accountTypeService := services.NewAccountTypeService(repos.User, repos.AccountTypeConversion, repos.AuditLog, mfaService)

	// Initialize recommendation service
	recommendationService := services.NewRecommendationServiceWithConfig(
		repos.Recommendation,
		infra.Redis.GetClient(),
		cfg.Recommendations.ContentWeight,
		cfg.Recommendations.CollaborativeWeight,
		cfg.Recommendations.TrendingWeight,
		cfg.Recommendations.EnableHybrid,
		cfg.Recommendations.CacheTTLHours,
		cfg.Recommendations.TrendingWindowDays,
		cfg.Recommendations.TrendingMinScore,
		cfg.Recommendations.PopularityWindowDays,
		cfg.Recommendations.PopularityMinViews,
	)

	// Initialize playlist service
	playlistService := services.NewPlaylistService(repos.Playlist, repos.Clip, cfg.Server.BaseURL)
	// Note: clipSyncService is initialized later (line ~268) and may be nil when Twitch is not configured.
	// We set it on playlistScriptService after clipSyncService is created.
	playlistScriptService := services.NewPlaylistScriptService(repos.PlaylistScript, repos.Playlist, repos.Clip, repos.PlaylistCuration, nil)

	// Initialize queue service
	queueService := services.NewQueueService(repos.Queue, repos.Clip, playlistService)

	// Initialize clip extraction job service for FFmpeg processing
	clipExtractionJobService := services.NewClipExtractionJobService(infra.Redis)

	// Initialize watch party service and hub manager
	watchPartyService := services.NewWatchPartyService(repos.WatchParty, repos.Playlist, repos.Clip, cfg.Server.BaseURL)

	// Initialize rate limiters for watch party (distributed via Redis)
	// 10 messages per minute per user
	chatRateLimiter := services.NewDistributedRateLimiter(infra.Redis, 10, time.Minute)
	// 30 reactions per minute per user
	reactRateLimiter := services.NewDistributedRateLimiter(infra.Redis, 30, time.Minute)

	watchPartyHubManager := services.NewWatchPartyHubManager(repos.WatchParty, chatRateLimiter, reactRateLimiter)

	// Initialize event tracker for feed analytics
	eventTracker := services.NewEventTracker(pool, 100, 5*time.Second)
	// Start event tracker in background with cancellable context for graceful shutdown
	eventTrackerCtx, cancelEventTracker := context.WithCancel(context.Background())
	go eventTracker.Start(eventTrackerCtx)

	// Initialize export service with exports directory
	exportDir := cfg.Server.ExportDir
	if exportDir == "" {
		exportDir = "./exports"
	}
	// Default retention period is 7 days
	exportRetentionDays := 7
	exportService := services.NewExportService(repos.Export, repos.User, emailService, notificationService, exportDir, cfg.Server.BaseURL, exportRetentionDays)

	// Initialize search and embedding services
	var searchIndexerService *services.SearchIndexerService
	var openSearchService *services.OpenSearchService
	var hybridSearchService *services.HybridSearchService
	var embeddingService *services.EmbeddingService

	// Initialize embedding service if enabled and configured
	if cfg.Embedding.Enabled {
		if cfg.Embedding.OpenAIAPIKey == "" && cfg.Embedding.APIBaseURL == "" {
			log.Println("WARNING: Embedding is enabled but OPENAI_API_KEY is not set; disabling embeddings")
		} else {
			embeddingService = services.NewEmbeddingService(&services.EmbeddingConfig{
				APIKey:            cfg.Embedding.OpenAIAPIKey,
				APIBaseURL:        cfg.Embedding.APIBaseURL,
				Model:             cfg.Embedding.Model,
				RedisClient:       infra.Redis.GetClient(),
				RequestsPerMinute: cfg.Embedding.RequestsPerMinute,
			})
			log.Printf("Embedding service initialized (model: %s)", cfg.Embedding.Model)
		}
	}
	if infra.OpenSearch != nil {
		searchIndexerService = services.NewSearchIndexerService(infra.OpenSearch)
		openSearchService = services.NewOpenSearchService(infra.OpenSearch)

		// Initialize indices in background
		go func() {
			ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
			defer cancel()
			if err := searchIndexerService.InitializeIndices(ctx); err != nil {
				log.Printf("WARNING: Failed to initialize search indices: %v", err)
			} else {
				log.Println("Search indices initialized successfully")
			}
		}()

		// Initialize hybrid search when OpenSearch is available
		hybridSearchService = services.NewHybridSearchService(&services.HybridSearchConfig{
			Pool:              pool,
			OpenSearchService: openSearchService,
			EmbeddingService:  embeddingService,
			RedisClient:       infra.Redis.GetClient(),
		})
	}

	var clipSyncService *services.ClipSyncService
	var submissionService *services.SubmissionService
	var liveStatusService *services.LiveStatusService
	outboundWebhookService := services.NewOutboundWebhookService(repos.OutboundWebhook)
	if infra.TwitchClient != nil {
		clipSyncService = services.NewClipSyncService(infra.TwitchClient, repos.Clip, repos.Tag, repos.User, infra.Redis)
		submissionService = services.NewSubmissionService(repos.Submission, repos.Clip, repos.DiscoveryClip, repos.User, repos.Vote, repos.AuditLog, infra.TwitchClient, notificationService, infra.Redis, outboundWebhookService, cacheService, cfg)
		liveStatusService = services.NewLiveStatusService(repos.Broadcaster, repos.StreamFollow, infra.TwitchClient)
		// Set notification service for live status notifications
		liveStatusService.SetNotificationService(notificationService)
		// Enable Twitch-powered playlist strategies
		playlistScriptService.SetClipSyncService(clipSyncService)
	}

	// Initialize Twitch-related services
	var twitchBanSyncService *services.TwitchBanSyncService
	var twitchModerationService *services.TwitchModerationService
	if infra.TwitchClient != nil {
		twitchBanSyncService = services.NewTwitchBanSyncService(infra.TwitchClient, repos.TwitchAuth, repos.TwitchBan, repos.User)
		twitchModerationService = services.NewTwitchModerationService(infra.TwitchClient, repos.TwitchAuth, repos.User, repos.AuditLog)
	}

	// Initialize WebSocket server
	wsServer := websocket.NewServer(pool, infra.Redis.GetClient(), &cfg.WebSocket)

	return &Services{
		Auth:                 authService,
		Email:                emailService,
		MFA:                  mfaService,
		Notification:         notificationService,
		ToxicityClassifier:   toxicityClassifier,
		NSFWDetector:         nsfwDetector,
		Comment:              commentService,
		Clip:                 clipService,
		AutoTag:              autoTagService,
		Reputation:           reputationService,
		Analytics:            analyticsService,
		Engagement:           engagementService,
		AuditLog:             auditLogService,
		AccountMerge:         accountMergeService,
		Dunning:              dunningService,
		Subscription:         subscriptionService,
		WebhookRetry:         webhookRetryService,
		UserSettings:         userSettingsService,
		Revenue:              revenueService,
		Ad:                   adService,
		EmailMetrics:         emailMetricsService,
		Cache:                cacheService,
		Feed:                 feedService,
		FilterPreset:         filterPresetService,
		Community:            communityService,
		Moderation:           moderationService,
		BanReasonTemplate:    banReasonTemplateService,
		AccountType:          accountTypeService,
		Recommendation:       recommendationService,
		Playlist:             playlistService,
		PlaylistScript:       playlistScriptService,
		Queue:                queueService,
		ClipExtractionJob:    clipExtractionJobService,
		WatchParty:           watchPartyService,
		WatchPartyHubManager: watchPartyHubManager,
		EventTracker:         eventTracker,
		Export:               exportService,
		SearchIndexer:        searchIndexerService,
		OpenSearch:           openSearchService,
		HybridSearch:         hybridSearchService,
		Embedding:            embeddingService,
		ClipSync:             clipSyncService,
		Submission:           submissionService,
		LiveStatus:           liveStatusService,
		OutboundWebhook:      outboundWebhookService,
		TwitchBanSync:        twitchBanSyncService,
		TwitchModeration:     twitchModerationService,
		WSServer:             wsServer,
		CancelEventTracker:   cancelEventTracker,
		Logger:               logger,
	}
}
