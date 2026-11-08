package services

import (
	"context"
	"errors"
	"os"
	"os/exec"
	"path/filepath"
	"testing"

	"git.subcult.tv/subculture-collective/clpr/internal/countdownrender"
	"github.com/prometheus/client_golang/prometheus/testutil"
)

type fakeCaptioner struct {
	words map[string][]countdownrender.Word
	err   error
	paths []string
}

func (f *fakeCaptioner) Words(_ context.Context, path string) ([]countdownrender.Word, error) {
	f.paths = append(f.paths, path)
	return f.words[path], f.err
}

func TestCountdownRenderWorkerCaptionsShorts(t *testing.T) {
	store, worker := renderFixture(t, 2)
	worker.Metrics = NewCountdownRenderMetrics(nil)
	first := filepath.Join(worker.MediaDir, store.media[0].FileName)
	captioner := &fakeCaptioner{words: map[string][]countdownrender.Word{first: {{Start: 0.1, End: 0.4, Text: "hello"}}}}
	worker.Captioner = captioner
	renderer := worker.Renderer.(*fakeRenderer)
	if err := worker.RunOnce(context.Background()); err != nil {
		t.Fatal(err)
	}
	if len(captioner.paths) != 2 {
		t.Fatalf("captioner ran for %d clips, want the 2 Shorts only", len(captioner.paths))
	}
	for _, short := range renderer.shorts {
		if want := map[bool]int{true: 1, false: 0}[short.Path == first]; len(short.Words) != want {
			t.Fatalf("short #%d (%s) has %d words, want %d", short.Rank, short.Path, len(short.Words), want)
		}
	}
	if len(renderer.countdown.Clips[0].Words) != 0 {
		t.Fatal("the countdown video got captions")
	}
	if got := testutil.ToFloat64(worker.Metrics.Captions.WithLabelValues("captioned")); got != 1 {
		t.Fatalf("captioned = %v", got)
	}
	if got := testutil.ToFloat64(worker.Metrics.Captions.WithLabelValues("none")); got != 1 {
		t.Fatalf("none = %v", got)
	}
}

func TestCountdownRenderWorkerRendersShortsWhenCaptionsFail(t *testing.T) {
	store, worker := renderFixture(t, 1)
	worker.Metrics = NewCountdownRenderMetrics(nil)
	worker.Captioner = &fakeCaptioner{err: errors.New("whisper runner crashed")}
	if err := worker.RunOnce(context.Background()); err != nil {
		t.Fatal(err)
	}
	if store.status != "rendered" || len(worker.Renderer.(*fakeRenderer).shorts) != 1 {
		t.Fatalf("status %q, shorts %d", store.status, len(worker.Renderer.(*fakeRenderer).shorts))
	}
	if got := testutil.ToFloat64(worker.Metrics.Captions.WithLabelValues("failed")); got != 1 {
		t.Fatalf("failed = %v", got)
	}
}

type fakeWordTranscriber struct{ result *WhisperResult }

func (f fakeWordTranscriber) TranscribeWords(_ context.Context, wav string) (*WhisperResult, error) {
	if _, err := os.Stat(wav); err != nil {
		return nil, err
	}
	return f.result, nil
}

func TestWhisperCaptionerKeepsConfidentEnglishSpeech(t *testing.T) {
	ffmpeg, err := exec.LookPath("ffmpeg")
	if err != nil {
		t.Skip("ffmpeg is not installed")
	}
	dir := t.TempDir()
	clip := filepath.Join(dir, "clip.mp4")
	if out, err := exec.Command(ffmpeg, "-nostdin", "-loglevel", "error", "-f", "lavfi", "-i", "sine=f=440:d=1", clip).CombinedOutput(); err != nil {
		t.Fatalf("fixture: %v %s", err, out)
	}
	result := &WhisperResult{Language: "en", Segments: []WhisperSegment{
		{AvgLogprob: -0.3, NoSpeechProb: 0.1, Words: []WhisperWord{{Start: 0.1, End: 0.3, Word: "nice"}, {Start: 0.3, End: 0.5, Word: "shot"}}},
		{AvgLogprob: -0.2, NoSpeechProb: 0.9, Words: []WhisperWord{{Start: 0.6, End: 0.7, Word: "music"}}},
		{AvgLogprob: -1.6, NoSpeechProb: 0.1, Words: []WhisperWord{{Start: 0.7, End: 0.9, Word: "thanks for watching"}}},
	}}
	work := filepath.Join(dir, "work")
	c := &WhisperCaptioner{Whisper: fakeWordTranscriber{result}, FFmpegPath: ffmpeg, WorkDir: work}
	words, err := c.Words(context.Background(), clip)
	if err != nil {
		t.Fatal(err)
	}
	if len(words) != 2 || words[0].Text != "nice" || words[1].Text != "shot" {
		t.Fatalf("words = %+v", words)
	}
	if entries, _ := os.ReadDir(work); len(entries) != 0 {
		t.Fatalf("temporary audio left behind: %v", entries)
	}

	result.Language = "de"
	if words, err := c.Words(context.Background(), clip); err != nil || words != nil {
		t.Fatalf("a German clip got captions: %v, %v", words, err)
	}
}
