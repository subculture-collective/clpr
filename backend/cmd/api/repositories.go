package main

import (
	"github.com/jackc/pgx/v5/pgxpool"
	"git.subcult.tv/subculture-collective/clpr/internal/repository"
)

// Repositories holds all database repository instances.
type Repositories struct {
	User                  *repository.UserRepository
	RefreshToken          *repository.RefreshTokenRepository
	UserSettings          *repository.UserSettingsRepository
	AccountDeletion       *repository.AccountDeletionRepository
	Consent               *repository.ConsentRepository
	Clip                  *repository.ClipRepository
	Comment               *repository.CommentRepository
	Vote                  *repository.VoteRepository
	Favorite              *repository.FavoriteRepository
	Tag                   *repository.TagRepository
	Search                *repository.SearchRepository
	Submission            *repository.SubmissionRepository
	Report                *repository.ReportRepository
	Reputation            *repository.ReputationRepository
	Notification          *repository.NotificationRepository
	EmailNotification     *repository.EmailNotificationRepository
	Analytics             *repository.AnalyticsRepository
	AuditLog              *repository.AuditLogRepository
	Subscription          *repository.SubscriptionRepository
	Webhook               *repository.WebhookRepository
	OutboundWebhook       *repository.OutboundWebhookRepository
	Dunning               *repository.DunningRepository
	Contact               *repository.ContactRepository
	Revenue               *repository.RevenueRepository
	Ad                    *repository.AdRepository
	Export                *repository.ExportRepository
	Broadcaster           *repository.BroadcasterRepository
	EmailLog              *repository.EmailLogRepository
	Feed                  *repository.FeedRepository
	FilterPreset          *repository.FilterPresetRepository
	DiscoveryList         *repository.DiscoveryListRepository
	Category              *repository.CategoryRepository
	Game                  *repository.GameRepository
	Community             *repository.CommunityRepository
	AccountTypeConversion *repository.AccountTypeConversionRepository
	Verification          *repository.VerificationRepository
	Recommendation        *repository.RecommendationRepository
	Playlist              *repository.PlaylistRepository
	PlaylistScript        *repository.PlaylistScriptRepository
	PlaylistCuration      *repository.PlaylistCurationRepository
	Queue                 *repository.QueueRepository
	WatchHistory          *repository.WatchHistoryRepository
	Stream                *repository.StreamRepository
	StreamFollow          *repository.StreamFollowRepository
	WatchParty            *repository.WatchPartyRepository
	TwitchAuth            *repository.TwitchAuthRepository
	TwitchBan             *repository.TwitchBanRepository
	ApplicationLog        *repository.ApplicationLogRepository
	BanReasonTemplate     *repository.BanReasonTemplateRepository
	MFA                   *repository.MFARepository
	DiscoveryClip         *repository.DiscoveryClipRepository
}

func initRepositories(pool *pgxpool.Pool) *Repositories {
	return &Repositories{
		User:                  repository.NewUserRepository(pool),
		RefreshToken:          repository.NewRefreshTokenRepository(pool),
		UserSettings:          repository.NewUserSettingsRepository(pool),
		AccountDeletion:       repository.NewAccountDeletionRepository(pool),
		Consent:               repository.NewConsentRepository(pool),
		Clip:                  repository.NewClipRepository(pool),
		Comment:               repository.NewCommentRepository(pool),
		Vote:                  repository.NewVoteRepository(pool),
		Favorite:              repository.NewFavoriteRepository(pool),
		Tag:                   repository.NewTagRepository(pool),
		Search:                repository.NewSearchRepository(pool),
		Submission:            repository.NewSubmissionRepository(pool),
		Report:                repository.NewReportRepository(pool),
		Reputation:            repository.NewReputationRepository(pool),
		Notification:          repository.NewNotificationRepository(pool),
		EmailNotification:     repository.NewEmailNotificationRepository(pool),
		Analytics:             repository.NewAnalyticsRepository(pool),
		AuditLog:              repository.NewAuditLogRepository(pool),
		Subscription:          repository.NewSubscriptionRepository(pool),
		Webhook:               repository.NewWebhookRepository(pool),
		OutboundWebhook:       repository.NewOutboundWebhookRepository(pool),
		Dunning:               repository.NewDunningRepository(pool),
		Contact:               repository.NewContactRepository(pool),
		Revenue:               repository.NewRevenueRepository(pool),
		Ad:                    repository.NewAdRepository(pool),
		Export:                repository.NewExportRepository(pool),
		Broadcaster:           repository.NewBroadcasterRepository(pool),
		EmailLog:              repository.NewEmailLogRepository(pool),
		Feed:                  repository.NewFeedRepository(pool),
		FilterPreset:          repository.NewFilterPresetRepository(pool),
		DiscoveryList:         repository.NewDiscoveryListRepository(pool),
		Category:              repository.NewCategoryRepository(pool),
		Game:                  repository.NewGameRepository(pool),
		Community:             repository.NewCommunityRepository(pool),
		AccountTypeConversion: repository.NewAccountTypeConversionRepository(pool),
		Verification:          repository.NewVerificationRepository(pool),
		Recommendation:        repository.NewRecommendationRepository(pool),
		Playlist:              repository.NewPlaylistRepository(pool),
		PlaylistScript:        repository.NewPlaylistScriptRepository(pool),
		PlaylistCuration:      repository.NewPlaylistCurationRepository(pool),
		Queue:                 repository.NewQueueRepository(pool),
		WatchHistory:          repository.NewWatchHistoryRepository(pool),
		Stream:                repository.NewStreamRepository(pool),
		StreamFollow:          repository.NewStreamFollowRepository(pool),
		WatchParty:            repository.NewWatchPartyRepository(pool),
		TwitchAuth:            repository.NewTwitchAuthRepository(pool),
		TwitchBan:             repository.NewTwitchBanRepository(pool),
		ApplicationLog:        repository.NewApplicationLogRepository(pool),
		BanReasonTemplate:     repository.NewBanReasonTemplateRepository(pool),
		MFA:                   repository.NewMFARepository(pool),
		DiscoveryClip:         repository.NewDiscoveryClipRepository(pool),
	}
}
