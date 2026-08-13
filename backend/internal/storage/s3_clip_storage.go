package storage

import (
	"context"
	"fmt"
	"net/url"
	"strings"

	"git.subcult.tv/subculture-collective/clpr/config"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/service/s3"
)

type s3Client interface {
	PutObject(ctx context.Context, params *s3.PutObjectInput, optFns ...func(*s3.Options)) (*s3.PutObjectOutput, error)
	CopyObject(ctx context.Context, params *s3.CopyObjectInput, optFns ...func(*s3.Options)) (*s3.CopyObjectOutput, error)
	DeleteObject(ctx context.Context, params *s3.DeleteObjectInput, optFns ...func(*s3.Options)) (*s3.DeleteObjectOutput, error)
}

type S3ClipStorage struct {
	client s3Client
	cfg    config.ClipStorageConfig
	base   *url.URL
}

func NewS3ClipStorage(cfg config.ClipStorageConfig) (*S3ClipStorage, error) {
	if err := ValidateS3ClipStorageConfig(cfg); err != nil {
		return nil, err
	}

	normalized := normalizeS3ClipStorageConfig(cfg)
	baseURL, _ := url.Parse(normalized.PublicBaseURL)
	client := s3.New(s3.Options{
		Region:       normalized.Region,
		Credentials:  credentials.NewStaticCredentialsProvider(normalized.AccessKey, normalized.SecretKey, ""),
		BaseEndpoint: aws.String(normalized.Endpoint),
		UsePathStyle: normalized.ForcePathStyle,
	})

	return &S3ClipStorage{client: client, cfg: normalized, base: baseURL}, nil
}

func newS3ClipStorageWithClient(cfg config.ClipStorageConfig, client s3Client) (*S3ClipStorage, error) {
	if err := ValidateS3ClipStorageConfig(cfg); err != nil {
		return nil, err
	}

	normalized := normalizeS3ClipStorageConfig(cfg)
	baseURL, _ := url.Parse(normalized.PublicBaseURL)

	return &S3ClipStorage{client: client, cfg: normalized, base: baseURL}, nil
}

func (s *S3ClipStorage) PutObject(ctx context.Context, input PutObjectInput) (PutObjectResult, error) {
	if s == nil {
		return PutObjectResult{}, fmt.Errorf("storage is nil")
	}
	if strings.TrimSpace(input.Key) == "" {
		return PutObjectResult{}, fmt.Errorf("object key is required")
	}
	if input.Body == nil {
		return PutObjectResult{}, fmt.Errorf("object body is required")
	}

	putInput := &s3.PutObjectInput{
		Bucket:        aws.String(s.cfg.Bucket),
		Key:           aws.String(input.Key),
		Body:          input.Body,
		ContentLength: aws.Int64(input.Size),
	}
	if input.ContentType != "" {
		putInput.ContentType = aws.String(input.ContentType)
	}
	if len(input.Metadata) > 0 {
		putInput.Metadata = make(map[string]string, len(input.Metadata))
		for k, v := range input.Metadata {
			putInput.Metadata[k] = v
		}
	}

	if _, err := s.client.PutObject(ctx, putInput); err != nil {
		return PutObjectResult{}, fmt.Errorf("upload clip object: %w", err)
	}

	return PutObjectResult{
		Provider:  s.cfg.Provider,
		Bucket:    s.cfg.Bucket,
		Key:       input.Key,
		PublicURL: s.PublicURL(input.Key),
	}, nil
}

func (s *S3ClipStorage) CopyObject(ctx context.Context, srcKey, dstKey, contentType string) (PutObjectResult, error) {
	if s == nil {
		return PutObjectResult{}, fmt.Errorf("storage is nil")
	}
	if strings.TrimSpace(srcKey) == "" {
		return PutObjectResult{}, fmt.Errorf("source key is required")
	}
	if strings.TrimSpace(dstKey) == "" {
		return PutObjectResult{}, fmt.Errorf("destination key is required")
	}

	copyInput := &s3.CopyObjectInput{
		Bucket:     aws.String(s.cfg.Bucket),
		CopySource: aws.String(fmt.Sprintf("%s/%s", s.cfg.Bucket, url.PathEscape(strings.TrimSpace(srcKey)))),
		Key:        aws.String(strings.TrimSpace(dstKey)),
	}
	if contentType != "" {
		copyInput.ContentType = aws.String(contentType)
	}

	if _, err := s.client.CopyObject(ctx, copyInput); err != nil {
		return PutObjectResult{}, fmt.Errorf("copy clip object: %w", err)
	}

	return PutObjectResult{
		Provider:  s.cfg.Provider,
		Bucket:    s.cfg.Bucket,
		Key:       strings.TrimSpace(dstKey),
		PublicURL: s.PublicURL(dstKey),
	}, nil

}

func (s *S3ClipStorage) DeleteObject(ctx context.Context, key string) error {
	if s == nil {
		return fmt.Errorf("storage is nil")
	}
	if strings.TrimSpace(key) == "" {
		return fmt.Errorf("object key is required")
	}

	_, err := s.client.DeleteObject(ctx, &s3.DeleteObjectInput{
		Bucket: aws.String(s.cfg.Bucket),
		Key:    aws.String(key),
	})
	if err != nil {
		return fmt.Errorf("delete clip object: %w", err)
	}

	return nil
}

func (s *S3ClipStorage) PublicURL(key string) string {
	if s == nil || s.base == nil {
		return ""
	}
	if !IsPublicClipStorageKey(key) {
		return ""
	}

	cloned := *s.base
	cloned.Path = joinURLPath(cloned.Path, key)
	cloned.RawPath = ""
	return cloned.String()
}

func joinURLPath(basePath, key string) string {
	basePath = strings.TrimRight(basePath, "/")
	key = strings.TrimLeft(key, "/")

	switch {
	case basePath == "" && key == "":
		return ""
	case basePath == "":
		return "/" + key
	case key == "":
		return basePath
	default:
		return basePath + "/" + key
	}
}

var _ ClipStorage = (*S3ClipStorage)(nil)
