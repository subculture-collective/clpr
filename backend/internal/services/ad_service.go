package services

import (
	"context"
	"fmt"
	"hash/fnv"
	"math/rand/v2"
	"sort"
	"time"

	"github.com/google/uuid"
	"git.subcult.tv/subculture-collective/clpr/internal/models"
	"git.subcult.tv/subculture-collective/clpr/internal/repository"
	redispkg "git.subcult.tv/subculture-collective/clpr/pkg/redis"
)

// experimentBucketCount is the number of buckets used for A/B experiment user distribution
// A higher number provides finer-grained distribution but 10000 provides adequate precision
const experimentBucketCount = 10000

// AdService handles business logic for ad delivery
type AdService struct {
	adRepo      *repository.AdRepository
	redisClient *redispkg.Client
}

// NewAdService creates a new AdService
func NewAdService(adRepo *repository.AdRepository, redisClient *redispkg.Client) *AdService {
	return &AdService{
		adRepo:      adRepo,
		redisClient: redisClient,
	}
}

// SelectAd selects an appropriate ad for display based on targeting, frequency caps, and fraud prevention
func (s *AdService) SelectAd(ctx context.Context, req models.AdSelectionRequest, userID *uuid.UUID, ipAddress string) (*models.AdSelectionResponse, error) {
	// Check if personalized ads are allowed
	// If not, we only use contextual targeting (game, language, platform)
	// and skip user-specific targeting (interests, user history, country/device)
	isPersonalized := req.Personalized != nil && *req.Personalized

	// Get all active ads matching the request criteria
	ads, err := s.adRepo.GetActiveAds(ctx, req.AdType, req.Width, req.Height)
	if err != nil {
		return nil, fmt.Errorf("failed to get active ads: %w", err)
	}

	if len(ads) == 0 {
		return &models.AdSelectionResponse{}, nil
	}

	// Filter ads based on targeting criteria
	// If personalized is false, only use contextual targeting (game, language, platform)
	if isPersonalized {
		// Full targeting including user-specific criteria
		ads = s.filterByTargeting(ads, req)
	} else {
		// Contextual-only targeting (no user-specific data)
		ads = s.filterByContextualTargeting(ads, req)
	}

	// Filter ads based on targeting rules (structured rules from database)
	// Only apply user-specific rules if personalized
	if isPersonalized {
		ads, err = s.filterByTargetingRules(ctx, ads, req)
		if err != nil {
			return nil, fmt.Errorf("failed to filter by targeting rules: %w", err)
		}
	} else {
		// Only apply contextual targeting rules
		ads, err = s.filterByContextualTargetingRules(ctx, ads, req)
		if err != nil {
			return nil, fmt.Errorf("failed to filter by contextual targeting rules: %w", err)
		}
	}

	if len(ads) == 0 {
		return &models.AdSelectionResponse{}, nil
	}

	// Filter ads based on frequency caps
	// Only apply if personalized (requires user/session tracking)
	if isPersonalized {
		ads, err = s.filterByFrequencyCaps(ctx, ads, userID, req.SessionID)
		if err != nil {
			return nil, fmt.Errorf("failed to filter by frequency caps: %w", err)
		}
	}

	if len(ads) == 0 {
		return &models.AdSelectionResponse{}, nil
	}

	// Basic fraud prevention: check for rapid impressions from same IP
	// This is non-personalized as it uses IP address for security, not targeting
	if ipAddress != "" {
		ads, err = s.filterByFraudPrevention(ctx, ads, ipAddress)
		if err != nil {
			return nil, fmt.Errorf("failed to filter by fraud prevention: %w", err)
		}
	}

	if len(ads) == 0 {
		return &models.AdSelectionResponse{}, nil
	}

	// Apply experiment selection if applicable (only if personalized)
	var selectedAd models.Ad
	if isPersonalized {
		selectedAd = s.selectAdWithExperiment(ads, userID, req.SessionID)
	} else {
		// Random selection without user-specific bucketing
		selectedAd = s.weightedRandomSelect(ads)
	}

	// Create impression record with enhanced tracking fields
	// If not personalized, we don't record user-identifiable information
	impressionID := uuid.New()
	impression := &models.AdImpression{
		ID:        impressionID,
		AdID:      selectedAd.ID,
		Platform:  req.Platform,
		PageURL:   req.PageURL,
		SlotID:    req.SlotID,
		CreatedAt: time.Now().UTC(),
	}

	// Only include user-identifiable data if personalized
	if isPersonalized {
		impression.UserID = userID
		impression.SessionID = req.SessionID
		impression.IPAddress = &ipAddress
		impression.Country = req.Country
		impression.DeviceType = req.DeviceType
		impression.ExperimentID = selectedAd.ExperimentID
		impression.ExperimentVariant = selectedAd.ExperimentVariant
	}

	if err := s.adRepo.CreateImpression(ctx, impression); err != nil {
		return nil, fmt.Errorf("failed to create impression: %w", err)
	}

	// Update frequency caps (async to not block response)
	// Only if personalized
	if isPersonalized {
		go s.updateFrequencyCaps(context.Background(), selectedAd.ID, userID, req.SessionID)
	}

	return &models.AdSelectionResponse{
		Ad:           &selectedAd,
		ImpressionID: impressionID.String(),
		TrackingURL:  fmt.Sprintf("/api/v1/ads/track/%s", impressionID.String()),
	}, nil
}

// TrackImpression updates an impression with viewability and click data
func (s *AdService) TrackImpression(ctx context.Context, req models.AdTrackingRequest) error {
	impressionID, err := uuid.Parse(req.ImpressionID)
	if err != nil {
		return fmt.Errorf("invalid impression ID: %w", err)
	}

	// Get the impression to validate it exists
	impression, err := s.adRepo.GetImpressionByID(ctx, impressionID)
	if err != nil {
		return fmt.Errorf("impression not found: %w", err)
	}

	// Determine if viewability threshold is met (IAB standard: 50% visible for 1s)
	isViewable := req.ViewabilityTimeMs >= models.ViewabilityThresholdMs

	// Update the impression
	if err := s.adRepo.UpdateImpression(ctx, impressionID, req.ViewabilityTimeMs, isViewable, req.IsClicked); err != nil {
		return fmt.Errorf("failed to update impression: %w", err)
	}

	// If viewable, charge the advertiser (CPM based)
	if isViewable && !impression.IsViewable {
		// Calculate cost (CPM / 1000 = cost per impression)
		ad, err := s.adRepo.GetAdByID(ctx, impression.AdID)
		if err == nil {
			costCents := ad.CPMCents / 1000 // Cost per single impression
			if costCents < 1 {
				costCents = 1 // Minimum 1 cent per impression
			}
			// Update ad spend (async)
			go func() {
				_ = s.adRepo.IncrementAdSpend(context.Background(), ad.ID, costCents)
			}()
		}
	}

	return nil
}

// filterByTargeting filters ads based on targeting criteria
func (s *AdService) filterByTargeting(ads []models.Ad, req models.AdSelectionRequest) []models.Ad {
	var filtered []models.Ad

	for _, ad := range ads {
		if ad.TargetingCriteria == nil {
			// No targeting = show to everyone
			filtered = append(filtered, ad)
			continue
		}

		match := true

		// Check game targeting
		if targetGames, ok := ad.TargetingCriteria["game_ids"].([]interface{}); ok && len(targetGames) > 0 {
			if req.GameID == nil {
				// Ad targets specific games but request doesn't specify a game
				match = false
			} else {
				gameMatch := false
				for _, g := range targetGames {
					if gID, ok := g.(string); ok && gID == *req.GameID {
						gameMatch = true
						break
					}
				}
				if !gameMatch {
					match = false
				}
			}
		}

		// Check language targeting
		if match {
			if targetLanguages, ok := ad.TargetingCriteria["languages"].([]interface{}); ok && len(targetLanguages) > 0 {
				if req.Language == nil {
					// Ad targets specific languages but request doesn't specify a language
					match = false
				} else {
					langMatch := false
					for _, l := range targetLanguages {
						if lang, ok := l.(string); ok && lang == *req.Language {
							langMatch = true
							break
						}
					}
					if !langMatch {
						match = false
					}
				}
			}
		}

		// Check platform targeting
		if match {
			if targetPlatforms, ok := ad.TargetingCriteria["platforms"].([]interface{}); ok && len(targetPlatforms) > 0 {
				platformMatch := false
				for _, p := range targetPlatforms {
					if platform, ok := p.(string); ok && platform == req.Platform {
						platformMatch = true
						break
					}
				}
				if !platformMatch {
					match = false
				}
			}
		}

		// Check country targeting (enhanced targeting)
		if match {
			if targetCountries, ok := ad.TargetingCriteria["countries"].([]interface{}); ok && len(targetCountries) > 0 {
				if req.Country == nil {
					// Ad targets specific countries but request doesn't specify a country
					match = false
				} else {
					countryMatch := false
					for _, c := range targetCountries {
						if country, ok := c.(string); ok && country == *req.Country {
							countryMatch = true
							break
						}
					}
					if !countryMatch {
						match = false
					}
				}
			}
		}

		// Check device type targeting (enhanced targeting)
		if match {
			if targetDevices, ok := ad.TargetingCriteria["devices"].([]interface{}); ok && len(targetDevices) > 0 {
				if req.DeviceType == nil {
					// Ad targets specific devices but request doesn't specify a device
					match = false
				} else {
					deviceMatch := false
					for _, d := range targetDevices {
						if device, ok := d.(string); ok && device == *req.DeviceType {
							deviceMatch = true
							break
						}
					}
					if !deviceMatch {
						match = false
					}
				}
			}
		}

		// Check interests targeting (enhanced targeting)
		if match {
			if targetInterests, ok := ad.TargetingCriteria["interests"].([]interface{}); ok && len(targetInterests) > 0 {
				if len(req.Interests) == 0 {
					// Ad targets specific interests but request doesn't specify interests
					match = false
				} else {
					// Check if any request interest matches any target interest
					interestMatch := false
					for _, ti := range targetInterests {
						if targetInterest, ok := ti.(string); ok {
							for _, reqInterest := range req.Interests {
								if targetInterest == reqInterest {
									interestMatch = true
									break
								}
							}
							if interestMatch {
								break
							}
						}
					}
					if !interestMatch {
						match = false
					}
				}
			}
		}

		if match {
			filtered = append(filtered, ad)
		}
	}

	return filtered
}

// filterByContextualTargeting filters ads using only contextual (non-user-specific) criteria
// This is used when user has not consented to personalized ads
func (s *AdService) filterByContextualTargeting(ads []models.Ad, req models.AdSelectionRequest) []models.Ad {
	var filtered []models.Ad

	for _, ad := range ads {
		if ad.TargetingCriteria == nil {
			// No targeting = show to everyone
			filtered = append(filtered, ad)
			continue
		}

		match := true

		// Check game targeting (contextual - based on page content, not user)
		if targetGames, ok := ad.TargetingCriteria["game_ids"].([]interface{}); ok && len(targetGames) > 0 {
			if req.GameID == nil {
				match = false
			} else {
				gameMatch := false
				for _, g := range targetGames {
					if gID, ok := g.(string); ok && gID == *req.GameID {
						gameMatch = true
						break
					}
				}
				if !gameMatch {
					match = false
				}
			}
		}

		// Check language targeting (contextual - based on page language)
		if match {
			if targetLanguages, ok := ad.TargetingCriteria["languages"].([]interface{}); ok && len(targetLanguages) > 0 {
				if req.Language == nil {
					match = false
				} else {
					langMatch := false
					for _, l := range targetLanguages {
						if lang, ok := l.(string); ok && lang == *req.Language {
							langMatch = true
							break
						}
					}
					if !langMatch {
						match = false
					}
				}
			}
		}

		// Check platform targeting (contextual - based on device platform)
		if match {
			if targetPlatforms, ok := ad.TargetingCriteria["platforms"].([]interface{}); ok && len(targetPlatforms) > 0 {
				platformMatch := false
				for _, p := range targetPlatforms {
					if platform, ok := p.(string); ok && platform == req.Platform {
						platformMatch = true
						break
					}
				}
				if !platformMatch {
					match = false
				}
			}
		}

		// NOTE: We skip country, device, and interests targeting in contextual mode
		// as these are considered user-specific personalization

		if match {
			filtered = append(filtered, ad)
		}
	}

	return filtered
}

// filterByFrequencyCaps filters ads based on per-user/session frequency caps
func (s *AdService) filterByFrequencyCaps(ctx context.Context, ads []models.Ad, userID *uuid.UUID, sessionID *string) ([]models.Ad, error) {
	if userID == nil && sessionID == nil {
		// No user identification, can't enforce caps
		return ads, nil
	}

	var filtered []models.Ad

	for _, ad := range ads {
		// Get frequency limits for this ad
		limits, err := s.adRepo.GetFrequencyLimits(ctx, ad.ID)
		if err != nil {
			// If we can't get limits, assume no limits
			filtered = append(filtered, ad)
			continue
		}

		if len(limits) == 0 {
			// No limits configured
			filtered = append(filtered, ad)
			continue
		}

		// Check each frequency limit
		capExceeded := false
		for _, limit := range limits {
			count, err := s.adRepo.GetUserImpressionCount(ctx, ad.ID, userID, sessionID, limit.WindowType)
			if err != nil {
				continue
			}
			if count >= limit.MaxImpressions {
				capExceeded = true
				break
			}
		}

		if !capExceeded {
			filtered = append(filtered, ad)
		}
	}

	return filtered, nil
}

// filterByFraudPrevention filters ads to prevent fraud (rapid impressions from same IP)
func (s *AdService) filterByFraudPrevention(ctx context.Context, ads []models.Ad, ipAddress string) ([]models.Ad, error) {
	var filtered []models.Ad

	// Maximum impressions per minute from same IP (basic fraud threshold)
	const maxImpressionsPerMinute = 5

	for _, ad := range ads {
		count, err := s.adRepo.CountRecentImpressions(ctx, ad.ID, ipAddress, 1)
		if err != nil {
			// If we can't check, include the ad
			filtered = append(filtered, ad)
			continue
		}

		if count < maxImpressionsPerMinute {
			filtered = append(filtered, ad)
		}
	}

	return filtered, nil
}

// weightedRandomSelect selects an ad using weighted random selection
func (s *AdService) weightedRandomSelect(ads []models.Ad) models.Ad {
	if len(ads) == 1 {
		return ads[0]
	}

	// Group ads by priority
	highestPriority := ads[0].Priority
	var priorityAds []models.Ad

	for _, ad := range ads {
		if ad.Priority > highestPriority {
			highestPriority = ad.Priority
			priorityAds = []models.Ad{ad}
		} else if ad.Priority == highestPriority {
			priorityAds = append(priorityAds, ad)
		}
	}

	if len(priorityAds) == 1 {
		return priorityAds[0]
	}

	// Calculate total weight
	totalWeight := 0
	for _, ad := range priorityAds {
		totalWeight += ad.Weight
	}

	// Random selection based on weight (using v2 package-level random)
	randomWeight := rand.IntN(totalWeight)

	cumulative := 0
	for _, ad := range priorityAds {
		cumulative += ad.Weight
		if randomWeight < cumulative {
			return ad
		}
	}

	// Fallback to first ad
	return priorityAds[0]
}

// updateFrequencyCaps updates frequency caps for all window types
func (s *AdService) updateFrequencyCaps(ctx context.Context, adID uuid.UUID, userID *uuid.UUID, sessionID *string) {
	if userID == nil && sessionID == nil {
		return
	}

	windowTypes := []string{
		models.FrequencyWindowHourly,
		models.FrequencyWindowDaily,
		models.FrequencyWindowWeekly,
		models.FrequencyWindowLifetime,
	}

	for _, windowType := range windowTypes {
		windowStart := s.calculateWindowStart(windowType)
		_ = s.adRepo.UpsertFrequencyCap(ctx, adID, userID, sessionID, windowType, windowStart)
	}
}

// calculateWindowStart returns the start time for a given window type
func (s *AdService) calculateWindowStart(windowType string) time.Time {
	now := time.Now().UTC()
	switch windowType {
	case models.FrequencyWindowHourly:
		return now.Truncate(time.Hour)
	case models.FrequencyWindowDaily:
		return time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, time.UTC)
	case models.FrequencyWindowWeekly:
		daysSinceSunday := int(now.Weekday())
		return time.Date(now.Year(), now.Month(), now.Day()-daysSinceSunday, 0, 0, 0, 0, time.UTC)
	case models.FrequencyWindowLifetime:
		return time.Time{}
	default:
		return now.Truncate(time.Hour)
	}
}

// GetAdByID retrieves an ad by its ID
func (s *AdService) GetAdByID(ctx context.Context, id uuid.UUID) (*models.Ad, error) {
	return s.adRepo.GetAdByID(ctx, id)
}

// ResetDailySpend resets daily spend for all ads
func (s *AdService) ResetDailySpend(ctx context.Context) error {
	return s.adRepo.ResetDailySpend(ctx)
}

// filterByTargetingRules applies structured targeting rules from the database
func (s *AdService) filterByTargetingRules(ctx context.Context, ads []models.Ad, req models.AdSelectionRequest) ([]models.Ad, error) {
	var filtered []models.Ad

	for _, ad := range ads {
		rules, err := s.adRepo.GetTargetingRules(ctx, ad.ID)
		if err != nil {
			// If we can't get rules, include the ad (fail open)
			filtered = append(filtered, ad)
			continue
		}

		if len(rules) == 0 {
			// No rules = show to everyone
			filtered = append(filtered, ad)
			continue
		}

		match := true
		for _, rule := range rules {
			ruleMatch := s.evaluateTargetingRule(rule, req)
			if rule.Operator == models.TargetingOperatorInclude && !ruleMatch {
				match = false
				break
			}
			if rule.Operator == models.TargetingOperatorExclude && ruleMatch {
				match = false
				break
			}
		}

		if match {
			filtered = append(filtered, ad)
		}
	}

	return filtered, nil
}

// filterByContextualTargetingRules applies only contextual (non-user-specific) targeting rules
// Used when user has not consented to personalized ads
func (s *AdService) filterByContextualTargetingRules(ctx context.Context, ads []models.Ad, req models.AdSelectionRequest) ([]models.Ad, error) {
	var filtered []models.Ad

	// Contextual rule types (non-user-specific)
	contextualRuleTypes := map[string]bool{
		models.TargetingRuleTypePlatform: true,
		models.TargetingRuleTypeLanguage: true,
		models.TargetingRuleTypeGame:     true,
	}

	for _, ad := range ads {
		rules, err := s.adRepo.GetTargetingRules(ctx, ad.ID)
		if err != nil {
			// If we can't get rules, include the ad (fail open)
			filtered = append(filtered, ad)
			continue
		}

		if len(rules) == 0 {
			// No rules = show to everyone
			filtered = append(filtered, ad)
			continue
		}

		match := true
		hasContextualRules := false
		for _, rule := range rules {
			// Skip non-contextual rules (user-specific targeting)
			if !contextualRuleTypes[rule.RuleType] {
				continue
			}
			hasContextualRules = true

			ruleMatch := s.evaluateTargetingRule(rule, req)
			if rule.Operator == models.TargetingOperatorInclude && !ruleMatch {
				match = false
				break
			}
			if rule.Operator == models.TargetingOperatorExclude && ruleMatch {
				match = false
				break
			}
		}

		// If ad has rules but none are contextual, exclude it in contextual mode
		if len(rules) > 0 && !hasContextualRules {
			match = false
		}

		if match {
			filtered = append(filtered, ad)
		}
	}

	return filtered, nil
}

// evaluateTargetingRule checks if a single targeting rule matches the request
func (s *AdService) evaluateTargetingRule(rule models.AdTargetingRule, req models.AdSelectionRequest) bool {
	switch rule.RuleType {
	case models.TargetingRuleTypeCountry:
		if req.Country == nil {
			return false
		}
		return containsString(rule.Values, *req.Country)
	case models.TargetingRuleTypeDevice:
		if req.DeviceType == nil {
			return false
		}
		return containsString(rule.Values, *req.DeviceType)
	case models.TargetingRuleTypeInterest:
		if len(req.Interests) == 0 {
			return false
		}
		// Check if any interest matches
		for _, interest := range req.Interests {
			if containsString(rule.Values, interest) {
				return true
			}
		}
		return false
	case models.TargetingRuleTypePlatform:
		return containsString(rule.Values, req.Platform)
	case models.TargetingRuleTypeLanguage:
		if req.Language == nil {
			return false
		}
		return containsString(rule.Values, *req.Language)
	case models.TargetingRuleTypeGame:
		if req.GameID == nil {
			return false
		}
		return containsString(rule.Values, *req.GameID)
	default:
		return true // Unknown rule types are ignored
	}
}

// containsString checks if a string slice contains a value
func containsString(slice []string, value string) bool {
	for _, item := range slice {
		if item == value {
			return true
		}
	}
	return false
}

// selectAdWithExperiment selects an ad considering A/B experiments
func (s *AdService) selectAdWithExperiment(ads []models.Ad, userID *uuid.UUID, sessionID *string) models.Ad {
	// Group ads by experiment
	experimentAds := make(map[uuid.UUID][]models.Ad)
	nonExperimentAds := []models.Ad{}

	for _, ad := range ads {
		if ad.ExperimentID != nil {
			experimentAds[*ad.ExperimentID] = append(experimentAds[*ad.ExperimentID], ad)
		} else {
			nonExperimentAds = append(nonExperimentAds, ad)
		}
	}

	// Sort experiment IDs for deterministic iteration order
	experimentIDs := make([]uuid.UUID, 0, len(experimentAds))
	for expID := range experimentAds {
		experimentIDs = append(experimentIDs, expID)
	}
	sort.Slice(experimentIDs, func(i, j int) bool {
		return experimentIDs[i].String() < experimentIDs[j].String()
	})

	// If there are experiment ads, try to select from them based on consistent bucketing
	for _, experimentID := range experimentIDs {
		expAds := experimentAds[experimentID]
		if len(expAds) > 1 {
			// Use consistent bucketing based on user/session ID
			selectedVariant := s.selectExperimentVariant(expAds, userID, sessionID, experimentID)
			return selectedVariant
		} else if len(expAds) == 1 {
			return expAds[0]
		}
	}

	// Fall back to weighted random selection from non-experiment ads
	if len(nonExperimentAds) > 0 {
		return s.weightedRandomSelect(nonExperimentAds)
	}

	// If somehow we have no ads, return the first one
	return ads[0]
}

// selectExperimentVariant selects a variant for an experiment using consistent bucketing
func (s *AdService) selectExperimentVariant(ads []models.Ad, userID *uuid.UUID, sessionID *string, experimentID uuid.UUID) models.Ad {
	// Create a deterministic bucket key from user/session ID and experiment ID
	var bucketKey string
	if userID != nil {
		bucketKey = userID.String() + experimentID.String()
	} else if sessionID != nil {
		bucketKey = *sessionID + experimentID.String()
	} else {
		// No stable identifier, use random selection
		return s.weightedRandomSelect(ads)
	}

	// Use FNV hash for better distribution in A/B experiment bucketing
	h := fnv.New32a()
	h.Write([]byte(bucketKey))
	hash := int(h.Sum32()) % experimentBucketCount

	// Group ads by variant
	variantAds := make(map[string][]models.Ad)
	for _, ad := range ads {
		variant := "default"
		if ad.ExperimentVariant != nil {
			variant = *ad.ExperimentVariant
		}
		variantAds[variant] = append(variantAds[variant], ad)
	}

	// Distribute bucket across variants evenly with deterministic ordering
	variants := make([]string, 0, len(variantAds))
	for v := range variantAds {
		variants = append(variants, v)
	}
	sort.Strings(variants) // Ensure deterministic ordering

	if len(variants) == 0 {
		return ads[0]
	}

	selectedVariant := variants[hash%len(variants)]
	variantGroup := variantAds[selectedVariant]

	if len(variantGroup) == 1 {
		return variantGroup[0]
	}

	return s.weightedRandomSelect(variantGroup)
}

// GetCTRReportByCampaign retrieves CTR report grouped by campaign
func (s *AdService) GetCTRReportByCampaign(ctx context.Context, since time.Time) ([]models.AdCTRReport, error) {
	return s.adRepo.GetCTRReportByCampaign(ctx, since)
}

// GetCTRReportBySlot retrieves CTR report grouped by ad slot
func (s *AdService) GetCTRReportBySlot(ctx context.Context, since time.Time) ([]models.AdSlotReport, error) {
	return s.adRepo.GetCTRReportBySlot(ctx, since)
}

// GetExperimentReport retrieves analytics for an experiment
func (s *AdService) GetExperimentReport(ctx context.Context, experimentID uuid.UUID, since time.Time) (*models.AdExperimentReport, error) {
	return s.adRepo.GetExperimentReport(ctx, experimentID, since)
}

// GetRunningExperiments retrieves all running experiments
func (s *AdService) GetRunningExperiments(ctx context.Context) ([]models.AdExperiment, error) {
	return s.adRepo.GetRunningExperiments(ctx)
}

// Campaign CRUD methods

// ListCampaigns retrieves all campaigns with optional filtering
func (s *AdService) ListCampaigns(ctx context.Context, page, limit int, status *string) ([]models.Ad, int, error) {
	return s.adRepo.ListCampaigns(ctx, page, limit, status)
}

// ValidateCampaign validates campaign data without creating it
func (s *AdService) ValidateCampaign(ad *models.Ad) error {
	// Validate required fields
	if ad.Name == "" {
		return fmt.Errorf("campaign name is required")
	}
	if ad.AdvertiserName == "" {
		return fmt.Errorf("advertiser name is required")
	}
	if ad.AdType == "" {
		return fmt.Errorf("ad type is required")
	}
	if ad.ContentURL == "" {
		return fmt.Errorf("content URL is required")
	}

	// Validate ad type
	validAdTypes := map[string]bool{"banner": true, "video": true, "native": true}
	if !validAdTypes[ad.AdType] {
		return fmt.Errorf("invalid ad type: must be banner, video, or native")
	}

	// Validate dates if both provided
	if ad.StartDate != nil && ad.EndDate != nil && ad.EndDate.Before(*ad.StartDate) {
		return fmt.Errorf("end date must be after start date")
	}

	return nil
}

// CreateCampaign creates a new ad campaign with validation
func (s *AdService) CreateCampaign(ctx context.Context, ad *models.Ad) error {
	// Validate campaign data
	if err := s.ValidateCampaign(ad); err != nil {
		return err
	}

	// Set defaults
	if ad.Priority == 0 {
		ad.Priority = 1
	}
	if ad.Weight == 0 {
		ad.Weight = 100
	}
	if ad.CPMCents == 0 {
		ad.CPMCents = 100 // Default $1 CPM
	}

	return s.adRepo.CreateCampaign(ctx, ad)
}

// UpdateCampaign updates an existing campaign
func (s *AdService) UpdateCampaign(ctx context.Context, ad *models.Ad) error {
	// Verify campaign exists
	existing, err := s.adRepo.GetAdByID(ctx, ad.ID)
	if err != nil {
		return fmt.Errorf("campaign not found")
	}

	// Validate ad type if changed
	if ad.AdType != "" {
		validAdTypes := map[string]bool{"banner": true, "video": true, "native": true}
		if !validAdTypes[ad.AdType] {
			return fmt.Errorf("invalid ad type: must be banner, video, or native")
		}
	} else {
		ad.AdType = existing.AdType
	}

	// Validate dates if both provided
	if ad.StartDate != nil && ad.EndDate != nil && ad.EndDate.Before(*ad.StartDate) {
		return fmt.Errorf("end date must be after start date")
	}

	return s.adRepo.UpdateCampaign(ctx, ad)
}

// DeleteCampaign deletes a campaign by ID
func (s *AdService) DeleteCampaign(ctx context.Context, id uuid.UUID) error {
	// Verify campaign exists
	_, err := s.adRepo.GetAdByID(ctx, id)
	if err != nil {
		return fmt.Errorf("campaign not found")
	}

	return s.adRepo.DeleteCampaign(ctx, id)
}

// GetCampaignReportByDate retrieves campaign performance report by date range
func (s *AdService) GetCampaignReportByDate(ctx context.Context, adID *uuid.UUID, startDate, endDate time.Time) ([]models.AdCampaignAnalytics, error) {
	return s.adRepo.GetCampaignReportByDate(ctx, adID, startDate, endDate)
}

// GetCampaignReportByPlacement retrieves campaign performance report grouped by placement/slot
func (s *AdService) GetCampaignReportByPlacement(ctx context.Context, adID *uuid.UUID, since time.Time) ([]models.AdSlotReport, error) {
	return s.adRepo.GetCampaignReportByPlacement(ctx, adID, since)
}

// ValidateCreative validates an ad creative URL and dimensions
func (s *AdService) ValidateCreative(ctx context.Context, contentURL string, adType string, width, height *int) error {
	// Validate content URL is not empty
	if contentURL == "" {
		return fmt.Errorf("content URL is required")
	}

	// Validate ad type
	validAdTypes := map[string]bool{"banner": true, "video": true, "native": true}
	if !validAdTypes[adType] {
		return fmt.Errorf("invalid ad type: must be banner, video, or native")
	}

	// Validate dimensions for banner ads
	if adType == "banner" {
		if width == nil || height == nil {
			return fmt.Errorf("width and height are required for banner ads")
		}
		// Standard banner sizes (IAB)
		validSizes := map[string]bool{
			"728x90":  true, // Leaderboard
			"300x250": true, // Medium Rectangle
			"336x280": true, // Large Rectangle
			"300x600": true, // Half Page
			"970x250": true, // Billboard
			"320x50":  true, // Mobile Leaderboard
			"160x600": true, // Wide Skyscraper
			"300x50":  true, // Mobile Banner
			"970x90":  true, // Large Leaderboard
			"250x250": true, // Square
			"200x200": true, // Small Square
		}
		sizeKey := fmt.Sprintf("%dx%d", *width, *height)
		if !validSizes[sizeKey] {
			return fmt.Errorf("invalid banner size: %s. Supported sizes: 728x90, 300x250, 336x280, 300x600, 970x250, 320x50, 160x600, 300x50, 970x90, 250x250, 200x200", sizeKey)
		}
	}

	return nil
}
