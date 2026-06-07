package services

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"math"
	"time"

	"github.com/google/uuid"
	redispkg "git.subcult.tv/subculture-collective/clpr/pkg/redis"
)

// AnomalyScorer scores actions for abuse and anomaly patterns
type AnomalyScorer struct {
	redisClient        *redispkg.Client
	featureExtractor   *AbuseFeatureExtractor
	moderationEventSvc *ModerationEventService
}

// NewAnomalyScorer creates a new anomaly scorer
func NewAnomalyScorer(redisClient *redispkg.Client, featureExtractor *AbuseFeatureExtractor, moderationEventSvc *ModerationEventService) *AnomalyScorer {
	return &AnomalyScorer{
		redisClient:        redisClient,
		featureExtractor:   featureExtractor,
		moderationEventSvc: moderationEventSvc,
	}
}

// AnomalyScore represents the result of anomaly scoring
type AnomalyScore struct {
	OverallScore    float64            `json:"overall_score"`    // 0.0-1.0, higher = more suspicious
	ConfidenceScore float64            `json:"confidence_score"` // 0.0-1.0, confidence in the score
	IsAnomaly       bool               `json:"is_anomaly"`       // true if score exceeds threshold
	Severity        string             `json:"severity"`         // "low", "medium", "high", "critical"
	ReasonCodes     []string           `json:"reason_codes"`     // Why it was flagged
	ComponentScores map[string]float64 `json:"component_scores"` // Individual feature scores
	Features        *AbuseFeatures     `json:"features"`         // Extracted features
	ShouldAutoFlag  bool               `json:"should_auto_flag"` // Should create moderation queue entry
}

const (
	// Anomaly thresholds
	anomalyThresholdLow      = 0.30
	anomalyThresholdMedium   = 0.50
	anomalyThresholdHigh     = 0.70
	anomalyThresholdCritical = 0.85

	// Auto-flag thresholds (more conservative to keep FPR < 2%)
	autoFlagThreshold = 0.75

	// Feature weights for scoring
	weightVelocity     = 0.25
	weightIPUA         = 0.20
	weightGraphPattern = 0.25
	weightBehavioral   = 0.15
	weightTrustScore   = 0.15
)

// ScoreVoteAction scores a vote action for anomalies
func (s *AnomalyScorer) ScoreVoteAction(ctx context.Context, userID uuid.UUID, clipID uuid.UUID, voteType int16, ip string, userAgent string, trustScore int, accountCreatedAt time.Time) (*AnomalyScore, error) {
	// Extract features
	features, err := s.featureExtractor.ExtractVoteFeatures(ctx, userID, clipID, ip, userAgent, trustScore, accountCreatedAt)
	if err != nil {
		return nil, fmt.Errorf("failed to extract features: %w", err)
	}

	// Calculate component scores
	componentScores := make(map[string]float64)
	reasonCodes := make([]string, 0)

	// Velocity scoring
	velocityScore := s.scoreVelocity(features.VotesLast5Min, features.VotesLastHour, 2, 10)
	componentScores["velocity"] = velocityScore
	if velocityScore > 0.7 {
		reasonCodes = append(reasonCodes, "VOTE_VELOCITY_HIGH")
	}

	// IP/UA scoring
	ipuaScore := s.scoreIPUA(features.IPSharedUserCount, features.UASharedUserCount, features.IPChangeFrequency)
	componentScores["ip_ua"] = ipuaScore
	if features.IPSharedUserCount > 10 {
		reasonCodes = append(reasonCodes, "IP_SHARED_MULTIPLE_ACCOUNTS")
	}
	if features.IPChangeFrequency > 5 {
		reasonCodes = append(reasonCodes, "IP_HOPPING_DETECTED")
	}

	// Graph pattern scoring
	graphScore := s.scoreGraphPatterns(features.CoordinatedVoteScore, features.CircularFollowScore, features.BurstScore)
	componentScores["graph_pattern"] = graphScore
	if features.CoordinatedVoteScore > 0.5 {
		reasonCodes = append(reasonCodes, "COORDINATED_VOTING_DETECTED")
	}
	if features.BurstScore > 0.7 {
		reasonCodes = append(reasonCodes, "BURST_ACTIVITY_DETECTED")
	}

	// Behavioral scoring
	behavioralScore := s.scoreBehavioral(features.VotePatternDiversity, features.TimingEntropy, features.AccountAge)
	componentScores["behavioral"] = behavioralScore
	if features.VotePatternDiversity < 0.2 {
		reasonCodes = append(reasonCodes, "VOTE_PATTERN_MONOTONOUS")
	}
	if features.TimingEntropy < 0.1 {
		reasonCodes = append(reasonCodes, "TIMING_PATTERN_SUSPICIOUS")
	}

	// Trust score contribution (inverse - low trust = higher anomaly)
	trustScoreComponent := s.scoreTrustScore(features.TrustScore)
	componentScores["trust_score"] = trustScoreComponent
	if features.TrustScore < 30 {
		reasonCodes = append(reasonCodes, "LOW_TRUST_SCORE")
	}

	// Calculate overall score (weighted average)
	overallScore := (velocityScore * weightVelocity) +
		(ipuaScore * weightIPUA) +
		(graphScore * weightGraphPattern) +
		(behavioralScore * weightBehavioral) +
		(trustScoreComponent * weightTrustScore)

	// Calculate confidence based on number of data points
	confidence := s.calculateConfidence(features)

	// Determine severity and auto-flag decision
	severity := s.determineSeverity(overallScore)
	shouldAutoFlag := overallScore >= autoFlagThreshold && confidence >= 0.6

	score := &AnomalyScore{
		OverallScore:    overallScore,
		ConfidenceScore: confidence,
		IsAnomaly:       overallScore >= anomalyThresholdMedium,
		Severity:        severity,
		ReasonCodes:     reasonCodes,
		ComponentScores: componentScores,
		Features:        features,
		ShouldAutoFlag:  shouldAutoFlag,
	}

	// Store score for metrics
	s.storeScoreMetrics(ctx, "vote", score)

	return score, nil
}

// ScoreFollowAction scores a follow action for anomalies
func (s *AnomalyScorer) ScoreFollowAction(ctx context.Context, followerID uuid.UUID, followingID uuid.UUID, ip string, userAgent string, trustScore int, accountCreatedAt time.Time) (*AnomalyScore, error) {
	// Extract features
	features, err := s.featureExtractor.ExtractFollowFeatures(ctx, followerID, followingID, ip, userAgent, trustScore, accountCreatedAt)
	if err != nil {
		return nil, fmt.Errorf("failed to extract features: %w", err)
	}

	// Calculate component scores
	componentScores := make(map[string]float64)
	reasonCodes := make([]string, 0)

	// Velocity scoring
	velocityScore := s.scoreVelocity(features.FollowsLast5Min, features.FollowsLastHour, 3, 15)
	componentScores["velocity"] = velocityScore
	if velocityScore > 0.7 {
		reasonCodes = append(reasonCodes, "FOLLOW_VELOCITY_HIGH")
	}

	// IP/UA scoring
	ipuaScore := s.scoreIPUA(features.IPSharedUserCount, features.UASharedUserCount, features.IPChangeFrequency)
	componentScores["ip_ua"] = ipuaScore
	if features.IPSharedUserCount > 10 {
		reasonCodes = append(reasonCodes, "IP_SHARED_MULTIPLE_ACCOUNTS")
	}

	// Graph pattern scoring
	graphScore := s.scoreGraphPatterns(features.CoordinatedVoteScore, features.CircularFollowScore, features.BurstScore)
	componentScores["graph_pattern"] = graphScore
	if features.CircularFollowScore > 0.3 {
		reasonCodes = append(reasonCodes, "CIRCULAR_FOLLOW_PATTERN")
	}
	if features.BurstScore > 0.7 {
		reasonCodes = append(reasonCodes, "BURST_ACTIVITY_DETECTED")
	}

	// Behavioral scoring
	behavioralScore := s.scoreBehavioral(1.0, features.TimingEntropy, features.AccountAge) // no vote diversity for follows
	componentScores["behavioral"] = behavioralScore
	if features.TimingEntropy < 0.1 {
		reasonCodes = append(reasonCodes, "TIMING_PATTERN_SUSPICIOUS")
	}

	// Trust score contribution
	trustScoreComponent := s.scoreTrustScore(features.TrustScore)
	componentScores["trust_score"] = trustScoreComponent
	if features.TrustScore < 30 {
		reasonCodes = append(reasonCodes, "LOW_TRUST_SCORE")
	}

	// Calculate overall score
	overallScore := (velocityScore * weightVelocity) +
		(ipuaScore * weightIPUA) +
		(graphScore * weightGraphPattern) +
		(behavioralScore * weightBehavioral) +
		(trustScoreComponent * weightTrustScore)

	confidence := s.calculateConfidence(features)
	severity := s.determineSeverity(overallScore)
	shouldAutoFlag := overallScore >= autoFlagThreshold && confidence >= 0.6

	score := &AnomalyScore{
		OverallScore:    overallScore,
		ConfidenceScore: confidence,
		IsAnomaly:       overallScore >= anomalyThresholdMedium,
		Severity:        severity,
		ReasonCodes:     reasonCodes,
		ComponentScores: componentScores,
		Features:        features,
		ShouldAutoFlag:  shouldAutoFlag,
	}

	// Store score for metrics
	s.storeScoreMetrics(ctx, "follow", score)

	return score, nil
}

// ScoreSubmissionAction scores a submission action for anomalies
func (s *AnomalyScorer) ScoreSubmissionAction(ctx context.Context, userID uuid.UUID, ip string, userAgent string, trustScore int, accountCreatedAt time.Time) (*AnomalyScore, error) {
	// Extract features
	features, err := s.featureExtractor.ExtractSubmissionFeatures(ctx, userID, ip, userAgent, trustScore, accountCreatedAt)
	if err != nil {
		return nil, fmt.Errorf("failed to extract features: %w", err)
	}

	// Calculate component scores
	componentScores := make(map[string]float64)
	reasonCodes := make([]string, 0)

	// Velocity scoring (submissions only tracked hourly, not in 5-min windows)
	velocityScore := s.scoreVelocity(0, features.SubmissionsLastHour, 3, 8)
	componentScores["velocity"] = velocityScore
	if velocityScore > 0.7 {
		reasonCodes = append(reasonCodes, "SUBMISSION_VELOCITY_HIGH")
	}

	// IP/UA scoring
	ipuaScore := s.scoreIPUA(features.IPSharedUserCount, features.UASharedUserCount, features.IPChangeFrequency)
	componentScores["ip_ua"] = ipuaScore
	if features.IPSharedUserCount > 10 {
		reasonCodes = append(reasonCodes, "IP_SHARED_MULTIPLE_ACCOUNTS")
	}

	// Behavioral scoring
	behavioralScore := s.scoreBehavioral(1.0, features.TimingEntropy, features.AccountAge)
	componentScores["behavioral"] = behavioralScore
	if features.TimingEntropy < 0.1 {
		reasonCodes = append(reasonCodes, "TIMING_PATTERN_SUSPICIOUS")
	}
	if features.AccountAge < 1 {
		reasonCodes = append(reasonCodes, "NEW_ACCOUNT")
	}

	// Trust score contribution
	trustScoreComponent := s.scoreTrustScore(features.TrustScore)
	componentScores["trust_score"] = trustScoreComponent
	if features.TrustScore < 30 {
		reasonCodes = append(reasonCodes, "LOW_TRUST_SCORE")
	}

	// Calculate overall score (different weights for submissions)
	overallScore := (velocityScore * 0.30) +
		(ipuaScore * 0.25) +
		(behavioralScore * 0.20) +
		(trustScoreComponent * 0.25)

	confidence := s.calculateConfidence(features)
	severity := s.determineSeverity(overallScore)
	shouldAutoFlag := overallScore >= autoFlagThreshold && confidence >= 0.6

	score := &AnomalyScore{
		OverallScore:    overallScore,
		ConfidenceScore: confidence,
		IsAnomaly:       overallScore >= anomalyThresholdMedium,
		Severity:        severity,
		ReasonCodes:     reasonCodes,
		ComponentScores: componentScores,
		Features:        features,
		ShouldAutoFlag:  shouldAutoFlag,
	}

	// Store score for metrics
	s.storeScoreMetrics(ctx, "submission", score)

	return score, nil
}

// scoreVelocity scores velocity features
func (s *AnomalyScorer) scoreVelocity(shortTermCount int64, longTermCount int64, shortThreshold int64, longThreshold int64) float64 {
	// Validate thresholds to prevent division by zero
	if shortThreshold <= 0 {
		shortThreshold = 1
	}
	if longThreshold <= 0 {
		longThreshold = 1
	}

	// Calculate score based on how much thresholds are exceeded
	shortScore := 0.0
	if shortTermCount > shortThreshold {
		shortScore = math.Min(1.0, float64(shortTermCount-shortThreshold)/float64(shortThreshold))
	}

	longScore := 0.0
	if longTermCount > longThreshold {
		longScore = math.Min(1.0, float64(longTermCount-longThreshold)/float64(longThreshold))
	}

	// Weight short term more heavily (more indicative of abuse)
	return (shortScore * 0.7) + (longScore * 0.3)
}

// scoreIPUA scores IP and user agent features
func (s *AnomalyScorer) scoreIPUA(ipUserCount int64, uaUserCount int64, ipChangeFreq float64) float64 {
	// Score based on shared IP/UA
	ipScore := 0.0
	if ipUserCount > 5 {
		ipScore = math.Min(1.0, float64(ipUserCount-5)/10.0)
	}

	uaScore := 0.0
	if uaUserCount > 3 {
		uaScore = math.Min(1.0, float64(uaUserCount-3)/7.0)
	}

	// Score based on IP hopping
	ipHopScore := math.Min(1.0, ipChangeFreq/10.0)

	// Combine scores
	return (ipScore * 0.5) + (uaScore * 0.2) + (ipHopScore * 0.3)
}

// scoreGraphPatterns scores graph pattern features
func (s *AnomalyScorer) scoreGraphPatterns(coordVoteScore float64, circularFollowScore float64, burstScore float64) float64 {
	// All scores are already 0.0-1.0
	return (coordVoteScore * 0.4) + (circularFollowScore * 0.3) + (burstScore * 0.3)
}

// scoreBehavioral scores behavioral features
func (s *AnomalyScorer) scoreBehavioral(voteDiversity float64, timingEntropy float64, accountAgeDays int64) float64 {
	// Low diversity and entropy are suspicious
	diversityScore := 1.0 - voteDiversity // invert so low diversity = high score
	entropyScore := 1.0 - timingEntropy   // invert so low entropy = high score

	// New accounts are more risky
	ageScore := 0.0
	if accountAgeDays < 7 {
		ageScore = 0.8
	} else if accountAgeDays < 30 {
		ageScore = 0.5
	} else if accountAgeDays < 90 {
		ageScore = 0.2
	}

	return (diversityScore * 0.3) + (entropyScore * 0.4) + (ageScore * 0.3)
}

// scoreTrustScore scores based on trust score (inverse)
func (s *AnomalyScorer) scoreTrustScore(trustScore int) float64 {
	// Trust score ranges 0-100, high trust = low anomaly
	// Invert and normalize to 0.0-1.0
	if trustScore >= 80 {
		return 0.0
	} else if trustScore >= 50 {
		return 0.3
	} else if trustScore >= 30 {
		return 0.6
	} else {
		return 0.9
	}
}

// calculateConfidence calculates confidence in the score based on data availability
func (s *AnomalyScorer) calculateConfidence(features *AbuseFeatures) float64 {
	confidence := 0.0
	dataPoints := 0

	// Check which features have meaningful data
	if features.VotesLastHour > 0 || features.FollowsLastHour > 0 || features.SubmissionsLastHour > 0 {
		confidence += 0.2
		dataPoints++
	}

	if features.IPSharedUserCount > 1 {
		confidence += 0.2
		dataPoints++
	}

	if features.AccountAge > 0 {
		confidence += 0.2
		dataPoints++
	}

	if features.TimingEntropy > 0 {
		confidence += 0.2
		dataPoints++
	}

	if features.TrustScore > 0 {
		confidence += 0.2
		dataPoints++
	}

	// Older accounts with more history = higher confidence
	if features.AccountAge > 30 {
		confidence += 0.1
	}

	return math.Min(1.0, confidence)
}

// determineSeverity determines severity level from score
func (s *AnomalyScorer) determineSeverity(score float64) string {
	if score >= anomalyThresholdCritical {
		return "critical"
	} else if score >= anomalyThresholdHigh {
		return "high"
	} else if score >= anomalyThresholdMedium {
		return "medium"
	} else if score >= anomalyThresholdLow {
		return "low"
	}
	return "none"
}

// storeScoreMetrics stores scoring metrics for monitoring
func (s *AnomalyScorer) storeScoreMetrics(ctx context.Context, actionType string, score *AnomalyScore) {
	// Store in Redis for dashboard metrics
	timestamp := time.Now().Unix()

	// Store overall score distribution
	scoreKey := fmt.Sprintf("abuse:metrics:scores:%s:%d", actionType, timestamp/3600) // hourly buckets
	scoreJSON, err := json.Marshal(score)
	if err != nil {
		log.Printf("Failed to marshal score for metrics: %v", err)
		return
	}
	if err := s.redisClient.ListPush(ctx, scoreKey, string(scoreJSON)); err != nil {
		log.Printf("Failed to store score metrics: %v", err)
	}
	s.redisClient.Expire(ctx, scoreKey, 7*24*time.Hour) // keep for 7 days

	// Increment counters for dashboard
	if score.IsAnomaly {
		anomalyKey := fmt.Sprintf("abuse:metrics:anomalies:%s", actionType)
		s.redisClient.Increment(ctx, anomalyKey)
		s.redisClient.Expire(ctx, anomalyKey, 24*time.Hour)
	}

	if score.ShouldAutoFlag {
		flagKey := fmt.Sprintf("abuse:metrics:auto_flagged:%s", actionType)
		s.redisClient.Increment(ctx, flagKey)
		s.redisClient.Expire(ctx, flagKey, 24*time.Hour)
	}

	// Store by severity
	severityKey := fmt.Sprintf("abuse:metrics:severity:%s:%s", actionType, score.Severity)
	s.redisClient.Increment(ctx, severityKey)
	s.redisClient.Expire(ctx, severityKey, 24*time.Hour)
}

// GetMetrics returns aggregated abuse detection metrics
func (s *AnomalyScorer) GetMetrics(ctx context.Context) (map[string]interface{}, error) {
	metrics := make(map[string]interface{})

	// Get anomaly counts by action type
	for _, actionType := range []string{"vote", "follow", "submission"} {
		anomalyKey := fmt.Sprintf("abuse:metrics:anomalies:%s", actionType)
		count, err := s.redisClient.Get(ctx, anomalyKey)
		if err == nil {
			metrics[fmt.Sprintf("anomalies_%s", actionType)] = count
		}

		flagKey := fmt.Sprintf("abuse:metrics:auto_flagged:%s", actionType)
		flagCount, err := s.redisClient.Get(ctx, flagKey)
		if err == nil {
			metrics[fmt.Sprintf("auto_flagged_%s", actionType)] = flagCount
		}

		// Get severity breakdown
		severityBreakdown := make(map[string]string)
		for _, severity := range []string{"low", "medium", "high", "critical"} {
			severityKey := fmt.Sprintf("abuse:metrics:severity:%s:%s", actionType, severity)
			sevCount, err := s.redisClient.Get(ctx, severityKey)
			if err == nil {
				severityBreakdown[severity] = sevCount
			}
		}
		metrics[fmt.Sprintf("severity_%s", actionType)] = severityBreakdown
	}

	return metrics, nil
}
