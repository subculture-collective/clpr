package services

import (
	"context"
	"strings"
	"testing"

	"git.subcult.tv/subculture-collective/clpr/config"
)

func TestUploadValidatorRejectsUnsupportedMIMEType(t *testing.T) {
	called := false
	validator := NewUploadValidator(config.ClipSourceConfig{
		MaxDurationSeconds:     600,
		MaxUploadBytes:         1024,
		AllowedUploadMimeTypes: []string{"video/mp4", "video/webm", "video/quicktime"},
	}, func(ctx context.Context, filePath string) (int64, error) {
		called = true
		return 1, nil
	})

	_, err := validator.Validate(context.Background(), "image/png", 100, "/tmp/upload")
	if err == nil {
		t.Fatal("Validate() error = nil, want error")
	}
	if called {
		t.Fatal("ffprobe probe was called for unsupported MIME type")
	}
	if !strings.Contains(err.Error(), "Unsupported upload MIME type") {
		t.Fatalf("Validate() error = %v, want unsupported MIME type", err)
	}
}

func TestUploadValidatorRejectsOversizeUpload(t *testing.T) {
	called := false
	validator := NewUploadValidator(config.ClipSourceConfig{
		MaxDurationSeconds:     600,
		MaxUploadBytes:         1024,
		AllowedUploadMimeTypes: []string{"video/mp4"},
	}, func(ctx context.Context, filePath string) (int64, error) {
		called = true
		return 1, nil
	})

	_, err := validator.Validate(context.Background(), "video/mp4", 2048, "/tmp/upload")
	if err == nil {
		t.Fatal("Validate() error = nil, want error")
	}
	if called {
		t.Fatal("ffprobe probe was called for oversize upload")
	}
	if !strings.Contains(err.Error(), "Upload is too large") {
		t.Fatalf("Validate() error = %v, want size error", err)
	}
}

func TestUploadValidatorRejectsLongDuration(t *testing.T) {
	validator := NewUploadValidator(config.ClipSourceConfig{
		MaxDurationSeconds:     600,
		MaxUploadBytes:         1024,
		AllowedUploadMimeTypes: []string{"video/mp4"},
	}, func(ctx context.Context, filePath string) (int64, error) {
		return 601, nil
	})

	_, err := validator.Validate(context.Background(), "video/mp4", 100, "/tmp/upload")
	if err == nil {
		t.Fatal("Validate() error = nil, want error")
	}
	if !strings.Contains(err.Error(), "Upload is too long") {
		t.Fatalf("Validate() error = %v, want duration error", err)
	}
}

func TestUploadValidatorAcceptsAllowedMimeTypes(t *testing.T) {
	tests := []struct {
		name     string
		mimeType string
	}{
		{name: "mp4", mimeType: "video/mp4"},
		{name: "webm", mimeType: "video/webm"},
		{name: "quicktime with params", mimeType: "video/quicktime; codecs=jpeg"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			validator := NewUploadValidator(config.ClipSourceConfig{
				MaxDurationSeconds:     600,
				MaxUploadBytes:         1024,
				AllowedUploadMimeTypes: []string{"video/mp4", "video/webm", "video/quicktime"},
			}, func(ctx context.Context, filePath string) (int64, error) {
				return 599, nil
			})

			result, err := validator.Validate(context.Background(), tt.mimeType, 100, "/tmp/upload")
			if err != nil {
				t.Fatalf("Validate() error = %v, want nil", err)
			}
			if !result.DurationVerified {
				t.Fatal("DurationVerified = false, want true")
			}
			if got, want := result.DurationSeconds, int64(599); got != want {
				t.Fatalf("DurationSeconds = %d, want %d", got, want)
			}
			if got, want := result.FileSizeBytes, int64(100); got != want {
				t.Fatalf("FileSizeBytes = %d, want %d", got, want)
			}
			if got, want := result.MimeType, strings.Split(strings.ToLower(tt.mimeType), ";")[0]; got != want {
				t.Fatalf("MimeType = %q, want %q", got, want)
			}
		})
	}
}
