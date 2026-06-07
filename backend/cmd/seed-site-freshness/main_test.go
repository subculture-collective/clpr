package main

import (
	"testing"

	"git.subcult.tv/subculture-collective/clpr/internal/models"
)

func TestDefaultSiteFreshnessPresetsWithoutTwitch(t *testing.T) {
	presets := defaultSiteFreshnessPresets(false)

	if len(presets) != 9 {
		t.Fatalf("expected 9 non-Twitch presets, got %d", len(presets))
	}

	for _, preset := range presets {
		if preset.RequiresTwitch {
			t.Fatalf("preset %q unexpectedly requires Twitch", preset.Name)
		}
		if preset.Visibility != models.PlaylistVisibilityPublic {
			t.Fatalf("preset %q should be public, got %q", preset.Name, preset.Visibility)
		}
		if preset.Schedule == "manual" {
			t.Fatalf("preset %q should be scheduled, got manual", preset.Name)
		}
	}
}

func TestDefaultSiteFreshnessPresetsWithTwitch(t *testing.T) {
	presets := defaultSiteFreshnessPresets(true)

	if len(presets) != 11 {
		t.Fatalf("expected 11 presets with Twitch enabled, got %d", len(presets))
	}

	requiresTwitch := 0
	for _, preset := range presets {
		if preset.RequiresTwitch {
			requiresTwitch++
		}
	}

	if requiresTwitch != 2 {
		t.Fatalf("expected 2 Twitch presets, got %d", requiresTwitch)
	}
}

func TestPresetToCreateRequest(t *testing.T) {
	preset := siteFreshnessPreset{
		Name:           "Trending Now",
		Description:    "Auto-generated playlist",
		Sort:           "trending",
		Timeframe:      stringPtr("day"),
		ClipLimit:      25,
		Visibility:     models.PlaylistVisibilityPublic,
		Schedule:       "daily",
		Strategy:       "twitch_trending",
		RetentionDays:  7,
		TitleTemplate:  "Trending Now • {date}",
		RequiresTwitch: true,
	}

	req := preset.toCreateRequest()

	if req.Name != preset.Name {
		t.Fatalf("expected name %q, got %q", preset.Name, req.Name)
	}
	if req.Strategy == nil || *req.Strategy != preset.Strategy {
		t.Fatalf("expected strategy %q, got %#v", preset.Strategy, req.Strategy)
	}
	if req.Schedule == nil || *req.Schedule != preset.Schedule {
		t.Fatalf("expected schedule %q, got %#v", preset.Schedule, req.Schedule)
	}
	if req.RetentionDays == nil || *req.RetentionDays != preset.RetentionDays {
		t.Fatalf("expected retention %d, got %#v", preset.RetentionDays, req.RetentionDays)
	}
	if req.TitleTemplate == nil || *req.TitleTemplate != preset.TitleTemplate {
		t.Fatalf("expected title template %q, got %#v", preset.TitleTemplate, req.TitleTemplate)
	}
	if req.ExcludeNSFW == nil || !*req.ExcludeNSFW {
		t.Fatalf("expected ExcludeNSFW to default to true")
	}
}
