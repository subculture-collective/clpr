package handlers

import (
	"bytes"
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"git.subcult.tv/subculture-collective/clpr/config"
	"git.subcult.tv/subculture-collective/clpr/internal/models"
	"git.subcult.tv/subculture-collective/clpr/internal/services"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

type fakeTokenIssuer struct{ calls int }

func (f *fakeTokenIssuer) GenerateTokensForUser(_ context.Context, user *models.User) (string, string, error) {
	f.calls++
	if user.IsBanned {
		return "", "", services.ErrUserBanned
	}
	return "access-" + user.ID.String(), "refresh-" + user.ID.String(), nil
}

type fakeMFAGate struct {
	required bool
	err      error
}

func (f fakeMFAGate) LoginRequiresMFA(context.Context, uuid.UUID, string) (bool, error) {
	return f.required, f.err
}

type fakeChallengeStore struct {
	byToken  map[string]uuid.UUID
	failures map[string]int
	max      int
}

func newFakeChallengeStore() *fakeChallengeStore {
	return &fakeChallengeStore{byToken: map[string]uuid.UUID{}, failures: map[string]int{}, max: 3}
}

func (f *fakeChallengeStore) Create(_ context.Context, userID uuid.UUID) (string, error) {
	token := "challenge-" + uuid.NewString()
	f.byToken[token] = userID
	return token, nil
}

func (f *fakeChallengeStore) Resolve(_ context.Context, token string) (uuid.UUID, error) {
	id, ok := f.byToken[token]
	if !ok {
		return uuid.Nil, services.ErrMFAChallengeNotFound
	}
	return id, nil
}

func (f *fakeChallengeStore) RecordFailure(_ context.Context, token string) (bool, error) {
	f.failures[token]++
	if f.failures[token] >= f.max {
		delete(f.byToken, token)
		return true, nil
	}
	return false, nil
}

func (f *fakeChallengeStore) Consume(_ context.Context, token string) (uuid.UUID, error) {
	id, ok := f.byToken[token]
	if !ok {
		return uuid.Nil, services.ErrMFAChallengeNotFound
	}
	delete(f.byToken, token)
	return id, nil
}

func responseCookies(w *httptest.ResponseRecorder) map[string]*http.Cookie {
	cookies := map[string]*http.Cookie{}
	for _, cookie := range (&http.Response{Header: w.Header()}).Cookies() {
		cookies[cookie.Name] = cookie
	}
	return cookies
}

func runBegin(t *testing.T, session *loginSession, user *models.User) (loginOutcome, sessionTokens, error, *httptest.ResponseRecorder) {
	t.Helper()
	gin.SetMode(gin.TestMode)
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(http.MethodPost, "/api/v1/auth/twitch/callback", nil)
	c.Request.Header.Set("User-Agent", "test-agent")
	outcome, tokens, err := session.begin(c, user)
	return outcome, tokens, err, w
}

func TestLoginSessionRequiresMFAForEnrolledUsers(t *testing.T) {
	issuer := &fakeTokenIssuer{}
	store := newFakeChallengeStore()
	session := newLoginSession(&config.Config{}, issuer, fakeMFAGate{required: true}, store)
	user := &models.User{ID: uuid.New()}

	outcome, tokens, err, w := runBegin(t, session, user)
	if err != nil || outcome != loginMFARequired {
		t.Fatalf("begin = %v, %v; want loginMFARequired", outcome, err)
	}
	if issuer.calls != 0 || tokens.access != "" {
		t.Fatalf("session tokens issued before MFA: calls=%d tokens=%+v", issuer.calls, tokens)
	}
	cookies := responseCookies(w)
	challenge := cookies[mfaChallengeCookie]
	if challenge == nil || challenge.Value == "" || !challenge.HttpOnly {
		t.Fatalf("missing HttpOnly challenge cookie: %+v", challenge)
	}
	if store.byToken[challenge.Value] != user.ID {
		t.Fatalf("challenge cookie does not name a stored challenge for the user")
	}
	if access := cookies["access_token"]; access == nil || access.Value != "" || access.MaxAge >= 0 {
		t.Fatalf("existing session cookie not cleared: %+v", access)
	}
}

func TestLoginSessionIssuesTokensWhenMFANotRequired(t *testing.T) {
	issuer := &fakeTokenIssuer{}
	session := newLoginSession(&config.Config{}, issuer, fakeMFAGate{required: false}, newFakeChallengeStore())
	user := &models.User{ID: uuid.New()}

	outcome, tokens, err, w := runBegin(t, session, user)
	if err != nil || outcome != loginCompleted {
		t.Fatalf("begin = %v, %v; want loginCompleted", outcome, err)
	}
	if tokens.access == "" || tokens.refresh == "" {
		t.Fatalf("tokens not returned: %+v", tokens)
	}
	if access := responseCookies(w)["access_token"]; access == nil || access.Value != tokens.access {
		t.Fatalf("access cookie = %+v", access)
	}
}

func TestLoginSessionFailsClosedWhenMFAStatusUnavailable(t *testing.T) {
	issuer := &fakeTokenIssuer{}
	session := newLoginSession(&config.Config{}, issuer, fakeMFAGate{err: errors.New("db down")}, newFakeChallengeStore())

	_, _, err, w := runBegin(t, session, &models.User{ID: uuid.New()})
	if err == nil {
		t.Fatal("begin succeeded although the MFA check failed")
	}
	if issuer.calls != 0 || responseCookies(w)["access_token"] != nil {
		t.Fatal("session issued although the MFA check failed")
	}
}

func TestLoginSessionRejectsBannedUsers(t *testing.T) {
	session := newLoginSession(&config.Config{}, &fakeTokenIssuer{}, fakeMFAGate{required: true}, newFakeChallengeStore())
	if _, _, err, _ := runBegin(t, session, &models.User{ID: uuid.New(), IsBanned: true}); !errors.Is(err, services.ErrUserBanned) {
		t.Fatalf("begin error = %v, want ErrUserBanned", err)
	}
}

type fakeLoginVerifier struct {
	totp          string
	backup        string
	trustedDevice bool
	notEnabled    bool
}

func (f *fakeLoginVerifier) VerifyTOTP(_ context.Context, _ uuid.UUID, code string, _, _ *string) error {
	if f.notEnabled {
		return services.ErrMFANotEnabled
	}
	if code != f.totp {
		return services.ErrInvalidMFACode
	}
	return nil
}

func (f *fakeLoginVerifier) VerifyBackupCode(_ context.Context, _ uuid.UUID, code string) error {
	if code != f.backup {
		return services.ErrInvalidMFACode
	}
	return nil
}

func (f *fakeLoginVerifier) CreateTrustedDevice(context.Context, uuid.UUID, string, string, *string, *string) error {
	f.trustedDevice = true
	return nil
}

type fakeLoginUsers map[uuid.UUID]*models.User

func (f fakeLoginUsers) GetUserByID(_ context.Context, id uuid.UUID) (*models.User, error) {
	user, ok := f[id]
	if !ok {
		return nil, errors.New("not found")
	}
	return user, nil
}

type verifyLoginFixture struct {
	handler  *MFAHandler
	store    *fakeChallengeStore
	verifier *fakeLoginVerifier
	issuer   *fakeTokenIssuer
	user     *models.User
	token    string
}

func newVerifyLoginFixture(t *testing.T) *verifyLoginFixture {
	t.Helper()
	user := &models.User{ID: uuid.New(), Username: "mfa-user"}
	store := newFakeChallengeStore()
	token, _ := store.Create(context.Background(), user.ID)
	issuer := &fakeTokenIssuer{}
	verifier := &fakeLoginVerifier{totp: "123456", backup: "ABCD2345"}
	handler := &MFAHandler{
		cfg:           &config.Config{},
		loginVerifier: verifier,
		users:         fakeLoginUsers{user.ID: user},
		challenges:    store,
		session:       newLoginSession(&config.Config{}, issuer, nil, store),
	}
	return &verifyLoginFixture{handler: handler, store: store, verifier: verifier, issuer: issuer, user: user, token: token}
}

func (f *verifyLoginFixture) post(t *testing.T, token, body string) *httptest.ResponseRecorder {
	t.Helper()
	gin.SetMode(gin.TestMode)
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(http.MethodPost, "/api/v1/auth/mfa/verify-login", bytes.NewBufferString(body))
	c.Request.Header.Set("Content-Type", "application/json")
	if token != "" {
		c.Request.AddCookie(&http.Cookie{Name: mfaChallengeCookie, Value: token})
	}
	f.handler.VerifyLogin(c)
	return w
}

func TestVerifyLoginRequiresAChallenge(t *testing.T) {
	f := newVerifyLoginFixture(t)
	if w := f.post(t, "", `{"code":"123456"}`); w.Code != http.StatusUnauthorized {
		t.Fatalf("no cookie: status %d", w.Code)
	}
	w := f.post(t, "challenge-unknown", `{"code":"123456"}`)
	if w.Code != http.StatusUnauthorized {
		t.Fatalf("unknown challenge: status %d", w.Code)
	}
	if c := responseCookies(w)[mfaChallengeCookie]; c == nil || c.MaxAge >= 0 {
		t.Fatalf("stale challenge cookie not cleared: %+v", c)
	}
	if f.issuer.calls != 0 {
		t.Fatal("session issued without a valid challenge")
	}
}

func TestVerifyLoginIssuesSessionForValidCode(t *testing.T) {
	f := newVerifyLoginFixture(t)
	w := f.post(t, f.token, `{"code":"123 456","trust_device":true}`)
	if w.Code != http.StatusOK {
		t.Fatalf("status %d: %s", w.Code, w.Body.String())
	}
	cookies := responseCookies(w)
	if access := cookies["access_token"]; access == nil || access.Value != "access-"+f.user.ID.String() {
		t.Fatalf("access cookie = %+v", access)
	}
	if c := cookies[mfaChallengeCookie]; c == nil || c.MaxAge >= 0 {
		t.Fatalf("challenge cookie not cleared: %+v", c)
	}
	if !f.verifier.trustedDevice {
		t.Fatal("trusted device not recorded")
	}
	// The challenge completes only one sign-in.
	if w := f.post(t, f.token, `{"code":"123456"}`); w.Code != http.StatusUnauthorized {
		t.Fatalf("reused challenge: status %d", w.Code)
	}
	if f.issuer.calls != 1 {
		t.Fatalf("issued %d sessions, want 1", f.issuer.calls)
	}
}

func TestVerifyLoginAcceptsBackupCodes(t *testing.T) {
	f := newVerifyLoginFixture(t)
	if w := f.post(t, f.token, `{"code":"abcd-2345"}`); w.Code != http.StatusOK {
		t.Fatalf("backup code: status %d: %s", w.Code, w.Body.String())
	}
}

func TestVerifyLoginLimitsWrongCodesPerChallenge(t *testing.T) {
	f := newVerifyLoginFixture(t)
	for i := 1; i < f.store.max; i++ {
		if w := f.post(t, f.token, `{"code":"000000"}`); w.Code != http.StatusBadRequest {
			t.Fatalf("wrong code %d: status %d", i, w.Code)
		}
	}
	w := f.post(t, f.token, `{"code":"000000"}`)
	if w.Code != http.StatusTooManyRequests {
		t.Fatalf("final wrong code: status %d", w.Code)
	}
	// Even the right code no longer works for an exhausted challenge.
	if w := f.post(t, f.token, `{"code":"123456"}`); w.Code != http.StatusUnauthorized {
		t.Fatalf("correct code after exhaustion: status %d", w.Code)
	}
	if f.issuer.calls != 0 {
		t.Fatal("session issued after wrong codes")
	}
}

func TestVerifyLoginRestartsWhenMFAWasDisabled(t *testing.T) {
	f := newVerifyLoginFixture(t)
	f.verifier.notEnabled = true
	if w := f.post(t, f.token, `{"code":"123456"}`); w.Code != http.StatusUnauthorized {
		t.Fatalf("status %d", w.Code)
	}
	if _, ok := f.store.byToken[f.token]; ok {
		t.Fatal("challenge kept after MFA was disabled")
	}
	if f.issuer.calls != 0 {
		t.Fatal("session issued without MFA verification")
	}
}
