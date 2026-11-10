package scheduler

import (
	"context"
	"errors"
	"testing"
	"time"

	"git.subcult.tv/subculture-collective/clpr/internal/repository"
	"github.com/google/uuid"
)

type fakeTranscriptQueue struct {
	pending []repository.PendingTranscript
	marked  []repository.PendingTranscript
}

func (f *fakeTranscriptQueue) ListTranscriptsToIndex(_ context.Context, limit int) ([]repository.PendingTranscript, error) {
	return f.pending[:min(limit, len(f.pending))], nil
}

func (f *fakeTranscriptQueue) MarkTranscriptsIndexed(_ context.Context, done []repository.PendingTranscript) error {
	f.marked = append(f.marked, done...)
	f.pending = f.pending[len(done):]
	return nil
}

type fakeTranscriptIndex struct {
	calls []map[uuid.UUID]string
	err   error
}

func (f *fakeTranscriptIndex) UpdateClipTranscripts(_ context.Context, t map[uuid.UUID]string) error {
	f.calls = append(f.calls, t)
	return f.err
}

func TestTranscriptIndexSchedulerCopiesInBatches(t *testing.T) {
	queue := &fakeTranscriptQueue{}
	for i := 0; i < 5; i++ {
		queue.pending = append(queue.pending, repository.PendingTranscript{ClipID: uuid.New(), FullText: "words", UpdatedAt: time.Now()})
	}
	index := &fakeTranscriptIndex{}
	s := NewTranscriptIndexScheduler(queue, index, 0)
	s.batch = 2
	if copied := s.RunOnce(context.Background()); copied != 5 || len(index.calls) != 3 || len(queue.marked) != 5 {
		t.Fatalf("copied %d in %d calls, marked %d", copied, len(index.calls), len(queue.marked))
	}
}

func TestTranscriptIndexSchedulerKeepsPendingOnFailure(t *testing.T) {
	queue := &fakeTranscriptQueue{pending: []repository.PendingTranscript{{ClipID: uuid.New(), FullText: "x"}}}
	s := NewTranscriptIndexScheduler(queue, &fakeTranscriptIndex{err: errors.New("opensearch down")}, 0)
	if copied := s.RunOnce(context.Background()); copied != 0 || len(queue.marked) != 0 {
		t.Fatalf("copied %d, marked %v after a failed update", copied, queue.marked)
	}
}
