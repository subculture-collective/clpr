package main

import (
	"context"
	"errors"
	"strings"
	"testing"

	"git.subcult.tv/subculture-collective/clpr/internal/models"
	"github.com/google/uuid"
)

type fakeGenerator struct {
	calls int
	fail  map[uuid.UUID]bool
}

func (f *fakeGenerator) GenerateClipEmbedding(_ context.Context, clip *models.Clip) ([]float32, error) {
	f.calls++
	if f.fail[clip.ID] {
		return nil, errors.New("generation failed")
	}
	return []float32{0.1, 0.2}, nil
}

func (f *fakeGenerator) GetModel() string { return "test-model" }

func TestBackfillBatchQuery_UsesKeysetPagination(t *testing.T) {
	for _, force := range []bool{false, true} {
		first := backfillBatchQuery(force, false)
		next := backfillBatchQuery(force, true)
		for _, q := range []string{first, next} {
			if strings.Contains(q, "OFFSET") {
				t.Fatalf("query must not page with OFFSET: %s", q)
			}
			if !strings.Contains(q, "ORDER BY id ASC") {
				t.Fatalf("query must order by id for keyset paging: %s", q)
			}
			if strings.Contains(q, "embedding IS NULL") == force {
				t.Fatalf("force=%t: unexpected embedding filter in %s", force, q)
			}
		}
		if strings.Contains(first, "id > $2") || !strings.Contains(next, "id > $2") {
			t.Fatalf("cursor condition must appear only after the first page")
		}
	}
}

func TestProcessBatch_DryRunDoesNotCallEmbeddingAPI(t *testing.T) {
	gen := &fakeGenerator{}
	saved := 0
	save := func(context.Context, *models.Clip, []float32, string) error { saved++; return nil }
	stats := &BackfillStats{TotalClips: 2}
	clips := []models.Clip{{ID: uuid.New()}, {ID: uuid.New()}}

	processBatch(context.Background(), clips, gen, save, false, true, stats)

	if gen.calls != 0 || saved != 0 {
		t.Fatalf("dry run called generator %d times and saved %d times", gen.calls, saved)
	}
	if stats.ProcessedClips != 2 {
		t.Fatalf("processed = %d, want 2", stats.ProcessedClips)
	}
}

func TestProcessBatch_CountsFailuresAndSkips(t *testing.T) {
	failing := uuid.New()
	gen := &fakeGenerator{fail: map[uuid.UUID]bool{failing: true}}
	saveErrID := uuid.New()
	save := func(_ context.Context, clip *models.Clip, _ []float32, model string) error {
		if model != "test-model" {
			t.Fatalf("model = %q", model)
		}
		if clip.ID == saveErrID {
			return errors.New("save failed")
		}
		return nil
	}
	stats := &BackfillStats{TotalClips: 4}
	clips := []models.Clip{
		{ID: uuid.New()},
		{ID: failing},
		{ID: saveErrID},
		{ID: uuid.New(), Embedding: []float32{1}},
	}

	processBatch(context.Background(), clips, gen, save, false, false, stats)

	if stats.ProcessedClips != 1 || stats.FailedClips != 2 || stats.SkippedClips != 1 {
		t.Fatalf("stats = processed %d failed %d skipped %d, want 1/2/1",
			stats.ProcessedClips, stats.FailedClips, stats.SkippedClips)
	}
}
