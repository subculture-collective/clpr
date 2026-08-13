package services

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"regexp"
	"strings"
	"time"

	"golang.org/x/net/html"
)

// ExternalMetadata captures best-effort metadata for non-Twitch sources.
type ExternalMetadata struct {
	Title            string
	AuthorName       string
	ThumbnailURL     string
	DurationSeconds  *int64
	DurationVerified bool
	EmbedURL         string
	Raw              map[string]any
}

// ExternalMetadataFetcher fetches metadata for a detected external source.
type ExternalMetadataFetcher interface {
	Fetch(ctx context.Context, source DetectedSource) (ExternalMetadata, error)
}

type httpExternalMetadataFetcher struct {
	client *http.Client
}

var oEmbedIframeSrcRe = regexp.MustCompile(`(?i)src=["']([^"']+)`)

// NewExternalMetadataFetcher creates a metadata fetcher with a sensible default client.
func NewExternalMetadataFetcher(client *http.Client) ExternalMetadataFetcher {
	if client == nil {
		client = &http.Client{Timeout: 10 * time.Second}
	}
	return &httpExternalMetadataFetcher{client: client}
}

func (f *httpExternalMetadataFetcher) Fetch(ctx context.Context, source DetectedSource) (ExternalMetadata, error) {
	switch source.Platform {
	case SourcePlatformYouTube, SourcePlatformYouTubeShorts:
		return f.fetchYouTube(ctx, source)
	case SourcePlatformKick, SourcePlatformTikTok:
		return f.fetchOpenGraph(ctx, source)
	default:
		return ExternalMetadata{}, fmt.Errorf("unsupported external metadata platform: %s", source.Platform)
	}
}

func (f *httpExternalMetadataFetcher) fetchYouTube(ctx context.Context, source DetectedSource) (ExternalMetadata, error) {
	if metadata, err := f.fetchYouTubeOEmbed(ctx, source); err == nil {
		return metadata, nil
	}

	return f.fetchOpenGraph(ctx, source)
}

func (f *httpExternalMetadataFetcher) fetchYouTubeOEmbed(ctx context.Context, source DetectedSource) (ExternalMetadata, error) {
	pageURL := source.NormalizedURL
	if pageURL == "" {
		pageURL = source.RawURL
	}
	if pageURL == "" {
		return ExternalMetadata{}, fmt.Errorf("missing source url")
	}

	requestURL, err := buildOEmbedURL(pageURL)
	if err != nil {
		return ExternalMetadata{}, err
	}

	body, err := f.fetchBody(ctx, requestURL)
	if err != nil {
		return ExternalMetadata{}, err
	}

	var payload map[string]any
	if err := json.Unmarshal(body, &payload); err != nil {
		return ExternalMetadata{}, fmt.Errorf("decode oembed response: %w", err)
	}

	metadata := ExternalMetadata{
		Title:            stringFromMap(payload, "title"),
		AuthorName:       stringFromMap(payload, "author_name"),
		ThumbnailURL:     stringFromMap(payload, "thumbnail_url"),
		DurationVerified: false,
		EmbedURL:         source.NormalizedURL,
		Raw:              map[string]any{"oembed": payload, "source_url": pageURL},
	}
	if embedURL := extractOEmbedEmbedURL(stringFromMap(payload, "html")); embedURL != "" {
		metadata.EmbedURL = embedURL
	}

	return metadata, nil
}

func (f *httpExternalMetadataFetcher) fetchOpenGraph(ctx context.Context, source DetectedSource) (ExternalMetadata, error) {
	pageURL := source.NormalizedURL
	if pageURL == "" {
		pageURL = source.RawURL
	}
	if pageURL == "" {
		return ExternalMetadata{}, fmt.Errorf("missing source url")
	}

	body, err := f.fetchBody(ctx, pageURL)
	if err != nil {
		return ExternalMetadata{
			DurationVerified: false,
			EmbedURL:         pageURL,
			Raw: map[string]any{
				"source_url": pageURL,
				"error":      err.Error(),
			},
		}, nil
	}

	parsedURL, err := url.Parse(pageURL)
	if err != nil {
		return ExternalMetadata{}, fmt.Errorf("parse source url: %w", err)
	}

	tags, pageTitle, canonicalURL, err := parseMetadataTags(body, parsedURL)
	if err != nil {
		return ExternalMetadata{
			DurationVerified: false,
			EmbedURL:         pageURL,
			Raw: map[string]any{
				"source_url": pageURL,
				"error":      err.Error(),
			},
		}, nil
	}

	metadata := ExternalMetadata{
		Title:            firstNonEmpty(tags["og:title"], tags["twitter:title"], pageTitle),
		AuthorName:       firstNonEmpty(tags["article:author"], tags["twitter:creator"], tags["og:site_name"]),
		ThumbnailURL:     resolveMetadataURL(parsedURL, firstNonEmpty(tags["og:image"], tags["twitter:image"])),
		DurationVerified: false,
		EmbedURL:         resolveMetadataURL(parsedURL, firstNonEmpty(tags["og:video:url"], tags["twitter:player"], canonicalURL, pageURL)),
		Raw: map[string]any{
			"source_url": pageURL,
			"title":      pageTitle,
			"canonical":  canonicalURL,
			"open_graph": tags,
		},
	}

	return metadata, nil
}

func (f *httpExternalMetadataFetcher) fetchBody(ctx context.Context, rawURL string) ([]byte, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, rawURL, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Accept", "text/html,application/xhtml+xml,application/xml;q=0.9,*/*;q=0.8")
	req.Header.Set("User-Agent", "clpr-external-metadata/1.0")

	resp, err := f.client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return nil, fmt.Errorf("unexpected status %s", resp.Status)
	}

	return io.ReadAll(io.LimitReader(resp.Body, 1<<20))
}

func buildOEmbedURL(pageURL string) (string, error) {
	parsed, err := url.Parse(pageURL)
	if err != nil {
		return "", err
	}

	oembedURL := &url.URL{Scheme: parsed.Scheme, Host: parsed.Host, Path: "/oembed"}
	query := oembedURL.Query()
	query.Set("url", pageURL)
	query.Set("format", "json")
	oembedURL.RawQuery = query.Encode()
	return oembedURL.String(), nil
}

func extractOEmbedEmbedURL(htmlSnippet string) string {
	htmlSnippet = strings.TrimSpace(htmlSnippet)
	if htmlSnippet == "" {
		return ""
	}
	match := oEmbedIframeSrcRe.FindStringSubmatch(htmlSnippet)
	if len(match) < 2 {
		return ""
	}
	return strings.TrimSpace(match[1])
}

func parseMetadataTags(body []byte, baseURL *url.URL) (map[string]string, string, string, error) {
	tags := make(map[string]string)
	tokenizer := html.NewTokenizer(strings.NewReader(string(body)))
	var pageTitle string
	var canonicalURL string

	for {
		tt := tokenizer.Next()
		switch tt {
		case html.ErrorToken:
			if tokenizer.Err() == io.EOF {
				if canonicalURL == "" {
					canonicalURL = baseURL.String()
				}
				return tags, pageTitle, canonicalURL, nil
			}
			return nil, "", "", tokenizer.Err()
		case html.StartTagToken, html.SelfClosingTagToken:
			tag := tokenizer.Token()
			switch strings.ToLower(tag.Data) {
			case "meta":
				attrs := tokenAttrMap(tag.Attr)
				key := firstNonEmpty(attrs["property"], attrs["name"], attrs["itemprop"])
				content := strings.TrimSpace(attrs["content"])
				if key != "" && content != "" {
					tags[strings.ToLower(key)] = content
				}
			case "title":
				pageTitle = readTextUntilEndTag(tokenizer, "title")
			case "link":
				attrs := tokenAttrMap(tag.Attr)
				if strings.Contains(strings.ToLower(attrs["rel"]), "canonical") {
					canonicalURL = resolveMetadataURL(baseURL, attrs["href"])
				}
			}
		}
	}
}

func readTextUntilEndTag(tokenizer *html.Tokenizer, endTag string) string {
	var b strings.Builder
	for {
		tt := tokenizer.Next()
		switch tt {
		case html.ErrorToken:
			return strings.TrimSpace(b.String())
		case html.TextToken:
			b.WriteString(tokenizer.Token().Data)
		case html.EndTagToken:
			if strings.EqualFold(tokenizer.Token().Data, endTag) {
				return strings.TrimSpace(b.String())
			}
		}
	}
}

func tokenAttrMap(attrs []html.Attribute) map[string]string {
	result := make(map[string]string, len(attrs))
	for _, attr := range attrs {
		result[strings.ToLower(attr.Key)] = strings.TrimSpace(attr.Val)
	}
	return result
}

func firstNonEmpty(values ...string) string {
	for _, value := range values {
		if trimmed := strings.TrimSpace(value); trimmed != "" {
			return trimmed
		}
	}
	return ""
}

func resolveMetadataURL(baseURL *url.URL, raw string) string {
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return ""
	}
	parsed, err := url.Parse(raw)
	if err != nil {
		return raw
	}
	if parsed.IsAbs() {
		return parsed.String()
	}
	if baseURL == nil {
		return parsed.String()
	}
	return baseURL.ResolveReference(parsed).String()
}

func stringFromMap(values map[string]any, key string) string {
	if values == nil {
		return ""
	}
	value, ok := values[key]
	if !ok || value == nil {
		return ""
	}
	switch v := value.(type) {
	case string:
		return v
	default:
		return fmt.Sprint(v)
	}
}

func encodeExternalMetadata(source DetectedSource, metadata ExternalMetadata) ([]byte, error) {
	payload := map[string]any{
		"source_url":        source.NormalizedURL,
		"source_raw_url":    source.RawURL,
		"source_id":         source.SourceID,
		"source_platform":   source.Platform,
		"title":             metadata.Title,
		"author_name":       metadata.AuthorName,
		"thumbnail_url":     metadata.ThumbnailURL,
		"duration_verified": metadata.DurationVerified,
		"embed_url":         metadata.EmbedURL,
		"raw":               metadata.Raw,
	}
	if metadata.DurationSeconds != nil {
		payload["duration_seconds"] = *metadata.DurationSeconds
	}
	return json.Marshal(payload)
}
