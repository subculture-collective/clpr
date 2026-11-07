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

// An imported upload keeps its archive provenance from submission to approved
// clip, and its playback stays the uploaded file, never the passage URL.
func TestArchiveProvenanceCarriesFromUploadToApprovedClip(t *testing.T) {
	handoff, err := ParseArchiveHandoff(loadArchiveHandoffFixture(t))
	if err != nil {
		t.Fatal(err)
	}
	record := BuildArchiveProvenanceRecord(handoff, nil, time.Now())

	service := &SubmissionService{cfg: &config.Config{}}
	userID, submissionID := uuid.New(), uuid.New()
	storageKey := BuildUploadStorageKey(userID, submissionID, ".mp4")
	submission, err := service.buildUploadSubmissionRecord(userID, submissionID, &SubmitUploadRequest{
		OriginalFilename:  handoff.Media.Filename,
		MimeType:          "video/mp4",
		FileSizeBytes:     handoff.Media.Bytes,
		DurationSeconds:   9,
		DurationVerified:  true,
		StorageProvider:   "s3",
		StorageBucket:     "clips",
		StorageKey:        storageKey,
		ArchiveProvenance: record,
	}, nil, nil, nil, time.Now())
	if err != nil {
		t.Fatal(err)
	}
	if submission.Status != "pending" || submission.SourceURL != nil {
		t.Fatalf("submission status/source_url = %q/%v; import must stay pending with no source URL", submission.Status, submission.SourceURL)
	}
	if got := ArchiveProvenanceFromMetadata(submission.SourceMetadata); got == nil || got.ExportID != handoff.ExportID {
		t.Fatalf("submission provenance = %+v", got)
	}

	publicURL := "https://cdn.example.com/uploads/public/video.mp4"
	var captured *models.Clip
	approver := &SubmissionService{
		clipStorage: &fakeUploadStorage{publicURL: publicURL},
		clipCreateFn: func(ctx context.Context, clip *models.Clip) error {
			captured = clip
			return nil
		},
	}
	if _, err := approver.createClipFromSubmission(context.Background(), submission); err != nil {
		t.Fatal(err)
	}
	clipRecord := ArchiveProvenanceFromMetadata(captured.SourceMetadata)
	if clipRecord == nil || clipRecord.Passage.URL != handoff.Passage.URL || len(clipRecord.Captions.Cues) != 4 {
		t.Fatalf("clip provenance = %+v", clipRecord)
	}
	for name, value := range map[string]*string{"video_url": captured.VideoURL, "source_url": captured.SourceURL} {
		if value == nil || *value != publicURL || strings.Contains(*value, "archive.example") {
			t.Fatalf("%s = %v, want the uploaded file %s", name, value, publicURL)
		}
	}
}
