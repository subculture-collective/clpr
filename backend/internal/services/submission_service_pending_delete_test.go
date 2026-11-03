package services

import (
	"context"
	"testing"

	"git.subcult.tv/subculture-collective/clpr/internal/models"
	"git.subcult.tv/subculture-collective/clpr/internal/storage"
)

type deleteRecordingStorage struct {
	deleted []string
}

func (d *deleteRecordingStorage) PutObject(context.Context, storage.PutObjectInput) (storage.PutObjectResult, error) {
	return storage.PutObjectResult{}, nil
}

func (d *deleteRecordingStorage) CopyObject(context.Context, string, string, string) (storage.PutObjectResult, error) {
	return storage.PutObjectResult{}, nil
}

func (d *deleteRecordingStorage) DeleteObject(_ context.Context, key string) error {
	d.deleted = append(d.deleted, key)
	return nil
}

func (d *deleteRecordingStorage) PublicURL(key string) string { return "https://cdn.example/" + key }

func TestDeletePendingUploadAfterApproval_OnlyDeletesPendingKeys(t *testing.T) {
	tests := []struct {
		name string
		key  *string
		want []string
	}{
		{"pending key is deleted", strPtr(" uploads/pending/u/clip.mp4 "), []string{"uploads/pending/u/clip.mp4"}},
		{"public key is the approved clip file and is kept", strPtr("uploads/public/u/clip.mp4"), nil},
		{"missing key is ignored", nil, nil},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			store := &deleteRecordingStorage{}
			svc := &SubmissionService{clipStorage: store}
			if err := svc.deletePendingUploadAfterApproval(context.Background(), &models.ClipSubmission{StorageKey: tt.key}); err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if len(store.deleted) != len(tt.want) {
				t.Fatalf("deleted %v, want %v", store.deleted, tt.want)
			}
			for i := range tt.want {
				if store.deleted[i] != tt.want[i] {
					t.Fatalf("deleted %v, want %v", store.deleted, tt.want)
				}
			}
		})
	}
}
