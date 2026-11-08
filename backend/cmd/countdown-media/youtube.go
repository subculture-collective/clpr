package main

import (
	"context"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"net"
	"net/http"
	"net/url"
	"os"
	"strconv"
	"strings"
	"time"

	"git.subcult.tv/subculture-collective/clpr/internal/services"
)

// secretEnv reads NAME, or the file named by NAME_FILE, so secrets can come
// from the secret store as files.
func secretEnv(name string) (string, error) {
	if value := os.Getenv(name); value != "" {
		return value, nil
	}
	path := os.Getenv(name + "_FILE")
	if path == "" {
		return "", nil
	}
	raw, err := os.ReadFile(path) // #nosec G304 G703 -- the operator names the secret file
	if err != nil {
		return "", fmt.Errorf("reading %s_FILE: %w", name, err)
	}
	return strings.TrimSpace(string(raw)), nil
}

// youTubeClientFromEnv returns nil when YouTube publishing is not set up.
func youTubeClientFromEnv() (*services.YouTubeClient, error) {
	id := os.Getenv("YOUTUBE_CLIENT_ID")
	secret, err := secretEnv("YOUTUBE_CLIENT_SECRET")
	if err != nil {
		return nil, err
	}
	refresh, err := secretEnv("YOUTUBE_REFRESH_TOKEN")
	if err != nil {
		return nil, err
	}
	if id == "" && secret == "" && refresh == "" {
		return nil, nil
	}
	if id == "" || secret == "" || refresh == "" {
		return nil, errors.New("YouTube publishing needs YOUTUBE_CLIENT_ID, YOUTUBE_CLIENT_SECRET and YOUTUBE_REFRESH_TOKEN")
	}
	return &services.YouTubeClient{ClientID: id, ClientSecret: secret, RefreshToken: refresh}, nil
}

func youTubePrivacyFromEnv() (string, int, error) {
	privacy := envString("YOUTUBE_PRIVACY", services.YouTubePrivacyPrivate)
	switch privacy {
	case services.YouTubePrivacyPrivate, services.YouTubePrivacyUnlisted, services.YouTubePrivacyPublic, services.YouTubePrivacyScheduled:
	default:
		return "", 0, fmt.Errorf("YOUTUBE_PRIVACY must be private, unlisted, public or scheduled, not %q", privacy)
	}
	hour := 15
	if value := os.Getenv("YOUTUBE_PUBLISH_HOUR_UTC"); value != "" {
		n, err := strconv.Atoi(value)
		if err != nil || n < 0 || n > 23 {
			return "", 0, fmt.Errorf("YOUTUBE_PUBLISH_HOUR_UTC must be 0 to 23")
		}
		hour = n
	}
	return privacy, hour, nil
}

const (
	googleAuthURL  = "https://accounts.google.com/o/oauth2/v2/auth"
	googleOAuthExchange = "https://oauth2.googleapis.com/token"
	youTubeScope   = "https://www.googleapis.com/auth/youtube"
)

func randomURLString(n int) (string, error) {
	b := make([]byte, n)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}
	return base64.RawURLEncoding.EncodeToString(b), nil
}

// youTubeAuth runs the OAuth consent for the channel and writes the refresh
// token to a file readable only by this user. The channel owner signs in in
// a browser on the same machine; the token never appears on screen.
func youTubeAuth(args []string) error {
	fs := flag.NewFlagSet("youtube-auth", flag.ContinueOnError)
	out := fs.String("out", "", "file to write the refresh token to (created with mode 0600)")
	listen := fs.String("listen", "127.0.0.1:8085", "loopback address for the OAuth redirect")
	if err := fs.Parse(args); err != nil {
		return err
	}
	clientID := os.Getenv("YOUTUBE_CLIENT_ID")
	clientSecret, err := secretEnv("YOUTUBE_CLIENT_SECRET")
	if err != nil {
		return err
	}
	if *out == "" || clientID == "" || clientSecret == "" {
		return errors.New("usage: YOUTUBE_CLIENT_ID=... YOUTUBE_CLIENT_SECRET_FILE=... countdown-media youtube-auth -out FILE")
	}
	host, _, err := net.SplitHostPort(*listen)
	if err != nil {
		return err
	}
	if ip := net.ParseIP(host); ip == nil || !ip.IsLoopback() {
		return errors.New("-listen must be a loopback address")
	}
	state, err := randomURLString(24)
	if err != nil {
		return err
	}
	verifier, err := randomURLString(48)
	if err != nil {
		return err
	}
	sum := sha256.Sum256([]byte(verifier))
	redirect := "http://" + *listen + "/"
	authURL := googleAuthURL + "?" + url.Values{
		"client_id":             {clientID},
		"redirect_uri":          {redirect},
		"response_type":         {"code"},
		"scope":                 {youTubeScope},
		"access_type":           {"offline"},
		"prompt":                {"consent"},
		"state":                 {state},
		"code_challenge":        {base64.RawURLEncoding.EncodeToString(sum[:])},
		"code_challenge_method": {"S256"},
	}.Encode()

	ln, err := net.Listen("tcp", *listen)
	if err != nil {
		return err
	}
	done := make(chan error, 1)
	server := &http.Server{ReadHeaderTimeout: 10 * time.Second, Handler: http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		q := r.URL.Query()
		if q.Get("state") != state || q.Get("code") == "" {
			http.Error(w, "Sign-in did not complete. Close this tab and run the command again.", http.StatusBadRequest)
			done <- fmt.Errorf("authorization failed: %s", q.Get("error"))
			return
		}
		err := exchangeYouTubeCode(r.Context(), clientID, clientSecret, q.Get("code"), verifier, redirect, *out)
		if err != nil {
			http.Error(w, "The token exchange failed; see the terminal.", http.StatusBadGateway)
		} else {
			fmt.Fprintln(w, "clpr can now upload to this YouTube channel. You can close this tab.")
		}
		done <- err
	})}
	go func() { _ = server.Serve(ln) }()
	defer server.Close()
	fmt.Printf("Open this URL in a browser on this machine and sign in as the channel owner:\n\n%s\n\n", authURL)
	if err := <-done; err != nil {
		return err
	}
	fmt.Printf("Refresh token written to %s. Store it in the secret store and set YOUTUBE_REFRESH_TOKEN_FILE.\n", *out)
	return nil
}

func exchangeYouTubeCode(ctx context.Context, clientID, clientSecret, code, verifier, redirect, out string) error {
	form := url.Values{
		"grant_type":    {"authorization_code"},
		"client_id":     {clientID},
		"client_secret": {clientSecret},
		"code":          {code},
		"code_verifier": {verifier},
		"redirect_uri":  {redirect},
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, googleOAuthExchange, strings.NewReader(form.Encode()))
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	resp, err := (&http.Client{Timeout: 30 * time.Second}).Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	var body struct {
		RefreshToken string `json:"refresh_token"`
		Error        string `json:"error"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&body); err != nil {
		return fmt.Errorf("token exchange: HTTP %d", resp.StatusCode)
	}
	if body.RefreshToken == "" {
		return fmt.Errorf("token exchange returned no refresh token (HTTP %d %s)", resp.StatusCode, body.Error)
	}
	return os.WriteFile(out, []byte(body.RefreshToken+"\n"), 0o600)
}
