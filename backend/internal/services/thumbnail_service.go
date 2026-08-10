package services

import (
	"bytes"
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"
)

// contentTagSlugs is the set of valid content tag slugs the vision model can choose from.
var contentTagSlugs = []string{
	"clutch",
	"funny",
	"fail",
	"educational",
	"highlights",
	"reaction",
	"speedrun",
	"music",
	"creative",
	"irl",
}

// ErrVisionAPIUnavailable is returned when the vision API is not configured.
var ErrVisionAPIUnavailable = fmt.Errorf("vision API is disabled or not configured")

// ThumbnailService handles thumbnail extraction via ffmpeg and content
// classification via a configurable vision AI API.
type ThumbnailService struct {
	ffmpegPath string
	outputDir  string

	provider   string
	httpClient *http.Client
	apiKey     string
	apiURL     string
	model      string
	siteURL    string
	siteName   string
	enabled    bool
}

// NewThumbnailService creates a new ThumbnailService.
//
// ffmpegPath is the path to the ffmpeg binary. outputDir is where extracted
// thumbnails are written. provider is one of "openai", "openrouter", "anthropic".
// apiKey, apiURL, and model configure the vision AI call.
// siteURL and siteName are used for the OpenRouter HTTP-Referer / X-Title headers.
func NewThumbnailService(
	ffmpegPath, outputDir string,
	provider, apiKey, apiURL, model string,
	siteURL, siteName string,
	enabled bool,
	timeoutSeconds int,
) *ThumbnailService {
	if timeoutSeconds <= 0 {
		timeoutSeconds = 30
	}
	return &ThumbnailService{
		ffmpegPath: ffmpegPath,
		outputDir:  outputDir,
		provider:   provider,
		httpClient: &http.Client{
			Timeout: time.Duration(timeoutSeconds) * time.Second,
		},
		apiKey:   apiKey,
		apiURL:   apiURL,
		model:    model,
		siteURL:  siteURL,
		siteName: siteName,
		enabled:  enabled,
	}
}

// Operational reports whether the vision API is configured and ready.
func (ts *ThumbnailService) Operational() bool {
	return ts != nil && ts.enabled && ts.apiKey != "" && ts.apiURL != ""
}

// Enabled reports the configured feature flag.
func (ts *ThumbnailService) Enabled() bool {
	return ts != nil && ts.enabled
}

// ExtractThumbnails extracts 3 frames at 25%, 50%, and 75% of the clip
// duration using ffmpeg. The returned slice contains absolute paths to the
// generated JPEG files in the service's output directory.
func (ts *ThumbnailService) ExtractThumbnails(ctx context.Context, videoPath string, duration float64) ([]string, error) {
	positions := []float64{0.25, 0.50, 0.75}
	thumbnails := make([]string, 0, len(positions))

	baseName := filepath.Base(videoPath)
	ext := filepath.Ext(baseName)
	stem := strings.TrimSuffix(baseName, ext)

	for i, pos := range positions {
		timestamp := duration * pos
		outPath := filepath.Join(ts.outputDir, fmt.Sprintf("%s_thumb_%d.jpg", stem, i))

		cmd := exec.CommandContext(ctx, ts.ffmpegPath,
			"-ss", fmt.Sprintf("%.1f", timestamp),
			"-i", videoPath,
			"-vframes", "1",
			"-q:v", "2",
			"-y", outPath,
		)

		// Suppress stderr noise from ffmpeg (it's verbose by default)
		var stderr bytes.Buffer
		cmd.Stderr = &stderr

		if err := cmd.Run(); err != nil {
			return nil, fmt.Errorf("extracting thumbnail at %.1fs from %s: %w\nffmpeg stderr: %s",
				timestamp, videoPath, err, stderr.String())
		}
		thumbnails = append(thumbnails, outPath)
	}
	return thumbnails, nil
}

// ClassifyThumbnails sends base64-encoded images to a vision AI API for
// content classification. It reads each image file at imagePaths, encodes it
// as base64, constructs a multi-image prompt with the game name, and calls the
// configured OpenAI-compatible chat completions endpoint. The response is
// parsed into a slice of content tag slugs.
func (ts *ThumbnailService) ClassifyThumbnails(ctx context.Context, imagePaths []string, gameName string) ([]string, error) {
	if !ts.Operational() {
		return nil, ErrVisionAPIUnavailable
	}

	// Read and base64-encode images.
	type base64Image struct {
		path     string
		data     []byte
		mimeType string
	}
	images := make([]base64Image, len(imagePaths))
	for i, p := range imagePaths {
		data, err := os.ReadFile(p)
		if err != nil {
			return nil, fmt.Errorf("reading image %s: %w", p, err)
		}
		mimeType := "image/jpeg"
		if strings.HasSuffix(strings.ToLower(p), ".png") {
			mimeType = "image/png"
		}
		images[i] = base64Image{path: p, data: data, mimeType: mimeType}
	}

	// Construct the system + user message with embedded images.
	tagsList := strings.Join(contentTagSlugs, `", "`)
	systemPrompt := `You are a content classifier for Twitch clips. Your task is to analyze video frames and select the most appropriate content tags. Only return tags you are confident about. Return exactly a JSON object: {"tags": ["tag1", "tag2"]}`

	userPrompt := fmt.Sprintf(
		`These are %d frames from a Twitch clip in the game/category "%s". Select 1-3 content tags that best describe what is happening. Choose only from: ["%s"]. Respond with JSON only.`,
		len(images), gameName, tagsList,
	)

	// Build the OpenAI-compatible multi-image message.
	contentParts := make([]map[string]interface{}, 0, len(images)+1)
	contentParts = append(contentParts, map[string]interface{}{
		"type": "text",
		"text": userPrompt,
	})
	for _, img := range images {
		b64 := base64.StdEncoding.EncodeToString(img.data)
		contentParts = append(contentParts, map[string]interface{}{
			"type": "image_url",
			"image_url": map[string]string{
				"url":    fmt.Sprintf("data:%s;base64,%s", img.mimeType, b64),
				"detail": "low",
			},
		})
	}

	reqBody := map[string]interface{}{
		"model": ts.model,
		"messages": []map[string]interface{}{
			{
				"role":    "system",
				"content": systemPrompt,
			},
			{
				"role":    "user",
				"content": contentParts,
			},
		},
		"response_format": map[string]string{"type": "json_object"},
		"max_tokens":      100,
		"temperature":     0.0,
	}

	jsonData, err := json.Marshal(reqBody)
	if err != nil {
		return nil, fmt.Errorf("marshalling vision API request: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, "POST", ts.apiURL, bytes.NewBuffer(jsonData))
	if err != nil {
		return nil, fmt.Errorf("creating vision API request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+ts.apiKey)

	// OpenRouter-specific headers for ranking and attribution
	if ts.provider == "openrouter" {
		req.Header.Set("HTTP-Referer", ts.siteURL)
		req.Header.Set("X-Title", ts.siteName)
	}

	resp, err := ts.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("vision API request failed: %w", err)
	}
	defer resp.Body.Close()

	bodyBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("reading vision API response: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("vision API returned status %d: %s", resp.StatusCode, string(bodyBytes))
	}

	return ts.parseVisionResponse(bodyBytes)
}

// parseVisionResponse extracts the content tags from the vision API response.
// It expects the OpenAI chat completion JSON format and treats the message
// content as a JSON object with a "tags" array.
func (ts *ThumbnailService) parseVisionResponse(body []byte) ([]string, error) {
	// OpenAI chat completions response structure.
	var apiResp struct {
		Choices []struct {
			Message struct {
				Content string `json:"content"`
			} `json:"message"`
		} `json:"choices"`
	}
	if err := json.Unmarshal(body, &apiResp); err != nil {
		return nil, fmt.Errorf("parsing vision API response: %w", err)
	}

	if len(apiResp.Choices) == 0 {
		return nil, fmt.Errorf("vision API returned no choices")
	}

	content := strings.TrimSpace(apiResp.Choices[0].Message.Content)
	if content == "" {
		return nil, fmt.Errorf("vision API returned empty content")
	}

	// The content should be a JSON object: {"tags": [...]}
	var tagResp struct {
		Tags []string `json:"tags"`
	}
	if err := json.Unmarshal([]byte(content), &tagResp); err != nil {
		return nil, fmt.Errorf("parsing vision API content: %w (content: %s)", err, content)
	}

	// Validate and normalize slugs.
	validSet := make(map[string]bool, len(contentTagSlugs))
	for _, s := range contentTagSlugs {
		validSet[s] = true
	}

	seen := make(map[string]bool, len(tagResp.Tags))
	tags := make([]string, 0, len(tagResp.Tags))
	for _, t := range tagResp.Tags {
		t = strings.ToLower(strings.TrimSpace(t))
		if t == "" {
			continue
		}
		if !validSet[t] {
			continue // Skip unknown tags
		}
		if seen[t] {
			continue
		}
		seen[t] = true
		tags = append(tags, t)
	}

	return tags, nil
}