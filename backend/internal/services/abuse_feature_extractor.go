package services

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"log"
	"strings"
	"time"

	"github.com/google/uuid"
	redispkg "git.subcult.tv/subculture-collective/clpr/pkg/redis"
)

// AbuseFeatureExtractor extracts features for anomaly detection
type AbuseFeatureExtractor struct {
	redisClient *redispkg.Client
}

// NewAbuseFeatureExtractor creates a new feature extractor
func NewAbuseFeatureExtractor(redisClient *redispkg.Client) *AbuseFeatureExtractor {
	return &AbuseFeatureExtractor{
		redisClient: redisClient,
	}
}

// AbuseFeatures represents extracted features for anomaly detection
type AbuseFeatures struct {
	// Velocity features
	VotesLast5Min       int64 `json:"votes_last_5min"`
	VotesLastHour       int64 `json:"votes_last_hour"`
	FollowsLast5Min     int64 `json:"follows_last_5min"`
	FollowsLastHour     int64 `json:"follows_last_hour"`
	SubmissionsLastHour int64 `json:"submissions_last_hour"`

	// IP/UA features
	IPSharedUserCount int64   `json:"ip_shared_user_count"`
	UASharedUserCount int64   `json:"ua_shared_user_count"`
	IPChangeFrequency float64 `json:"ip_change_frequency"`

	// Graph pattern features
	CircularFollowScore  float64 `json:"circular_follow_score"`
	CoordinatedVoteScore float64 `json:"coordinated_vote_score"`
	BurstScore           float64 `json:"burst_score"`

	// Behavioral features
	VotePatternDiversity float64 `json:"vote_pattern_diversity"`
	TimingEntropy        float64 `json:"timing_entropy"`
	AccountAge           int64   `json:"account_age_days"`
	TrustScore           int     `json:"trust_score"`
}

const (
	// Time windows for velocity tracking
	velocityWindow5Min   = 5 * time.Minute
	velocityWindow1Hour  = 1 * time.Hour
	velocityWindow24Hour = 24 * time.Hour

	// Feature tracking windows
	ipTrackingWindow    = 24 * time.Hour
	uaTrackingWindow    = 24 * time.Hour
	graphTrackingWindow = 7 * 24 * time.Hour
)

// ExtractVoteFeatures extracts features for a vote action
func (e *AbuseFeatureExtractor) ExtractVoteFeatures(ctx context.Context, userID uuid.UUID, clipID uuid.UUID, ip string, userAgent string, trustScore int, accountCreatedAt time.Time) (*AbuseFeatures, error) {
	features := &AbuseFeatures{
		TrustScore: trustScore,
		AccountAge: int64(time.Since(accountCreatedAt).Hours() / 24),
	}

	// Extract velocity features
	features.VotesLast5Min = e.getVelocityCount(ctx, userID, "vote", velocityWindow5Min)
	features.VotesLastHour = e.getVelocityCount(ctx, userID, "vote", velocityWindow1Hour)

	// Extract IP/UA features
	features.IPSharedUserCount = e.getIPUserCount(ctx, ip)
	features.UASharedUserCount = e.getUAUserCount(ctx, userAgent)
	features.IPChangeFrequency = e.getIPChangeFrequency(ctx, userID)

	// Extract graph pattern features
	features.CoordinatedVoteScore = e.getCoordinatedVoteScore(ctx, userID, clipID)
	features.BurstScore = e.calculateBurstScore(features.VotesLast5Min, 2) // threshold of 2 votes in 5 min

	// Extract behavioral features
	features.VotePatternDiversity = e.getVotePatternDiversity(ctx, userID)
	features.TimingEntropy = e.getTimingEntropy(ctx, userID, "vote")

	// Track this action for future feature extraction
	e.trackVoteAction(ctx, userID, clipID, ip, userAgent)

	return features, nil
}

// ExtractFollowFeatures extracts features for a follow action
func (e *AbuseFeatureExtractor) ExtractFollowFeatures(ctx context.Context, followerID uuid.UUID, followingID uuid.UUID, ip string, userAgent string, trustScore int, accountCreatedAt time.Time) (*AbuseFeatures, error) {
	features := &AbuseFeatures{
		TrustScore: trustScore,
		AccountAge: int64(time.Since(accountCreatedAt).Hours() / 24),
	}

	// Extract velocity features
	features.FollowsLast5Min = e.getVelocityCount(ctx, followerID, "follow", velocityWindow5Min)
	features.FollowsLastHour = e.getVelocityCount(ctx, followerID, "follow", velocityWindow1Hour)

	// Extract IP/UA features
	features.IPSharedUserCount = e.getIPUserCount(ctx, ip)
	features.UASharedUserCount = e.getUAUserCount(ctx, userAgent)
	features.IPChangeFrequency = e.getIPChangeFrequency(ctx, followerID)

	// Extract graph pattern features
	features.CircularFollowScore = e.getCircularFollowScore(ctx, followerID, followingID)
	features.BurstScore = e.calculateBurstScore(features.FollowsLast5Min, 3) // threshold of 3 follows in 5 min

	// Extract behavioral features
	features.TimingEntropy = e.getTimingEntropy(ctx, followerID, "follow")

	// Track this action for future feature extraction
	e.trackFollowAction(ctx, followerID, followingID, ip, userAgent)

	return features, nil
}

// ExtractSubmissionFeatures extracts features for a submission action
func (e *AbuseFeatureExtractor) ExtractSubmissionFeatures(ctx context.Context, userID uuid.UUID, ip string, userAgent string, trustScore int, accountCreatedAt time.Time) (*AbuseFeatures, error) {
	features := &AbuseFeatures{
		TrustScore: trustScore,
		AccountAge: int64(time.Since(accountCreatedAt).Hours() / 24),
	}

	// Extract velocity features
	features.SubmissionsLastHour = e.getVelocityCount(ctx, userID, "submission", velocityWindow1Hour)

	// Extract IP/UA features
	features.IPSharedUserCount = e.getIPUserCount(ctx, ip)
	features.UASharedUserCount = e.getUAUserCount(ctx, userAgent)
	features.IPChangeFrequency = e.getIPChangeFrequency(ctx, userID)

	// Extract behavioral features
	features.BurstScore = e.calculateBurstScore(features.SubmissionsLastHour, 3) // threshold of 3 submissions in 1 hour
	features.TimingEntropy = e.getTimingEntropy(ctx, userID, "submission")

	// Track this action for future feature extraction
	e.trackSubmissionAction(ctx, userID, ip, userAgent)

	return features, nil
}

// getVelocityCount gets the count of actions in the time window
func (e *AbuseFeatureExtractor) getVelocityCount(ctx context.Context, userID uuid.UUID, actionType string, window time.Duration) int64 {
	key := fmt.Sprintf("abuse:velocity:%s:%s", actionType, userID.String())
	count, err := e.redisClient.Get(ctx, key)
	if err != nil {
		return 0
	}

	var currentCount int64
	if _, err := fmt.Sscanf(count, "%d", &currentCount); err != nil {
		return 0
	}

	return currentCount
}

// getIPUserCount gets the number of unique users from this IP
func (e *AbuseFeatureExtractor) getIPUserCount(ctx context.Context, ip string) int64 {
	key := fmt.Sprintf("abuse:ip:%s", hashIP(ip))
	count, err := e.redisClient.SetCard(ctx, key)
	if err != nil {
		return 0
	}
	return count
}

// getUAUserCount gets the number of unique users with this user agent
func (e *AbuseFeatureExtractor) getUAUserCount(ctx context.Context, userAgent string) int64 {
	// Normalize user agent to avoid minor variations
	normalized := normalizeUserAgent(userAgent)
	key := fmt.Sprintf("abuse:ua:%s", normalized)
	count, err := e.redisClient.SetCard(ctx, key)
	if err != nil {
		return 0
	}
	return count
}

// getIPChangeFrequency calculates how frequently a user changes IPs
func (e *AbuseFeatureExtractor) getIPChangeFrequency(ctx context.Context, userID uuid.UUID) float64 {
	key := fmt.Sprintf("abuse:ip:history:%s", userID.String())
	count, err := e.redisClient.SetCard(ctx, key)
	if err != nil || count <= 1 {
		return 0.0
	}

	// Return number of unique IPs in the last 24 hours
	// High count indicates IP hopping behavior
	return float64(count)
}

// getCoordinatedVoteScore detects coordinated voting patterns
func (e *AbuseFeatureExtractor) getCoordinatedVoteScore(ctx context.Context, userID uuid.UUID, clipID uuid.UUID) float64 {
	// Get users who recently voted on the same clip
	key := fmt.Sprintf("abuse:clip:voters:%s", clipID.String())
	voters, err := e.redisClient.SetMembers(ctx, key)
	if err != nil || len(voters) < 2 {
		return 0.0
	}

	// Check for common IPs among recent voters
	commonIPCount := 0
	userIPs := make(map[string]int)

	for _, voterID := range voters {
		ipKey := fmt.Sprintf("abuse:user:last_ip:%s", voterID)
		ip, err := e.redisClient.Get(ctx, ipKey)
		if err == nil && ip != "" {
			userIPs[ip]++
			if userIPs[ip] > 1 {
				commonIPCount++
			}
		}
	}

	// Score based on percentage of voters from common IPs
	if len(voters) > 0 {
		return float64(commonIPCount) / float64(len(voters))
	}

	return 0.0
}

// getCircularFollowScore detects circular follow patterns (A follows B, B follows C, C follows A)
func (e *AbuseFeatureExtractor) getCircularFollowScore(ctx context.Context, followerID uuid.UUID, followingID uuid.UUID) float64 {
	// Check if followingID follows followerID (mutual follow)
	mutualKey := fmt.Sprintf("abuse:follows:%s", followingID.String())
	isMutual, err := e.redisClient.SetIsMember(ctx, mutualKey, followerID.String())
	if err != nil {
		return 0.0
	}

	if isMutual {
		// Mutual follows are somewhat suspicious but common
		return 0.3
	}

	// Check for longer circular patterns (computationally expensive, so we limit depth)
	// This is a simplified check - production systems would use graph algorithms
	return 0.0
}

// calculateBurstScore calculates a burst score based on velocity
func (e *AbuseFeatureExtractor) calculateBurstScore(count int64, threshold int64) float64 {
	if count <= threshold {
		return 0.0
	}

	// Exponential scoring for exceeding threshold
	excess := float64(count - threshold)
	score := 1.0 - (1.0 / (1.0 + excess/float64(threshold)))

	if score > 1.0 {
		score = 1.0
	}

	return score
}

// getVotePatternDiversity measures diversity in voting patterns
func (e *AbuseFeatureExtractor) getVotePatternDiversity(ctx context.Context, userID uuid.UUID) float64 {
	// Track upvote/downvote ratio
	upvoteKey := fmt.Sprintf("abuse:votes:up:%s", userID.String())
	downvoteKey := fmt.Sprintf("abuse:votes:down:%s", userID.String())

	upvotes, err1 := e.redisClient.Get(ctx, upvoteKey)
	downvotes, err2 := e.redisClient.Get(ctx, downvoteKey)

	if err1 != nil || err2 != nil {
		return 0.5 // neutral diversity - no data available
	}

	var upCount, downCount int64
	_, err1 = fmt.Sscanf(upvotes, "%d", &upCount)
	_, err2 = fmt.Sscanf(downvotes, "%d", &downCount)

	if err1 != nil || err2 != nil {
		// Data corrupted or in unexpected format
		log.Printf("Warning: failed to parse vote counts for user %s: up_err=%v down_err=%v", userID, err1, err2)
		return 0.5 // neutral diversity - corrupted data
	}

	total := upCount + downCount
	if total == 0 {
		return 0.5
	}

	// Calculate entropy-like diversity score
	// 0.5 = balanced voting, 0.0 or 1.0 = only one type of vote
	ratio := float64(upCount) / float64(total)

	// Penalize extreme ratios (all upvotes or all downvotes)
	diversity := 1.0 - (2.0 * abs(ratio-0.5))

	return diversity
}

// getTimingEntropy measures randomness in action timing
func (e *AbuseFeatureExtractor) getTimingEntropy(ctx context.Context, userID uuid.UUID, actionType string) float64 {
	// Get recent action timestamps
	key := fmt.Sprintf("abuse:timing:%s:%s", actionType, userID.String())
	timestamps, err := e.redisClient.ListRange(ctx, key, 0, 9) // last 10 actions
	if err != nil || len(timestamps) < 2 {
		return 0.5 // default neutral entropy
	}

	// Calculate intervals between actions
	intervals := make([]float64, 0, len(timestamps)-1)
	for i := 1; i < len(timestamps); i++ {
		var t1, t2 int64
		fmt.Sscanf(timestamps[i-1], "%d", &t1)
		fmt.Sscanf(timestamps[i], "%d", &t2)

		interval := float64(t2 - t1)
		if interval > 0 {
			intervals = append(intervals, interval)
		}
	}

	if len(intervals) == 0 {
		return 0.5
	}

	// Calculate standard deviation as entropy measure
	mean := 0.0
	for _, interval := range intervals {
		mean += interval
	}
	mean /= float64(len(intervals))

	variance := 0.0
	for _, interval := range intervals {
		diff := interval - mean
		variance += diff * diff
	}
	variance /= float64(len(intervals))

	// Normalize entropy score (higher std dev = higher entropy = more natural)
	// Very low entropy (< 1 second std dev) is suspicious
	entropy := 0.5 // default neutral entropy
	if mean != 0 {
		entropy = variance / (mean * mean)
		if entropy > 1.0 {
			entropy = 1.0
		}
	}

	return entropy
}

// trackVoteAction tracks a vote action for future feature extraction
func (e *AbuseFeatureExtractor) trackVoteAction(ctx context.Context, userID uuid.UUID, clipID uuid.UUID, ip string, userAgent string) {
	// Track velocity
	velocityKey := fmt.Sprintf("abuse:velocity:vote:%s", userID.String())
	count, _ := e.redisClient.Increment(ctx, velocityKey)
	if count == 1 {
		e.redisClient.Expire(ctx, velocityKey, velocityWindow1Hour)
	}

	// Track IP (hashed for privacy)
	ipKey := fmt.Sprintf("abuse:ip:%s", hashIP(ip))
	e.redisClient.SetAdd(ctx, ipKey, userID.String())
	e.redisClient.Expire(ctx, ipKey, ipTrackingWindow)

	// Track user's last IP (hashed for privacy)
	userIPKey := fmt.Sprintf("abuse:user:last_ip:%s", userID.String())
	e.redisClient.Set(ctx, userIPKey, hashIP(ip), ipTrackingWindow)

	// Track IP history (hashed for privacy)
	ipHistoryKey := fmt.Sprintf("abuse:ip:history:%s", userID.String())
	e.redisClient.SetAdd(ctx, ipHistoryKey, hashIP(ip))
	e.redisClient.Expire(ctx, ipHistoryKey, ipTrackingWindow)

	// Track UA
	normalized := normalizeUserAgent(userAgent)
	uaKey := fmt.Sprintf("abuse:ua:%s", normalized)
	e.redisClient.SetAdd(ctx, uaKey, userID.String())
	e.redisClient.Expire(ctx, uaKey, uaTrackingWindow)

	// Track voters on this clip
	clipVotersKey := fmt.Sprintf("abuse:clip:voters:%s", clipID.String())
	e.redisClient.SetAdd(ctx, clipVotersKey, userID.String())
	e.redisClient.Expire(ctx, clipVotersKey, 1*time.Hour)

	// Track timing
	timingKey := fmt.Sprintf("abuse:timing:vote:%s", userID.String())
	timestamp := fmt.Sprintf("%d", time.Now().Unix())
	e.redisClient.ListPush(ctx, timingKey, timestamp)
	e.redisClient.ListTrim(ctx, timingKey, 0, 19) // keep last 20
	e.redisClient.Expire(ctx, timingKey, velocityWindow24Hour)
}

// trackFollowAction tracks a follow action for future feature extraction
func (e *AbuseFeatureExtractor) trackFollowAction(ctx context.Context, followerID uuid.UUID, followingID uuid.UUID, ip string, userAgent string) {
	// Track velocity
	velocityKey := fmt.Sprintf("abuse:velocity:follow:%s", followerID.String())
	count, _ := e.redisClient.Increment(ctx, velocityKey)
	if count == 1 {
		e.redisClient.Expire(ctx, velocityKey, velocityWindow1Hour)
	}

	// Track IP (hashed for privacy)
	ipKey := fmt.Sprintf("abuse:ip:%s", hashIP(ip))
	e.redisClient.SetAdd(ctx, ipKey, followerID.String())
	e.redisClient.Expire(ctx, ipKey, ipTrackingWindow)

	// Track user's last IP (hashed for privacy)
	userIPKey := fmt.Sprintf("abuse:user:last_ip:%s", followerID.String())
	e.redisClient.Set(ctx, userIPKey, hashIP(ip), ipTrackingWindow)

	// Track IP history (hashed for privacy)
	ipHistoryKey := fmt.Sprintf("abuse:ip:history:%s", followerID.String())
	e.redisClient.SetAdd(ctx, ipHistoryKey, hashIP(ip))
	e.redisClient.Expire(ctx, ipHistoryKey, ipTrackingWindow)

	// Track UA
	normalized := normalizeUserAgent(userAgent)
	uaKey := fmt.Sprintf("abuse:ua:%s", normalized)
	e.redisClient.SetAdd(ctx, uaKey, followerID.String())
	e.redisClient.Expire(ctx, uaKey, uaTrackingWindow)

	// Track follow relationships for circular detection
	followsKey := fmt.Sprintf("abuse:follows:%s", followerID.String())
	e.redisClient.SetAdd(ctx, followsKey, followingID.String())
	e.redisClient.Expire(ctx, followsKey, graphTrackingWindow)

	// Track timing
	timingKey := fmt.Sprintf("abuse:timing:follow:%s", followerID.String())
	timestamp := fmt.Sprintf("%d", time.Now().Unix())
	e.redisClient.ListPush(ctx, timingKey, timestamp)
	e.redisClient.ListTrim(ctx, timingKey, 0, 19) // keep last 20
	e.redisClient.Expire(ctx, timingKey, velocityWindow24Hour)
}

// trackSubmissionAction tracks a submission action for future feature extraction
func (e *AbuseFeatureExtractor) trackSubmissionAction(ctx context.Context, userID uuid.UUID, ip string, userAgent string) {
	// Track velocity
	velocityKey := fmt.Sprintf("abuse:velocity:submission:%s", userID.String())
	count, _ := e.redisClient.Increment(ctx, velocityKey)
	if count == 1 {
		e.redisClient.Expire(ctx, velocityKey, velocityWindow1Hour)
	}

	// Track IP (hashed for privacy)
	ipKey := fmt.Sprintf("abuse:ip:%s", hashIP(ip))
	e.redisClient.SetAdd(ctx, ipKey, userID.String())
	e.redisClient.Expire(ctx, ipKey, ipTrackingWindow)

	// Track user's last IP (hashed for privacy)
	userIPKey := fmt.Sprintf("abuse:user:last_ip:%s", userID.String())
	e.redisClient.Set(ctx, userIPKey, hashIP(ip), ipTrackingWindow)

	// Track IP history (hashed for privacy)
	ipHistoryKey := fmt.Sprintf("abuse:ip:history:%s", userID.String())
	e.redisClient.SetAdd(ctx, ipHistoryKey, hashIP(ip))
	e.redisClient.Expire(ctx, ipHistoryKey, ipTrackingWindow)

	// Track UA
	normalized := normalizeUserAgent(userAgent)
	uaKey := fmt.Sprintf("abuse:ua:%s", normalized)
	e.redisClient.SetAdd(ctx, uaKey, userID.String())
	e.redisClient.Expire(ctx, uaKey, uaTrackingWindow)

	// Track timing
	timingKey := fmt.Sprintf("abuse:timing:submission:%s", userID.String())
	timestamp := fmt.Sprintf("%d", time.Now().Unix())
	e.redisClient.ListPush(ctx, timingKey, timestamp)
	e.redisClient.ListTrim(ctx, timingKey, 0, 19) // keep last 20
	e.redisClient.Expire(ctx, timingKey, velocityWindow24Hour)
}

// normalizeUserAgent normalizes a user agent string for comparison by hashing
func normalizeUserAgent(ua string) string {
	// Convert to lowercase for case-insensitive comparison
	ua = strings.ToLower(ua)

	// Hash the user agent to create a fixed-size key while preserving privacy
	hash := sha256.Sum256([]byte(ua))
	return hex.EncodeToString(hash[:16]) // Use first 16 bytes (32 hex chars)
}

// abs returns absolute value of float64
func abs(x float64) float64 {
	if x < 0 {
		return -x
	}
	return x
}

// hashIP hashes an IP address for privacy protection while maintaining uniqueness
func hashIP(ip string) string {
	hash := sha256.Sum256([]byte(ip))
	return hex.EncodeToString(hash[:16]) // Use first 16 bytes (32 hex chars)
}
