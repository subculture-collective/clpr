package repository

import (
	"context"
	"encoding/json"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgconn"
)

// Transcript sources (#87).
const (
	TranscriptSourceAuthorized = "twitch_authorized_whisper"
	TranscriptSourceCountdown  = "countdown_render_whisper"
)

type execer interface {
	Exec(context.Context, string, ...any) (pgconn.CommandTag, error)
}

// TranscriptionUsage is one transcription to add to the day's totals.
type TranscriptionUsage struct {
	Day    time.Time
	Source string
	// Transcribed is whether it produced a stored transcript.
	Transcribed    bool
	AudioSeconds   float64
	ComputeSeconds float64
}

func recordTranscriptionUsage(ctx context.Context, db execer, u TranscriptionUsage) error {
	transcripts := 0
	if u.Transcribed {
		transcripts = 1
	}
	_, err := db.Exec(ctx, `
		INSERT INTO transcription_usage_daily (day, source, clips, transcripts, audio_seconds, compute_seconds)
		VALUES ($1::date, $2, 1, $3, $4, $5)
		ON CONFLICT (day, source) DO UPDATE SET
			clips = transcription_usage_daily.clips + 1,
			transcripts = transcription_usage_daily.transcripts + EXCLUDED.transcripts,
			audio_seconds = transcription_usage_daily.audio_seconds + EXCLUDED.audio_seconds,
			compute_seconds = transcription_usage_daily.compute_seconds + EXCLUDED.compute_seconds,
			updated_at = NOW()`,
		u.Day.UTC().Format("2006-01-02"), u.Source, transcripts, max(u.AudioSeconds, 0), max(u.ComputeSeconds, 0))
	return err
}

// RecordTranscriptionUsage adds one transcription to the day's totals.
func (r *ClipRepository) RecordTranscriptionUsage(ctx context.Context, u TranscriptionUsage) error {
	return recordTranscriptionUsage(ctx, r.pool, u)
}

// RecordTranscriptionUsage adds one transcription to the day's totals.
func (r *CountdownRepository) RecordTranscriptionUsage(ctx context.Context, u TranscriptionUsage) error {
	return recordTranscriptionUsage(ctx, r.pool, u)
}

// SaveCountdownTranscript stores the render host's transcript of a countdown
// clip, unless the clip already has one from the broadcaster's own
// authorization. It reports whether it stored the transcript.
func (r *CountdownRepository) SaveCountdownTranscript(ctx context.Context, clipID uuid.UUID, language, fullText string, segments json.RawMessage) (bool, error) {
	tag, err := r.pool.Exec(ctx, `
		INSERT INTO clip_transcripts (clip_id, language, full_text, segments, source)
		VALUES ($1, NULLIF($2, ''), $3, $4, 'countdown_render_whisper')
		ON CONFLICT (clip_id) DO UPDATE SET
			language = EXCLUDED.language, full_text = EXCLUDED.full_text, segments = EXCLUDED.segments, updated_at = NOW()
		WHERE clip_transcripts.source = 'countdown_render_whisper'`, clipID, language, fullText, segments)
	if err != nil || tag.RowsAffected() == 0 {
		return false, err
	}
	return true, nil
}

// PendingTranscript is a transcript the search index does not have yet.
type PendingTranscript struct {
	ClipID    uuid.UUID
	FullText  string
	UpdatedAt time.Time
}

// ListTranscriptsToIndex returns transcripts written or changed since they
// were last copied into the search index, oldest first.
func (r *ClipRepository) ListTranscriptsToIndex(ctx context.Context, limit int) ([]PendingTranscript, error) {
	rows, err := r.pool.Query(ctx, `
		SELECT clip_id, full_text, updated_at
		FROM clip_transcripts
		WHERE search_indexed_at IS NULL OR search_indexed_at < updated_at
		ORDER BY updated_at
		LIMIT $1`, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := []PendingTranscript{}
	for rows.Next() {
		var p PendingTranscript
		if err := rows.Scan(&p.ClipID, &p.FullText, &p.UpdatedAt); err != nil {
			return nil, err
		}
		out = append(out, p)
	}
	return out, rows.Err()
}

// MarkTranscriptsIndexed records that the search index has these transcript
// versions. A transcript changed since it was read stays pending.
func (r *ClipRepository) MarkTranscriptsIndexed(ctx context.Context, indexed []PendingTranscript) error {
	if len(indexed) == 0 {
		return nil
	}
	ids := make([]uuid.UUID, len(indexed))
	versions := make([]time.Time, len(indexed))
	for i, p := range indexed {
		ids[i], versions[i] = p.ClipID, p.UpdatedAt
	}
	_, err := r.pool.Exec(ctx, `
		UPDATE clip_transcripts t SET search_indexed_at = v.version
		FROM unnest($1::uuid[], $2::timestamptz[]) AS v(clip_id, version)
		WHERE t.clip_id = v.clip_id`, ids, versions)
	return err
}
