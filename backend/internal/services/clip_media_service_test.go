package services

import (
	"context"
	"net/http"
	"net/http/httptest"
	"os"
	"os/exec"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestExtractAudioFromURLStreamsMediaWithoutVideoFile(t *testing.T) {
	if _, err := exec.LookPath("ffmpeg"); err != nil {
		t.Skip("ffmpeg not available")
	}

	tmpDir := t.TempDir()
	videoPath := filepath.Join(tmpDir, "source.mp4")
	generate := exec.Command("ffmpeg", "-f", "lavfi", "-i", "sine=frequency=440:duration=1", "-c:a", "aac", "-y", videoPath)
	output, err := generate.CombinedOutput()
	require.NoError(t, err, "generate fixture: %s", output)

	server := httptest.NewServer(http.FileServer(http.Dir(tmpDir)))
	defer server.Close()
	audioDir := filepath.Join(tmpDir, "audio")

	wavPath, err := ExtractAudioFromURL(context.Background(), "ffmpeg", server.URL+"/source.mp4", audioDir)

	require.NoError(t, err)
	info, err := os.Stat(wavPath)
	require.NoError(t, err)
	assert.Greater(t, info.Size(), int64(44))
	entries, err := os.ReadDir(audioDir)
	require.NoError(t, err)
	assert.Len(t, entries, 1)
	assert.Equal(t, ".wav", filepath.Ext(entries[0].Name()))
}
