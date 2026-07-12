package handlers

import (
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

func TestTwitchOAuthStateIsUserBoundAndTamperEvident(t *testing.T) {
	now := time.Unix(1_700_000_000, 0)
	userID := uuid.New()
	state, err := signTwitchOAuthState(userID, "test-secret", now)
	if err != nil {
		t.Fatal(err)
	}
	if !verifyTwitchOAuthState(state, userID, "test-secret", now.Add(time.Minute)) {
		t.Fatal("valid state rejected")
	}
	if verifyTwitchOAuthState(state, uuid.New(), "test-secret", now.Add(time.Minute)) {
		t.Fatal("state accepted for another user")
	}
	if verifyTwitchOAuthState(state, userID, "wrong-secret", now.Add(time.Minute)) {
		t.Fatal("state accepted with wrong secret")
	}
	parts := strings.Split(state, ".")
	parts[0] = "A" + parts[0][1:]
	if verifyTwitchOAuthState(strings.Join(parts, "."), userID, "test-secret", now.Add(time.Minute)) {
		t.Fatal("tampered state accepted")
	}
}

func TestTwitchOAuthCallbackRejectsMissingStateBeforeExchange(t *testing.T) {
	gin.SetMode(gin.TestMode)
	os.Setenv("TWITCH_CLIENT_SECRET", "test-secret")
	defer os.Unsetenv("TWITCH_CLIENT_SECRET")
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(http.MethodGet, "/api/v1/twitch/oauth/callback?code=attacker-code", nil)
	c.Set("user_id", uuid.New())
	(&TwitchOAuthHandler{}).TwitchOAuthCallback(c)
	if w.Code != http.StatusTemporaryRedirect {
		t.Fatalf("expected 307, got %d", w.Code)
	}
	if w.Header().Get("Location") != "/streams?error=invalid_oauth_state" {
		t.Fatalf("unexpected redirect: %s", w.Header().Get("Location"))
	}
}

func TestTwitchOAuthStateExpires(t *testing.T) {
	now := time.Unix(1_700_000_000, 0)
	userID := uuid.New()
	state, err := signTwitchOAuthState(userID, "test-secret", now)
	if err != nil {
		t.Fatal(err)
	}
	if verifyTwitchOAuthState(state, userID, "test-secret", now.Add(11*time.Minute)) {
		t.Fatal("expired state accepted")
	}
}
