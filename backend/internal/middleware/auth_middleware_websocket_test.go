package middleware

import (
	"encoding/base64"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
)

func TestExtractToken_SubprotocolAndHeaders(t *testing.T) {
	gin.SetMode(gin.TestMode)

	tests := []struct {
		name          string
		setupRequest  func(*gin.Context)
		expectedToken string
		description   string
	}{
		{
			name: "token from Authorization header",
			setupRequest: func(c *gin.Context) {
				c.Request.Header.Set("Authorization", "Bearer test-token-header")
			},
			expectedToken: "test-token-header",
			description:   "Should extract token from Authorization header",
		},
		{
			name: "token from WebSocket subprotocol",
			setupRequest: func(c *gin.Context) {
				token := base64.StdEncoding.EncodeToString([]byte("test-token-subprotocol"))
				c.Request.Header.Set("Sec-WebSocket-Protocol", "auth.bearer."+token)
			},
			expectedToken: "test-token-subprotocol",
			description:   "Should extract token from WebSocket subprotocol header",
		},
		{
			name: "header takes precedence over subprotocol",
			setupRequest: func(c *gin.Context) {
				c.Request.Header.Set("Authorization", "Bearer test-token-header")
				token := base64.StdEncoding.EncodeToString([]byte("test-token-subprotocol"))
				c.Request.Header.Set("Sec-WebSocket-Protocol", "auth.bearer."+token)
			},
			expectedToken: "test-token-header",
			description:   "Authorization header should take precedence over subprotocol",
		},
		{
			name: "token from cookie",
			setupRequest: func(c *gin.Context) {
				c.Request.AddCookie(&http.Cookie{
					Name:  "access_token",
					Value: "test-token-cookie",
				})
			},
			expectedToken: "test-token-cookie",
			description:   "Should extract token from cookie as fallback",
		},
		{
			name: "subprotocol takes precedence over cookie",
			setupRequest: func(c *gin.Context) {
				token := base64.StdEncoding.EncodeToString([]byte("test-token-subprotocol"))
				c.Request.Header.Set("Sec-WebSocket-Protocol", "auth.bearer."+token)
				c.Request.AddCookie(&http.Cookie{
					Name:  "access_token",
					Value: "test-token-cookie",
				})
			},
			expectedToken: "test-token-subprotocol",
			description:   "Subprotocol should take precedence over cookie",
		},
		{
			name: "no token provided",
			setupRequest: func(c *gin.Context) {
				// No token in any location
			},
			expectedToken: "",
			description:   "Should return empty string when no token provided",
		},
		{
			name: "malformed Authorization header",
			setupRequest: func(c *gin.Context) {
				c.Request.Header.Set("Authorization", "InvalidFormat")
			},
			expectedToken: "",
			description:   "Should return empty string for malformed Authorization header",
		},
		{
			name: "malformed subprotocol format",
			setupRequest: func(c *gin.Context) {
				c.Request.Header.Set("Sec-WebSocket-Protocol", "invalid-format")
			},
			expectedToken: "",
			description:   "Should return empty string for malformed subprotocol",
		},
		{
			name: "invalid base64 in subprotocol",
			setupRequest: func(c *gin.Context) {
				c.Request.Header.Set("Sec-WebSocket-Protocol", "auth.bearer.!!!invalid-base64!!!")
			},
			expectedToken: "",
			description:   "Should return empty string for invalid base64 in subprotocol",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			w := httptest.NewRecorder()
			c, _ := gin.CreateTestContext(w)
			c.Request = httptest.NewRequest(http.MethodGet, "/test", nil)

			tt.setupRequest(c)

			token := extractToken(c)
			assert.Equal(t, tt.expectedToken, token, tt.description)
		})
	}
}

func TestExtractToken_WebSocketScenario(t *testing.T) {
	gin.SetMode(gin.TestMode)

	// Simulate a WebSocket upgrade request where token is passed via subprotocol
	// This prevents tokens from appearing in URLs which could be logged
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(
		http.MethodGet,
		"/api/v1/watch-parties/123/ws",
		nil,
	)
	c.Request.Header.Set("Upgrade", "websocket")
	c.Request.Header.Set("Connection", "Upgrade")

	// Token passed via Sec-WebSocket-Protocol header (not in URL)
	jwtToken := "fixture-jwt-token"
	encodedToken := base64.StdEncoding.EncodeToString([]byte(jwtToken))
	c.Request.Header.Set("Sec-WebSocket-Protocol", "auth.bearer."+encodedToken)

	token := extractToken(c)

	assert.Equal(t, jwtToken, token,
		"Should extract JWT token from subprotocol for WebSocket connections")
	assert.NotEmpty(t, token, "Token should not be empty for authenticated WebSocket")
}
