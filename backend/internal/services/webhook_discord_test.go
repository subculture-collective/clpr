package services

import (
	"context"
	"encoding/json"
	"strings"
	"testing"

	"git.subcult.tv/subculture-collective/clpr/internal/models"
	"github.com/google/uuid"
)

func TestValidateDiscordWebhookURL(t *testing.T) {
	good := []string{
		"https://discord.com/api/webhooks/123456789012345678/abcDEF_-123",
		"https://discordapp.com/api/webhooks/1/x",
		"https://canary.discord.com/api/webhooks/1/x",
	}
	for _, u := range good {
		if err := validateDiscordWebhookURL(u); err != nil {
			t.Errorf("%s rejected: %v", u, err)
		}
	}
	bad := []string{
		"http://discord.com/api/webhooks/1/x",
		"https://discord.com.evil.example/api/webhooks/1/x",
		"https://evil.example/api/webhooks/1/x",
		"https://discord.com:8443/api/webhooks/1/x",
		"https://user@discord.com/api/webhooks/1/x",
		"https://discord.com/api/webhooks/1/x?wait=true",
		"https://discord.com/api/webhooks/abc/x",
		"https://discord.com/api/webhooks/1/x/slack",
		"https://discord.com/api/v10/channels/1/messages",
	}
	for _, u := range bad {
		if err := validateDiscordWebhookURL(u); err == nil {
			t.Errorf("%s accepted", u)
		}
	}
}

func TestDiscordTargetEvents(t *testing.T) {
	s := &OutboundWebhookService{}
	if err := s.validateTargetEvents(models.WebhookTargetDiscord, []string{models.WebhookEventClipApproved}); err == nil {
		t.Fatal("a Discord target accepted clip.approved, which has no Discord format")
	}
	if err := s.validateTargetEvents(models.WebhookTargetDiscord, models.GetDiscordWebhookEvents()); err != nil {
		t.Fatal(err)
	}
	if err := s.validateTargetEvents(models.WebhookTargetGeneric, []string{models.WebhookEventDailyTopClips, models.WebhookEventClipApproved}); err != nil {
		t.Fatal(err)
	}
}

func eventPayload(t *testing.T, eventType string, data any) string {
	t.Helper()
	m, err := webhookEventData(data)
	if err != nil {
		t.Fatal(err)
	}
	raw, err := json.Marshal(models.WebhookEventPayload{Event: eventType, Data: m})
	if err != nil {
		t.Fatal(err)
	}
	return string(raw)
}

func TestDiscordBodyCountdown(t *testing.T) {
	payload := eventPayload(t, models.WebhookEventCountdownPublished, CountdownPublishedData{
		EditionDate: "2026-11-07", SeriesSlug: "overall", SeriesName: "Overall", Size: 2,
		URL: "https://clpr.tv/countdowns/2026-11-07/overall",
		Clips: []WebhookClip{
			{Position: 1, Title: "@everyone [click](https://evil.example) *ace*", BroadcasterName: "alpha_one", URL: "https://clpr.tv/clip/1", ThumbnailURL: "https://clips-media.example/1.jpg"},
			{Position: 2, Title: "plain", BroadcasterName: "bravo", URL: "javascript:alert(1)"},
		},
	})
	raw, err := discordBody(models.WebhookEventCountdownPublished, payload)
	if err != nil {
		t.Fatal(err)
	}
	var msg discordMessage
	if err := json.Unmarshal(raw, &msg); err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(string(raw), `"allowed_mentions":{"parse":[]}`) {
		t.Fatalf("mentions are not disabled: %s", raw)
	}
	if len(msg.Embeds) != 1 {
		t.Fatalf("embeds = %d", len(msg.Embeds))
	}
	e := msg.Embeds[0]
	if e.Title != "Top 2: Overall · Nov 7, 2026" || e.URL != "https://clpr.tv/countdowns/2026-11-07/overall" {
		t.Fatalf("embed = %+v", e)
	}
	if e.Image == nil || e.Image.URL != "https://clips-media.example/1.jpg" {
		t.Fatalf("image = %+v", e.Image)
	}
	wantFirst := `**#1** [@` + "​" + `everyone \[click\]\(https://evil.example\) \*ace\*](https://clpr.tv/clip/1) · alpha\_one`
	lines := strings.Split(e.Description, "\n")
	if len(lines) != 2 || lines[0] != wantFirst || lines[1] != "**#2** plain · bravo" {
		t.Fatalf("description lines = %q", lines)
	}
}

func TestDiscordBodyDailyTopClips(t *testing.T) {
	clips := make([]WebhookClip, 12)
	for i := range clips {
		clips[i] = WebhookClip{Position: i + 1, Title: strings.Repeat("long ", 80), BroadcasterName: "b", URL: "https://clpr.tv/clip/x",
			ThumbnailURL: "https://clips-media.example/x.jpg", ViewCount: 1200}
	}
	raw, err := discordBody(models.WebhookEventDailyTopClips, eventPayload(t, models.WebhookEventDailyTopClips, DailyTopClipsData{Date: "2026-11-07", Clips: clips}))
	if err != nil {
		t.Fatal(err)
	}
	var msg discordMessage
	if err := json.Unmarshal(raw, &msg); err != nil {
		t.Fatal(err)
	}
	if msg.Content != "**Top Twitch clips of Nov 7, 2026**" || len(msg.Embeds) != 10 {
		t.Fatalf("content %q, %d embeds", msg.Content, len(msg.Embeds))
	}
	for i, e := range msg.Embeds {
		if len([]rune(e.Title)) > discordTitleLimit || e.Thumbnail == nil || e.Description != "b · 1200 views" {
			t.Fatalf("embed %d = %+v", i, e)
		}
	}
	if msg.Embeds[9].Footer == nil || msg.Embeds[0].Footer != nil {
		t.Fatal("the footer belongs on the last embed only")
	}
	if _, err := discordBody(models.WebhookEventClipApproved, `{"data":{}}`); err == nil {
		t.Fatal("clip.approved got a Discord body")
	}
}

type fakePublicReader struct {
	edition *models.PublicCountdownEdition
}

func (f fakePublicReader) GetPublicEdition(context.Context, string, string) (*models.PublicCountdownEdition, error) {
	return f.edition, nil
}

type fakeTrigger struct {
	eventType string
	eventID   uuid.UUID
	data      map[string]interface{}
}

func (f *fakeTrigger) TriggerEvent(_ context.Context, eventType string, eventID uuid.UUID, data map[string]interface{}) error {
	f.eventType, f.eventID, f.data = eventType, eventID, data
	return nil
}

func TestCountdownPublishNotifier(t *testing.T) {
	clipID := uuid.New()
	thumb := "https://clips-media.example/1.jpg"
	trigger := &fakeTrigger{}
	n := &CountdownPublishNotifier{
		Editions: fakePublicReader{&models.PublicCountdownEdition{EditionDate: "2026-11-07", SeriesSlug: "overall", SeriesName: "Overall", Size: 10,
			Clips: []models.PublicCountdownClip{{Position: 1, ClipID: clipID, Title: "ace", BroadcasterName: "alpha", TwitchClipURL: "https://clips.twitch.tv/x", ThumbnailURL: &thumb}}}},
		Webhooks: trigger,
		BaseURL:  "https://clpr.tv/",
	}
	edition := &models.CountdownEdition{ID: uuid.New(), EditionDate: "2026-11-07", SeriesSlug: "overall"}
	if err := n.EditionPublished(context.Background(), edition); err != nil {
		t.Fatal(err)
	}
	if trigger.eventType != models.WebhookEventCountdownPublished || trigger.eventID != edition.ID {
		t.Fatalf("event %s %s", trigger.eventType, trigger.eventID)
	}
	raw, _ := json.Marshal(trigger.data)
	var data CountdownPublishedData
	if err := json.Unmarshal(raw, &data); err != nil {
		t.Fatal(err)
	}
	if data.URL != "https://clpr.tv/countdowns/2026-11-07/overall" || len(data.Clips) != 1 ||
		data.Clips[0].URL != "https://clpr.tv/clip/"+clipID.String() || data.Clips[0].ThumbnailURL != thumb {
		t.Fatalf("data = %+v", data)
	}

	// An edition that is not public sends nothing.
	trigger.eventType = ""
	n.Editions = fakePublicReader{}
	if err := n.EditionPublished(context.Background(), edition); err != nil || trigger.eventType != "" {
		t.Fatalf("unpublished edition announced: %q, %v", trigger.eventType, err)
	}
}
