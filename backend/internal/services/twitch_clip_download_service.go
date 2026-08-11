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
)

// TwitchClipDownloadService obtains temporary clip media URLs through Twitch's
// official, broadcaster-authorized Helix endpoint.
type TwitchClipDownloadService struct {
	clientID   string
	baseURL    string
	httpClient *http.Client
}

func NewTwitchClipDownloadService(clientID, baseURL string, httpClient *http.Client) *TwitchClipDownloadService {
	if strings.TrimSpace(baseURL) == "" {
		baseURL = "https://api.twitch.tv/helix"
	}
	if httpClient == nil {
		httpClient = &http.Client{Timeout: 30 * time.Second}
	}
	return &TwitchClipDownloadService{
		clientID:   strings.TrimSpace(clientID),
		baseURL:    strings.TrimRight(baseURL, "/"),
		httpClient: httpClient,
	}
}

func (s *TwitchClipDownloadService) GetDownloadURL(ctx context.Context, broadcasterID, clipID, accessToken string) (string, error) {
	if s == nil || s.clientID == "" {
		return "", fmt.Errorf("Twitch clip download client is not configured")
	}
	broadcasterID = strings.TrimSpace(broadcasterID)
	clipID = strings.TrimSpace(clipID)
	accessToken = strings.TrimSpace(accessToken)
	if broadcasterID == "" || clipID == "" || accessToken == "" {
		return "", fmt.Errorf("broadcaster ID, clip ID, and access token are required")
	}

	query := url.Values{}
	query.Set("broadcaster_id", broadcasterID)
	query.Set("editor_id", broadcasterID)
	query.Add("clip_id", clipID)
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, s.baseURL+"/clips/downloads?"+query.Encode(), nil)
	if err != nil {
		return "", fmt.Errorf("creating Twitch clip download request: %w", err)
	}
	req.Header.Set("Authorization", "Bearer "+accessToken)
	req.Header.Set("Client-Id", s.clientID)

	resp, err := s.httpClient.Do(req)
	if err != nil {
		return "", fmt.Errorf("requesting Twitch clip download URL: %w", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(io.LimitReader(resp.Body, 2048))
		return "", fmt.Errorf("Twitch clip download API returned status %d: %s", resp.StatusCode, strings.TrimSpace(string(body)))
	}

	var payload struct {
		Data []struct {
			ClipID               string  `json:"clip_id"`
			LandscapeDownloadURL *string `json:"landscape_download_url"`
		} `json:"data"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&payload); err != nil {
		return "", fmt.Errorf("decoding Twitch clip download response: %w", err)
	}
	if len(payload.Data) != 1 || payload.Data[0].ClipID != clipID || payload.Data[0].LandscapeDownloadURL == nil {
		return "", fmt.Errorf("Twitch did not return a landscape download URL for clip %s", clipID)
	}
	downloadURL := strings.TrimSpace(*payload.Data[0].LandscapeDownloadURL)
	parsed, err := url.Parse(downloadURL)
	if err != nil || parsed.Scheme != "https" || !isTwitchMediaHost(parsed.Hostname()) {
		return "", fmt.Errorf("Twitch returned an invalid clip media URL")
	}
	return downloadURL, nil
}

func isTwitchMediaHost(host string) bool {
	host = strings.ToLower(strings.TrimSpace(host))
	return host == "twitchcdn.net" || strings.HasSuffix(host, ".twitchcdn.net") ||
		host == "ttvnw.net" || strings.HasSuffix(host, ".ttvnw.net")
}
