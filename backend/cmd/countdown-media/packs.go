package main

import (
	"net/http"
	"os"
	"path/filepath"
	"time"

	"git.subcult.tv/subculture-collective/clpr/config"
	"git.subcult.tv/subculture-collective/clpr/internal/countdownrender"
	"git.subcult.tv/subculture-collective/clpr/internal/repository"
	"git.subcult.tv/subculture-collective/clpr/internal/services"
	"github.com/prometheus/client_golang/prometheus"
)

// clipPacksFromEnv sets up weekly clip packs (#86) when CLIP_PACK_LINK_SECRET
// (or CLIP_PACK_LINK_SECRET_FILE) is set and Helix downloads are available.
// It returns nil without them.
func clipPacksFromEnv(cfg *config.Config, store *repository.CountdownRepository, fetcher *services.TwitchClipMediaFetcher,
	prober services.MediaProber, renderer *countdownrender.Renderer, captioner services.CountdownCaptioner,
	mediaDir string, registry prometheus.Registerer) (*services.ClipPackWorker, *http.Server, error) {
	secret, err := secretEnv("CLIP_PACK_LINK_SECRET")
	if err != nil || secret == "" || fetcher.Downloads == nil {
		return nil, nil, err
	}
	packDir := envString("COUNTDOWN_PACK_DIR", "/var/lib/clpr-clip-packs")
	if err := os.MkdirAll(packDir, 0o700); err != nil {
		return nil, nil, err
	}
	// A pack holds the broadcaster's own clips, fetched only through their
	// own authorization.
	helixOnly := *fetcher
	helixOnly.HelixOnly = true
	// The highlight reel is a countdown of the week, closed by its own outro.
	reel := *renderer
	reel.Template.OutroHeading = "This week's top clips"
	reel.Template.OutroFooter = "Clips by the broadcaster, ranked by views. Made with clpr.tv."
	worker := &services.ClipPackWorker{
		Store:     store,
		Fetcher:   &helixOnly,
		Prober:    prober,
		Renderer:  &reel,
		Captioner: captioner,
		// Under the media directory, so its 24-hour sweep catches leftovers.
		WorkDir: filepath.Join(mediaDir, ".clip-packs"),
		PackDir: packDir,
		Metrics: services.NewClipPackMetrics(registry),
	}
	server := &http.Server{
		Addr: envString("CLIP_PACK_ADDR", ":9465"),
		Handler: &services.ClipPackServer{Store: store, Dir: packDir,
			Links: services.ClipPackLinks{Secret: []byte(secret), BaseURL: cfg.ClipPacks.BaseURL}},
		ReadHeaderTimeout: 5 * time.Second,
	}
	return worker, server, nil
}
