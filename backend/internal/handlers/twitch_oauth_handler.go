package handlers

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"os"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"git.subcult.tv/subculture-collective/clpr/internal/models"
	"git.subcult.tv/subculture-collective/clpr/internal/repository"
	"git.subcult.tv/subculture-collective/clpr/pkg/utils"
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
	AccessToken  string `json:"access_token"`
	RefreshToken string `json:"refresh_token"`
	ExpiresIn    int    `json:"expires_in"`
	Scope        string `json:"scope"`
	TokenType    string `json:"token_type"`
}

// TwitchUserResponse represents the response from Twitch users endpoint
type TwitchUserResponse struct {
	Data []struct {
		ID    string `json:"id"`
		Login string `json:"login"`
	} `json:"data"`
}

// InitiateTwitchOAuth initiates the Twitch OAuth flow for chat
// GET /api/v1/twitch/oauth/authorize
func (h *TwitchOAuthHandler) InitiateTwitchOAuth(c *gin.Context) {
	clientID := os.Getenv("TWITCH_CLIENT_ID")
	redirectURI := os.Getenv("TWITCH_REDIRECT_URI")

	// Validate required environment variables
	if clientID == "" || redirectURI == "" {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Twitch OAuth is not configured. Please set TWITCH_CLIENT_ID and TWITCH_REDIRECT_URI.",
		})
		return
	}

	// For chat integration and ban management, we need:
	// - chat:read chat:edit: for chat functionality
	// - moderator:manage:banned_users: for moderators to ban/unban users
	// - channel:manage:banned_users: for broadcasters to ban/unban users
	scopes := "chat:read chat:edit moderator:manage:banned_users channel:manage:banned_users"

	authURL := fmt.Sprintf(
		"https://id.twitch.tv/oauth2/authorize?client_id=%s&redirect_uri=%s&response_type=code&scope=%s",
		clientID,
		url.QueryEscape(redirectURI),
		url.QueryEscape(scopes),
	)

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
	if code == "" {
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

	ctx := c.Request.Context()

	// Exchange code for tokens
	clientID := os.Getenv("TWITCH_CLIENT_ID")
	clientSecret := os.Getenv("TWITCH_CLIENT_SECRET")
	redirectURI := os.Getenv("TWITCH_REDIRECT_URI")

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
		Scopes:         tokens.Scope,
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

	// Redirect back to streams page with success message
	c.Redirect(http.StatusFound, "/streams?twitch_connected=true")
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
		Authenticated:  true,
		TwitchUsername: &twitchUsername,
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
	if err := h.twitchAuthRepo.RefreshToken(ctx, auth.UserID, tokens.AccessToken, tokens.RefreshToken, tokens.Scope, expiresAt); err != nil {
		return fmt.Errorf("failed to update tokens: %w", err)
	}

	// Update the auth object with new values
	auth.AccessToken = tokens.AccessToken
	auth.RefreshToken = tokens.RefreshToken
	auth.Scopes = tokens.Scope
	auth.ExpiresAt = expiresAt

	utils.GetLogger().Info("Twitch token refreshed successfully", map[string]interface{}{
		"user_id": auth.UserID.String(),
	})

	return nil
}
