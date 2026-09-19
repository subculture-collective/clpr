package scheduler

import (
	"context"
	"errors"
	"time"

	"git.subcult.tv/subculture-collective/clpr/internal/repository"
	"git.subcult.tv/subculture-collective/clpr/pkg/metrics"
	"git.subcult.tv/subculture-collective/clpr/pkg/utils"
)

const tagPromotionJobName = "tag_promotion_detection"

type TagPromotionServiceInterface interface {
	CheckPromotionCandidates(context.Context) ([]string, error)
	PendingPromotionCount(context.Context) (int, error)
}

type TagPromotionScheduler struct {
	service      TagPromotionServiceInterface
	interval     time.Duration
	initialDelay time.Duration
}

func NewTagPromotionScheduler(service TagPromotionServiceInterface, interval, initialDelay time.Duration) *TagPromotionScheduler {
	if interval <= 0 {
		interval = 15 * time.Minute
	}
	return &TagPromotionScheduler{service: service, interval: interval, initialDelay: initialDelay}
}

func (s *TagPromotionScheduler) Start(ctx context.Context) {
	if s.initialDelay > 0 {
		timer := time.NewTimer(s.initialDelay)
		defer timer.Stop()
		select {
		case <-timer.C:
		case <-ctx.Done():
			return
		}
	}
	s.run(ctx)
	ticker := time.NewTicker(s.interval)
	defer ticker.Stop()
	for {
		select {
		case <-ticker.C:
			s.run(ctx)
		case <-ctx.Done():
			return
		}
	}
}

func (s *TagPromotionScheduler) Stop() {}

func (s *TagPromotionScheduler) run(ctx context.Context) {
	started := time.Now()
	queued, err := s.service.CheckPromotionCandidates(ctx)
	metrics.JobExecutionDuration.WithLabelValues(tagPromotionJobName).Observe(time.Since(started).Seconds())
	if errors.Is(err, repository.ErrSchedulerLockUnavailable) {
		metrics.JobExecutionTotal.WithLabelValues(tagPromotionJobName, "skipped").Inc()
		return
	}
	if err != nil {
		status := "failed"
		if len(queued) > 0 {
			status = "partial"
		}
		metrics.JobExecutionTotal.WithLabelValues(tagPromotionJobName, status).Inc()
		metrics.JobItemsProcessed.WithLabelValues(tagPromotionJobName, "success").Add(float64(len(queued)))
		metrics.JobItemsProcessed.WithLabelValues(tagPromotionJobName, "failed").Inc()
		if count, countErr := s.service.PendingPromotionCount(ctx); countErr == nil {
			metrics.JobQueueSize.WithLabelValues(tagPromotionJobName).Set(float64(count))
		}
		utils.Error("Tag promotion detection failed", err, map[string]interface{}{"job": tagPromotionJobName})
		return
	}
	metrics.JobExecutionTotal.WithLabelValues(tagPromotionJobName, "success").Inc()
	metrics.JobLastSuccessTimestamp.WithLabelValues(tagPromotionJobName).Set(float64(time.Now().Unix()))
	metrics.JobItemsProcessed.WithLabelValues(tagPromotionJobName, "success").Add(float64(len(queued)))
	if count, countErr := s.service.PendingPromotionCount(ctx); countErr == nil {
		metrics.JobQueueSize.WithLabelValues(tagPromotionJobName).Set(float64(count))
	}
}
