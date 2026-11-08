// Command countdown-media runs on the render host beside the database. Each
// cycle it fetches the clip media of selected countdown editions (#77) and
// renders fetched editions into a countdown video, a thumbnail and Shorts
// (#78), then uploads rendered editions to YouTube when the channel is
// authorized (#79; see "countdown-media youtube-auth"). Downloaded clips are deleted when their edition finishes rendering,
// and within 24 hours in any case (ADR-002). It never runs on the API host
// and never writes media to object storage.
package main

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"path/filepath"
	"strconv"
	"syscall"
	"time"

	"git.subcult.tv/subculture-collective/clpr/config"
	"git.subcult.tv/subculture-collective/clpr/internal/countdownrender"
	"git.subcult.tv/subculture-collective/clpr/internal/repository"
	"git.subcult.tv/subculture-collective/clpr/internal/services"
	"git.subcult.tv/subculture-collective/clpr/pkg/database"
	"git.subcult.tv/subculture-collective/clpr/pkg/utils"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

func envDuration(name string, fallback time.Duration, unit time.Duration) time.Duration {
	value := os.Getenv(name)
	if value == "" {
		return fallback
	}
	n, err := strconv.Atoi(value)
	if err != nil || n <= 0 {
		utils.Warn("Ignoring invalid duration setting", map[string]interface{}{"name": name})
		return fallback
	}
	return time.Duration(n) * unit
}

func envString(name, fallback string) string {
	if value := os.Getenv(name); value != "" {
		return value
	}
	return fallback
}

func main() {
	if len(os.Args) > 1 && os.Args[1] == "youtube-auth" {
		if err := youTubeAuth(os.Args[2:]); err != nil {
			fmt.Fprintln(os.Stderr, err)
			os.Exit(1)
		}
		return
	}
	if err := run(); err != nil {
		utils.Error("Countdown media worker stopped", err)
		os.Exit(1)
	}
}

func run() error {
	// Media files and their directories are readable by this user only.
	syscall.Umask(0o077)
	cfg, err := config.Load()
	if err != nil {
		return err
	}
	db, err := database.NewDB(&cfg.Database)
	if err != nil {
		return err
	}
	defer db.Close()

	dir := envString("COUNTDOWN_MEDIA_DIR", "/var/lib/clpr-countdown-media")
	if err := os.MkdirAll(dir, 0o700); err != nil {
		return err
	}
	renderDir := envString("COUNTDOWN_RENDER_DIR", "/var/lib/clpr-countdown-renders")
	if err := os.MkdirAll(renderDir, 0o700); err != nil {
		return err
	}
	template, err := countdownrender.LoadTemplate(os.Getenv("COUNTDOWN_RENDER_TEMPLATE"))
	if err != nil {
		return err
	}
	interval := envDuration("COUNTDOWN_MEDIA_INTERVAL_SECONDS", time.Minute, time.Second)
	retention := envDuration("COUNTDOWN_MEDIA_RETENTION_HOURS", 24*time.Hour, time.Hour)
	if retention > 24*time.Hour {
		retention = 24 * time.Hour // ADR-002
	}

	fetcher := &services.TwitchClipMediaFetcher{YTDLPPath: envString("YTDLP_PATH", "yt-dlp")}
	if cfg.Twitch.ClientID != "" {
		// Helix is used only for broadcasters who granted channel:manage:clips.
		fetcher.Auth = repository.NewTwitchAuthRepository(db.Pool)
		fetcher.Downloads = services.NewTwitchClipDownloadService(cfg.Twitch.ClientID, "", nil)
	}
	registry := prometheus.NewRegistry()
	store := repository.NewCountdownRepository(db.Pool)
	prober := &services.FFprobeMediaProber{FFprobePath: envString("FFPROBE_PATH", "ffprobe")}
	worker := &services.CountdownMediaWorker{
		Store:     store,
		Fetcher:   fetcher,
		Prober:    prober,
		Dir:       dir,
		Retention: retention,
		MaxSkips:  3,
		Metrics:   services.NewCountdownMediaMetrics(registry),
	}
	renderer := &countdownrender.Renderer{
		FFmpegPath:  envString("FFMPEG_PATH", "ffmpeg"),
		FFprobePath: envString("FFPROBE_PATH", "ffprobe"),
		Template:    template,
		// Intermediate segments hold clip footage, so they live under the
		// media directory and its 24-hour sweep.
		WorkDir: filepath.Join(dir, ".render-work"),
	}
	renders := &services.CountdownRenderWorker{
		Store:     store,
		Renderer:  renderer,
		Prober:    prober,
		MediaDir:  dir,
		RenderDir: renderDir,
		Retention: envDuration("COUNTDOWN_RENDER_RETENTION_DAYS", 7*24*time.Hour, 24*time.Hour),
		Metrics:   services.NewCountdownRenderMetrics(registry),
	}
	// YouTube publishing (#79) runs once the channel is authorized.
	youTube, err := youTubeClientFromEnv()
	if err != nil {
		return err
	}
	privacy, publishHour, err := youTubePrivacyFromEnv()
	if err != nil {
		return err
	}
	var publisher *services.CountdownPublisher
	if youTube != nil {
		publisher = &services.CountdownPublisher{
			Store:           store,
			YouTube:         youTube,
			RenderDir:       renderDir,
			BaseURL:         cfg.Server.BaseURL,
			Privacy:         privacy,
			PublishHourUTC:  publishHour,
			RankCardSeconds: template.RankCardSeconds,
			Metrics:         services.NewCountdownPublishMetrics(registry),
		}
	}
	lastRun := prometheus.NewGauge(prometheus.GaugeOpts{
		Name: "clpr_countdown_media_last_run_timestamp_seconds",
		Help: "Unix time the countdown media worker last finished a cycle without error.",
	})
	registry.MustRegister(lastRun)

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	if addr := envString("COUNTDOWN_MEDIA_METRICS_ADDR", ":9464"); addr != "off" {
		mux := http.NewServeMux()
		mux.Handle("/metrics", promhttp.HandlerFor(registry, promhttp.HandlerOpts{}))
		mux.HandleFunc("/healthz", func(w http.ResponseWriter, _ *http.Request) { w.WriteHeader(http.StatusOK) })
		server := &http.Server{Addr: addr, Handler: mux, ReadHeaderTimeout: 5 * time.Second}
		go func() {
			if err := server.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
				utils.Error("Countdown media metrics server failed", err)
			}
		}()
		defer server.Close()
	}

	utils.Info("Countdown media worker started", map[string]interface{}{
		"dir": dir, "render_dir": renderDir, "interval": interval.String(), "retention": retention.String(),
		"render_retention": renders.Retention.String(), "helix": fetcher.Downloads != nil,
		"youtube": publisher != nil, "youtube_privacy": privacy,
	})
	ticker := time.NewTicker(interval)
	defer ticker.Stop()
	for {
		err := errors.Join(worker.RunOnce(ctx), renders.RunOnce(ctx))
		if publisher != nil {
			err = errors.Join(err, publisher.RunOnce(ctx))
		}
		if err != nil && ctx.Err() == nil {
			utils.Error("Countdown media cycle failed", err)
		} else if err == nil {
			lastRun.SetToCurrentTime()
		}
		select {
		case <-ctx.Done():
			return nil
		case <-ticker.C:
		}
	}
}
