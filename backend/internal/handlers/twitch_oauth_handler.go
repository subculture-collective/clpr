package handlers

import (
	"context"
	"crypto/hmac"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"os"
	"strings"
	"time"

	"git.subcult.tv/subculture-collective/clpr/internal/models"
	"git.subcult.tv/subculture-collective/clpr/internal/repository"
	"git.subcult.tv/subculture-collective/clpr/pkg/utils"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

// TwitchOAuthHandler handles Twitch OAuth for chat integration
type TwitchOAuthHandler struct {
	twitchAuthRepo *repository.TwitchAuthRepository
}

// NewTwitchOAuthHandler creates a new Twitch OAuth handler
func NewTwitchOAuthHandler(twitchAuthRepo *repository.TwitchAuthRepository) *TwitchOAuthHandler {
	return &TwitchOAuthHandler{
		twitchAuthRepo: twitchAuthRepo,
	}
}

// TwitchTokenResponse represents the response from Twitch token endpoint
type TwitchTokenResponse struct {
	AccessToken  string   `json:"access_token"`
	RefreshToken string   `json:"refresh_token"`
	ExpiresIn    int      `json:"expires_in"`
	Scope        []string `json:"scope"`
	TokenType    string   `json:"token_type"`
}

// TwitchUserResponse represents the response from Twitch users endpoint
type TwitchUserResponse struct {
	Data []struct {
		ID    string `json:"id"`
		Login string `json:"login"`
	} `json:"data"`
}

type twitchOAuthState struct {
	UserID    string `json:"uid"`
	ExpiresAt int64  `json:"exp"`
	Nonce     string `json:"nonce"`
	ReturnTo  string `json:"return_to,omitempty"`
}

func signTwitchOAuthState(userID uuid.UUID, secret string, now time.Time, returnTo ...string) (string, error) {
	nonce := make([]byte, 24)
	if _, err := rand.Read(nonce); err != nil {
		return "", err
	}
	safeReturnTo := ""
	if len(returnTo) > 0 {
		safeReturnTo = sanitizeOAuthReturnTo(returnTo[0])
	}
	payload, err := json.Marshal(twitchOAuthState{UserID: userID.String(), ExpiresAt: now.Add(10 * time.Minute).Unix(), Nonce: base64.RawURLEncoding.EncodeToString(nonce), ReturnTo: safeReturnTo})
	if err != nil {
		return "", err
	}
	mac := hmac.New(sha256.New, []byte(secret))
	_, _ = mac.Write(payload)
	return base64.RawURLEncoding.EncodeToString(payload) + "." + base64.RawURLEncoding.EncodeToString(mac.Sum(nil)), nil
}

func decodeTwitchOAuthState(value string) (twitchOAuthState, bool) {
	parts := strings.Split(value, ".")
	if len(parts) != 2 {
		return twitchOAuthState{}, false
	}
	payload, err := base64.RawURLEncoding.DecodeString(parts[0])
	if err != nil {
		return twitchOAuthState{}, false
	}
	var state twitchOAuthState
	if json.Unmarshal(payload, &state) != nil {
		return twitchOAuthState{}, false
	}
	return state, true
}

func sanitizeOAuthReturnTo(value string) string {
	value = strings.TrimSpace(value)
	if value == "" || len(value) > 512 || !strings.HasPrefix(value, "/") || strings.HasPrefix(value, "//") || strings.Contains(value, `\`) {
		return ""
	}
	parsed, err := url.Parse(value)
	if err != nil || parsed.IsAbs() || parsed.Host != "" {
		return ""
	}
	return value
}

func appendOAuthResult(returnTo, key, value string) string {
	if returnTo == "" {
		returnTo = "/streams"
	}
	parsed, err := url.Parse(returnTo)
	if err != nil {
		return "/streams"
	}
	query := parsed.Query()
	query.Set(key, value)
	parsed.RawQuery = query.Encode()
	return parsed.String()
}

func hasTwitchScope(scopes, required string) bool {
	for _, scope := range strings.Fields(scopes) {
		if scope == required {
			return true
		}
	}
	return false
}

func verifyTwitchOAuthState(value string, userID uuid.UUID, secret string, now time.Time) bool {
	parts := strings.Split(value, ".")
	if len(parts) != 2 || secret == "" {
		return false
	}
	payload, err := base64.RawURLEncoding.DecodeString(parts[0])
	if err != nil {
		return false
	}
	signature, err := base64.RawURLEncoding.DecodeString(parts[1])
	if err != nil {
		return false
	}
	mac := hmac.New(sha256.New, []byte(secret))
	_, _ = mac.Write(payload)
	if !hmac.Equal(signature, mac.Sum(nil)) {
		return false
	}
	var state twitchOAuthState
	if json.Unmarshal(payload, &state) != nil || state.Nonce == "" || state.UserID != userID.String() {
		return false
	}
	return state.ExpiresAt >= now.Unix() && state.ExpiresAt <= now.Add(10*time.Minute).Unix()
}

// InitiateTwitchOAuth initiates the Twitch OAuth flow for chat
// GET /api/v1/twitch/oauth/authorize
func (h *TwitchOAuthHandler) InitiateTwitchOAuth(c *gin.Context) {
	clientID := os.Getenv("TWITCH_CLIENT_ID")
	clientSecret := os.Getenv("TWITCH_CLIENT_SECRET")
	redirectURI := os.Getenv("TWITCH_PLATFORM_REDIRECT_URI")
	if redirectURI == "" {
		redirectURI = os.Getenv("TWITCH_REDIRECT_URI")
	}

	// Validate required environment variables
	if clientID == "" || clientSecret == "" || redirectURI == "" {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Twitch OAuth is not configured",
		})
		return
	}
	userID, ok := currentUserID(c)
	if !ok {
		return
	}
	state, err := signTwitchOAuthState(userID, clientSecret, time.Now(), c.Query("return_to"))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to initialize OAuth state"})
		return
	}

	// For chat integration, bot authorization, clip downloads, and ban management, we need:
	// - chat:read chat:edit: for chat functionality
	// - channel:bot: let the Clpr cloud bot join and read this broadcaster's chat
	// - channel:manage:clips: obtain official temporary download URLs for this broadcaster's clips
	// - moderator:manage:banned_users: for moderators to ban/unban users
	// - channel:manage:banned_users: for broadcasters to ban/unban users
	scopes := "chat:read chat:edit channel:bot channel:manage:clips moderator:manage:banned_users channel:manage:banned_users"

	params := url.Values{"client_id": {clientID}, "redirect_uri": {redirectURI}, "response_type": {"code"}, "scope": {scopes}, "state": {state}}
	authURL := "https://id.twitch.tv/oauth2/authorize?" + params.Encode()

	c.Redirect(http.StatusTemporaryRedirect, authURL)
}

// TwitchOAuthCallback handles the OAuth callback from Twitch
// GET /api/v1/twitch/oauth/callback
func (h *TwitchOAuthHandler) TwitchOAuthCallback(c *gin.Context) {
	// Check for OAuth errors from Twitch
	if errorParam := c.Query("error"); errorParam != "" {
		errorDesc := c.Query("error_description")
		utils.GetLogger().Warn("Twitch OAuth error", map[string]interface{}{
			"error":             errorParam,
			"error_description": errorDesc,
		})
		c.Redirect(http.StatusTemporaryRedirect, "/streams?error=oauth_denied")
		return
	}

	code := c.Query("code")
	if code == "" || len(code) > 2048 || len(c.Query("state")) > 4096 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "authorization code is required"})
		return
	}

	// Get user ID from context (middleware sets this)
	userIDVal, exists := c.Get("user_id")
	if !exists {
		// Redirect to login if not authenticated
		c.Redirect(http.StatusTemporaryRedirect, "/login?error=authentication_required")
		return
	}
	userID, ok := userIDVal.(uuid.UUID)
	if !ok {
		// Redirect to login if user ID is not in the expected format
		c.Redirect(http.StatusTemporaryRedirect, "/login?error=invalid_user")
		return
	}
	clientSecret := os.Getenv("TWITCH_CLIENT_SECRET")
	if !verifyTwitchOAuthState(c.Query("state"), userID, clientSecret, time.Now()) {
		c.Redirect(http.StatusTemporaryRedirect, "/streams?error=invalid_oauth_state")
		return
	}
	state, _ := decodeTwitchOAuthState(c.Query("state"))
	returnTo := sanitizeOAuthReturnTo(state.ReturnTo)

	ctx := c.Request.Context()

	// Exchange code for tokens
	clientID := os.Getenv("TWITCH_CLIENT_ID")
	redirectURI := os.Getenv("TWITCH_PLATFORM_REDIRECT_URI")
	if redirectURI == "" {
		redirectURI = os.Getenv("TWITCH_REDIRECT_URI")
	}

	// Validate required environment variables
	if clientID == "" || clientSecret == "" || redirectURI == "" {
		utils.GetLogger().Error("Twitch OAuth configuration missing", nil, map[string]interface{}{
			"user_id": userID.String(),
		})
		c.Redirect(http.StatusTemporaryRedirect, "/streams?error=oauth_config_error")
		return
	}

	// Create HTTP client with timeout
	httpClient := &http.Client{
		Timeout: 10 * time.Second,
	}

	tokenResp, err := httpClient.PostForm("https://id.twitch.tv/oauth2/token", url.Values{
		"client_id":     {clientID},
		"client_secret": {clientSecret},
		"code":          {code},
		"grant_type":    {"authorization_code"},
		"redirect_uri":  {redirectURI},
	})

	if err != nil {
		utils.GetLogger().Error("Failed to exchange code for tokens", err, map[string]interface{}{
			"user_id": userID.String(),
		})
		c.Redirect(http.StatusTemporaryRedirect, "/streams?error=oauth_failed")
		return
	}
	defer tokenResp.Body.Close()

	if tokenResp.StatusCode != http.StatusOK {
		utils.GetLogger().Error("Twitch token endpoint returned error", nil, map[string]interface{}{
			"status_code": tokenResp.StatusCode,
			"user_id":     userID.String(),
		})
		c.Redirect(http.StatusTemporaryRedirect, "/streams?error=oauth_failed")
		return
	}

	var tokens TwitchTokenResponse
	if err := json.NewDecoder(tokenResp.Body).Decode(&tokens); err != nil {
		utils.GetLogger().Error("Failed to decode token response", err, map[string]interface{}{
			"user_id": userID.String(),
		})
		c.Redirect(http.StatusTemporaryRedirect, "/streams?error=oauth_failed")
		return
	}

	// Get Twitch user info
	req, err := http.NewRequest("GET", "https://api.twitch.tv/helix/users", nil)
	if err != nil {
		utils.GetLogger().Error("Failed to create Twitch API request", err, map[string]interface{}{
			"user_id": userID.String(),
		})
		c.Redirect(http.StatusTemporaryRedirect, "/streams?error=oauth_failed")
		return
	}
	req.Header.Set("Authorization", "Bearer "+tokens.AccessToken)
	req.Header.Set("Client-Id", clientID)

	var userResp *http.Response
	userResp, err = httpClient.Do(req)
	if err != nil {
		utils.GetLogger().Error("Failed to get Twitch user info", err, map[string]interface{}{
			"user_id": userID.String(),
		})
		c.Redirect(http.StatusTemporaryRedirect, "/streams?error=oauth_failed")
		return
	}
	defer userResp.Body.Close()

	if userResp.StatusCode != http.StatusOK {
		utils.GetLogger().Error("Twitch users endpoint returned error", nil, map[string]interface{}{
			"status_code": userResp.StatusCode,
			"user_id":     userID.String(),
		})
		c.Redirect(http.StatusTemporaryRedirect, "/streams?error=oauth_failed")
		return
	}

	var userData TwitchUserResponse
	if err := json.NewDecoder(userResp.Body).Decode(&userData); err != nil {
		utils.GetLogger().Error("Failed to decode user response", err, map[string]interface{}{
			"user_id": userID.String(),
		})
		c.Redirect(http.StatusTemporaryRedirect, "/streams?error=oauth_failed")
		return
	}

	if len(userData.Data) == 0 {
		utils.GetLogger().Error("No user data returned from Twitch", nil, map[string]interface{}{
			"user_id": userID.String(),
		})
		c.Redirect(http.StatusTemporaryRedirect, "/streams?error=oauth_failed")
		return
	}

	// Store OAuth credentials
	expiresAt := time.Now().Add(time.Duration(tokens.ExpiresIn) * time.Second)
	auth := &models.TwitchAuth{
		UserID:         userID,
		TwitchUserID:   userData.Data[0].ID,
		TwitchUsername: userData.Data[0].Login,
		AccessToken:    tokens.AccessToken,
		RefreshToken:   tokens.RefreshToken,
		Scopes:         strings.Join(tokens.Scope, " "),
		ExpiresAt:      expiresAt,
	}

	if err := h.twitchAuthRepo.UpsertTwitchAuth(ctx, auth); err != nil {
		utils.GetLogger().Error("Failed to store Twitch OAuth credentials", err, map[string]interface{}{
			"user_id": userID.String(),
		})
		c.Redirect(http.StatusTemporaryRedirect, "/streams?error=oauth_failed")
		return
	}

	utils.GetLogger().Info("Twitch OAuth completed successfully", map[string]interface{}{
		"user_id":         userID.String(),
		"twitch_username": userData.Data[0].Login,
	})

	// Return to the feature that initiated authorization when possible.
	c.Redirect(http.StatusFound, appendOAuthResult(returnTo, "twitch_connected", "true"))
}

// GetTwitchAuthStatus returns the Twitch authentication status for the current user
// GET /api/v1/twitch/auth/status
func (h *TwitchOAuthHandler) GetTwitchAuthStatus(c *gin.Context) {
	// Get user ID from context (middleware sets this)
	userIDVal, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusOK, models.TwitchAuthStatusResponse{
			Authenticated: false,
		})
		return
	}
	userID, ok := userIDVal.(uuid.UUID)
	if !ok {
		utils.GetLogger().Error("Invalid user_id type in context", nil, map[string]interface{}{
			"user_id_type": fmt.Sprintf("%T", userIDVal),
		})
		c.JSON(http.StatusInternalServerError, gin.H{"error": "internal server error"})
		return
	}

	ctx := c.Request.Context()

	// Get Twitch auth credentials
	auth, err := h.twitchAuthRepo.GetTwitchAuth(ctx, userID)
	if err != nil {
		utils.GetLogger().Error("Failed to get Twitch auth", err, map[string]interface{}{
			"user_id": userID.String(),
		})
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to check authentication status"})
		return
	}

	if auth == nil {
		c.JSON(http.StatusOK, models.TwitchAuthStatusResponse{
			Authenticated: false,
		})
		return
	}

	// Check if token needs refresh
	if h.twitchAuthRepo.IsTokenExpired(auth) {
		// Attempt to refresh the token
		if err := h.refreshTwitchToken(ctx, auth); err != nil {
			utils.GetLogger().Error("Failed to refresh Twitch token", err, map[string]interface{}{
				"user_id": userID.String(),
			})
			// Token refresh failed, return not authenticated
			c.JSON(http.StatusOK, models.TwitchAuthStatusResponse{
				Authenticated: false,
			})
			return
		}
	}

	twitchUsername := auth.TwitchUsername
	c.JSON(http.StatusOK, models.TwitchAuthStatusResponse{
		Authenticated:          true,
		BotAuthorized:          hasTwitchScope(auth.Scopes, "channel:bot"),
		ClipDownloadAuthorized: hasTwitchScope(auth.Scopes, "channel:manage:clips"),
		TwitchUsername:         &twitchUsername,
	})
}

// RevokeTwitchAuth revokes Twitch OAuth credentials
// DELETE /api/v1/twitch/auth
func (h *TwitchOAuthHandler) RevokeTwitchAuth(c *gin.Context) {
	// Get user ID from context (middleware sets this)
	userIDVal, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "authentication required"})
		return
	}
	userID, ok := userIDVal.(uuid.UUID)
	if !ok {
		utils.GetLogger().Error("Invalid user_id type in context", nil, map[string]interface{}{
			"user_id_type": fmt.Sprintf("%T", userIDVal),
		})
		c.JSON(http.StatusInternalServerError, gin.H{"error": "internal server error"})
		return
	}

	ctx := c.Request.Context()

	// Delete Twitch auth credentials
	if err := h.twitchAuthRepo.DeleteTwitchAuth(ctx, userID); err != nil {
		utils.GetLogger().Error("Failed to revoke Twitch auth", err, map[string]interface{}{
			"user_id": userID.String(),
		})
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to revoke authentication"})
		return
	}

	utils.GetLogger().Info("Twitch OAuth revoked", map[string]interface{}{
		"user_id": userID.String(),
	})

	c.JSON(http.StatusOK, gin.H{"message": "Twitch authentication revoked successfully"})
}

// refreshTwitchToken refreshes an expired Twitch token
func (h *TwitchOAuthHandler) refreshTwitchToken(ctx context.Context, auth *models.TwitchAuth) error {
	clientID := os.Getenv("TWITCH_CLIENT_ID")
	clientSecret := os.Getenv("TWITCH_CLIENT_SECRET")
	if clientID == "" || clientSecret == "" {
		return fmt.Errorf("Twitch OAuth is not configured")
	}

	// Create HTTP client with timeout
	httpClient := &http.Client{
		Timeout: 10 * time.Second,
	}

	tokenResp, err := httpClient.PostForm("https://id.twitch.tv/oauth2/token", url.Values{
		"client_id":     {clientID},
		"client_secret": {clientSecret},
		"refresh_token": {auth.RefreshToken},
		"grant_type":    {"refresh_token"},
	})

	if err != nil {
		return fmt.Errorf("failed to refresh token: %w", err)
	}
	defer tokenResp.Body.Close()

	if tokenResp.StatusCode != http.StatusOK {
		return fmt.Errorf("token refresh failed with status: %d", tokenResp.StatusCode)
	}

	var tokens TwitchTokenResponse
	if err := json.NewDecoder(tokenResp.Body).Decode(&tokens); err != nil {
		return fmt.Errorf("failed to decode token response: %w", err)
	}

	// Update tokens in database
	expiresAt := time.Now().Add(time.Duration(tokens.ExpiresIn) * time.Second)
	scopes := strings.Join(tokens.Scope, " ")
	if err := h.twitchAuthRepo.RefreshToken(ctx, auth.UserID, tokens.AccessToken, tokens.RefreshToken, scopes, expiresAt); err != nil {
		return fmt.Errorf("failed to update tokens: %w", err)
	}

	// Update the auth object with new values
	auth.AccessToken = tokens.AccessToken
	auth.RefreshToken = tokens.RefreshToken
	auth.Scopes = scopes
	auth.ExpiresAt = expiresAt

	utils.GetLogger().Info("Twitch token refreshed successfully", map[string]interface{}{
		"user_id": auth.UserID.String(),
	})

	return nil
}
