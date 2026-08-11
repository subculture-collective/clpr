package services

import (
	"context"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestWhisperServiceTranscribesThroughPackagedRunner(t *testing.T) {
	tmpDir := t.TempDir()
	pythonPath := filepath.Join(tmpDir, "python3")
	runnerPath := filepath.Join(tmpDir, "whisper_runner.py")
	wavPath := filepath.Join(tmpDir, "clip.wav")
	fakePython := "#!/bin/sh\n" +
		"[ \"$1\" = \"" + runnerPath + "\" ] || exit 21\n" +
		"[ \"$2\" = \"" + wavPath + "\" ] || exit 22\n" +
		"printf '%s' '{\"segments\":[{\"start\":0,\"end\":1.2,\"text\":\"what a save\",\"avg_logprob\":-0.1}],\"language\":\"en\",\"full_text\":\"what a save\"}'\n"
	require.NoError(t, os.WriteFile(pythonPath, []byte(fakePython), 0755))
	require.NoError(t, os.WriteFile(runnerPath, []byte("# test runner"), 0644))
	require.NoError(t, os.WriteFile(wavPath, []byte("wav"), 0644))

	result, err := NewWhisperService(pythonPath, tmpDir).TranscribeAudio(context.Background(), wavPath)

	require.NoError(t, err)
	require.NotNil(t, result)
	assert.Equal(t, "what a save", result.FullText)
	assert.Equal(t, "en", result.Language)
	require.Len(t, result.Segments, 1)
	assert.Equal(t, 1.2, result.Segments[0].End)
}

func TestWhisperService_TranscribeAudio(t *testing.T) {
	if os.Getenv("WHISPER_TEST") != "1" {
		t.Skip("Skipping Whisper test (set WHISPER_TEST=1 to run)")
	}

	// Defaults: python3, runner dir relative to backend/internal/services
	// (i.e. ../../whisper). Adjust WHISPER_PYTHON / WHISPER_RUNNER_DIR to
	// override for CI or local dev.
	pythonPath := os.Getenv("WHISPER_PYTHON")
	if pythonPath == "" {
		pythonPath = "python3"
	}
	runnerDir := os.Getenv("WHISPER_RUNNER_DIR")
	if runnerDir == "" {
		runnerDir = "../../whisper"
	}

	svc := NewWhisperService(pythonPath, runnerDir)

	// Use testdata/sample.wav if present; otherwise skip with a clear message.
	wavPath := "testdata/sample.wav"
	if _, err := os.Stat(wavPath); os.IsNotExist(err) {
		t.Skipf("Test WAV file not found at %s — provide one to run the Whisper integration test", wavPath)
	}

	result, err := svc.TranscribeAudio(context.Background(), wavPath)
	require.NoError(t, err)
	require.NotNil(t, result)
	assert.NotEmpty(t, result.FullText)
	t.Logf("language=%s full_text=%q", result.Language, result.FullText)
}
