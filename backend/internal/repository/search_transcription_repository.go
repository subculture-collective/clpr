package repository

import (
	"context"
	"encoding/json"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
)

// TranscriptSourceSearch marks transcripts the render host made for search
// alone (ADR-003).
const TranscriptSourceSearch = "search_render_whisper"

// Search transcription attempt outcomes.
const (
	SearchTranscriptionTranscribed = "transcribed"
	SearchTranscriptionNoSpeech    = "no_speech"
	SearchTranscriptionFailed      = "failed"
)

// SearchTranscriptionRepository selects public clips to transcribe for
// search and stores the results (ADR-003).
type SearchTranscriptionRepository struct {
	pool *pgxpool.Pool
}

// NewSearchTranscriptionRepository returns a repository on the pool.
func NewSearchTranscriptionRepository(pool *pgxpool.Pool) *SearchTranscriptionRepository {
	return &SearchTranscriptionRepository{pool: pool}
}

// SearchTranscriptionQuery bounds which clips are transcribed.
type SearchTranscriptionQuery struct {
	MinViews int
	// CreatedAfter skips older clips.
	CreatedAfter time.Time
	// RetryBefore: a failed clip is tried again once its last attempt is
	// older than this, up to MaxAttempts attempts.
	RetryBefore time.Time
	MaxAttempts int
	Limit       int
}

// SearchTranscriptionClip is a clip to download and transcribe.
type SearchTranscriptionClip struct {
	ClipID        uuid.UUID
	TwitchClipID  string
	TwitchClipURL string
	BroadcasterID *string
}

// NextClipsToTranscribe returns public Twitch clips above the view threshold
// that have no transcript, most viewed first. It leaves out clips that are
// removed, hidden, NSFW or taken down, clips and broadcasters excluded on
// request, clips without English speech, and failed clips until they are
// due for a retry.
func (r *SearchTranscriptionRepository) NextClipsToTranscribe(ctx context.Context, q SearchTranscriptionQuery) ([]SearchTranscriptionClip, error) {
	rows, err := r.pool.Query(ctx, `
		SELECT c.id, c.twitch_clip_id, c.twitch_clip_url, c.broadcaster_id
		FROM clips c
		WHERE c.source_type = 'twitch'
		  AND c.is_removed = false
		  AND c.is_hidden = false
		  AND c.is_nsfw = false
		  AND COALESCE(c.dmca_removed, false) = false
		  AND (c.language IS NULL OR c.language = 'en')
		  AND c.view_count >= $1
		  AND c.created_at >= $2
		  AND NOT EXISTS (SELECT 1 FROM clip_transcripts t WHERE t.clip_id = c.id)
		  AND NOT EXISTS (
		      SELECT 1 FROM countdown_exclusions x
		      WHERE x.clip_id = c.id OR (c.broadcaster_id IS NOT NULL AND x.broadcaster_id = c.broadcaster_id)
		  )
		  AND NOT EXISTS (
		      SELECT 1 FROM search_transcription_attempts a
		      WHERE a.clip_id = c.id
		        AND (a.outcome <> 'failed' OR a.attempts >= $4 OR a.attempted_at > $3)
		  )
		ORDER BY c.view_count DESC, c.id
		LIMIT $5`, q.MinViews, q.CreatedAfter, q.RetryBefore, q.MaxAttempts, q.Limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := []SearchTranscriptionClip{}
	for rows.Next() {
		var c SearchTranscriptionClip
		if err := rows.Scan(&c.ClipID, &c.TwitchClipID, &c.TwitchClipURL, &c.BroadcasterID); err != nil {
			return nil, err
		}
		out = append(out, c)
	}
	return out, rows.Err()
}

// SaveSearchTranscript stores a search transcript unless the clip got a
// transcript from another source meanwhile. It reports whether it stored it.
func (r *SearchTranscriptionRepository) SaveSearchTranscript(ctx context.Context, clipID uuid.UUID, language, fullText string, segments json.RawMessage) (bool, error) {
	tag, err := r.pool.Exec(ctx, `
		INSERT INTO clip_transcripts (clip_id, language, full_text, segments, source)
		VALUES ($1, NULLIF($2, ''), $3, $4, 'search_render_whisper')
		ON CONFLICT (clip_id) DO NOTHING`, clipID, language, fullText, segments)
	if err != nil {
		return false, err
	}
	return tag.RowsAffected() > 0, nil
}

// RecordSearchTranscriptionAttempt records the outcome of one attempt.
func (r *SearchTranscriptionRepository) RecordSearchTranscriptionAttempt(ctx context.Context, clipID uuid.UUID, outcome string) error {
	_, err := r.pool.Exec(ctx, `
		INSERT INTO search_transcription_attempts (clip_id, outcome)
		VALUES ($1, $2)
		ON CONFLICT (clip_id) DO UPDATE SET
			outcome = EXCLUDED.outcome,
			attempts = search_transcription_attempts.attempts + 1,
			attempted_at = NOW()`, clipID, outcome)
	return err
}

// SearchClipsAttemptedOn returns how many clips were attempted for search on
// the UTC day of t, for the daily limit.
func (r *SearchTranscriptionRepository) SearchClipsAttemptedOn(ctx context.Context, t time.Time) (int, error) {
	var clips int
	err := r.pool.QueryRow(ctx, `
		SELECT COALESCE(SUM(clips), 0) FROM transcription_usage_daily
		WHERE day = $1::date AND source = 'search_render_whisper'`, t.UTC().Format("2006-01-02")).Scan(&clips)
	return clips, err
}

// ClearExcludedTranscripts empties the render host's transcripts of clips
// and broadcasters excluded on request. Emptying, rather than deleting,
// changes the transcript's version, so the search index drops the text on
// its next copy. It returns how many it emptied.
func (r *SearchTranscriptionRepository) ClearExcludedTranscripts(ctx context.Context) (int64, error) {
	tag, err := r.pool.Exec(ctx, `
		UPDATE clip_transcripts t
		SET full_text = '', segments = '[]'::jsonb, updated_at = NOW()
		FROM clips c
		WHERE c.id = t.clip_id
		  AND t.source IN ('countdown_render_whisper', 'search_render_whisper')
		  AND t.full_text <> ''
		  AND EXISTS (
		      SELECT 1 FROM countdown_exclusions x
		      WHERE x.clip_id = c.id OR (c.broadcaster_id IS NOT NULL AND x.broadcaster_id = c.broadcaster_id)
		  )`)
	if err != nil {
		return 0, err
	}
	return tag.RowsAffected(), nil
}

// RecordTranscriptionUsage adds one transcription to the day's totals.
func (r *SearchTranscriptionRepository) RecordTranscriptionUsage(ctx context.Context, u TranscriptionUsage) error {
	return recordTranscriptionUsage(ctx, r.pool, u)
}
