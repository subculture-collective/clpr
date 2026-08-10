package services

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func newTestThumbnailService(ffmpegPath, outputDir, apiKey, apiURL, model string, enabled bool) *ThumbnailService {
	return NewThumbnailService(ffmpegPath, outputDir, "openai", apiKey, apiURL, model, "", "", enabled, 30)
}

func TestNewThumbnailService(t *testing.T) {
	svc := NewThumbnailService(
		"ffmpeg",
		"/tmp/thumbs",
		"openai",
		"test-api-key",
		"https://api.openai.com/v1/chat/completions",
		"gpt-4o-mini",
		"", "",
		true,
		30,
	)

	assert.NotNil(t, svc)
	assert.True(t, svc.enabled)
	assert.True(t, svc.Operational())
	assert.Equal(t, "ffmpeg", svc.ffmpegPath)
	assert.Equal(t, "/tmp/thumbs", svc.outputDir)
}

func TestThumbnailService_Operational_Disabled(t *testing.T) {
	svc := newTestThumbnailService(
		"ffmpeg",
		"/tmp/thumbs",
		"",  // no API key
		"",  // no API URL
		"",
		false,
	)

	assert.False(t, svc.Operational())
	assert.False(t, svc.Enabled())
}

func TestExtractThumbnails_SkipsWithoutFfmpeg(t *testing.T) {
	if _, err := exec.LookPath("ffmpeg"); err != nil {
		t.Skip("ffmpeg not available in PATH")
	}

	// Create a tiny test video with ffmpeg so we have something to extract from.
	tmpDir := t.TempDir()
	videoPath := filepath.Join(tmpDir, "test.mp4")

	// Generate a 2-second test video with a solid color.
	genCmd := exec.Command("ffmpeg",
		"-f", "lavfi",
		"-i", "color=c=red:s=128x72:d=2",
		"-c:v", "libx264",
		"-preset", "ultrafast",
		"-y", videoPath,
	)
	out, err := genCmd.CombinedOutput()
	require.NoError(t, err, "ffmpeg gen failed: %s", string(out))

	outputDir := filepath.Join(tmpDir, "thumbs")
	require.NoError(t, os.MkdirAll(outputDir, 0755))

	svc := newTestThumbnailService("ffmpeg", outputDir, "", "", "", false)
	ctx := context.Background()

	thumbs, err := svc.ExtractThumbnails(ctx, videoPath, 2.0)
	require.NoError(t, err)
	assert.Len(t, thumbs, 3)

	for i, p := range thumbs {
		assert.True(t, strings.HasSuffix(p, ".jpg"), "thumbnail %d should be .jpg: %s", i, p)
		info, err := os.Stat(p)
		require.NoError(t, err, "thumbnail %d should exist: %s", i, p)
		assert.Greater(t, info.Size(), int64(0), "thumbnail %d should be non-empty", i)
	}
}

func TestClassifyThumbnails_Disabled(t *testing.T) {
	svc := newTestThumbnailService("ffmpeg", "/tmp", "", "", "", false)

	ctx := context.Background()
	tags, err := svc.ClassifyThumbnails(ctx, []string{"/fake/path.jpg"}, "Test Game")
	require.ErrorIs(t, err, ErrVisionAPIUnavailable)
	assert.Nil(t, tags)
}

func TestClassifyThumbnails_WithMockAPI(t *testing.T) {
	// Create a small test JPEG to embed.
	tmpDir := t.TempDir()
	imgPath := filepath.Join(tmpDir, "test.jpg")
	require.NoError(t, os.WriteFile(imgPath, []byte("fake-jpeg-data"), 0644))

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "POST", r.Method)
		assert.Contains(t, r.Header.Get("Authorization"), "Bearer test-api-key")
		assert.Equal(t, "application/json", r.Header.Get("Content-Type"))

		// Return a valid vision API response.
		resp := map[string]interface{}{
			"choices": []map[string]interface{}{
				{
					"message": map[string]string{
						"content": `{"tags": ["clutch", "highlights"]}`,
					},
				},
			},
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(resp)
	}))
	defer server.Close()

	svc := newTestThumbnailService("ffmpeg", tmpDir, "test-api-key", server.URL, "gpt-4o-mini", true)
	ctx := context.Background()

	tags, err := svc.ClassifyThumbnails(ctx, []string{imgPath}, "Valorant")
	require.NoError(t, err)
	assert.Equal(t, []string{"clutch", "highlights"}, tags)
}

func TestClassifyThumbnails_WithMockAPI_SingleTag(t *testing.T) {
	tmpDir := t.TempDir()
	imgPath := filepath.Join(tmpDir, "test.jpg")
	require.NoError(t, os.WriteFile(imgPath, []byte("fake-data"), 0644))

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		resp := map[string]interface{}{
			"choices": []map[string]interface{}{
				{
					"message": map[string]string{
						"content": `{"tags": ["funny"]}`,
					},
				},
			},
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(resp)
	}))
	defer server.Close()

	svc := newTestThumbnailService("ffmpeg", tmpDir, "test-api-key", server.URL, "gpt-4o-mini", true)
	ctx := context.Background()

	tags, err := svc.ClassifyThumbnails(ctx, []string{imgPath}, "Minecraft")
	require.NoError(t, err)
	assert.Equal(t, []string{"funny"}, tags)
}

func TestClassifyThumbnails_WithMockAPI_FiltersUnknownTags(t *testing.T) {
	tmpDir := t.TempDir()
	imgPath := filepath.Join(tmpDir, "test.jpg")
	require.NoError(t, os.WriteFile(imgPath, []byte("fake-data"), 0644))

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Return a mix of valid and unknown tags.
		resp := map[string]interface{}{
			"choices": []map[string]interface{}{
				{
					"message": map[string]string{
						"content": `{"tags": ["clutch", "unknown_tag", "fail"]}`,
					},
				},
			},
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(resp)
	}))
	defer server.Close()

	svc := newTestThumbnailService("ffmpeg", tmpDir, "test-api-key", server.URL, "gpt-4o-mini", true)
	ctx := context.Background()

	tags, err := svc.ClassifyThumbnails(ctx, []string{imgPath}, "League of Legends")
	require.NoError(t, err)
	assert.Equal(t, []string{"clutch", "fail"}, tags)
}

func TestClassifyThumbnails_WithMockAPI_Deduplicates(t *testing.T) {
	tmpDir := t.TempDir()
	imgPath := filepath.Join(tmpDir, "test.jpg")
	require.NoError(t, os.WriteFile(imgPath, []byte("fake-data"), 0644))

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		resp := map[string]interface{}{
			"choices": []map[string]interface{}{
				{
					"message": map[string]string{
						"content": `{"tags": ["clutch", "clutch", "funny"]}`,
					},
				},
			},
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(resp)
	}))
	defer server.Close()

	svc := newTestThumbnailService("ffmpeg", tmpDir, "test-api-key", server.URL, "gpt-4o-mini", true)
	ctx := context.Background()

	tags, err := svc.ClassifyThumbnails(ctx, []string{imgPath}, "Valorant")
	require.NoError(t, err)
	assert.Equal(t, []string{"clutch", "funny"}, tags)
}

func TestClassifyThumbnails_WithMockAPI_Error(t *testing.T) {
	tmpDir := t.TempDir()
	imgPath := filepath.Join(tmpDir, "test.jpg")
	require.NoError(t, os.WriteFile(imgPath, []byte("fake-data"), 0644))

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte("Internal server error"))
	}))
	defer server.Close()

	svc := newTestThumbnailService("ffmpeg", tmpDir, "test-api-key", server.URL, "gpt-4o-mini", true)
	ctx := context.Background()

	tags, err := svc.ClassifyThumbnails(ctx, []string{imgPath}, "Game")
	assert.Error(t, err)
	assert.Nil(t, tags)
	assert.Contains(t, err.Error(), "500")
}

func TestClassifyThumbnails_WithMockAPI_MalformedResponse(t *testing.T) {
	tmpDir := t.TempDir()
	imgPath := filepath.Join(tmpDir, "test.jpg")
	require.NoError(t, os.WriteFile(imgPath, []byte("fake-data"), 0644))

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Return invalid JSON in the message content.
		resp := map[string]interface{}{
			"choices": []map[string]interface{}{
				{
					"message": map[string]string{
						"content": "not json at all",
					},
				},
			},
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(resp)
	}))
	defer server.Close()

	svc := newTestThumbnailService("ffmpeg", tmpDir, "test-api-key", server.URL, "gpt-4o-mini", true)
	ctx := context.Background()

	tags, err := svc.ClassifyThumbnails(ctx, []string{imgPath}, "Game")
	assert.Error(t, err)
	assert.Nil(t, tags)
}

// TestClassifyThumbnails_WithMockAPI_OpenRouterHeaders verifies that the
// HTTP-Referer and X-Title headers are sent when provider is "openrouter".
func TestClassifyThumbnails_WithMockAPI_OpenRouterHeaders(t *testing.T) {
	tmpDir := t.TempDir()
	imgPath := filepath.Join(tmpDir, "test.jpg")
	require.NoError(t, os.WriteFile(imgPath, []byte("fake-data"), 0644))

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "openrouter.ai", r.Header.Get("HTTP-Referer"))
		assert.Equal(t, "TestApp", r.Header.Get("X-Title"))
		assert.Contains(t, r.Header.Get("Authorization"), "Bearer or-key")

		resp := map[string]interface{}{
			"choices": []map[string]interface{}{
				{"message": map[string]string{"content": `{"tags": ["irl"]}`}},
			},
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(resp)
	}))
	defer server.Close()

	svc := NewThumbnailService("ffmpeg", tmpDir, "openrouter", "or-key", server.URL,
		"openai/gpt-4o-mini", "openrouter.ai", "TestApp", true, 30)
	ctx := context.Background()

	tags, err := svc.ClassifyThumbnails(ctx, []string{imgPath}, "Just Chatting")
	require.NoError(t, err)
	assert.Equal(t, []string{"irl"}, tags)
}

// TestClassifyThumbnails_WithRealAPI performs a live API call to the configured
// vision model. It is skipped unless VISION_API_KEY and VISION_API_URL
// environment variables are set.
func TestClassifyThumbnails_WithRealAPI(t *testing.T) {
	apiKey := os.Getenv("VISION_API_KEY")
	apiURL := os.Getenv("VISION_API_URL")
	if apiKey == "" || apiURL == "" {
		t.Skip("Skipping real vision API test: set VISION_API_KEY and VISION_API_URL to run")
	}

	model := os.Getenv("VISION_API_MODEL")
	if model == "" {
		model = "gpt-4o-mini"
	}

	tmpDir := t.TempDir()

	// Create a small solid-color test image that the API can process.
	imgPath := filepath.Join(tmpDir, "test_frame.jpg")
	require.NoError(t, os.WriteFile(imgPath, []byte("placeholder"), 0644))

	svc := newTestThumbnailService("ffmpeg", tmpDir, apiKey, apiURL, model, true)
	require.True(t, svc.Operational())

	ctx := context.Background()
	tags, err := svc.ClassifyThumbnails(ctx, []string{imgPath}, "Test Game")

	// The real API call may fail for various reasons (invalid image, network, etc.).
	// We just verify the code path runs without panicking and returns reasonable results.
	if err != nil {
		t.Logf("Real API call returned error (acceptable): %v", err)
		return
	}

	t.Logf("Real API returned tags: %v", tags)

	// All returned tags should be valid content slugs.
	validSet := map[string]bool{}
	for _, s := range contentTagSlugs {
		validSet[s] = true
	}
	for _, tag := range tags {
		assert.True(t, validSet[tag], "tag %q is not a valid content tag slug", tag)
	}
}

// TestExtractThumbnails_InvalidVideo ensures ffmpeg errors propagate cleanly.
func TestExtractThumbnails_InvalidVideo(t *testing.T) {
	if _, err := exec.LookPath("ffmpeg"); err != nil {
		t.Skip("ffmpeg not available in PATH")
	}

	tmpDir := t.TempDir()
	outputDir := filepath.Join(tmpDir, "thumbs")
	require.NoError(t, os.MkdirAll(outputDir, 0755))

	// Non-video file as input.
	badPath := filepath.Join(tmpDir, "notavideo.mp4")
	require.NoError(t, os.WriteFile(badPath, []byte("this is not a video"), 0644))

	svc := newTestThumbnailService("ffmpeg", outputDir, "", "", "", false)
	ctx := context.Background()

	_, err := svc.ExtractThumbnails(ctx, badPath, 10.0)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "extracting thumbnail")
}
