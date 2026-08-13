package storage

import (
	"context"
	"fmt"
	"io"
	"net/url"
	"strings"

	"git.subcult.tv/subculture-collective/clpr/config"
)

type PutObjectInput struct {
	Key         string
	Body        io.Reader
	Size        int64
	ContentType string
	Metadata    map[string]string
}

type PutObjectResult struct {
	Provider  string
	Bucket    string
	Key       string
	PublicURL string
}

const (
	pendingUploadPrefix = "uploads/pending/"
	publicUploadPrefix  = "uploads/public/"
)

type ClipStorage interface {
	PutObject(ctx context.Context, input PutObjectInput) (PutObjectResult, error)
	CopyObject(ctx context.Context, srcKey, dstKey, contentType string) (PutObjectResult, error)
	DeleteObject(ctx context.Context, key string) error
	PublicURL(key string) string
}

func IsPendingClipStorageKey(key string) bool {
	return strings.HasPrefix(strings.TrimSpace(key), pendingUploadPrefix)
}

func IsPublicClipStorageKey(key string) bool {
	return strings.HasPrefix(strings.TrimSpace(key), publicUploadPrefix)
}

func PublicClipStorageKeyFromPendingKey(key string) (string, bool) {
	trimmed := strings.TrimSpace(key)
	if !strings.HasPrefix(trimmed, pendingUploadPrefix) {
		return "", false
	}
	return publicUploadPrefix + strings.TrimPrefix(trimmed, pendingUploadPrefix), true
}

func ValidateS3ClipStorageConfig(cfg config.ClipStorageConfig) error {
	var problems []string

	provider := strings.TrimSpace(cfg.Provider)
	endpoint := strings.TrimSpace(cfg.Endpoint)
	bucket := strings.TrimSpace(cfg.Bucket)
	region := strings.TrimSpace(cfg.Region)
	accessKey := strings.TrimSpace(cfg.AccessKey)
	secretKey := strings.TrimSpace(cfg.SecretKey)
	publicBaseURL := strings.TrimSpace(cfg.PublicBaseURL)

	if provider == "" {
		problems = append(problems, "provider is required")
	}
	if endpoint == "" {
		problems = append(problems, "endpoint is required")
	} else if err := validateHTTPURL(endpoint); err != nil {
		problems = append(problems, fmt.Sprintf("endpoint is invalid: %v", err))
	}
	if bucket == "" {
		problems = append(problems, "bucket is required")
	}
	if region == "" {
		problems = append(problems, "region is required")
	}
	if accessKey == "" {
		problems = append(problems, "access key is required")
	}
	if secretKey == "" {
		problems = append(problems, "secret key is required")
	}
	if publicBaseURL == "" {
		problems = append(problems, "public base URL is required")
	} else if err := validateHTTPURL(publicBaseURL); err != nil {
		problems = append(problems, fmt.Sprintf("public base URL is invalid: %v", err))
	}

	if len(problems) > 0 {
		return fmt.Errorf("invalid clip storage config: %s", strings.Join(problems, "; "))
	}

	return nil
}

func normalizeS3ClipStorageConfig(cfg config.ClipStorageConfig) config.ClipStorageConfig {
	cfg.Provider = strings.TrimSpace(cfg.Provider)
	cfg.Endpoint = strings.TrimSpace(cfg.Endpoint)
	cfg.Bucket = strings.TrimSpace(cfg.Bucket)
	cfg.Region = strings.TrimSpace(cfg.Region)
	cfg.AccessKey = strings.TrimSpace(cfg.AccessKey)
	cfg.SecretKey = strings.TrimSpace(cfg.SecretKey)
	cfg.PublicBaseURL = strings.TrimSpace(cfg.PublicBaseURL)
	return cfg
}

func validateHTTPURL(raw string) error {
	parsed, err := url.Parse(raw)
	if err != nil {
		return err
	}
	if parsed.Scheme == "" || parsed.Host == "" {
		return fmt.Errorf("must include scheme and host")
	}
	if parsed.User != nil {
		return fmt.Errorf("must not include user info")
	}
	switch strings.ToLower(parsed.Scheme) {
	case "http", "https":
	default:
		return fmt.Errorf("must use http or https")
	}
	return nil
}
