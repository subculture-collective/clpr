package services

import (
	"net/url"
	"regexp"
	"strings"
)

type DetectedClipLink struct {
	SourceURL    string
	SourceType   string
	ClprClipID   string
	TwitchClipID string
}

var httpURLPattern = regexp.MustCompile(`https?://[^\s<>()]+`)

func ExtractClipLinks(message string) []DetectedClipLink {
	matches := httpURLPattern.FindAllString(message, -1)
	seen := map[string]bool{}
	links := make([]DetectedClipLink, 0, len(matches))

	for _, raw := range matches {
		parsed, err := url.Parse(strings.TrimRight(raw, ".,!?)]}"))
		if err != nil || parsed.Host == "" {
			continue
		}

		host := strings.ToLower(parsed.Host)
		path := strings.Trim(parsed.Path, "/")
		parts := strings.Split(path, "/")

		// Normalize so the same clip posted with a different query string,
		// fragment or host casing is recognized as the same link.
		cleaned := strings.ToLower(parsed.Scheme) + "://" + host + "/" + path
		if seen[cleaned] {
			continue
		}

		if (host == "clpr.tv" || strings.HasSuffix(host, ".clpr.tv")) && len(parts) == 2 && (parts[0] == "clip" || parts[0] == "clips") {
			seen[cleaned] = true
			links = append(links, DetectedClipLink{SourceURL: cleaned, SourceType: "clpr", ClprClipID: parts[1]})
			continue
		}

		if host == "clips.twitch.tv" && len(parts) >= 1 && parts[0] != "" {
			seen[cleaned] = true
			links = append(links, DetectedClipLink{SourceURL: cleaned, SourceType: "twitch", TwitchClipID: parts[0]})
			continue
		}

		if (host == "www.twitch.tv" || host == "twitch.tv") && len(parts) >= 3 && parts[1] == "clip" {
			seen[cleaned] = true
			links = append(links, DetectedClipLink{SourceURL: cleaned, SourceType: "twitch", TwitchClipID: parts[2]})
		}
	}

	return links
}
