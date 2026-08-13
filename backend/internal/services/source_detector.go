package services

import (
	"fmt"
	"net/url"
	"strings"
)

type SourcePlatform string

const (
	SourcePlatformTwitch        SourcePlatform = "twitch"
	SourcePlatformKick          SourcePlatform = "kick"
	SourcePlatformYouTube       SourcePlatform = "youtube"
	SourcePlatformYouTubeShorts SourcePlatform = "youtube_shorts"
	SourcePlatformTikTok        SourcePlatform = "tiktok"
)

type SourceType string

const (
	SourceTypeTwitch   SourceType = "twitch"
	SourceTypeExternal SourceType = "external"
)

type DetectedSource struct {
	RawURL        string         `json:"raw_url"`
	NormalizedURL string         `json:"normalized_url"`
	Platform      SourcePlatform `json:"platform"`
	SourceType    SourceType     `json:"source_type"`
	SourceID      string         `json:"source_id"`
}

// DetectClipSource identifies the platform and normalizes supported clip URLs.
func DetectClipSource(rawURL string) (DetectedSource, error) {
	rawURL = strings.TrimSpace(rawURL)
	if rawURL == "" {
		return DetectedSource{}, fmt.Errorf("unsupported source/platform")
	}

	parsed, err := url.Parse(rawURL)
	if err != nil || parsed.Scheme == "" || parsed.Host == "" {
		return DetectedSource{}, fmt.Errorf("unsupported source/platform")
	}

	host := strings.ToLower(parsed.Hostname())
	segments, err := splitStrictPathSegments(parsed.EscapedPath())
	if err != nil {
		return DetectedSource{}, fmt.Errorf("unsupported source/platform")
	}

	switch {
	case host == "clips.twitch.tv":
		if len(segments) != 1 {
			return DetectedSource{}, fmt.Errorf("unsupported source/platform")
		}
		return DetectedSource{
			RawURL:        rawURL,
			NormalizedURL: normalizeHTTPSURL("clips.twitch.tv", "/"+segments[0], ""),
			Platform:      SourcePlatformTwitch,
			SourceType:    SourceTypeTwitch,
			SourceID:      segments[0],
		}, nil
	case host == "twitch.tv" || strings.HasSuffix(host, ".twitch.tv"):
		if len(segments) == 3 && segments[1] == "clip" {
			return DetectedSource{
				RawURL:        rawURL,
				NormalizedURL: normalizeHTTPSURL("clips.twitch.tv", "/"+segments[2], ""),
				Platform:      SourcePlatformTwitch,
				SourceType:    SourceTypeTwitch,
				SourceID:      segments[2],
			}, nil
		}
		return DetectedSource{}, fmt.Errorf("unsupported source/platform")
	case host == "kick.com" || strings.HasSuffix(host, ".kick.com"):
		clipID := parsed.Query().Get("clip")
		if clipID == "" || len(segments) != 1 {
			return DetectedSource{}, fmt.Errorf("unsupported source/platform")
		}
		return DetectedSource{
			RawURL:        rawURL,
			NormalizedURL: normalizeHTTPSURL("kick.com", "/"+segments[0], "clip="+url.QueryEscape(clipID)),
			Platform:      SourcePlatformKick,
			SourceType:    SourceTypeExternal,
			SourceID:      clipID,
		}, nil
	case host == "youtube.com" || strings.HasSuffix(host, ".youtube.com"):
		if len(segments) == 2 && segments[0] == "shorts" {
			return DetectedSource{
				RawURL:        rawURL,
				NormalizedURL: normalizeHTTPSURL("www.youtube.com", "/shorts/"+segments[1], ""),
				Platform:      SourcePlatformYouTubeShorts,
				SourceType:    SourceTypeExternal,
				SourceID:      segments[1],
			}, nil
		}
		if len(segments) == 1 && segments[0] == "watch" {
			videoID := parsed.Query().Get("v")
			if videoID == "" {
				return DetectedSource{}, fmt.Errorf("unsupported source/platform")
			}
			return DetectedSource{
				RawURL:        rawURL,
				NormalizedURL: normalizeHTTPSURL("www.youtube.com", "/watch", "v="+url.QueryEscape(videoID)),
				Platform:      SourcePlatformYouTube,
				SourceType:    SourceTypeExternal,
				SourceID:      videoID,
			}, nil
		}
		return DetectedSource{}, fmt.Errorf("unsupported source/platform")
	case host == "youtu.be":
		if len(segments) != 1 {
			return DetectedSource{}, fmt.Errorf("unsupported source/platform")
		}
		return DetectedSource{
			RawURL:        rawURL,
			NormalizedURL: normalizeHTTPSURL("www.youtube.com", "/watch", "v="+url.QueryEscape(segments[0])),
			Platform:      SourcePlatformYouTube,
			SourceType:    SourceTypeExternal,
			SourceID:      segments[0],
		}, nil
	case host == "tiktok.com" || strings.HasSuffix(host, ".tiktok.com"):
		if len(segments) == 3 && strings.HasPrefix(segments[0], "@") && segments[1] == "video" {
			return DetectedSource{
				RawURL:        rawURL,
				NormalizedURL: normalizeHTTPSURL("www.tiktok.com", "/"+segments[0]+"/video/"+segments[2], ""),
				Platform:      SourcePlatformTikTok,
				SourceType:    SourceTypeExternal,
				SourceID:      segments[2],
			}, nil
		}
		return DetectedSource{}, fmt.Errorf("unsupported source/platform")
	case host == "x.com" || host == "twitter.com" || strings.HasSuffix(host, ".x.com") || strings.HasSuffix(host, ".twitter.com"):
		return DetectedSource{}, fmt.Errorf("unsupported source/platform")
	case host == "instagram.com" || strings.HasSuffix(host, ".instagram.com"):
		return DetectedSource{}, fmt.Errorf("unsupported source/platform")
	default:
		return DetectedSource{}, fmt.Errorf("unsupported source/platform")
	}
}

func splitStrictPathSegments(escapedPath string) ([]string, error) {
	if escapedPath == "" || escapedPath == "/" {
		return nil, nil
	}

	trimmed := strings.TrimPrefix(escapedPath, "/")
	if trimmed == "" {
		return nil, nil
	}

	rawSegments := strings.Split(trimmed, "/")
	segments := make([]string, 0, len(rawSegments))
	for _, rawSegment := range rawSegments {
		if rawSegment == "" || rawSegment == "." || rawSegment == ".." {
			return nil, fmt.Errorf("invalid path segment")
		}

		segment, err := url.PathUnescape(rawSegment)
		if err != nil || segment == "" || segment == "." || segment == ".." || strings.Contains(segment, "/") {
			return nil, fmt.Errorf("invalid path segment")
		}

		segments = append(segments, segment)
	}

	return segments, nil
}

func normalizeHTTPSURL(host, pth, rawQuery string) string {
	return (&url.URL{Scheme: "https", Host: host, Path: pth, RawQuery: rawQuery}).String()
}
