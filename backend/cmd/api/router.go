package main

import (
	"fmt"

	"git.subcult.tv/subculture-collective/clpr/internal/middleware"
	"github.com/gin-gonic/gin"
)

// newRouter builds the API engine. Hierarchical tag slugs such as
// "content/funny" travel as one encoded path segment (content%2Ffunny), so
// routes match on the raw path and then unescape parameter values.
func newRouter() *gin.Engine {
	r := gin.New()
	r.UseRawPath = true
	r.UnescapePathValues = true
	return r
}

// publicCache caches anonymous responses of a hot public listing route. Each
// route gets its own instance so in-flight deduplication stays per route.
func publicCache(infra *Infrastructure) gin.HandlerFunc {
	cfg := middleware.PublicResponseCacheConfig{}
	if infra.Config != nil {
		cfg.TTL = infra.Config.Server.PublicCacheTTL
		cfg.StaleTTL = infra.Config.Server.PublicCacheStaleTTL
	}
	return middleware.PublicResponseCache(infra.Redis, cfg)
}

// configureClientIP makes c.ClientIP() return the visitor address instead of
// a value the visitor can choose.
//
// Production path: Cloudflare -> cloudflared -> shared Caddy -> backend.
// Cloudflare appends the visitor address to any X-Forwarded-For the visitor
// sent, and Caddy (which trusts only cloudflared) appends cloudflared's
// address, giving "<spoofed...>, <visitor>, <cloudflared>". Gin walks the
// header from the right and returns the first address that is not a trusted
// proxy. With Gin's default of trusting every address, the walk ran off the
// left end and returned the spoofed value; trusting only private/container
// networks stops at the visitor's public address.
func configureClientIP(r *gin.Engine, trustedProxies []string) error {
	if err := r.SetTrustedProxies(trustedProxies); err != nil {
		return fmt.Errorf("invalid TRUSTED_PROXIES: %w", err)
	}
	// X-Real-IP is not consulted: Caddy does not set it for this site, so it
	// would only ever carry a visitor-supplied value.
	r.RemoteIPHeaders = []string{"X-Forwarded-For"}
	r.ForwardedByClientIP = true
	return nil
}
