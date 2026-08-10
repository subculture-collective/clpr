package services

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

// deriveVideoURLFromThumbnail converts a Twitch clip thumbnail URL into a
// downloadable video URL by replacing the preview suffix with .mp4.
//
// Thumbnail URL: https://clips-media-assets2.twitch.tv/AT-cm-123456789-preview-480x272.jpg
// Video URL:     https://clips-media-assets2.twitch.tv/AT-cm-123456789.mp4
func deriveVideoURLFromThumbnail(thumbnailURL string) string {
	// Remove known preview size suffixes
	suffixes := []string{
		"-preview-480x272.jpg",
		"-preview-260x147.jpg",
		"-preview-130x73.jpg",
		"-preview-86x45.jpg",
	}
	for _, suffix := range suffixes {
		if strings.HasSuffix(thumbnailURL, suffix) {
			return strings.TrimSuffix(thumbnailURL, suffix) + ".mp4"
		}
	}
	// Fallback: replace any trailing .jpg/.png with .mp4
	ext := filepath.Ext(thumbnailURL)
	if ext == ".jpg" || ext == ".png" || ext == ".jpeg" {
		return thumbnailURL[:len(thumbnailURL)-len(ext)] + ".mp4"
	}
	return thumbnailURL
}

// DownloadClipVideo downloads a clip's video from its thumbnail URL
// (deriving the actual mp4 URL) into a temp file under workDir.
// Returns the path to the downloaded video file.
func DownloadClipVideo(ctx context.Context, thumbnailURL string, workDir string) (string, error) {
	videoURL := deriveVideoURLFromThumbnail(thumbnailURL)

	if err := os.MkdirAll(workDir, 0755); err != nil {
		return "", fmt.Errorf("creating work dir %s: %w", workDir, err)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, videoURL, nil)
	if err != nil {
		return "", fmt.Errorf("creating download request: %w", err)
	}
	// Twitch requires a User-Agent for CDN requests
	req.Header.Set("User-Agent", "clpr/1.0 (media processing)")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return "", fmt.Errorf("downloading clip video from %s: %w", videoURL, err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("clip video download returned status %d", resp.StatusCode)
	}

	// Create a temp file for the video
	f, err := os.CreateTemp(workDir, "clip-*.mp4")
	if err != nil {
		return "", fmt.Errorf("creating temp file: %w", err)
	}

	written, err := io.Copy(f, resp.Body)
	closeErr := f.Close()
	if err != nil {
		os.Remove(f.Name())
		return "", fmt.Errorf("writing clip video: %w", err)
	}
	if closeErr != nil {
		os.Remove(f.Name())
		return "", fmt.Errorf("closing temp file: %w", closeErr)
	}
	if written == 0 {
		os.Remove(f.Name())
		return "", fmt.Errorf("downloaded empty clip video")
	}

	return f.Name(), nil
}

// ExtractAudio extracts WAV audio (16kHz mono PCM) from a video file using
// ffmpeg.  Returns the path to the extracted WAV file.
func ExtractAudio(ctx context.Context, ffmpegPath string, videoPath string, workDir string) (string, error) {
	if err := os.MkdirAll(workDir, 0755); err != nil {
		return "", fmt.Errorf("creating work dir %s: %w", workDir, err)
	}

	baseName := filepath.Base(videoPath)
	ext := filepath.Ext(baseName)
	stem := strings.TrimSuffix(baseName, ext)
	wavPath := filepath.Join(workDir, stem+".wav")

	cmd := exec.CommandContext(ctx, ffmpegPath,
		"-i", videoPath,
		"-vn",                   // no video
		"-acodec", "pcm_s16le",  // 16-bit PCM
		"-ar", "16000",          // 16kHz sample rate
		"-ac", "1",              // mono
		"-y",                    // overwrite
		wavPath,
	)

	output, err := cmd.CombinedOutput()
	if err != nil {
		return "", fmt.Errorf("ffmpeg audio extraction failed: %w\n%s", err, string(output))
	}

	return wavPath, nil
}