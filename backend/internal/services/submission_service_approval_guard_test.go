package services

import (
	"context"
	"encoding/json"
	"testing"
	"time"

	"git.subcult.tv/subculture-collective/clpr/internal/models"
	"git.subcult.tv/subculture-collective/clpr/internal/storage"
	"github.com/google/uuid"
)

type fakeCreatorModerationChecker struct {
	allowed bool
	message string
	err     error
}

func (f fakeCreatorModerationChecker) CanInteract(ctx context.Context, creatorID uuid.UUID, userID uuid.UUID) (bool, string, error) {
	return f.allowed, f.message, f.err
}

func (f fakeCreatorModerationChecker) CanSubmit(ctx context.Context, creatorID uuid.UUID, userID uuid.UUID) (bool, string, error) {
	return f.allowed, f.message, f.err
}

func (f fakeCreatorModerationChecker) CanComment(ctx context.Context, creatorID uuid.UUID, userID uuid.UUID) (bool, string, error) {
	return f.allowed, f.message, f.err
}

func TestSubmissionService_EnsureSubmissionCanBeApproved(t *testing.T) {
	service := &SubmissionService{}

	tests := []struct {
		name    string
		sub     *models.ClipSubmission
		wantErr string
	}{
		{
			name:    "twitch submission is allowed",
			sub:     &models.ClipSubmission{SourceType: string(SourceTypeTwitch)},
			wantErr: "",
		},
		{
			name:    "external submission is allowed",
			sub:     &models.ClipSubmission{SourceType: string(SourceTypeExternal)},
			wantErr: "",
		},
		{
			name:    "upload submission is allowed",
			sub:     &models.ClipSubmission{SourceType: "upload"},
			wantErr: "",
		},
		{
			name:    "legacy empty source type is blocked",
			sub:     &models.ClipSubmission{SourceType: ""},
			wantErr: "submission source type is required",
		},
		{
			name:    "unknown source type is blocked",
			sub:     &models.ClipSubmission{SourceType: "vimeo"},
			wantErr: "unsupported submission source type \"vimeo\"",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := service.ensureSubmissionCanBeApproved(tt.sub)
			if tt.wantErr == "" {
				if err != nil {
					t.Fatalf("ensureSubmissionCanBeApproved() error = %v, want nil", err)
				}
				return
			}
			if err == nil {
				t.Fatalf("ensureSubmissionCanBeApproved() error = nil, want %q", tt.wantErr)
			}
			if err.Error() != tt.wantErr {
				t.Fatalf("ensureSubmissionCanBeApproved() error = %q, want %q", err.Error(), tt.wantErr)
			}
		})
	}
}

type fakeUploadStorage struct {
	publicURL string
	copyFrom  string
	copyTo    string
	copyType  string
}

func (f fakeUploadStorage) PutObject(ctx context.Context, input storage.PutObjectInput) (storage.PutObjectResult, error) {
	return storage.PutObjectResult{}, nil
}

func (f *fakeUploadStorage) CopyObject(ctx context.Context, srcKey, dstKey, contentType string) (storage.PutObjectResult, error) {
	f.copyFrom = srcKey
	f.copyTo = dstKey
	f.copyType = contentType
	return storage.PutObjectResult{Provider: "s3", Bucket: "bucket", Key: dstKey, PublicURL: f.PublicURL(dstKey)}, nil
}

func (f fakeUploadStorage) DeleteObject(ctx context.Context, key string) error { return nil }
func (f fakeUploadStorage) PublicURL(key string) string                        { return f.publicURL }

func TestSubmissionService_CreateClipFromSubmission_SourceAware(t *testing.T) {
	publicURL := "https://cdn.example.com/uploads/public/video.mp4"
	storageBackend := &fakeUploadStorage{publicURL: publicURL}

	tests := []struct {
		name          string
		submission    *models.ClipSubmission
		wantSourceTyp string
		check         func(t *testing.T, clip *models.Clip)
	}{
		{
			name: "twitch submission",
			submission: &models.ClipSubmission{
				ID:               uuid.New(),
				UserID:           uuid.New(),
				TwitchClipID:     "Clip123",
				TwitchClipURL:    "https://clips.twitch.tv/Clip123",
				Title:            ptrString("Original title"),
				CreatorName:      ptrString("Creator"),
				BroadcasterName:  ptrString("Broadcaster"),
				CreatedAt:        time.Now().Add(-time.Minute),
				SourceType:       string(SourceTypeTwitch),
				SourcePlatform:   string(SourcePlatformTwitch),
				SourceMetadata:   json.RawMessage(`{}`),
				Duration:         float64Ptr(12.5),
				DurationVerified: true,
			},
			wantSourceTyp: string(SourceTypeTwitch),
			check: func(t *testing.T, clip *models.Clip) {
				if clip.EmbedURL != "https://clips.twitch.tv/embed?clip=Clip123" {
					t.Fatalf("EmbedURL = %q, want twitch embed", clip.EmbedURL)
				}
				if clip.VideoURL != nil {
					t.Fatalf("VideoURL = %v, want nil", clip.VideoURL)
				}
			},
		},
		{
			name: "external submission",
			submission: &models.ClipSubmission{
				ID:               uuid.New(),
				UserID:           uuid.New(),
				TwitchClipID:     "external:youtube:legacy-id",
				TwitchClipURL:    "https://www.youtube.com/watch?v=abc123",
				Title:            ptrString("External title"),
				CustomTitle:      ptrString("Custom title"),
				CreatorName:      ptrString("Author Name"),
				BroadcasterName:  ptrString("Channel Name"),
				CreatedAt:        time.Now().Add(-time.Minute),
				SourceType:       string(SourceTypeExternal),
				SourcePlatform:   string(SourcePlatformYouTube),
				SourceURL:        ptrString("https://www.youtube.com/watch?v=abc123"),
				SourceID:         ptrString("abc123"),
				SourceMetadata:   json.RawMessage(`{"embed_url":"https://player.example/embed/abc123"}`),
				DurationSeconds:  ptrInt(42),
				DurationVerified: true,
			},
			wantSourceTyp: string(SourceTypeExternal),
			check: func(t *testing.T, clip *models.Clip) {
				if clip.StreamSource == nil || *clip.StreamSource != "external" {
					t.Fatalf("StreamSource = %v, want external", clip.StreamSource)
				}
				if clip.SourcePlatform != "youtube" {
					t.Fatalf("SourcePlatform = %q, want youtube", clip.SourcePlatform)
				}
				if clip.EmbedURL != "https://player.example/embed/abc123" {
					t.Fatalf("EmbedURL = %q, want metadata embed url", clip.EmbedURL)
				}
				if clip.VideoURL != nil {
					t.Fatalf("VideoURL = %v, want nil", clip.VideoURL)
				}
				if clip.SourceURL == nil || *clip.SourceURL != "https://www.youtube.com/watch?v=abc123" {
					t.Fatalf("SourceURL = %v, want normalized url", clip.SourceURL)
				}
				if clip.SourceID == nil || *clip.SourceID != "abc123" {
					t.Fatalf("SourceID = %v, want abc123", clip.SourceID)
				}
				if clip.TwitchClipID != "external:youtube:legacy-id" {
					t.Fatalf("TwitchClipID = %q, want preserved external legacy id", clip.TwitchClipID)
				}
			},
		},
		{
			name: "upload submission",
			submission: &models.ClipSubmission{
				ID:               uuid.New(),
				UserID:           uuid.New(),
				TwitchClipID:     "upload:sub-123",
				TwitchClipURL:    "",
				Title:            ptrString("Upload title"),
				CustomTitle:      ptrString("Upload custom title"),
				CreatedAt:        time.Now().Add(-time.Minute),
				SourceType:       "upload",
				SourcePlatform:   "upload",
				SourceID:         ptrString("sub-123"),
				SourceMetadata:   json.RawMessage(`{"storage_key":"uploads/pending/u/s/original.mp4"}`),
				DurationSeconds:  ptrInt(55),
				DurationVerified: true,
				StorageProvider:  ptrString("s3"),
				StorageBucket:    ptrString("bucket"),
				StorageKey:       ptrString("uploads/pending/u/s/original.mp4"),
				OriginalFilename: ptrString("clip.mp4"),
				MimeType:         ptrString("video/mp4"),
				FileSizeBytes:    ptrInt64(2048),
			},
			wantSourceTyp: "upload",
			check: func(t *testing.T, clip *models.Clip) {
				if clip.StreamSource == nil || *clip.StreamSource != "upload" {
					t.Fatalf("StreamSource = %v, want upload", clip.StreamSource)
				}
				if clip.VideoURL == nil || *clip.VideoURL != publicURL {
					t.Fatalf("VideoURL = %v, want public url", clip.VideoURL)
				}
				if clip.SourceURL == nil || *clip.SourceURL != publicURL {
					t.Fatalf("SourceURL = %v, want public url", clip.SourceURL)
				}
				if clip.StorageProvider == nil || *clip.StorageProvider != "s3" {
					t.Fatalf("StorageProvider = %v, want s3", clip.StorageProvider)
				}
				if clip.StorageKey == nil || *clip.StorageKey != "uploads/public/u/s/original.mp4" {
					t.Fatalf("StorageKey = %v, want public storage key", clip.StorageKey)
				}
				if clip.TwitchClipID != "upload:sub-123" {
					t.Fatalf("TwitchClipID = %q, want upload legacy id", clip.TwitchClipID)
				}
				if len(clip.TwitchClipID) > 100 {
					t.Fatalf("TwitchClipID length = %d, want <= 100", len(clip.TwitchClipID))
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var captured *models.Clip
			service := &SubmissionService{
				clipStorage: storageBackend,
				clipCreateFn: func(ctx context.Context, clip *models.Clip) error {
					captured = clip
					return nil
				},
			}

			clipID, err := service.createClipFromSubmission(context.Background(), tt.submission)
			if err != nil {
				t.Fatalf("createClipFromSubmission() error = %v", err)
			}
			if clipID == uuid.Nil {
				t.Fatalf("createClipFromSubmission() returned nil clip id")
			}
			if captured == nil {
				t.Fatalf("expected clip to be captured")
			}
			if captured.SourceType != tt.wantSourceTyp {
				t.Fatalf("SourceType = %q, want %q", captured.SourceType, tt.wantSourceTyp)
			}
			tt.check(t, captured)
			if tt.submission.SourceType == "upload" {
				if storageBackend.copyFrom != "uploads/pending/u/s/original.mp4" {
					t.Fatalf("CopyObject source = %q, want pending key", storageBackend.copyFrom)
				}
				if storageBackend.copyTo != "uploads/public/u/s/original.mp4" {
					t.Fatalf("CopyObject destination = %q, want public key", storageBackend.copyTo)
				}
				if storageBackend.copyType != "video/mp4" {
					t.Fatalf("CopyObject content type = %q, want video/mp4", storageBackend.copyType)
				}
			}
		})
	}
}

func TestSubmissionService_CreateClipFromSubmission_UnsupportedSource(t *testing.T) {
	service := &SubmissionService{}
	_, err := service.createClipFromSubmission(context.Background(), &models.ClipSubmission{SourceType: "vimeo"})
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	if err.Error() != "unsupported submission source type \"vimeo\"" {
		t.Fatalf("error = %q, want unsupported submission source type", err.Error())
	}
}

func TestSubmissionService_ShouldAutoUpvoteClaimedClip(t *testing.T) {
	creatorID := uuid.New()
	userID := uuid.New()

	t.Run("skips when interaction is denied", func(t *testing.T) {
		service := &SubmissionService{creatorModeration: fakeCreatorModerationChecker{allowed: false, message: "denied"}}
		if service.shouldAutoUpvoteClaimedClip(context.Background(), &creatorID, userID) {
			t.Fatal("expected auto-upvote to be skipped when interaction is denied")
		}
	})

	t.Run("allows when interaction is permitted", func(t *testing.T) {
		service := &SubmissionService{creatorModeration: fakeCreatorModerationChecker{allowed: true}}
		if !service.shouldAutoUpvoteClaimedClip(context.Background(), &creatorID, userID) {
			t.Fatal("expected auto-upvote to proceed when interaction is allowed")
		}
	})
}

func ptrString(s string) *string    { return &s }
func ptrInt(v int) *int             { return &v }
func ptrInt64(v int64) *int64       { return &v }
func ptrFloat64(v float64) *float64 { return &v }
