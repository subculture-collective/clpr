package services

import (
	"context"
	"fmt"
	"math"
	"mime"
	"os/exec"
	"strconv"
	"strings"

	"git.subcult.tv/subculture-collective/clpr/config"
)

// UploadValidationResult captures normalized upload metadata after validation.
type UploadValidationResult struct {
	DurationSeconds  int64
	DurationVerified bool
	MimeType         string
	FileSizeBytes    int64
}

// DurationProbeFunc inspects a file and returns its duration in seconds.
type DurationProbeFunc func(ctx context.Context, filePath string) (int64, error)

// UploadValidator validates hosted clip uploads before persistence.
type UploadValidator struct {
	cfg          config.ClipSourceConfig
	probe        DurationProbeFunc
	allowedTypes map[string]struct{}
}

// MaxUploadBytes returns the configured upload size limit.
func (v *UploadValidator) MaxUploadBytes() int64 {
	if v == nil {
		return 0
	}
	return v.cfg.MaxUploadBytes
}

// NewUploadValidator creates an UploadValidator.
func NewUploadValidator(cfg config.ClipSourceConfig, probe DurationProbeFunc) *UploadValidator {
	allowed := make(map[string]struct{}, len(cfg.AllowedUploadMimeTypes))
	for _, mimeType := range cfg.AllowedUploadMimeTypes {
		if normalized := normalizeUploadMimeType(mimeType); normalized != "" {
			allowed[normalized] = struct{}{}
		}
	}
	if probe == nil {
		probe = probeUploadDurationWithFFProbe
	}

	return &UploadValidator{cfg: cfg, probe: probe, allowedTypes: allowed}
}

// Validate checks MIME type, size, and duration limits for a hosted upload.
func (v *UploadValidator) Validate(ctx context.Context, mimeType string, fileSizeBytes int64, filePath string) (*UploadValidationResult, error) {
	if v == nil {
		return nil, fmt.Errorf("upload validator is not configured")
	}

	normalizedMimeType := normalizeUploadMimeType(mimeType)
	if normalizedMimeType == "" {
		return nil, &ValidationError{Field: "mime_type", Message: "Unsupported upload MIME type"}
	}
	if _, ok := v.allowedTypes[normalizedMimeType]; !ok {
		return nil, &ValidationError{Field: "mime_type", Message: fmt.Sprintf("Unsupported upload MIME type: %s", normalizedMimeType)}
	}

	if fileSizeBytes > v.cfg.MaxUploadBytes {
		return nil, &ValidationError{Field: "file_size_bytes", Message: fmt.Sprintf("Upload is too large (maximum %d bytes)", v.cfg.MaxUploadBytes)}
	}

	durationSeconds, err := v.probe(ctx, filePath)
	if err != nil {
		return nil, &ValidationError{Field: "duration_seconds", Message: "Unable to inspect uploaded video duration"}
	}
	if durationSeconds > v.cfg.MaxDurationSeconds {
		return nil, &ValidationError{Field: "duration_seconds", Message: fmt.Sprintf("Upload is too long (maximum %d seconds)", v.cfg.MaxDurationSeconds)}
	}

	return &UploadValidationResult{
		DurationSeconds:  durationSeconds,
		DurationVerified: true,
		MimeType:         normalizedMimeType,
		FileSizeBytes:    fileSizeBytes,
	}, nil
}

func normalizeUploadMimeType(raw string) string {
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return ""
	}
	mediaType, _, err := mime.ParseMediaType(raw)
	if err != nil {
		return ""
	}
	return strings.ToLower(strings.TrimSpace(mediaType))
}

func probeUploadDurationWithFFProbe(ctx context.Context, filePath string) (int64, error) {
	output, err := exec.CommandContext(ctx, "ffprobe",
		"-v", "error",
		"-show_entries", "format=duration",
		"-of", "default=noprint_wrappers=1:nokey=1",
		filePath,
	).CombinedOutput()
	if err != nil {
		return 0, fmt.Errorf("ffprobe failed: %w: %s", err, strings.TrimSpace(string(output)))
	}

	trimmed := strings.TrimSpace(string(output))
	if trimmed == "" {
		return 0, fmt.Errorf("ffprobe returned empty duration")
	}

	seconds, err := strconv.ParseFloat(trimmed, 64)
	if err != nil {
		return 0, fmt.Errorf("parse ffprobe duration: %w", err)
	}
	if seconds < 0 {
		return 0, fmt.Errorf("ffprobe returned negative duration")
	}

	return int64(math.Ceil(seconds)), nil
}
