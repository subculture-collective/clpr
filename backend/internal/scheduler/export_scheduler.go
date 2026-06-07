package scheduler

import (
	"context"
	"sync"
	"time"

	"git.subcult.tv/subculture-collective/clpr/internal/models"
	"git.subcult.tv/subculture-collective/clpr/internal/services"
	"git.subcult.tv/subculture-collective/clpr/pkg/utils"
)

const exportSchedulerName = "exports"

// ExportScheduler manages periodic processing of export requests
type ExportScheduler struct {
	exportService *services.ExportService
	exportRepo    services.ExportRepositoryInterface
	interval      time.Duration
	batchSize     int
	stopChan      chan struct{}
	stopOnce      sync.Once
}

// NewExportScheduler creates a new export scheduler
func NewExportScheduler(
	exportService *services.ExportService,
	exportRepo services.ExportRepositoryInterface,
	intervalMinutes int,
	batchSize int,
) *ExportScheduler {
	return &ExportScheduler{
		exportService: exportService,
		exportRepo:    exportRepo,
		interval:      time.Duration(intervalMinutes) * time.Minute,
		batchSize:     batchSize,
		stopChan:      make(chan struct{}),
	}
}

// Start begins the periodic export processing
func (s *ExportScheduler) Start(ctx context.Context) {
	logger := utils.GetLogger()
	logger.Info("Starting export scheduler", map[string]interface{}{
		"scheduler":  exportSchedulerName,
		"interval":   s.interval.String(),
		"batch_size": s.batchSize,
	})

	ticker := time.NewTicker(s.interval)
	defer ticker.Stop()

	// Run initial processing
	s.processExports(ctx)

	for {
		select {
		case <-ticker.C:
			s.processExports(ctx)
		case <-s.stopChan:
			logger.Info("Export scheduler stopped", map[string]interface{}{
				"scheduler": exportSchedulerName,
			})
			return
		case <-ctx.Done():
			logger.Info("Export scheduler stopped due to context cancellation", map[string]interface{}{
				"scheduler": exportSchedulerName,
			})
			return
		}
	}
}

// Stop stops the scheduler in a thread-safe manner
func (s *ExportScheduler) Stop() {
	s.stopOnce.Do(func() {
		close(s.stopChan)
	})
}

// processExports processes pending export requests
func (s *ExportScheduler) processExports(ctx context.Context) {
	logger := utils.GetLogger()
	logger.Debug("Processing pending export requests", map[string]interface{}{
		"scheduler": exportSchedulerName,
	})

	// Get pending export requests
	requests, err := s.exportRepo.GetPendingExportRequests(ctx, s.batchSize)
	if err != nil {
		logger.Error("Failed to get pending export requests", err, map[string]interface{}{
			"scheduler":  exportSchedulerName,
			"batch_size": s.batchSize,
		})
		return
	}

	if len(requests) == 0 {
		logger.Debug("No pending export requests to process", map[string]interface{}{
			"scheduler": exportSchedulerName,
		})
		return
	}

	logger.Info("Processing pending export requests", map[string]interface{}{
		"scheduler": exportSchedulerName,
		"count":     len(requests),
	})

	// Process each request concurrently using a worker pool
	// Cap workers at a reasonable maximum to avoid excessive goroutines
	const maxWorkers = 10
	numWorkers := s.batchSize
	if numWorkers > maxWorkers {
		numWorkers = maxWorkers
	}
	if numWorkers > len(requests) {
		numWorkers = len(requests)
	}
	requestCh := make(chan *models.ExportRequest, len(requests))
	var wg sync.WaitGroup

	// Start workers
	for i := 0; i < numWorkers; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for req := range requestCh {
				if err := s.exportService.ProcessExportRequest(ctx, req); err != nil {
					logger.Error("Failed to process export request", err, map[string]interface{}{
						"scheduler": exportSchedulerName,
						"export_id": req.ID.String(),
					})
					continue
				}
				logger.Info("Successfully processed export request", map[string]interface{}{
					"scheduler":    exportSchedulerName,
					"export_id":    req.ID.String(),
					"creator_name": req.CreatorName,
					"format":       req.Format,
				})
			}
		}()
	}

	// Send requests to workers
	for _, req := range requests {
		requestCh <- req
	}
	close(requestCh)

	// Wait for all workers to finish
	wg.Wait()

	// Clean up expired exports
	if err := s.exportService.CleanupExpiredExports(ctx); err != nil {
		logger.Error("Failed to cleanup expired exports", err, map[string]interface{}{
			"scheduler": exportSchedulerName,
		})
	}
}
