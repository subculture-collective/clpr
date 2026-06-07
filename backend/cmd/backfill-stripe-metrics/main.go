package main

import (
	"context"
	"flag"
	"log"
	"time"

	"github.com/stripe/stripe-go/v81"
	"github.com/stripe/stripe-go/v81/invoice"
	stripeSub "github.com/stripe/stripe-go/v81/subscription"
	"git.subcult.tv/subculture-collective/clpr/config"
	"git.subcult.tv/subculture-collective/clpr/pkg/database"
)

// BackfillStats tracks statistics for the backfill operation
type BackfillStats struct {
	TotalSubscriptions     int
	ProcessedSubscriptions int
	FailedSubscriptions    int
	TotalInvoices          int
	ProcessedInvoices      int
	FailedInvoices         int
	StartTime              time.Time
	EndTime                time.Time
	LastError              error
}

func main() {
	dryRun := flag.Bool("dry-run", false, "Dry run mode - don't save to database")
	limit := flag.Int("limit", 100, "Maximum number of records to sync per type")
	flag.Parse()

	log.Println("Starting Stripe metrics backfill job...")
	log.Printf("Configuration: dry_run=%t, limit=%d", *dryRun, *limit)

	// Load configuration
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("Failed to load configuration: %v", err)
	}

	if cfg.Stripe.SecretKey == "" {
		log.Fatalf("STRIPE_SECRET_KEY is not set. Please set it in your environment or .env file")
	}

	// Initialize Stripe
	stripe.Key = cfg.Stripe.SecretKey
	log.Println("Stripe client initialized")

	// Initialize database connection
	db, err := database.NewDB(&cfg.Database)
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}
	defer db.Close()
	log.Println("Database connection established")

	// Run backfill
	ctx := context.Background()
	stats, err := backfillStripeMetrics(ctx, db, *dryRun, *limit)
	if err != nil {
		log.Fatalf("Backfill failed: %v", err)
	}

	// Print summary
	duration := stats.EndTime.Sub(stats.StartTime)
	log.Println("\n=== Backfill Summary ===")
	log.Printf("Subscriptions processed: %d/%d (failed: %d)", stats.ProcessedSubscriptions, stats.TotalSubscriptions, stats.FailedSubscriptions)
	log.Printf("Invoices processed: %d/%d (failed: %d)", stats.ProcessedInvoices, stats.TotalInvoices, stats.FailedInvoices)
	log.Printf("Duration: %v", duration)

	if stats.LastError != nil {
		log.Printf("Last error: %v", stats.LastError)
	}

	log.Println("\n✓ Backfill completed successfully!")
}

func backfillStripeMetrics(
	ctx context.Context,
	db *database.DB,
	dryRun bool,
	limit int,
) (*BackfillStats, error) {
	stats := &BackfillStats{
		StartTime: time.Now(),
	}

	// Sync subscriptions from Stripe
	log.Println("\n--- Syncing Subscriptions ---")
	if err := syncSubscriptions(ctx, db, stats, dryRun, limit); err != nil {
		log.Printf("WARNING: Error syncing subscriptions: %v", err)
		stats.LastError = err
	}

	// Sync paid invoices from Stripe
	log.Println("\n--- Syncing Invoices ---")
	if err := syncInvoices(ctx, db, stats, dryRun, limit); err != nil {
		log.Printf("WARNING: Error syncing invoices: %v", err)
		stats.LastError = err
	}

	stats.EndTime = time.Now()
	return stats, nil
}

func syncSubscriptions(
	ctx context.Context,
	db *database.DB,
	stats *BackfillStats,
	dryRun bool,
	limit int,
) error {
	// List subscriptions from Stripe
	params := &stripe.SubscriptionListParams{
		ListParams: stripe.ListParams{
			Limit: stripe.Int64(int64(limit)),
		},
	}
	params.AddExpand("data.customer")

	i := stripeSub.List(params)
	for i.Next() {
		stats.TotalSubscriptions++
		sub := i.Subscription()

		log.Printf("Processing subscription: %s (status: %s, customer: %s)",
			sub.ID, sub.Status, sub.Customer.ID)

		if dryRun {
			stats.ProcessedSubscriptions++
			continue
		}

		// Check if we have this customer in our database
		var userID string
		query := `SELECT user_id::text FROM subscriptions WHERE stripe_customer_id = $1`
		err := db.Pool.QueryRow(ctx, query, sub.Customer.ID).Scan(&userID)
		if err != nil {
			log.Printf("No local user found for customer %s, skipping", sub.Customer.ID)
			stats.FailedSubscriptions++
			continue
		}

		// Update subscription in our database
		updateQuery := `
			UPDATE subscriptions
			SET stripe_subscription_id = $1,
			    stripe_price_id = $2,
			    status = $3,
			    tier = $4,
			    current_period_start = $5,
			    current_period_end = $6,
			    cancel_at_period_end = $7,
			    canceled_at = $8,
			    trial_start = $9,
			    trial_end = $10,
			    updated_at = NOW()
			WHERE stripe_customer_id = $11
		`

		var priceID string
		if len(sub.Items.Data) > 0 && sub.Items.Data[0].Price != nil {
			priceID = sub.Items.Data[0].Price.ID
		}

		tier := "free"
		if sub.Status == stripe.SubscriptionStatusActive || sub.Status == stripe.SubscriptionStatusTrialing || sub.Status == stripe.SubscriptionStatusCanceled {
			tier = "pro" // Preserve pro tier for canceled subs for historical revenue data
		}

		var canceledAt *time.Time
		if sub.CanceledAt > 0 {
			t := time.Unix(sub.CanceledAt, 0)
			canceledAt = &t
		}

		var trialStart, trialEnd *time.Time
		if sub.TrialStart > 0 {
			t := time.Unix(sub.TrialStart, 0)
			trialStart = &t
		}
		if sub.TrialEnd > 0 {
			t := time.Unix(sub.TrialEnd, 0)
			trialEnd = &t
		}

		periodStart := time.Unix(sub.CurrentPeriodStart, 0)
		periodEnd := time.Unix(sub.CurrentPeriodEnd, 0)

		_, err = db.Pool.Exec(ctx, updateQuery,
			sub.ID,
			priceID,
			string(sub.Status),
			tier,
			periodStart,
			periodEnd,
			sub.CancelAtPeriodEnd,
			canceledAt,
			trialStart,
			trialEnd,
			sub.Customer.ID,
		)
		if err != nil {
			log.Printf("Failed to update subscription %s: %v", sub.ID, err)
			stats.FailedSubscriptions++
			stats.LastError = err
			continue
		}

		stats.ProcessedSubscriptions++
	}

	if err := i.Err(); err != nil {
		return err
	}

	return nil
}

func syncInvoices(
	ctx context.Context,
	db *database.DB,
	stats *BackfillStats,
	dryRun bool,
	limit int,
) error {
	// List paid invoices from Stripe
	params := &stripe.InvoiceListParams{
		ListParams: stripe.ListParams{
			Limit: stripe.Int64(int64(limit)),
		},
		Status: stripe.String("paid"),
	}
	params.AddExpand("data.subscription")
	params.AddExpand("data.customer")

	i := invoice.List(params)
	for i.Next() {
		stats.TotalInvoices++
		inv := i.Invoice()

		if inv.Subscription == nil {
			log.Printf("Invoice %s has no subscription, skipping", inv.ID)
			continue
		}

		log.Printf("Processing invoice: %s (amount: %d, customer: %s)",
			inv.ID, inv.AmountPaid, inv.Customer.ID)

		if dryRun {
			stats.ProcessedInvoices++
			continue
		}

		// Check if we have this subscription in our database
		var subscriptionID string
		query := `SELECT id::text FROM subscriptions WHERE stripe_subscription_id = $1`
		err := db.Pool.QueryRow(ctx, query, inv.Subscription.ID).Scan(&subscriptionID)
		if err != nil {
			log.Printf("No local subscription found for Stripe subscription %s, skipping", inv.Subscription.ID)
			stats.FailedInvoices++
			continue
		}

		// Check if we already have this event logged
		eventQuery := `SELECT id FROM subscription_events WHERE stripe_event_id = $1`
		var existingID string
		err = db.Pool.QueryRow(ctx, eventQuery, inv.ID).Scan(&existingID)
		if err == nil {
			log.Printf("Invoice %s already logged, skipping", inv.ID)
			stats.ProcessedInvoices++
			continue
		}

		// Log the invoice as a subscription event
		insertQuery := `
			INSERT INTO subscription_events (subscription_id, event_type, stripe_event_id, payload)
			VALUES ($1::uuid, $2, $3, $4::jsonb)
		`

		// Create a simple payload
		payload := map[string]interface{}{
			"invoice_id":      inv.ID,
			"amount_paid":     inv.AmountPaid,
			"currency":        inv.Currency,
			"customer_id":     inv.Customer.ID,
			"subscription_id": inv.Subscription.ID,
			"created":         time.Unix(inv.Created, 0).Format(time.RFC3339),
		}

		_, err = db.Pool.Exec(ctx, insertQuery,
			subscriptionID,
			"invoice_paid",
			inv.ID,
			payload,
		)
		if err != nil {
			log.Printf("Failed to log invoice %s: %v", inv.ID, err)
			stats.FailedInvoices++
			stats.LastError = err
			continue
		}

		stats.ProcessedInvoices++
	}

	if err := i.Err(); err != nil {
		return err
	}

	return nil
}
