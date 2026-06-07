package storage

import (
	"bytes"
	"context"
	"io"
	"strings"
	"testing"

	"git.subcult.tv/subculture-collective/clpr/config"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/s3"
)

type fakeS3Client struct {
	putInput    *s3.PutObjectInput
	copyInput   *s3.CopyObjectInput
	deleteInput *s3.DeleteObjectInput
	putBody     []byte
	putErr      error
	copyErr     error
	deleteErr   error
}

func (f *fakeS3Client) PutObject(ctx context.Context, params *s3.PutObjectInput, optFns ...func(*s3.Options)) (*s3.PutObjectOutput, error) {
	f.putInput = params
	if params != nil && params.Body != nil {
		body, _ := io.ReadAll(params.Body)
		f.putBody = body
	}
	return &s3.PutObjectOutput{}, f.putErr
}

func (f *fakeS3Client) CopyObject(ctx context.Context, params *s3.CopyObjectInput, optFns ...func(*s3.Options)) (*s3.CopyObjectOutput, error) {
	f.copyInput = params
	return &s3.CopyObjectOutput{}, f.copyErr
}

func (f *fakeS3Client) DeleteObject(ctx context.Context, params *s3.DeleteObjectInput, optFns ...func(*s3.Options)) (*s3.DeleteObjectOutput, error) {
	f.deleteInput = params
	return &s3.DeleteObjectOutput{}, f.deleteErr
}

func validClipStorageConfig() config.ClipStorageConfig {
	return config.ClipStorageConfig{
		Provider:       "s3",
		Endpoint:       "http://localhost:9000",
		Bucket:         "clpr-clips",
		Region:         "us-east-1",
		AccessKey:      "minioadmin",
		SecretKey:      "minioadmin",
		ForcePathStyle: true,
		PublicBaseURL:  "http://localhost:9000/clpr-clips",
	}
}

func TestValidateS3ClipStorageConfig(t *testing.T) {
	tests := []struct {
		name       string
		mutate     func(*config.ClipStorageConfig)
		wantSubstr string
	}{
		{
			name:       "endpoint ftp scheme",
			mutate:     func(cfg *config.ClipStorageConfig) { cfg.Endpoint = "ftp://localhost:9000" },
			wantSubstr: "endpoint is invalid: must use http or https",
		},
		{
			name:       "endpoint javascript scheme",
			mutate:     func(cfg *config.ClipStorageConfig) { cfg.Endpoint = "javascript:alert(1)" },
			wantSubstr: "endpoint is invalid: must include scheme and host",
		},
		{
			name:       "endpoint userinfo",
			mutate:     func(cfg *config.ClipStorageConfig) { cfg.Endpoint = "https://user:pass@localhost:9000" },
			wantSubstr: "endpoint is invalid: must not include user info",
		},
		{
			name:       "missing bucket",
			mutate:     func(cfg *config.ClipStorageConfig) { cfg.Bucket = "" },
			wantSubstr: "bucket is required",
		},
		{
			name:       "missing endpoint",
			mutate:     func(cfg *config.ClipStorageConfig) { cfg.Endpoint = "" },
			wantSubstr: "endpoint is required",
		},
		{
			name:       "missing credentials",
			mutate:     func(cfg *config.ClipStorageConfig) { cfg.AccessKey, cfg.SecretKey = "", "" },
			wantSubstr: "access key is required",
		},
		{
			name:       "public base url ftp scheme",
			mutate:     func(cfg *config.ClipStorageConfig) { cfg.PublicBaseURL = "ftp://localhost:9000/clpr-clips" },
			wantSubstr: "public base URL is invalid: must use http or https",
		},
		{
			name:       "public base url javascript scheme",
			mutate:     func(cfg *config.ClipStorageConfig) { cfg.PublicBaseURL = "javascript:alert(1)" },
			wantSubstr: "public base URL is invalid: must include scheme and host",
		},
		{
			name:       "public base url userinfo",
			mutate:     func(cfg *config.ClipStorageConfig) { cfg.PublicBaseURL = "https://user:pass@localhost:9000/clpr-clips" },
			wantSubstr: "public base URL is invalid: must not include user info",
		},
		{
			name:       "missing public base url",
			mutate:     func(cfg *config.ClipStorageConfig) { cfg.PublicBaseURL = "" },
			wantSubstr: "public base URL is required",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cfg := validClipStorageConfig()
			tt.mutate(&cfg)

			err := ValidateS3ClipStorageConfig(cfg)
			if err == nil {
				t.Fatal("ValidateS3ClipStorageConfig() error = nil, want error")
			}
			if !strings.Contains(err.Error(), tt.wantSubstr) {
				t.Fatalf("ValidateS3ClipStorageConfig() error = %v, want substring %q", err, tt.wantSubstr)
			}
		})
	}
}

func TestValidateHTTPURL(t *testing.T) {
	tests := []struct {
		name       string
		raw        string
		wantSubstr string
	}{
		{name: "http ok", raw: "http://localhost:9000"},
		{name: "https ok", raw: "https://cdn.example.com/clips"},
		{name: "ftp rejected", raw: "ftp://localhost:9000", wantSubstr: "must use http or https"},
		{name: "javascript rejected", raw: "javascript:alert(1)", wantSubstr: "must include scheme and host"},
		{name: "userinfo rejected", raw: "https://user:pass@localhost:9000", wantSubstr: "must not include user info"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validateHTTPURL(tt.raw)
			if tt.wantSubstr == "" {
				if err != nil {
					t.Fatalf("validateHTTPURL(%q) error = %v, want nil", tt.raw, err)
				}
				return
			}
			if err == nil {
				t.Fatalf("validateHTTPURL(%q) error = nil, want error", tt.raw)
			}
			if !strings.Contains(err.Error(), tt.wantSubstr) {
				t.Fatalf("validateHTTPURL(%q) error = %v, want substring %q", tt.raw, err, tt.wantSubstr)
			}
		})
	}
}

func TestS3ClipStoragePublicURL(t *testing.T) {
	storage, err := newS3ClipStorageWithClient(validClipStorageConfig(), &fakeS3Client{})
	if err != nil {
		t.Fatalf("newS3ClipStorageWithClient() error = %v", err)
	}

	if got := storage.PublicURL("uploads/pending/user-1/submission-2/original.mp4"); got != "" {
		t.Fatalf("PublicURL(pending) = %q, want empty", got)
	}

	got := storage.PublicURL("uploads/public/user-1/submission-2/original.mp4")
	want := "http://localhost:9000/clpr-clips/uploads/public/user-1/submission-2/original.mp4"
	if got != want {
		t.Fatalf("PublicURL() = %q, want %q", got, want)
	}
}

func TestS3ClipStoragePutObject(t *testing.T) {
	fake := &fakeS3Client{}
	storage, err := newS3ClipStorageWithClient(validClipStorageConfig(), fake)
	if err != nil {
		t.Fatalf("newS3ClipStorageWithClient() error = %v", err)
	}

	result, err := storage.PutObject(context.Background(), PutObjectInput{
		Key:         "uploads/pending/user-1/submission-2/original.webm",
		Body:        bytes.NewReader([]byte("clip-bytes")),
		Size:        int64(len("clip-bytes")),
		ContentType: "video/webm",
		Metadata: map[string]string{
			"original_filename": "original.webm",
		},
	})
	if err != nil {
		t.Fatalf("PutObject() error = %v", err)
	}

	if got, want := result.Provider, "s3"; got != want {
		t.Fatalf("Provider = %q, want %q", got, want)
	}
	if got, want := result.Bucket, "clpr-clips"; got != want {
		t.Fatalf("Bucket = %q, want %q", got, want)
	}
	if got, want := result.Key, "uploads/pending/user-1/submission-2/original.webm"; got != want {
		t.Fatalf("Key = %q, want %q", got, want)
	}
	if got := result.PublicURL; got != "" {
		t.Fatalf("PublicURL = %q, want empty", got)
	}

	if fake.putInput == nil {
		t.Fatal("PutObject was not called")
	}
	if got, want := aws.ToString(fake.putInput.Bucket), "clpr-clips"; got != want {
		t.Fatalf("PutObject Bucket = %q, want %q", got, want)
	}
	if got, want := aws.ToString(fake.putInput.Key), "uploads/pending/user-1/submission-2/original.webm"; got != want {
		t.Fatalf("PutObject Key = %q, want %q", got, want)
	}
	if got, want := aws.ToString(fake.putInput.ContentType), "video/webm"; got != want {
		t.Fatalf("PutObject ContentType = %q, want %q", got, want)
	}
	if got, want := aws.ToInt64(fake.putInput.ContentLength), int64(len("clip-bytes")); got != want {
		t.Fatalf("PutObject ContentLength = %d, want %d", got, want)
	}
	if got, want := string(fake.putBody), "clip-bytes"; got != want {
		t.Fatalf("PutObject body = %q, want %q", got, want)
	}
	if got, want := fake.putInput.Metadata["original_filename"], "original.webm"; got != want {
		t.Fatalf("PutObject metadata = %q, want %q", got, want)
	}
}

func TestS3ClipStorageCopyObject(t *testing.T) {
	fake := &fakeS3Client{}
	storage, err := newS3ClipStorageWithClient(validClipStorageConfig(), fake)
	if err != nil {
		t.Fatalf("newS3ClipStorageWithClient() error = %v", err)
	}

	result, err := storage.CopyObject(context.Background(), "uploads/pending/user-1/submission-2/original.webm", "uploads/public/user-1/submission-2/original.webm", "video/webm")
	if err != nil {
		t.Fatalf("CopyObject() error = %v", err)
	}

	if got, want := result.Key, "uploads/public/user-1/submission-2/original.webm"; got != want {
		t.Fatalf("Key = %q, want %q", got, want)
	}
	if got, want := result.PublicURL, "http://localhost:9000/clpr-clips/uploads/public/user-1/submission-2/original.webm"; got != want {
		t.Fatalf("PublicURL = %q, want %q", got, want)
	}
	if fake.copyInput == nil {
		t.Fatal("CopyObject was not called")
	}
	if got, want := aws.ToString(fake.copyInput.Bucket), "clpr-clips"; got != want {
		t.Fatalf("CopyObject Bucket = %q, want %q", got, want)
	}
	if got, want := aws.ToString(fake.copyInput.Key), "uploads/public/user-1/submission-2/original.webm"; got != want {
		t.Fatalf("CopyObject Key = %q, want %q", got, want)
	}
	if got, want := aws.ToString(fake.copyInput.CopySource), "clpr-clips/uploads%2Fpending%2Fuser-1%2Fsubmission-2%2Foriginal.webm"; got != want {
		t.Fatalf("CopyObject CopySource = %q, want %q", got, want)
	}
	if got, want := aws.ToString(fake.copyInput.ContentType), "video/webm"; got != want {
		t.Fatalf("CopyObject ContentType = %q, want %q", got, want)
	}
}

func TestS3ClipStorageDeleteObject(t *testing.T) {
	fake := &fakeS3Client{}
	storage, err := newS3ClipStorageWithClient(validClipStorageConfig(), fake)
	if err != nil {
		t.Fatalf("newS3ClipStorageWithClient() error = %v", err)
	}

	if err := storage.DeleteObject(context.Background(), "uploads/pending/user-1/submission-2/original.webm"); err != nil {
		t.Fatalf("DeleteObject() error = %v", err)
	}

	if fake.deleteInput == nil {
		t.Fatal("DeleteObject was not called")
	}
	if got, want := aws.ToString(fake.deleteInput.Bucket), "clpr-clips"; got != want {
		t.Fatalf("DeleteObject Bucket = %q, want %q", got, want)
	}
	if got, want := aws.ToString(fake.deleteInput.Key), "uploads/pending/user-1/submission-2/original.webm"; got != want {
		t.Fatalf("DeleteObject Key = %q, want %q", got, want)
	}
}
