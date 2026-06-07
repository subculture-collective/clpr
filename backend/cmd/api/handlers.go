package main

import (
	"log"

	"git.subcult.tv/subculture-collective/clpr/internal/handlers"
)

// Handlers holds all HTTP handler instances.
type Handlers struct {
	Auth                *handlers.AuthHandler
	MFA                 *handlers.MFAHandler
	Monitoring          *handlers.MonitoringHandler
	WebhookMonitoring   *handlers.WebhookMonitoringHandler
	Comment             *handlers.CommentHandler
	Clip                *handlers.ClipHandler
	Favorite            *handlers.FavoriteHandler
	Tag                 *handlers.TagHandler
	Search              *handlers.SearchHandler
	Report              *handlers.ReportHandler
	Reputation          *handlers.ReputationHandler
	Notification        *handlers.NotificationHandler
	Analytics           *handlers.AnalyticsHandler
	Engagement          *handlers.EngagementHandler
	AuditLog            *handlers.AuditLogHandler
	Subscription        *handlers.SubscriptionHandler
	User                *handlers.UserHandler
	AdminUser           *handlers.AdminUserHandler
	UserSettings        *handlers.UserSettingsHandler
	Consent             *handlers.ConsentHandler
	Contact             *handlers.ContactHandler
	SEO                 *handlers.SEOHandler
	Pages               *handlers.PagesHandler
	Docs                *handlers.DocsHandler
	Revenue             *handlers.RevenueHandler
	Ad                  *handlers.AdHandler
	Export              *handlers.ExportHandler
	WebhookSubscription *handlers.WebhookSubscriptionHandler
	WebhookDLQ          *handlers.WebhookDLQHandler
	Config              *handlers.ConfigHandler
	Broadcaster         *handlers.BroadcasterHandler
	EmailMetrics        *handlers.EmailMetricsHandler
	SendGridWebhook     *handlers.SendGridWebhookHandler
	Feed                *handlers.FeedHandler
	FilterPreset        *handlers.FilterPresetHandler
	Community           *handlers.CommunityHandler
	DiscoveryList       *handlers.DiscoveryListHandler
	Category            *handlers.CategoryHandler
	Game                *handlers.GameHandler
	AccountType         *handlers.AccountTypeHandler
	Verification        *handlers.VerificationHandler
	Chat                *handlers.ChatHandler
	ApplicationLog      *handlers.ApplicationLogHandler
	BanReasonTemplate   *handlers.BanReasonTemplateHandler
	WebSocket           *handlers.WebSocketHandler
	Recommendation      *handlers.RecommendationHandler
	Playlist            *handlers.PlaylistHandler
	PlaylistScript      *handlers.PlaylistScriptHandler
	Queue               *handlers.QueueHandler
	WatchHistory        *handlers.WatchHistoryHandler
	WatchParty          *handlers.WatchPartyHandler
	Event               *handlers.EventHandler
	ClipSync            *handlers.ClipSyncHandler       // may be nil
	Submission          *handlers.SubmissionHandler      // may be nil
	Moderation          *handlers.ModerationHandler      // may be nil
	LiveStatus          *handlers.LiveStatusHandler      // may be nil
	Stream              *handlers.StreamHandler          // may be nil
	TwitchOAuth         *handlers.TwitchOAuthHandler     // may be nil
	Forum               *handlers.ForumHandler
	ForumModeration     *handlers.ForumModerationHandler
	NSFW                *handlers.NSFWHandler
}

func initHandlers(svcs *Services, repos *Repositories, infra *Infrastructure) *Handlers {
	pool := infra.DB.Pool
	cfg := infra.Config

	authHandler := handlers.NewAuthHandler(svcs.Auth, cfg)
	mfaHandler := handlers.NewMFAHandler(svcs.MFA, cfg)
	monitoringHandler := handlers.NewMonitoringHandler(infra.Redis)
	webhookMonitoringHandler := handlers.NewWebhookMonitoringHandler(svcs.WebhookRetry, svcs.OutboundWebhook)
	commentHandler := handlers.NewCommentHandler(svcs.Comment)
	clipHandler := handlers.NewClipHandler(
		svcs.Clip,
		svcs.Auth,
		handlers.WithClipExtractionJobService(svcs.ClipExtractionJob),
	)
	favoriteHandler := handlers.NewFavoriteHandler(repos.Favorite, repos.Vote, svcs.Clip)
	tagHandler := handlers.NewTagHandler(repos.Tag, repos.Clip, svcs.AutoTag)
	searchHandler := handlers.NewSearchHandler(repos.Search, svcs.Auth)
	if svcs.HybridSearch != nil {
		// Use hybrid search (BM25 + vector similarity)
		searchHandler = handlers.NewSearchHandlerWithHybridSearch(repos.Search, svcs.HybridSearch, svcs.Auth)
		log.Println("Using hybrid search handler (BM25 + vector similarity)")
	} else if svcs.OpenSearch != nil {
		// Use OpenSearch-enhanced handler (BM25 only)
		searchHandler = handlers.NewSearchHandlerWithOpenSearch(repos.Search, svcs.OpenSearch, svcs.Auth)
		log.Println("Using OpenSearch handler (BM25 only)")
	} else {
		log.Println("Using PostgreSQL FTS handler (fallback)")
	}
	reportHandler := handlers.NewReportHandler(repos.Report, repos.Clip, repos.Comment, repos.User, svcs.Auth)
	reputationHandler := handlers.NewReputationHandler(svcs.Reputation, svcs.Auth)
	notificationHandler := handlers.NewNotificationHandler(svcs.Notification, svcs.Email)
	analyticsHandler := handlers.NewAnalyticsHandler(svcs.Analytics, svcs.Auth)
	engagementHandler := handlers.NewEngagementHandler(svcs.Engagement, svcs.Auth)
	auditLogHandler := handlers.NewAuditLogHandler(svcs.AuditLog)
	subscriptionHandler := handlers.NewSubscriptionHandler(svcs.Subscription)
	userHandler := handlers.NewUserHandler(repos.Clip, repos.Vote, repos.Comment, repos.User, repos.Broadcaster, svcs.AccountMerge)
	adminUserHandler := handlers.NewAdminUserHandler(repos.User, repos.AuditLog, svcs.Auth)
	userSettingsHandler := handlers.NewUserSettingsHandler(svcs.UserSettings, svcs.Auth)
	consentHandler := handlers.NewConsentHandler(repos.Consent)
	contactHandler := handlers.NewContactHandler(repos.Contact, svcs.Auth)
	seoHandler := handlers.NewSEOHandler(repos.Clip, repos.Game)
	pagesHandler := handlers.NewPagesHandler(repos.Clip, repos.Broadcaster, repos.Game)
	docsHandler := handlers.NewDocsHandler(cfg.Server.DocsPath, "subculture-collective", "clpr", "main")
	revenueHandler := handlers.NewRevenueHandler(svcs.Revenue)
	adHandler := handlers.NewAdHandler(svcs.Ad)
	exportHandler := handlers.NewExportHandler(svcs.Export, svcs.Auth, repos.User)
	webhookSubscriptionHandler := handlers.NewWebhookSubscriptionHandler(svcs.OutboundWebhook)
	webhookDLQHandler := handlers.NewWebhookDLQHandler(svcs.OutboundWebhook)
	configHandler := handlers.NewConfigHandler(cfg)
	broadcasterHandler := handlers.NewBroadcasterHandler(repos.Broadcaster, repos.Clip, infra.TwitchClient, svcs.Auth)
	emailMetricsHandler := handlers.NewEmailMetricsHandler(svcs.EmailMetrics, repos.EmailLog)
	sendgridWebhookHandler := handlers.NewSendGridWebhookHandler(repos.EmailLog, cfg.Email.SendGridWebhookPublicKey)
	feedHandler := handlers.NewFeedHandler(svcs.Feed, svcs.Auth, repos.Vote, repos.Favorite, repos.User)
	filterPresetHandler := handlers.NewFilterPresetHandler(svcs.FilterPreset)
	communityHandler := handlers.NewCommunityHandler(svcs.Community, svcs.Auth)
	discoveryListHandler := handlers.NewDiscoveryListHandler(repos.DiscoveryList, repos.Analytics)
	categoryHandler := handlers.NewCategoryHandler(repos.Category, repos.Clip)
	gameHandler := handlers.NewGameHandler(repos.Game, repos.Clip, svcs.Auth)
	accountTypeHandler := handlers.NewAccountTypeHandler(svcs.AccountType, svcs.Auth)
	verificationHandler := handlers.NewVerificationHandler(repos.Verification, svcs.Notification, pool)
	chatHandler := handlers.NewChatHandler(pool)
	applicationLogHandler := handlers.NewApplicationLogHandler(repos.ApplicationLog)
	banReasonTemplateHandler := handlers.NewBanReasonTemplateHandler(svcs.BanReasonTemplate, svcs.Logger)

	// Initialize WebSocket handler
	websocketHandler := handlers.NewWebSocketHandler(pool, svcs.WSServer)

	recommendationHandler := handlers.NewRecommendationHandler(svcs.Recommendation, svcs.Auth)
	playlistHandler := handlers.NewPlaylistHandler(svcs.Playlist)
	playlistScriptHandler := handlers.NewPlaylistScriptHandler(svcs.PlaylistScript)
	queueHandler := handlers.NewQueueHandler(svcs.Queue)
	watchHistoryHandler := handlers.NewWatchHistoryHandler(repos.WatchHistory)
	watchPartyHandler := handlers.NewWatchPartyHandler(svcs.WatchParty, svcs.WatchPartyHubManager, repos.WatchParty, repos.Analytics, cfg)
	eventHandler := handlers.NewEventHandler(svcs.EventTracker)

	var clipSyncHandler *handlers.ClipSyncHandler
	var submissionHandler *handlers.SubmissionHandler
	var moderationHandler *handlers.ModerationHandler
	var liveStatusHandler *handlers.LiveStatusHandler
	var streamHandler *handlers.StreamHandler
	var twitchOAuthHandler *handlers.TwitchOAuthHandler

	if svcs.ClipSync != nil {
		clipSyncHandler = handlers.NewClipSyncHandler(svcs.ClipSync, cfg)
	}

	if svcs.LiveStatus != nil {
		liveStatusHandler = handlers.NewLiveStatusHandler(svcs.LiveStatus, svcs.Auth)
	}

	// Initialize Twitch-related handlers
	if infra.TwitchClient != nil {
		streamHandler = handlers.NewStreamHandler(infra.TwitchClient, repos.Stream, repos.Clip, repos.StreamFollow, svcs.ClipExtractionJob)
		twitchOAuthHandler = handlers.NewTwitchOAuthHandler(repos.TwitchAuth)
	}

	if svcs.Submission != nil {
		submissionHandler = handlers.NewSubmissionHandler(svcs.Submission)
		// Create moderation handler using services from submission service
		abuseDetector := svcs.Submission.GetAbuseDetector()
		moderationEventService := svcs.Submission.GetModerationEventService()
		if abuseDetector != nil && moderationEventService != nil {
			moderationHandler = handlers.NewModerationHandler(moderationEventService, svcs.Moderation, abuseDetector, svcs.ToxicityClassifier, svcs.TwitchBanSync, repos.Community, repos.AuditLog, pool)
			// Set Twitch moderation service if available
			if svcs.TwitchModeration != nil {
				moderationHandler.SetTwitchModerationService(svcs.TwitchModeration)
			}
		}
	}

	// Initialize forum handlers
	forumHandler := handlers.NewForumHandler(pool)
	forumModerationHandler := handlers.NewForumModerationHandler(pool)

	// Initialize NSFW handler
	nsfwHandler := handlers.NewNSFWHandler(svcs.NSFWDetector)

	return &Handlers{
		Auth:                authHandler,
		MFA:                 mfaHandler,
		Monitoring:          monitoringHandler,
		WebhookMonitoring:   webhookMonitoringHandler,
		Comment:             commentHandler,
		Clip:                clipHandler,
		Favorite:            favoriteHandler,
		Tag:                 tagHandler,
		Search:              searchHandler,
		Report:              reportHandler,
		Reputation:          reputationHandler,
		Notification:        notificationHandler,
		Analytics:           analyticsHandler,
		Engagement:          engagementHandler,
		AuditLog:            auditLogHandler,
		Subscription:        subscriptionHandler,
		User:                userHandler,
		AdminUser:           adminUserHandler,
		UserSettings:        userSettingsHandler,
		Consent:             consentHandler,
		Contact:             contactHandler,
		SEO:                 seoHandler,
		Pages:               pagesHandler,
		Docs:                docsHandler,
		Revenue:             revenueHandler,
		Ad:                  adHandler,
		Export:              exportHandler,
		WebhookSubscription: webhookSubscriptionHandler,
		WebhookDLQ:          webhookDLQHandler,
		Config:              configHandler,
		Broadcaster:         broadcasterHandler,
		EmailMetrics:        emailMetricsHandler,
		SendGridWebhook:     sendgridWebhookHandler,
		Feed:                feedHandler,
		FilterPreset:        filterPresetHandler,
		Community:           communityHandler,
		DiscoveryList:       discoveryListHandler,
		Category:            categoryHandler,
		Game:                gameHandler,
		AccountType:         accountTypeHandler,
		Verification:        verificationHandler,
		Chat:                chatHandler,
		ApplicationLog:      applicationLogHandler,
		BanReasonTemplate:   banReasonTemplateHandler,
		WebSocket:           websocketHandler,
		Recommendation:      recommendationHandler,
		Playlist:            playlistHandler,
		PlaylistScript:      playlistScriptHandler,
		Queue:               queueHandler,
		WatchHistory:        watchHistoryHandler,
		WatchParty:          watchPartyHandler,
		Event:               eventHandler,
		ClipSync:            clipSyncHandler,
		Submission:          submissionHandler,
		Moderation:          moderationHandler,
		LiveStatus:          liveStatusHandler,
		Stream:              streamHandler,
		TwitchOAuth:         twitchOAuthHandler,
		Forum:               forumHandler,
		ForumModeration:     forumModerationHandler,
		NSFW:                nsfwHandler,
	}
}
