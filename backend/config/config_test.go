package config

import (
	"reflect"
	"testing"
)

func TestLoadClipConfigDefaults(t *testing.T) {
	t.Setenv("CLIP_MAX_DURATION_SECONDS", "")
	t.Setenv("CLIP_RECOMMENDED_DURATION_SECONDS", "")
	t.Setenv("CLIP_MAX_UPLOAD_BYTES", "")
	t.Setenv("CLIP_ALLOWED_UPLOAD_MIME_TYPES", "")
	t.Setenv("CLIP_REQUIRE_MODERATION_FOR_UPLOAD", "")
	t.Setenv("CLIP_STORAGE_PROVIDER", "")
	t.Setenv("CLIP_STORAGE_ENDPOINT", "")
	t.Setenv("CLIP_STORAGE_BUCKET", "")
	t.Setenv("CLIP_STORAGE_REGION", "")
	t.Setenv("CLIP_STORAGE_ACCESS_KEY", "")
	t.Setenv("CLIP_STORAGE_SECRET_KEY", "")
	t.Setenv("CLIP_STORAGE_FORCE_PATH_STYLE", "")
	t.Setenv("CLIP_STORAGE_PUBLIC_BASE_URL", "")
	t.Setenv("CLIP_MEDIA_PUBLIC_BASE_URL", "")

	cfg, err := Load()
	if err != nil {
		t.Fatalf("Load() error = %v", err)
	}

	if cfg.Clip.MaxDurationSeconds != 60 {
		t.Fatalf("MaxDurationSeconds = %d, want 60", cfg.Clip.MaxDurationSeconds)
	}
	if cfg.Clip.RecommendedDurationSeconds != 60 {
		t.Fatalf("RecommendedDurationSeconds = %d, want 60", cfg.Clip.RecommendedDurationSeconds)
	}
	if cfg.Clip.MaxUploadBytes != 104857600 {
		t.Fatalf("MaxUploadBytes = %d, want 104857600", cfg.Clip.MaxUploadBytes)
	}
	if !reflect.DeepEqual(cfg.Clip.AllowedUploadMimeTypes, []string{"video/mp4", "video/webm", "video/quicktime"}) {
		t.Fatalf("AllowedUploadMimeTypes = %#v, want default list", cfg.Clip.AllowedUploadMimeTypes)
	}
	if cfg.Clip.RequireModerationForUpload {
		t.Fatal("RequireModerationForUpload = true, want false")
	}
	if cfg.Clip.StorageProvider != "local" {
		t.Fatalf("StorageProvider = %q, want local", cfg.Clip.StorageProvider)
	}
	if cfg.Clip.StorageRegion != "us-east-1" {
		t.Fatalf("StorageRegion = %q, want us-east-1", cfg.Clip.StorageRegion)
	}
}

func TestLoadClipConfigOverrides(t *testing.T) {
	t.Setenv("CLIP_MAX_DURATION_SECONDS", "90")
	t.Setenv("CLIP_RECOMMENDED_DURATION_SECONDS", "45")
	t.Setenv("CLIP_MAX_UPLOAD_BYTES", "250000000")
	t.Setenv("CLIP_ALLOWED_UPLOAD_MIME_TYPES", "video/mp4, video/ogg")
	t.Setenv("CLIP_REQUIRE_MODERATION_FOR_UPLOAD", "true")
	t.Setenv("CLIP_STORAGE_PROVIDER", "s3")
	t.Setenv("CLIP_STORAGE_ENDPOINT", "https://s3.example.invalid")
	t.Setenv("CLIP_STORAGE_BUCKET", "clips-bucket")
	t.Setenv("CLIP_STORAGE_REGION", "eu-west-1")
	t.Setenv("CLIP_STORAGE_ACCESS_KEY", "access-key")
	t.Setenv("CLIP_STORAGE_SECRET_KEY", "secret-key")
	t.Setenv("CLIP_STORAGE_FORCE_PATH_STYLE", "true")
	t.Setenv("CLIP_STORAGE_PUBLIC_BASE_URL", "https://cdn.example.invalid/clips")
	t.Setenv("CLIP_MEDIA_PUBLIC_BASE_URL", "https://clpr.example.invalid/api/v1/clips")

	cfg, err := Load()
	if err != nil {
		t.Fatalf("Load() error = %v", err)
	}

	if cfg.Clip.MaxDurationSeconds != 90 {
		t.Fatalf("MaxDurationSeconds = %d, want 90", cfg.Clip.MaxDurationSeconds)
	}
	if cfg.Clip.RecommendedDurationSeconds != 45 {
		t.Fatalf("RecommendedDurationSeconds = %d, want 45", cfg.Clip.RecommendedDurationSeconds)
	}
	if cfg.Clip.MaxUploadBytes != 250000000 {
		t.Fatalf("MaxUploadBytes = %d, want 250000000", cfg.Clip.MaxUploadBytes)
	}
	if !reflect.DeepEqual(cfg.Clip.AllowedUploadMimeTypes, []string{"video/mp4", "video/ogg"}) {
		t.Fatalf("AllowedUploadMimeTypes = %#v, want trimmed list", cfg.Clip.AllowedUploadMimeTypes)
	}
	if !cfg.Clip.RequireModerationForUpload {
		t.Fatal("RequireModerationForUpload = false, want true")
	}
	if cfg.Clip.StorageProvider != "s3" || cfg.Clip.StorageEndpoint != "https://s3.example.invalid" || cfg.Clip.StorageBucket != "clips-bucket" {
		t.Fatal("storage config did not load expected overrides")
	}
	if cfg.Clip.StorageRegion != "eu-west-1" || cfg.Clip.StorageAccessKey != "access-key" || cfg.Clip.StorageSecretKey != "secret-key" {
		t.Fatal("storage credentials/region did not load expected overrides")
	}
	if !cfg.Clip.StorageForcePathStyle {
		t.Fatal("StorageForcePathStyle = false, want true")
	}
	if cfg.Clip.StoragePublicBaseURL != "https://cdn.example.invalid/clips" {
		t.Fatalf("StoragePublicBaseURL = %q, want override", cfg.Clip.StoragePublicBaseURL)
	}
	if cfg.Clip.MediaPublicBaseURL != "https://clpr.example.invalid/api/v1/clips" {
		t.Fatalf("MediaPublicBaseURL = %q, want override", cfg.Clip.MediaPublicBaseURL)
	}
}
