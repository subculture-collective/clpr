package services

import (
	"context"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

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