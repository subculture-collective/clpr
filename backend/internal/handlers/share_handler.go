package handlers

import (
	"bytes"
	"context"
	"database/sql"
	"errors"
	"fmt"
	"html/template"
	"log"
	"math"
	"net/http"
	"net/url"
	"regexp"
	"strings"

	"git.subcult.tv/subculture-collective/clpr/internal/models"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
)

// ClipLookupForShare defines the clip repository methods needed by ShareHandler.
type ClipLookupForShare interface {
	GetByID(ctx context.Context, id uuid.UUID) (*models.Clip, error)
	GetByTwitchClipID(ctx context.Context, twitchClipID string) (*models.Clip, error)
}

// ShareHandler renders small link-preview documents for chat and social
// crawlers (Discord, Slack, X, ...). The frontend edge proxies known preview
// user agents here for /clip/:id so shared links show the clip instead of the
// site-wide defaults baked into the SPA shell.
type ShareHandler struct {
	clipRepo ClipLookupForShare
}

// NewShareHandler creates a new ShareHandler.
func NewShareHandler(clipRepo ClipLookupForShare) *ShareHandler {
	return &ShareHandler{clipRepo: clipRepo}
}

// shareClipIDPattern matches database UUIDs and Twitch clip slugs.
var shareClipIDPattern = regexp.MustCompile(`^[A-Za-z0-9_-]{1,128}$`)

const (
	shareSiteName           = "clpr"
	shareDefaultTitle       = "clpr - Discover Creators and Live Moments"
	shareDefaultDescription = "Discover the creators and moments shaping live culture. Browse Twitch clips by creator, topic, tag, or collection."
	shareCacheControl       = "public, max-age=300"
)

type clipPreview struct {
	Title        string
	Description  string
	PageURL      string
	ImageURL     string
	ImageAlt     string
	IsDefault    bool
	DefaultImage bool
	LinkText     string
}

var clipPreviewTemplate = template.Must(template.New("clip_preview").Parse(`<!doctype html>
<html lang="en">
<head>
<meta charset="utf-8">
<meta name="viewport" content="width=device-width, initial-scale=1">
<title>{{.Title}}</title>
<meta name="description" content="{{.Description}}">
<link rel="canonical" href="{{.PageURL}}">
<meta property="og:site_name" content="clpr">
<meta property="og:locale" content="en_US">
<meta property="og:type" content="{{if .IsDefault}}website{{else}}video.other{{end}}">
<meta property="og:url" content="{{.PageURL}}">
<meta property="og:title" content="{{.Title}}">
<meta property="og:description" content="{{.Description}}">
<meta property="og:image" content="{{.ImageURL}}">
<meta property="og:image:secure_url" content="{{.ImageURL}}">
<meta property="og:image:alt" content="{{.ImageAlt}}">
{{- if .DefaultImage}}
<meta property="og:image:type" content="image/png">
<meta property="og:image:width" content="1200">
<meta property="og:image:height" content="630">
{{- end}}
<meta name="twitter:card" content="summary_large_image">
<meta name="twitter:site" content="@clpr_tv">
<meta name="twitter:url" content="{{.PageURL}}">
<meta name="twitter:title" content="{{.Title}}">
<meta name="twitter:description" content="{{.Description}}">
<meta name="twitter:image" content="{{.ImageURL}}">
<meta name="twitter:image:alt" content="{{.ImageAlt}}">
<meta name="theme-color" content="#0E0C13">
</head>
<body>
<h1>{{.Title}}</h1>
<p>{{.Description}}</p>
<p><a href="{{.PageURL}}">{{.LinkText}}</a></p>
</body>
</html>
`))

// GetClipPreview handles GET /api/v1/share/clips/:id.
//
// It always returns an HTML document with OpenGraph and Twitter card tags.
// Unknown, removed, or hidden clips return 404 with the site-wide defaults so
// crawlers still render a clpr card.
func (h *ShareHandler) GetClipPreview(c *gin.Context) {
	baseURL := getBaseURL(c)
	clipID := c.Param("id")

	if !shareClipIDPattern.MatchString(clipID) {
		h.render(c, http.StatusNotFound, defaultClipPreview(baseURL))
		return
	}

	clip, err := h.lookup(c.Request.Context(), clipID)
	if err != nil {
		if isNotFound(err) {
			h.render(c, http.StatusNotFound, defaultClipPreview(baseURL))
			return
		}
		log.Printf("Error loading clip %q for share preview: %v", clipID, err)
		c.Header("Cache-Control", "no-store")
		h.renderWithoutCache(c, http.StatusServiceUnavailable, defaultClipPreview(baseURL))
		return
	}
	if clip.IsRemoved || clip.IsHidden || clip.DMCARemoved {
		h.render(c, http.StatusNotFound, defaultClipPreview(baseURL))
		return
	}

	h.render(c, http.StatusOK, buildClipPreview(baseURL, clip))
}

func (h *ShareHandler) lookup(ctx context.Context, clipID string) (*models.Clip, error) {
	if id, err := uuid.Parse(clipID); err == nil {
		return h.clipRepo.GetByID(ctx, id)
	}
	return h.clipRepo.GetByTwitchClipID(ctx, clipID)
}

func isNotFound(err error) bool {
	return errors.Is(err, pgx.ErrNoRows) || errors.Is(err, sql.ErrNoRows)
}

func (h *ShareHandler) render(c *gin.Context, status int, preview clipPreview) {
	c.Header("Cache-Control", shareCacheControl)
	h.renderWithoutCache(c, status, preview)
}

func (h *ShareHandler) renderWithoutCache(c *gin.Context, status int, preview clipPreview) {
	var body bytes.Buffer
	if err := clipPreviewTemplate.Execute(&body, preview); err != nil {
		log.Printf("Error rendering share preview: %v", err)
		c.Header("Cache-Control", "no-store")
		c.Data(http.StatusInternalServerError, "text/plain; charset=utf-8", []byte("preview unavailable\n"))
		return
	}
	c.Header("X-Robots-Tag", "noindex")
	c.Data(status, "text/html; charset=utf-8", body.Bytes())
}

func defaultClipPreview(baseURL string) clipPreview {
	return clipPreview{
		Title:        shareDefaultTitle,
		Description:  shareDefaultDescription,
		PageURL:      baseURL + "/",
		ImageURL:     baseURL + "/social-card.png",
		ImageAlt:     "clpr — Discover creators and live moments",
		IsDefault:    true,
		DefaultImage: true,
		LinkText:     "Open clpr",
	}
}

func buildClipPreview(baseURL string, clip *models.Clip) clipPreview {
	title := strings.TrimSpace(clip.Title)
	if title == "" {
		title = "Twitch clip"
	}
	pageID := clip.ID.String()
	if clip.ID == uuid.Nil && clip.TwitchClipID != "" {
		pageID = clip.TwitchClipID
	}

	preview := clipPreview{
		Title:       title,
		Description: clipPreviewDescription(clip),
		PageURL:     baseURL + "/clip/" + url.PathEscape(pageID),
		ImageURL:    baseURL + "/social-card.png",
		ImageAlt:    title,
		LinkText:    "Watch on clpr",
	}
	if image := absoluteImageURL(baseURL, clip.ThumbnailURL); image != "" && !clip.IsNSFW {
		preview.ImageURL = image
	} else {
		preview.DefaultImage = true
	}
	return preview
}

func clipPreviewDescription(clip *models.Clip) string {
	parts := make([]string, 0, 4)
	if name := strings.TrimSpace(clip.BroadcasterName); name != "" {
		parts = append(parts, "Clip from "+name)
	}
	if clip.GameName != nil && strings.TrimSpace(*clip.GameName) != "" {
		parts = append(parts, strings.TrimSpace(*clip.GameName))
	}
	parts = append(parts, formatCount(clip.ViewCount, "view"))
	if clip.Duration != nil && *clip.Duration > 0 {
		parts = append(parts, formatClipDuration(*clip.Duration))
	}
	return strings.Join(parts, " · ") + " — on " + shareSiteName
}

func formatCount(count int, noun string) string {
	if count == 1 {
		return "1 " + noun
	}
	return fmt.Sprintf("%s %ss", groupThousands(count), noun)
}

func groupThousands(n int) string {
	sign := ""
	if n < 0 {
		sign = "-"
		n = -n
	}
	digits := fmt.Sprintf("%d", n)
	var out strings.Builder
	for i, r := range digits {
		if i > 0 && (len(digits)-i)%3 == 0 {
			out.WriteByte(',')
		}
		out.WriteRune(r)
	}
	return sign + out.String()
}

func formatClipDuration(seconds float64) string {
	total := int(math.Round(seconds))
	if total < 60 {
		return fmt.Sprintf("%ds", total)
	}
	return fmt.Sprintf("%d:%02d", total/60, total%60)
}

// absoluteImageURL returns an absolute HTTP(S) image URL, resolving
// root-relative paths against the public base URL. Other schemes are dropped.
func absoluteImageURL(baseURL string, raw *string) string {
	if raw == nil {
		return ""
	}
	value := strings.TrimSpace(*raw)
	if value == "" {
		return ""
	}
	if strings.HasPrefix(value, "/") && !strings.HasPrefix(value, "//") {
		return baseURL + value
	}
	parsed, err := url.Parse(value)
	if err != nil || parsed.Host == "" || (parsed.Scheme != "https" && parsed.Scheme != "http") {
		return ""
	}
	return parsed.String()
}
