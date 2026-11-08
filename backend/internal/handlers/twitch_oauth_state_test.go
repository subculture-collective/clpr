package handlers

import (
	"net/http"
	"net/http/httptest"
	"net/url"
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
	if w.Header().Get("Location") != "/settings?error=invalid_oauth_state" {
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

func runTwitchOAuthCallback(t *testing.T, query string, userID uuid.UUID) string {
	t.Helper()
	gin.SetMode(gin.TestMode)
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(http.MethodGet, "/api/v1/twitch/oauth/callback?"+query, nil)
	c.Set("user_id", userID)
	(&TwitchOAuthHandler{}).TwitchOAuthCallback(c)
	return w.Header().Get("Location")
}

// Twitch-connect failures must land on a page the frontend serves: the signed
// return path when the state verifies, /settings otherwise. Never /streams.
func TestTwitchOAuthCallbackErrorsRedirectToSignedReturnPath(t *testing.T) {
	t.Setenv("TWITCH_CLIENT_SECRET", "test-secret")
	t.Setenv("TWITCH_CLIENT_ID", "")
	t.Setenv("TWITCH_PLATFORM_REDIRECT_URI", "")
	t.Setenv("TWITCH_REDIRECT_URI", "")
	userID := uuid.New()
	sign := func(returnTo string) string {
		state, err := signTwitchOAuthState(userID, "test-secret", time.Now(), returnTo)
		if err != nil {
			t.Fatal(err)
		}
		return url.QueryEscape(state)
	}
	clipRoom := sign("/streamer-tools/teststreamer/clips")
	otherUserState, err := signTwitchOAuthState(uuid.New(), "test-secret", time.Now(), "/streamer-tools/teststreamer/clips")
	if err != nil {
		t.Fatal(err)
	}

	tests := []struct {
		name  string
		query string
		want  string
	}{
		{"denied with signed return path", "error=access_denied&state=" + clipRoom, "/streamer-tools/teststreamer/clips?error=oauth_denied"},
		{"denied without state", "error=access_denied", "/settings?error=oauth_denied"},
		{"denied with another user's state", "error=access_denied&state=" + url.QueryEscape(otherUserState), "/settings?error=oauth_denied"},
		{"denied with external return path", "error=access_denied&state=" + sign("https://evil.example/steal"), "/settings?error=oauth_denied"},
		{"denied with protocol-relative return path", "error=access_denied&state=" + sign("//evil.example/steal"), "/settings?error=oauth_denied"},
		{"invalid state", "code=abc&state=forged", "/settings?error=invalid_oauth_state"},
		{"config error after verified state", "code=abc&state=" + clipRoom, "/streamer-tools/teststreamer/clips?error=oauth_config_error"},
		{"config error without return path", "code=abc&state=" + sign(""), "/settings?error=oauth_config_error"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := runTwitchOAuthCallback(t, tt.query, userID); got != tt.want {
				t.Fatalf("redirect = %q, want %q", got, tt.want)
			}
		})
	}
}

func TestTwitchOAuthStateCarriesFeaturePurpose(t *testing.T) {
	userID := uuid.New()
	signed, err := signTwitchOAuthStateFor(userID, "secret", time.Now(), "/settings", "feature")
	if err != nil {
		t.Fatal(err)
	}
	state, ok := decodeTwitchOAuthState(signed)
	if !ok || !verifyTwitchOAuthState(signed, userID, "secret", time.Now()) || state.Purpose != "feature" || state.ReturnTo != "/settings" {
		t.Fatalf("state = %+v, ok %v", state, ok)
	}
	plain, _ := signTwitchOAuthState(userID, "secret", time.Now())
	if s, _ := decodeTwitchOAuthState(plain); s.Purpose != "" {
		t.Fatalf("plain state has purpose %q", s.Purpose)
	}
}
