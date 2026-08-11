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
	"regexp"
	"strings"
	"time"

	"git.subcult.tv/subculture-collective/clpr/internal/models"
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

var nonWordTitleCharacters = regexp.MustCompile(`[^\p{L}\p{N}]+`)

// ClipThumbnailEnrichment is the model's evidence-bound interpretation of a
// Twitch clip thumbnail and Twitch-provided metadata.
type ClipThumbnailEnrichment struct {
	SuggestedTitle string   `json:"suggested_title"`
	Confidence     float64  `json:"confidence"`
	Basis          string   `json:"basis"`
	Evidence       []string `json:"evidence"`
	Tags           []string `json:"tags"`
}

// ShouldApplySuggestedTitle is the deterministic safety gate between a model
// suggestion and the public clip title. It only repairs weak automated titles;
// an informative Twitch or human title always wins.
func ShouldApplySuggestedTitle(clip *models.Clip, result *ClipThumbnailEnrichment) bool {
	if clip == nil || result == nil || clip.SubmittedByUserID != nil {
		return false
	}
	if !isWeakSourceTitle(clip.Title) || result.Confidence < 0.90 {
		return false
	}
	basis := strings.ToLower(strings.TrimSpace(result.Basis))
	if basis != "source_title" && basis != "visible" && basis != "metadata" && basis != "transcript" {
		return false
	}
	candidate := strings.TrimSpace(result.SuggestedTitle)
	if len([]rune(candidate)) < 8 || len([]rune(candidate)) > 100 {
		return false
	}
	if strings.ContainsAny(candidate, `"“”`) && !strings.ContainsAny(clip.Title, `"“”`) {
		return false
	}
	return !isWeakSourceTitle(candidate)
}

func isWeakSourceTitle(title string) bool {
	normalized := strings.ToLower(strings.TrimSpace(title))
	words := strings.Fields(nonWordTitleCharacters.ReplaceAllString(normalized, " "))
	if len(words) == 0 {
		return true
	}
	weak := map[string]bool{
		"clip": true, "twitch": true, "lol": true, "lmao": true,
		"wow": true, "omg": true, "wtf": true, "pog": true, "poggers": true,
	}
	if len(words) == 1 {
		return weak[words[0]] || len([]rune(words[0])) < 4
	}
	if len(words) == 2 {
		return weak[words[0]] && weak[words[1]]
	}
	return false
}

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

// AnalyzeClipThumbnail asks the configured vision model to enrich a clip using
// only fields Twitch exposes through its API and the public thumbnail URL.
func (ts *ThumbnailService) AnalyzeClipThumbnail(ctx context.Context, clip *models.Clip) (*ClipThumbnailEnrichment, error) {
	return ts.analyzeClip(ctx, clip, "")
}

// AnalyzeClipWithTranscript combines Twitch metadata, the public thumbnail,
// and authorized Whisper output. Spoken words are stronger evidence than a
// thumbnail but must still not be expanded into invented events.
func (ts *ThumbnailService) AnalyzeClipWithTranscript(ctx context.Context, clip *models.Clip, transcript string) (*ClipThumbnailEnrichment, error) {
	return ts.analyzeClip(ctx, clip, strings.TrimSpace(transcript))
}

func (ts *ThumbnailService) analyzeClip(ctx context.Context, clip *models.Clip, transcript string) (*ClipThumbnailEnrichment, error) {
	if !ts.Operational() {
		return nil, ErrVisionAPIUnavailable
	}
	if clip == nil || clip.ThumbnailURL == nil || strings.TrimSpace(*clip.ThumbnailURL) == "" {
		return nil, fmt.Errorf("clip thumbnail URL is required")
	}

	metadata := map[string]interface{}{
		"source_title":     clip.Title,
		"broadcaster_name": clip.BroadcasterName,
		"creator_name":     clip.CreatorName,
		"game_name":        gameName(clip),
	}
	if clip.Language != nil {
		metadata["language"] = *clip.Language
	}
	if clip.Duration != nil {
		metadata["duration_seconds"] = *clip.Duration
	}
	metadataJSON, err := json.Marshal(metadata)
	if err != nil {
		return nil, fmt.Errorf("marshalling clip metadata: %w", err)
	}

	tagsList := strings.Join(contentTagSlugs, `", "`)
	systemPrompt := `You enrich Twitch clip metadata using only the supplied Twitch metadata, authorized transcript when present, and thumbnail. Do not invent dialogue, identities, events, causes, or outcomes. A transcript is evidence of spoken words but not proof that an event occurred; a thumbnail is weak visual evidence. Prefer a cleaned version of the source title when it is informative. Return exactly one JSON object with suggested_title, confidence (0 to 1), basis (source_title, transcript, visible, metadata, or insufficient), evidence (short strings), and tags.`
	transcriptContext := "No authorized transcript is available."
	if transcript != "" {
		runes := []rune(transcript)
		if len(runes) > 4000 {
			runes = runes[:4000]
		}
		transcriptContext = "Authorized Whisper transcript: " + string(runes)
	}
	userPrompt := fmt.Sprintf(
		`Twitch metadata: %s. %s Suggest an accurate concise title and 0-3 tags chosen only from ["%s"]. If evidence is insufficient, preserve the source title and use basis "insufficient".`,
		string(metadataJSON), transcriptContext, tagsList,
	)

	reqBody := map[string]interface{}{
		"model": ts.model,
		"messages": []map[string]interface{}{
			{"role": "system", "content": systemPrompt},
			{"role": "user", "content": []map[string]interface{}{
				{"type": "text", "text": userPrompt},
				{"type": "image_url", "image_url": map[string]string{"url": *clip.ThumbnailURL, "detail": "low"}},
			}},
		},
		"response_format": map[string]string{"type": "json_object"},
		"max_tokens":      300,
		"temperature":     0.0,
	}

	body, err := ts.doVisionRequest(ctx, reqBody)
	if err != nil {
		return nil, err
	}
	result, err := ts.parseClipEnrichmentResponse(body)
	if err != nil {
		return nil, err
	}
	if result.Basis == "transcript" && transcript == "" {
		return nil, fmt.Errorf("vision API returned transcript basis without a transcript")
	}
	return result, nil
}

func (ts *ThumbnailService) doVisionRequest(ctx context.Context, reqBody map[string]interface{}) ([]byte, error) {
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
	if ts.provider == "openrouter" {
		req.Header.Set("HTTP-Referer", ts.siteURL)
		req.Header.Set("X-Title", ts.siteName)
	}
	resp, err := ts.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("vision API request failed: %w", err)
	}
	defer resp.Body.Close()
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("reading vision API response: %w", err)
	}
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("vision API returned status %d: %s", resp.StatusCode, string(body))
	}
	return body, nil
}

func (ts *ThumbnailService) parseClipEnrichmentResponse(body []byte) (*ClipThumbnailEnrichment, error) {
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
	var result ClipThumbnailEnrichment
	if err := json.Unmarshal([]byte(content), &result); err != nil {
		return nil, fmt.Errorf("parsing clip enrichment content: %w (content: %s)", err, content)
	}
	result.Basis = strings.ToLower(strings.TrimSpace(result.Basis))
	validBasis := result.Basis == "source_title" || result.Basis == "transcript" ||
		result.Basis == "visible" || result.Basis == "metadata" || result.Basis == "insufficient"
	if !validBasis {
		return nil, fmt.Errorf("vision API returned unsupported evidence basis %q", result.Basis)
	}
	if result.Confidence < 0 || result.Confidence > 1 {
		return nil, fmt.Errorf("vision API returned confidence outside 0..1: %f", result.Confidence)
	}
	result.SuggestedTitle = strings.TrimSpace(result.SuggestedTitle)
	result.Tags = filterContentTags(result.Tags)
	return &result, nil
}

func gameName(clip *models.Clip) string {
	if clip != nil && clip.GameName != nil {
		return *clip.GameName
	}
	return "Unknown Game"
}

func filterContentTags(input []string) []string {
	validSet := make(map[string]bool, len(contentTagSlugs))
	for _, slug := range contentTagSlugs {
		validSet[slug] = true
	}
	seen := make(map[string]bool, len(input))
	result := make([]string, 0, len(input))
	for _, tag := range input {
		tag = strings.ToLower(strings.TrimSpace(tag))
		if tag != "" && validSet[tag] && !seen[tag] {
			seen[tag] = true
			result = append(result, tag)
		}
	}
	return result
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
