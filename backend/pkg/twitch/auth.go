package twitch

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"sync"
	"time"

	"git.subcult.tv/subculture-collective/clpr/pkg/utils"
)

const (
	tokenURL = "https://id.twitch.tv/oauth2/token" // #nosec G101 -- not a credential, just OAuth endpoint URL
)

// AuthManager handles OAuth token management for Twitch API
type AuthManager struct {
	clientID     string
	clientSecret string
	httpClient   *http.Client
	cache        Cache
	mu           sync.RWMutex
	accessToken  string
	tokenExpiry  time.Time
}

// NewAuthManager creates a new authentication manager
func NewAuthManager(clientID, clientSecret string, httpClient *http.Client, cache Cache) *AuthManager {
	return &AuthManager{
		clientID:     clientID,
		clientSecret: clientSecret,
		httpClient:   httpClient,
		cache:        cache,
	}
}

// GetToken returns the current access token, refreshing if necessary
func (am *AuthManager) GetToken(ctx context.Context) (string, error) {
	am.mu.RLock()
	needsRefresh := time.Now().After(am.tokenExpiry)
	token := am.accessToken
	am.mu.RUnlock()

	if needsRefresh {
		if err := am.RefreshToken(ctx); err != nil {
			return "", err
		}
		am.mu.RLock()
		token = am.accessToken
		am.mu.RUnlock()
	}

	return token, nil
}

// LoadFromCache loads the access token from cache
func (am *AuthManager) LoadFromCache(ctx context.Context) error {
	token, ok := am.cache.Get(cacheKeyPrefix + "access_token")
	if !ok {
		return fmt.Errorf("token not found in cache")
	}

	expiryVal, ok := am.cache.Get(cacheKeyPrefix + "token_expiry")
	if !ok {
		return fmt.Errorf("token expiry not found in cache")
	}

	expiryStr, ok := expiryVal.(string)
	if !ok {
		return fmt.Errorf("invalid token expiry format in cache")
	}

	expiry, err := time.Parse(time.RFC3339, expiryStr)
	if err != nil {
		return fmt.Errorf("failed to parse token expiry: %w", err)
	}

	tokenStr, ok := token.(string)
	if !ok {
		return fmt.Errorf("invalid token format in cache")
	}

	am.mu.Lock()
	am.accessToken = tokenStr
	am.tokenExpiry = expiry
	am.mu.Unlock()

	return nil
}

// SaveToCache saves the access token to cache
func (am *AuthManager) SaveToCache(ctx context.Context) error {
	am.mu.RLock()
	defer am.mu.RUnlock()

	ttl := time.Until(am.tokenExpiry)
	if ttl <= 0 {
		return nil
	}

	am.cache.Set(cacheKeyPrefix+"access_token", am.accessToken, ttl)
	am.cache.Set(cacheKeyPrefix+"token_expiry", am.tokenExpiry.Format(time.RFC3339), ttl)

	return nil
}

// RefreshToken obtains a new app access token
func (am *AuthManager) RefreshToken(ctx context.Context) error {
	data := url.Values{}
	data.Set("client_id", am.clientID)
	data.Set("client_secret", am.clientSecret)
	data.Set("grant_type", "client_credentials")

	req, err := http.NewRequestWithContext(ctx, "POST", tokenURL, http.NoBody)
	if err != nil {
		return &AuthError{Message: "failed to create token request", Err: err}
	}

	req.URL.RawQuery = data.Encode()
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	resp, err := am.httpClient.Do(req)
	if err != nil {
		return &AuthError{Message: "failed to request token", Err: err}
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		const maxErrBody = 4096 // 4KB
		body, _ := io.ReadAll(io.LimitReader(resp.Body, maxErrBody))
		msg := fmt.Sprintf("token request failed with status %d: %s", resp.StatusCode, string(body))
		if len(body) == maxErrBody {
			msg += " [truncated]"
		}
		return &AuthError{
			Message: msg,
		}
	}

	var tokenResp struct {
		AccessToken string `json:"access_token"`
		ExpiresIn   int    `json:"expires_in"`
		TokenType   string `json:"token_type"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&tokenResp); err != nil {
		return &AuthError{Message: "failed to decode token response", Err: err}
	}

	am.mu.Lock()
	am.accessToken = tokenResp.AccessToken
	// Refresh 5 minutes before actual expiry to be safe
	am.tokenExpiry = time.Now().Add(time.Duration(tokenResp.ExpiresIn-300) * time.Second)
	am.mu.Unlock()

	// Save to cache
	if err := am.SaveToCache(ctx); err != nil {
		logger := utils.GetLogger()
		logger.Warn("Failed to cache token", map[string]interface{}{
			"error": err.Error(),
		})
	}

	logger := utils.GetLogger()
	logger.Info("Obtained new Twitch access token", map[string]interface{}{
		"expires_in": tokenResp.ExpiresIn,
	})
	return nil
}
