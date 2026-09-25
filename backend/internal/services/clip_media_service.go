package services

import (
	"context"
	"fmt"
	"os"
	"os/exec"
)

// ExtractAudioFromURL streams remote media through ffmpeg and writes only the
// normalized WAV needed by Whisper. The source video is never persisted.
func ExtractAudioFromURL(ctx context.Context, ffmpegPath, mediaURL, workDir string) (string, error) {
	if err := os.MkdirAll(workDir, 0755); err != nil {
		return "", fmt.Errorf("creating work dir %s: %w", workDir, err)
	}
	temp, err := os.CreateTemp(workDir, "clip-audio-*.wav")
	if err != nil {
		return "", fmt.Errorf("creating audio temp file: %w", err)
	}
	wavPath := temp.Name()
	if err := temp.Close(); err != nil {
		os.Remove(wavPath)
		return "", fmt.Errorf("closing audio temp file: %w", err)
	}

	cmd := exec.CommandContext(ctx, ffmpegPath,
		"-nostdin",
		"-v", "error",
		"-i", mediaURL,
		"-vn",
		"-acodec", "pcm_s16le",
		"-ar", "16000",
		"-ac", "1",
		"-y",
		wavPath,
	)
	output, err := cmd.CombinedOutput()
	if err != nil {
		os.Remove(wavPath)
		return "", fmt.Errorf("ffmpeg streamed audio extraction failed: %w\n%s", err, string(output))
	}
	info, err := os.Stat(wavPath)
	if err != nil || info.Size() <= 44 {
		os.Remove(wavPath)
		return "", fmt.Errorf("ffmpeg produced empty audio")
	}
	return wavPath, nil
}
