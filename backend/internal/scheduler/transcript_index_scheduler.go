package scheduler

import (
	"context"
	"time"

	"git.subcult.tv/subculture-collective/clpr/internal/repository"
	"git.subcult.tv/subculture-collective/clpr/pkg/utils"
	"github.com/google/uuid"
)

// TranscriptIndexStore lists the transcripts the search index lacks.
type TranscriptIndexStore interface {
	ListTranscriptsToIndex(ctx context.Context, limit int) ([]repository.PendingTranscript, error)
	MarkTranscriptsIndexed(ctx context.Context, indexed []repository.PendingTranscript) error
}

// TranscriptIndexer writes transcripts into the clips search index.
type TranscriptIndexer interface {
	UpdateClipTranscripts(ctx context.Context, transcripts map[uuid.UUID]string) error
}

// TranscriptIndexScheduler copies new and changed clip transcripts into the
// search index (#87). Transcripts are written by the API's authorized
// transcription and by the render host, which has no search client.
type TranscriptIndexScheduler struct {
	store    TranscriptIndexStore
	index    TranscriptIndexer
	interval time.Duration
	batch    int
}

// NewTranscriptIndexScheduler creates the scheduler; interval defaults to
// five minutes.
func NewTranscriptIndexScheduler(store TranscriptIndexStore, index TranscriptIndexer, interval time.Duration) *TranscriptIndexScheduler {
	if interval <= 0 {
		interval = 5 * time.Minute
	}
	return &TranscriptIndexScheduler{store: store, index: index, interval: interval, batch: 200}
}

// Start runs until ctx is done.
func (s *TranscriptIndexScheduler) Start(ctx context.Context) {
	s.RunOnce(ctx)
	ticker := time.NewTicker(s.interval)
	defer ticker.Stop()
	for {
		select {
		case <-ticker.C:
			s.RunOnce(ctx)
		case <-ctx.Done():
			return
		}
	}
}

// RunOnce copies pending transcripts in batches until none are left or a
// step fails, and returns how many it copied.
func (s *TranscriptIndexScheduler) RunOnce(ctx context.Context) int {
	copied := 0
	for ctx.Err() == nil {
		pending, err := s.store.ListTranscriptsToIndex(ctx, s.batch)
		if err != nil {
			utils.Error("Listing transcripts to index failed", err, nil)
			return copied
		}
		if len(pending) == 0 {
			return copied
		}
		texts := make(map[uuid.UUID]string, len(pending))
		for _, p := range pending {
			texts[p.ClipID] = p.FullText
		}
		if err := s.index.UpdateClipTranscripts(ctx, texts); err != nil {
			utils.Warn("Indexing clip transcripts failed", map[string]interface{}{"error": err.Error(), "transcripts": len(pending)})
			return copied
		}
		if err := s.store.MarkTranscriptsIndexed(ctx, pending); err != nil {
			utils.Error("Marking transcripts indexed failed", err, nil)
			return copied
		}
		copied += len(pending)
		if len(pending) < s.batch {
			return copied
		}
	}
	return copied
}
