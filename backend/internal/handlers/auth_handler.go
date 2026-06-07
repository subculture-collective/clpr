package handlers

import (
	"errors"
	"fmt"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"git.subcult.tv/subculture-collective/clpr/config"
	"git.subcult.tv/subculture-collective/clpr/internal/models"
	"git.subcult.tv/subculture-collective/clpr/internal/repository"
	"git.subcult.tv/subculture-collective/clpr/internal/services"
)

// AuthHandler handles authentication endpoints
type AuthHandler struct {
	authService *services.AuthService
	cfg         *config.Config
}

// NewAuthHandler creates a new auth handler
func NewAuthHandler(authService *services.AuthService, cfg *config.Config) *AuthHandler {
	return &AuthHandler{
		authService: authService,
		cfg:         cfg,
	}
}

// InitiateOAuth handles GET /auth/twitch
// Supports PKCE (code_challenge, code_challenge_method parameters)
func (h *AuthHandler) InitiateOAuth(c *gin.Context) {
	// Get PKCE parameters if provided
	codeChallenge := c.Query("code_challenge")
	codeChallengeMethod := c.Query("code_challenge_method")
	clientState := c.Query("state")

	authURL, err := h.authService.GenerateAuthURL(c.Request.Context(), codeChallenge, codeChallengeMethod, clientState)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to generate auth URL",
		})
		return
	}

	// Redirect to Twitch authorization page
	c.Redirect(http.StatusTemporaryRedirect, authURL)
}

// HandleCallback handles GET /auth/twitch/callback
// For PKCE flow: returns code and state for frontend to complete
// For non-PKCE flow: directly authenticates and sets cookies
func (h *AuthHandler) HandleCallback(c *gin.Context) {
	code := c.Query("code")
	state := c.Query("state")

	if code == "" || state == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Missing code or state parameter",
		})
		return
	}

	// Check if PKCE was used by validating state in Redis
	// If PKCE challenge exists in Redis, we need code_verifier from frontend
	stateKey := fmt.Sprintf("oauth:state:%s", state)
	stateValue, err := h.authService.GetStateValue(c.Request.Context(), stateKey)

	if err != nil || stateValue == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Invalid state parameter",
		})
		return
	}

	// Check if PKCE was used (stateValue contains code challenge)
	if strings.Contains(stateValue, ":") {
		// PKCE flow: return code and state to frontend to complete exchange
		// Frontend will POST to /auth/twitch/callback with code_verifier
		frontendURL := "http://localhost:3000"
		origins := strings.Split(h.cfg.CORS.AllowedOrigins, ",")
		if len(origins) > 0 {
			frontendURL = origins[0]
		}

		// Redirect to frontend callback route to complete PKCE exchange
		c.Redirect(http.StatusTemporaryRedirect, frontendURL+"/auth/success?code="+code+"&state="+state)
		return
	}

	// Non-PKCE flow: complete authentication directly
	_, accessToken, refreshToken, err := h.authService.HandleCallback(c.Request.Context(), code, state, "")
	if err != nil {
		if err == services.ErrInvalidState {
			c.JSON(http.StatusBadRequest, gin.H{
				"error": "Invalid state parameter",
			})
			return
		}
		if err == services.ErrUserBanned {
			c.JSON(http.StatusForbidden, gin.H{
				"error": "Your account has been banned",
			})
			return
		}
		// Log the actual error for debugging
		c.Error(err) // This will be logged by Gin's logger middleware
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Authentication failed",
		})
		return
	}

	// Set HTTP-only secure cookies
	h.setAuthCookies(c, accessToken, refreshToken)

	// Get frontend URL from allowed origins (first one)
	frontendURL := "http://localhost:3000"
	origins := strings.Split(h.cfg.CORS.AllowedOrigins, ",")
	if len(origins) > 0 {
		frontendURL = origins[0]
	}

	// Redirect to frontend with success
	c.Redirect(http.StatusTemporaryRedirect, frontendURL+"/auth/success?success=true")
}

// HandlePKCECallback handles POST /auth/twitch/callback
// For PKCE flow with code_verifier
func (h *AuthHandler) HandlePKCECallback(c *gin.Context) {
	var body struct {
		Code         string `json:"code" binding:"required"`
		State        string `json:"state" binding:"required"`
		CodeVerifier string `json:"code_verifier" binding:"required"`
	}

	if err := c.ShouldBindJSON(&body); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Invalid request body",
		})
		return
	}

	// Handle OAuth callback with PKCE
	_, accessToken, refreshToken, err := h.authService.HandleCallback(c.Request.Context(), body.Code, body.State, body.CodeVerifier)
	if err != nil {
		if err == services.ErrInvalidState {
			c.JSON(http.StatusBadRequest, gin.H{
				"error": "Invalid state parameter",
			})
			return
		}
		if err == services.ErrInvalidCodeVerifier {
			c.JSON(http.StatusBadRequest, gin.H{
				"error": "Invalid code verifier",
			})
			return
		}
		if err == services.ErrUserBanned {
			c.JSON(http.StatusForbidden, gin.H{
				"error": "Your account has been banned",
			})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Authentication failed",
		})
		return
	}

	// Set HTTP-only secure cookies
	h.setAuthCookies(c, accessToken, refreshToken)

	c.JSON(http.StatusOK, gin.H{
		"message": "Authentication successful",
	})
}

// TestLogin provides a deterministic login flow for E2E/local environments without Twitch OAuth
func (h *AuthHandler) TestLogin(c *gin.Context) {
	// Guard: disable in production/release modes
	if h.cfg.Server.GinMode == "release" || strings.EqualFold(h.cfg.Server.Environment, "production") {
		c.JSON(http.StatusForbidden, gin.H{"error": "Test login is disabled in production"})
		return
	}

	var body struct {
		UserID   string `json:"user_id"`
		Username string `json:"username"`
	}

	if err := c.ShouldBindJSON(&body); err != nil || (body.UserID == "" && body.Username == "") {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Provide user_id or username"})
		return
	}

	ctx := c.Request.Context()

	// Resolve user by ID (if provided) otherwise by username
	var userErr error
	var user *models.User
	if body.UserID != "" {
		userID, err := uuid.Parse(body.UserID)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid user_id"})
			return
		}
		user, userErr = h.authService.GetUserByID(ctx, userID)
	} else {
		user, userErr = h.authService.GetUserByUsername(ctx, body.Username)
	}

	if userErr != nil {
		if errors.Is(userErr, repository.ErrUserNotFound) {
			c.JSON(http.StatusNotFound, gin.H{"error": "User not found"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to lookup user"})
		return
	}

	accessToken, refreshToken, err := h.authService.GenerateTokensForUser(ctx, user)
	if err != nil {
		if errors.Is(err, services.ErrUserBanned) {
			c.JSON(http.StatusForbidden, gin.H{"error": "User is banned"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to generate tokens"})
		return
	}

	h.setAuthCookies(c, accessToken, refreshToken)

	c.JSON(http.StatusOK, gin.H{
		"user":          user,
		"access_token":  accessToken,
		"refresh_token": refreshToken,
	})
}

// RefreshToken handles POST /auth/refresh
func (h *AuthHandler) RefreshToken(c *gin.Context) {
	// Get refresh token from cookie or body
	refreshToken, err := c.Cookie("refresh_token")
	if err != nil {
		var body struct {
			RefreshToken string `json:"refresh_token"`
		}
		if err := c.ShouldBindJSON(&body); err != nil || body.RefreshToken == "" {
			c.JSON(http.StatusUnauthorized, gin.H{
				"error": "Missing refresh token",
			})
			return
		}
		refreshToken = body.RefreshToken
	}

	// Refresh tokens
	newAccessToken, newRefreshToken, err := h.authService.RefreshAccessToken(c.Request.Context(), refreshToken)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{
			"error": "Failed to refresh token",
		})
		return
	}

	// Set new cookies
	h.setAuthCookies(c, newAccessToken, newRefreshToken)

	c.JSON(http.StatusOK, gin.H{
		"access_token":  newAccessToken,
		"refresh_token": newRefreshToken,
	})
}

// Logout handles POST /auth/logout
func (h *AuthHandler) Logout(c *gin.Context) {
	// Get refresh token from cookie or body
	refreshToken, err := c.Cookie("refresh_token")
	if err != nil {
		var body struct {
			RefreshToken string `json:"refresh_token"`
		}
		if err := c.ShouldBindJSON(&body); err == nil && body.RefreshToken != "" {
			refreshToken = body.RefreshToken
		}
	}

	// Revoke refresh token
	if refreshToken != "" {
		_ = h.authService.Logout(c.Request.Context(), refreshToken)
	}

	// Clear cookies
	h.clearAuthCookies(c)

	c.JSON(http.StatusOK, gin.H{
		"message": "Logged out successfully",
	})
}

// GetCurrentUser handles GET /auth/me
func (h *AuthHandler) GetCurrentUser(c *gin.Context) {
	// Get user from context (set by auth middleware)
	userInterface, exists := c.Get("user")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{
			"error": "Not authenticated",
		})
		return
	}

	c.JSON(http.StatusOK, userInterface)
}

// ReauthorizeTwitch handles POST /auth/twitch/reauthorize
// Initiates a new OAuth flow to refresh Twitch profile metadata
func (h *AuthHandler) ReauthorizeTwitch(c *gin.Context) {
	// Get user from context (set by auth middleware) - just verify they're authenticated
	_, exists := c.Get("user")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{
			"error": "Not authenticated",
		})
		return
	}

	// Generate new auth URL (no PKCE for reauthorize)
	authURL, err := h.authService.GenerateAuthURL(c.Request.Context(), "", "", "")
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to generate auth URL",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"auth_url": authURL,
	})
}

// setAuthCookies sets authentication cookies
func (h *AuthHandler) setAuthCookies(c *gin.Context, accessToken, refreshToken string) {
	isProduction := h.cfg.Server.GinMode == "release"

	// Access token cookie (15 minutes)
	c.SetCookie(
		"access_token",
		accessToken,
		900, // 15 minutes
		"/",
		"",
		isProduction, // Secure only in production
		true,         // HttpOnly
	)

	// Refresh token cookie (7 days)
	c.SetCookie(
		"refresh_token",
		refreshToken,
		604800, // 7 days
		"/",
		"",
		isProduction, // Secure only in production
		true,         // HttpOnly
	)
}

// clearAuthCookies clears authentication cookies
func (h *AuthHandler) clearAuthCookies(c *gin.Context) {
	c.SetCookie("access_token", "", -1, "/", "", false, true)
	c.SetCookie("refresh_token", "", -1, "/", "", false, true)
}
