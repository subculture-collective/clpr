package main

import (
	"context"
	"time"

	"git.subcult.tv/subculture-collective/clpr/internal/scheduler"
)

// SchedulerGroup holds all background scheduler instances for graceful shutdown.
type SchedulerGroup struct {
	ClipSync        *scheduler.ClipSyncScheduler
	Reputation      *scheduler.ReputationScheduler
	HotScore        *scheduler.HotScoreScheduler
	TrendingScore   *scheduler.TrendingScoreScheduler
	WebhookRetry    *scheduler.WebhookRetryScheduler
	OutboundWebhook *scheduler.OutboundWebhookScheduler
	Embedding       *scheduler.EmbeddingScheduler // may be nil
	Export          *scheduler.ExportScheduler
	EmailMetrics    *scheduler.EmailMetricsScheduler
	LiveStatus      *scheduler.LiveStatusScheduler       // may be nil
	PlaylistScript  *scheduler.PlaylistScriptScheduler
}

func startSchedulers(svcs *Services, repos *Repositories, infra *Infrastructure) *SchedulerGroup {
	cfg := infra.Config
	sg := &SchedulerGroup{}

	// Start background scheduler if Twitch client is available
	if svcs.ClipSync != nil {
		// Start scheduler to run every 15 minutes
		sg.ClipSync = scheduler.NewClipSyncScheduler(svcs.ClipSync, 15)
		go sg.ClipSync.Start(context.Background())
	}

	// Start reputation scheduler (runs every 6 hours)
	sg.Reputation = scheduler.NewReputationScheduler(svcs.Reputation, repos.User, 6)
	go sg.Reputation.Start(context.Background())

	// Start hot score scheduler (runs every 5 minutes)
	sg.HotScore = scheduler.NewHotScoreScheduler(repos.Clip, cfg.Jobs.HotClipsRefreshIntervalMinutes)
	go sg.HotScore.Start(context.Background())

	// Start trending score scheduler (runs every 60 minutes)
	sg.TrendingScore = scheduler.NewTrendingScoreScheduler(repos.Clip, 60)
	go sg.TrendingScore.Start(context.Background())

	// Start webhook retry scheduler (runs every 1 minute)
	sg.WebhookRetry = scheduler.NewWebhookRetryScheduler(svcs.WebhookRetry, cfg.Jobs.WebhookRetryIntervalMinutes, cfg.Jobs.WebhookRetryBatchSize)
	go sg.WebhookRetry.Start(context.Background())

	// Start outbound webhook delivery scheduler (runs every 30 seconds, batch size 50)
	sg.OutboundWebhook = scheduler.NewOutboundWebhookScheduler(svcs.OutboundWebhook, 30*time.Second, 50)
	go sg.OutboundWebhook.Start(context.Background())

	// Start embedding scheduler if embedding service is available (runs based on configured interval)
	if svcs.Embedding != nil {
		sg.Embedding = scheduler.NewEmbeddingScheduler(infra.DB, svcs.Embedding, cfg.Embedding.SchedulerIntervalMinutes, cfg.Embedding.Model)
		go sg.Embedding.Start(context.Background())
	}

	// Start export scheduler (runs every 2 minutes, batch size 10)
	sg.Export = scheduler.NewExportScheduler(svcs.Export, repos.Export, 2, 10)
	go sg.Export.Start(context.Background())

	// Start email metrics scheduler
	// - Calculate daily metrics every 24 hours
	// - Check alerts every 30 minutes
	// - Cleanup old logs every 7 days
	sg.EmailMetrics = scheduler.NewEmailMetricsScheduler(svcs.EmailMetrics, 24, 30, 7)
	go sg.EmailMetrics.Start(context.Background())

	// Start live status scheduler (runs every 30 seconds if Twitch client is available)
	if svcs.LiveStatus != nil {
		sg.LiveStatus = scheduler.NewLiveStatusScheduler(svcs.LiveStatus, repos.Broadcaster, 30)
		go sg.LiveStatus.Start(context.Background())
	}

	// Start playlist script scheduler (checks every 5 minutes for due scripts)
	sg.PlaylistScript = scheduler.NewPlaylistScriptScheduler(svcs.PlaylistScript, 5)
	go sg.PlaylistScript.Start(context.Background())

	return sg
}
