package services

import "testing"

func TestExtractClipLinks(t *testing.T) {
	input := "watch https://clpr.tv/clip/123e4567-e89b-12d3-a456-426614174000 and https://clips.twitch.tv/FunnySlug"
	links := ExtractClipLinks(input)
	if len(links) != 2 {
		t.Fatalf("expected 2 links, got %d", len(links))
	}
	if links[0].SourceType != "clpr" || links[0].ClprClipID != "123e4567-e89b-12d3-a456-426614174000" {
		t.Fatalf("unexpected first link: %#v", links[0])
	}
	if links[1].SourceType != "twitch" || links[1].TwitchClipID != "FunnySlug" {
		t.Fatalf("unexpected second link: %#v", links[1])
	}
}

func TestExtractClipLinksDeduplicates(t *testing.T) {
	input := "https://clips.twitch.tv/FunnySlug https://clips.twitch.tv/FunnySlug"
	links := ExtractClipLinks(input)
	if len(links) != 1 {
		t.Fatalf("expected 1 link, got %d", len(links))
	}
}
