package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"time"

	"git.subcult.tv/subculture-collective/clpr/config"
	"git.subcult.tv/subculture-collective/clpr/internal/models"
	"git.subcult.tv/subculture-collective/clpr/internal/repository"
	"git.subcult.tv/subculture-collective/clpr/pkg/database"
)

// AuditStats tracks statistics for the audit operation
type AuditStats struct {
	TotalUsersAudited int
	PassedAudits      int
	FlaggedAudits     int
	RevokedAudits     int
	StartTime         time.Time
	EndTime           time.Time
	LastError         error
}

func main() {
	dryRun := flag.Bool("dry-run", false, "Dry run mode - don't save audit results to database")
	limit := flag.Int("limit", 100, "Maximum number of users to audit")
	auditPeriodDays := flag.Int("audit-period", 90, "Audit users who haven't been audited in this many days")
	flag.Parse()

	log.Println("Starting creator verification audit job...")
	log.Printf("Configuration: dry_run=%t, limit=%d, audit_period=%d days", *dryRun, *limit, *auditPeriodDays)

	// Load configuration
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("Failed to load configuration: %v", err)
	}

	// Initialize database connection
	db, err := database.NewDB(&cfg.Database)
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}
	defer db.Close()
	log.Println("Database connection established")

	// Initialize repository
	verificationRepo := repository.NewVerificationRepository(db.Pool)

	// Run audit
	ctx := context.Background()
	stats, err := runVerificationAudit(ctx, verificationRepo, *dryRun, *limit, *auditPeriodDays)
	if err != nil {
		log.Fatalf("Audit failed: %v", err)
	}

	// Print summary
	duration := stats.EndTime.Sub(stats.StartTime)
	log.Println("\n=== Audit Summary ===")
	log.Printf("Users audited: %d", stats.TotalUsersAudited)
	log.Printf("  - Passed: %d", stats.PassedAudits)
	log.Printf("  - Flagged: %d", stats.FlaggedAudits)
	log.Printf("  - Revoked: %d", stats.RevokedAudits)
	log.Printf("Duration: %v", duration)

	if stats.LastError != nil {
		log.Printf("Last error: %v", stats.LastError)
	}

	log.Println("\n✓ Audit completed successfully!")
}

func runVerificationAudit(
	ctx context.Context,
	repo *repository.VerificationRepository,
	dryRun bool,
	limit int,
	auditPeriodDays int,
) (*AuditStats, error) {
	stats := &AuditStats{
		StartTime: time.Now(),
	}

	log.Printf("\n--- Retrieving verified users for audit ---")

	// Get verified users that need auditing
	users, err := repo.GetVerifiedUsersForAudit(ctx, auditPeriodDays, limit)
	if err != nil {
		return stats, fmt.Errorf("failed to retrieve users for audit: %w", err)
	}

	log.Printf("Found %d verified users requiring audit", len(users))

	if len(users) == 0 {
		log.Println("No users require auditing at this time")
		stats.EndTime = time.Now()
		return stats, nil
	}

	log.Println("\n--- Auditing verified users ---")

	for i, user := range users {
		log.Printf("\n[%d/%d] Auditing user: %s (ID: %s)", i+1, len(users), user.Username, user.ID)

		// Perform audit checks
		auditResult := performUserAudit(ctx, repo, user)

		stats.TotalUsersAudited++

		switch auditResult.Status {
		case models.AuditStatusPassed:
			stats.PassedAudits++
			log.Printf("  ✓ Audit passed")
		case models.AuditStatusFlagged:
			stats.FlaggedAudits++
			log.Printf("  ⚠ Audit flagged for review: %s", *auditResult.Notes)
		case models.AuditStatusRevoked:
			stats.RevokedAudits++
			log.Printf("  ✗ Verification revoked: %s", *auditResult.Notes)

			// Actually revoke the user's verification status
			if !dryRun {
				if err := repo.RevokeUserVerification(ctx, user.ID); err != nil {
					log.Printf("  ERROR: Failed to revoke verification: %v", err)
					stats.LastError = err
					// Continue to save audit log even if revocation fails
				} else {
					log.Printf("  ✓ Verification status revoked in database")
				}
			} else {
				log.Printf("  [DRY RUN] Would revoke verification status")
			}
		}

		// Save audit log (unless dry run)
		if !dryRun {
			err := repo.CreateAuditLog(ctx, auditResult)
			if err != nil {
				log.Printf("  ERROR: Failed to save audit log: %v", err)
				stats.LastError = err
				continue
			}
			log.Printf("  Audit log saved (ID: %s)", auditResult.ID)
		} else {
			log.Printf("  [DRY RUN] Would save audit log: status=%s", auditResult.Status)
		}
	}

	stats.EndTime = time.Now()
	return stats, nil
}

// performUserAudit performs audit checks on a verified user
func performUserAudit(ctx context.Context, repo *repository.VerificationRepository, user *models.User) *models.VerificationAuditLog {
	findings := make(map[string]interface{})
	status := models.AuditStatusPassed
	actionTaken := models.AuditActionNone
	var notes string

	// Check 1: User account status
	if user.IsBanned {
		status = models.AuditStatusRevoked
		actionTaken = models.AuditActionVerificationRevoked
		notes = "User account is banned"
		findings["banned"] = true
	}

	// Check 2: Trust score (if below threshold, flag for review)
	if user.TrustScore < 50 && status != models.AuditStatusRevoked {
		status = models.AuditStatusFlagged
		actionTaken = models.AuditActionFurtherReviewRequired
		notes = fmt.Sprintf("Low trust score: %d (threshold: 50)", user.TrustScore)
		findings["trust_score"] = user.TrustScore
		findings["trust_score_threshold"] = 50
	}

	// Check 3: Karma points (if negative, flag for review)
	if user.KarmaPoints < 0 && status == models.AuditStatusPassed {
		status = models.AuditStatusFlagged
		actionTaken = models.AuditActionFurtherReviewRequired
		notes = fmt.Sprintf("Negative karma points: %d", user.KarmaPoints)
		findings["karma_points"] = user.KarmaPoints
	}

	// Check 4: DMCA strikes or termination
	if user.DMCATerminated {
		status = models.AuditStatusRevoked
		actionTaken = models.AuditActionVerificationRevoked
		notes = "User has DMCA termination"
		findings["dmca_terminated"] = true
	} else if user.DMCAStrikesCount >= 2 && status != models.AuditStatusRevoked {
		status = models.AuditStatusFlagged
		actionTaken = models.AuditActionFurtherReviewRequired
		notes = fmt.Sprintf("Multiple DMCA strikes: %d", user.DMCAStrikesCount)
		findings["dmca_strikes"] = user.DMCAStrikesCount
	}

	// Check 5: Verification age (recently verified users get a pass on some checks)
	if user.VerifiedAt != nil {
		daysSinceVerification := time.Since(*user.VerifiedAt).Hours() / 24
		findings["days_since_verification"] = int(daysSinceVerification)
	}

	// Record audit metadata
	findings["user_id"] = user.ID.String()
	findings["username"] = user.Username
	findings["audit_timestamp"] = time.Now().Format(time.RFC3339)

	if status == models.AuditStatusPassed {
		notes = "All audit checks passed"
	}

	return &models.VerificationAuditLog{
		UserID:      user.ID,
		AuditType:   models.AuditTypePeriodicCheck,
		Status:      status,
		Findings:    findings,
		Notes:       &notes,
		AuditedBy:   nil, // Automated audit
		ActionTaken: &actionTaken,
	}
}
