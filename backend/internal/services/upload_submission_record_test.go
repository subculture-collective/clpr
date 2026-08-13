package services

import (
	"strings"
	"testing"
	"time"

	"git.subcult.tv/subculture-collective/clpr/config"
	"github.com/google/uuid"
)

func TestBuildUploadStorageKey(t *testing.T) {
	userID := uuid.MustParse("11111111-1111-1111-1111-111111111111")
	submissionID := uuid.MustParse("22222222-2222-2222-2222-222222222222")

	got := BuildUploadStorageKey(userID, submissionID, ".mp4")
	want := "uploads/pending/11111111-1111-1111-1111-111111111111/22222222-2222-2222-2222-222222222222/original.mp4"
	if got != want {
		t.Fatalf("buildUploadStorageKey() = %q, want %q", got, want)
	}
}

func TestBuildUploadSubmissionRecord_RemainsPendingAndPrivate(t *testing.T) {
	service := &SubmissionService{cfg: &config.Config{ClipSource: config.ClipSourceConfig{RequireModerationForUpload: true}}}
	userID := uuid.MustParse("11111111-1111-1111-1111-111111111111")
	submissionID := uuid.MustParse("22222222-2222-2222-2222-222222222222")
	customTitle := "Great Moment"
	reason := "Because it is great"
	req := &SubmitUploadRequest{
		OriginalFilename: "clip.mp4",
		MimeType:         "video/mp4",
		FileSizeBytes:    2048,
		DurationSeconds:  42,
		DurationVerified: true,
		StorageProvider:  "s3",
		StorageBucket:    "clips",
		StorageKey:       "uploads/pending/11111111-1111-1111-1111-111111111111/22222222-2222-2222-2222-222222222222/original.mp4",
		CustomTitle:      &customTitle,
		SubmissionReason: &reason,
	}

	submission, err := service.buildUploadSubmissionRecord(userID, submissionID, req, &customTitle, &customTitle, &reason, time.Unix(1700000000, 0))
	if err != nil {
		t.Fatalf("buildUploadSubmissionRecord() error = %v", err)
	}

	if submission.Status != "pending" {
		t.Fatalf("Status = %q, want pending", submission.Status)
	}
	if submission.UploadStatus != "validated" {
		t.Fatalf("UploadStatus = %q, want validated", submission.UploadStatus)
	}
	if submission.StorageVisibility != "private" {
		t.Fatalf("StorageVisibility = %q, want private", submission.StorageVisibility)
	}
	if submission.SourceType != "upload" || submission.SourcePlatform != "upload" {
		t.Fatalf("source fields = %q/%q, want upload/upload", submission.SourceType, submission.SourcePlatform)
	}
	if submission.Title == nil || *submission.Title != customTitle {
		t.Fatalf("Title = %v, want %q", submission.Title, customTitle)
	}
	if submission.SourceID == nil || !strings.HasPrefix(*submission.SourceID, submissionID.String()) {
		t.Fatalf("SourceID = %v, want prefix %q", submission.SourceID, submissionID.String())
	}
	if !strings.HasPrefix(submission.TwitchClipID, "upload:") {
		t.Fatalf("TwitchClipID = %q, want upload legacy prefix", submission.TwitchClipID)
	}
	if len(submission.TwitchClipID) > 100 {
		t.Fatalf("TwitchClipID length = %d, want <= 100", len(submission.TwitchClipID))
	}
	if submission.SubmissionReason == nil || *submission.SubmissionReason != reason {
		t.Fatalf("SubmissionReason = %v, want %q", submission.SubmissionReason, reason)
	}
}
