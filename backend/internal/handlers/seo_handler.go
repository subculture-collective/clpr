package handlers

import (
	"context"
	"fmt"
	"html"
	"log"
	"net/http"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"git.subcult.tv/subculture-collective/clpr/internal/models"
	"git.subcult.tv/subculture-collective/clpr/pkg/utils"
)

// ClipRepositoryForSEO defines the interface for clip repository methods needed by SEO handler
type ClipRepositoryForSEO interface {
	ListForSitemap(ctx context.Context) ([]models.Clip, error)
	ListForSitemapBroadcasters(ctx context.Context) ([]models.BroadcasterWithClipCount, error)
}

// GameRepositoryForSEO defines the interface for game repository methods needed by SEO handler
type GameRepositoryForSEO interface {
	ListAllWithClipCounts(ctx context.Context, limit, offset int) ([]*models.GameWithStats, error)
}

// SEOHandler handles SEO-related endpoints (sitemap, robots.txt)
type SEOHandler struct {
	clipRepo ClipRepositoryForSEO
	gameRepo GameRepositoryForSEO
}

// NewSEOHandler creates a new SEO handler
func NewSEOHandler(clipRepo ClipRepositoryForSEO, gameRepo GameRepositoryForSEO) *SEOHandler {
	return &SEOHandler{
		clipRepo: clipRepo,
		gameRepo: gameRepo,
	}
}

// GetSitemap generates and returns an XML sitemap
func (h *SEOHandler) GetSitemap(c *gin.Context) {
	// Get all clips with basic info for sitemap
	ctx := c.Request.Context()
	clips, err := h.clipRepo.ListForSitemap(ctx)
	if err != nil {
		log.Printf("Error fetching clips for sitemap: %v", err)
		c.String(http.StatusInternalServerError, "Error generating sitemap")
		return
	}

	// Base URL from config or environment
	baseURL := c.GetString("base_url")
	if baseURL == "" {
		baseURL = "https://clpr.app" // Default, should be configured
	}

	// Build sitemap XML using strings.Builder for better performance
	var builder strings.Builder
	builder.WriteString(`<?xml version="1.0" encoding="UTF-8"?>
<urlset xmlns="http://www.sitemaps.org/schemas/sitemap/0.9">
`)

	// Static pages
	staticPages := []struct {
		path       string
		priority   string
		changefreq string
	}{
		{"/", "1.0", "daily"},
		{"/discover", "0.9", "daily"},
		{"/new", "0.9", "daily"},
		{"/top", "0.9", "daily"},
		{"/rising", "0.9", "daily"},
		{"/search", "0.8", "weekly"},
		{"/leaderboards", "0.7", "weekly"},
		{"/about", "0.5", "monthly"},
		{"/community-rules", "0.5", "monthly"},
		{"/terms", "0.4", "monthly"},
		{"/privacy", "0.4", "monthly"},
		{"/pricing", "0.6", "weekly"},
	}

	for _, page := range staticPages {
		builder.WriteString(fmt.Sprintf(`  <url>
    <loc>%s%s</loc>
    <changefreq>%s</changefreq>
    <priority>%s</priority>
  </url>
`, baseURL, page.path, page.changefreq, page.priority))
	}

	// Best-of pages (static temporal)
	bestOfPages := []struct {
		path     string
		priority string
	}{
		{"/clips/best/this-week", "0.9"},
		{"/clips/best/this-month", "0.9"},
	}
	for _, page := range bestOfPages {
		builder.WriteString(fmt.Sprintf(`  <url>
    <loc>%s%s</loc>
    <changefreq>daily</changefreq>
    <priority>%s</priority>
  </url>
`, baseURL, page.path, page.priority))
	}

	// Monthly archive pages (last 12 months)
	now := time.Now().UTC()
	for i := 0; i < 12; i++ {
		t := now.AddDate(0, -i, 0)
		path := fmt.Sprintf("/clips/best/%d/%02d", t.Year(), t.Month())
		builder.WriteString(fmt.Sprintf(`  <url>
    <loc>%s%s</loc>
    <changefreq>monthly</changefreq>
    <priority>0.7</priority>
  </url>
`, baseURL, path))
	}

	// Streamer pSEO pages
	broadcasters, err := h.clipRepo.ListForSitemapBroadcasters(ctx)
	if err != nil {
		log.Printf("Error fetching broadcasters for sitemap: %v", err)
	} else {
		for _, b := range broadcasters {
			builder.WriteString(fmt.Sprintf(`  <url>
    <loc>%s/clips/streamer/%s</loc>
    <changefreq>weekly</changefreq>
    <priority>0.7</priority>
  </url>
`, baseURL, html.EscapeString(b.BroadcasterName)))
		}
	}

	// Game pSEO pages
	if h.gameRepo != nil {
		games, err := h.gameRepo.ListAllWithClipCounts(ctx, 500, 0)
		if err != nil {
			log.Printf("Error fetching games for sitemap: %v", err)
		} else {
			for _, g := range games {
				slug := utils.Slugify(g.Name)
				if slug == "" {
					continue
				}
				builder.WriteString(fmt.Sprintf(`  <url>
    <loc>%s/clips/game/%s</loc>
    <changefreq>weekly</changefreq>
    <priority>0.7</priority>
  </url>
`, baseURL, slug))
			}
		}
	}

	// Dynamic clip pages
	for _, clip := range clips {
		builder.WriteString(fmt.Sprintf(`  <url>
    <loc>%s/clip/%s</loc>
    <lastmod>%s</lastmod>
    <changefreq>weekly</changefreq>
    <priority>0.8</priority>
  </url>
`, baseURL, clip.ID, clip.CreatedAt.Format(time.RFC3339)))
	}

	builder.WriteString(`</urlset>`)

	c.Header("Content-Type", "application/xml")
	c.String(http.StatusOK, builder.String())
}

// GetRobotsTxt returns the robots.txt file
func (h *SEOHandler) GetRobotsTxt(c *gin.Context) {
	// Get environment and base URL from context
	env := c.GetString("environment")
	baseURL := c.GetString("base_url")
	if baseURL == "" {
		baseURL = "https://clpr.app" // Default, should be configured
	}

	var robotsTxt string
	if env == "production" {
		// Production: Allow all crawlers
		robotsTxt = fmt.Sprintf(`User-agent: *
Allow: /

# Sitemap location
Sitemap: %s/sitemap.xml
# Disallow admin and API routes
Disallow: /api/
Disallow: /admin/
Disallow: /auth/
Disallow: /settings
Disallow: /profile
Disallow: /favorites
Disallow: /notifications
Disallow: /submit
Disallow: /submissions

# Crawl-delay for polite crawling
Crawl-delay: 1
`, baseURL)
	} else {
		// Non-production: Disallow all
		robotsTxt = `User-agent: *
Disallow: /

# This is a staging/development environment
`
	}

	c.Header("Content-Type", "text/plain")
	c.String(http.StatusOK, robotsTxt)
}
