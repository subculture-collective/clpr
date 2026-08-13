package services

import (
	"context"
	"strings"
	"testing"
	"time"

	"git.subcult.tv/subculture-collective/clpr/config"
	"git.subcult.tv/subculture-collective/clpr/internal/models"
	"github.com/google/uuid"
)

type fakeExternalMetadataFetcher struct {
	metadata ExternalMetadata
	err      error
}

func (f fakeExternalMetadataFetcher) Fetch(ctx context.Context, source DetectedSource) (ExternalMetadata, error) {
	return f.metadata, f.err
}

type trackingExternalMetadataFetcher struct {
	called *bool
}

func (f trackingExternalMetadataFetcher) Fetch(ctx context.Context, source DetectedSource) (ExternalMetadata, error) {
	if f.called != nil {
		*f.called = true
	}
	return ExternalMetadata{}, nil
}

func TestSubmitClip_ExternalSubmissionUnknownDurationAllowed(t *testing.T) {
	var captured *models.ClipSubmission
	service := &SubmissionService{
		cfg: &config.Config{
			ClipSource: config.ClipSourceConfig{MaxDurationSeconds: 600},
			Karma:      config.KarmaConfig{RequireKarmaForSubmission: false},
		},
		bypassRateLimits: true,
		userLookupFn: func(ctx context.Context, userID uuid.UUID) (*models.User, error) {
			return &models.User{Role: models.RoleUser, KarmaPoints: 0}, nil
		},
		submissionCreateFn: func(ctx context.Context, submission *models.ClipSubmission) error {
			captured = submission
			return nil
		},
		externalMetadataFetcher: fakeExternalMetadataFetcher{metadata: ExternalMetadata{
			Title:            "Great play",
			AuthorName:       "CreatorOne",
			ThumbnailURL:     "https://cdn.example/thumb.jpg",
			DurationVerified: false,
			EmbedURL:         "https://player.example/embed",
			Raw:              map[string]any{"source": "youtube"},
		}},
	}

	userID := uuid.New()
	req := &SubmitClipRequest{ClipURL: "https://www.youtube.com/watch?v=abc123"}
	submission, err := service.SubmitClip(context.Background(), userID, req, "127.0.0.1", "fingerprint")
	if err != nil {
		t.Fatalf("SubmitClip() error = %v", err)
	}
	if submission == nil || captured == nil {
		t.Fatalf("expected submission to be created")
	}
	if submission.Status != "pending" {
		t.Fatalf("Status = %q, want pending", submission.Status)
	}
	if submission.SourceType != string(SourceTypeExternal) || submission.SourcePlatform != string(SourcePlatformYouTube) {
		t.Fatalf("source fields = %q/%q, want external/youtube", submission.SourceType, submission.SourcePlatform)
	}
	if submission.SourceID == nil || *submission.SourceID != "abc123" {
		t.Fatalf("SourceID = %v, want abc123", submission.SourceID)
	}
	if submission.TwitchClipID == "abc123" {
		t.Fatalf("TwitchClipID reused raw source id")
	}
	if !strings.HasPrefix(submission.TwitchClipID, "external:youtube:") {
		t.Fatalf("TwitchClipID = %q, want external youtube legacy id", submission.TwitchClipID)
	}
	if len(submission.TwitchClipID) > 100 {
		t.Fatalf("TwitchClipID length = %d, want <= 100", len(submission.TwitchClipID))
	}
	if submission.DurationVerified {
		t.Fatalf("DurationVerified = true, want false")
	}
	if submission.DurationSeconds != nil {
		t.Fatalf("DurationSeconds = %v, want nil", submission.DurationSeconds)
	}
	if submission.TwitchClipURL != "https://www.youtube.com/watch?v=abc123" {
		t.Fatalf("TwitchClipURL = %q, want normalized URL", submission.TwitchClipURL)
	}
	if submission.CreatorName == nil || *submission.CreatorName != "CreatorOne" {
		t.Fatalf("CreatorName = %v, want CreatorOne", submission.CreatorName)
	}
	if !strings.Contains(string(submission.SourceMetadata), "CreatorOne") {
		t.Fatalf("SourceMetadata = %s, want encoded metadata", string(submission.SourceMetadata))
	}
}

func TestSubmitClip_ExternalSubmissionLegacyTwitchClipIDIsDistinctAndBounded(t *testing.T) {
	service := &SubmissionService{
		cfg: &config.Config{
			ClipSource: config.ClipSourceConfig{MaxDurationSeconds: 600},
			Karma:      config.KarmaConfig{RequireKarmaForSubmission: false},
		},
		bypassRateLimits: true,
		userLookupFn: func(ctx context.Context, userID uuid.UUID) (*models.User, error) {
			return &models.User{Role: models.RoleUser, KarmaPoints: 0}, nil
		},
		submissionCreateFn: func(ctx context.Context, submission *models.ClipSubmission) error {
			return nil
		},
		externalMetadataFetcher: fakeExternalMetadataFetcher{metadata: ExternalMetadata{
			Title:            "Great play",
			AuthorName:       "CreatorOne",
			DurationVerified: false,
		}},
	}

	userID := uuid.New()
	submit := func(url string) *models.ClipSubmission {
		t.Helper()
		submission, err := service.SubmitClip(context.Background(), userID, &SubmitClipRequest{ClipURL: url}, "127.0.0.1", "fingerprint")
		if err != nil {
			t.Fatalf("SubmitClip(%q) error = %v", url, err)
		}
		return submission
	}

	youtube := submit("https://www.youtube.com/watch?v=Clip123")
	shorts := submit("https://www.youtube.com/shorts/Clip123")
	longSourceID := strings.Repeat("a", 180)
	long := submit("https://www.youtube.com/watch?v=" + longSourceID)

	if youtube.SourceID == nil || *youtube.SourceID != "Clip123" {
		t.Fatalf("youtube SourceID = %v, want Clip123", youtube.SourceID)
	}
	if shorts.SourceID == nil || *shorts.SourceID != "Clip123" {
		t.Fatalf("shorts SourceID = %v, want Clip123", shorts.SourceID)
	}
	if youtube.TwitchClipID == "Clip123" || shorts.TwitchClipID == "Clip123" {
		t.Fatalf("legacy TwitchClipID reused raw source id")
	}
	if youtube.TwitchClipID == shorts.TwitchClipID {
		t.Fatalf("youtube and shorts legacy TwitchClipID should differ, got %q", youtube.TwitchClipID)
	}
	if !strings.HasPrefix(youtube.TwitchClipID, "external:youtube:") {
		t.Fatalf("youtube TwitchClipID = %q, want external youtube legacy id", youtube.TwitchClipID)
	}
	if !strings.HasPrefix(shorts.TwitchClipID, "external:youtube_shorts:") {
		t.Fatalf("shorts TwitchClipID = %q, want external youtube shorts legacy id", shorts.TwitchClipID)
	}
	if len(youtube.TwitchClipID) > 100 || len(shorts.TwitchClipID) > 100 {
		t.Fatalf("legacy TwitchClipID lengths = %d/%d, want <= 100", len(youtube.TwitchClipID), len(shorts.TwitchClipID))
	}
	if long.SourceID == nil || *long.SourceID != longSourceID {
		t.Fatalf("long SourceID = %v, want raw long source id", long.SourceID)
	}
	if long.TwitchClipID == longSourceID {
		t.Fatalf("long TwitchClipID reused raw source id")
	}
	if len(long.TwitchClipID) > 100 {
		t.Fatalf("long TwitchClipID length = %d, want <= 100", len(long.TwitchClipID))
	}
}

func TestSubmitClip_ExternalSubmissionVerifiedOverDurationRejected(t *testing.T) {
	service := &SubmissionService{
		cfg: &config.Config{
			ClipSource: config.ClipSourceConfig{MaxDurationSeconds: 600},
			Karma:      config.KarmaConfig{RequireKarmaForSubmission: false},
		},
		bypassRateLimits: true,
		userLookupFn: func(ctx context.Context, userID uuid.UUID) (*models.User, error) {
			return &models.User{Role: models.RoleUser}, nil
		},
		externalMetadataFetcher: fakeExternalMetadataFetcher{metadata: ExternalMetadata{
			Title:            "Too long",
			DurationSeconds:  externalInt64Ptr(601),
			DurationVerified: true,
		}},
	}

	_, err := service.SubmitClip(context.Background(), uuid.New(), &SubmitClipRequest{ClipURL: "https://www.youtube.com/watch?v=abc123"}, "127.0.0.1", "fingerprint")
	if err == nil {
		t.Fatalf("expected error, got nil")
	}
	valErr, ok := err.(*ValidationError)
	if !ok {
		t.Fatalf("expected ValidationError, got %T", err)
	}
	if valErr.Field != "clip_url" {
		t.Fatalf("Field = %q, want clip_url", valErr.Field)
	}
	if valErr.Message == "" {
		t.Fatalf("expected non-empty validation message")
	}
}

func TestSubmitClip_ExternalDuplicateRejectedBySourceIdentity(t *testing.T) {
	userID := uuid.New()
	sourceID := "abc123"
	inputURL := "https://www.youtube.com/watch?v=abc123"
	metadataFetchCalled := false

	tests := []struct {
		name         string
		existing     *models.ClipSubmission
		wantContains string
		wantClipID   bool
	}{
		{
			name: "pending duplicate is rejected",
			existing: &models.ClipSubmission{
				ID:             uuid.New(),
				Status:         "pending",
				SourceType:     string(SourceTypeExternal),
				SourcePlatform: string(SourcePlatformYouTube),
				SourceID:       &sourceID,
				CreatedAt:      time.Now().Add(-2 * time.Hour),
			},
			wantContains: "already pending review",
		},
		{
			name: "approved duplicate includes existing clip id",
			existing: &models.ClipSubmission{
				ID:             uuid.New(),
				Status:         "approved",
				ClipID:         func() *uuid.UUID { id := uuid.New(); return &id }(),
				SourceType:     string(SourceTypeExternal),
				SourcePlatform: string(SourcePlatformYouTube),
				SourceID:       &sourceID,
				CreatedAt:      time.Now().Add(-2 * time.Hour),
			},
			wantContains: "already been approved and added to our database",
			wantClipID:   true,
		},
		{
			name: "recently rejected duplicate is rejected",
			existing: &models.ClipSubmission{
				ID:             uuid.New(),
				Status:         "rejected",
				SourceType:     string(SourceTypeExternal),
				SourcePlatform: string(SourcePlatformYouTube),
				SourceID:       &sourceID,
				CreatedAt:      time.Now().Add(-48 * time.Hour),
			},
			wantContains: "recently rejected",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			lookupCalled := false
			metadataFetchCalled = false
			service := &SubmissionService{
				cfg: &config.Config{
					ClipSource: config.ClipSourceConfig{MaxDurationSeconds: 600},
					Karma:      config.KarmaConfig{RequireKarmaForSubmission: false},
				},
				bypassRateLimits: true,
				userLookupFn: func(ctx context.Context, id uuid.UUID) (*models.User, error) {
					return &models.User{Role: models.RoleUser}, nil
				},
				submissionLookupBySourceFn: func(ctx context.Context, platform string, source string) (*models.ClipSubmission, error) {
					lookupCalled = true
					if platform != string(SourcePlatformYouTube) || source != sourceID {
						t.Fatalf("lookup key = %s/%s, want youtube/%s", platform, source, sourceID)
					}
					return tt.existing, nil
				},
				externalMetadataFetcher: trackingExternalMetadataFetcher{called: &metadataFetchCalled},
			}

			_, err := service.SubmitClip(context.Background(), userID, &SubmitClipRequest{ClipURL: inputURL}, "127.0.0.1", "fingerprint")
			if err == nil {
				t.Fatalf("expected duplicate error, got nil")
			}
			valErr, ok := err.(*ValidationError)
			if !ok {
				t.Fatalf("expected ValidationError, got %T", err)
			}
			if valErr.Field != "clip_url" {
				t.Fatalf("Field = %q, want clip_url", valErr.Field)
			}
			if !strings.Contains(valErr.Message, tt.wantContains) {
				t.Fatalf("Message = %q, want to contain %q", valErr.Message, tt.wantContains)
			}
			if tt.wantClipID && tt.existing.ClipID != nil && !strings.Contains(valErr.Message, tt.existing.ClipID.String()) {
				t.Fatalf("Message = %q, want existing clip id %s", valErr.Message, tt.existing.ClipID.String())
			}
			if !lookupCalled {
				t.Fatal("expected source identity lookup to be called")
			}
			if metadataFetchCalled {
				t.Fatal("expected metadata fetch to be skipped for duplicate submissions")
			}
		})
	}
}

func externalInt64Ptr(v int64) *int64 { return &v }
