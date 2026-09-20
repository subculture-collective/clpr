package main

import (
	"context"
	"sync"
	"time"

	"git.subcult.tv/subculture-collective/clpr/internal/scheduler"
)

// SchedulerGroup holds all background scheduler instances for graceful shutdown.
type SchedulerGroup struct {
	Engagement      *scheduler.EngagementScheduler
	ClipSync        *scheduler.ClipSyncScheduler
	Reputation      *scheduler.ReputationScheduler
	HotScore        *scheduler.HotScoreScheduler
	TrendingScore   *scheduler.TrendingScoreScheduler
	OutboundWebhook *scheduler.OutboundWebhookScheduler
	Embedding       *scheduler.EmbeddingScheduler // may be nil
	Export          *scheduler.ExportScheduler
	EmailMetrics    *scheduler.EmailMetricsScheduler
	LiveStatus      *scheduler.LiveStatusScheduler // may be nil
	PlaylistScript  *scheduler.PlaylistScriptScheduler
	AutoTag         *scheduler.AutoTagScheduler
	TagPromotion    *scheduler.TagPromotionScheduler
	cancel          context.CancelFunc
	wg              sync.WaitGroup
}

func (s *SchedulerGroup) launch(ctx context.Context, delay time.Duration, start func(context.Context)) {
	s.wg.Add(1)
	go func() {
		defer s.wg.Done()
		if delay > 0 {
			timer := time.NewTimer(delay)
			defer timer.Stop()
			select {
			case <-timer.C:
			case <-ctx.Done():
				return
			}
		}
		start(ctx)
	}()
}

func (s *SchedulerGroup) CancelAndWait(ctx context.Context) bool {
	s.cancel()
	done := make(chan struct{})
	go func() { s.wg.Wait(); close(done) }()
	select {
	case <-done:
		return true
	case <-ctx.Done():
		return false
	}
}

func startSchedulers(svcs *Services, repos *Repositories, infra *Infrastructure) *SchedulerGroup {
	cfg := infra.Config
	ctx, cancel := context.WithCancel(context.Background())
	sg := &SchedulerGroup{cancel: cancel}
	var engagementProvider scheduler.EngagementProvider
	if infra.TwitchClient != nil {
		engagementProvider = infra.TwitchClient
	}
	sg.Engagement = scheduler.NewEngagementScheduler(repos.Engagement, engagementProvider)
	sg.launch(ctx, 0, sg.Engagement.Start)

	// Start background scheduler if Twitch client is available
	if svcs.ClipSync != nil {
		// Start scheduler to run every 15 minutes
		sg.ClipSync = scheduler.NewClipSyncScheduler(svcs.ClipSync, 15)
		sg.launch(ctx, 0, sg.ClipSync.Start)
	}

	// Start reputation scheduler (runs every 6 hours)
	sg.Reputation = scheduler.NewReputationScheduler(svcs.Reputation, repos.User, 6)
	sg.launch(ctx, 0, sg.Reputation.Start)

	// Start hot score scheduler (runs every 5 minutes)
	sg.HotScore = scheduler.NewHotScoreScheduler(repos.Clip, cfg.Jobs.HotClipsRefreshIntervalMinutes)
	sg.launch(ctx, time.Duration(cfg.Jobs.HotScoreStartDelaySeconds)*time.Second, sg.HotScore.Start)

	// Start trending score scheduler (runs every 60 minutes)
	sg.TrendingScore = scheduler.NewTrendingScoreScheduler(repos.Clip, cfg.Jobs.TrendingScoreIntervalMinutes).SetBatchSize(cfg.Jobs.TrendingScoreBatchSize)
	sg.launch(ctx, time.Duration(cfg.Jobs.TrendingStartDelaySeconds)*time.Second, sg.TrendingScore.Start)

	// Start outbound webhook delivery scheduler (runs every 30 seconds, batch size 50)
	sg.OutboundWebhook = scheduler.NewOutboundWebhookScheduler(svcs.OutboundWebhook, 30*time.Second, 50)
	sg.launch(ctx, 0, sg.OutboundWebhook.Start)

	// Start embedding scheduler if embedding service is available (runs based on configured interval)
	if svcs.Embedding != nil {
		sg.Embedding = scheduler.NewEmbeddingScheduler(infra.DB, svcs.Embedding, cfg.Embedding.SchedulerIntervalMinutes, cfg.Embedding.Model)
		sg.launch(ctx, time.Duration(cfg.Jobs.EmbeddingStartDelaySeconds)*time.Second, sg.Embedding.Start)
	}

	// Start export scheduler (runs every 2 minutes, batch size 10)
	sg.Export = scheduler.NewExportScheduler(svcs.Export, repos.Export, 2, 10)
	sg.launch(ctx, 0, sg.Export.Start)

	// Start email metrics scheduler
	// - Calculate daily metrics every 24 hours
	// - Check alerts every 30 minutes
	// - Cleanup old logs every 7 days
	sg.EmailMetrics = scheduler.NewEmailMetricsScheduler(svcs.EmailMetrics, 24, 30, 7)
	sg.launch(ctx, 0, sg.EmailMetrics.Start)

	// Start live status scheduler (runs every 30 seconds if Twitch client is available)
	if svcs.LiveStatus != nil {
		sg.LiveStatus = scheduler.NewLiveStatusScheduler(svcs.LiveStatus, repos.Broadcaster, 30)
		sg.launch(ctx, 0, sg.LiveStatus.Start)
	}

	// Start playlist script scheduler (checks every 5 minutes for due scripts)
	sg.PlaylistScript = scheduler.NewPlaylistScriptScheduler(svcs.PlaylistScript, 5)
	sg.launch(ctx, time.Duration(cfg.Jobs.PlaylistStartDelaySeconds)*time.Second, sg.PlaylistScript.Start)

	// Start auto-tag scheduler (runs every 30 seconds to tag newly synced clips)
	if svcs.AutoTag != nil {
		sg.AutoTag = scheduler.NewAutoTagScheduler(
			svcs.AutoTag,
			svcs.Thumbnail,
			svcs.ClipTranscription,
			svcs.TopicClassification,
			repos.Clip,
			repos.Tag,
			cfg.Jobs.AutoTagIntervalSeconds,
			scheduler.WithVisionQueue(cfg.Vision.BatchSize, cfg.Vision.CreatedAfter),
		)
		sg.launch(ctx, time.Duration(cfg.Jobs.AutoTagStartDelaySeconds)*time.Second, sg.AutoTag.Start)
	}

	sg.TagPromotion = scheduler.NewTagPromotionScheduler(
		svcs.TagPromotion,
		time.Duration(cfg.Jobs.TagPromotionIntervalMinutes)*time.Minute,
		time.Duration(cfg.Jobs.TagPromotionStartDelaySeconds)*time.Second,
	)
	sg.launch(ctx, 0, sg.TagPromotion.Start)

	return sg
}
