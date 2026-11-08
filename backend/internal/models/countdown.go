package models

import (
	"time"

	"github.com/google/uuid"
)

// Countdown edition statuses. Selection writes "selected" or "failed"; the
// fetch, render and publish stages (#77 to #79) move an edition forward.
const (
	CountdownEditionSelected  = "selected"
	CountdownEditionFetching  = "fetching"
	CountdownEditionFetched   = "fetched"
	CountdownEditionRendering = "rendering"
	CountdownEditionRendered  = "rendered"
	CountdownEditionPublished = "published"
	CountdownEditionFailed    = "failed"
)

// Countdown exclusion reasons and sources.
const (
	CountdownExclusionRequest  = "request"
	CountdownExclusionDMCA     = "dmca"
	CountdownExclusionTakedown = "takedown"

	CountdownExclusionSourceAdmin       = "admin"
	CountdownExclusionSourceSelfService = "self_service"
)

// CountdownSeries is one daily ranked list: overall, one Twitch category
// (GameID) or one language.
type CountdownSeries struct {
	ID                uuid.UUID `json:"id" db:"id"`
	Slug              string    `json:"slug" db:"slug"`
	Name              string    `json:"name" db:"name"`
	GameID            *string   `json:"game_id,omitempty" db:"game_id"`
	Language          *string   `json:"language,omitempty" db:"language"`
	Size              int       `json:"size" db:"size"`
	WindowHours       int       `json:"window_hours" db:"window_hours"`
	PerBroadcasterCap int       `json:"per_broadcaster_cap" db:"per_broadcaster_cap"`
	SelectionHourUTC  int       `json:"selection_hour_utc" db:"selection_hour_utc"`
	Priority          int       `json:"priority" db:"priority"`
	Enabled           bool      `json:"enabled" db:"enabled"`
	ShortsCount       int       `json:"shorts_count" db:"shorts_count"`
	CreatedAt         time.Time `json:"created_at" db:"created_at"`
	UpdatedAt         time.Time `json:"updated_at" db:"updated_at"`
}

// CountdownSeriesInput creates or replaces a series configuration.
type CountdownSeriesInput struct {
	Slug              string  `json:"slug"`
	Name              string  `json:"name"`
	GameID            *string `json:"game_id"`
	Language          *string `json:"language"`
	Size              int     `json:"size"`
	WindowHours       int     `json:"window_hours"`
	PerBroadcasterCap int     `json:"per_broadcaster_cap"`
	SelectionHourUTC  int     `json:"selection_hour_utc"`
	Priority          int     `json:"priority"`
	Enabled           bool    `json:"enabled"`
	ShortsCount       int     `json:"shorts_count"`
}

// CountdownEdition is one day's selection for a series.
type CountdownEdition struct {
	ID            uuid.UUID              `json:"id" db:"id"`
	SeriesID      uuid.UUID              `json:"series_id" db:"series_id"`
	SeriesSlug    string                 `json:"series_slug" db:"series_slug"`
	EditionDate   string                 `json:"edition_date" db:"edition_date"`
	Status        string                 `json:"status" db:"status"`
	WindowStart   time.Time              `json:"window_start" db:"window_start"`
	WindowEnd     time.Time              `json:"window_end" db:"window_end"`
	FailureReason *string                `json:"failure_reason,omitempty" db:"failure_reason"`
	CreatedAt     time.Time              `json:"created_at" db:"created_at"`
	UpdatedAt     time.Time              `json:"updated_at" db:"updated_at"`
	Clips         []CountdownEditionClip `json:"clips"`
}

// CountdownEditionClip is one ranked entry; position 1 is the top clip.
type CountdownEditionClip struct {
	Position        int       `json:"position" db:"position"`
	ClipID          uuid.UUID `json:"clip_id" db:"clip_id"`
	ViewVelocity    float64   `json:"view_velocity" db:"view_velocity"`
	Title           string    `json:"title" db:"title"`
	BroadcasterName string    `json:"broadcaster_name" db:"broadcaster_name"`
	BroadcasterID   *string   `json:"broadcaster_id,omitempty" db:"broadcaster_id"`
	TwitchClipID    string    `json:"twitch_clip_id" db:"twitch_clip_id"`
	TwitchClipURL   string    `json:"twitch_clip_url" db:"twitch_clip_url"`
	Duration        *float64  `json:"duration,omitempty" db:"duration"`
}

// CountdownEditionMedia is one fetched clip file on the render host. FileName
// is relative to the media directory.
type CountdownEditionMedia struct {
	EditionID       uuid.UUID  `json:"edition_id" db:"edition_id"`
	ClipID          uuid.UUID  `json:"clip_id" db:"clip_id"`
	Source          string     `json:"source" db:"source"`
	FileName        string     `json:"file_name" db:"file_name"`
	SHA256          string     `json:"sha256" db:"sha256"`
	Bytes           int64      `json:"bytes" db:"bytes"`
	DurationSeconds float64    `json:"duration_seconds" db:"duration_seconds"`
	Width           int        `json:"width" db:"width"`
	Height          int        `json:"height" db:"height"`
	FetchedAt       time.Time  `json:"fetched_at" db:"fetched_at"`
	DeletedAt       *time.Time `json:"deleted_at,omitempty" db:"deleted_at"`
}

// Countdown render kinds.
const (
	CountdownRenderCountdown = "countdown"
	CountdownRenderShort     = "short"
	CountdownRenderThumbnail = "thumbnail"
)

// CountdownRender is one rendered file on the render host. FileName is
// relative to the render directory; ClipID is set for Shorts only.
type CountdownRender struct {
	ID              uuid.UUID  `json:"id" db:"id"`
	EditionID       uuid.UUID  `json:"edition_id" db:"edition_id"`
	Kind            string     `json:"kind" db:"kind"`
	ClipID          *uuid.UUID `json:"clip_id,omitempty" db:"clip_id"`
	FileName        string     `json:"file_name" db:"file_name"`
	SHA256          string     `json:"sha256" db:"sha256"`
	Bytes           int64      `json:"bytes" db:"bytes"`
	DurationSeconds *float64   `json:"duration_seconds,omitempty" db:"duration_seconds"`
	RenderedAt      time.Time  `json:"rendered_at" db:"rendered_at"`
	DeletedAt       *time.Time `json:"deleted_at,omitempty" db:"deleted_at"`
}

// Countdown media sources: Twitch's official Helix download endpoint, used
// when the broadcaster granted channel:manage:clips, or the playback files
// Twitch serves to its own player, fetched with yt-dlp (ADR-002).
const (
	CountdownMediaSourceHelix  = "helix"
	CountdownMediaSourcePlayer = "player"
)

// CountdownExclusion leaves a broadcaster or one clip out of future editions.
type CountdownExclusion struct {
	ID            uuid.UUID  `json:"id" db:"id"`
	BroadcasterID *string    `json:"broadcaster_id,omitempty" db:"broadcaster_id"`
	ClipID        *uuid.UUID `json:"clip_id,omitempty" db:"clip_id"`
	Reason        string     `json:"reason" db:"reason"`
	Source        string     `json:"source" db:"source"`
	Note          *string    `json:"note,omitempty" db:"note"`
	CreatedBy     *uuid.UUID `json:"created_by,omitempty" db:"created_by"`
	CreatedAt     time.Time  `json:"created_at" db:"created_at"`
}

// CountdownExclusionInput is the admin request to add an exclusion.
type CountdownExclusionInput struct {
	BroadcasterID *string    `json:"broadcaster_id"`
	ClipID        *uuid.UUID `json:"clip_id"`
	Reason        string     `json:"reason"`
	Note          *string    `json:"note"`
}

// PublicCountdownEdition is a rendered or published edition as the public
// countdown pages show it. Size is the number of ranked clips in the edition;
// Clips leaves out clips removed or excluded since, so positions can skip.
type PublicCountdownEdition struct {
	EditionDate string                `json:"edition_date"`
	SeriesSlug  string                `json:"series_slug"`
	SeriesName  string                `json:"series_name"`
	Size        int                   `json:"size"`
	Clips       []PublicCountdownClip `json:"clips"`
}

// PublicCountdownClip is one ranked entry of a public edition.
type PublicCountdownClip struct {
	Position        int       `json:"position"`
	ClipID          uuid.UUID `json:"clip_id"`
	Title           string    `json:"title"`
	BroadcasterName string    `json:"broadcaster_name"`
	BroadcasterID   *string   `json:"broadcaster_id,omitempty"`
	TwitchClipID    string    `json:"twitch_clip_id"`
	TwitchClipURL   string    `json:"twitch_clip_url"`
	ThumbnailURL    *string   `json:"thumbnail_url,omitempty"`
	Duration        *float64  `json:"duration,omitempty"`
}

// CountdownFeature records that a clip placed in a public edition.
type CountdownFeature struct {
	EditionDate string    `json:"edition_date"`
	SeriesSlug  string    `json:"series_slug"`
	SeriesName  string    `json:"series_name"`
	Position    int       `json:"position"`
	ClipID      uuid.UUID `json:"clip_id"`
	Title       string    `json:"title"`
}

// CountdownEditionKey names one public edition page.
type CountdownEditionKey struct {
	EditionDate string
	SeriesSlug  string
}
