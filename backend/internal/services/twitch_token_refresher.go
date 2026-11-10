package services

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"

	"git.subcult.tv/subculture-collective/clpr/internal/models"
	"github.com/google/uuid"
)

const twitchOAuthEndpoint = "https://id.twitch.tv/oauth2/token"

// TwitchTokenStore saves a refreshed broadcaster token.
type TwitchTokenStore interface {
	RefreshToken(ctx context.Context, userID uuid.UUID, accessToken, refreshToken, scopes string, expiresAt time.Time) error
}

// TwitchTokenRefresher renews a broadcaster's expired user token with its
// refresh token, so the render host can use Helix downloads without the
// broadcaster visiting clpr first. Twitch user tokens last about four hours.
type TwitchTokenRefresher struct {
	ClientID     string
	ClientSecret string
	Store        TwitchTokenStore
	HTTP         *http.Client
	TokenURL     string
	Now          func() time.Time
}

// Refresh renews auth in place and saves it.
func (r *TwitchTokenRefresher) Refresh(ctx context.Context, auth *models.TwitchAuth) error {
	if r.ClientID == "" || r.ClientSecret == "" {
		return fmt.Errorf("twitch token refresh is not configured")
	}
	if auth.RefreshToken == "" {
		return fmt.Errorf("the broadcaster's Twitch connection has no refresh token")
	}
	form := url.Values{
		"client_id":     {r.ClientID},
		"client_secret": {r.ClientSecret},
		"refresh_token": {auth.RefreshToken},
		"grant_type":    {"refresh_token"},
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, orDefault(r.TokenURL, twitchOAuthEndpoint), strings.NewReader(form.Encode()))
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	client := r.HTTP
	if client == nil {
		client = &http.Client{Timeout: 10 * time.Second}
	}
	resp, err := client.Do(req)
	if err != nil {
		return fmt.Errorf("twitch token refresh: %w", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		// The body can name the token; report only the status.
		return fmt.Errorf("twitch token refresh failed with status %d", resp.StatusCode)
	}
	var body struct {
		AccessToken  string   `json:"access_token"`
		RefreshToken string   `json:"refresh_token"`
		ExpiresIn    int      `json:"expires_in"`
		Scope        []string `json:"scope"`
	}
	if err := json.NewDecoder(io.LimitReader(resp.Body, 1<<16)).Decode(&body); err != nil || body.AccessToken == "" {
		return fmt.Errorf("twitch token refresh returned no token")
	}
	now := time.Now()
	if r.Now != nil {
		now = r.Now()
	}
	refresh := body.RefreshToken
	if refresh == "" {
		refresh = auth.RefreshToken
	}
	scopes := strings.Join(body.Scope, " ")
	if scopes == "" {
		scopes = auth.Scopes
	}
	expiresAt := now.Add(time.Duration(body.ExpiresIn) * time.Second)
	if err := r.Store.RefreshToken(ctx, auth.UserID, body.AccessToken, refresh, scopes, expiresAt); err != nil {
		return fmt.Errorf("save the refreshed twitch token: %w", err)
	}
	auth.AccessToken, auth.RefreshToken, auth.Scopes, auth.ExpiresAt = body.AccessToken, refresh, scopes, expiresAt
	return nil
}
