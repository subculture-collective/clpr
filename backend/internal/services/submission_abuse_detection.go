package services

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	redispkg "git.subcult.tv/subculture-collective/clpr/pkg/redis"
	"git.subcult.tv/subculture-collective/clpr/pkg/utils"
)

// SubmissionAbuseDetector detects and prevents abusive submission patterns
type SubmissionAbuseDetector struct {
	redisClient *redispkg.Client
}

// NewSubmissionAbuseDetector creates a new abuse detector
func NewSubmissionAbuseDetector(redisClient *redispkg.Client) *SubmissionAbuseDetector {
	return &SubmissionAbuseDetector{
		redisClient: redisClient,
	}
}

const (
	// Velocity limits - detect rapid submissions
	velocityWindow    = 5 * time.Minute
	velocityThreshold = 3 // max 3 submissions in 5 minutes
	velocityCooldown  = 30 * time.Minute

	// Duplicate detection
	duplicateWindow    = 1 * time.Hour
	duplicateThreshold = 3 // max 3 duplicate attempts in 1 hour
	duplicateCooldown  = 1 * time.Hour

	// Same IP, different users
	ipSharedWindow    = 1 * time.Hour
	ipSharedThreshold = 5 // max 5 different users from same IP

	// Burst detection
	burstWindow    = 1 * time.Minute
	burstThreshold = 2 // max 2 submissions in 1 minute
	burstCooldown  = 15 * time.Minute
)

// AbuseCheckResult contains the result of an abuse check
type AbuseCheckResult struct {
	Allowed       bool
	Reason        string
	CooldownUntil *time.Time
	Severity      string // "warning", "throttle", "block"
}

// CheckSubmissionAbuse performs comprehensive abuse detection
func (d *SubmissionAbuseDetector) CheckSubmissionAbuse(ctx context.Context, userID uuid.UUID, ip string, deviceFingerprint string) (*AbuseCheckResult, error) {
	// Check if user is in cooldown
	if blocked, until := d.checkCooldown(ctx, userID); blocked {
		return &AbuseCheckResult{
			Allowed:       false,
			Reason:        "You are temporarily blocked from submitting due to suspicious activity. Please try again later.",
			CooldownUntil: &until,
			Severity:      "block",
		}, nil
	}

	// Check burst detection (most immediate threat)
	if violated, err := d.checkBurstViolation(ctx, userID); err != nil {
		utils.Warn("Error checking burst violation", map[string]interface{}{"error": err})
	} else if violated {
		cooldownUntil := time.Now().Add(burstCooldown)
		_ = d.setCooldown(ctx, userID, burstCooldown, "burst")
		d.logAbuseAttempt(ctx, userID, ip, "burst_detection", fmt.Sprintf("Exceeded %d submissions in %v", burstThreshold, burstWindow))

		return &AbuseCheckResult{
			Allowed:       false,
			Reason:        "You are submitting too quickly. Please slow down and try again in a few minutes.",
			CooldownUntil: &cooldownUntil,
			Severity:      "throttle",
		}, nil
	}

	// Check velocity (sustained rapid submissions)
	if violated, err := d.checkVelocityViolation(ctx, userID); err != nil {
		utils.Warn("Error checking velocity violation", map[string]interface{}{"error": err})
	} else if violated {
		cooldownUntil := time.Now().Add(velocityCooldown)
		_ = d.setCooldown(ctx, userID, velocityCooldown, "velocity")
		d.logAbuseAttempt(ctx, userID, ip, "velocity_exceeded", fmt.Sprintf("Exceeded %d submissions in %v", velocityThreshold, velocityWindow))

		return &AbuseCheckResult{
			Allowed:       false,
			Reason:        "You have been submitting clips too rapidly. Please wait before submitting more clips.",
			CooldownUntil: &cooldownUntil,
			Severity:      "throttle",
		}, nil
	}

	// Check IP-based patterns (multiple users from same IP)
	if warning, err := d.checkIPSharing(ctx, userID, ip); err != nil {
		utils.Warn("Error checking IP sharing", map[string]interface{}{"error": err})
	} else if warning {
		// Log but don't block (could be legitimate shared network)
		d.logAbuseAttempt(ctx, userID, ip, "ip_sharing_detected", fmt.Sprintf("Multiple users from IP %s", ip))

		return &AbuseCheckResult{
			Allowed:       true, // Allow but flag
			Reason:        "",
			CooldownUntil: nil,
			Severity:      "warning",
		}, nil
	}

	// Track this submission for future checks
	_ = d.trackSubmission(ctx, userID, ip, deviceFingerprint)

	return &AbuseCheckResult{
		Allowed:       true,
		Reason:        "",
		CooldownUntil: nil,
		Severity:      "",
	}, nil
}

// TrackDuplicateAttempt tracks when a user tries to submit a duplicate clip
func (d *SubmissionAbuseDetector) TrackDuplicateAttempt(ctx context.Context, userID uuid.UUID, ip string, clipID string) error {
	key := fmt.Sprintf("submission:duplicate:%s", userID.String())
	count, err := d.redisClient.Increment(ctx, key)
	if err != nil {
		return err
	}

	if count == 1 {
		_ = d.redisClient.Expire(ctx, key, duplicateWindow)
	}

	// Check if threshold exceeded
	if count >= int64(duplicateThreshold) {
		_ = d.setCooldown(ctx, userID, duplicateCooldown, "duplicate")
		d.logAbuseAttempt(ctx, userID, ip, "duplicate_attempts", fmt.Sprintf("Attempted to submit duplicate clip %s %d times", clipID, count))
	}

	return nil
}

// checkBurstViolation checks if user is submitting too fast
func (d *SubmissionAbuseDetector) checkBurstViolation(ctx context.Context, userID uuid.UUID) (bool, error) {
	key := fmt.Sprintf("submission:burst:%s", userID.String())
	count, err := d.redisClient.Get(ctx, key)
	if err != nil {
		return false, nil // Key doesn't exist, no violation
	}

	var currentCount int64
	if _, err := fmt.Sscanf(count, "%d", &currentCount); err != nil {
		return false, err
	}

	return currentCount >= int64(burstThreshold), nil
}

// checkVelocityViolation checks if user has too many submissions in the window
func (d *SubmissionAbuseDetector) checkVelocityViolation(ctx context.Context, userID uuid.UUID) (bool, error) {
	key := fmt.Sprintf("submission:velocity:%s", userID.String())
	count, err := d.redisClient.Get(ctx, key)
	if err != nil {
		return false, nil // Key doesn't exist, no violation
	}

	var currentCount int64
	if _, err := fmt.Sscanf(count, "%d", &currentCount); err != nil {
		return false, err
	}

	return currentCount >= int64(velocityThreshold), nil
}

// checkIPSharing checks if too many users are submitting from the same IP
func (d *SubmissionAbuseDetector) checkIPSharing(ctx context.Context, userID uuid.UUID, ip string) (bool, error) {
	key := fmt.Sprintf("submission:ip:%s", ip)

	// Add user to set
	if err := d.redisClient.SetAdd(ctx, key, userID.String()); err != nil {
		return false, err
	}

	// Always set expiration to ensure it's maintained (Redis preserves TTL if key exists)
	_ = d.redisClient.Expire(ctx, key, ipSharedWindow)

	// Check count
	count, err := d.redisClient.SetCard(ctx, key)
	if err != nil {
		return false, err
	}

	return count >= int64(ipSharedThreshold), nil
}

// checkCooldown checks if user is currently in cooldown
func (d *SubmissionAbuseDetector) checkCooldown(ctx context.Context, userID uuid.UUID) (bool, time.Time) {
	key := fmt.Sprintf("submission:cooldown:%s", userID.String())
	ttl, err := d.redisClient.TTL(ctx, key)
	if err != nil || ttl <= 0 {
		return false, time.Time{}
	}

	return true, time.Now().Add(time.Duration(ttl) * time.Second)
}

// setCooldown sets a cooldown period for a user
func (d *SubmissionAbuseDetector) setCooldown(ctx context.Context, userID uuid.UUID, duration time.Duration, reason string) error {
	key := fmt.Sprintf("submission:cooldown:%s", userID.String())
	return d.redisClient.Set(ctx, key, reason, duration)
}

// trackSubmission tracks a successful submission for abuse detection
func (d *SubmissionAbuseDetector) trackSubmission(ctx context.Context, userID uuid.UUID, ip string, deviceFingerprint string) error {
	// Track burst
	burstKey := fmt.Sprintf("submission:burst:%s", userID.String())
	count, err := d.redisClient.Increment(ctx, burstKey)
	if err != nil {
		return err
	}
	if count == 1 {
		_ = d.redisClient.Expire(ctx, burstKey, burstWindow)
	}

	// Track velocity
	velocityKey := fmt.Sprintf("submission:velocity:%s", userID.String())
	count, err = d.redisClient.Increment(ctx, velocityKey)
	if err != nil {
		return err
	}
	if count == 1 {
		_ = d.redisClient.Expire(ctx, velocityKey, velocityWindow)
	}

	// Track IP
	ipKey := fmt.Sprintf("submission:ip:%s", ip)
	_ = d.redisClient.SetAdd(ctx, ipKey, userID.String())
	_ = d.redisClient.Expire(ctx, ipKey, ipSharedWindow)

	return nil
}

// logAbuseAttempt logs an abuse attempt for monitoring
func (d *SubmissionAbuseDetector) logAbuseAttempt(ctx context.Context, userID uuid.UUID, ip string, abuseType string, details string) {
	// Log to application logs
	utils.Warn("Abuse detection triggered", map[string]interface{}{
		"type":    abuseType,
		"user_id": userID,
		"ip":      ip,
		"details": details,
	})

	// Store in Redis for admin dashboard
	key := fmt.Sprintf("abuse:log:%s:%d", abuseType, time.Now().Unix())
	data := fmt.Sprintf("user_id=%s,ip=%s,details=%s", userID, ip, details)
	_ = d.redisClient.Set(ctx, key, data, 7*24*time.Hour) // Keep for 7 days
}

// GetAbuseStats returns abuse statistics for a user (admin function)
func (d *SubmissionAbuseDetector) GetAbuseStats(ctx context.Context, userID uuid.UUID) (map[string]interface{}, error) {
	stats := make(map[string]interface{})

	// Check cooldown status
	inCooldown, cooldownUntil := d.checkCooldown(ctx, userID)
	stats["in_cooldown"] = inCooldown
	if inCooldown {
		stats["cooldown_until"] = cooldownUntil
	}

	// Get burst count
	burstKey := fmt.Sprintf("submission:burst:%s", userID.String())
	if count, err := d.redisClient.Get(ctx, burstKey); err == nil {
		stats["burst_count"] = count
	}

	// Get velocity count
	velocityKey := fmt.Sprintf("submission:velocity:%s", userID.String())
	if count, err := d.redisClient.Get(ctx, velocityKey); err == nil {
		stats["velocity_count"] = count
	}

	return stats, nil
}
