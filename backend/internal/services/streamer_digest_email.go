package services

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	htmltemplate "html/template"
	"net/url"
	"strings"
	texttemplate "text/template"
	"time"

	"git.subcult.tv/subculture-collective/clpr/internal/models"
	"github.com/google/uuid"
)

// ErrEmailDisabled means the email service is turned off.
var ErrEmailDisabled = errors.New("email is disabled")

// streamerDigestCategory tags digest emails in SendGrid, so delivery, opens
// and clicks are counted per template by the email metrics.
const streamerDigestCategory = "streamer_digest"

type streamerDigestView struct {
	Name         string
	WeekLabel    string
	Clips        []streamerDigestClipView
	Placements   []streamerDigestPlacementView
	PackReady    bool
	PackPending  bool
	PackClips    int
	DashboardURL string
	SettingsURL  string
	Unsubscribe  string
}

type streamerDigestClipView struct {
	Title string
	Views string
	URL   string
}

type streamerDigestPlacementView struct {
	Label string
	Title string
	URL   string
}

var streamerDigestHTML = htmltemplate.Must(htmltemplate.New("digest").Parse(`<!DOCTYPE html>
<html>
<head><meta charset="UTF-8"><meta name="viewport" content="width=device-width, initial-scale=1.0"><title>Your week on clpr</title></head>
<body style="font-family: Arial, sans-serif; line-height: 1.5; color: #17141F; max-width: 600px; margin: 0 auto; padding: 20px;">
  <div style="border-top: 6px solid #8C5CFF; padding: 24px 0 8px;">
    <p style="margin: 0; font-size: 12px; letter-spacing: 0.12em; text-transform: uppercase; color: #6B6880;">clpr weekly digest</p>
    <h1 style="margin: 4px 0 0; font-size: 24px;">Your week on clpr, {{.Name}}</h1>
    <p style="margin: 4px 0 0; color: #6B6880;">{{.WeekLabel}}</p>
  </div>
  {{if .Clips}}
  <h2 style="font-size: 18px; margin: 24px 0 8px;">Your top clips</h2>
  <ol style="padding-left: 20px; margin: 0;">
    {{range .Clips}}<li style="margin-bottom: 6px;"><a href="{{.URL}}" style="color: #5B2FD9;">{{.Title}}</a> <span style="color: #6B6880;">· {{.Views}} views</span></li>
    {{end}}
  </ol>
  {{end}}
  {{if .Placements}}
  <h2 style="font-size: 18px; margin: 24px 0 8px;">In the daily countdowns</h2>
  <ul style="padding-left: 20px; margin: 0;">
    {{range .Placements}}<li style="margin-bottom: 6px;"><a href="{{.URL}}" style="color: #5B2FD9;">{{.Label}}</a>: {{.Title}}</li>
    {{end}}
  </ul>
  {{end}}
  <h2 style="font-size: 18px; margin: 24px 0 8px;">Your clip pack</h2>
  {{if .PackReady}}<p style="margin: 0;">Your {{.PackClips}} top clips of the week are ready to post, as captioned vertical videos and a highlight reel. Download them from your dashboard within 7 days.</p>
  {{else if .PackPending}}<p style="margin: 0;">Your clip pack for the week is being made. It will be on your dashboard soon.</p>
  {{else}}<p style="margin: 0;">No clip pack this week.</p>{{end}}
  <p style="margin: 28px 0;"><a href="{{.DashboardURL}}" style="display: inline-block; background: #8C5CFF; color: #0E0C13; padding: 12px 24px; text-decoration: none; font-weight: bold;">Open your creator dashboard</a></p>
  <hr style="border: none; border-top: 1px solid #DDDAE8; margin: 28px 0 12px;">
  <p style="font-size: 12px; color: #6B6880;">You get this email because you feature your clips on clpr.
    {{if .Unsubscribe}}<a href="{{.Unsubscribe}}" style="color: #5B2FD9;">Stop the weekly digest</a> ·{{end}}
    <a href="{{.SettingsURL}}" style="color: #5B2FD9;">Email settings</a></p>
</body>
</html>`))

var streamerDigestText = texttemplate.Must(texttemplate.New("digest").Parse(`Your week on clpr, {{.Name}}
{{.WeekLabel}}
{{if .Clips}}
Your top clips:
{{range $i, $c := .Clips}}{{$c.Title}} ({{$c.Views}} views): {{$c.URL}}
{{end}}{{end}}{{if .Placements}}
In the daily countdowns:
{{range .Placements}}{{.Label}}: {{.Title}} {{.URL}}
{{end}}{{end}}
Your clip pack:
{{if .PackReady}}Your {{.PackClips}} top clips of the week are ready to post, as captioned vertical videos and a highlight reel. Download them from your dashboard within 7 days.{{else if .PackPending}}Your clip pack for the week is being made. It will be on your dashboard soon.{{else}}No clip pack this week.{{end}}

Open your creator dashboard: {{.DashboardURL}}

---
You get this email because you feature your clips on clpr.
{{if .Unsubscribe}}Stop the weekly digest: {{.Unsubscribe}}
{{end}}Email settings: {{.SettingsURL}}
`))

func formatDigestDate(date string, layout string) string {
	t, err := time.Parse("2006-01-02", date)
	if err != nil {
		return date
	}
	return t.Format(layout)
}

func groupThousands(n int64) string {
	s := fmt.Sprint(n)
	if n < 0 {
		return s
	}
	var b strings.Builder
	for i, r := range s {
		if i > 0 && (len(s)-i)%3 == 0 {
			b.WriteByte(',')
		}
		b.WriteRune(r)
	}
	return b.String()
}

// renderStreamerDigest returns the subject, HTML and text of a digest email.
func renderStreamerDigest(baseURL string, d *models.StreamerDigest, unsubscribeURL string) (subject, html, text string, err error) {
	base := strings.TrimRight(baseURL, "/")
	view := streamerDigestView{
		Name:         d.Name,
		WeekLabel:    "Week of " + formatDigestDate(d.WeekStart, "January 2, 2006"),
		PackReady:    d.PackStatus == models.ClipPackReady,
		PackPending:  d.PackStatus == models.ClipPackPending || d.PackStatus == models.ClipPackBuilding,
		PackClips:    d.PackClipCount,
		DashboardURL: base + "/creator/" + url.PathEscape(d.BroadcasterID) + "/dashboard",
		SettingsURL:  base + "/settings",
		Unsubscribe:  unsubscribeURL,
	}
	for _, c := range d.TopClips {
		view.Clips = append(view.Clips, streamerDigestClipView{Title: c.Title, Views: groupThousands(c.ViewCount),
			URL: base + "/clip/" + c.ClipID.String()})
	}
	for _, p := range d.Placements {
		view.Placements = append(view.Placements, streamerDigestPlacementView{
			Label: fmt.Sprintf("#%d in %s, %s", p.Position, p.SeriesName, formatDigestDate(p.EditionDate, "Jan 2")),
			Title: p.Title,
			URL:   base + "/countdowns/" + p.EditionDate + "/" + url.PathEscape(p.SeriesSlug),
		})
	}
	var h, t bytes.Buffer
	if err := streamerDigestHTML.Execute(&h, view); err != nil {
		return "", "", "", err
	}
	if err := streamerDigestText.Execute(&t, view); err != nil {
		return "", "", "", err
	}
	subject = "Your week on clpr: " + view.WeekLabel
	return subject, h.String(), t.String(), nil
}

// SendStreamerDigest sends a weekly streamer digest (#85) to one broadcaster,
// with a link that stops only the digest, and logs it. The caller decides who
// gets one; this checks only that email is on.
func (s *EmailService) SendStreamerDigest(ctx context.Context, rcpt models.StreamerDigestRecipient, d *models.StreamerDigest) error {
	if !s.enabled {
		return ErrEmailDisabled
	}
	notificationType := models.NotificationTypeStreamerDigest
	unsubscribe := ""
	if token, err := s.generateUnsubscribeToken(ctx, rcpt.UserID, &notificationType); err == nil {
		unsubscribe = fmt.Sprintf("%s/api/v1/notifications/unsubscribe?token=%s", strings.TrimRight(s.baseURL, "/"), token)
	} else {
		return fmt.Errorf("create the unsubscribe link: %w", err)
	}
	subject, htmlBody, textBody, err := renderStreamerDigest(s.baseURL, d, unsubscribe)
	if err != nil {
		return fmt.Errorf("render the digest: %w", err)
	}
	now := time.Now()
	logEntry := &models.EmailNotificationLog{ID: uuid.New(), UserID: rcpt.UserID, NotificationType: notificationType,
		RecipientEmail: rcpt.Email, Subject: subject, Status: models.EmailStatusPending, CreatedAt: now, UpdatedAt: now}
	if err := s.repo.CreateLog(ctx, logEntry); err != nil {
		return fmt.Errorf("log the digest: %w", err)
	}
	messageID, sendErr := s.sendViaSendGrid(rcpt.Email, subject, htmlBody, textBody, streamerDigestCategory)
	logEntry.UpdatedAt = time.Now()
	if sendErr != nil {
		logEntry.Status = models.EmailStatusFailed
		message := sendErr.Error()
		logEntry.ErrorMessage = &message
		_ = s.repo.UpdateLog(ctx, logEntry)
		return fmt.Errorf("send the digest: %w", sendErr)
	}
	logEntry.Status = models.EmailStatusSent
	logEntry.ProviderMessageID = &messageID
	logEntry.SentAt = &logEntry.UpdatedAt
	if err := s.repo.UpdateLog(ctx, logEntry); err != nil {
		s.logger.Error("Failed to update the digest email log", err, map[string]interface{}{"log_id": logEntry.ID.String()})
	}
	return nil
}
