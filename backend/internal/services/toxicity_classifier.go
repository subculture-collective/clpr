package services

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"sync"
	"time"
	"unicode"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
	"gopkg.in/yaml.v3"
)

// Category represents a toxicity category
type Category string

const (
	CategoryHateSpeech    Category = "hate_speech"
	CategoryProfanity     Category = "profanity"
	CategoryHarassment    Category = "harassment"
	CategorySexualContent Category = "sexual_content"
	CategoryViolence      Category = "violence"
	CategorySpam          Category = "spam"
)

// Severity represents the severity level of a rule
type Severity string

const (
	SeverityLow    Severity = "low"
	SeverityMedium Severity = "medium"
	SeverityHigh   Severity = "high"
)

// Rule represents a toxicity detection rule
type Rule struct {
	Pattern     string   `yaml:"pattern"`
	Category    Category `yaml:"category"`
	Severity    Severity `yaml:"severity"`
	Weight      float64  `yaml:"weight"`
	Description string   `yaml:"description"`
	compiled    *regexp.Regexp
}

// RulesConfig represents the toxicity rules configuration
type RulesConfig struct {
	Rules     []Rule   `yaml:"rules"`
	Whitelist []string `yaml:"whitelist"`
}

// ToxicityClassifier handles toxicity detection for comments
type ToxicityClassifier struct {
	httpClient     *http.Client
	apiKey         string
	apiURL         string
	enabled        bool
	threshold      float64
	db             *pgxpool.Pool
	rulesConfig    *RulesConfig
	whitelistMap   map[string]bool
	configLoadOnce sync.Once
	configLoadErr  error // Stores any error from loading config
}

// ToxicityScore represents the result of toxicity classification
type ToxicityScore struct {
	Toxic           bool               `json:"toxic"`
	ConfidenceScore float64            `json:"confidence_score"`
	Categories      map[string]float64 `json:"categories"`
	ReasonCodes     []string           `json:"reason_codes"`
}

// PerspectiveAPIRequest represents a request to the Perspective API
type PerspectiveAPIRequest struct {
	Comment             CommentText            `json:"comment"`
	RequestedAttributes map[string]interface{} `json:"requestedAttributes"`
	Languages           []string               `json:"languages"`
}

// CommentText represents the comment text for Perspective API
type CommentText struct {
	Text string `json:"text"`
}

// PerspectiveAPIResponse represents a response from the Perspective API
type PerspectiveAPIResponse struct {
	AttributeScores map[string]AttributeScore `json:"attributeScores"`
	Languages       []string                  `json:"languages"`
}

// AttributeScore represents the score for a specific attribute
type AttributeScore struct {
	SummaryScore SummaryScore `json:"summaryScore"`
}

// SummaryScore represents the summary score
type SummaryScore struct {
	Value float64 `json:"value"`
	Type  string  `json:"type"`
}

// NewToxicityClassifier creates a new ToxicityClassifier
func NewToxicityClassifier(apiKey, apiURL string, enabled bool, threshold float64, db *pgxpool.Pool) *ToxicityClassifier {
	tc := &ToxicityClassifier{
		httpClient: &http.Client{
			Timeout: 10 * time.Second,
		},
		apiKey:       apiKey,
		apiURL:       apiURL,
		enabled:      enabled,
		threshold:    threshold,
		db:           db,
		whitelistMap: make(map[string]bool),
	}
	// Load rules config on first use
	return tc
}

// ClassifyComment analyzes a comment for toxicity
func (tc *ToxicityClassifier) ClassifyComment(ctx context.Context, content string) (*ToxicityScore, error) {
	// If classifier is disabled, return safe default
	if !tc.enabled {
		return &ToxicityScore{
			Toxic:           false,
			ConfidenceScore: 0.0,
			Categories:      make(map[string]float64),
			ReasonCodes:     []string{},
		}, nil
	}

	// Use Perspective API if configured
	if tc.apiKey != "" && tc.apiURL != "" {
		return tc.classifyWithPerspectiveAPI(ctx, content)
	}

	// Fallback to rule-based classification
	return tc.classifyWithRules(content)
}

// classifyWithPerspectiveAPI uses Google's Perspective API for toxicity detection
func (tc *ToxicityClassifier) classifyWithPerspectiveAPI(ctx context.Context, content string) (*ToxicityScore, error) {
	// Construct Perspective API request
	reqBody := PerspectiveAPIRequest{
		Comment: CommentText{
			Text: content,
		},
		RequestedAttributes: map[string]interface{}{
			"TOXICITY":          struct{}{},
			"SEVERE_TOXICITY":   struct{}{},
			"IDENTITY_ATTACK":   struct{}{},
			"INSULT":            struct{}{},
			"PROFANITY":         struct{}{},
			"THREAT":            struct{}{},
			"SEXUALLY_EXPLICIT": struct{}{},
		},
		Languages: []string{"en"},
	}

	jsonData, err := json.Marshal(reqBody)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request: %w", err)
	}

	// Make API request
	// Note: API key is passed via header to avoid exposure in logs
	req, err := http.NewRequestWithContext(ctx, "POST", tc.apiURL, bytes.NewBuffer(jsonData))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-Goog-Api-Key", tc.apiKey)

	resp, err := tc.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to make API request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("API request failed with status %d: %s", resp.StatusCode, string(body))
	}

	// Parse response
	var apiResp PerspectiveAPIResponse
	if err := json.NewDecoder(resp.Body).Decode(&apiResp); err != nil {
		return nil, fmt.Errorf("failed to decode API response: %w", err)
	}

	// Extract scores and determine toxicity
	categories := make(map[string]float64)
	reasonCodes := []string{}
	maxScore := 0.0

	for attr, score := range apiResp.AttributeScores {
		value := score.SummaryScore.Value
		categories[attr] = value

		if value > maxScore {
			maxScore = value
		}

		// Add to reason codes if above threshold
		if value >= tc.threshold {
			reasonCodes = append(reasonCodes, attr)
		}
	}

	return &ToxicityScore{
		Toxic:           maxScore >= tc.threshold,
		ConfidenceScore: maxScore,
		Categories:      categories,
		ReasonCodes:     reasonCodes,
	}, nil
}

// classifyWithRules uses rule-based classification as fallback
func (tc *ToxicityClassifier) classifyWithRules(content string) (*ToxicityScore, error) {
	// Load rules config on first use
	tc.configLoadOnce.Do(func() {
		tc.configLoadErr = tc.loadRulesConfig()
	})

	// If no rules loaded, return safe default
	if tc.rulesConfig == nil {
		// For testing/debugging - rules config failed to load but we continue safely
		return &ToxicityScore{
			Toxic:           false,
			ConfidenceScore: 0.0,
			Categories:      map[string]float64{},
			ReasonCodes:     []string{},
		}, nil
	}

	// Check whitelist first
	if tc.isWhitelisted(content) {
		return &ToxicityScore{
			Toxic:           false,
			ConfidenceScore: 0.0,
			Categories:      map[string]float64{},
			ReasonCodes:     []string{},
		}, nil
	}

	// Normalize text for better pattern matching
	normalizedContent := tc.normalizeText(content)

	// Apply context awareness - reduce scoring for quoted text, URLs, etc.
	contextMultiplier := tc.calculateContextMultiplier(content)

	// Track scores by category
	categoryScores := make(map[Category]float64)

	// Apply all rules
	for _, rule := range tc.rulesConfig.Rules {
		if rule.compiled == nil {
			continue
		}

		// Check if pattern matches
		if rule.compiled.MatchString(normalizedContent) {
			// Add weighted score for this category
			categoryScores[rule.Category] += rule.Weight
		}
	}

	// Apply context multiplier to all category scores
	for category := range categoryScores {
		categoryScores[category] *= contextMultiplier
	}

	// Calculate overall toxicity score (max category score)
	maxScore := 0.0
	reasonCodes := []string{}
	categories := make(map[string]float64)

	for category, score := range categoryScores {
		// Normalize category score to 0-1 range
		normalizedScore := score
		if normalizedScore > 1.0 {
			normalizedScore = 1.0
		}

		categories[string(category)] = normalizedScore

		if normalizedScore > maxScore {
			maxScore = normalizedScore
		}

		// Add to reason codes if above threshold
		if normalizedScore >= tc.threshold {
			reasonCodes = append(reasonCodes, string(category))
		}
	}

	return &ToxicityScore{
		Toxic:           maxScore >= tc.threshold,
		ConfidenceScore: maxScore,
		Categories:      categories,
		ReasonCodes:     reasonCodes,
	}, nil
}

// loadRulesConfig loads the toxicity rules configuration
func (tc *ToxicityClassifier) loadRulesConfig() error {
	// First, allow an explicit override via environment variable
	envPath := os.Getenv("TOXICITY_RULES_CONFIG_PATH")

	var possiblePaths []string
	if envPath != "" {
		if filepath.IsAbs(envPath) || strings.HasPrefix(filepath.Clean(envPath), ".."+string(filepath.Separator)) {
			return fmt.Errorf("TOXICITY_RULES_CONFIG_PATH must be a repository-relative path")
		}
		possiblePaths = append(possiblePaths, envPath)
	}

	possiblePaths = append(possiblePaths,
		"config/toxicity_rules.yaml",         // From backend directory
		"backend/config/toxicity_rules.yaml", // From repository root
		"../../config/toxicity_rules.yaml",   // From test files in internal/services
	)

	var configPath string
	var found bool
	for _, path := range possiblePaths {
		cleanPath := filepath.Clean(path)
		// Paths are fixed defaults or a validated repository-relative operator
		// setting; HTTP/user input cannot reach this startup-only loader.
		if _, err := os.Stat(cleanPath); err == nil { // #nosec G304 G703 -- validated operator-controlled path.
			configPath = cleanPath
			found = true
			break
		}
	}

	if !found {
		// Config file not found - continue with empty rules
		// This is not a fatal error - allows classifier to work without rules
		// In production, rules should be loaded from a database or external service
		return fmt.Errorf("failed to find rules config in any expected location")
	}

	data, err := os.ReadFile(configPath) // #nosec G304 G703 -- path selected from validated startup configuration.
	if err != nil {
		return fmt.Errorf("failed to load rules config from %s: %w", configPath, err)
	}

	var config RulesConfig
	if err := yaml.Unmarshal(data, &config); err != nil {
		return fmt.Errorf("failed to parse rules config: %w", err)
	}

	// Validate that we got some rules
	if len(config.Rules) == 0 {
		return fmt.Errorf("no rules found in config file")
	}

	// Compile regex patterns
	for i := range config.Rules {
		pattern, err := regexp.Compile("(?i)" + config.Rules[i].Pattern)
		if err != nil {
			return fmt.Errorf("failed to compile pattern '%s': %w", config.Rules[i].Pattern, err)
		}
		config.Rules[i].compiled = pattern
	}

	// Build whitelist map
	tc.whitelistMap = make(map[string]bool)
	for _, word := range config.Whitelist {
		tc.whitelistMap[strings.ToLower(word)] = true
	}

	tc.rulesConfig = &config
	return nil
}

// GetRulesCount returns the number of loaded rules (for testing)
func (tc *ToxicityClassifier) GetRulesCount() int {
	// Force config loading
	tc.configLoadOnce.Do(func() {
		tc.configLoadErr = tc.loadRulesConfig()
	})

	if tc.rulesConfig == nil {
		return 0
	}
	return len(tc.rulesConfig.Rules)
}

// GetConfigLoadError returns any error from loading the config (for testing)
func (tc *ToxicityClassifier) GetConfigLoadError() error {
	// Force config loading
	tc.configLoadOnce.Do(func() {
		tc.configLoadErr = tc.loadRulesConfig()
	})

	return tc.configLoadErr
}

// normalizeText normalizes text for better pattern matching
// Handles l33tspeak, obfuscation, and common substitutions
func (tc *ToxicityClassifier) normalizeText(text string) string {
	text = strings.ToLower(text)

	// Common l33tspeak and obfuscation substitutions
	// Order matters - do letter replacements before removing symbols
	replacements := []struct{ old, new string }{
		{"@", "a"},
		{"4", "a"},
		{"3", "e"},
		{"1", "i"},
		{"!", "i"},
		{"0", "o"},
		{"$", "s"},
		{"5", "s"},
		{"7", "t"},
		{"+", "t"},
		{"*", "u"}, // Asterisk often used to censor 'u' (f*ck -> fuck)
		{"_", ""},  // Remove underscores
		{"-", ""},  // Remove dashes
		{"..", ""}, // Remove double dots
	}

	result := text
	for _, r := range replacements {
		result = strings.ReplaceAll(result, r.old, r.new)
	}

	// Normalize repeated characters (e.g., "asssss" -> "ass")
	normalized := ""
	lastChar := rune(0)
	repeatCount := 0

	for _, char := range result {
		if char == lastChar {
			repeatCount++
			// Allow up to 2 repetitions
			if repeatCount < 3 {
				normalized += string(char)
			}
		} else {
			normalized += string(char)
			lastChar = char
			repeatCount = 1
		}
	}

	// Remove extra whitespace
	return strings.Join(strings.Fields(normalized), " ")
}

// isWhitelisted checks if content contains only whitelisted terms
func (tc *ToxicityClassifier) isWhitelisted(content string) bool {
	words := strings.Fields(strings.ToLower(content))

	if len(words) == 0 {
		return false
	}

	// Check if ALL words are whitelisted (not just any one word)
	for _, word := range words {
		// Clean the word of punctuation
		cleaned := strings.TrimFunc(word, func(r rune) bool {
			return !unicode.IsLetter(r) && !unicode.IsNumber(r)
		})

		// Skip empty tokens
		if cleaned == "" {
			continue
		}

		// If any word is NOT whitelisted, content is not fully whitelisted
		if !tc.whitelistMap[cleaned] {
			return false
		}
	}

	// All non-empty words are whitelisted
	return true
}

// calculateContextMultiplier determines a multiplier based on context
// Reduces false positives by detecting quoted text, URLs, code, etc.
func (tc *ToxicityClassifier) calculateContextMultiplier(content string) float64 {
	multiplier := 1.0

	// Check if content is very short (might be part of a larger context)
	if len(strings.TrimSpace(content)) < 10 {
		multiplier *= 0.8
	}

	// Reduce score if content appears to be quoted (surrounded by quotes)
	if (strings.HasPrefix(content, "\"") && strings.HasSuffix(content, "\"")) ||
		(strings.HasPrefix(content, "'") && strings.HasSuffix(content, "'")) {
		multiplier *= 0.5
	}

	// Reduce score if content contains code-like patterns
	if strings.Contains(content, "```") || strings.Contains(content, "function") ||
		strings.Contains(content, "class ") || strings.Contains(content, "def ") {
		multiplier *= 0.6
	}

	// Reduce score if content contains URLs (might be sharing links)
	if strings.Contains(content, "http://") || strings.Contains(content, "https://") ||
		strings.Contains(content, "www.") {
		multiplier *= 0.7
	}

	// Reduce score if content contains @mentions (might be quoting someone)
	if strings.Contains(content, "@") && strings.Contains(content, " ") {
		multiplier *= 0.8
	}

	return multiplier
}

// AddToModerationQueue adds a toxic comment to the moderation queue
func (tc *ToxicityClassifier) AddToModerationQueue(ctx context.Context, commentID uuid.UUID, score *ToxicityScore) error {
	if tc.db == nil {
		return fmt.Errorf("database connection not available")
	}

	// Determine reason based on the highest scoring category
	reason := "toxic"
	if len(score.ReasonCodes) > 0 {
		// Map Perspective API categories to our reason codes
		categoryMap := map[string]string{
			"TOXICITY":          "toxic",
			"SEVERE_TOXICITY":   "toxic",
			"IDENTITY_ATTACK":   "harassment",
			"INSULT":            "offensive",
			"PROFANITY":         "offensive",
			"THREAT":            "harassment",
			"SEXUALLY_EXPLICIT": "inappropriate",
		}

		if mappedReason, ok := categoryMap[score.ReasonCodes[0]]; ok {
			reason = mappedReason
		}
	}

	// Calculate priority based on confidence score
	// Higher confidence = higher priority (0-100 scale)
	priority := int(score.ConfidenceScore * 100)
	if priority > 100 {
		priority = 100
	} else if priority < 50 {
		priority = 50 // Minimum priority for auto-flagged items
	}

	// Insert into moderation queue
	_, err := tc.db.Exec(ctx, `
		INSERT INTO moderation_queue 
		(content_type, content_id, reason, priority, status, auto_flagged, confidence_score)
		VALUES ($1, $2, $3, $4, 'pending', true, $5)
		ON CONFLICT (content_type, content_id) WHERE status = 'pending'
		DO UPDATE SET 
			confidence_score = EXCLUDED.confidence_score,
			priority = EXCLUDED.priority,
			reason = EXCLUDED.reason
	`, "comment", commentID, reason, priority, score.ConfidenceScore)

	return err
}

// RecordPrediction records a toxicity prediction for metrics tracking
func (tc *ToxicityClassifier) RecordPrediction(ctx context.Context, commentID uuid.UUID, score *ToxicityScore) error {
	if tc.db == nil {
		return fmt.Errorf("database connection not available")
	}

	// Store prediction in a metrics table for tracking precision/recall
	categoriesJSON, err := json.Marshal(score.Categories)
	if err != nil {
		return fmt.Errorf("failed to marshal categories: %w", err)
	}

	_, err = tc.db.Exec(ctx, `
		INSERT INTO toxicity_predictions 
		(comment_id, toxic, confidence_score, categories, reason_codes, created_at)
		VALUES ($1, $2, $3, $4, $5, NOW())
		ON CONFLICT (comment_id) 
		DO UPDATE SET 
			toxic = EXCLUDED.toxic,
			confidence_score = EXCLUDED.confidence_score,
			categories = EXCLUDED.categories,
			reason_codes = EXCLUDED.reason_codes,
			updated_at = NOW()
	`, commentID, score.Toxic, score.ConfidenceScore, categoriesJSON, score.ReasonCodes)

	return err
}

// GetMetrics retrieves toxicity classification metrics
func (tc *ToxicityClassifier) GetMetrics(ctx context.Context, startDate, endDate time.Time) (map[string]interface{}, error) {
	if tc.db == nil {
		return nil, fmt.Errorf("database connection not available")
	}

	metrics := make(map[string]interface{})

	// Total predictions
	var totalPredictions int
	err := tc.db.QueryRow(ctx, `
		SELECT COUNT(*) 
		FROM toxicity_predictions 
		WHERE created_at >= $1 AND created_at <= $2
	`, startDate, endDate).Scan(&totalPredictions)
	if err != nil {
		return nil, fmt.Errorf("failed to get total predictions: %w", err)
	}
	metrics["total_predictions"] = totalPredictions

	// Total flagged as toxic
	var totalToxic int
	err = tc.db.QueryRow(ctx, `
		SELECT COUNT(*) 
		FROM toxicity_predictions 
		WHERE toxic = true AND created_at >= $1 AND created_at <= $2
	`, startDate, endDate).Scan(&totalToxic)
	if err != nil {
		return nil, fmt.Errorf("failed to get total toxic: %w", err)
	}
	metrics["total_toxic"] = totalToxic
	metrics["toxicity_rate"] = 0.0
	if totalPredictions > 0 {
		metrics["toxicity_rate"] = float64(totalToxic) / float64(totalPredictions)
	}

	// Get precision/recall if we have human review data
	// This would require joining with moderation decisions
	var tp, fp, fn int
	var tn int
	err = tc.db.QueryRow(ctx, `
		SELECT 
			COUNT(*) FILTER (WHERE tp.toxic = true AND md.action = 'reject') as true_positives,
			COUNT(*) FILTER (WHERE tp.toxic = true AND md.action = 'approve') as false_positives,
			COUNT(*) FILTER (WHERE tp.toxic = false AND md.action = 'reject') as false_negatives,
			COUNT(*) FILTER (WHERE tp.toxic = false AND md.action = 'approve') as true_negatives
		FROM toxicity_predictions tp
		LEFT JOIN moderation_queue mq ON tp.comment_id = mq.content_id AND mq.content_type = 'comment'
		LEFT JOIN moderation_decisions md ON mq.id = md.queue_item_id
		WHERE tp.created_at >= $1 AND tp.created_at <= $2
			AND md.created_at IS NOT NULL
	`, startDate, endDate).Scan(&tp, &fp, &fn, &tn)

	if err == nil {
		precision := 0.0
		recall := 0.0
		fpr := 0.0

		if tp+fp > 0 {
			precision = float64(tp) / float64(tp+fp)
		}
		if tp+fn > 0 {
			recall = float64(tp) / float64(tp+fn)
		}
		// FPR = FP / (FP + TN)
		if fp+tn > 0 {
			fpr = float64(fp) / float64(fp+tn)
		}

		metrics["precision"] = precision
		metrics["recall"] = recall
		metrics["false_positive_rate"] = fpr
		metrics["true_positives"] = tp
		metrics["false_positives"] = fp
		metrics["false_negatives"] = fn
		metrics["true_negatives"] = tn
	}

	return metrics, nil
}
