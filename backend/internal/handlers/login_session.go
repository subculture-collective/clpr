package handlers

import (
	"context"

	"git.subcult.tv/subculture-collective/clpr/config"
	"git.subcult.tv/subculture-collective/clpr/internal/models"
	"git.subcult.tv/subculture-collective/clpr/internal/services"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

// mfaChallengeCookie holds the opaque token of a pending MFA sign-in.
const mfaChallengeCookie = "mfa_challenge"

// sessionTokenIssuer issues access and refresh tokens for a user.
type sessionTokenIssuer interface {
	GenerateTokensForUser(ctx context.Context, user *models.User) (string, string, error)
}

// loginMFAGate decides whether a sign-in needs a second factor.
type loginMFAGate interface {
	LoginRequiresMFA(ctx context.Context, userID uuid.UUID, deviceFingerprint string) (bool, error)
}

// mfaChallengeStore keeps pending MFA sign-ins.
type mfaChallengeStore interface {
	Create(ctx context.Context, userID uuid.UUID) (string, error)
	Resolve(ctx context.Context, token string) (uuid.UUID, error)
	RecordFailure(ctx context.Context, token string) (exhausted bool, err error)
	Consume(ctx context.Context, token string) (uuid.UUID, error)
}

// loginOutcome is the result of starting a session after the first sign-in
// step.
type loginOutcome int

const (
	// loginCompleted means session cookies were set.
	loginCompleted loginOutcome = iota
	// loginMFARequired means a challenge cookie was set and the user must
	// submit an MFA code to POST /auth/mfa/verify-login.
	loginMFARequired
)

// loginSession issues sessions for every sign-in path, so no path can skip
// the MFA check.
type loginSession struct {
	cfg        *config.Config
	tokens     sessionTokenIssuer
	mfa        loginMFAGate
	challenges mfaChallengeStore
}

func newLoginSession(cfg *config.Config, tokens sessionTokenIssuer, mfa loginMFAGate, challenges mfaChallengeStore) *loginSession {
	return &loginSession{cfg: cfg, tokens: tokens, mfa: mfa, challenges: challenges}
}

// requestDeviceFingerprint identifies the browser for trusted-device checks.
func requestDeviceFingerprint(c *gin.Context) string {
	return services.GenerateDeviceFingerprint(c.GetHeader("User-Agent"), c.ClientIP())
}

// requiresMFA reports whether this sign-in needs a second factor.
func (s *loginSession) requiresMFA(c *gin.Context, user *models.User) (bool, error) {
	if s.mfa == nil {
		return false, nil
	}
	return s.mfa.LoginRequiresMFA(c.Request.Context(), user.ID, requestDeviceFingerprint(c))
}

// sessionTokens are the tokens issued when a sign-in completes.
type sessionTokens struct {
	access  string
	refresh string
}

// begin is called once the user's identity is established. Users with MFA
// enabled on an untrusted device get a challenge cookie instead of a session.
// Any error checking MFA status fails closed: no session is issued.
func (s *loginSession) begin(c *gin.Context, user *models.User) (loginOutcome, sessionTokens, error) {
	if user.IsBanned {
		return 0, sessionTokens{}, services.ErrUserBanned
	}
	required, err := s.requiresMFA(c, user)
	if err != nil {
		return 0, sessionTokens{}, err
	}
	if required {
		if s.challenges == nil {
			return 0, sessionTokens{}, services.ErrMFAChallengeNotFound
		}
		token, err := s.challenges.Create(c.Request.Context(), user.ID)
		if err != nil {
			return 0, sessionTokens{}, err
		}
		// Drop any earlier session in this browser, so the pending sign-in
		// cannot ride on another account's cookies.
		clearSessionCookies(c)
		s.setChallengeCookie(c, token)
		return loginMFARequired, sessionTokens{}, nil
	}
	access, refresh, err := s.issueTokens(c, user)
	if err != nil {
		return 0, sessionTokens{}, err
	}
	clearChallengeCookie(c)
	return loginCompleted, sessionTokens{access: access, refresh: refresh}, nil
}

// issueTokens sets session cookies for user and returns the issued tokens.
func (s *loginSession) issueTokens(c *gin.Context, user *models.User) (string, string, error) {
	accessToken, refreshToken, err := s.tokens.GenerateTokensForUser(c.Request.Context(), user)
	if err != nil {
		return "", "", err
	}
	setSessionCookies(c, s.cfg, accessToken, refreshToken)
	return accessToken, refreshToken, nil
}

func (s *loginSession) setChallengeCookie(c *gin.Context, token string) {
	c.SetCookie(mfaChallengeCookie, token, int(services.MFALoginChallengeTTL.Seconds()), "/", "", isReleaseMode(s.cfg), true)
}

func clearChallengeCookie(c *gin.Context) {
	c.SetCookie(mfaChallengeCookie, "", -1, "/", "", false, true)
}

func isReleaseMode(cfg *config.Config) bool {
	return cfg != nil && cfg.Server.GinMode == "release"
}

// setSessionCookies sets the HttpOnly access and refresh token cookies.
func setSessionCookies(c *gin.Context, cfg *config.Config, accessToken, refreshToken string) {
	secure := isReleaseMode(cfg)
	c.SetCookie("access_token", accessToken, 900, "/", "", secure, true)      // 15 minutes
	c.SetCookie("refresh_token", refreshToken, 604800, "/", "", secure, true) // 7 days
}

// clearSessionCookies removes the access and refresh token cookies.
func clearSessionCookies(c *gin.Context) {
	c.SetCookie("access_token", "", -1, "/", "", false, true)
	c.SetCookie("refresh_token", "", -1, "/", "", false, true)
}
