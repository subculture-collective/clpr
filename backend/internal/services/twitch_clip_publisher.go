package services

import (
	"context"
	"time"

	"git.subcult.tv/subculture-collective/clpr/internal/models"
	"git.subcult.tv/subculture-collective/clpr/internal/repository"
	internalutils "git.subcult.tv/subculture-collective/clpr/internal/utils"
	"git.subcult.tv/subculture-collective/clpr/pkg/twitch"
	"github.com/google/uuid"
)

type PublishDisposition string

const (
	PublishCreated PublishDisposition = "created"
	PublishUpdated PublishDisposition = "updated"
)

type PublishResult struct {
	Clip        *models.Clip
	Disposition PublishDisposition
}

// TwitchClipPublisher owns idempotent publication of automated Twitch clips
// into the main feed persistence model.
type TwitchClipPublisher struct {
	clips *repository.ClipRepository
}

func NewTwitchClipPublisher(clips *repository.ClipRepository) *TwitchClipPublisher {
	return &TwitchClipPublisher{clips: clips}
}

func (p *TwitchClipPublisher) Publish(ctx context.Context, twitchClip *twitch.Clip) (*PublishResult, error) {
	clip := clipModelFromTwitch(twitchClip)
	published, created, err := p.clips.PublishAutomatedClip(ctx, clip)
	if err != nil {
		return nil, err
	}

	disposition := PublishUpdated
	if created {
		disposition = PublishCreated
	}
	return &PublishResult{Clip: published, Disposition: disposition}, nil
}

func clipModelFromTwitch(twitchClip *twitch.Clip) *models.Clip {
	return &models.Clip{
		ID:              uuid.New(),
		TwitchClipID:    twitchClip.ID,
		TwitchClipURL:   twitchClip.URL,
		EmbedURL:        twitchClip.EmbedURL,
		Title:           twitchClip.Title,
		CreatorName:     twitchClip.CreatorName,
		CreatorID:       internalutils.StringPtr(twitchClip.CreatorID),
		BroadcasterName: twitchClip.BroadcasterName,
		BroadcasterID:   internalutils.StringPtr(twitchClip.BroadcasterID),
		GameID:          internalutils.StringPtr(twitchClip.GameID),
		Language:        internalutils.StringPtr(twitchClip.Language),
		ThumbnailURL:    internalutils.StringPtr(twitchClip.ThumbnailURL),
		Duration:        internalutils.Float64Ptr(twitchClip.Duration),
		ViewCount:       twitchClip.ViewCount,
		CreatedAt:       twitchClip.CreatedAt,
		ImportedAt:      time.Now(),
	}
}
