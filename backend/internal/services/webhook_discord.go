package services

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/url"
	"regexp"
	"strings"
	"time"
	"unicode/utf8"

	"git.subcult.tv/subculture-collective/clpr/internal/models"
)

// WebhookClip is one clip in a countdown or daily top-clips event.
type WebhookClip struct {
	Position        int    `json:"position"`
	Title           string `json:"title"`
	BroadcasterName string `json:"broadcaster_name"`
	URL             string `json:"url"`
	TwitchURL       string `json:"twitch_url"`
	ThumbnailURL    string `json:"thumbnail_url,omitempty"`
	ViewCount       int64  `json:"view_count,omitempty"`
}

// CountdownPublishedData is the data of a countdown.published event.
type CountdownPublishedData struct {
	EditionDate string        `json:"edition_date"`
	SeriesSlug  string        `json:"series_slug"`
	SeriesName  string        `json:"series_name"`
	Size        int           `json:"size"`
	URL         string        `json:"url"`
	Clips       []WebhookClip `json:"clips"`
}

// DailyTopClipsData is the data of a daily.top_clips event.
type DailyTopClipsData struct {
	Date  string        `json:"date"`
	Clips []WebhookClip `json:"clips"`
}

// webhookEventData turns typed event data into the map TriggerEvent takes.
func webhookEventData(v any) (map[string]interface{}, error) {
	raw, err := json.Marshal(v)
	if err != nil {
		return nil, err
	}
	var data map[string]interface{}
	return data, json.Unmarshal(raw, &data)
}

var (
	discordWebhookPath  = regexp.MustCompile(`^/api/webhooks/[0-9]{1,20}/[A-Za-z0-9_-]{1,100}$`)
	discordWebhookHosts = map[string]bool{
		"discord.com": true, "discordapp.com": true, "ptb.discord.com": true, "canary.discord.com": true,
	}
	errDiscordWebhookURL = errors.New("a Discord webhook URL looks like https://discord.com/api/webhooks/<id>/<token>")
)

// validateDiscordWebhookURL accepts only Discord's own webhook endpoints.
func validateDiscordWebhookURL(raw string) error {
	u, err := url.Parse(raw)
	if err != nil || u.Scheme != "https" || u.User != nil || u.Port() != "" || u.RawQuery != "" || u.Fragment != "" ||
		!discordWebhookHosts[strings.ToLower(u.Host)] || !discordWebhookPath.MatchString(u.Path) {
		return errDiscordWebhookURL
	}
	return nil
}

type discordImage struct {
	URL string `json:"url"`
}

type discordFooter struct {
	Text string `json:"text"`
}

type discordEmbed struct {
	Title       string         `json:"title,omitempty"`
	URL         string         `json:"url,omitempty"`
	Description string         `json:"description,omitempty"`
	Color       int            `json:"color,omitempty"`
	Thumbnail   *discordImage  `json:"thumbnail,omitempty"`
	Image       *discordImage  `json:"image,omitempty"`
	Footer      *discordFooter `json:"footer,omitempty"`
}

type discordMessage struct {
	Username        string         `json:"username"`
	Content         string         `json:"content,omitempty"`
	Embeds          []discordEmbed `json:"embeds,omitempty"`
	AllowedMentions struct {
		Parse []string `json:"parse"`
	} `json:"allowed_mentions"`
}

const (
	discordColor          = 0x9146FF
	discordFooterText     = "clpr.tv · every clip belongs to its broadcaster"
	discordTitleLimit     = 256
	discordDescriptionCap = 4000
	discordMaxEmbeds      = 10
)

var discordMarkdown = strings.NewReplacer(
	`\`, `\\`, `*`, `\*`, `_`, `\_`, `~`, `\~`, "`", "\\`", `|`, `\|`, `>`, `\>`,
	`[`, `\[`, `]`, `\]`, `(`, `\(`, `)`, `\)`, `#`, `\#`, `@`, "@​",
)

// discordText escapes markdown and mentions in text from clips.
func discordText(s string) string {
	return discordMarkdown.Replace(strings.Join(strings.Fields(s), " "))
}

func cutRunes(s string, limit int) string {
	if utf8.RuneCountInString(s) <= limit {
		return s
	}
	r := []rune(s)
	return strings.TrimSpace(string(r[:limit-1])) + "…"
}

// httpsURL keeps a link Discord will accept in an embed.
func httpsURL(raw string) string {
	u, err := url.Parse(raw)
	if err != nil || u.Scheme != "https" || u.Host == "" {
		return ""
	}
	return raw
}

func discordDate(date string) string {
	if t, err := time.Parse("2006-01-02", date); err == nil {
		return t.Format("Jan 2, 2006")
	}
	return date
}

// discordBody turns a stored event payload into a Discord message.
func discordBody(eventType, payload string) ([]byte, error) {
	var event struct {
		Data json.RawMessage `json:"data"`
	}
	if err := json.Unmarshal([]byte(payload), &event); err != nil {
		return nil, fmt.Errorf("decode event payload: %w", err)
	}
	msg := discordMessage{Username: "clpr"}
	msg.AllowedMentions.Parse = []string{}

	switch eventType {
	case models.WebhookEventCountdownPublished:
		var data CountdownPublishedData
		if err := json.Unmarshal(event.Data, &data); err != nil {
			return nil, fmt.Errorf("decode countdown event: %w", err)
		}
		var lines strings.Builder
		for _, clip := range data.Clips {
			line := fmt.Sprintf("**#%d** %s · %s\n", clip.Position, discordLink(cutRunes(clip.Title, 100), clip.URL), discordText(clip.BroadcasterName))
			if lines.Len()+len(line) > discordDescriptionCap {
				break
			}
			lines.WriteString(line)
		}
		embed := discordEmbed{
			Title:       cutRunes(fmt.Sprintf("Top %d: %s · %s", data.Size, data.SeriesName, discordDate(data.EditionDate)), discordTitleLimit),
			URL:         httpsURL(data.URL),
			Description: strings.TrimSpace(lines.String()),
			Color:       discordColor,
			Footer:      &discordFooter{Text: discordFooterText},
		}
		if len(data.Clips) > 0 {
			if thumb := httpsURL(data.Clips[0].ThumbnailURL); thumb != "" {
				embed.Image = &discordImage{URL: thumb}
			}
		}
		msg.Embeds = []discordEmbed{embed}

	case models.WebhookEventDailyTopClips:
		var data DailyTopClipsData
		if err := json.Unmarshal(event.Data, &data); err != nil {
			return nil, fmt.Errorf("decode daily event: %w", err)
		}
		msg.Content = fmt.Sprintf("**Top Twitch clips of %s**", discordDate(data.Date))
		for _, clip := range data.Clips {
			if len(msg.Embeds) == discordMaxEmbeds {
				break
			}
			embed := discordEmbed{
				Title:       cutRunes(fmt.Sprintf("#%d %s", clip.Position, strings.Join(strings.Fields(clip.Title), " ")), discordTitleLimit),
				URL:         httpsURL(clip.URL),
				Description: discordText(clip.BroadcasterName),
				Color:       discordColor,
			}
			if clip.ViewCount > 0 {
				embed.Description += fmt.Sprintf(" · %d views", clip.ViewCount)
			}
			if thumb := httpsURL(clip.ThumbnailURL); thumb != "" {
				embed.Thumbnail = &discordImage{URL: thumb}
			}
			msg.Embeds = append(msg.Embeds, embed)
		}
		if len(msg.Embeds) > 0 {
			msg.Embeds[len(msg.Embeds)-1].Footer = &discordFooter{Text: discordFooterText}
		}

	default:
		return nil, fmt.Errorf("event %s has no Discord format", eventType)
	}
	return json.Marshal(msg)
}

func discordLink(text, link string) string {
	if link = httpsURL(link); link == "" {
		return discordText(text)
	}
	return fmt.Sprintf("[%s](%s)", discordText(text), link)
}
