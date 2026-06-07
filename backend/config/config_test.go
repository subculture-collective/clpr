package config

import (
	"reflect"
	"testing"
)

func TestLoadClipDefaults(t *testing.T) {
	t.Chdir(t.TempDir())

	for _, key := range []string{
		"CLIP_MAX_DURATION_SECONDS",
		"CLIP_RECOMMENDED_DURATION_SECONDS",
		"CLIP_MAX_UPLOAD_BYTES",
		"CLIP_ALLOWED_UPLOAD_MIME_TYPES",
		"CLIP_REQUIRE_MODERATION_FOR_UPLOAD",
		"CLIP_STORAGE_PROVIDER",
		"CLIP_STORAGE_ENDPOINT",
		"CLIP_STORAGE_BUCKET",
		"CLIP_STORAGE_REGION",
		"CLIP_STORAGE_ACCESS_KEY",
		"CLIP_STORAGE_SECRET_KEY",
		"CLIP_STORAGE_FORCE_PATH_STYLE",
		"CLIP_STORAGE_PUBLIC_BASE_URL",
	} {
		t.Setenv(key, "")
	}

	cfg, err := Load()
	if err != nil {
		t.Fatalf("Load() error = %v", err)
	}

	if got, want := cfg.ClipSource.MaxDurationSeconds, int64(600); got != want {
		t.Fatalf("MaxDurationSeconds = %d, want %d", got, want)
	}
	if got, want := cfg.ClipSource.RecommendedDurationSeconds, int64(420); got != want {
		t.Fatalf("RecommendedDurationSeconds = %d, want %d", got, want)
	}
	if got, want := cfg.ClipSource.MaxUploadBytes, int64(1073741824); got != want {
		t.Fatalf("MaxUploadBytes = %d, want %d", got, want)
	}
	if got, want := cfg.ClipSource.AllowedUploadMimeTypes, []string{"video/mp4", "video/webm", "video/quicktime"}; !reflect.DeepEqual(got, want) {
		t.Fatalf("AllowedUploadMimeTypes = %#v, want %#v", got, want)
	}
	if !cfg.ClipSource.RequireModerationForUpload {
		t.Fatal("RequireModerationForUpload = false, want true")
	}

	if got, want := cfg.ClipStorage.Provider, "s3"; got != want {
		t.Fatalf("Provider = %q, want %q", got, want)
	}
	if got, want := cfg.ClipStorage.Endpoint, ""; got != want {
		t.Fatalf("Endpoint = %q, want %q", got, want)
	}
	if got, want := cfg.ClipStorage.Bucket, ""; got != want {
		t.Fatalf("Bucket = %q, want %q", got, want)
	}
	if got, want := cfg.ClipStorage.Region, "us-east-1"; got != want {
		t.Fatalf("Region = %q, want %q", got, want)
	}
	if got, want := cfg.ClipStorage.AccessKey, ""; got != want {
		t.Fatalf("AccessKey = %q, want %q", got, want)
	}
	if got, want := cfg.ClipStorage.SecretKey, ""; got != want {
		t.Fatalf("SecretKey = %q, want %q", got, want)
	}
	if !cfg.ClipStorage.ForcePathStyle {
		t.Fatal("ForcePathStyle = false, want true")
	}
	if got, want := cfg.ClipStorage.PublicBaseURL, ""; got != want {
		t.Fatalf("PublicBaseURL = %q, want %q", got, want)
	}
}

func TestLoadClipEnvOverrides(t *testing.T) {
	t.Chdir(t.TempDir())

	t.Setenv("CLIP_MAX_DURATION_SECONDS", "901")
	t.Setenv("CLIP_RECOMMENDED_DURATION_SECONDS", "777")
	t.Setenv("CLIP_MAX_UPLOAD_BYTES", "2048")
	t.Setenv("CLIP_ALLOWED_UPLOAD_MIME_TYPES", "video/mp4, video/mkv , video/webm")
	t.Setenv("CLIP_REQUIRE_MODERATION_FOR_UPLOAD", "false")
	t.Setenv("CLIP_STORAGE_PROVIDER", "aws")
	t.Setenv("CLIP_STORAGE_ENDPOINT", "https://storage.example.com")
	t.Setenv("CLIP_STORAGE_BUCKET", "clips-bucket")
	t.Setenv("CLIP_STORAGE_REGION", "eu-west-1")
	t.Setenv("CLIP_STORAGE_ACCESS_KEY", "access")
	t.Setenv("CLIP_STORAGE_SECRET_KEY", "secret")
	t.Setenv("CLIP_STORAGE_FORCE_PATH_STYLE", "false")
	t.Setenv("CLIP_STORAGE_PUBLIC_BASE_URL", "https://cdn.example.com/clips")

	cfg, err := Load()
	if err != nil {
		t.Fatalf("Load() error = %v", err)
	}

	if got, want := cfg.ClipSource.MaxDurationSeconds, int64(901); got != want {
		t.Fatalf("MaxDurationSeconds = %d, want %d", got, want)
	}
	if got, want := cfg.ClipSource.RecommendedDurationSeconds, int64(777); got != want {
		t.Fatalf("RecommendedDurationSeconds = %d, want %d", got, want)
	}
	if got, want := cfg.ClipSource.MaxUploadBytes, int64(2048); got != want {
		t.Fatalf("MaxUploadBytes = %d, want %d", got, want)
	}
	if got, want := cfg.ClipSource.AllowedUploadMimeTypes, []string{"video/mp4", "video/mkv", "video/webm"}; !reflect.DeepEqual(got, want) {
		t.Fatalf("AllowedUploadMimeTypes = %#v, want %#v", got, want)
	}
	if cfg.ClipSource.RequireModerationForUpload {
		t.Fatal("RequireModerationForUpload = true, want false")
	}

	if got, want := cfg.ClipStorage.Provider, "aws"; got != want {
		t.Fatalf("Provider = %q, want %q", got, want)
	}
	if got, want := cfg.ClipStorage.Endpoint, "https://storage.example.com"; got != want {
		t.Fatalf("Endpoint = %q, want %q", got, want)
	}
	if got, want := cfg.ClipStorage.Bucket, "clips-bucket"; got != want {
		t.Fatalf("Bucket = %q, want %q", got, want)
	}
	if got, want := cfg.ClipStorage.Region, "eu-west-1"; got != want {
		t.Fatalf("Region = %q, want %q", got, want)
	}
	if got, want := cfg.ClipStorage.AccessKey, "access"; got != want {
		t.Fatalf("AccessKey = %q, want %q", got, want)
	}
	if got, want := cfg.ClipStorage.SecretKey, "secret"; got != want {
		t.Fatalf("SecretKey = %q, want %q", got, want)
	}
	if cfg.ClipStorage.ForcePathStyle {
		t.Fatal("ForcePathStyle = true, want false")
	}
	if got, want := cfg.ClipStorage.PublicBaseURL, "https://cdn.example.com/clips"; got != want {
		t.Fatalf("PublicBaseURL = %q, want %q", got, want)
	}
}
