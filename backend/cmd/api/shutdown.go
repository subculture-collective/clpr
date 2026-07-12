package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"
)

func gracefulShutdown(srv *http.Server, svcs *Services, schedulers *SchedulerGroup, infra *Infrastructure) {
	// Wait for interrupt signal to gracefully shutdown the server
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit
	log.Println("Shutting down server...")

	// Shutdown WebSocket server first to close all connections
	svcs.WSServer.Shutdown()

	// Stop event tracker
	svcs.CancelEventTracker()

	// Stop schedulers
	if schedulers.ClipSync != nil {
		schedulers.ClipSync.Stop()
	}
	schedulers.Reputation.Stop()
	schedulers.HotScore.Stop()
	schedulers.TrendingScore.Stop()
	schedulers.WebhookRetry.Stop()
	schedulers.OutboundWebhook.Stop()
	schedulers.Export.Stop()
	schedulers.EmailMetrics.Stop()
	if schedulers.Embedding != nil {
		schedulers.Embedding.Stop()
	}
	if schedulers.LiveStatus != nil {
		schedulers.LiveStatus.Stop()
	}
	schedulers.PlaylistScript.Stop()

	// Close embedding service if running
	if svcs.Embedding != nil {
		svcs.Embedding.Close()
	}

	// Allow in-flight requests to drain while remaining bounded for orchestrators.
	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()

	if err := srv.Shutdown(ctx); err != nil {
		log.Printf("Server forced to shutdown: %v", err)
	}

	log.Println("Server exited")
}
