//go:build integration

package repository

import (
	"context"
	"sync"
	"testing"
	"time"

	"git.subcult.tv/subculture-collective/clpr/internal/testutil"
	"github.com/google/uuid"
	"github.com/stretchr/testify/require"
)

func TestWebhookRetryClaimsAreExclusiveAcrossWorkers(t *testing.T) {
	pool := testutil.SetupTestDB(t)
	t.Cleanup(pool.Close)
	ctx := context.Background()
	eventID := "evt_claim_contract_" + uuid.NewString()
	_, err := pool.Exec(ctx, `
		INSERT INTO webhook_retry_queue
			(stripe_event_id, event_type, payload, retry_count, max_retries, next_retry_at)
		VALUES ($1, 'invoice.payment_failed', '{}', 0, 3, $2)
	`, eventID, time.Now().Add(-time.Minute))
	require.NoError(t, err)
	t.Cleanup(func() {
		_, _ = pool.Exec(context.Background(), "DELETE FROM webhook_retry_queue WHERE stripe_event_id = $1", eventID)
	})

	start := make(chan struct{})
	type claimResult struct {
		count int
		err   error
	}
	results := make(chan claimResult, 2)
	var workers sync.WaitGroup
	for range 2 {
		workers.Add(1)
		go func() {
			defer workers.Done()
			<-start
			items, claimErr := NewWebhookRepository(pool).GetPendingRetries(ctx, 1)
			results <- claimResult{count: len(items), err: claimErr}
		}()
	}
	close(start)
	workers.Wait()
	close(results)

	totalClaimed := 0
	for result := range results {
		require.NoError(t, result.err)
		totalClaimed += result.count
	}
	require.Equal(t, 1, totalClaimed, "a retry must be leased by exactly one worker")
}
