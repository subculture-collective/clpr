package models

import (
	"fmt"
	"time"

	"github.com/google/uuid"
)

// User represents a user in the system
type User struct {
	ID                   uuid.UUID  `json:"id" db:"id"`
	TwitchID             *string    `json:"twitch_id,omitempty" db:"twitch_id"`
	Username             string     `json:"username" db:"username"`
	DisplayName          string     `json:"display_name" db:"display_name"`
	Email                *string    `json:"email,omitempty" db:"email"`
	AvatarURL            *string    `json:"avatar_url,omitempty" db:"avatar_url"`
	Bio                  *string    `json:"bio,omitempty" db:"bio"`
	SocialLinks          *string    `json:"social_links,omitempty" db:"social_links"` // JSONB stored as string
	KarmaPoints          int        `json:"karma_points" db:"karma_points"`
	TrustScore           int        `json:"trust_score" db:"trust_score"`
	TrustScoreUpdatedAt  *time.Time `json:"trust_score_updated_at,omitempty" db:"trust_score_updated_at"`
	Role                 string     `json:"role" db:"role"`
	AccountType          string     `json:"account_type" db:"account_type"`
	AccountTypeUpdatedAt *time.Time `json:"account_type_updated_at,omitempty" db:"account_type_updated_at"`
	// Moderator metadata fields
	ModeratorScope      string      `json:"moderator_scope,omitempty" db:"moderator_scope"`
	ModerationChannels  []uuid.UUID `json:"moderation_channels,omitempty" db:"moderation_channels"`
	ModerationStartedAt *time.Time  `json:"moderation_started_at,omitempty" db:"moderation_started_at"`
	AccountStatus       string      `json:"account_status" db:"account_status"` // active, unclaimed, pending
	IsBanned            bool        `json:"is_banned" db:"is_banned"`
	DeviceToken         *string     `json:"device_token,omitempty" db:"device_token"`
	DevicePlatform      *string     `json:"device_platform,omitempty" db:"device_platform"`
	FollowerCount       int         `json:"follower_count" db:"follower_count"`
	FollowingCount      int         `json:"following_count" db:"following_count"`
	// DMCA-related fields
	DMCAStrikesCount   int        `json:"dmca_strikes_count" db:"dmca_strikes_count"`
	DMCASuspendedUntil *time.Time `json:"dmca_suspended_until,omitempty" db:"dmca_suspended_until"`
	DMCATerminated     bool       `json:"dmca_terminated" db:"dmca_terminated"`
	DMCATerminatedAt   *time.Time `json:"dmca_terminated_at,omitempty" db:"dmca_terminated_at"`
	// Verification fields
	IsVerified bool       `json:"is_verified" db:"is_verified"`
	VerifiedAt *time.Time `json:"verified_at,omitempty" db:"verified_at"`
	// Comment moderation fields
	CommentSuspendedUntil *time.Time `json:"comment_suspended_until,omitempty" db:"comment_suspended_until"`
	CommentsRequireReview bool       `json:"comments_require_review" db:"comments_require_review"`
	CommentWarningCount   int        `json:"comment_warning_count" db:"comment_warning_count"`
	CreatedAt             time.Time  `json:"created_at" db:"created_at"`
	UpdatedAt             time.Time  `json:"updated_at" db:"updated_at"`
	LastLoginAt           *time.Time `json:"last_login_at,omitempty" db:"last_login_at"`
}

// UserSettings represents user privacy and other settings
type UserSettings struct {
	UserID            uuid.UUID `json:"user_id" db:"user_id"`
	ProfileVisibility string    `json:"profile_visibility" db:"profile_visibility"` // public, private, followers
	ShowKarmaPublicly bool      `json:"show_karma_publicly" db:"show_karma_publicly"`
	CreatedAt         time.Time `json:"created_at" db:"created_at"`
	UpdatedAt         time.Time `json:"updated_at" db:"updated_at"`
}

// AccountDeletion represents a pending account deletion request
type AccountDeletion struct {
	ID           uuid.UUID  `json:"id" db:"id"`
	UserID       uuid.UUID  `json:"user_id" db:"user_id"`
	RequestedAt  time.Time  `json:"requested_at" db:"requested_at"`
	ScheduledFor time.Time  `json:"scheduled_for" db:"scheduled_for"`
	Reason       *string    `json:"reason,omitempty" db:"reason"`
	IsCancelled  bool       `json:"is_cancelled" db:"is_cancelled"`
	CancelledAt  *time.Time `json:"cancelled_at,omitempty" db:"cancelled_at"`
	CompletedAt  *time.Time `json:"completed_at,omitempty" db:"completed_at"`
}

// UpdateProfileRequest represents the request to update user profile
type UpdateProfileRequest struct {
	DisplayName string  `json:"display_name" binding:"required,min=1,max=100"`
	Bio         *string `json:"bio" binding:"omitempty,max=500"`
}

// UpdateUserSettingsRequest represents the request to update user settings
type UpdateUserSettingsRequest struct {
	ProfileVisibility *string `json:"profile_visibility,omitempty" binding:"omitempty,oneof=public private followers"`
	ShowKarmaPublicly *bool   `json:"show_karma_publicly,omitempty"`
}

// DeleteAccountRequest represents the request to delete an account
type DeleteAccountRequest struct {
	Reason       *string `json:"reason,omitempty" binding:"omitempty,max=1000"`
	Confirmation string  `json:"confirmation" binding:"required,eq=DELETE MY ACCOUNT"`
}

// CookieConsent represents a user's cookie consent preferences
type CookieConsent struct {
	ID          uuid.UUID `json:"id" db:"id"`
	UserID      uuid.UUID `json:"user_id" db:"user_id"`
	Essential   bool      `json:"essential" db:"essential"`
	Functional  bool      `json:"functional" db:"functional"`
	Analytics   bool      `json:"analytics" db:"analytics"`
	Advertising bool      `json:"advertising" db:"advertising"`
	ConsentDate time.Time `json:"consent_date" db:"consent_date"`
	IPAddress   *string   `json:"ip_address,omitempty" db:"ip_address"`
	UserAgent   *string   `json:"user_agent,omitempty" db:"user_agent"`
	ExpiresAt   time.Time `json:"expires_at" db:"expires_at"`
	CreatedAt   time.Time `json:"created_at" db:"created_at"`
	UpdatedAt   time.Time `json:"updated_at" db:"updated_at"`
}

// UpdateConsentRequest represents the request to update cookie consent
type UpdateConsentRequest struct {
	Essential   bool `json:"essential"`
	Functional  bool `json:"functional"`
	Analytics   bool `json:"analytics"`
	Advertising bool `json:"advertising"`
}

// Clip represents a Twitch clip
type Clip struct {
	ID                   uuid.UUID  `json:"id" db:"id"`
	TwitchClipID         string     `json:"twitch_clip_id" db:"twitch_clip_id"`
	TwitchClipURL        string     `json:"twitch_clip_url" db:"twitch_clip_url"`
	EmbedURL             string     `json:"embed_url" db:"embed_url"`
	Title                string     `json:"title" db:"title"`
	CreatorName          string     `json:"creator_name" db:"creator_name"`
	CreatorID            *string    `json:"creator_id,omitempty" db:"creator_id"`
	BroadcasterName      string     `json:"broadcaster_name" db:"broadcaster_name"`
	BroadcasterID        *string    `json:"broadcaster_id,omitempty" db:"broadcaster_id"`
	GameID               *string    `json:"game_id,omitempty" db:"game_id"`
	GameName             *string    `json:"game_name,omitempty" db:"game_name"`
	Language             *string    `json:"language,omitempty" db:"language"`
	ThumbnailURL         *string    `json:"thumbnail_url,omitempty" db:"thumbnail_url"`
	Duration             *float64   `json:"duration,omitempty" db:"duration"`
	ViewCount            int        `json:"view_count" db:"view_count"`
	CreatedAt            time.Time  `json:"created_at" db:"created_at"`
	ImportedAt           time.Time  `json:"imported_at" db:"imported_at"`
	VoteScore            int        `json:"vote_score" db:"vote_score"`
	CommentCount         int        `json:"comment_count" db:"comment_count"`
	FavoriteCount        int        `json:"favorite_count" db:"favorite_count"`
	IsFeatured           bool       `json:"is_featured" db:"is_featured"`
	IsNSFW               bool       `json:"is_nsfw" db:"is_nsfw"`
	IsRemoved            bool       `json:"is_removed" db:"is_removed"`
	RemovedReason        *string    `json:"removed_reason,omitempty" db:"removed_reason"`
	IsHidden             bool       `json:"is_hidden" db:"is_hidden"`
	Embedding            []float32  `json:"embedding,omitempty" db:"embedding"`
	EmbeddingGeneratedAt *time.Time `json:"embedding_generated_at,omitempty" db:"embedding_generated_at"`
	EmbeddingModel       *string    `json:"embedding_model,omitempty" db:"embedding_model"`
	SubmittedByUserID    *uuid.UUID `json:"submitted_by_user_id,omitempty" db:"submitted_by_user_id"`
	SubmittedAt          *time.Time `json:"submitted_at,omitempty" db:"submitted_at"`
	// Trending and popularity metrics
	TrendingScore   float64 `json:"trending_score,omitempty" db:"trending_score"`
	HotScore        float64 `json:"hot_score,omitempty" db:"hot_score"`
	PopularityIndex int     `json:"popularity_index,omitempty" db:"popularity_index"`
	EngagementCount int     `json:"engagement_count,omitempty" db:"engagement_count"`
	// DMCA-related fields
	DMCARemoved      bool       `json:"dmca_removed" db:"dmca_removed"`
	DMCANoticeID     *uuid.UUID `json:"dmca_notice_id,omitempty" db:"dmca_notice_id"`
	DMCARemovedAt    *time.Time `json:"dmca_removed_at,omitempty" db:"dmca_removed_at"`
	DMCAReinstatedAt *time.Time `json:"dmca_reinstated_at,omitempty" db:"dmca_reinstated_at"`
	// Stream clip fields
	StreamSource *string    `json:"stream_source,omitempty" db:"stream_source"` // 'twitch' or 'stream'
	Status       *string    `json:"status,omitempty" db:"status"`               // 'ready', 'processing', 'failed'
	VideoURL     *string    `json:"video_url,omitempty" db:"video_url"`
	ProcessedAt  *time.Time `json:"processed_at,omitempty" db:"processed_at"`
	Quality      *string    `json:"quality,omitempty" db:"quality"` // 'source', '1080p', '720p'
	StartTime    *float64   `json:"start_time,omitempty" db:"start_time"`
	EndTime      *float64   `json:"end_time,omitempty" db:"end_time"`
	// CDN and mirror fields
	PrimaryCDNURL    *string    `json:"primary_cdn_url,omitempty" db:"primary_cdn_url"`
	CDNProvider      *string    `json:"cdn_provider,omitempty" db:"cdn_provider"`
	IsMirrored       bool       `json:"is_mirrored" db:"is_mirrored"`
	MirrorCount      int        `json:"mirror_count" db:"mirror_count"`
	LastMirrorSyncAt *time.Time `json:"last_mirror_sync_at,omitempty" db:"last_mirror_sync_at"`
	// Watch progress (populated from watch history, not in database)
	WatchProgress *WatchProgressInfo `json:"watch_progress,omitempty" db:"-"`
}

// WatchProgressInfo represents watch progress for a clip (used in API responses)
// Note: WatchedAt is optional and may be empty for performance reasons in list views
type WatchProgressInfo struct {
	ProgressSeconds int     `json:"progress_seconds"`
	DurationSeconds int     `json:"duration_seconds"`
	ProgressPercent float64 `json:"progress_percent"`
	Completed       bool    `json:"completed"`
	WatchedAt       string  `json:"watched_at,omitempty"` // Optional: May be omitted for performance
}

// DiscoveryClip represents a scraped clip in the discovery_clips staging table.
// These clips have not been posted by any user yet. When a user claims one,
// it is moved into the main clips table and deleted from discovery_clips.
type DiscoveryClip struct {
	ID              uuid.UUID `json:"id" db:"id"`
	TwitchClipID    string    `json:"twitch_clip_id" db:"twitch_clip_id"`
	TwitchClipURL   string    `json:"twitch_clip_url" db:"twitch_clip_url"`
	EmbedURL        string    `json:"embed_url" db:"embed_url"`
	Title           string    `json:"title" db:"title"`
	CreatorName     string    `json:"creator_name" db:"creator_name"`
	CreatorID       *string   `json:"creator_id,omitempty" db:"creator_id"`
	BroadcasterName string    `json:"broadcaster_name" db:"broadcaster_name"`
	BroadcasterID   *string   `json:"broadcaster_id,omitempty" db:"broadcaster_id"`
	GameID          *string   `json:"game_id,omitempty" db:"game_id"`
	GameName        *string   `json:"game_name,omitempty" db:"game_name"`
	Language        *string   `json:"language,omitempty" db:"language"`
	ThumbnailURL    *string   `json:"thumbnail_url,omitempty" db:"thumbnail_url"`
	Duration        *float64  `json:"duration,omitempty" db:"duration"`
	ViewCount       int       `json:"view_count" db:"view_count"`
	CreatedAt       time.Time `json:"created_at" db:"created_at"`
	ImportedAt      time.Time `json:"imported_at" db:"imported_at"`
	IsNSFW          bool      `json:"is_nsfw" db:"is_nsfw"`
	IsRemoved       bool      `json:"is_removed" db:"is_removed"`
	IsHidden        bool      `json:"is_hidden" db:"is_hidden"`
}

// Vote represents a user's vote on a clip
type Vote struct {
	ID        uuid.UUID `json:"id" db:"id"`
	UserID    uuid.UUID `json:"user_id" db:"user_id"`
	ClipID    uuid.UUID `json:"clip_id" db:"clip_id"`
	VoteType  int16     `json:"vote_type" db:"vote_type"` // 1 for upvote, -1 for downvote
	CreatedAt time.Time `json:"created_at" db:"created_at"`
}

// Comment represents a user comment on a clip
type Comment struct {
	ID              uuid.UUID  `json:"id" db:"id"`
	ClipID          uuid.UUID  `json:"clip_id" db:"clip_id"`
	UserID          uuid.UUID  `json:"user_id" db:"user_id"`
	ParentCommentID *uuid.UUID `json:"parent_comment_id,omitempty" db:"parent_comment_id"`
	Content         string     `json:"content" db:"content"`
	VoteScore       int        `json:"vote_score" db:"vote_score"`
	ReplyCount      int        `json:"reply_count" db:"reply_count"`
	IsEdited        bool       `json:"is_edited" db:"is_edited"`
	IsRemoved       bool       `json:"is_removed" db:"is_removed"`
	RemovedReason   *string    `json:"removed_reason,omitempty" db:"removed_reason"`
	CreatedAt       time.Time  `json:"created_at" db:"created_at"`
	UpdatedAt       time.Time  `json:"updated_at" db:"updated_at"`
}

// CommentVote represents a user's vote on a comment
type CommentVote struct {
	ID        uuid.UUID `json:"id" db:"id"`
	UserID    uuid.UUID `json:"user_id" db:"user_id"`
	CommentID uuid.UUID `json:"comment_id" db:"comment_id"`
	VoteType  int16     `json:"vote_type" db:"vote_type"` // 1 for upvote, -1 for downvote
	CreatedAt time.Time `json:"created_at" db:"created_at"`
}

// Favorite represents a user's favorite clip
type Favorite struct {
	ID        uuid.UUID `json:"id" db:"id"`
	UserID    uuid.UUID `json:"user_id" db:"user_id"`
	ClipID    uuid.UUID `json:"clip_id" db:"clip_id"`
	CreatedAt time.Time `json:"created_at" db:"created_at"`
}

// Tag represents a categorization tag
type Tag struct {
	ID          uuid.UUID `json:"id" db:"id"`
	Name        string    `json:"name" db:"name"`
	Slug        string    `json:"slug" db:"slug"`
	Description *string   `json:"description,omitempty" db:"description"`
	Color       *string   `json:"color,omitempty" db:"color"`
	UsageCount  int       `json:"usage_count" db:"usage_count"`
	CreatedAt   time.Time `json:"created_at" db:"created_at"`
}

// ClipTag represents the many-to-many relationship between clips and tags
type ClipTag struct {
	ClipID    uuid.UUID `json:"clip_id" db:"clip_id"`
	TagID     uuid.UUID `json:"tag_id" db:"tag_id"`
	CreatedAt time.Time `json:"created_at" db:"created_at"`
}

// BlacklistedTag represents a tag pattern that should be excluded from listings
type BlacklistedTag struct {
	ID        uuid.UUID  `json:"id" db:"id"`
	Pattern   string     `json:"pattern" db:"pattern"`
	Reason    *string    `json:"reason,omitempty" db:"reason"`
	CreatedBy *uuid.UUID `json:"created_by,omitempty" db:"created_by"`
	CreatedAt time.Time  `json:"created_at" db:"created_at"`
}

// Report represents a user report for moderation
type Report struct {
	ID             uuid.UUID  `json:"id" db:"id"`
	ReporterID     uuid.UUID  `json:"reporter_id" db:"reporter_id"`
	ReportableType string     `json:"reportable_type" db:"reportable_type"` // 'clip', 'comment', 'user'
	ReportableID   uuid.UUID  `json:"reportable_id" db:"reportable_id"`
	Reason         string     `json:"reason" db:"reason"`
	Description    *string    `json:"description,omitempty" db:"description"`
	Status         string     `json:"status" db:"status"` // pending, reviewed, actioned, dismissed
	ReviewedBy     *uuid.UUID `json:"reviewed_by,omitempty" db:"reviewed_by"`
	ReviewedAt     *time.Time `json:"reviewed_at,omitempty" db:"reviewed_at"`
	CreatedAt      time.Time  `json:"created_at" db:"created_at"`
}

// ClipWithHotScore represents a clip with calculated hot score
type ClipWithHotScore struct {
	Clip
	HotScore float64 `json:"hot_score" db:"hot_score"`
}

// ClipSubmitterInfo represents basic info about the user who submitted a clip
type ClipSubmitterInfo struct {
	ID          uuid.UUID `json:"id"`
	Username    string    `json:"username"`
	DisplayName string    `json:"display_name"`
	AvatarURL   *string   `json:"avatar_url,omitempty"`
}

// ClipWithSubmitter represents a clip with submitter information
type ClipWithSubmitter struct {
	Clip
	SubmittedBy *ClipSubmitterInfo `json:"submitted_by,omitempty"`
}

// SearchRequest represents a search query request
type SearchRequest struct {
	Query     string   `json:"query" form:"q"`
	Type      string   `json:"type" form:"type"` // clips, creators, games, tags, all
	Sort      string   `json:"sort" form:"sort"` // relevance (default), recent, popular
	GameID    *string  `json:"game_id" form:"game_id"`
	CreatorID *string  `json:"creator_id" form:"creator_id"`
	Language  *string  `json:"language" form:"language"`
	Tags      []string `json:"tags" form:"tags"`
	MinVotes  *int     `json:"min_votes" form:"min_votes"`
	DateFrom  *string  `json:"date_from" form:"date_from"`
	DateTo    *string  `json:"date_to" form:"date_to"`
	Page      int      `json:"page" form:"page"`
	Limit     int      `json:"limit" form:"limit"`
}

// SearchResponse represents search results
type SearchResponse struct {
	Query   string              `json:"query"`
	Results SearchResultsByType `json:"results"`
	Counts  SearchCounts        `json:"counts"`
	Facets  SearchFacets        `json:"facets,omitempty"`
	Meta    SearchMeta          `json:"meta"`
}

// SearchResultsByType groups results by type
type SearchResultsByType struct {
	Clips    []Clip             `json:"clips,omitempty"`
	Creators []User             `json:"creators,omitempty"`
	Games    []GameSearchResult `json:"games,omitempty"`
	Tags     []Tag              `json:"tags,omitempty"`
}

// SearchCounts holds counts for each result type
type SearchCounts struct {
	Clips    int `json:"clips"`
	Creators int `json:"creators"`
	Games    int `json:"games"`
	Tags     int `json:"tags"`
}

// SearchMeta holds pagination and other metadata
type SearchMeta struct {
	Page       int `json:"page"`
	Limit      int `json:"limit"`
	TotalItems int `json:"total_items"`
	TotalPages int `json:"total_pages"`
}

// SearchFacets holds aggregated facet data for filtering
type SearchFacets struct {
	Languages []FacetBucket  `json:"languages,omitempty"`
	Games     []FacetBucket  `json:"games,omitempty"`
	Tags      []FacetBucket  `json:"tags,omitempty"`
	DateRange DateRangeFacet `json:"date_range,omitempty"`
}

// FacetBucket represents a single facet value with its count
type FacetBucket struct {
	Key   string `json:"key"`
	Label string `json:"label,omitempty"` // Human-readable label
	Count int    `json:"count"`
}

// DateRangeFacet represents date range distribution
type DateRangeFacet struct {
	LastHour  int `json:"last_hour"`
	LastDay   int `json:"last_day"`
	LastWeek  int `json:"last_week"`
	LastMonth int `json:"last_month"`
	Older     int `json:"older"`
}

// GameSearchResult represents a game in search results (aggregated from clips)
type GameSearchResult struct {
	ID        string `json:"id" db:"game_id"`
	Name      string `json:"name" db:"game_name"`
	ClipCount int    `json:"clip_count" db:"clip_count"`
}

// SearchSuggestion represents an autocomplete suggestion
type SearchSuggestion struct {
	Text string `json:"text"`
	Type string `json:"type"` // query, game, creator, tag
}

// SearchQuery tracks a search query for analytics
type SearchQuery struct {
	ID                uuid.UUID  `json:"id" db:"id"`
	UserID            *uuid.UUID `json:"user_id,omitempty" db:"user_id"`
	Query             string     `json:"query" db:"query"`
	Filters           *string    `json:"filters,omitempty" db:"filters"`
	ResultCount       int        `json:"result_count" db:"result_count"`
	ClickedResultID   *uuid.UUID `json:"clicked_result_id,omitempty" db:"clicked_result_id"`
	ClickedResultType *string    `json:"clicked_result_type,omitempty" db:"clicked_result_type"`
	CreatedAt         time.Time  `json:"created_at" db:"created_at"`
}

// ClipScore represents the relevance score for a clip in search results
type ClipScore struct {
	ClipID          uuid.UUID `json:"clip_id"`
	SimilarityScore float64   `json:"similarity_score"` // 0-1, higher is better
	SimilarityRank  int       `json:"similarity_rank"`  // 1-based ranking
}

// SearchResponseWithScores extends SearchResponse with similarity scores
type SearchResponseWithScores struct {
	SearchResponse
	Scores []ClipScore `json:"scores,omitempty"`
}

// TrendingSearch represents a popular search query
type TrendingSearch struct {
	Query       string `json:"query"`
	SearchCount int    `json:"search_count"`
	UniqueUsers int    `json:"unique_users"`
	AvgResults  int    `json:"avg_results"`
}

// FailedSearch represents a search query that returned no results
type FailedSearch struct {
	Query        string    `json:"query"`
	SearchCount  int       `json:"search_count"`
	LastSearched time.Time `json:"last_searched"`
}

// SearchHistoryItem represents a single search in a user's history
type SearchHistoryItem struct {
	Query       string    `json:"query"`
	ResultCount int       `json:"result_count"`
	CreatedAt   time.Time `json:"created_at"`
}

// SearchAnalyticsSummary represents overall search analytics
type SearchAnalyticsSummary struct {
	TotalSearches       int     `json:"total_searches"`
	UniqueUsers         int     `json:"unique_users"`
	FailedSearches      int     `json:"failed_searches"`
	AvgResultsPerSearch int     `json:"avg_results_per_search"`
	SuccessRate         float64 `json:"success_rate"`
}

// ClipSubmission represents a user-submitted clip pending moderation
type ClipSubmission struct {
	ID                      uuid.UUID  `json:"id" db:"id"`
	UserID                  uuid.UUID  `json:"user_id" db:"user_id"`
	ClipID                  *uuid.UUID `json:"clip_id,omitempty" db:"clip_id"` // Set when submission is approved
	TwitchClipID            string     `json:"twitch_clip_id" db:"twitch_clip_id"`
	TwitchClipURL           string     `json:"twitch_clip_url" db:"twitch_clip_url"`
	Title                   *string    `json:"title,omitempty" db:"title"`
	CustomTitle             *string    `json:"custom_title,omitempty" db:"custom_title"`
	BroadcasterNameOverride *string    `json:"broadcaster_name_override,omitempty" db:"broadcaster_name_override"`
	Tags                    []string   `json:"tags,omitempty" db:"tags"`
	IsNSFW                  bool       `json:"is_nsfw" db:"is_nsfw"`
	SubmissionReason        *string    `json:"submission_reason,omitempty" db:"submission_reason"`
	Status                  string     `json:"status" db:"status"` // pending, approved, rejected
	RejectionReason         *string    `json:"rejection_reason,omitempty" db:"rejection_reason"`
	ReviewedBy              *uuid.UUID `json:"reviewed_by,omitempty" db:"reviewed_by"`
	ReviewedAt              *time.Time `json:"reviewed_at,omitempty" db:"reviewed_at"`
	CreatedAt               time.Time  `json:"created_at" db:"created_at"`
	UpdatedAt               time.Time  `json:"updated_at" db:"updated_at"`
	// Metadata from Twitch
	CreatorName     *string  `json:"creator_name,omitempty" db:"creator_name"`
	CreatorID       *string  `json:"creator_id,omitempty" db:"creator_id"`
	BroadcasterName *string  `json:"broadcaster_name,omitempty" db:"broadcaster_name"`
	BroadcasterID   *string  `json:"broadcaster_id,omitempty" db:"broadcaster_id"`
	GameID          *string  `json:"game_id,omitempty" db:"game_id"`
	GameName        *string  `json:"game_name,omitempty" db:"game_name"`
	ThumbnailURL    *string  `json:"thumbnail_url,omitempty" db:"thumbnail_url"`
	Duration        *float64 `json:"duration,omitempty" db:"duration"`
	ViewCount       int      `json:"view_count" db:"view_count"`
}

// ClipSubmissionWithUser includes user information
type ClipSubmissionWithUser struct {
	ClipSubmission
	User *User `json:"user,omitempty"`
}

// SubmissionStats represents submission statistics for a user
type SubmissionStats struct {
	UserID        uuid.UUID `json:"user_id" db:"user_id"`
	TotalCount    int       `json:"total_submissions" db:"total_submissions"`
	ApprovedCount int       `json:"approved_count" db:"approved_count"`
	RejectedCount int       `json:"rejected_count" db:"rejected_count"`
	PendingCount  int       `json:"pending_count" db:"pending_count"`
	ApprovalRate  float64   `json:"approval_rate" db:"approval_rate"`
}

// ModerationAuditLog represents an audit log entry for moderation actions
type ModerationAuditLog struct {
	ID          uuid.UUID              `json:"id" db:"id"`
	Action      string                 `json:"action" db:"action"`           // approve, reject, bulk_approve, bulk_reject
	EntityType  string                 `json:"entity_type" db:"entity_type"` // clip_submission, clip, comment, user, channel
	EntityID    uuid.UUID              `json:"entity_id" db:"entity_id"`
	ModeratorID uuid.UUID              `json:"moderator_id" db:"moderator_id"`
	Reason      *string                `json:"reason,omitempty" db:"reason"`
	Metadata    map[string]interface{} `json:"metadata,omitempty" db:"metadata"`
	IPAddress   *string                `json:"ip_address,omitempty" db:"ip_address"`
	UserAgent   *string                `json:"user_agent,omitempty" db:"user_agent"`
	ChannelID   *uuid.UUID             `json:"channel_id,omitempty" db:"channel_id"`
	CreatedAt   time.Time              `json:"created_at" db:"created_at"`
}

// ModerationAuditLogWithUser includes moderator information
type ModerationAuditLogWithUser struct {
	ModerationAuditLog
	Moderator *User `json:"moderator,omitempty"`
}

// RejectionReason constants for common rejection reasons
const (
	RejectionReasonLowQuality         = "Low quality clip"
	RejectionReasonDuplicate          = "Duplicate content"
	RejectionReasonInappropriate      = "Inappropriate content"
	RejectionReasonOffTopic           = "Off-topic or irrelevant"
	RejectionReasonPoorTitle          = "Poor or misleading title"
	RejectionReasonTooShort           = "Clip too short"
	RejectionReasonTooLong            = "Clip too long"
	RejectionReasonSpam               = "Spam or promotional content"
	RejectionReasonViolatesGuidelines = "Violates community guidelines"
	RejectionReasonOther              = "Other (see notes)"
)

// GetRejectionReasonTemplates returns a list of common rejection reason templates
func GetRejectionReasonTemplates() []string {
	return []string{
		RejectionReasonLowQuality,
		RejectionReasonDuplicate,
		RejectionReasonInappropriate,
		RejectionReasonOffTopic,
		RejectionReasonPoorTitle,
		RejectionReasonTooShort,
		RejectionReasonTooLong,
		RejectionReasonSpam,
		RejectionReasonViolatesGuidelines,
		RejectionReasonOther,
	}
}

// UserBadge represents a badge awarded to a user
type UserBadge struct {
	ID        uuid.UUID  `json:"id" db:"id"`
	UserID    uuid.UUID  `json:"user_id" db:"user_id"`
	BadgeID   string     `json:"badge_id" db:"badge_id"`
	AwardedAt time.Time  `json:"awarded_at" db:"awarded_at"`
	AwardedBy *uuid.UUID `json:"awarded_by,omitempty" db:"awarded_by"`
}

// KarmaHistory represents a karma change event
type KarmaHistory struct {
	ID        uuid.UUID  `json:"id" db:"id"`
	UserID    uuid.UUID  `json:"user_id" db:"user_id"`
	Amount    int        `json:"amount" db:"amount"`
	Source    string     `json:"source" db:"source"`
	SourceID  *uuid.UUID `json:"source_id,omitempty" db:"source_id"`
	CreatedAt time.Time  `json:"created_at" db:"created_at"`
}

// UserStats represents user statistics for reputation
type UserStats struct {
	UserID           uuid.UUID  `json:"user_id" db:"user_id"`
	TrustScore       int        `json:"trust_score" db:"trust_score"`
	EngagementScore  int        `json:"engagement_score" db:"engagement_score"`
	TotalComments    int        `json:"total_comments" db:"total_comments"`
	TotalVotesCast   int        `json:"total_votes_cast" db:"total_votes_cast"`
	TotalClipsSubmit int        `json:"total_clips_submitted" db:"total_clips_submitted"`
	CorrectReports   int        `json:"correct_reports" db:"correct_reports"`
	IncorrectReports int        `json:"incorrect_reports" db:"incorrect_reports"`
	DaysActive       int        `json:"days_active" db:"days_active"`
	LastActiveDate   *time.Time `json:"last_active_date,omitempty" db:"last_active_date"`
	UpdatedAt        time.Time  `json:"updated_at" db:"updated_at"`
}

// Badge represents a badge definition
type Badge struct {
	ID          string `json:"id"`
	Name        string `json:"name"`
	Description string `json:"description"`
	Icon        string `json:"icon"`
	Category    string `json:"category"` // achievement, staff, supporter, special
	Requirement string `json:"requirement,omitempty"`
}

// UserReputation represents complete reputation info for a user
type UserReputation struct {
	UserID          uuid.UUID   `json:"user_id"`
	Username        string      `json:"username"`
	DisplayName     string      `json:"display_name"`
	AvatarURL       *string     `json:"avatar_url,omitempty"`
	KarmaPoints     int         `json:"karma_points"`
	Rank            string      `json:"rank"`
	TrustScore      int         `json:"trust_score"`
	EngagementScore int         `json:"engagement_score"`
	Badges          []UserBadge `json:"badges"`
	Stats           *UserStats  `json:"stats,omitempty"`
	CreatedAt       time.Time   `json:"created_at"`
}

// KarmaBreakdown represents karma sources breakdown
type KarmaBreakdown struct {
	ClipKarma    int `json:"clip_karma"`
	CommentKarma int `json:"comment_karma"`
	TotalKarma   int `json:"total_karma"`
}

// LeaderboardEntry represents a user entry in leaderboard
type LeaderboardEntry struct {
	Rank             int       `json:"rank"`
	UserID           uuid.UUID `json:"user_id" db:"id"`
	Username         string    `json:"username" db:"username"`
	DisplayName      string    `json:"display_name" db:"display_name"`
	AvatarURL        *string   `json:"avatar_url,omitempty" db:"avatar_url"`
	Score            int       `json:"score"` // Karma or engagement score
	UserRank         string    `json:"user_rank" db:"rank"`
	AccountAge       string    `json:"account_age,omitempty"`
	TotalComments    *int      `json:"total_comments,omitempty" db:"total_comments"`
	TotalVotesCast   *int      `json:"total_votes_cast,omitempty" db:"total_votes_cast"`
	TotalClipsSubmit *int      `json:"total_clips_submitted,omitempty" db:"total_clips_submitted"`
}

// Notification represents a notification for a user
type Notification struct {
	ID                uuid.UUID  `json:"id" db:"id"`
	UserID            uuid.UUID  `json:"user_id" db:"user_id"`
	Type              string     `json:"type" db:"type"`
	Title             string     `json:"title" db:"title"`
	Message           string     `json:"message" db:"message"`
	Link              *string    `json:"link,omitempty" db:"link"`
	IsRead            bool       `json:"is_read" db:"is_read"`
	CreatedAt         time.Time  `json:"created_at" db:"created_at"`
	ExpiresAt         *time.Time `json:"expires_at,omitempty" db:"expires_at"`
	SourceUserID      *uuid.UUID `json:"source_user_id,omitempty" db:"source_user_id"`
	SourceContentID   *uuid.UUID `json:"source_content_id,omitempty" db:"source_content_id"`
	SourceContentType *string    `json:"source_content_type,omitempty" db:"source_content_type"`
}

// NotificationWithSource includes source user information
type NotificationWithSource struct {
	Notification
	SourceUsername    *string `json:"source_username,omitempty" db:"source_username"`
	SourceDisplayName *string `json:"source_display_name,omitempty" db:"source_display_name"`
	SourceAvatarURL   *string `json:"source_avatar_url,omitempty" db:"source_avatar_url"`
}

// NotificationPreferences represents user's notification settings
type NotificationPreferences struct {
	UserID       uuid.UUID `json:"user_id" db:"user_id"`
	InAppEnabled bool      `json:"in_app_enabled" db:"in_app_enabled"`
	EmailEnabled bool      `json:"email_enabled" db:"email_enabled"`
	EmailDigest  string    `json:"email_digest" db:"email_digest"` // immediate, daily, weekly, never

	// Account & Security
	NotifyLoginNewDevice  bool `json:"notify_login_new_device" db:"notify_login_new_device"`
	NotifyFailedLogin     bool `json:"notify_failed_login" db:"notify_failed_login"`
	NotifyPasswordChanged bool `json:"notify_password_changed" db:"notify_password_changed"`
	NotifyEmailChanged    bool `json:"notify_email_changed" db:"notify_email_changed"`

	// Content notifications
	NotifyReplies              bool `json:"notify_replies" db:"notify_replies"`
	NotifyMentions             bool `json:"notify_mentions" db:"notify_mentions"`
	NotifySubmissionApproved   bool `json:"notify_submission_approved" db:"notify_submission_approved"`
	NotifySubmissionRejected   bool `json:"notify_submission_rejected" db:"notify_submission_rejected"`
	NotifyContentTrending      bool `json:"notify_content_trending" db:"notify_content_trending"`
	NotifyContentFlagged       bool `json:"notify_content_flagged" db:"notify_content_flagged"`
	NotifyVotes                bool `json:"notify_votes" db:"notify_votes"`
	NotifyFavoritedClipComment bool `json:"notify_favorited_clip_comment" db:"notify_favorited_clip_comment"`

	// Community notifications
	NotifyModeratorMessage bool `json:"notify_moderator_message" db:"notify_moderator_message"`
	NotifyUserFollowed     bool `json:"notify_user_followed" db:"notify_user_followed"`
	NotifyCommentOnContent bool `json:"notify_comment_on_content" db:"notify_comment_on_content"`
	NotifyDiscussionReply  bool `json:"notify_discussion_reply" db:"notify_discussion_reply"`
	NotifyBadges           bool `json:"notify_badges" db:"notify_badges"`
	NotifyRankUp           bool `json:"notify_rank_up" db:"notify_rank_up"`
	NotifyModeration       bool `json:"notify_moderation" db:"notify_moderation"`

	// Creator-specific notification preferences
	NotifyClipApproved  bool `json:"notify_clip_approved" db:"notify_clip_approved"`
	NotifyClipRejected  bool `json:"notify_clip_rejected" db:"notify_clip_rejected"`
	NotifyClipComments  bool `json:"notify_clip_comments" db:"notify_clip_comments"`
	NotifyClipThreshold bool `json:"notify_clip_threshold" db:"notify_clip_threshold"`

	// Broadcaster notifications
	NotifyBroadcasterLive bool `json:"notify_broadcaster_live" db:"notify_broadcaster_live"`

	// Stream notifications
	NotifyStreamLive bool `json:"notify_stream_live" db:"notify_stream_live"`

	// Global preferences
	NotifyMarketing             bool `json:"notify_marketing" db:"notify_marketing"`
	NotifyPolicyUpdates         bool `json:"notify_policy_updates" db:"notify_policy_updates"`
	NotifyPlatformAnnouncements bool `json:"notify_platform_announcements" db:"notify_platform_announcements"`

	UpdatedAt time.Time `json:"updated_at" db:"updated_at"`
}

// Notification types constants
const (
	NotificationTypeReply                = "reply"
	NotificationTypeMention              = "mention"
	NotificationTypeVoteMilestone        = "vote_milestone"
	NotificationTypeBadgeEarned          = "badge_earned"
	NotificationTypeRankUp               = "rank_up"
	NotificationTypeFavoritedClipComment = "favorited_clip_comment"
	NotificationTypeContentRemoved       = "content_removed"
	NotificationTypeWarning              = "warning"
	NotificationTypeBan                  = "ban"
	NotificationTypeAppealDecision       = "appeal_decision"
	NotificationTypeSubmissionApproved   = "submission_approved"
	NotificationTypeSubmissionRejected   = "submission_rejected"
	NotificationTypeNewReport            = "new_report"
	NotificationTypePendingSubmissions   = "pending_submissions"
	NotificationTypeSystemAlert          = "system_alert"
	// Dunning notification types
	NotificationTypePaymentFailed          = "payment_failed"
	NotificationTypePaymentRetry           = "payment_retry"
	NotificationTypeGracePeriodWarning     = "grace_period_warning"
	NotificationTypeSubscriptionDowngraded = "subscription_downgraded"
	// Invoice notification types
	NotificationTypeInvoiceFinalized = "invoice_finalized"
	// Export notification types
	NotificationTypeExportCompleted = "export_completed"
	NotificationTypeExportFailed    = "export_failed"
	// Creator clip notification types
	NotificationTypeClipComment       = "clip_comment"
	NotificationTypeClipViewThreshold = "clip_view_threshold"
	NotificationTypeClipVoteThreshold = "clip_vote_threshold"
	// Account & Security notification types
	NotificationTypeLoginNewDevice  = "login_new_device"
	NotificationTypeFailedLogin     = "failed_login"
	NotificationTypePasswordChanged = "password_changed"
	NotificationTypeEmailChanged    = "email_changed"
	// Content notification types (additional)
	NotificationTypeContentTrending = "content_trending"
	NotificationTypeContentFlagged  = "content_flagged"
	// Community notification types (additional)
	NotificationTypeModeratorMessage = "moderator_message"
	NotificationTypeUserFollowed     = "user_followed"
	NotificationTypeCommentOnContent = "comment_on_content"
	NotificationTypeDiscussionReply  = "discussion_reply"
	// Broadcaster notification types
	NotificationTypeBroadcasterLive = "broadcaster_live"
	// Stream notification types
	NotificationTypeStreamLive = "stream_live"
	// Global/Marketing notification types
	NotificationTypeMarketing            = "marketing"
	NotificationTypePolicyUpdate         = "policy_update"
	NotificationTypePlatformAnnouncement = "platform_announcement"
)

// AnalyticsEvent represents a tracked event for analytics
type AnalyticsEvent struct {
	ID        uuid.UUID  `json:"id" db:"id"`
	EventType string     `json:"event_type" db:"event_type"`
	UserID    *uuid.UUID `json:"user_id,omitempty" db:"user_id"`
	ClipID    *uuid.UUID `json:"clip_id,omitempty" db:"clip_id"`
	Metadata  *string    `json:"metadata,omitempty" db:"metadata"` // JSON string
	IPAddress *string    `json:"ip_address,omitempty" db:"ip_address"`
	UserAgent *string    `json:"user_agent,omitempty" db:"user_agent"`
	Referrer  *string    `json:"referrer,omitempty" db:"referrer"`
	CreatedAt time.Time  `json:"created_at" db:"created_at"`
}

// DailyAnalytics represents pre-aggregated daily metrics
type DailyAnalytics struct {
	ID         uuid.UUID `json:"id" db:"id"`
	Date       time.Time `json:"date" db:"date"`
	MetricType string    `json:"metric_type" db:"metric_type"`
	EntityType *string   `json:"entity_type,omitempty" db:"entity_type"`
	EntityID   *string   `json:"entity_id,omitempty" db:"entity_id"`
	Value      int64     `json:"value" db:"value"`
	Metadata   *string   `json:"metadata,omitempty" db:"metadata"` // JSON string
	CreatedAt  time.Time `json:"created_at" db:"created_at"`
	UpdatedAt  time.Time `json:"updated_at" db:"updated_at"`
}

// ClipAnalytics represents analytics for a specific clip
type ClipAnalytics struct {
	ClipID              uuid.UUID  `json:"clip_id" db:"clip_id"`
	TotalViews          int64      `json:"total_views" db:"total_views"`
	UniqueViewers       int64      `json:"unique_viewers" db:"unique_viewers"`
	AvgViewDuration     *float64   `json:"avg_view_duration,omitempty" db:"avg_view_duration"`
	TotalShares         int64      `json:"total_shares" db:"total_shares"`
	PeakConcurrentViews int        `json:"peak_concurrent_viewers" db:"peak_concurrent_viewers"`
	RetentionRate       *float64   `json:"retention_rate,omitempty" db:"retention_rate"`
	FirstViewedAt       *time.Time `json:"first_viewed_at,omitempty" db:"first_viewed_at"`
	LastViewedAt        *time.Time `json:"last_viewed_at,omitempty" db:"last_viewed_at"`
	UpdatedAt           time.Time  `json:"updated_at" db:"updated_at"`
}

// CreatorAnalytics represents analytics for a content creator
type CreatorAnalytics struct {
	CreatorName       string    `json:"creator_name" db:"creator_name"`
	CreatorID         *string   `json:"creator_id,omitempty" db:"creator_id"`
	TotalClips        int       `json:"total_clips" db:"total_clips"`
	TotalViews        int64     `json:"total_views" db:"total_views"`
	TotalUpvotes      int64     `json:"total_upvotes" db:"total_upvotes"`
	TotalDownvotes    int64     `json:"total_downvotes" db:"total_downvotes"`
	TotalComments     int64     `json:"total_comments" db:"total_comments"`
	TotalFavorites    int64     `json:"total_favorites" db:"total_favorites"`
	AvgEngagementRate *float64  `json:"avg_engagement_rate,omitempty" db:"avg_engagement_rate"`
	FollowerCount     int       `json:"follower_count" db:"follower_count"`
	UpdatedAt         time.Time `json:"updated_at" db:"updated_at"`
}

// UserAnalytics represents personal statistics for a user
type UserAnalytics struct {
	UserID            uuid.UUID  `json:"user_id" db:"user_id"`
	ClipsUpvoted      int        `json:"clips_upvoted" db:"clips_upvoted"`
	ClipsDownvoted    int        `json:"clips_downvoted" db:"clips_downvoted"`
	CommentsPosted    int        `json:"comments_posted" db:"comments_posted"`
	ClipsFavorited    int        `json:"clips_favorited" db:"clips_favorited"`
	SearchesPerformed int        `json:"searches_performed" db:"searches_performed"`
	DaysActive        int        `json:"days_active" db:"days_active"`
	TotalKarmaEarned  int        `json:"total_karma_earned" db:"total_karma_earned"`
	LastActiveAt      *time.Time `json:"last_active_at,omitempty" db:"last_active_at"`
	UpdatedAt         time.Time  `json:"updated_at" db:"updated_at"`
}

// PlatformAnalytics represents global platform statistics
type PlatformAnalytics struct {
	ID                 uuid.UUID `json:"id" db:"id"`
	Date               time.Time `json:"date" db:"date"`
	TotalUsers         int64     `json:"total_users" db:"total_users"`
	ActiveUsersDaily   int       `json:"active_users_daily" db:"active_users_daily"`
	ActiveUsersWeekly  int       `json:"active_users_weekly" db:"active_users_weekly"`
	ActiveUsersMonthly int       `json:"active_users_monthly" db:"active_users_monthly"`
	NewUsersToday      int       `json:"new_users_today" db:"new_users_today"`
	TotalClips         int64     `json:"total_clips" db:"total_clips"`
	NewClipsToday      int       `json:"new_clips_today" db:"new_clips_today"`
	TotalVotes         int64     `json:"total_votes" db:"total_votes"`
	VotesToday         int       `json:"votes_today" db:"votes_today"`
	TotalComments      int64     `json:"total_comments" db:"total_comments"`
	CommentsToday      int       `json:"comments_today" db:"comments_today"`
	TotalViews         int64     `json:"total_views" db:"total_views"`
	ViewsToday         int64     `json:"views_today" db:"views_today"`
	AvgSessionDuration *float64  `json:"avg_session_duration,omitempty" db:"avg_session_duration"`
	Metadata           *string   `json:"metadata,omitempty" db:"metadata"` // JSON string
	CreatedAt          time.Time `json:"created_at" db:"created_at"`
}

// CreatorAnalyticsOverview represents summary metrics for creator dashboard
type CreatorAnalyticsOverview struct {
	TotalClips        int     `json:"total_clips"`
	TotalViews        int64   `json:"total_views"`
	TotalUpvotes      int64   `json:"total_upvotes"`
	TotalComments     int64   `json:"total_comments"`
	AvgEngagementRate float64 `json:"avg_engagement_rate"`
	FollowerCount     int     `json:"follower_count"`
}

// CreatorTopClip represents a top-performing clip for creator analytics
type CreatorTopClip struct {
	Clip
	Views          int64   `json:"views"`
	EngagementRate float64 `json:"engagement_rate"`
}

// TrendDataPoint represents a data point in a time series
type TrendDataPoint struct {
	Date  time.Time `json:"date"`
	Value int64     `json:"value"`
}

// PlatformOverviewMetrics represents KPIs for admin dashboard
type PlatformOverviewMetrics struct {
	TotalUsers         int64   `json:"total_users"`
	ActiveUsersDaily   int     `json:"active_users_daily"`
	ActiveUsersMonthly int     `json:"active_users_monthly"`
	TotalClips         int64   `json:"total_clips"`
	ClipsAddedToday    int     `json:"clips_added_today"`
	TotalVotes         int64   `json:"total_votes"`
	TotalComments      int64   `json:"total_comments"`
	AvgSessionDuration float64 `json:"avg_session_duration"`
}

// ContentMetrics represents content-related metrics for admin dashboard
type ContentMetrics struct {
	MostPopularGames    []GameMetric    `json:"most_popular_games"`
	MostPopularCreators []CreatorMetric `json:"most_popular_creators"`
	TrendingTags        []TagMetric     `json:"trending_tags"`
	AvgClipVoteScore    float64         `json:"avg_clip_vote_score"`
}

// GameMetric represents game popularity metrics
type GameMetric struct {
	GameID    *string `json:"game_id"`
	GameName  string  `json:"game_name"`
	ClipCount int     `json:"clip_count"`
	ViewCount int64   `json:"view_count"`
}

// CreatorMetric represents creator popularity metrics
type CreatorMetric struct {
	CreatorID   *string `json:"creator_id"`
	CreatorName string  `json:"creator_name"`
	ClipCount   int     `json:"clip_count"`
	ViewCount   int64   `json:"view_count"`
	VoteScore   int64   `json:"vote_score"`
}

// TagMetric represents tag usage metrics
type TagMetric struct {
	TagID      uuid.UUID `json:"tag_id"`
	TagName    string    `json:"tag_name"`
	UsageCount int       `json:"usage_count"`
}

// UserEngagementScore represents a user's engagement score and its components
type UserEngagementScore struct {
	UserID       uuid.UUID                `json:"user_id" db:"user_id"`
	Score        int                      `json:"score" db:"score"` // 0-100
	Tier         string                   `json:"tier" db:"tier"`   // Inactive, Low, Moderate, High, Very High
	Components   UserEngagementComponents `json:"components"`
	CalculatedAt time.Time                `json:"calculated_at" db:"calculated_at"`
	UpdatedAt    time.Time                `json:"updated_at" db:"updated_at"`
}

// UserEngagementComponents represents the individual components of engagement score
type UserEngagementComponents struct {
	Posts          EngagementComponent `json:"posts"`
	Comments       EngagementComponent `json:"comments"`
	Votes          EngagementComponent `json:"votes"`
	LoginFrequency EngagementComponent `json:"login_frequency"`
	TimeSpent      EngagementComponent `json:"time_spent"`
}

// EngagementComponent represents a single component of the engagement score
type EngagementComponent struct {
	Score  int     `json:"score"`  // 0-100
	Count  int     `json:"count"`  // Raw count of activities
	Weight float64 `json:"weight"` // Weight in overall score (e.g., 0.20 for 20%)
}

// PlatformHealthMetrics represents platform-wide health indicators
type PlatformHealthMetrics struct {
	DAU              int            `json:"dau" db:"dau"`
	WAU              int            `json:"wau" db:"wau"`
	MAU              int            `json:"mau" db:"mau"`
	Stickiness       float64        `json:"stickiness" db:"stickiness"` // DAU/MAU ratio
	RetentionRates   RetentionRates `json:"retention"`
	ChurnRateMonthly float64        `json:"churn_rate_monthly" db:"churn_rate_monthly"`
	Trends           PlatformTrends `json:"trends"`
	CalculatedAt     time.Time      `json:"calculated_at" db:"calculated_at"`
}

// RetentionRates represents retention percentages for different periods
type RetentionRates struct {
	Day1  float64 `json:"day1" db:"day1_retention"`   // Day 1 retention rate
	Day7  float64 `json:"day7" db:"day7_retention"`   // Day 7 retention rate
	Day30 float64 `json:"day30" db:"day30_retention"` // Day 30 retention rate
}

// PlatformTrends represents week-over-week and month-over-month changes
type PlatformTrends struct {
	DAUChangeWoW float64 `json:"dau_change_wow" db:"dau_change_wow"` // Week-over-week % change
	MAUChangeMoM float64 `json:"mau_change_mom" db:"mau_change_mom"` // Month-over-month % change
}

// TrendingMetrics represents trending data with week-over-week changes
type TrendingMetrics struct {
	Metric             string              `json:"metric"`
	PeriodDays         int                 `json:"period_days"`
	Data               []TrendingDataPoint `json:"data"`
	WeekOverWeekChange float64             `json:"week_over_week_change"`
	Summary            TrendSummary        `json:"summary"`
}

// TrendingDataPoint represents a single data point in trending metrics with change calculation
type TrendingDataPoint struct {
	Date               time.Time `json:"date"`
	Value              int64     `json:"value"`
	ChangeFromPrevious float64   `json:"change_from_previous"`
}

// FromTrendDataPoint converts a TrendDataPoint to TrendingDataPoint
func (tdp *TrendingDataPoint) FromTrendDataPoint(t TrendDataPoint, prevValue int64) {
	tdp.Date = t.Date
	tdp.Value = t.Value
	if prevValue > 0 {
		tdp.ChangeFromPrevious = ((float64(t.Value) - float64(prevValue)) / float64(prevValue)) * 100
	}
}

// TrendSummary provides summary statistics for a trend
type TrendSummary struct {
	Min   int64  `json:"min"`
	Max   int64  `json:"max"`
	Avg   int64  `json:"avg"`
	Trend string `json:"trend"` // increasing, decreasing, stable
}

// ContentEngagementScore represents engagement metrics for a piece of content
type ContentEngagementScore struct {
	ClipID             uuid.UUID `json:"clip_id" db:"clip_id"`
	Score              int       `json:"score" db:"score"` // 0-100 composite score
	NormalizedViews    int       `json:"normalized_views" db:"normalized_views"`
	VoteRatio          float64   `json:"vote_ratio" db:"vote_ratio"`
	NormalizedComments int       `json:"normalized_comments" db:"normalized_comments"`
	NormalizedShares   int       `json:"normalized_shares" db:"normalized_shares"`
	FavoriteRate       float64   `json:"favorite_rate" db:"favorite_rate"`
	CalculatedAt       time.Time `json:"calculated_at" db:"calculated_at"`
}

// EngagementAlert represents an alert for engagement metrics
type EngagementAlert struct {
	ID             uuid.UUID              `json:"id" db:"id"`
	AlertType      string                 `json:"alert_type" db:"alert_type"` // dau_drop, churn_spike, etc.
	Severity       string                 `json:"severity" db:"severity"`     // P1, P2, P3
	Metric         string                 `json:"metric" db:"metric"`         // Which metric triggered the alert
	CurrentValue   float64                `json:"current_value" db:"current_value"`
	ThresholdValue float64                `json:"threshold_value" db:"threshold_value"`
	Message        string                 `json:"message" db:"message"`
	Metadata       map[string]interface{} `json:"metadata,omitempty" db:"metadata"`
	TriggeredAt    time.Time              `json:"triggered_at" db:"triggered_at"`
	AcknowledgedAt *time.Time             `json:"acknowledged_at,omitempty" db:"acknowledged_at"`
	AcknowledgedBy *uuid.UUID             `json:"acknowledged_by,omitempty" db:"acknowledged_by"`
	ResolvedAt     *time.Time             `json:"resolved_at,omitempty" db:"resolved_at"`
}

// CohortRetention represents retention data for a user cohort
type CohortRetention struct {
	CohortDate     time.Time `json:"cohort_date" db:"cohort_date"`         // Start date of cohort (e.g., signup month)
	CohortSize     int       `json:"cohort_size" db:"cohort_size"`         // Total users in cohort
	Day1Active     int       `json:"day1_active" db:"day1_active"`         // Users active after 1 day
	Day7Active     int       `json:"day7_active" db:"day7_active"`         // Users active after 7 days
	Day30Active    int       `json:"day30_active" db:"day30_active"`       // Users active after 30 days
	Day1Retention  float64   `json:"day1_retention" db:"day1_retention"`   // Percentage
	Day7Retention  float64   `json:"day7_retention" db:"day7_retention"`   // Percentage
	Day30Retention float64   `json:"day30_retention" db:"day30_retention"` // Percentage
	CalculatedAt   time.Time `json:"calculated_at" db:"calculated_at"`
}

// GeographyMetric represents audience distribution by country
type GeographyMetric struct {
	Country    string  `json:"country"`    // ISO country code (e.g., "US", "GB")
	ViewCount  int64   `json:"view_count"` // Number of views from this country
	Percentage float64 `json:"percentage"` // Percentage of total views
}

// DeviceMetric represents audience distribution by device type
type DeviceMetric struct {
	DeviceType string  `json:"device_type"` // "mobile", "desktop", "tablet", "unknown"
	ViewCount  int64   `json:"view_count"`  // Number of views from this device type
	Percentage float64 `json:"percentage"`  // Percentage of total views
}

// CreatorAudienceInsights represents audience insights for a creator
type CreatorAudienceInsights struct {
	TopCountries []GeographyMetric `json:"top_countries"` // Top countries by view count
	DeviceTypes  []DeviceMetric    `json:"device_types"`  // Distribution by device type
	TotalViews   int64             `json:"total_views"`   // Total views analyzed
}

// Subscription represents a user's subscription status
type Subscription struct {
	ID                   uuid.UUID  `json:"id" db:"id"`
	UserID               uuid.UUID  `json:"user_id" db:"user_id"`
	StripeCustomerID     string     `json:"stripe_customer_id" db:"stripe_customer_id"`
	StripeSubscriptionID *string    `json:"stripe_subscription_id,omitempty" db:"stripe_subscription_id"`
	StripePriceID        *string    `json:"stripe_price_id,omitempty" db:"stripe_price_id"`
	Status               string     `json:"status" db:"status"` // inactive, active, trialing, past_due, canceled, unpaid
	Tier                 string     `json:"tier" db:"tier"`     // free, pro
	CurrentPeriodStart   *time.Time `json:"current_period_start,omitempty" db:"current_period_start"`
	CurrentPeriodEnd     *time.Time `json:"current_period_end,omitempty" db:"current_period_end"`
	CancelAtPeriodEnd    bool       `json:"cancel_at_period_end" db:"cancel_at_period_end"`
	CanceledAt           *time.Time `json:"canceled_at,omitempty" db:"canceled_at"`
	TrialStart           *time.Time `json:"trial_start,omitempty" db:"trial_start"`
	TrialEnd             *time.Time `json:"trial_end,omitempty" db:"trial_end"`
	GracePeriodEnd       *time.Time `json:"grace_period_end,omitempty" db:"grace_period_end"`
	CreatedAt            time.Time  `json:"created_at" db:"created_at"`
	UpdatedAt            time.Time  `json:"updated_at" db:"updated_at"`
}

// SubscriptionEvent represents an event in subscription lifecycle for audit logging
type SubscriptionEvent struct {
	ID             uuid.UUID  `json:"id" db:"id"`
	SubscriptionID *uuid.UUID `json:"subscription_id,omitempty" db:"subscription_id"`
	EventType      string     `json:"event_type" db:"event_type"`
	StripeEventID  *string    `json:"stripe_event_id,omitempty" db:"stripe_event_id"`
	Payload        string     `json:"payload" db:"payload"` // JSONB stored as string
	CreatedAt      time.Time  `json:"created_at" db:"created_at"`
}

// WebhookRetryQueue represents a webhook event pending retry
type WebhookRetryQueue struct {
	ID            uuid.UUID  `json:"id" db:"id"`
	StripeEventID string     `json:"stripe_event_id" db:"stripe_event_id"`
	EventType     string     `json:"event_type" db:"event_type"`
	Payload       string     `json:"payload" db:"payload"` // JSONB stored as string
	RetryCount    int        `json:"retry_count" db:"retry_count"`
	MaxRetries    int        `json:"max_retries" db:"max_retries"`
	NextRetryAt   *time.Time `json:"next_retry_at,omitempty" db:"next_retry_at"`
	LastError     *string    `json:"last_error,omitempty" db:"last_error"`
	CreatedAt     time.Time  `json:"created_at" db:"created_at"`
	UpdatedAt     time.Time  `json:"updated_at" db:"updated_at"`
}

// WebhookDeadLetterQueue represents a permanently failed webhook event
type WebhookDeadLetterQueue struct {
	ID                uuid.UUID `json:"id" db:"id"`
	StripeEventID     string    `json:"stripe_event_id" db:"stripe_event_id"`
	EventType         string    `json:"event_type" db:"event_type"`
	Payload           string    `json:"payload" db:"payload"` // JSONB stored as string
	RetryCount        int       `json:"retry_count" db:"retry_count"`
	Error             string    `json:"error" db:"error"`
	OriginalTimestamp time.Time `json:"original_timestamp" db:"original_timestamp"`
	CreatedAt         time.Time `json:"created_at" db:"created_at"`
}

// UserWithSubscription represents a user with their subscription information
type UserWithSubscription struct {
	User
	Subscription *Subscription `json:"subscription,omitempty"`
}

// CreateCheckoutSessionRequest represents a request to create a Stripe checkout session
type CreateCheckoutSessionRequest struct {
	PriceID    string  `json:"price_id" binding:"required"`
	CouponCode *string `json:"coupon_code,omitempty"`
}

// ChangeSubscriptionPlanRequest represents a request to change subscription plan
type ChangeSubscriptionPlanRequest struct {
	PriceID string `json:"price_id" binding:"required"`
}

// CancelSubscriptionRequest represents a request to cancel a subscription
type CancelSubscriptionRequest struct {
	Immediate bool `json:"immediate"` // If true, cancel immediately. Otherwise, cancel at period end.
}

// CreateCheckoutSessionResponse represents the response with checkout session URL
type CreateCheckoutSessionResponse struct {
	SessionID  string `json:"session_id"`
	SessionURL string `json:"session_url"`
}

// CreatePortalSessionResponse represents the response with portal session URL
type CreatePortalSessionResponse struct {
	PortalURL string `json:"portal_url"`
}

// PaymentFailure represents a failed payment attempt for a subscription
type PaymentFailure struct {
	ID                    uuid.UUID  `json:"id" db:"id"`
	SubscriptionID        uuid.UUID  `json:"subscription_id" db:"subscription_id"`
	StripeInvoiceID       string     `json:"stripe_invoice_id" db:"stripe_invoice_id"`
	StripePaymentIntentID *string    `json:"stripe_payment_intent_id,omitempty" db:"stripe_payment_intent_id"`
	AmountDue             int64      `json:"amount_due" db:"amount_due"` // Amount in cents
	Currency              string     `json:"currency" db:"currency"`
	AttemptCount          int        `json:"attempt_count" db:"attempt_count"`
	FailureReason         *string    `json:"failure_reason,omitempty" db:"failure_reason"`
	NextRetryAt           *time.Time `json:"next_retry_at,omitempty" db:"next_retry_at"`
	Resolved              bool       `json:"resolved" db:"resolved"`
	ResolvedAt            *time.Time `json:"resolved_at,omitempty" db:"resolved_at"`
	CreatedAt             time.Time  `json:"created_at" db:"created_at"`
	UpdatedAt             time.Time  `json:"updated_at" db:"updated_at"`
}

// DunningAttempt represents a communication attempt to a user about failed payment
type DunningAttempt struct {
	ID               uuid.UUID  `json:"id" db:"id"`
	PaymentFailureID uuid.UUID  `json:"payment_failure_id" db:"payment_failure_id"`
	UserID           uuid.UUID  `json:"user_id" db:"user_id"`
	AttemptNumber    int        `json:"attempt_number" db:"attempt_number"`
	NotificationType string     `json:"notification_type" db:"notification_type"` // payment_failed, payment_retry, grace_period_warning, subscription_downgraded
	EmailSent        bool       `json:"email_sent" db:"email_sent"`
	EmailSentAt      *time.Time `json:"email_sent_at,omitempty" db:"email_sent_at"`
	CreatedAt        time.Time  `json:"created_at" db:"created_at"`
}

// ContactMessage represents a contact form submission
type ContactMessage struct {
	ID        uuid.UUID  `json:"id" db:"id"`
	UserID    *uuid.UUID `json:"user_id,omitempty" db:"user_id"`       // Nullable for logged-out users
	Email     string     `json:"email" db:"email"`                     // Required for contact
	Category  string     `json:"category" db:"category"`               // abuse, account, billing, feedback
	Subject   string     `json:"subject" db:"subject"`                 // Brief subject line
	Message   string     `json:"message" db:"message"`                 // Full message content
	Status    string     `json:"status" db:"status"`                   // pending, reviewed, resolved
	IPAddress *string    `json:"ip_address,omitempty" db:"ip_address"` // For abuse prevention
	UserAgent *string    `json:"user_agent,omitempty" db:"user_agent"` // For abuse prevention
	CreatedAt time.Time  `json:"created_at" db:"created_at"`
	UpdatedAt time.Time  `json:"updated_at" db:"updated_at"`
}

// CreateContactMessageRequest represents the request to submit a contact form
type CreateContactMessageRequest struct {
	Email    string `json:"email" binding:"required,email,max=255"`
	Category string `json:"category" binding:"required,oneof=abuse account billing feedback"`
	Subject  string `json:"subject" binding:"required,min=3,max=200"`
	Message  string `json:"message" binding:"required,min=10,max=5000"`
}

// EmailNotificationLog represents an audit log for sent emails
type EmailNotificationLog struct {
	ID                uuid.UUID  `json:"id" db:"id"`
	UserID            uuid.UUID  `json:"user_id" db:"user_id"`
	NotificationID    *uuid.UUID `json:"notification_id,omitempty" db:"notification_id"`
	NotificationType  string     `json:"notification_type" db:"notification_type"`
	RecipientEmail    string     `json:"recipient_email" db:"recipient_email"`
	Subject           string     `json:"subject" db:"subject"`
	Status            string     `json:"status" db:"status"` // pending, sent, failed, bounced
	ProviderMessageID *string    `json:"provider_message_id,omitempty" db:"provider_message_id"`
	ErrorMessage      *string    `json:"error_message,omitempty" db:"error_message"`
	SentAt            *time.Time `json:"sent_at,omitempty" db:"sent_at"`
	CreatedAt         time.Time  `json:"created_at" db:"created_at"`
	UpdatedAt         time.Time  `json:"updated_at" db:"updated_at"`
}

// EmailUnsubscribeToken represents an unsubscribe token
type EmailUnsubscribeToken struct {
	ID               uuid.UUID  `json:"id" db:"id"`
	UserID           uuid.UUID  `json:"user_id" db:"user_id"`
	Token            string     `json:"token" db:"token"`
	NotificationType *string    `json:"notification_type,omitempty" db:"notification_type"` // null means all
	CreatedAt        time.Time  `json:"created_at" db:"created_at"`
	ExpiresAt        time.Time  `json:"expires_at" db:"expires_at"`
	UsedAt           *time.Time `json:"used_at,omitempty" db:"used_at"`
}

// EmailRateLimit represents rate limiting for email notifications
type EmailRateLimit struct {
	ID          uuid.UUID `json:"id" db:"id"`
	UserID      uuid.UUID `json:"user_id" db:"user_id"`
	WindowStart time.Time `json:"window_start" db:"window_start"`
	EmailCount  int       `json:"email_count" db:"email_count"`
	CreatedAt   time.Time `json:"created_at" db:"created_at"`
	UpdatedAt   time.Time `json:"updated_at" db:"updated_at"`
}

// Email notification log status constants
const (
	EmailStatusPending = "pending"
	EmailStatusSent    = "sent"
	EmailStatusFailed  = "failed"
	EmailStatusBounced = "bounced"
)

// RegisterDeviceTokenRequest represents the request to register a device token
type RegisterDeviceTokenRequest struct {
	DeviceToken    string `json:"device_token" binding:"required"`
	DevicePlatform string `json:"device_platform" binding:"required,oneof=ios android web"`
}

// UnregisterDeviceTokenRequest represents the request to unregister a device token
type UnregisterDeviceTokenRequest struct {
	DeviceToken string `json:"device_token" binding:"required"`
}

// RevenueMetrics represents subscription revenue metrics for admin dashboard
type RevenueMetrics struct {
	MRR                  float64                  `json:"mrr"`                    // Monthly Recurring Revenue in cents
	Churn                float64                  `json:"churn"`                  // Churn rate as percentage
	ARPU                 float64                  `json:"arpu"`                   // Average Revenue Per User in cents
	ActiveSubscribers    int                      `json:"active_subscribers"`     // Total active subscribers
	TotalRevenue         float64                  `json:"total_revenue"`          // Total revenue to date in cents
	PlanDistribution     []PlanDistributionMetric `json:"plan_distribution"`      // Distribution by plan
	CohortRetention      []CohortRetentionMetric  `json:"cohort_retention"`       // Cohort retention data
	ChurnedSubscribers   int                      `json:"churned_subscribers"`    // Subscribers churned this month
	NewSubscribers       int                      `json:"new_subscribers"`        // New subscribers this month
	TrialConversionRate  float64                  `json:"trial_conversion_rate"`  // Trial to paid conversion rate
	GracePeriodRecovery  float64                  `json:"grace_period_recovery"`  // Grace period recovery rate
	AverageLifetimeValue float64                  `json:"average_lifetime_value"` // Average customer LTV in cents
	RevenueByMonth       []RevenueByMonthMetric   `json:"revenue_by_month"`       // Revenue trend by month
	SubscriberGrowth     []SubscriberGrowthMetric `json:"subscriber_growth"`      // Subscriber growth trend
	UpdatedAt            time.Time                `json:"updated_at"`
}

// PlanDistributionMetric represents distribution of subscribers by plan
type PlanDistributionMetric struct {
	PlanID       string  `json:"plan_id"`
	PlanName     string  `json:"plan_name"`
	Subscribers  int     `json:"subscribers"`
	Percentage   float64 `json:"percentage"`
	MonthlyValue float64 `json:"monthly_value"` // in cents
}

// CohortRetentionMetric represents retention data for a specific cohort
type CohortRetentionMetric struct {
	CohortMonth    string    `json:"cohort_month"`    // YYYY-MM format
	InitialSize    int       `json:"initial_size"`    // Number of subscribers in cohort
	RetentionRates []float64 `json:"retention_rates"` // Retention % for each month after signup
}

// RevenueByMonthMetric represents revenue data for a specific month
type RevenueByMonthMetric struct {
	Month   string  `json:"month"`   // YYYY-MM format
	Revenue float64 `json:"revenue"` // Revenue in cents
	MRR     float64 `json:"mrr"`     // MRR at end of month in cents
}

// SubscriberGrowthMetric represents subscriber growth data for a specific month
type SubscriberGrowthMetric struct {
	Month     string `json:"month"`      // YYYY-MM format
	Total     int    `json:"total"`      // Total subscribers at end of month
	New       int    `json:"new"`        // New subscribers that month
	Churned   int    `json:"churned"`    // Churned subscribers that month
	NetChange int    `json:"net_change"` // Net subscriber change
}

// ExportRequest represents a creator's data export request
type ExportRequest struct {
	ID            uuid.UUID  `json:"id" db:"id"`
	UserID        uuid.UUID  `json:"user_id" db:"user_id"`
	CreatorName   string     `json:"creator_name" db:"creator_name"`
	Format        string     `json:"format" db:"format"` // csv, json
	Status        string     `json:"status" db:"status"` // pending, processing, completed, failed, expired
	FilePath      *string    `json:"file_path,omitempty" db:"file_path"`
	FileSizeBytes *int64     `json:"file_size_bytes,omitempty" db:"file_size_bytes"`
	ErrorMessage  *string    `json:"error_message,omitempty" db:"error_message"`
	ExpiresAt     *time.Time `json:"expires_at,omitempty" db:"expires_at"`
	EmailSent     bool       `json:"email_sent" db:"email_sent"`
	CreatedAt     time.Time  `json:"created_at" db:"created_at"`
	UpdatedAt     time.Time  `json:"updated_at" db:"updated_at"`
	CompletedAt   *time.Time `json:"completed_at,omitempty" db:"completed_at"`
}

// Export request status constants
const (
	ExportStatusPending    = "pending"
	ExportStatusProcessing = "processing"
	ExportStatusCompleted  = "completed"
	ExportStatusFailed     = "failed"
	ExportStatusExpired    = "expired"
)

// Export format constants
const (
	ExportFormatCSV  = "csv"
	ExportFormatJSON = "json"
)

// CreateExportRequest represents the request to create a data export
type CreateExportRequest struct {
	Format string `json:"format" binding:"required,oneof=csv json"`
}

// ExportRequestResponse represents the response for an export request
type ExportRequestResponse struct {
	ExportRequest
	DownloadURL *string `json:"download_url,omitempty"`
}

// Ad represents an advertisement campaign
type Ad struct {
	ID                uuid.UUID              `json:"id" db:"id"`
	Name              string                 `json:"name" db:"name"`
	AdvertiserName    string                 `json:"advertiser_name" db:"advertiser_name"`
	AdType            string                 `json:"ad_type" db:"ad_type"` // banner, video, native
	ContentURL        string                 `json:"content_url" db:"content_url"`
	ClickURL          *string                `json:"click_url,omitempty" db:"click_url"`
	AltText           *string                `json:"alt_text,omitempty" db:"alt_text"`
	Width             *int                   `json:"width,omitempty" db:"width"`
	Height            *int                   `json:"height,omitempty" db:"height"`
	Priority          int                    `json:"priority" db:"priority"`
	Weight            int                    `json:"weight" db:"weight"`
	DailyBudgetCents  *int64                 `json:"daily_budget_cents,omitempty" db:"daily_budget_cents"`
	TotalBudgetCents  *int64                 `json:"total_budget_cents,omitempty" db:"total_budget_cents"`
	SpentTodayCents   int64                  `json:"spent_today_cents" db:"spent_today_cents"`
	SpentTotalCents   int64                  `json:"spent_total_cents" db:"spent_total_cents"`
	CPMCents          int                    `json:"cpm_cents" db:"cpm_cents"` // Cost per 1000 impressions
	IsActive          bool                   `json:"is_active" db:"is_active"`
	StartDate         *time.Time             `json:"start_date,omitempty" db:"start_date"`
	EndDate           *time.Time             `json:"end_date,omitempty" db:"end_date"`
	TargetingCriteria map[string]interface{} `json:"targeting_criteria,omitempty" db:"targeting_criteria"`
	// New fields for slots and experiments
	SlotID            *string    `json:"slot_id,omitempty" db:"slot_id"`
	ExperimentID      *uuid.UUID `json:"experiment_id,omitempty" db:"experiment_id"`
	ExperimentVariant *string    `json:"experiment_variant,omitempty" db:"experiment_variant"`
	CreatedAt         time.Time  `json:"created_at" db:"created_at"`
	UpdatedAt         time.Time  `json:"updated_at" db:"updated_at"`
}

// AdImpression represents a tracked ad impression
type AdImpression struct {
	ID                uuid.UUID  `json:"id" db:"id"`
	AdID              uuid.UUID  `json:"ad_id" db:"ad_id"`
	UserID            *uuid.UUID `json:"user_id,omitempty" db:"user_id"`
	SessionID         *string    `json:"session_id,omitempty" db:"session_id"`
	Platform          string     `json:"platform" db:"platform"` // web, ios, android
	IPAddress         *string    `json:"ip_address,omitempty" db:"ip_address"`
	UserAgent         *string    `json:"user_agent,omitempty" db:"user_agent"`
	PageURL           *string    `json:"page_url,omitempty" db:"page_url"`
	ViewabilityTimeMs int        `json:"viewability_time_ms" db:"viewability_time_ms"`
	IsViewable        bool       `json:"is_viewable" db:"is_viewable"`
	IsClicked         bool       `json:"is_clicked" db:"is_clicked"`
	ClickedAt         *time.Time `json:"clicked_at,omitempty" db:"clicked_at"`
	CostCents         int        `json:"cost_cents" db:"cost_cents"`
	// New fields for enhanced tracking
	SlotID            *string    `json:"slot_id,omitempty" db:"slot_id"`
	Country           *string    `json:"country,omitempty" db:"country"`
	DeviceType        *string    `json:"device_type,omitempty" db:"device_type"` // desktop, mobile, tablet
	ExperimentID      *uuid.UUID `json:"experiment_id,omitempty" db:"experiment_id"`
	ExperimentVariant *string    `json:"experiment_variant,omitempty" db:"experiment_variant"`
	CreatedAt         time.Time  `json:"created_at" db:"created_at"`
}

// AdFrequencyCap represents per-user/session impression tracking for frequency capping
type AdFrequencyCap struct {
	ID              uuid.UUID  `json:"id" db:"id"`
	AdID            uuid.UUID  `json:"ad_id" db:"ad_id"`
	UserID          *uuid.UUID `json:"user_id,omitempty" db:"user_id"`
	SessionID       *string    `json:"session_id,omitempty" db:"session_id"`
	ImpressionCount int        `json:"impression_count" db:"impression_count"`
	WindowStart     time.Time  `json:"window_start" db:"window_start"`
	WindowType      string     `json:"window_type" db:"window_type"` // hourly, daily, weekly, lifetime
	CreatedAt       time.Time  `json:"created_at" db:"created_at"`
	UpdatedAt       time.Time  `json:"updated_at" db:"updated_at"`
}

// AdFrequencyLimit represents configurable frequency limits per ad
type AdFrequencyLimit struct {
	ID             uuid.UUID `json:"id" db:"id"`
	AdID           uuid.UUID `json:"ad_id" db:"ad_id"`
	WindowType     string    `json:"window_type" db:"window_type"` // hourly, daily, weekly, lifetime
	MaxImpressions int       `json:"max_impressions" db:"max_impressions"`
	CreatedAt      time.Time `json:"created_at" db:"created_at"`
}

// AdSelectionRequest represents a request to select an ad
type AdSelectionRequest struct {
	Platform  string  `json:"platform" form:"platform" binding:"required,oneof=web ios android"`
	PageURL   *string `json:"page_url,omitempty" form:"page_url"`
	AdType    *string `json:"ad_type,omitempty" form:"ad_type"` // Filter by ad type
	Width     *int    `json:"width,omitempty" form:"width"`     // Filter by dimensions
	Height    *int    `json:"height,omitempty" form:"height"`
	SessionID *string `json:"session_id,omitempty" form:"session_id"` // For anonymous users
	GameID    *string `json:"game_id,omitempty" form:"game_id"`       // For targeting
	Language  *string `json:"language,omitempty" form:"language"`
	// Enhanced targeting fields
	SlotID     *string  `json:"slot_id,omitempty" form:"slot_id"`         // Ad placement slot identifier
	Country    *string  `json:"country,omitempty" form:"country"`         // ISO 3166-1 alpha-2 country code
	DeviceType *string  `json:"device_type,omitempty" form:"device_type"` // desktop, mobile, tablet
	Interests  []string `json:"interests,omitempty" form:"interests"`     // User interest categories
	// Privacy/consent fields
	Personalized *bool `json:"personalized,omitempty" form:"personalized"` // Whether user consented to personalized ads
}

// AdSelectionResponse represents a selected ad for display
type AdSelectionResponse struct {
	Ad           *Ad    `json:"ad,omitempty"`
	ImpressionID string `json:"impression_id,omitempty"` // UUID for tracking
	TrackingURL  string `json:"tracking_url,omitempty"`  // URL to call for viewability
}

// AdTrackingRequest represents a tracking update for an impression
type AdTrackingRequest struct {
	ImpressionID      string `json:"impression_id" binding:"required"`
	ViewabilityTimeMs int    `json:"viewability_time_ms"`
	IsViewable        bool   `json:"is_viewable"`
	IsClicked         bool   `json:"is_clicked"`
}

// AdFrequencyCapWindow represents the time window types for frequency capping
const (
	FrequencyWindowHourly   = "hourly"
	FrequencyWindowDaily    = "daily"
	FrequencyWindowWeekly   = "weekly"
	FrequencyWindowLifetime = "lifetime"
)

// ViewabilityThresholdMs is the minimum time (in ms) an ad must be viewable to count
// IAB standard: 50% of pixels visible for 1000ms (1 second)
const ViewabilityThresholdMs = 1000

// AdExperiment represents an A/B experiment for comparing ad variants
type AdExperiment struct {
	ID             uuid.UUID  `json:"id" db:"id"`
	Name           string     `json:"name" db:"name"`
	Description    *string    `json:"description,omitempty" db:"description"`
	Status         string     `json:"status" db:"status"` // draft, running, paused, completed
	StartDate      *time.Time `json:"start_date,omitempty" db:"start_date"`
	EndDate        *time.Time `json:"end_date,omitempty" db:"end_date"`
	TrafficPercent int        `json:"traffic_percent" db:"traffic_percent"` // 0-100
	WinningVariant *string    `json:"winning_variant,omitempty" db:"winning_variant"`
	CreatedAt      time.Time  `json:"created_at" db:"created_at"`
	UpdatedAt      time.Time  `json:"updated_at" db:"updated_at"`
}

// AdTargetingRule represents a structured targeting rule for an ad
type AdTargetingRule struct {
	ID        uuid.UUID `json:"id" db:"id"`
	AdID      uuid.UUID `json:"ad_id" db:"ad_id"`
	RuleType  string    `json:"rule_type" db:"rule_type"` // country, device, interest, platform, language, game
	Operator  string    `json:"operator" db:"operator"`   // include, exclude
	Values    []string  `json:"values" db:"values"`       // Array of values to match
	CreatedAt time.Time `json:"created_at" db:"created_at"`
}

// TargetingRuleType constants
const (
	TargetingRuleTypeCountry  = "country"
	TargetingRuleTypeDevice   = "device"
	TargetingRuleTypeInterest = "interest"
	TargetingRuleTypePlatform = "platform"
	TargetingRuleTypeLanguage = "language"
	TargetingRuleTypeGame     = "game"
)

// TargetingRuleOperator constants
const (
	TargetingOperatorInclude = "include"
	TargetingOperatorExclude = "exclude"
)

// ExperimentStatus constants
const (
	ExperimentStatusDraft     = "draft"
	ExperimentStatusRunning   = "running"
	ExperimentStatusPaused    = "paused"
	ExperimentStatusCompleted = "completed"
)

// AdCampaignAnalytics represents aggregated campaign analytics by date and slot
type AdCampaignAnalytics struct {
	ID                  uuid.UUID `json:"id" db:"id"`
	AdID                uuid.UUID `json:"ad_id" db:"ad_id"`
	Date                time.Time `json:"date" db:"date"`
	SlotID              *string   `json:"slot_id,omitempty" db:"slot_id"`
	Impressions         int       `json:"impressions" db:"impressions"`
	ViewableImpressions int       `json:"viewable_impressions" db:"viewable_impressions"`
	Clicks              int       `json:"clicks" db:"clicks"`
	SpendCents          int64     `json:"spend_cents" db:"spend_cents"`
	UniqueUsers         int       `json:"unique_users" db:"unique_users"`
	CreatedAt           time.Time `json:"created_at" db:"created_at"`
	UpdatedAt           time.Time `json:"updated_at" db:"updated_at"`
}

// AdExperimentAnalytics represents aggregated experiment analytics by variant and date
type AdExperimentAnalytics struct {
	ID                  uuid.UUID `json:"id" db:"id"`
	ExperimentID        uuid.UUID `json:"experiment_id" db:"experiment_id"`
	Variant             string    `json:"variant" db:"variant"`
	Date                time.Time `json:"date" db:"date"`
	Impressions         int       `json:"impressions" db:"impressions"`
	ViewableImpressions int       `json:"viewable_impressions" db:"viewable_impressions"`
	Clicks              int       `json:"clicks" db:"clicks"`
	Conversions         int       `json:"conversions" db:"conversions"`
	CreatedAt           time.Time `json:"created_at" db:"created_at"`
}

// AdCTRReport represents a CTR report for campaigns/slots
type AdCTRReport struct {
	AdID                uuid.UUID `json:"ad_id" db:"ad_id"`
	AdName              string    `json:"ad_name" db:"ad_name"`
	SlotID              *string   `json:"slot_id,omitempty" db:"slot_id"`
	Impressions         int64     `json:"impressions" db:"impressions"`
	ViewableImpressions int64     `json:"viewable_impressions" db:"viewable_impressions"`
	Clicks              int64     `json:"clicks" db:"clicks"`
	CTR                 float64   `json:"ctr"`              // Click-through rate (clicks / viewable impressions)
	ViewabilityRate     float64   `json:"viewability_rate"` // viewable impressions / total impressions
	SpendCents          int64     `json:"spend_cents" db:"spend_cents"`
}

// AdSlotReport represents CTR metrics grouped by slot
type AdSlotReport struct {
	SlotID              string  `json:"slot_id" db:"slot_id"`
	Impressions         int64   `json:"impressions" db:"impressions"`
	ViewableImpressions int64   `json:"viewable_impressions" db:"viewable_impressions"`
	Clicks              int64   `json:"clicks" db:"clicks"`
	CTR                 float64 `json:"ctr"`
	ViewabilityRate     float64 `json:"viewability_rate"`
	SpendCents          int64   `json:"spend_cents" db:"spend_cents"`
	UniqueAds           int     `json:"unique_ads" db:"unique_ads"`
}

// AdExperimentReport represents experiment results with variant comparison
type AdExperimentReport struct {
	ExperimentID   uuid.UUID                   `json:"experiment_id" db:"experiment_id"`
	ExperimentName string                      `json:"experiment_name" db:"experiment_name"`
	Status         string                      `json:"status" db:"status"`
	Variants       []AdExperimentVariantReport `json:"variants"`
}

// AdExperimentVariantReport represents metrics for a single experiment variant
type AdExperimentVariantReport struct {
	Variant             string  `json:"variant" db:"variant"`
	Impressions         int64   `json:"impressions" db:"impressions"`
	ViewableImpressions int64   `json:"viewable_impressions" db:"viewable_impressions"`
	Clicks              int64   `json:"clicks" db:"clicks"`
	CTR                 float64 `json:"ctr"`
	Conversions         int64   `json:"conversions" db:"conversions"`
	ConversionRate      float64 `json:"conversion_rate"` // conversions / clicks
}

// AdSelectionContext represents context information for ad targeting
type AdSelectionContext struct {
	Country    *string  `json:"country,omitempty" form:"country"`
	DeviceType *string  `json:"device_type,omitempty" form:"device_type"` // desktop, mobile, tablet
	Interests  []string `json:"interests,omitempty" form:"interests"`
	SlotID     *string  `json:"slot_id,omitempty" form:"slot_id"`
}

// UpdateClipMetadataRequest represents a request to update clip metadata (title, tags)
type UpdateClipMetadataRequest struct {
	Title *string  `json:"title,omitempty" binding:"omitempty,min=1,max=255"`
	Tags  []string `json:"tags,omitempty" binding:"omitempty,max=10,dive,min=1,max=50"`
}

// UpdateClipVisibilityRequest represents a request to update clip visibility
type UpdateClipVisibilityRequest struct {
	IsHidden bool `json:"is_hidden"`
}

// WebhookSubscription represents a webhook subscription for third-party integrations
type WebhookSubscription struct {
	ID             uuid.UUID  `json:"id" db:"id"`
	UserID         uuid.UUID  `json:"user_id" db:"user_id"`
	URL            string     `json:"url" db:"url"`
	Secret         string     `json:"-" db:"secret"` // Never expose in JSON responses
	Events         []string   `json:"events" db:"events"`
	IsActive       bool       `json:"is_active" db:"is_active"`
	Description    *string    `json:"description,omitempty" db:"description"`
	CreatedAt      time.Time  `json:"created_at" db:"created_at"`
	UpdatedAt      time.Time  `json:"updated_at" db:"updated_at"`
	LastDeliveryAt *time.Time `json:"last_delivery_at,omitempty" db:"last_delivery_at"`
}

// WebhookDelivery represents a webhook delivery attempt
type WebhookDelivery struct {
	ID             uuid.UUID  `json:"id" db:"id"`
	SubscriptionID uuid.UUID  `json:"subscription_id" db:"subscription_id"`
	EventType      string     `json:"event_type" db:"event_type"`
	EventID        uuid.UUID  `json:"event_id" db:"event_id"`
	Payload        string     `json:"payload" db:"payload"`
	Status         string     `json:"status" db:"status"`
	HTTPStatusCode *int       `json:"http_status_code,omitempty" db:"http_status_code"`
	ResponseBody   *string    `json:"response_body,omitempty" db:"response_body"`
	ErrorMessage   *string    `json:"error_message,omitempty" db:"error_message"`
	AttemptCount   int        `json:"attempt_count" db:"attempt_count"`
	MaxAttempts    int        `json:"max_attempts" db:"max_attempts"`
	NextAttemptAt  *time.Time `json:"next_attempt_at,omitempty" db:"next_attempt_at"`
	DeliveredAt    *time.Time `json:"delivered_at,omitempty" db:"delivered_at"`
	CreatedAt      time.Time  `json:"created_at" db:"created_at"`
	UpdatedAt      time.Time  `json:"updated_at" db:"updated_at"`
}

// WebhookEventPayload represents the payload sent to webhook endpoints
type WebhookEventPayload struct {
	Event     string                 `json:"event"`
	Timestamp time.Time              `json:"timestamp"`
	Data      map[string]interface{} `json:"data"`
}

// CreateWebhookSubscriptionRequest represents a request to create a webhook subscription
type CreateWebhookSubscriptionRequest struct {
	URL         string   `json:"url" binding:"required,url,max=2048"`
	Events      []string `json:"events" binding:"required,min=1,max=10"`
	Description *string  `json:"description,omitempty" binding:"omitempty,max=500"`
}

// UpdateWebhookSubscriptionRequest represents a request to update a webhook subscription
type UpdateWebhookSubscriptionRequest struct {
	URL         *string  `json:"url,omitempty" binding:"omitempty,url,max=2048"`
	Events      []string `json:"events,omitempty" binding:"omitempty,min=1,max=10"`
	IsActive    *bool    `json:"is_active,omitempty"`
	Description *string  `json:"description,omitempty" binding:"omitempty,max=500"`
}

// WebhookEvent constants for supported webhook events
const (
	WebhookEventClipSubmitted = "clip.submitted"
	WebhookEventClipApproved  = "clip.approved"
	WebhookEventClipRejected  = "clip.rejected"
)

// GetSupportedWebhookEvents returns the list of supported webhook events
func GetSupportedWebhookEvents() []string {
	return []string{
		WebhookEventClipSubmitted,
		WebhookEventClipApproved,
		WebhookEventClipRejected,
	}
}

// OutboundWebhookDeadLetterQueue represents a permanently failed outbound webhook delivery
type OutboundWebhookDeadLetterQueue struct {
	ID                uuid.UUID  `json:"id" db:"id"`
	SubscriptionID    uuid.UUID  `json:"subscription_id" db:"subscription_id"`
	DeliveryID        uuid.UUID  `json:"delivery_id" db:"delivery_id"`
	EventType         string     `json:"event_type" db:"event_type"`
	EventID           uuid.UUID  `json:"event_id" db:"event_id"`
	Payload           string     `json:"payload" db:"payload"`
	ErrorMessage      string     `json:"error_message" db:"error_message"`
	HTTPStatusCode    *int       `json:"http_status_code,omitempty" db:"http_status_code"`
	ResponseBody      *string    `json:"response_body,omitempty" db:"response_body"`
	AttemptCount      int        `json:"attempt_count" db:"attempt_count"`
	OriginalCreatedAt time.Time  `json:"original_created_at" db:"original_created_at"`
	MovedToDLQAt      time.Time  `json:"moved_to_dlq_at" db:"moved_to_dlq_at"`
	ReplayedAt        *time.Time `json:"replayed_at,omitempty" db:"replayed_at"`
	ReplaySuccessful  *bool      `json:"replay_successful,omitempty" db:"replay_successful"`
	CreatedAt         time.Time  `json:"created_at" db:"created_at"`
}

// TrustScoreHistory represents a trust score change audit log entry
type TrustScoreHistory struct {
	ID              uuid.UUID              `json:"id" db:"id"`
	UserID          uuid.UUID              `json:"user_id" db:"user_id"`
	OldScore        int                    `json:"old_score" db:"old_score"`
	NewScore        int                    `json:"new_score" db:"new_score"`
	ChangeReason    string                 `json:"change_reason" db:"change_reason"`
	ComponentScores map[string]interface{} `json:"component_scores,omitempty" db:"component_scores"`
	ChangedBy       *uuid.UUID             `json:"changed_by,omitempty" db:"changed_by"`
	Notes           *string                `json:"notes,omitempty" db:"notes"`
	CreatedAt       time.Time              `json:"created_at" db:"created_at"`
}

// TrustScoreHistoryWithUser includes user and admin information
type TrustScoreHistoryWithUser struct {
	TrustScoreHistory
	User      *User `json:"user,omitempty"`
	ChangedBy *User `json:"changed_by_user,omitempty"`
}

// TrustScoreBreakdown represents the detailed breakdown of a trust score calculation
type TrustScoreBreakdown struct {
	TotalScore       int     `json:"total_score"`
	AccountAgeScore  int     `json:"account_age_score"`
	KarmaScore       int     `json:"karma_score"`
	ReportAccuracy   int     `json:"report_accuracy_score"`
	ActivityScore    int     `json:"activity_score"`
	MaxScore         int     `json:"max_score"`
	AccountAgeDays   int     `json:"account_age_days"`
	KarmaPoints      int     `json:"karma_points"`
	CorrectReports   int     `json:"correct_reports"`
	IncorrectReports int     `json:"incorrect_reports"`
	TotalComments    int     `json:"total_comments"`
	TotalVotes       int     `json:"total_votes"`
	DaysActive       int     `json:"days_active"`
	IsBanned         bool    `json:"is_banned"`
	BanPenalty       float64 `json:"ban_penalty,omitempty"`
}

// TrustScoreChangeReason constants for common change reasons
const (
	TrustScoreReasonScheduledRecalc    = "scheduled_recalc"
	TrustScoreReasonSubmissionApproved = "submission_approved"
	TrustScoreReasonSubmissionRejected = "submission_rejected"
	TrustScoreReasonReportActioned     = "report_actioned"
	TrustScoreReasonManualAdjustment   = "manual_adjustment"
	TrustScoreReasonNewActivity        = "new_activity"
	TrustScoreReasonBanned             = "banned"
	TrustScoreReasonUnbanned           = "unbanned"
)

// ManualTrustScoreAdjustment represents a request to manually adjust a user's trust score
type ManualTrustScoreAdjustment struct {
	NewScore int     `json:"new_score" binding:"required,min=0,max=100"`
	Reason   string  `json:"reason" binding:"required,min=3,max=100"`
	Notes    *string `json:"notes,omitempty" binding:"omitempty,max=1000"`
}

// CommentSuspensionHistory represents a comment suspension action record
type CommentSuspensionHistory struct {
	ID             uuid.UUID  `json:"id" db:"id"`
	UserID         uuid.UUID  `json:"user_id" db:"user_id"`
	SuspendedBy    uuid.UUID  `json:"suspended_by" db:"suspended_by"`
	SuspensionType string     `json:"suspension_type" db:"suspension_type"` // warning, temporary, permanent
	Reason         string     `json:"reason" db:"reason"`
	DurationHours  *int       `json:"duration_hours,omitempty" db:"duration_hours"`
	SuspendedAt    time.Time  `json:"suspended_at" db:"suspended_at"`
	ExpiresAt      *time.Time `json:"expires_at,omitempty" db:"expires_at"`
	IsActive       bool       `json:"is_active" db:"is_active"`
	LiftedAt       *time.Time `json:"lifted_at,omitempty" db:"lifted_at"`
	LiftedBy       *uuid.UUID `json:"lifted_by,omitempty" db:"lifted_by"`
	LiftReason     *string    `json:"lift_reason,omitempty" db:"lift_reason"`
	Metadata       *string    `json:"metadata,omitempty" db:"metadata"` // JSONB stored as string
}

// CommentSuspensionRequest represents a request to suspend comment privileges
type CommentSuspensionRequest struct {
	SuspensionType string `json:"suspension_type" binding:"required,oneof=warning temporary permanent"`
	Reason         string `json:"reason" binding:"required,min=10,max=1000"`
	DurationHours  *int   `json:"duration_hours,omitempty" binding:"omitempty,min=1,max=8760"` // max 1 year
}

// LiftSuspensionRequest represents a request to lift a comment suspension
type LiftSuspensionRequest struct {
	Reason string `json:"reason" binding:"required,min=10,max=500"`
}

// Comment suspension type constants
const (
	SuspensionTypeWarning   = "warning"
	SuspensionTypeTemporary = "temporary"
	SuspensionTypePermanent = "permanent"
)

// BroadcasterFollow represents a user following a broadcaster
type BroadcasterFollow struct {
	ID              uuid.UUID `json:"id" db:"id"`
	UserID          uuid.UUID `json:"user_id" db:"user_id"`
	BroadcasterID   string    `json:"broadcaster_id" db:"broadcaster_id"`
	BroadcasterName string    `json:"broadcaster_name" db:"broadcaster_name"`
	CreatedAt       time.Time `json:"created_at" db:"created_at"`
}

// BroadcasterProfile represents a broadcaster's profile information
type BroadcasterProfile struct {
	BroadcasterID   string    `json:"broadcaster_id"`
	BroadcasterName string    `json:"broadcaster_name"`
	DisplayName     string    `json:"display_name"`
	AvatarURL       *string   `json:"avatar_url,omitempty"`
	Bio             *string   `json:"bio,omitempty"`
	TwitchURL       string    `json:"twitch_url"`
	TotalClips      int       `json:"total_clips"`
	FollowerCount   int       `json:"follower_count"`
	TotalViews      int64     `json:"total_views"`
	AvgVoteScore    float64   `json:"avg_vote_score"`
	IsFollowing     bool      `json:"is_following"` // Whether the current user is following
	UpdatedAt       time.Time `json:"updated_at"`
}

// PopularBroadcaster represents a broadcaster summary for navigation/discovery
type PopularBroadcaster struct {
	BroadcasterID   string `json:"broadcaster_id"`
	BroadcasterName string `json:"broadcaster_name"`
	ClipCount       int    `json:"clip_count"`
}

// BroadcasterRanking represents a broadcaster's engagement-weighted rank
type BroadcasterRanking struct {
	BroadcasterID       string    `json:"broadcaster_id" db:"broadcaster_id"`
	BroadcasterName     string    `json:"broadcaster_name" db:"broadcaster_name"`
	TotalClips          int       `json:"total_clips" db:"total_clips"`
	HumanSubmittedClips int       `json:"human_submitted_clips" db:"human_submitted_clips"`
	TotalVoteScore      int64     `json:"total_vote_score" db:"total_vote_score"`
	TotalViews          int64     `json:"total_views" db:"total_views"`
	TotalComments       int64     `json:"total_comments" db:"total_comments"`
	UniqueCommenters    int64     `json:"unique_commenters" db:"unique_commenters"`
	EngagementScore     float64   `json:"engagement_score" db:"engagement_score"`
	FollowerCount       int       `json:"follower_count" db:"follower_count"`
	LastCalculated      time.Time `json:"last_calculated" db:"last_calculated"`
}

// EmailLog represents a comprehensive email event log from SendGrid webhooks
type EmailLog struct {
	ID                uuid.UUID  `json:"id" db:"id"`
	UserID            *uuid.UUID `json:"user_id,omitempty" db:"user_id"`
	Template          *string    `json:"template,omitempty" db:"template"`
	Recipient         string     `json:"recipient" db:"recipient"`
	Status            string     `json:"status" db:"status"`
	EventType         string     `json:"event_type" db:"event_type"`
	SendGridMessageID *string    `json:"sendgrid_message_id,omitempty" db:"sendgrid_message_id"`
	SendGridEventID   *string    `json:"sendgrid_event_id,omitempty" db:"sendgrid_event_id"`
	BounceType        *string    `json:"bounce_type,omitempty" db:"bounce_type"`
	BounceReason      *string    `json:"bounce_reason,omitempty" db:"bounce_reason"`
	SpamReportReason  *string    `json:"spam_report_reason,omitempty" db:"spam_report_reason"`
	LinkURL           *string    `json:"link_url,omitempty" db:"link_url"`
	IPAddress         *string    `json:"ip_address,omitempty" db:"ip_address"`
	UserAgent         *string    `json:"user_agent,omitempty" db:"user_agent"`
	Metadata          *string    `json:"metadata,omitempty" db:"metadata"` // JSONB stored as string
	SentAt            *time.Time `json:"sent_at,omitempty" db:"sent_at"`
	DeliveredAt       *time.Time `json:"delivered_at,omitempty" db:"delivered_at"`
	OpenedAt          *time.Time `json:"opened_at,omitempty" db:"opened_at"`
	ClickedAt         *time.Time `json:"clicked_at,omitempty" db:"clicked_at"`
	BouncedAt         *time.Time `json:"bounced_at,omitempty" db:"bounced_at"`
	SpamReportedAt    *time.Time `json:"spam_reported_at,omitempty" db:"spam_reported_at"`
	UnsubscribedAt    *time.Time `json:"unsubscribed_at,omitempty" db:"unsubscribed_at"`
	CreatedAt         time.Time  `json:"created_at" db:"created_at"`
	UpdatedAt         time.Time  `json:"updated_at" db:"updated_at"`
}

// EmailMetricsSummary represents aggregated email metrics
type EmailMetricsSummary struct {
	ID                uuid.UUID `json:"id" db:"id"`
	PeriodStart       time.Time `json:"period_start" db:"period_start"`
	PeriodEnd         time.Time `json:"period_end" db:"period_end"`
	Granularity       string    `json:"granularity" db:"granularity"` // hourly, daily
	Template          *string   `json:"template,omitempty" db:"template"`
	TotalSent         int       `json:"total_sent" db:"total_sent"`
	TotalDelivered    int       `json:"total_delivered" db:"total_delivered"`
	TotalBounced      int       `json:"total_bounced" db:"total_bounced"`
	TotalHardBounced  int       `json:"total_hard_bounced" db:"total_hard_bounced"`
	TotalSoftBounced  int       `json:"total_soft_bounced" db:"total_soft_bounced"`
	TotalDropped      int       `json:"total_dropped" db:"total_dropped"`
	TotalOpened       int       `json:"total_opened" db:"total_opened"`
	TotalClicked      int       `json:"total_clicked" db:"total_clicked"`
	TotalSpamReports  int       `json:"total_spam_reports" db:"total_spam_reports"`
	TotalUnsubscribes int       `json:"total_unsubscribes" db:"total_unsubscribes"`
	UniqueOpened      int       `json:"unique_opened" db:"unique_opened"`
	UniqueClicked     int       `json:"unique_clicked" db:"unique_clicked"`
	BounceRate        *float64  `json:"bounce_rate,omitempty" db:"bounce_rate"`
	OpenRate          *float64  `json:"open_rate,omitempty" db:"open_rate"`
	ClickRate         *float64  `json:"click_rate,omitempty" db:"click_rate"`
	SpamRate          *float64  `json:"spam_rate,omitempty" db:"spam_rate"`
	CreatedAt         time.Time `json:"created_at" db:"created_at"`
	UpdatedAt         time.Time `json:"updated_at" db:"updated_at"`
}

// EmailAlert represents an alert triggered by email metrics
type EmailAlert struct {
	ID             uuid.UUID  `json:"id" db:"id"`
	AlertType      string     `json:"alert_type" db:"alert_type"` // high_bounce_rate, high_complaint_rate, etc.
	Severity       string     `json:"severity" db:"severity"`     // warning, critical
	MetricName     string     `json:"metric_name" db:"metric_name"`
	CurrentValue   *float64   `json:"current_value,omitempty" db:"current_value"`
	ThresholdValue *float64   `json:"threshold_value,omitempty" db:"threshold_value"`
	PeriodStart    time.Time  `json:"period_start" db:"period_start"`
	PeriodEnd      time.Time  `json:"period_end" db:"period_end"`
	Message        string     `json:"message" db:"message"`
	Metadata       *string    `json:"metadata,omitempty" db:"metadata"` // JSONB stored as string
	TriggeredAt    time.Time  `json:"triggered_at" db:"triggered_at"`
	AcknowledgedAt *time.Time `json:"acknowledged_at,omitempty" db:"acknowledged_at"`
	AcknowledgedBy *uuid.UUID `json:"acknowledged_by,omitempty" db:"acknowledged_by"`
	ResolvedAt     *time.Time `json:"resolved_at,omitempty" db:"resolved_at"`
	CreatedAt      time.Time  `json:"created_at" db:"created_at"`
}

// Email log status constants
const (
	EmailLogStatusSent        = "sent"
	EmailLogStatusDelivered   = "delivered"
	EmailLogStatusBounce      = "bounce"
	EmailLogStatusDropped     = "dropped"
	EmailLogStatusOpen        = "open"
	EmailLogStatusClick       = "click"
	EmailLogStatusSpamReport  = "spam_report"
	EmailLogStatusUnsubscribe = "unsubscribe"
	EmailLogStatusDeferred    = "deferred"
	EmailLogStatusProcessed   = "processed"
)

// Email alert types
const (
	EmailAlertTypeHighBounceRate    = "high_bounce_rate"
	EmailAlertTypeHighComplaintRate = "high_complaint_rate"
	EmailAlertTypeSendErrors        = "send_errors"
	EmailAlertTypeOpenRateDrop      = "open_rate_drop"
	EmailAlertTypeUnsubscribeSpike  = "unsubscribe_spike"
)

// Email alert severities
const (
	EmailAlertSeverityWarning  = "warning"
	EmailAlertSeverityCritical = "critical"
)

// SendGridWebhookEvent represents an incoming webhook event from SendGrid
type SendGridWebhookEvent struct {
	Email                 string                 `json:"email"`
	Timestamp             int64                  `json:"timestamp"`
	Event                 string                 `json:"event"`
	SgMessageID           string                 `json:"sg_message_id"`
	SgEventID             string                 `json:"sg_event_id"`
	Category              []string               `json:"category,omitempty"`
	Type                  string                 `json:"type,omitempty"`      // For bounce events: bounce, blocked, etc.
	Reason                string                 `json:"reason,omitempty"`    // Bounce/drop reason
	Status                string                 `json:"status,omitempty"`    // Bounce status code
	URL                   string                 `json:"url,omitempty"`       // Clicked URL
	IP                    string                 `json:"ip,omitempty"`        // IP address
	UserAgent             string                 `json:"useragent,omitempty"` // User agent
	Response              string                 `json:"response,omitempty"`  // SMTP response
	Attempt               string                 `json:"attempt,omitempty"`   // Deferred attempt number
	CustomArgs            map[string]interface{} `json:"custom_args,omitempty"`
	ASMGroupID            int                    `json:"asm_group_id,omitempty"` // Unsubscribe group ID
	MarketingCampaignID   string                 `json:"marketing_campaign_id,omitempty"`
	MarketingCampaignName string                 `json:"marketing_campaign_name,omitempty"`
}

// Feed represents a user-created feed
type Feed struct {
	ID            uuid.UUID `json:"id" db:"id"`
	UserID        uuid.UUID `json:"user_id" db:"user_id"`
	Name          string    `json:"name" db:"name"`
	Description   *string   `json:"description,omitempty" db:"description"`
	Icon          *string   `json:"icon,omitempty" db:"icon"`
	IsPublic      bool      `json:"is_public" db:"is_public"`
	FollowerCount int       `json:"follower_count" db:"follower_count"`
	CreatedAt     time.Time `json:"created_at" db:"created_at"`
	UpdatedAt     time.Time `json:"updated_at" db:"updated_at"`
}

// FeedWithOwner includes owner information
type FeedWithOwner struct {
	Feed
	Owner *User `json:"owner,omitempty"`
}

// FeedItem represents a clip in a feed
type FeedItem struct {
	ID       uuid.UUID `json:"id" db:"id"`
	FeedID   uuid.UUID `json:"feed_id" db:"feed_id"`
	ClipID   uuid.UUID `json:"clip_id" db:"clip_id"`
	Position int       `json:"position" db:"position"`
	AddedAt  time.Time `json:"added_at" db:"added_at"`
}

// FeedItemWithClip includes clip information
type FeedItemWithClip struct {
	FeedItem
	Clip *Clip `json:"clip,omitempty"`
}

// FeedFollow represents a user following a feed
type FeedFollow struct {
	ID         uuid.UUID `json:"id" db:"id"`
	UserID     uuid.UUID `json:"user_id" db:"user_id"`
	FeedID     uuid.UUID `json:"feed_id" db:"feed_id"`
	FollowedAt time.Time `json:"followed_at" db:"followed_at"`
}

// CreateFeedRequest represents the request to create a feed
type CreateFeedRequest struct {
	Name        string  `json:"name" binding:"required,min=1,max=255"`
	Description *string `json:"description,omitempty" binding:"omitempty,max=1000"`
	Icon        *string `json:"icon,omitempty" binding:"omitempty,max=100"`
	IsPublic    *bool   `json:"is_public,omitempty"`
}

// UpdateFeedRequest represents the request to update a feed
type UpdateFeedRequest struct {
	Name        *string `json:"name,omitempty" binding:"omitempty,min=1,max=255"`
	Description *string `json:"description,omitempty" binding:"omitempty,max=1000"`
	Icon        *string `json:"icon,omitempty" binding:"omitempty,max=100"`
	IsPublic    *bool   `json:"is_public,omitempty"`
}

// AddClipToFeedRequest represents the request to add a clip to a feed
type AddClipToFeedRequest struct {
	ClipID uuid.UUID `json:"clip_id" binding:"required"`
}

// ReorderFeedClipsRequest represents the request to reorder clips in a feed
type ReorderFeedClipsRequest struct {
	ClipIDs []uuid.UUID `json:"clip_ids" binding:"required"`
}

// UserFollow represents a user following another user
type UserFollow struct {
	ID          uuid.UUID `json:"id" db:"id"`
	FollowerID  uuid.UUID `json:"follower_id" db:"follower_id"`
	FollowingID uuid.UUID `json:"following_id" db:"following_id"`
	CreatedAt   time.Time `json:"created_at" db:"created_at"`
}

// UserBlock represents a user blocking another user
type UserBlock struct {
	ID            uuid.UUID `json:"id" db:"id"`
	UserID        uuid.UUID `json:"user_id" db:"user_id"`
	BlockedUserID uuid.UUID `json:"blocked_user_id" db:"blocked_user_id"`
	BlockedAt     time.Time `json:"blocked_at" db:"blocked_at"`
}

// BlockedUser represents a user in a blocked users list
type BlockedUser struct {
	ID          uuid.UUID `json:"id" db:"id"`
	Username    string    `json:"username" db:"username"`
	DisplayName string    `json:"display_name" db:"display_name"`
	AvatarURL   *string   `json:"avatar_url,omitempty" db:"avatar_url"`
	Bio         *string   `json:"bio,omitempty" db:"bio"`
	KarmaPoints int       `json:"karma_points" db:"karma_points"`
	BlockedAt   time.Time `json:"blocked_at" db:"blocked_at"`
}

// UserActivity represents a user activity entry
type UserActivity struct {
	ID           uuid.UUID  `json:"id" db:"id"`
	UserID       uuid.UUID  `json:"user_id" db:"user_id"`
	ActivityType string     `json:"activity_type" db:"activity_type"`
	TargetID     *uuid.UUID `json:"target_id,omitempty" db:"target_id"`
	TargetType   *string    `json:"target_type,omitempty" db:"target_type"`
	Metadata     *string    `json:"metadata,omitempty" db:"metadata"` // JSONB stored as string
	CreatedAt    time.Time  `json:"created_at" db:"created_at"`
}

// Activity type constants
const (
	ActivityTypeClipSubmitted       = "clip_submitted"
	ActivityTypeUpvote              = "upvote"
	ActivityTypeDownvote            = "downvote"
	ActivityTypeComment             = "comment"
	ActivityTypeUserFollowed        = "user_followed"
	ActivityTypeBroadcasterFollowed = "broadcaster_followed"
)

// UserProfile represents a complete user profile with stats
type UserProfile struct {
	User
	Stats        UserProfileStats `json:"stats"`
	IsFollowing  bool             `json:"is_following"`   // Whether the current user is following this user
	IsFollowedBy bool             `json:"is_followed_by"` // Whether this user is following the current user
}

// UserProfileStats represents statistics for a user profile
type UserProfileStats struct {
	ClipsSubmitted       int `json:"clips_submitted"`
	TotalUpvotes         int `json:"total_upvotes"`
	TotalComments        int `json:"total_comments"`
	ClipsFeatured        int `json:"clips_featured"`
	BroadcastersFollowed int `json:"broadcasters_followed"`
}

// UserActivityItem represents a single activity item with expanded data
type UserActivityItem struct {
	UserActivity
	Username    string  `json:"username"`
	UserAvatar  *string `json:"user_avatar"`
	ClipTitle   *string `json:"clip_title,omitempty"`
	ClipID      *string `json:"clip_id,omitempty"`
	CommentText *string `json:"comment_text,omitempty"`
	TargetUser  *string `json:"target_user,omitempty"`
}

// SocialLinks represents social media links
type SocialLinks struct {
	Twitter *string `json:"twitter,omitempty"`
	Twitch  *string `json:"twitch,omitempty"`
	Discord *string `json:"discord,omitempty"`
	YouTube *string `json:"youtube,omitempty"`
	Website *string `json:"website,omitempty"`
}

// UpdateSocialLinksRequest represents the request to update social links
type UpdateSocialLinksRequest struct {
	Twitter *string `json:"twitter,omitempty" binding:"omitempty,max=255"`
	Twitch  *string `json:"twitch,omitempty" binding:"omitempty,max=255"`
	Discord *string `json:"discord,omitempty" binding:"omitempty,max=255"`
	YouTube *string `json:"youtube,omitempty" binding:"omitempty,max=255"`
	Website *string `json:"website,omitempty" binding:"omitempty,url,max=255"`
}

// FollowerUser represents a user in a followers/following list
type FollowerUser struct {
	ID          uuid.UUID `json:"id" db:"id"`
	Username    string    `json:"username" db:"username"`
	DisplayName string    `json:"display_name" db:"display_name"`
	AvatarURL   *string   `json:"avatar_url,omitempty" db:"avatar_url"`
	Bio         *string   `json:"bio,omitempty" db:"bio"`
	KarmaPoints int       `json:"karma_points" db:"karma_points"`
	FollowedAt  time.Time `json:"followed_at" db:"followed_at"`
	IsFollowing bool      `json:"is_following"` // Whether the current user is following this user
}

// Category represents a high-level content category
type Category struct {
	ID          uuid.UUID `json:"id" db:"id"`
	Name        string    `json:"name" db:"name"`
	Slug        string    `json:"slug" db:"slug"`
	Description *string   `json:"description,omitempty" db:"description"`
	Icon        *string   `json:"icon,omitempty" db:"icon"`
	Position    int       `json:"position" db:"position"`
	CategoryType string   `json:"category_type" db:"category_type"`
	IsFeatured   bool     `json:"is_featured" db:"is_featured"`
	IsCustom     bool     `json:"is_custom" db:"is_custom"`
	CreatedByUserID *uuid.UUID `json:"created_by_user_id,omitempty" db:"created_by_user_id"`
	CreatedAt   time.Time `json:"created_at" db:"created_at"`
	UpdatedAt   time.Time `json:"updated_at" db:"updated_at"`
}

// GameEntity represents a full game entity from Twitch (with database fields)
type GameEntity struct {
	ID           uuid.UUID `json:"id" db:"id"`
	TwitchGameID string    `json:"twitch_game_id" db:"twitch_game_id"`
	Name         string    `json:"name" db:"name"`
	Slug         string    `json:"slug" db:"slug"`
	BoxArtURL    *string   `json:"box_art_url,omitempty" db:"box_art_url"`
	IGDBID       *string   `json:"igdb_id,omitempty" db:"igdb_id"`
	CreatedAt    time.Time `json:"created_at" db:"created_at"`
	UpdatedAt    time.Time `json:"updated_at" db:"updated_at"`
}

// Game is an alias for GameEntity for backward compatibility
type Game = GameEntity

// GameWithStats represents a game with additional statistics
type GameWithStats struct {
	GameEntity
	ClipCount     int  `json:"clip_count" db:"clip_count"`
	FollowerCount int  `json:"follower_count" db:"follower_count"`
	IsFollowing   bool `json:"is_following"` // Whether the current user is following
}

// CategoryGame represents the many-to-many relationship between categories and games
type CategoryGame struct {
	GameID     uuid.UUID `json:"game_id" db:"game_id"`
	CategoryID uuid.UUID `json:"category_id" db:"category_id"`
	CreatedAt  time.Time `json:"created_at" db:"created_at"`
}

// GameFollow represents a user following a game
type GameFollow struct {
	ID         uuid.UUID `json:"id" db:"id"`
	UserID     uuid.UUID `json:"user_id" db:"user_id"`
	GameID     uuid.UUID `json:"game_id" db:"game_id"`
	FollowedAt time.Time `json:"followed_at" db:"followed_at"`
}

// TrendingGame represents a game with trending statistics
type TrendingGame struct {
	ID              uuid.UUID `json:"id" db:"id"`
	TwitchGameID    string    `json:"twitch_game_id" db:"twitch_game_id"`
	Name            string    `json:"name" db:"name"`
	BoxArtURL       *string   `json:"box_art_url,omitempty" db:"box_art_url"`
	RecentClipCount int       `json:"recent_clip_count" db:"recent_clip_count"`
	TotalVoteScore  int       `json:"total_vote_score" db:"total_vote_score"`
	FollowerCount   int       `json:"follower_count" db:"follower_count"`
}

// DiscoveryList represents a curated list of clips
type DiscoveryList struct {
	ID           uuid.UUID  `json:"id" db:"id"`
	Name         string     `json:"name" db:"name"`
	Slug         string     `json:"slug" db:"slug"`
	Description  *string    `json:"description,omitempty" db:"description"`
	IsFeatured   bool       `json:"is_featured" db:"is_featured"`
	IsActive     bool       `json:"is_active" db:"is_active"`
	DisplayOrder int        `json:"display_order" db:"display_order"`
	CreatedBy    *uuid.UUID `json:"created_by,omitempty" db:"created_by"`
	CreatedAt    time.Time  `json:"created_at" db:"created_at"`
	UpdatedAt    time.Time  `json:"updated_at" db:"updated_at"`
}

// DiscoveryListWithStats includes additional statistics about the list
type DiscoveryListWithStats struct {
	DiscoveryList
	ClipCount     int    `json:"clip_count" db:"clip_count"`
	FollowerCount int    `json:"follower_count" db:"follower_count"`
	IsFollowing   bool   `json:"is_following"`
	IsBookmarked  bool   `json:"is_bookmarked"`
	PreviewClips  []Clip `json:"preview_clips,omitempty"`
}

// DiscoveryListClip represents the relationship between a list and a clip
type DiscoveryListClip struct {
	ID           uuid.UUID `json:"id" db:"id"`
	ListID       uuid.UUID `json:"list_id" db:"list_id"`
	ClipID       uuid.UUID `json:"clip_id" db:"clip_id"`
	DisplayOrder int       `json:"display_order" db:"display_order"`
	AddedAt      time.Time `json:"added_at" db:"added_at"`
}

// CreateDiscoveryListRequest represents the request to create a discovery list
type CreateDiscoveryListRequest struct {
	Name        string  `json:"name" binding:"required,min=1,max=200"`
	Description *string `json:"description,omitempty" binding:"omitempty,max=1000"`
	IsFeatured  *bool   `json:"is_featured,omitempty"`
}

// UpdateDiscoveryListRequest represents the request to update a discovery list
type UpdateDiscoveryListRequest struct {
	Name        *string `json:"name,omitempty" binding:"omitempty,min=1,max=200"`
	Description *string `json:"description,omitempty" binding:"omitempty,max=1000"`
	IsFeatured  *bool   `json:"is_featured,omitempty"`
	IsActive    *bool   `json:"is_active,omitempty"`
}

// AddClipToListRequest represents the request to add a clip to a discovery list
type AddClipToListRequest struct {
	ClipID uuid.UUID `json:"clip_id" binding:"required"`
}

// ReorderListClipsRequest represents the request to reorder clips in a discovery list
type ReorderListClipsRequest struct {
	ClipIDs []uuid.UUID `json:"clip_ids" binding:"required,min=1,max=200"`
}

// BroadcasterLiveStatus represents the live streaming status of a broadcaster
type BroadcasterLiveStatus struct {
	BroadcasterID string     `json:"broadcaster_id" db:"broadcaster_id"`
	UserLogin     *string    `json:"user_login,omitempty" db:"user_login"`
	UserName      *string    `json:"user_name,omitempty" db:"user_name"`
	IsLive        bool       `json:"is_live" db:"is_live"`
	StreamTitle   *string    `json:"stream_title,omitempty" db:"stream_title"`
	GameName      *string    `json:"game_name,omitempty" db:"game_name"`
	ViewerCount   int        `json:"viewer_count" db:"viewer_count"`
	StartedAt     *time.Time `json:"started_at,omitempty" db:"started_at"`
	LastChecked   time.Time  `json:"last_checked" db:"last_checked"`
	CreatedAt     time.Time  `json:"created_at" db:"created_at"`
	UpdatedAt     time.Time  `json:"updated_at" db:"updated_at"`
}

// BroadcasterSyncStatus represents the sync status for broadcaster live tracking
type BroadcasterSyncStatus struct {
	BroadcasterID   string     `json:"broadcaster_id" db:"broadcaster_id"`
	IsLive          bool       `json:"is_live" db:"is_live"`
	StreamStartedAt *time.Time `json:"stream_started_at,omitempty" db:"stream_started_at"`
	LastSynced      time.Time  `json:"last_synced" db:"last_synced"`
	GameName        *string    `json:"game_name,omitempty" db:"game_name"`
	ViewerCount     int        `json:"viewer_count" db:"viewer_count"`
	StreamTitle     *string    `json:"stream_title,omitempty" db:"stream_title"`
	CreatedAt       time.Time  `json:"created_at" db:"created_at"`
	UpdatedAt       time.Time  `json:"updated_at" db:"updated_at"`
}

// BroadcasterSyncLog represents a log entry for broadcaster sync events
type BroadcasterSyncLog struct {
	ID            uuid.UUID `json:"id" db:"id"`
	BroadcasterID string    `json:"broadcaster_id" db:"broadcaster_id"`
	SyncTime      time.Time `json:"sync_time" db:"sync_time"`
	StatusChange  *string   `json:"status_change,omitempty" db:"status_change"`
	Error         *string   `json:"error,omitempty" db:"error"`
	CreatedAt     time.Time `json:"created_at" db:"created_at"`
}

// Stream represents a Twitch stream with metadata and status
type Stream struct {
	ID               uuid.UUID  `json:"id" db:"id"`
	StreamerUsername string     `json:"streamer_username" db:"streamer_username"`
	StreamerUserID   *string    `json:"streamer_user_id,omitempty" db:"streamer_user_id"`
	DisplayName      *string    `json:"display_name,omitempty" db:"display_name"`
	IsLive           bool       `json:"is_live" db:"is_live"`
	LastWentLive     *time.Time `json:"last_went_live,omitempty" db:"last_went_live"`
	LastWentOffline  *time.Time `json:"last_went_offline,omitempty" db:"last_went_offline"`
	GameName         *string    `json:"game_name,omitempty" db:"game_name"`
	Title            *string    `json:"title,omitempty" db:"title"`
	ViewerCount      int        `json:"viewer_count" db:"viewer_count"`
	CreatedAt        time.Time  `json:"created_at" db:"created_at"`
	UpdatedAt        time.Time  `json:"updated_at" db:"updated_at"`
}

// StreamSession represents a user's watch session for a stream
type StreamSession struct {
	ID                   uuid.UUID  `json:"id" db:"id"`
	UserID               uuid.UUID  `json:"user_id" db:"user_id"`
	StreamID             uuid.UUID  `json:"stream_id" db:"stream_id"`
	StartedAt            time.Time  `json:"started_at" db:"started_at"`
	EndedAt              *time.Time `json:"ended_at,omitempty" db:"ended_at"`
	WatchDurationSeconds int        `json:"watch_duration_seconds" db:"watch_duration_seconds"`
}

// StreamInfo represents stream status information returned to the frontend
type StreamInfo struct {
	StreamerUsername string     `json:"streamer_username"`
	IsLive           bool       `json:"is_live"`
	Title            *string    `json:"title,omitempty"`
	GameName         *string    `json:"game_name,omitempty"`
	ViewerCount      int        `json:"viewer_count"`
	ThumbnailURL     *string    `json:"thumbnail_url,omitempty"`
	StartedAt        *time.Time `json:"started_at,omitempty"`
	LastWentOffline  *time.Time `json:"last_went_offline,omitempty"`
}

// StreamFollow represents a user following a streamer for live notifications
type StreamFollow struct {
	ID                   uuid.UUID `json:"id" db:"id"`
	UserID               uuid.UUID `json:"user_id" db:"user_id"`
	StreamerUsername     string    `json:"streamer_username" db:"streamer_username"`
	NotificationsEnabled bool      `json:"notifications_enabled" db:"notifications_enabled"`
	CreatedAt            time.Time `json:"created_at" db:"created_at"`
	UpdatedAt            time.Time `json:"updated_at" db:"updated_at"`
}

// StreamFollowRequest represents a request to follow/unfollow a streamer
type StreamFollowRequest struct {
	NotificationsEnabled *bool `json:"notifications_enabled,omitempty"`
}

// ClipFromStreamRequest represents a request to create a clip from a live stream
type ClipFromStreamRequest struct {
	StreamerUsername string  `json:"streamer_username" binding:"required"`
	StartTime        float64 `json:"start_time" binding:"required,min=0"` // Seconds into VOD
	EndTime          float64 `json:"end_time" binding:"required,min=0,gtfield=StartTime"`
	Quality          string  `json:"quality" binding:"required,oneof=source 1080p 720p"`
	Title            string  `json:"title" binding:"required,min=3,max=255"`
}

// ClipFromStreamResponse represents the response when creating a clip from a stream
type ClipFromStreamResponse struct {
	ClipID string `json:"clip_id"`
	Status string `json:"status"` // 'processing'
}

// ClipExtractionJob represents a job for extracting and processing a clip from a stream VOD
type ClipExtractionJob struct {
	ClipID    string  `json:"clip_id"`
	VODURL    string  `json:"vod_url"`
	StartTime float64 `json:"start_time"`
	EndTime   float64 `json:"end_time"`
	Quality   string  `json:"quality"`
}

// Community represents a community space
type Community struct {
	ID          uuid.UUID `json:"id" db:"id"`
	Name        string    `json:"name" db:"name"`
	Slug        string    `json:"slug" db:"slug"`
	Description *string   `json:"description,omitempty" db:"description"`
	Icon        *string   `json:"icon,omitempty" db:"icon"`
	OwnerID     uuid.UUID `json:"owner_id" db:"owner_id"`
	IsPublic    bool      `json:"is_public" db:"is_public"`
	MemberCount int       `json:"member_count" db:"member_count"`
	Rules       *string   `json:"rules,omitempty" db:"rules"`
	CreatedAt   time.Time `json:"created_at" db:"created_at"`
	UpdatedAt   time.Time `json:"updated_at" db:"updated_at"`
}

// CommunityWithOwner includes owner information
type CommunityWithOwner struct {
	Community
	Owner *User `json:"owner,omitempty"`
}

// CommunityWithStats includes additional statistics
type CommunityWithStats struct {
	Community
	ClipCount       int     `json:"clip_count" db:"clip_count"`
	DiscussionCount int     `json:"discussion_count" db:"discussion_count"`
	IsMember        bool    `json:"is_member"`
	UserRole        *string `json:"user_role,omitempty"` // admin, mod, member, or null if not a member
}

// CommunityMember represents a member of a community
type CommunityMember struct {
	ID          uuid.UUID `json:"id" db:"id"`
	CommunityID uuid.UUID `json:"community_id" db:"community_id"`
	UserID      uuid.UUID `json:"user_id" db:"user_id"`
	Role        string    `json:"role" db:"role"` // admin, mod, member
	JoinedAt    time.Time `json:"joined_at" db:"joined_at"`
}

// CommunityMemberWithUser includes user information
type CommunityMemberWithUser struct {
	CommunityMember
	User *User `json:"user,omitempty"`
}

// CommunityBan represents a banned user in a community
type CommunityBan struct {
	ID             uuid.UUID  `json:"id" db:"id"`
	CommunityID    uuid.UUID  `json:"community_id" db:"community_id"`
	BannedUserID   uuid.UUID  `json:"banned_user_id" db:"banned_user_id"`
	BannedByUserID *uuid.UUID `json:"banned_by_user_id,omitempty" db:"banned_by_user_id"`
	Reason         *string    `json:"reason,omitempty" db:"reason"`
	BannedAt       time.Time  `json:"banned_at" db:"banned_at"`
}

// CommunityBanWithUser includes user information
type CommunityBanWithUser struct {
	CommunityBan
	BannedUser   *User `json:"banned_user,omitempty"`
	BannedByUser *User `json:"banned_by_user,omitempty"`
}

// CommunityClip represents a clip in a community feed
type CommunityClip struct {
	ID            uuid.UUID  `json:"id" db:"id"`
	CommunityID   uuid.UUID  `json:"community_id" db:"community_id"`
	ClipID        uuid.UUID  `json:"clip_id" db:"clip_id"`
	AddedByUserID *uuid.UUID `json:"added_by_user_id,omitempty" db:"added_by_user_id"`
	AddedAt       time.Time  `json:"added_at" db:"added_at"`
}

// CommunityClipWithClip includes clip information
type CommunityClipWithClip struct {
	CommunityClip
	Clip *Clip `json:"clip,omitempty"`
}

// CommunityDiscussion represents a discussion thread in a community
type CommunityDiscussion struct {
	ID           uuid.UUID `json:"id" db:"id"`
	CommunityID  uuid.UUID `json:"community_id" db:"community_id"`
	UserID       uuid.UUID `json:"user_id" db:"user_id"`
	Title        string    `json:"title" db:"title"`
	Content      string    `json:"content" db:"content"`
	IsPinned     bool      `json:"is_pinned" db:"is_pinned"`
	IsResolved   bool      `json:"is_resolved" db:"is_resolved"`
	VoteScore    int       `json:"vote_score" db:"vote_score"`
	CommentCount int       `json:"comment_count" db:"comment_count"`
	CreatedAt    time.Time `json:"created_at" db:"created_at"`
	UpdatedAt    time.Time `json:"updated_at" db:"updated_at"`
}

// CommunityDiscussionWithUser includes user information
type CommunityDiscussionWithUser struct {
	CommunityDiscussion
	User *User `json:"user,omitempty"`
}

// CommunityDiscussionComment represents a comment on a discussion thread
type CommunityDiscussionComment struct {
	ID              uuid.UUID  `json:"id" db:"id"`
	DiscussionID    uuid.UUID  `json:"discussion_id" db:"discussion_id"`
	UserID          uuid.UUID  `json:"user_id" db:"user_id"`
	ParentCommentID *uuid.UUID `json:"parent_comment_id,omitempty" db:"parent_comment_id"`
	Content         string     `json:"content" db:"content"`
	VoteScore       int        `json:"vote_score" db:"vote_score"`
	IsEdited        bool       `json:"is_edited" db:"is_edited"`
	IsRemoved       bool       `json:"is_removed" db:"is_removed"`
	RemovedReason   *string    `json:"removed_reason,omitempty" db:"removed_reason"`
	CreatedAt       time.Time  `json:"created_at" db:"created_at"`
	UpdatedAt       time.Time  `json:"updated_at" db:"updated_at"`
}

// CommunityDiscussionCommentWithUser includes user information
type CommunityDiscussionCommentWithUser struct {
	CommunityDiscussionComment
	User *User `json:"user,omitempty"`
}

// CommunityDiscussionVote represents a vote on a discussion or comment
type CommunityDiscussionVote struct {
	ID           uuid.UUID  `json:"id" db:"id"`
	UserID       uuid.UUID  `json:"user_id" db:"user_id"`
	DiscussionID *uuid.UUID `json:"discussion_id,omitempty" db:"discussion_id"`
	CommentID    *uuid.UUID `json:"comment_id,omitempty" db:"comment_id"`
	VoteType     int16      `json:"vote_type" db:"vote_type"` // 1 for upvote, -1 for downvote
	CreatedAt    time.Time  `json:"created_at" db:"created_at"`
}

// CreateCommunityRequest represents the request to create a community
type CreateCommunityRequest struct {
	Name        string  `json:"name" binding:"required,min=3,max=255"`
	Description *string `json:"description,omitempty" binding:"omitempty,max=5000"`
	Icon        *string `json:"icon,omitempty" binding:"omitempty,max=100"`
	IsPublic    *bool   `json:"is_public,omitempty"`
	Rules       *string `json:"rules,omitempty" binding:"omitempty,max=10000"`
}

// UpdateCommunityRequest represents the request to update a community
type UpdateCommunityRequest struct {
	Name        *string `json:"name,omitempty" binding:"omitempty,min=3,max=255"`
	Description *string `json:"description,omitempty" binding:"omitempty,max=5000"`
	Icon        *string `json:"icon,omitempty" binding:"omitempty,max=100"`
	IsPublic    *bool   `json:"is_public,omitempty"`
	Rules       *string `json:"rules,omitempty" binding:"omitempty,max=10000"`
}

// AddMemberRequest represents the request to add a member to a community
type AddMemberRequest struct {
	UserID uuid.UUID `json:"user_id" binding:"required"`
	Role   *string   `json:"role,omitempty" binding:"omitempty,oneof=admin mod member"`
}

// UpdateMemberRoleRequest represents the request to update a member's role
type UpdateMemberRoleRequest struct {
	Role string `json:"role" binding:"required,oneof=admin mod member"`
}

// BanMemberRequest represents the request to ban a member from a community
type BanMemberRequest struct {
	UserID uuid.UUID `json:"user_id" binding:"required"`
	Reason *string   `json:"reason,omitempty" binding:"omitempty,max=1000"`
}

// AddClipToCommunityRequest represents the request to add a clip to a community
type AddClipToCommunityRequest struct {
	ClipID uuid.UUID `json:"clip_id" binding:"required"`
}

// CreateDiscussionRequest represents the request to create a discussion thread
type CreateDiscussionRequest struct {
	Title   string `json:"title" binding:"required,min=3,max=500"`
	Content string `json:"content" binding:"required,min=10,max=10000"`
}

// UpdateDiscussionRequest represents the request to update a discussion thread
type UpdateDiscussionRequest struct {
	Title      *string `json:"title,omitempty" binding:"omitempty,min=3,max=500"`
	Content    *string `json:"content,omitempty" binding:"omitempty,min=10,max=10000"`
	IsPinned   *bool   `json:"is_pinned,omitempty"`
	IsResolved *bool   `json:"is_resolved,omitempty"`
}

// CreateDiscussionCommentRequest represents the request to create a comment on a discussion
type CreateDiscussionCommentRequest struct {
	Content         string     `json:"content" binding:"required,min=1,max=10000"`
	ParentCommentID *uuid.UUID `json:"parent_comment_id,omitempty"`
}

// UpdateDiscussionCommentRequest represents the request to update a discussion comment
type UpdateDiscussionCommentRequest struct {
	Content string `json:"content" binding:"required,min=1,max=10000"`
}

// Community role constants
const (
	CommunityRoleAdmin  = "admin"
	CommunityRoleMod    = "mod"
	CommunityRoleMember = "member"
)

// AccountTypeConversion represents a conversion from one account type to another
type AccountTypeConversion struct {
	ID          uuid.UUID              `json:"id" db:"id"`
	UserID      uuid.UUID              `json:"user_id" db:"user_id"`
	OldType     string                 `json:"old_type" db:"old_type"`
	NewType     string                 `json:"new_type" db:"new_type"`
	Reason      *string                `json:"reason,omitempty" db:"reason"`
	ConvertedBy *uuid.UUID             `json:"converted_by,omitempty" db:"converted_by"`
	ConvertedAt time.Time              `json:"converted_at" db:"converted_at"`
	Metadata    map[string]interface{} `json:"metadata,omitempty" db:"metadata"`
	CreatedAt   time.Time              `json:"created_at" db:"created_at"`
}

// AccountTypeConversionWithUser includes user and converter information
type AccountTypeConversionWithUser struct {
	AccountTypeConversion
	User        *User `json:"user,omitempty"`
	ConvertedBy *User `json:"converted_by_user,omitempty"`
}

// ConvertToBroadcasterRequest represents the request to convert to broadcaster account type
type ConvertToBroadcasterRequest struct {
	Reason *string `json:"reason,omitempty" binding:"omitempty,max=500"`
}

// ConvertToModeratorRequest represents the request to convert to moderator account type
type ConvertToModeratorRequest struct {
	Reason *string `json:"reason,omitempty" binding:"omitempty,max=500"`
}

// AccountTypeResponse represents the response for account type queries
type AccountTypeResponse struct {
	AccountType       string                  `json:"account_type"`
	UpdatedAt         *time.Time              `json:"updated_at,omitempty"`
	Permissions       []string                `json:"permissions"`
	ConversionHistory []AccountTypeConversion `json:"conversion_history,omitempty"`
}

// UserMFA represents multi-factor authentication configuration for a user
type UserMFA struct {
	ID                     int        `json:"id" db:"id"`
	UserID                 uuid.UUID  `json:"user_id" db:"user_id"`
	Secret                 string     `json:"-" db:"secret"` // Never expose encrypted secret in JSON
	Enabled                bool       `json:"enabled" db:"enabled"`
	EnrolledAt             *time.Time `json:"enrolled_at,omitempty" db:"enrolled_at"`
	BackupCodes            []string   `json:"-" db:"backup_codes"` // Never expose hashed codes
	BackupCodesGeneratedAt *time.Time `json:"backup_codes_generated_at,omitempty" db:"backup_codes_generated_at"`
	MFARequired            bool       `json:"mfa_required" db:"mfa_required"`
	MFARequiredAt          *time.Time `json:"mfa_required_at,omitempty" db:"mfa_required_at"`
	GracePeriodEnd         *time.Time `json:"grace_period_end,omitempty" db:"grace_period_end"`
	CreatedAt              time.Time  `json:"created_at" db:"created_at"`
	UpdatedAt              time.Time  `json:"updated_at" db:"updated_at"`
}

// MFATrustedDevice represents a trusted device that can skip MFA for a period
type MFATrustedDevice struct {
	ID                int       `json:"id" db:"id"`
	UserID            uuid.UUID `json:"user_id" db:"user_id"`
	DeviceFingerprint string    `json:"device_fingerprint" db:"device_fingerprint"`
	DeviceName        *string   `json:"device_name,omitempty" db:"device_name"`
	IPAddress         *string   `json:"ip_address,omitempty" db:"ip_address"`
	UserAgent         *string   `json:"user_agent,omitempty" db:"user_agent"`
	TrustedAt         time.Time `json:"trusted_at" db:"trusted_at"`
	ExpiresAt         time.Time `json:"expires_at" db:"expires_at"`
	LastUsedAt        time.Time `json:"last_used_at" db:"last_used_at"`
	CreatedAt         time.Time `json:"created_at" db:"created_at"`
}

// MFAAuditLog represents an audit log entry for MFA-related actions
type MFAAuditLog struct {
	ID        int        `json:"id" db:"id"`
	UserID    *uuid.UUID `json:"user_id,omitempty" db:"user_id"`
	Action    string     `json:"action" db:"action"`
	Success   bool       `json:"success" db:"success"`
	IPAddress *string    `json:"ip_address,omitempty" db:"ip_address"`
	UserAgent *string    `json:"user_agent,omitempty" db:"user_agent"`
	Details   *string    `json:"details,omitempty" db:"details"`
	CreatedAt time.Time  `json:"created_at" db:"created_at"`
}

// MFA action constants for audit logging
const (
	MFAActionEnrollStart          = "mfa_enroll_start"
	MFAActionEnrollComplete       = "mfa_enroll_complete"
	MFAActionEnrollFailed         = "mfa_enroll_failed"
	MFAActionLoginSuccess         = "mfa_login_success"
	MFAActionLoginFailed          = "mfa_login_failed"
	MFAActionBackupCodeUsed       = "mfa_backup_code_used"
	MFAActionBackupCodeFailed     = "mfa_backup_code_failed"
	MFAActionBackupCodeRegen      = "mfa_backup_codes_regenerated"
	MFAActionDisabled             = "mfa_disabled"
	MFAActionRecoveryRequested    = "mfa_recovery_requested"
	MFAActionRecoveryUsed         = "mfa_recovery_used"
	MFAActionTrustedDeviceAdded   = "mfa_trusted_device_added"
	MFAActionTrustedDeviceRevoked = "mfa_trusted_device_revoked"
)

// EnrollMFAResponse represents the response when starting MFA enrollment
type EnrollMFAResponse struct {
	Secret      string   `json:"secret"`       // Base32 encoded secret for manual entry
	QRCodeURL   string   `json:"qr_code_url"`  // Data URL for QR code image
	BackupCodes []string `json:"backup_codes"` // Plain text backup codes (shown once)
}

// VerifyMFAEnrollmentRequest represents the request to verify MFA enrollment
type VerifyMFAEnrollmentRequest struct {
	Code string `json:"code" binding:"required,len=6,numeric"`
}

// VerifyMFALoginRequest represents the request to verify MFA during login
type VerifyMFALoginRequest struct {
	Code        string `json:"code" binding:"required"`
	TrustDevice *bool  `json:"trust_device,omitempty"`
}

// RegenerateBackupCodesRequest represents the request to regenerate backup codes
type RegenerateBackupCodesRequest struct {
	Code string `json:"code" binding:"required,len=6,numeric"`
}

// RegenerateBackupCodesResponse represents the response with new backup codes
type RegenerateBackupCodesResponse struct {
	BackupCodes []string `json:"backup_codes"`
}

// DisableMFARequest represents the request to disable MFA
type DisableMFARequest struct {
	Code string `json:"code" binding:"required"`
}

// MFAStatusResponse represents the current MFA status for a user
type MFAStatusResponse struct {
	Enabled              bool       `json:"enabled"`
	EnrolledAt           *time.Time `json:"enrolled_at,omitempty"`
	BackupCodesRemaining int        `json:"backup_codes_remaining"`
	TrustedDevicesCount  int        `json:"trusted_devices_count"`
	Required             bool       `json:"required"`              // Whether MFA is required for this user
	RequiredAt           *time.Time `json:"required_at,omitempty"` // When MFA became required
	GracePeriodEnd       *time.Time `json:"grace_period_end,omitempty"`
	InGracePeriod        bool       `json:"in_grace_period"` // Whether user is in grace period
}

// ============================================================================
// DMCA System Models
// ============================================================================

// DMCANotice represents a DMCA takedown notice submitted by a copyright holder
type DMCANotice struct {
	ID                         uuid.UUID  `json:"id" db:"id"`
	ComplainantName            string     `json:"complainant_name" db:"complainant_name"`
	ComplainantEmail           string     `json:"complainant_email" db:"complainant_email"`
	ComplainantAddress         string     `json:"complainant_address" db:"complainant_address"`
	ComplainantPhone           *string    `json:"complainant_phone,omitempty" db:"complainant_phone"`
	Relationship               string     `json:"relationship" db:"relationship"` // 'owner' or 'agent'
	CopyrightedWorkDescription string     `json:"copyrighted_work_description" db:"copyrighted_work_description"`
	InfringingURLs             []string   `json:"infringing_urls" db:"infringing_urls"` // PostgreSQL array
	GoodFaithStatement         bool       `json:"good_faith_statement" db:"good_faith_statement"`
	AccuracyStatement          bool       `json:"accuracy_statement" db:"accuracy_statement"`
	Signature                  string     `json:"signature" db:"signature"`
	SubmittedAt                time.Time  `json:"submitted_at" db:"submitted_at"`
	ReviewedAt                 *time.Time `json:"reviewed_at,omitempty" db:"reviewed_at"`
	ReviewedBy                 *uuid.UUID `json:"reviewed_by,omitempty" db:"reviewed_by"`
	Status                     string     `json:"status" db:"status"` // pending, valid, invalid, processed
	Notes                      *string    `json:"notes,omitempty" db:"notes"`
	IPAddress                  *string    `json:"ip_address,omitempty" db:"ip_address"`
	UserAgent                  *string    `json:"user_agent,omitempty" db:"user_agent"`
	CreatedAt                  time.Time  `json:"created_at" db:"created_at"`
	UpdatedAt                  time.Time  `json:"updated_at" db:"updated_at"`
}

// DMCACounterNotice represents a DMCA counter-notice submitted by a user
type DMCACounterNotice struct {
	ID                         uuid.UUID  `json:"id" db:"id"`
	DMCANoticeID               uuid.UUID  `json:"dmca_notice_id" db:"dmca_notice_id"`
	UserID                     *uuid.UUID `json:"user_id,omitempty" db:"user_id"`
	UserName                   string     `json:"user_name" db:"user_name"`
	UserEmail                  string     `json:"user_email" db:"user_email"`
	UserAddress                string     `json:"user_address" db:"user_address"`
	UserPhone                  *string    `json:"user_phone,omitempty" db:"user_phone"`
	RemovedMaterialURL         string     `json:"removed_material_url" db:"removed_material_url"`
	RemovedMaterialDescription *string    `json:"removed_material_description,omitempty" db:"removed_material_description"`
	GoodFaithStatement         bool       `json:"good_faith_statement" db:"good_faith_statement"`
	ConsentToJurisdiction      bool       `json:"consent_to_jurisdiction" db:"consent_to_jurisdiction"`
	ConsentToService           bool       `json:"consent_to_service" db:"consent_to_service"`
	Signature                  string     `json:"signature" db:"signature"`
	SubmittedAt                time.Time  `json:"submitted_at" db:"submitted_at"`
	ForwardedAt                *time.Time `json:"forwarded_at,omitempty" db:"forwarded_at"`
	WaitingPeriodEnds          *time.Time `json:"waiting_period_ends,omitempty" db:"waiting_period_ends"`
	Status                     string     `json:"status" db:"status"` // pending, forwarded, waiting, reinstated, rejected
	LawsuitFiled               bool       `json:"lawsuit_filed" db:"lawsuit_filed"`
	LawsuitFiledAt             *time.Time `json:"lawsuit_filed_at,omitempty" db:"lawsuit_filed_at"`
	Notes                      *string    `json:"notes,omitempty" db:"notes"`
	IPAddress                  *string    `json:"ip_address,omitempty" db:"ip_address"`
	UserAgent                  *string    `json:"user_agent,omitempty" db:"user_agent"`
	CreatedAt                  time.Time  `json:"created_at" db:"created_at"`
	UpdatedAt                  time.Time  `json:"updated_at" db:"updated_at"`
}

// DMCAStrike represents a copyright strike against a user
type DMCAStrike struct {
	ID            uuid.UUID  `json:"id" db:"id"`
	UserID        uuid.UUID  `json:"user_id" db:"user_id"`
	DMCANoticeID  uuid.UUID  `json:"dmca_notice_id" db:"dmca_notice_id"`
	ClipID        *uuid.UUID `json:"clip_id,omitempty" db:"clip_id"`
	SubmissionID  *uuid.UUID `json:"submission_id,omitempty" db:"submission_id"`
	StrikeNumber  int        `json:"strike_number" db:"strike_number"`
	IssuedAt      time.Time  `json:"issued_at" db:"issued_at"`
	ExpiresAt     time.Time  `json:"expires_at" db:"expires_at"`
	Status        string     `json:"status" db:"status"` // active, removed, expired
	RemovalReason *string    `json:"removal_reason,omitempty" db:"removal_reason"`
	RemovedAt     *time.Time `json:"removed_at,omitempty" db:"removed_at"`
	Notes         *string    `json:"notes,omitempty" db:"notes"`
	CreatedAt     time.Time  `json:"created_at" db:"created_at"`
	UpdatedAt     time.Time  `json:"updated_at" db:"updated_at"`
}

// SubmitDMCANoticeRequest represents the request to submit a DMCA takedown notice
type SubmitDMCANoticeRequest struct {
	ComplainantName            string   `json:"complainant_name" binding:"required,min=2,max=255"`
	ComplainantEmail           string   `json:"complainant_email" binding:"required,email,max=255"`
	ComplainantAddress         string   `json:"complainant_address" binding:"required,min=10"`
	ComplainantPhone           *string  `json:"complainant_phone,omitempty" binding:"omitempty,max=50"`
	Relationship               string   `json:"relationship" binding:"required,oneof=owner agent"`
	CopyrightedWorkDescription string   `json:"copyrighted_work_description" binding:"required,min=20"`
	InfringingURLs             []string `json:"infringing_urls" binding:"required,min=1,dive,url"`
	// Note: binding:"required" on booleans only validates presence, not truthiness
	// Service layer validates these are true (dmca_service.go lines 168-173)
	GoodFaithStatement bool   `json:"good_faith_statement" binding:"required"`
	AccuracyStatement  bool   `json:"accuracy_statement" binding:"required"`
	Signature          string `json:"signature" binding:"required,min=2,max=255"`
}

// SubmitDMCACounterNoticeRequest represents the request to submit a counter-notice
type SubmitDMCACounterNoticeRequest struct {
	DMCANoticeID               uuid.UUID `json:"dmca_notice_id" binding:"required"`
	UserName                   string    `json:"user_name" binding:"required,min=2,max=255"`
	UserEmail                  string    `json:"user_email" binding:"required,email,max=255"`
	UserAddress                string    `json:"user_address" binding:"required,min=10"`
	UserPhone                  *string   `json:"user_phone,omitempty" binding:"omitempty,max=50"`
	RemovedMaterialURL         string    `json:"removed_material_url" binding:"required,url"`
	RemovedMaterialDescription *string   `json:"removed_material_description,omitempty"`
	GoodFaithStatement         bool      `json:"good_faith_statement" binding:"required"`
	ConsentToJurisdiction      bool      `json:"consent_to_jurisdiction" binding:"required"`
	ConsentToService           bool      `json:"consent_to_service" binding:"required"`
	Signature                  string    `json:"signature" binding:"required,min=2,max=255"`
}

// UpdateDMCANoticeStatusRequest represents admin request to update notice status
type UpdateDMCANoticeStatusRequest struct {
	Status string  `json:"status" binding:"required,oneof=pending valid invalid processed"`
	Notes  *string `json:"notes,omitempty" binding:"omitempty,max=5000"`
}

// DMCANoticeListResponse represents a list of DMCA notices for admin panel
type DMCANoticeListResponse struct {
	Notices    []DMCANotice `json:"notices"`
	TotalCount int          `json:"total_count"`
	Page       int          `json:"page"`
	PageSize   int          `json:"page_size"`
}

// DMCAStrikeListResponse represents a list of strikes for a user
type DMCAStrikeListResponse struct {
	Strikes      []DMCAStrike `json:"strikes"`
	ActiveCount  int          `json:"active_count"`
	ExpiredCount int          `json:"expired_count"`
	RemovedCount int          `json:"removed_count"`
}

// DMCADashboardStats represents admin dashboard statistics
type DMCADashboardStats struct {
	PendingNotices               int `json:"pending_notices"`
	PendingCounterNotices        int `json:"pending_counter_notices"`
	ContentAwaitingRemoval       int `json:"content_awaiting_removal"`
	ContentAwaitingRestore       int `json:"content_awaiting_restore"`
	UsersWithActiveStrikes       int `json:"users_with_active_strikes"`
	UsersWithTwoStrikes          int `json:"users_with_two_strikes"`
	TotalTakedownsThisMonth      int `json:"total_takedowns_this_month"`
	TotalCounterNoticesThisMonth int `json:"total_counter_notices_this_month"`
}

// ModerationQueueItem represents an item in the moderation queue
type ModerationQueueItem struct {
	ID              uuid.UUID  `json:"id" db:"id"`
	ContentType     string     `json:"content_type" db:"content_type"`
	ContentID       uuid.UUID  `json:"content_id" db:"content_id"`
	Reason          string     `json:"reason" db:"reason"`
	Priority        int        `json:"priority" db:"priority"`
	Status          string     `json:"status" db:"status"`
	AssignedTo      *uuid.UUID `json:"assigned_to,omitempty" db:"assigned_to"`
	ReportedBy      []string   `json:"reported_by" db:"reported_by"` // PostgreSQL array
	ReportCount     int        `json:"report_count" db:"report_count"`
	AutoFlagged     bool       `json:"auto_flagged" db:"auto_flagged"`
	ConfidenceScore *float64   `json:"confidence_score,omitempty" db:"confidence_score"`
	CreatedAt       time.Time  `json:"created_at" db:"created_at"`
	ReviewedAt      *time.Time `json:"reviewed_at,omitempty" db:"reviewed_at"`
	ReviewedBy      *uuid.UUID `json:"reviewed_by,omitempty" db:"reviewed_by"`
	// Content will be joined separately
	Content interface{} `json:"content,omitempty" db:"-"`
}

// ModerationDecision represents a moderation decision audit entry
type ModerationDecision struct {
	ID          uuid.UUID `json:"id" db:"id"`
	QueueItemID uuid.UUID `json:"queue_item_id" db:"queue_item_id"`
	ModeratorID uuid.UUID `json:"moderator_id" db:"moderator_id"`
	Action      string    `json:"action" db:"action"`
	Reason      *string   `json:"reason,omitempty" db:"reason"`
	Metadata    *string   `json:"metadata,omitempty" db:"metadata"` // JSONB stored as string
	CreatedAt   time.Time `json:"created_at" db:"created_at"`
}

// ModerationQueueStats represents statistics about the moderation queue
type ModerationQueueStats struct {
	TotalPending      int            `json:"total_pending"`
	TotalApproved     int            `json:"total_approved"`
	TotalRejected     int            `json:"total_rejected"`
	TotalEscalated    int            `json:"total_escalated"`
	ByContentType     map[string]int `json:"by_content_type"`
	ByReason          map[string]int `json:"by_reason"`
	AutoFlaggedCount  int            `json:"auto_flagged_count"`
	UserReportedCount int            `json:"user_reported_count"`
	HighPriorityCount int            `json:"high_priority_count"`
	OldestPendingAge  *int           `json:"oldest_pending_age_hours,omitempty"`
}

// BulkModerationRequest represents a bulk moderation action request
type BulkModerationRequest struct {
	ItemIDs []string `json:"item_ids" binding:"required,min=1,max=100"`
	Action  string   `json:"action" binding:"required,oneof=approve reject escalate"`
	Reason  *string  `json:"reason,omitempty" binding:"omitempty,max=1000"`
}

// ModerationAppeal represents an appeal of a moderation decision
type ModerationAppeal struct {
	ID                 uuid.UUID  `json:"id" db:"id"`
	UserID             uuid.UUID  `json:"user_id" db:"user_id"`
	ModerationActionID uuid.UUID  `json:"moderation_action_id" db:"moderation_action_id"`
	Reason             string     `json:"reason" db:"reason"`
	Status             string     `json:"status" db:"status"` // pending, approved, rejected
	ResolvedBy         *uuid.UUID `json:"resolved_by,omitempty" db:"resolved_by"`
	Resolution         *string    `json:"resolution,omitempty" db:"resolution"`
	CreatedAt          time.Time  `json:"created_at" db:"created_at"`
	ResolvedAt         *time.Time `json:"resolved_at,omitempty" db:"resolved_at"`
}

// CreateAppealRequest represents the request to create an appeal
type CreateAppealRequest struct {
	ModerationActionID string `json:"moderation_action_id" binding:"required,uuid"`
	Reason             string `json:"reason" binding:"required,min=10,max=2000"`
}

// ResolveAppealRequest represents the request to resolve an appeal
type ResolveAppealRequest struct {
	Decision   string  `json:"decision" binding:"required,oneof=approve reject"`
	Resolution *string `json:"resolution,omitempty" binding:"omitempty,max=2000"`
}

// ModerationDecisionWithDetails represents a moderation decision with additional context
type ModerationDecisionWithDetails struct {
	ID            uuid.UUID `json:"id" db:"id"`
	QueueItemID   uuid.UUID `json:"queue_item_id" db:"queue_item_id"`
	ModeratorID   uuid.UUID `json:"moderator_id" db:"moderator_id"`
	ModeratorName string    `json:"moderator_name" db:"moderator_name"`
	Action        string    `json:"action" db:"action"`
	ContentType   string    `json:"content_type" db:"content_type"`
	ContentID     uuid.UUID `json:"content_id" db:"content_id"`
	Reason        *string   `json:"reason,omitempty" db:"reason"`
	Metadata      *string   `json:"metadata,omitempty" db:"metadata"`
	CreatedAt     time.Time `json:"created_at" db:"created_at"`
}

// TimeSeriesPoint represents a point in time series data
type TimeSeriesPoint struct {
	Date  string `json:"date"`
	Count int    `json:"count"`
}

// BannedUserStat represents statistics for a banned user
type BannedUserStat struct {
	UserID    string `json:"user_id"`
	Username  string `json:"username"`
	BanCount  int    `json:"ban_count"`
	LastBanAt string `json:"last_ban_at"`
}

// AppealStats represents appeal and reversal statistics
type AppealStats struct {
	TotalAppeals      int      `json:"total_appeals"`
	PendingAppeals    int      `json:"pending_appeals"`
	ApprovedAppeals   int      `json:"approved_appeals"`
	RejectedAppeals   int      `json:"rejected_appeals"`
	FalsePositiveRate *float64 `json:"false_positive_rate,omitempty"` // Percentage of approved appeals
}

// ModerationAnalytics represents analytics data for moderation actions
type ModerationAnalytics struct {
	TotalActions         int               `json:"total_actions"`
	ActionsByType        map[string]int    `json:"actions_by_type"`
	ActionsByModerator   map[string]int    `json:"actions_by_moderator"`
	ActionsOverTime      []TimeSeriesPoint `json:"actions_over_time"`
	ContentTypeBreakdown map[string]int    `json:"content_type_breakdown"`
	AverageResponseTime  *float64          `json:"average_response_time_minutes,omitempty"`
	BanReasons           map[string]int    `json:"ban_reasons"`
	MostBannedUsers      []BannedUserStat  `json:"most_banned_users"`
	Appeals              *AppealStats      `json:"appeals,omitempty"`
}

// ============================================================================
// Creator Verification System Models
// ============================================================================

// CreatorVerificationApplication represents a creator's verification application
type CreatorVerificationApplication struct {
	ID                 uuid.UUID              `json:"id" db:"id"`
	UserID             uuid.UUID              `json:"user_id" db:"user_id"`
	TwitchChannelURL   string                 `json:"twitch_channel_url" db:"twitch_channel_url"`
	FollowerCount      *int                   `json:"follower_count,omitempty" db:"follower_count"`
	SubscriberCount    *int                   `json:"subscriber_count,omitempty" db:"subscriber_count"`
	AvgViewers         *int                   `json:"avg_viewers,omitempty" db:"avg_viewers"`
	ContentDescription *string                `json:"content_description,omitempty" db:"content_description"`
	SocialMediaLinks   map[string]interface{} `json:"social_media_links,omitempty" db:"social_media_links"`
	Status             string                 `json:"status" db:"status"` // pending, approved, rejected
	Priority           int                    `json:"priority" db:"priority"`
	ReviewedBy         *uuid.UUID             `json:"reviewed_by,omitempty" db:"reviewed_by"`
	ReviewedAt         *time.Time             `json:"reviewed_at,omitempty" db:"reviewed_at"`
	ReviewerNotes      *string                `json:"reviewer_notes,omitempty" db:"reviewer_notes"`
	CreatedAt          time.Time              `json:"created_at" db:"created_at"`
	UpdatedAt          time.Time              `json:"updated_at" db:"updated_at"`
}

// CreatorVerificationApplicationWithUser includes user information
type CreatorVerificationApplicationWithUser struct {
	CreatorVerificationApplication
	User       *User `json:"user,omitempty"`
	ReviewedBy *User `json:"reviewed_by_user,omitempty"`
}

// CreatorVerificationDecision represents a verification decision audit entry
type CreatorVerificationDecision struct {
	ID            uuid.UUID              `json:"id" db:"id"`
	ApplicationID uuid.UUID              `json:"application_id" db:"application_id"`
	ReviewerID    uuid.UUID              `json:"reviewer_id" db:"reviewer_id"`
	Decision      string                 `json:"decision" db:"decision"` // approved, rejected
	Notes         *string                `json:"notes,omitempty" db:"notes"`
	Metadata      map[string]interface{} `json:"metadata,omitempty" db:"metadata"`
	CreatedAt     time.Time              `json:"created_at" db:"created_at"`
}

// VerificationApplicationStats represents statistics about verification applications
type VerificationApplicationStats struct {
	TotalPending  int `json:"total_pending"`
	TotalApproved int `json:"total_approved"`
	TotalRejected int `json:"total_rejected"`
	TotalVerified int `json:"total_verified"` // Total verified users
}

// CreateVerificationApplicationRequest represents the request to apply for verification
type CreateVerificationApplicationRequest struct {
	TwitchChannelURL   string            `json:"twitch_channel_url" binding:"required,url,max=500"`
	FollowerCount      *int              `json:"follower_count,omitempty" binding:"omitempty,min=0"`
	SubscriberCount    *int              `json:"subscriber_count,omitempty" binding:"omitempty,min=0"`
	AvgViewers         *int              `json:"avg_viewers,omitempty" binding:"omitempty,min=0"`
	ContentDescription *string           `json:"content_description,omitempty" binding:"omitempty,max=2000"`
	SocialMediaLinks   map[string]string `json:"social_media_links,omitempty"`
}

// ReviewVerificationApplicationRequest represents admin request to review an application
type ReviewVerificationApplicationRequest struct {
	Decision string  `json:"decision" binding:"required,oneof=approved rejected"`
	Notes    *string `json:"notes,omitempty" binding:"omitempty,max=2000"`
}

// Verification status constants
const (
	VerificationStatusPending  = "pending"
	VerificationStatusApproved = "approved"
	VerificationStatusRejected = "rejected"
)

// Verification decision constants
const (
	VerificationDecisionApproved = "approved"
	VerificationDecisionRejected = "rejected"
)

// VerificationAuditLog represents an audit log entry for verified users
type VerificationAuditLog struct {
	ID          uuid.UUID              `json:"id" db:"id"`
	UserID      uuid.UUID              `json:"user_id" db:"user_id"`
	AuditType   string                 `json:"audit_type" db:"audit_type"` // periodic_check, manual_review, abuse_detection
	Status      string                 `json:"status" db:"status"`         // passed, flagged, revoked
	Findings    map[string]interface{} `json:"findings,omitempty" db:"findings"`
	Notes       *string                `json:"notes,omitempty" db:"notes"`
	AuditedBy   *uuid.UUID             `json:"audited_by,omitempty" db:"audited_by"`
	ActionTaken *string                `json:"action_taken,omitempty" db:"action_taken"` // none, warning_sent, verification_revoked, further_review_required
	CreatedAt   time.Time              `json:"created_at" db:"created_at"`
}

// Audit type constants
const (
	AuditTypePeriodicCheck  = "periodic_check"
	AuditTypeManualReview   = "manual_review"
	AuditTypeAbuseDetection = "abuse_detection"
)

// Audit status constants
const (
	AuditStatusPassed  = "passed"
	AuditStatusFlagged = "flagged"
	AuditStatusRevoked = "revoked"
)

// Audit action constants
const (
	AuditActionNone                  = "none"
	AuditActionWarningSent           = "warning_sent"
	AuditActionVerificationRevoked   = "verification_revoked"
	AuditActionFurtherReviewRequired = "further_review_required"
)

// ChatChannel represents a chat channel
type ChatChannel struct {
	ID              uuid.UUID `json:"id" db:"id"`
	Name            string    `json:"name" db:"name"`
	Description     *string   `json:"description,omitempty" db:"description"`
	CreatorID       uuid.UUID `json:"creator_id" db:"creator_id"`
	ChannelType     string    `json:"channel_type" db:"channel_type"`
	IsActive        bool      `json:"is_active" db:"is_active"`
	MaxParticipants *int      `json:"max_participants,omitempty" db:"max_participants"`
	CreatedAt       time.Time `json:"created_at" db:"created_at"`
	UpdatedAt       time.Time `json:"updated_at" db:"updated_at"`
}

// ChatMessage represents a message in a chat channel
type ChatMessage struct {
	ID        uuid.UUID  `json:"id" db:"id"`
	ChannelID uuid.UUID  `json:"channel_id" db:"channel_id"`
	UserID    uuid.UUID  `json:"user_id" db:"user_id"`
	Content   string     `json:"content" db:"content"`
	IsDeleted bool       `json:"is_deleted" db:"is_deleted"`
	DeletedAt *time.Time `json:"deleted_at,omitempty" db:"deleted_at"`
	DeletedBy *uuid.UUID `json:"deleted_by,omitempty" db:"deleted_by"`
	CreatedAt time.Time  `json:"created_at" db:"created_at"`
	UpdatedAt time.Time  `json:"updated_at" db:"updated_at"`
	// Populated fields
	Username    string  `json:"username,omitempty" db:"username"`
	DisplayName string  `json:"display_name,omitempty" db:"display_name"`
	AvatarURL   *string `json:"avatar_url,omitempty" db:"avatar_url"`
}

// ChatBan represents a ban or mute for a user in a channel
type ChatBan struct {
	ID        uuid.UUID  `json:"id" db:"id"`
	ChannelID uuid.UUID  `json:"channel_id" db:"channel_id"`
	UserID    uuid.UUID  `json:"user_id" db:"user_id"`
	BannedBy  uuid.UUID  `json:"banned_by" db:"banned_by"`
	Reason    *string    `json:"reason,omitempty" db:"reason"`
	ExpiresAt *time.Time `json:"expires_at,omitempty" db:"expires_at"`
	CreatedAt time.Time  `json:"created_at" db:"created_at"`
	// Populated fields
	BannedByUsername string `json:"banned_by_username,omitempty" db:"banned_by_username"`
	TargetUsername   string `json:"target_username,omitempty" db:"target_username"`
}

// ChatModerationLog represents a log entry for moderation actions
type ChatModerationLog struct {
	ID           uuid.UUID  `json:"id" db:"id"`
	ChannelID    uuid.UUID  `json:"channel_id" db:"channel_id"`
	ModeratorID  uuid.UUID  `json:"moderator_id" db:"moderator_id"`
	TargetUserID *uuid.UUID `json:"target_user_id,omitempty" db:"target_user_id"`
	Action       string     `json:"action" db:"action"`
	Reason       *string    `json:"reason,omitempty" db:"reason"`
	Metadata     *string    `json:"metadata,omitempty" db:"metadata"` // JSONB stored as string
	CreatedAt    time.Time  `json:"created_at" db:"created_at"`
	// Populated fields
	ModeratorUsername string  `json:"moderator_username,omitempty" db:"moderator_username"`
	TargetUsername    *string `json:"target_username,omitempty" db:"target_username"`
}

// BanReasonTemplate represents a reusable ban reason template
type BanReasonTemplate struct {
	ID              uuid.UUID  `json:"id" db:"id"`
	Name            string     `json:"name" db:"name"`
	Reason          string     `json:"reason" db:"reason"`
	DurationSeconds *int       `json:"duration_seconds,omitempty" db:"duration_seconds"`
	IsDefault       bool       `json:"is_default" db:"is_default"`
	BroadcasterID   *string    `json:"broadcaster_id,omitempty" db:"broadcaster_id"`
	CreatedBy       *uuid.UUID `json:"created_by,omitempty" db:"created_by"`
	CreatedAt       time.Time  `json:"created_at" db:"created_at"`
	UpdatedAt       time.Time  `json:"updated_at" db:"updated_at"`
	UsageCount      int        `json:"usage_count" db:"usage_count"`
	LastUsedAt      *time.Time `json:"last_used_at,omitempty" db:"last_used_at"`
}

// CreateBanReasonTemplateRequest represents a request to create a template
type CreateBanReasonTemplateRequest struct {
	Name            string  `json:"name" binding:"required,max=100"`
	Reason          string  `json:"reason" binding:"required,max=1000"`
	DurationSeconds *int    `json:"duration_seconds,omitempty" binding:"omitempty,min=1,max=1209600"`
	BroadcasterID   *string `json:"broadcaster_id,omitempty" binding:"omitempty"`
}

// UpdateBanReasonTemplateRequest represents a request to update a template
type UpdateBanReasonTemplateRequest struct {
	Name            *string `json:"name,omitempty" binding:"omitempty,max=100"`
	Reason          *string `json:"reason,omitempty" binding:"omitempty,max=1000"`
	DurationSeconds *int    `json:"duration_seconds,omitempty" binding:"omitempty,min=1,max=1209600"`
}

// BanUserRequest represents a request to ban a user
type BanUserRequest struct {
	UserID          string     `json:"user_id" binding:"required,uuid"`
	Reason          string     `json:"reason" binding:"omitempty,max=1000"`
	DurationMinutes *int       `json:"duration_minutes,omitempty" binding:"omitempty,min=1"`
	TemplateID      *uuid.UUID `json:"template_id,omitempty" binding:"omitempty,uuid"`
}

// MuteUserRequest represents a request to mute a user
type MuteUserRequest struct {
	UserID          string `json:"user_id" binding:"required,uuid"`
	Reason          string `json:"reason" binding:"omitempty,max=1000"`
	DurationMinutes *int   `json:"duration_minutes,omitempty" binding:"omitempty,min=1"`
}

// TimeoutUserRequest represents a request to timeout a user
type TimeoutUserRequest struct {
	UserID          string `json:"user_id" binding:"required,uuid"`
	Reason          string `json:"reason" binding:"omitempty,max=1000"`
	DurationMinutes int    `json:"duration_minutes" binding:"required,min=1,max=43200"` // Max 30 days
}

// DeleteMessageRequest represents a request to delete a message
type DeleteMessageRequest struct {
	Reason string `json:"reason" binding:"omitempty,max=500"`
}

// CreateChannelRequest represents a request to create a chat channel
type CreateChannelRequest struct {
	Name            string  `json:"name" binding:"required,min=1,max=100"`
	Description     *string `json:"description,omitempty" binding:"omitempty,max=500"`
	ChannelType     string  `json:"channel_type" binding:"omitempty,oneof=public private"`
	MaxParticipants *int    `json:"max_participants,omitempty" binding:"omitempty,min=2"`
}

// UpdateChannelRequest represents a request to update a chat channel
type UpdateChannelRequest struct {
	Name            *string `json:"name,omitempty" binding:"omitempty,min=1,max=100"`
	Description     *string `json:"description,omitempty" binding:"omitempty,max=500"`
	IsActive        *bool   `json:"is_active,omitempty"`
	MaxParticipants *int    `json:"max_participants,omitempty" binding:"omitempty,min=2"`
}

// ChannelMember represents a member of a chat channel
type ChannelMember struct {
	ID        uuid.UUID  `json:"id" db:"id"`
	ChannelID uuid.UUID  `json:"channel_id" db:"channel_id"`
	UserID    uuid.UUID  `json:"user_id" db:"user_id"`
	Role      string     `json:"role" db:"role"`
	JoinedAt  time.Time  `json:"joined_at" db:"joined_at"`
	InvitedBy *uuid.UUID `json:"invited_by,omitempty" db:"invited_by"`
	// Populated fields
	Username    string  `json:"username,omitempty" db:"username"`
	DisplayName string  `json:"display_name,omitempty" db:"display_name"`
	AvatarURL   *string `json:"avatar_url,omitempty" db:"avatar_url"`
}

// AddChannelMemberRequest represents a request to add a member to a channel
type AddChannelMemberRequest struct {
	UserID string `json:"user_id" binding:"required,uuid"`
	Role   string `json:"role" binding:"omitempty,oneof=member moderator admin"`
}

// UpdateChannelMemberRequest represents a request to update a member's role
type UpdateChannelMemberRequest struct {
	Role string `json:"role" binding:"required,oneof=member moderator admin"`
}

// Chat moderation action constants
const (
	ChatActionBan     = "ban"
	ChatActionUnban   = "unban"
	ChatActionMute    = "mute"
	ChatActionUnmute  = "unmute"
	ChatActionTimeout = "timeout"
	ChatActionDelete  = "delete_message"
)

// UserFilterPreset represents a saved feed filter configuration
type UserFilterPreset struct {
	ID          uuid.UUID `json:"id" db:"id"`
	UserID      uuid.UUID `json:"user_id" db:"user_id"`
	Name        string    `json:"name" db:"name"`
	FiltersJSON string    `json:"filters_json" db:"filters_json"` // JSONB stored as string
	CreatedAt   time.Time `json:"created_at" db:"created_at"`
	UpdatedAt   time.Time `json:"updated_at" db:"updated_at"`
}

// FilterPresetFilters represents the filter configuration in a preset
type FilterPresetFilters struct {
	Games       []string `json:"games,omitempty"`
	Streamers   []string `json:"streamers,omitempty"`
	Tags        []string `json:"tags,omitempty"`
	ExcludeTags []string `json:"exclude_tags,omitempty"`
	DateFrom    *string  `json:"date_from,omitempty"`
	DateTo      *string  `json:"date_to,omitempty"`
	Sort        *string  `json:"sort,omitempty"`
	Language    *string  `json:"language,omitempty"`
	NSFWFilter  *bool    `json:"nsfw_filter,omitempty"`
}

// CreateFilterPresetRequest represents the request to create a filter preset
type CreateFilterPresetRequest struct {
	Name    string              `json:"name" binding:"required,min=1,max=100"`
	Filters FilterPresetFilters `json:"filters" binding:"required"`
}

// UpdateFilterPresetRequest represents the request to update a filter preset
type UpdateFilterPresetRequest struct {
	Name    *string              `json:"name,omitempty" binding:"omitempty,min=1,max=100"`
	Filters *FilterPresetFilters `json:"filters,omitempty"`
}

// ClipFiltersResponse represents enhanced clip feed response with filter metadata
type ClipFiltersResponse struct {
	Clips        interface{}   `json:"clips"`
	Pagination   interface{}   `json:"pagination"`
	FilterCounts *FilterCounts `json:"filter_counts,omitempty"`
}

// FilterCounts represents aggregated counts for filter options
type FilterCounts struct {
	Games     map[string]int `json:"games,omitempty"`
	Streamers map[string]int `json:"streamers,omitempty"`
	Tags      map[string]int `json:"tags,omitempty"`
}

// ============================================================================
// Recommendation System Models
// ============================================================================

// UserPreference represents a user's content preferences
type UserPreference struct {
	UserID                uuid.UUID   `json:"user_id" db:"user_id"`
	FavoriteGames         []string    `json:"favorite_games" db:"favorite_games"`
	FollowedStreamers     []string    `json:"followed_streamers" db:"followed_streamers"`
	PreferredCategories   []string    `json:"preferred_categories" db:"preferred_categories"`
	PreferredTags         []uuid.UUID `json:"preferred_tags" db:"preferred_tags"`
	OnboardingCompleted   bool        `json:"onboarding_completed" db:"onboarding_completed"`
	OnboardingCompletedAt *time.Time  `json:"onboarding_completed_at,omitempty" db:"onboarding_completed_at"`
	ColdStartSource       *string     `json:"cold_start_source,omitempty" db:"cold_start_source"` // 'onboarding', 'inferred', 'default'
	UpdatedAt             time.Time   `json:"updated_at" db:"updated_at"`
	CreatedAt             time.Time   `json:"created_at" db:"created_at"`
}

// UserClipInteraction represents a user's interaction with a clip
type UserClipInteraction struct {
	ID              uuid.UUID `json:"id" db:"id"`
	UserID          uuid.UUID `json:"user_id" db:"user_id"`
	ClipID          uuid.UUID `json:"clip_id" db:"clip_id"`
	InteractionType string    `json:"interaction_type" db:"interaction_type"` // 'view', 'like', 'share', 'dwell'
	DwellTime       *int      `json:"dwell_time,omitempty" db:"dwell_time"`
	Timestamp       time.Time `json:"timestamp" db:"timestamp"`
}

// ClipRecommendation represents a recommended clip with score and reason
type ClipRecommendation struct {
	Clip
	Score     float64 `json:"score" db:"score"`
	Reason    string  `json:"reason" db:"reason"`
	Algorithm string  `json:"algorithm" db:"algorithm"`
}

// RecommendationRequest represents a request for clip recommendations
type RecommendationRequest struct {
	UserID    uuid.UUID `json:"user_id" form:"user_id"`
	Limit     int       `json:"limit" form:"limit" binding:"omitempty,min=1,max=100"`
	Algorithm string    `json:"algorithm" form:"algorithm" binding:"omitempty,oneof=content collaborative hybrid trending"`
}

// RecommendationResponse represents the response with recommended clips
type RecommendationResponse struct {
	Recommendations []ClipRecommendation   `json:"recommendations"`
	Metadata        RecommendationMetadata `json:"metadata"`
}

// RecommendationMetadata contains metadata about the recommendation process
type RecommendationMetadata struct {
	AlgorithmUsed    string `json:"algorithm_used"`
	DiversityApplied bool   `json:"diversity_applied"`
	ColdStart        bool   `json:"cold_start"`
	CacheHit         bool   `json:"cache_hit"`
	ProcessingTimeMs int64  `json:"processing_time_ms"`
}

// RecommendationFeedback represents user feedback on a recommendation
type RecommendationFeedback struct {
	ID           uuid.UUID `json:"id" db:"id"`
	UserID       uuid.UUID `json:"user_id" db:"user_id"`
	ClipID       uuid.UUID `json:"clip_id" db:"clip_id"`
	FeedbackType string    `json:"feedback_type" db:"feedback_type"` // 'positive', 'negative'
	Algorithm    string    `json:"algorithm" db:"algorithm"`
	Score        float64   `json:"score" db:"score"`
	CreatedAt    time.Time `json:"created_at" db:"created_at"`
}

// SubmitFeedbackRequest represents a request to submit feedback on a recommendation
type SubmitFeedbackRequest struct {
	ClipID       uuid.UUID `json:"clip_id" binding:"required"`
	FeedbackType string    `json:"feedback_type" binding:"required,oneof=positive negative"`
	Algorithm    *string   `json:"algorithm,omitempty"`
	Score        *float64  `json:"score,omitempty"`
}

// UpdatePreferencesRequest represents a request to update user preferences
type UpdatePreferencesRequest struct {
	FavoriteGames       *[]string    `json:"favorite_games,omitempty"`
	FollowedStreamers   *[]string    `json:"followed_streamers,omitempty"`
	PreferredCategories *[]string    `json:"preferred_categories,omitempty"`
	PreferredTags       *[]uuid.UUID `json:"preferred_tags,omitempty"`
}

// OnboardingPreferencesRequest represents initial onboarding preferences
// At least one preference type (games, streamers, categories, or tags) must be provided
type OnboardingPreferencesRequest struct {
	FavoriteGames       []string    `json:"favorite_games,omitempty" binding:"omitempty,max=10,dive,required"`
	FollowedStreamers   []string    `json:"followed_streamers,omitempty" binding:"omitempty,max=10,dive,required"`
	PreferredCategories []string    `json:"preferred_categories,omitempty" binding:"omitempty,max=5,dive,required"`
	PreferredTags       []uuid.UUID `json:"preferred_tags,omitempty" binding:"omitempty,max=10"`
}

// Validate ensures at least one preference type is provided
func (r *OnboardingPreferencesRequest) Validate() error {
	if len(r.FavoriteGames) == 0 && len(r.FollowedStreamers) == 0 &&
		len(r.PreferredCategories) == 0 && len(r.PreferredTags) == 0 {
		return fmt.Errorf("at least one preference type must be provided")
	}
	return nil
}

// Interaction type constants
const (
	InteractionTypeView    = "view"
	InteractionTypeLike    = "like"
	InteractionTypeShare   = "share"
	InteractionTypeDwell   = "dwell"
	InteractionTypeDislike = "dislike"
)

// Recommendation algorithm constants
const (
	AlgorithmContent       = "content"
	AlgorithmCollaborative = "collaborative"
	AlgorithmHybrid        = "hybrid"
	AlgorithmTrending      = "trending"
)

// ============================================================================
// Feed Events System Models (Analytics & Performance Monitoring)
// ============================================================================

// Event represents a feed interaction event for analytics
type Event struct {
	ID         uuid.UUID              `json:"id" db:"id"`
	EventType  string                 `json:"event_type" db:"event_type"`
	UserID     *uuid.UUID             `json:"user_id,omitempty" db:"user_id"`
	SessionID  string                 `json:"session_id" db:"session_id"`
	Timestamp  time.Time              `json:"timestamp" db:"timestamp"`
	Properties map[string]interface{} `json:"properties,omitempty" db:"properties"`
	CreatedAt  time.Time              `json:"created_at" db:"created_at"`
}

// HourlyMetric represents aggregated hourly event metrics
type HourlyMetric struct {
	Hour           time.Time `json:"hour" db:"hour"`
	EventType      string    `json:"event_type" db:"event_type"`
	Count          int64     `json:"count" db:"count"`
	UniqueUsers    int64     `json:"unique_users" db:"unique_users"`
	UniqueSessions int64     `json:"unique_sessions" db:"unique_sessions"`
}

// TrackEventRequest represents a request to track an event
type TrackEventRequest struct {
	EventType  string                 `json:"event_type" binding:"required"`
	Properties map[string]interface{} `json:"properties,omitempty"`
}

// BatchEventsRequest represents a request to track multiple events
type BatchEventsRequest struct {
	Events []TrackEventRequest `json:"events" binding:"required,min=1,max=100"`
}

// Event type constants for feed analytics
const (
	EventFeedViewed            = "feed_viewed"
	EventFilterApplied         = "filter_applied"
	EventSortChanged           = "sort_changed"
	EventRecommendationClicked = "recommendation_clicked"
	EventFeedEngaged           = "feed_engaged"
)

// ============================================================================
// Playlist System Models
// ============================================================================

// Playlist represents a user-created collection of clips
type Playlist struct {
	ID             uuid.UUID  `json:"id" db:"id"`
	UserID         uuid.UUID  `json:"user_id" db:"user_id"`
	Title          string     `json:"title" db:"title"`
	Description    *string    `json:"description,omitempty" db:"description"`
	CoverURL       *string    `json:"cover_url,omitempty" db:"cover_url"`
	Visibility     string     `json:"visibility" db:"visibility"` // private, public, unlisted
	ShareToken     *string    `json:"share_token,omitempty" db:"share_token"`
	ViewCount      int        `json:"view_count" db:"view_count"`
	ShareCount     int        `json:"share_count" db:"share_count"`
	LikeCount      int        `json:"like_count" db:"like_count"`
	FollowerCount  int        `json:"follower_count" db:"follower_count"`
	BookmarkCount  int        `json:"bookmark_count" db:"bookmark_count"`
	IsCurated      bool       `json:"is_curated" db:"is_curated"`
	IsFeatured     bool       `json:"is_featured" db:"is_featured"`
	DisplayOrder   int        `json:"display_order" db:"display_order"`
	ScriptID       *uuid.UUID `json:"script_id,omitempty" db:"script_id"`
	Slug           *string    `json:"slug,omitempty" db:"slug"`
	CreatedAt      time.Time  `json:"created_at" db:"created_at"`
	UpdatedAt      time.Time  `json:"updated_at" db:"updated_at"`
	DeletedAt      *time.Time `json:"deleted_at,omitempty" db:"deleted_at"`
}

// PlaylistItem represents a clip in a playlist
type PlaylistItem struct {
	ID         int       `json:"id" db:"id"`
	PlaylistID uuid.UUID `json:"playlist_id" db:"playlist_id"`
	ClipID     uuid.UUID `json:"clip_id" db:"clip_id"`
	OrderIndex int       `json:"order_index" db:"order_index"`
	AddedAt    time.Time `json:"added_at" db:"added_at"`
}

// PlaylistLike represents a user's like on a playlist
type PlaylistLike struct {
	ID         uuid.UUID `json:"id" db:"id"`
	UserID     uuid.UUID `json:"user_id" db:"user_id"`
	PlaylistID uuid.UUID `json:"playlist_id" db:"playlist_id"`
	CreatedAt  time.Time `json:"created_at" db:"created_at"`
}

// PlaylistFollow represents a user following a playlist
type PlaylistFollow struct {
	ID         uuid.UUID `json:"id" db:"id"`
	UserID     uuid.UUID `json:"user_id" db:"user_id"`
	PlaylistID uuid.UUID `json:"playlist_id" db:"playlist_id"`
	FollowedAt time.Time `json:"followed_at" db:"followed_at"`
}

// PlaylistBookmark represents a user bookmarking a playlist
type PlaylistBookmark struct {
	ID           uuid.UUID `json:"id" db:"id"`
	UserID       uuid.UUID `json:"user_id" db:"user_id"`
	PlaylistID   uuid.UUID `json:"playlist_id" db:"playlist_id"`
	BookmarkedAt time.Time `json:"bookmarked_at" db:"bookmarked_at"`
}

// PlaylistScript represents an automated playlist definition
type PlaylistScript struct {
	ID                      uuid.UUID  `json:"id" db:"id"`
	Name                    string     `json:"name" db:"name"`
	Description             *string    `json:"description,omitempty" db:"description"`
	Sort                    string     `json:"sort" db:"sort"`
	Timeframe               *string    `json:"timeframe,omitempty" db:"timeframe"`
	ClipLimit               int        `json:"clip_limit" db:"clip_limit"`
	Visibility              string     `json:"visibility" db:"visibility"`
	IsActive                bool       `json:"is_active" db:"is_active"`
	Schedule                string     `json:"schedule" db:"schedule"`
	Strategy                string     `json:"strategy" db:"strategy"`
	GameID                  *string    `json:"game_id,omitempty" db:"game_id"`
	GameIDs                 []string   `json:"game_ids,omitempty" db:"game_ids"`
	BroadcasterID           *string    `json:"broadcaster_id,omitempty" db:"broadcaster_id"`
	Tag                     *string    `json:"tag,omitempty" db:"tag"`
	ExcludeTags             []string   `json:"exclude_tags,omitempty" db:"exclude_tags"`
	Language                *string    `json:"language,omitempty" db:"language"`
	MinVoteScore            *int       `json:"min_vote_score,omitempty" db:"min_vote_score"`
	MinViewCount            *int       `json:"min_view_count,omitempty" db:"min_view_count"`
	ExcludeNSFW             bool       `json:"exclude_nsfw" db:"exclude_nsfw"`
	Top10kStreamers         bool       `json:"top_10k_streamers" db:"top_10k_streamers"`
	SeedClipID              *uuid.UUID `json:"seed_clip_id,omitempty" db:"seed_clip_id"`
	RetentionDays           int        `json:"retention_days" db:"retention_days"`
	TitleTemplate           *string    `json:"title_template,omitempty" db:"title_template"`
	CreatedBy               *uuid.UUID `json:"created_by,omitempty" db:"created_by"`
	CreatedAt               time.Time  `json:"created_at" db:"created_at"`
	UpdatedAt               time.Time  `json:"updated_at" db:"updated_at"`
	LastRunAt               *time.Time `json:"last_run_at,omitempty" db:"last_run_at"`
	LastGeneratedPlaylistID *uuid.UUID `json:"last_generated_playlist_id,omitempty" db:"last_generated_playlist_id"`
}

// GeneratedPlaylist represents a playlist generated from a script
type GeneratedPlaylist struct {
	ID          uuid.UUID `json:"id" db:"id"`
	ScriptID    uuid.UUID `json:"script_id" db:"script_id"`
	PlaylistID  uuid.UUID `json:"playlist_id" db:"playlist_id"`
	GeneratedAt time.Time `json:"generated_at" db:"generated_at"`
}

// PlaylistListItem represents a playlist in list views with clip count
// PlaylistListItem represents a playlist in list views with clip count
type PlaylistListItem struct {
	Playlist
	ClipCount           int    `json:"clip_count" db:"clip_count"`
	HasProcessingClips  bool   `json:"has_processing_clips" db:"has_processing_clips"`
	IsLiked             bool   `json:"is_liked"`
	IsBookmarked        bool   `json:"is_bookmarked"`
	PreviewClips        []Clip `json:"preview_clips,omitempty"`
}

// PlaylistWithClips represents a playlist with its clips
type PlaylistWithClips struct {
	Playlist
	ClipCount    int               `json:"clip_count" db:"clip_count"`
	Clips        []PlaylistClipRef `json:"clips,omitempty"`
	PreviewClips []Clip            `json:"preview_clips,omitempty"`
	IsLiked      bool              `json:"is_liked"`
	IsFollowed   bool              `json:"is_followed"`
	IsBookmarked bool              `json:"is_bookmarked"`
	Creator      *User             `json:"creator,omitempty"`
	CurrentUserPermission *string  `json:"current_user_permission,omitempty"`
}

// PlaylistClipRef represents a clip reference in a playlist with ordering
type PlaylistClipRef struct {
	Clip
	OrderIndex int `json:"order" db:"order_index"`
}

// CreatePlaylistRequest represents the request to create a playlist
type CreatePlaylistRequest struct {
	Title       string  `json:"title" binding:"required,min=1,max=100"`
	Description *string `json:"description,omitempty" binding:"omitempty,max=500"`
	CoverURL    *string `json:"cover_url,omitempty" binding:"omitempty,max=255,url"`
	Visibility  *string `json:"visibility,omitempty" binding:"omitempty,oneof=private public unlisted"`
}

// UpdatePlaylistRequest represents the request to update a playlist
type UpdatePlaylistRequest struct {
	Title       *string `json:"title,omitempty" binding:"omitempty,min=1,max=100"`
	Description *string `json:"description,omitempty" binding:"omitempty,max=500"`
	CoverURL    *string `json:"cover_url,omitempty" binding:"omitempty,max=255,url"`
	Visibility  *string `json:"visibility,omitempty" binding:"omitempty,oneof=private public unlisted"`
}

// CopyPlaylistRequest represents the request to copy a playlist
type CopyPlaylistRequest struct {
	Title       *string `json:"title,omitempty" binding:"omitempty,min=1,max=100"`
	Description *string `json:"description,omitempty" binding:"omitempty,max=500"`
	CoverURL    *string `json:"cover_url,omitempty" binding:"omitempty,max=255,url"`
	Visibility  *string `json:"visibility,omitempty" binding:"omitempty,oneof=private public unlisted"`
}

// CreatePlaylistScriptRequest represents the request to create a playlist script
type CreatePlaylistScriptRequest struct {
	Name            string   `json:"name" binding:"required,min=1,max=100"`
	Description     *string  `json:"description,omitempty" binding:"omitempty,max=500"`
	Sort            string   `json:"sort" binding:"required,oneof=hot new top rising discussed trending popular"`
	Timeframe       *string  `json:"timeframe,omitempty" binding:"omitempty,oneof=hour day week month year"`
	ClipLimit       int      `json:"clip_limit" binding:"required,min=1,max=100"`
	Visibility      *string  `json:"visibility,omitempty" binding:"omitempty,oneof=private public unlisted"`
	IsActive        *bool    `json:"is_active,omitempty"`
	Schedule        *string  `json:"schedule,omitempty" binding:"omitempty,oneof=manual hourly daily weekly monthly"`
	Strategy        *string  `json:"strategy,omitempty" binding:"omitempty,oneof=standard sleeper_hits viral_velocity community_favorites deep_cuts fresh_faces one_per_creator similar_vibes cross_game_hits controversial binge_worthy rising_stars twitch_top_game twitch_top_broadcaster twitch_trending twitch_discovery"`
	GameID          *string  `json:"game_id,omitempty" binding:"omitempty,max=50"`
	GameIDs         []string `json:"game_ids,omitempty"`
	BroadcasterID   *string  `json:"broadcaster_id,omitempty" binding:"omitempty,max=50"`
	Tag             *string  `json:"tag,omitempty" binding:"omitempty,max=100"`
	ExcludeTags     []string `json:"exclude_tags,omitempty"`
	Language        *string  `json:"language,omitempty" binding:"omitempty,max=10"`
	MinVoteScore    *int     `json:"min_vote_score,omitempty" binding:"omitempty,min=0"`
	MinViewCount    *int     `json:"min_view_count,omitempty" binding:"omitempty,min=0"`
	ExcludeNSFW     *bool    `json:"exclude_nsfw,omitempty"`
	Top10kStreamers *bool    `json:"top_10k_streamers,omitempty"`
	SeedClipID      *string  `json:"seed_clip_id,omitempty" binding:"omitempty,uuid"`
	RetentionDays   *int     `json:"retention_days,omitempty" binding:"omitempty,min=1,max=365"`
	TitleTemplate   *string  `json:"title_template,omitempty" binding:"omitempty,max=200"`
}

// UpdatePlaylistScriptRequest represents the request to update a playlist script
type UpdatePlaylistScriptRequest struct {
	Name            *string  `json:"name,omitempty" binding:"omitempty,min=1,max=100"`
	Description     *string  `json:"description,omitempty" binding:"omitempty,max=500"`
	Sort            *string  `json:"sort,omitempty" binding:"omitempty,oneof=hot new top rising discussed trending popular"`
	Timeframe       *string  `json:"timeframe,omitempty" binding:"omitempty,oneof=hour day week month year"`
	ClipLimit       *int     `json:"clip_limit,omitempty" binding:"omitempty,min=1,max=100"`
	Visibility      *string  `json:"visibility,omitempty" binding:"omitempty,oneof=private public unlisted"`
	IsActive        *bool    `json:"is_active,omitempty"`
	Schedule        *string  `json:"schedule,omitempty" binding:"omitempty,oneof=manual hourly daily weekly monthly"`
	Strategy        *string  `json:"strategy,omitempty" binding:"omitempty,oneof=standard sleeper_hits viral_velocity community_favorites deep_cuts fresh_faces one_per_creator similar_vibes cross_game_hits controversial binge_worthy rising_stars twitch_top_game twitch_top_broadcaster twitch_trending twitch_discovery"`
	GameID          *string  `json:"game_id,omitempty" binding:"omitempty,max=50"`
	GameIDs         []string `json:"game_ids,omitempty"`
	BroadcasterID   *string  `json:"broadcaster_id,omitempty" binding:"omitempty,max=50"`
	Tag             *string  `json:"tag,omitempty" binding:"omitempty,max=100"`
	ExcludeTags     []string `json:"exclude_tags,omitempty"`
	Language        *string  `json:"language,omitempty" binding:"omitempty,max=10"`
	MinVoteScore    *int     `json:"min_vote_score,omitempty" binding:"omitempty,min=0"`
	MinViewCount    *int     `json:"min_view_count,omitempty" binding:"omitempty,min=0"`
	ExcludeNSFW     *bool    `json:"exclude_nsfw,omitempty"`
	Top10kStreamers *bool    `json:"top_10k_streamers,omitempty"`
	SeedClipID      *string  `json:"seed_clip_id,omitempty" binding:"omitempty,uuid"`
	RetentionDays   *int     `json:"retention_days,omitempty" binding:"omitempty,min=1,max=365"`
	TitleTemplate   *string  `json:"title_template,omitempty" binding:"omitempty,max=200"`
}

// AddClipsToPlaylistRequest represents the request to add clips to a playlist
type AddClipsToPlaylistRequest struct {
	ClipIDs []uuid.UUID `json:"clip_ids" binding:"required,min=1,max=100"`
}

// ReorderPlaylistClipsRequest represents the request to reorder clips in a playlist
type ReorderPlaylistClipsRequest struct {
	ClipIDs []uuid.UUID `json:"clip_ids" binding:"required,min=1"`
}

// Playlist visibility constants
const (
	PlaylistVisibilityPrivate  = "private"
	PlaylistVisibilityPublic   = "public"
	PlaylistVisibilityUnlisted = "unlisted"
)

// Playlist permission constants
const (
	PlaylistPermissionView  = "view"
	PlaylistPermissionEdit  = "edit"
	PlaylistPermissionAdmin = "admin"
)

// PlaylistCollaborator represents a collaborator on a playlist
type PlaylistCollaborator struct {
	ID         uuid.UUID  `json:"id" db:"id"`
	PlaylistID uuid.UUID  `json:"playlist_id" db:"playlist_id"`
	UserID     uuid.UUID  `json:"user_id" db:"user_id"`
	User       *User      `json:"user,omitempty"`
	Permission string     `json:"permission" db:"permission"` // view, edit, admin
	InvitedBy  *uuid.UUID `json:"invited_by,omitempty" db:"invited_by"`
	InvitedAt  time.Time  `json:"invited_at" db:"invited_at"`
	CreatedAt  time.Time  `json:"created_at" db:"created_at"`
	UpdatedAt  time.Time  `json:"updated_at" db:"updated_at"`
}

// PlaylistShare represents a share event for analytics
type PlaylistShare struct {
	ID         uuid.UUID `json:"id" db:"id"`
	PlaylistID uuid.UUID `json:"playlist_id" db:"playlist_id"`
	Platform   *string   `json:"platform,omitempty" db:"platform"` // twitter, facebook, discord, embed
	Referrer   *string   `json:"referrer,omitempty" db:"referrer"`
	SharedAt   time.Time `json:"shared_at" db:"shared_at"`
}

// GetShareLinkResponse represents the response for share link generation
type GetShareLinkResponse struct {
	ShareURL  string `json:"share_url"`
	EmbedCode string `json:"embed_code"`
}

// AddCollaboratorRequest represents the request to add a collaborator
type AddCollaboratorRequest struct {
	UserID     string `json:"user_id" binding:"required,uuid"`
	Permission string `json:"permission" binding:"required,oneof=view edit admin"`
}

// UpdateCollaboratorRequest represents the request to update a collaborator's permission
type UpdateCollaboratorRequest struct {
	Permission string `json:"permission" binding:"required,oneof=view edit admin"`
}

// TrackShareRequest represents the request to track a share event
type TrackShareRequest struct {
	Platform string  `json:"platform" binding:"required,oneof=twitter facebook discord embed link"`
	Referrer *string `json:"referrer,omitempty"`
}

// ============================================================================
// Queue System Models
// ============================================================================

// QueueItem represents a clip in a user's playback queue
type QueueItem struct {
	ID        uuid.UUID  `json:"id" db:"id"`
	UserID    uuid.UUID  `json:"user_id" db:"user_id"`
	ClipID    uuid.UUID  `json:"clip_id" db:"clip_id"`
	Position  int        `json:"position" db:"position"`
	AddedAt   time.Time  `json:"added_at" db:"added_at"`
	PlayedAt  *time.Time `json:"played_at,omitempty" db:"played_at"`
	CreatedAt time.Time  `json:"created_at" db:"created_at"`
	UpdatedAt time.Time  `json:"updated_at" db:"updated_at"`
}

// QueueItemWithClip represents a queue item with clip information
type QueueItemWithClip struct {
	QueueItem
	Clip *Clip `json:"clip,omitempty"`
}

// Queue represents a user's complete queue
type Queue struct {
	Items    []QueueItemWithClip `json:"items"`
	Total    int                 `json:"total"`
	NextClip *Clip               `json:"next_clip,omitempty"`
}

// AddToQueueRequest represents the request to add a clip to the queue
type AddToQueueRequest struct {
	ClipID string `json:"clip_id" binding:"required,uuid"`
	AtEnd  *bool  `json:"at_end,omitempty"` // true = add to end (default), false = add to top
}

// ReorderQueueRequest represents the request to reorder a queue item
type ReorderQueueRequest struct {
	ItemID      string `json:"item_id" binding:"required,uuid"`
	NewPosition int    `json:"new_position" binding:"required,min=1"`
}

// ConvertQueueToPlaylistRequest represents a request to convert queue to playlist
type ConvertQueueToPlaylistRequest struct {
	Title        string  `json:"title" binding:"required,min=1,max=255"`
	Description  *string `json:"description"`
	OnlyUnplayed bool    `json:"only_unplayed"` // If true, only convert unplayed items
	ClearQueue   bool    `json:"clear_queue"`   // If true, clear queue after conversion
}

// WatchHistoryEntry represents a watch history entry for a clip
type WatchHistoryEntry struct {
	ID              uuid.UUID `json:"id" db:"id"`
	UserID          uuid.UUID `json:"user_id" db:"user_id"`
	ClipID          uuid.UUID `json:"clip_id" db:"clip_id"`
	Clip            *Clip     `json:"clip,omitempty"`
	ProgressSeconds int       `json:"progress_seconds" db:"progress_seconds"`
	DurationSeconds int       `json:"duration_seconds" db:"duration_seconds"`
	ProgressPercent float64   `json:"progress_percent"`
	Completed       bool      `json:"completed" db:"completed"`
	SessionID       string    `json:"session_id" db:"session_id"`
	WatchedAt       time.Time `json:"watched_at" db:"watched_at"`
	CreatedAt       time.Time `json:"created_at" db:"created_at"`
	UpdatedAt       time.Time `json:"updated_at" db:"updated_at"`
}

// RecordWatchProgressRequest represents the request to record watch progress
type RecordWatchProgressRequest struct {
	ClipID          string `json:"clip_id" binding:"required,uuid"`
	ProgressSeconds int    `json:"progress_seconds" binding:"required,min=0"`
	DurationSeconds int    `json:"duration_seconds" binding:"required,min=1"`
	SessionID       string `json:"session_id" binding:"required,min=1,max=100"`
}

// WatchHistoryResponse represents the response containing watch history
type WatchHistoryResponse struct {
	History []WatchHistoryEntry `json:"history"`
	Total   int                 `json:"total"`
}

// ResumePositionResponse represents the response for resume position
type ResumePositionResponse struct {
	HasProgress     bool `json:"has_progress"`
	ProgressSeconds int  `json:"progress_seconds"`
	Completed       bool `json:"completed"`
}

// Watch Party System Models

// WatchParty represents a synchronized video watching session
type WatchParty struct {
	ID                     uuid.UUID               `json:"id" db:"id"`
	HostUserID             uuid.UUID               `json:"host_user_id" db:"host_user_id"`
	Title                  string                  `json:"title" db:"title"`
	PlaylistID             *uuid.UUID              `json:"playlist_id,omitempty" db:"playlist_id"`
	CurrentClipID          *uuid.UUID              `json:"current_clip_id,omitempty" db:"current_clip_id"`
	CurrentPositionSeconds int                     `json:"current_position_seconds" db:"current_position_seconds"`
	IsPlaying              bool                    `json:"is_playing" db:"is_playing"`
	Visibility             string                  `json:"visibility" db:"visibility"`
	Password               *string                 `json:"-" db:"password"` // Password hash, never sent to client
	InviteCode             string                  `json:"invite_code" db:"invite_code"`
	MaxParticipants        int                     `json:"max_participants" db:"max_participants"`
	CreatedAt              time.Time               `json:"created_at" db:"created_at"`
	StartedAt              *time.Time              `json:"started_at,omitempty" db:"started_at"`
	EndedAt                *time.Time              `json:"ended_at,omitempty" db:"ended_at"`
	Participants           []WatchPartyParticipant `json:"participants,omitempty" db:"-"`
	ActiveParticipantCount int                     `json:"active_participant_count,omitempty" db:"-"` // Computed field for discovery endpoints
}

// WatchPartyParticipant represents a user participating in a watch party
type WatchPartyParticipant struct {
	ID           uuid.UUID  `json:"id" db:"id"`
	PartyID      uuid.UUID  `json:"party_id" db:"party_id"`
	UserID       uuid.UUID  `json:"user_id" db:"user_id"`
	User         *User      `json:"user,omitempty" db:"-"`
	Role         string     `json:"role" db:"role"`
	JoinedAt     time.Time  `json:"joined_at" db:"joined_at"`
	LeftAt       *time.Time `json:"left_at,omitempty" db:"left_at"`
	LastSyncAt   *time.Time `json:"last_sync_at,omitempty" db:"last_sync_at"`
	SyncOffsetMS int        `json:"sync_offset_ms" db:"sync_offset_ms"`
}

// CreateWatchPartyRequest represents a request to create a watch party
type CreateWatchPartyRequest struct {
	Title           string     `json:"title" binding:"required,min=1,max=200"`
	PlaylistID      *uuid.UUID `json:"playlist_id,omitempty" binding:"omitempty,uuid"`
	Visibility      string     `json:"visibility,omitempty" binding:"omitempty,oneof=private public friends invite"`
	Password        *string    `json:"password,omitempty" binding:"omitempty,min=4,max=100"`
	MaxParticipants *int       `json:"max_participants,omitempty" binding:"omitempty,min=2,max=1000"`
}

// UpdateWatchPartySettingsRequest represents a request to update watch party settings
type UpdateWatchPartySettingsRequest struct {
	Visibility *string `json:"visibility,omitempty" binding:"omitempty,oneof=private public friends invite"`
	Password   *string `json:"password,omitempty" binding:"omitempty,min=4,max=100"`
}

// JoinWatchPartyResponse represents the response when joining a party
type JoinWatchPartyResponse struct {
	Party     WatchParty `json:"party"`
	InviteURL string     `json:"invite_url"`
}

// JoinWatchPartyRequest represents a request to join a watch party
type JoinWatchPartyRequest struct {
	Password *string `json:"password,omitempty"`
}

// WatchPartyHistoryEntry represents a past watch party
type WatchPartyHistoryEntry struct {
	ID               uuid.UUID  `json:"id" db:"id"`
	HostUserID       uuid.UUID  `json:"host_user_id" db:"host_user_id"`
	Title            string     `json:"title" db:"title"`
	Visibility       string     `json:"visibility" db:"visibility"`
	ParticipantCount int        `json:"participant_count" db:"participant_count"`
	CreatedAt        time.Time  `json:"created_at" db:"created_at"`
	StartedAt        *time.Time `json:"started_at,omitempty" db:"started_at"`
	EndedAt          *time.Time `json:"ended_at,omitempty" db:"ended_at"`
	Duration         *int       `json:"duration_seconds,omitempty"` // Calculated field
}

// WatchPartyCommand represents a command from client to server
type WatchPartyCommand struct {
	Type           string     `json:"type"` // play, pause, seek, skip, sync-request, chat, reaction, typing
	PartyID        string     `json:"party_id"`
	Position       *int       `json:"position,omitempty"`        // for seek (in seconds)
	ClipID         *uuid.UUID `json:"clip_id,omitempty"`         // for skip
	Message        string     `json:"message,omitempty"`         // for chat
	Emoji          string     `json:"emoji,omitempty"`           // for reaction
	VideoTimestamp *float64   `json:"video_timestamp,omitempty"` // for reaction
	IsTyping       bool       `json:"is_typing,omitempty"`       // for typing indicator
	Timestamp      int64      `json:"timestamp"`                 // client timestamp (Unix seconds)
}

// WatchPartySyncEvent represents a sync event from server to clients
type WatchPartySyncEvent struct {
	Type            string                     `json:"type"` // sync, play, pause, seek, skip, participant-joined, participant-left, chat_message, reaction, typing
	PartyID         string                     `json:"party_id"`
	ClipID          *uuid.UUID                 `json:"clip_id,omitempty"`
	Position        int                        `json:"position"` // playback position in seconds
	IsPlaying       bool                       `json:"is_playing"`
	ServerTimestamp int64                      `json:"server_timestamp"` // server timestamp (Unix seconds)
	Participant     *WatchPartyParticipantInfo `json:"participant,omitempty"`
	ChatMessage     *WatchPartyMessage         `json:"chat_message,omitempty"` // for chat_message events
	Reaction        *WatchPartyReaction        `json:"reaction,omitempty"`     // for reaction events
	UserID          *uuid.UUID                 `json:"user_id,omitempty"`      // for typing events
	IsTyping        bool                       `json:"is_typing,omitempty"`    // for typing events
}

// WatchPartyParticipantInfo represents basic participant info in events
type WatchPartyParticipantInfo struct {
	UserID      uuid.UUID `json:"user_id"`
	DisplayName string    `json:"display_name"`
	AvatarURL   *string   `json:"avatar_url,omitempty"`
	Role        string    `json:"role"`
}

// WatchPartyMessage represents a chat message in a watch party
type WatchPartyMessage struct {
	ID           uuid.UUID `json:"id" db:"id"`
	WatchPartyID uuid.UUID `json:"watch_party_id" db:"watch_party_id"`
	UserID       uuid.UUID `json:"user_id" db:"user_id"`
	Username     string    `json:"username,omitempty" db:"-"`
	DisplayName  string    `json:"display_name,omitempty" db:"-"`
	AvatarURL    *string   `json:"avatar_url,omitempty" db:"-"`
	Message      string    `json:"message" db:"message"`
	CreatedAt    time.Time `json:"created_at" db:"created_at"`
}

// WatchPartyReaction represents an emoji reaction in a watch party
type WatchPartyReaction struct {
	ID             uuid.UUID `json:"id" db:"id"`
	WatchPartyID   uuid.UUID `json:"watch_party_id" db:"watch_party_id"`
	UserID         uuid.UUID `json:"user_id" db:"user_id"`
	Username       string    `json:"username,omitempty" db:"-"`
	Emoji          string    `json:"emoji" db:"emoji"`
	VideoTimestamp *float64  `json:"video_timestamp,omitempty" db:"video_timestamp"`
	CreatedAt      time.Time `json:"created_at" db:"created_at"`
}

// SendMessageRequest represents a request to send a chat message
type SendMessageRequest struct {
	Message string `json:"message" binding:"required,min=1,max=1000"`
}

// SendReactionRequest represents a request to send an emoji reaction
type SendReactionRequest struct {
	Emoji          string   `json:"emoji" binding:"required,min=1,max=10"`
	VideoTimestamp *float64 `json:"video_timestamp,omitempty"`
}

// TwitchAuth represents Twitch OAuth authentication data
// Note: This model is for internal use only. The TwitchAuthStatusResponse
// is used for public API responses and excludes sensitive fields like tokens and scopes.
type TwitchAuth struct {
	UserID         uuid.UUID `json:"user_id" db:"user_id"`
	TwitchUserID   string    `json:"twitch_user_id" db:"twitch_user_id"`
	TwitchUsername string    `json:"twitch_username" db:"twitch_username"`
	AccessToken    string    `json:"access_token" db:"access_token"`   // Never expose in API responses
	RefreshToken   string    `json:"refresh_token" db:"refresh_token"` // Never expose in API responses
	Scopes         string    `json:"scopes" db:"scopes"`               // Space-separated list of granted scopes (internal use)
	ExpiresAt      time.Time `json:"expires_at" db:"expires_at"`
	CreatedAt      time.Time `json:"created_at" db:"created_at"`
	UpdatedAt      time.Time `json:"updated_at" db:"updated_at"`
}

// TwitchAuthStatusResponse represents the response for Twitch auth status
type TwitchAuthStatusResponse struct {
	Authenticated  bool       `json:"authenticated"`
	Connected      bool       `json:"connected"`
	TwitchUserID   *string    `json:"twitch_user_id,omitempty"`
	TwitchUsername *string    `json:"twitch_username,omitempty"`
	ExpiresAt      *time.Time `json:"expires_at,omitempty"`
}

// ============================================================================
// CDN and Mirror Models
// ============================================================================

// ClipMirror represents a mirrored clip in a specific region
type ClipMirror struct {
	ID              uuid.UUID  `json:"id" db:"id"`
	ClipID          uuid.UUID  `json:"clip_id" db:"clip_id"`
	Region          string     `json:"region" db:"region"`
	MirrorURL       string     `json:"mirror_url" db:"mirror_url"`
	Status          string     `json:"status" db:"status"` // pending, active, failed, expired
	StorageProvider string     `json:"storage_provider" db:"storage_provider"`
	SizeBytes       *int64     `json:"size_bytes,omitempty" db:"size_bytes"`
	CreatedAt       time.Time  `json:"created_at" db:"created_at"`
	LastAccessedAt  *time.Time `json:"last_accessed_at,omitempty" db:"last_accessed_at"`
	AccessCount     int        `json:"access_count" db:"access_count"`
	ExpiresAt       *time.Time `json:"expires_at,omitempty" db:"expires_at"`
	FailureReason   *string    `json:"failure_reason,omitempty" db:"failure_reason"`
}

// MirrorMetrics represents metrics for mirror performance
type MirrorMetrics struct {
	ID          uuid.UUID              `json:"id" db:"id"`
	ClipID      uuid.UUID              `json:"clip_id" db:"clip_id"`
	Region      string                 `json:"region" db:"region"`
	MetricType  string                 `json:"metric_type" db:"metric_type"` // access, failover, bandwidth, cost
	MetricValue float64                `json:"metric_value" db:"metric_value"`
	RecordedAt  time.Time              `json:"recorded_at" db:"recorded_at"`
	Metadata    map[string]interface{} `json:"metadata,omitempty" db:"metadata"`
}

// CDNConfiguration represents CDN provider configuration
type CDNConfiguration struct {
	ID        uuid.UUID              `json:"id" db:"id"`
	Provider  string                 `json:"provider" db:"provider"` // cloudflare, bunny, aws-cloudfront
	Region    *string                `json:"region,omitempty" db:"region"`
	IsActive  bool                   `json:"is_active" db:"is_active"`
	Priority  int                    `json:"priority" db:"priority"`
	Config    map[string]interface{} `json:"config" db:"config"`
	CreatedAt time.Time              `json:"created_at" db:"created_at"`
	UpdatedAt time.Time              `json:"updated_at" db:"updated_at"`
}

// CDNMetrics represents CDN performance and cost metrics
type CDNMetrics struct {
	ID          uuid.UUID              `json:"id" db:"id"`
	Provider    string                 `json:"provider" db:"provider"`
	Region      *string                `json:"region,omitempty" db:"region"`
	MetricType  string                 `json:"metric_type" db:"metric_type"` // latency, bandwidth, cost, cache_hit_rate, requests
	MetricValue float64                `json:"metric_value" db:"metric_value"`
	RecordedAt  time.Time              `json:"recorded_at" db:"recorded_at"`
	Metadata    map[string]interface{} `json:"metadata,omitempty" db:"metadata"`
}

// MirrorStatus constants
const (
	MirrorStatusPending = "pending"
	MirrorStatusActive  = "active"
	MirrorStatusFailed  = "failed"
	MirrorStatusExpired = "expired"
)

// CDN Provider constants
// Note: These constants use CamelCase (matching Go conventions), but their string values
// use lowercase-with-hyphens (matching configuration and documentation conventions)
const (
	CDNProviderCloudflare    = "cloudflare"     // Cloudflare CDN
	CDNProviderBunny         = "bunny"          // Bunny.net CDN
	CDNProviderAWSCloudFront = "aws-cloudfront" // AWS CloudFront CDN
)

// Mirror metric types
const (
	MirrorMetricTypeAccess    = "access"
	MirrorMetricTypeFailover  = "failover"
	MirrorMetricTypeBandwidth = "bandwidth"
	MirrorMetricTypeCost      = "cost"
)

// CDN metric types
const (
	CDNMetricTypeLatency      = "latency"
	CDNMetricTypeBandwidth    = "bandwidth"
	CDNMetricTypeCost         = "cost"
	CDNMetricTypeCacheHitRate = "cache_hit_rate"
	CDNMetricTypeRequests     = "requests"
)

// ServiceStatus represents the current status of a service
type ServiceStatus struct {
	ID             uuid.UUID              `json:"id" db:"id"`
	ServiceName    string                 `json:"service_name" db:"service_name"`
	Status         string                 `json:"status" db:"status"`
	StatusMessage  *string                `json:"status_message,omitempty" db:"status_message"`
	LastCheckAt    time.Time              `json:"last_check_at" db:"last_check_at"`
	ResponseTimeMs *int                   `json:"response_time_ms,omitempty" db:"response_time_ms"`
	ErrorRate      *float64               `json:"error_rate,omitempty" db:"error_rate"`
	Metadata       map[string]interface{} `json:"metadata,omitempty" db:"metadata"`
	CreatedAt      time.Time              `json:"created_at" db:"created_at"`
	UpdatedAt      time.Time              `json:"updated_at" db:"updated_at"`
}

// StatusHistory represents historical status data
type StatusHistory struct {
	ID             int64                  `json:"id" db:"id"`
	ServiceName    string                 `json:"service_name" db:"service_name"`
	Status         string                 `json:"status" db:"status"`
	ResponseTimeMs *int                   `json:"response_time_ms,omitempty" db:"response_time_ms"`
	ErrorRate      *float64               `json:"error_rate,omitempty" db:"error_rate"`
	CheckedAt      time.Time              `json:"checked_at" db:"checked_at"`
	Metadata       map[string]interface{} `json:"metadata,omitempty" db:"metadata"`
}

// StatusIncident represents a service incident
type StatusIncident struct {
	ID          uuid.UUID  `json:"id" db:"id"`
	ServiceName string     `json:"service_name" db:"service_name"`
	Title       string     `json:"title" db:"title"`
	Description *string    `json:"description,omitempty" db:"description"`
	Severity    string     `json:"severity" db:"severity"`
	Status      string     `json:"status" db:"status"`
	StartedAt   time.Time  `json:"started_at" db:"started_at"`
	ResolvedAt  *time.Time `json:"resolved_at,omitempty" db:"resolved_at"`
	CreatedBy   *uuid.UUID `json:"created_by,omitempty" db:"created_by"`
	CreatedAt   time.Time  `json:"created_at" db:"created_at"`
	UpdatedAt   time.Time  `json:"updated_at" db:"updated_at"`
}

// StatusIncidentUpdate represents an update to an incident
type StatusIncidentUpdate struct {
	ID         uuid.UUID  `json:"id" db:"id"`
	IncidentID uuid.UUID  `json:"incident_id" db:"incident_id"`
	Status     string     `json:"status" db:"status"`
	Message    string     `json:"message" db:"message"`
	CreatedBy  *uuid.UUID `json:"created_by,omitempty" db:"created_by"`
	CreatedAt  time.Time  `json:"created_at" db:"created_at"`
}

// StatusSubscription represents a user's subscription to status updates
type StatusSubscription struct {
	ID               uuid.UUID `json:"id" db:"id"`
	UserID           uuid.UUID `json:"user_id" db:"user_id"`
	ServiceName      *string   `json:"service_name,omitempty" db:"service_name"`
	NotificationType string    `json:"notification_type" db:"notification_type"`
	WebhookURL       *string   `json:"webhook_url,omitempty" db:"webhook_url"`
	IsActive         bool      `json:"is_active" db:"is_active"`
	CreatedAt        time.Time `json:"created_at" db:"created_at"`
	UpdatedAt        time.Time `json:"updated_at" db:"updated_at"`
}

// Service status constants
const (
	ServiceStatusHealthy   = "healthy"
	ServiceStatusDegraded  = "degraded"
	ServiceStatusUnhealthy = "unhealthy"
)

// Incident severity constants
const (
	IncidentSeverityCritical    = "critical"
	IncidentSeverityMajor       = "major"
	IncidentSeverityMinor       = "minor"
	IncidentSeverityMaintenance = "maintenance"
)

// Incident status constants
const (
	IncidentStatusInvestigating = "investigating"
	IncidentStatusIdentified    = "identified"
	IncidentStatusMonitoring    = "monitoring"
	IncidentStatusResolved      = "resolved"
)

// Notification type constants
const (
	NotificationTypeEmail   = "email"
	NotificationTypeWebhook = "webhook"
	NotificationTypeAll     = "all"
)

// CreateIncidentRequest represents the request to create an incident
type CreateIncidentRequest struct {
	ServiceName string  `json:"service_name" binding:"required,max=100"`
	Title       string  `json:"title" binding:"required,min=3,max=255"`
	Description *string `json:"description,omitempty" binding:"omitempty,max=5000"`
	Severity    string  `json:"severity" binding:"required,oneof=critical major minor maintenance"`
}

// UpdateIncidentRequest represents the request to update an incident
type UpdateIncidentRequest struct {
	Status  *string `json:"status,omitempty" binding:"omitempty,oneof=investigating identified monitoring resolved"`
	Message string  `json:"message" binding:"required,min=1,max=5000"`
}

// CreateSubscriptionRequest represents the request to create a status subscription
type CreateSubscriptionRequest struct {
	ServiceName      *string `json:"service_name,omitempty" binding:"omitempty,max=100"`
	NotificationType string  `json:"notification_type" binding:"required,oneof=email webhook all"`
	WebhookURL       *string `json:"webhook_url,omitempty" binding:"omitempty,url,max=2048"`
}
