package services

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"strconv"
	"strings"
	"sync"
	"time"
)

// YouTube Data API v3 endpoints. Tests point them at a fake server.
const (
	youTubeOAuthEndpoint = "https://oauth2.googleapis.com/token"
	youTubeAPIBase       = "https://www.googleapis.com/youtube/v3"
	youTubeUploadBase    = "https://www.googleapis.com/upload/youtube/v3"
)

// YouTubeVideo is what an upload sets on a video.
type YouTubeVideo struct {
	Title       string
	Description string
	Tags        []string
	CategoryID  string
	// PrivacyStatus is private, unlisted or public. An unaudited API project
	// can only upload private videos.
	PrivacyStatus string
	// PublishAt, with PrivacyStatus private, makes YouTube publish the video
	// at that time.
	PublishAt *time.Time
}

// YouTubeError is an error response from the YouTube Data API.
type YouTubeError struct {
	Status  int
	Reason  string
	Message string
}

func (e *YouTubeError) Error() string {
	return fmt.Sprintf("youtube: HTTP %d %s: %s", e.Status, e.Reason, e.Message)
}

// QuotaExhausted reports whether the error lasts until the daily quota
// resets: the project's quota or the channel's upload limit.
func (e *YouTubeError) QuotaExhausted() bool {
	switch e.Reason {
	case "quotaExceeded", "dailyLimitExceeded", "uploadLimitExceeded":
		return true
	}
	return false
}

// Transient reports whether the same request may succeed shortly.
func (e *YouTubeError) Transient() bool {
	return e.Status >= 500 || e.Status == http.StatusTooManyRequests ||
		e.Reason == "rateLimitExceeded" || e.Reason == "userRateLimitExceeded" || e.Reason == "backendError"
}

// AuthFailed reports whether the channel's authorization is missing,
// revoked or expired, which only the channel owner can fix.
func (e *YouTubeError) AuthFailed() bool {
	switch e.Reason {
	case "invalid_grant", "invalid_client", "unauthorized_client", "authError":
		return true
	}
	return e.Status == http.StatusUnauthorized
}

// IsYouTubeQuotaError reports whether err is a YouTube quota or upload-limit
// error.
func IsYouTubeQuotaError(err error) bool {
	var yt *YouTubeError
	return errors.As(err, &yt) && yt.QuotaExhausted()
}

// YouTubeClient uploads to one channel with an OAuth refresh token (#79).
type YouTubeClient struct {
	HTTP         *http.Client
	ClientID     string
	ClientSecret string
	RefreshToken string

	OAuthURL   string
	APIBase    string
	UploadBase string

	mu     sync.Mutex
	token  string
	expiry time.Time
}

func (c *YouTubeClient) http() *http.Client {
	if c.HTTP != nil {
		return c.HTTP
	}
	return &http.Client{Timeout: 30 * time.Minute}
}

func orDefault(value, fallback string) string {
	if value != "" {
		return value
	}
	return fallback
}

// accessToken returns a cached access token, refreshing it a minute before
// it expires.
func (c *YouTubeClient) accessToken(ctx context.Context) (string, error) {
	c.mu.Lock()
	defer c.mu.Unlock()
	if c.token != "" && time.Now().Before(c.expiry.Add(-time.Minute)) {
		return c.token, nil
	}
	form := url.Values{
		"grant_type":    {"refresh_token"},
		"client_id":     {c.ClientID},
		"client_secret": {c.ClientSecret},
		"refresh_token": {c.RefreshToken},
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, orDefault(c.OAuthURL, youTubeOAuthEndpoint), strings.NewReader(form.Encode()))
	if err != nil {
		return "", err
	}
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	resp, err := c.http().Do(req)
	if err != nil {
		return "", fmt.Errorf("youtube token refresh: %w", err)
	}
	defer resp.Body.Close()
	var body struct {
		AccessToken string `json:"access_token"`
		ExpiresIn   int    `json:"expires_in"`
		Error       string `json:"error"`
	}
	if err := json.NewDecoder(io.LimitReader(resp.Body, 1<<16)).Decode(&body); err != nil {
		return "", fmt.Errorf("youtube token refresh: HTTP %d: %w", resp.StatusCode, err)
	}
	if resp.StatusCode != http.StatusOK || body.AccessToken == "" {
		// The body can hold no token on failure; report only the error code.
		return "", &YouTubeError{Status: resp.StatusCode, Reason: body.Error, Message: "token refresh failed"}
	}
	c.token, c.expiry = body.AccessToken, time.Now().Add(time.Duration(body.ExpiresIn)*time.Second)
	return c.token, nil
}

// do sends one authorized request and decodes a JSON response into out.
func (c *YouTubeClient) do(ctx context.Context, method, target, contentType string, body io.Reader, size int64, out any) (*http.Response, error) {
	token, err := c.accessToken(ctx)
	if err != nil {
		return nil, err
	}
	req, err := http.NewRequestWithContext(ctx, method, target, body)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Authorization", "Bearer "+token)
	if contentType != "" {
		req.Header.Set("Content-Type", contentType)
	}
	if size >= 0 {
		req.ContentLength = size
	}
	resp, err := c.http().Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	raw, _ := io.ReadAll(io.LimitReader(resp.Body, 1<<20))
	if resp.StatusCode >= 300 {
		return resp, parseYouTubeError(resp.StatusCode, raw)
	}
	if out != nil && len(raw) > 0 {
		if err := json.Unmarshal(raw, out); err != nil {
			return resp, fmt.Errorf("youtube: decode response: %w", err)
		}
	}
	return resp, nil
}

func parseYouTubeError(status int, raw []byte) error {
	var body struct {
		Error struct {
			Message string `json:"message"`
			Errors  []struct {
				Reason string `json:"reason"`
			} `json:"errors"`
		} `json:"error"`
	}
	_ = json.Unmarshal(raw, &body)
	e := &YouTubeError{Status: status, Message: body.Error.Message}
	if len(body.Error.Errors) > 0 {
		e.Reason = body.Error.Errors[0].Reason
	}
	if e.Message == "" {
		e.Message = http.StatusText(status)
	}
	return e
}

// UploadVideo uploads a video file with a resumable upload session and
// returns the new video's ID.
func (c *YouTubeClient) UploadVideo(ctx context.Context, path string, v YouTubeVideo) (string, error) {
	f, err := os.Open(path)
	if err != nil {
		return "", err
	}
	defer f.Close()
	info, err := f.Stat()
	if err != nil {
		return "", err
	}
	status := map[string]any{"privacyStatus": v.PrivacyStatus, "selfDeclaredMadeForKids": false}
	if v.PublishAt != nil {
		status["publishAt"] = v.PublishAt.UTC().Format(time.RFC3339)
	}
	meta, err := json.Marshal(map[string]any{
		"snippet": map[string]any{"title": v.Title, "description": v.Description, "tags": v.Tags, "categoryId": v.CategoryID},
		"status":  status,
	})
	if err != nil {
		return "", err
	}
	token, err := c.accessToken(ctx)
	if err != nil {
		return "", err
	}
	start := orDefault(c.UploadBase, youTubeUploadBase) + "/videos?uploadType=resumable&part=snippet,status"
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, start, bytes.NewReader(meta))
	if err != nil {
		return "", err
	}
	req.Header.Set("Authorization", "Bearer "+token)
	req.Header.Set("Content-Type", "application/json; charset=UTF-8")
	req.Header.Set("X-Upload-Content-Type", "video/mp4")
	req.Header.Set("X-Upload-Content-Length", fmt.Sprint(info.Size()))
	resp, err := c.http().Do(req)
	if err != nil {
		return "", fmt.Errorf("youtube: start upload: %w", err)
	}
	raw, _ := io.ReadAll(io.LimitReader(resp.Body, 1<<20))
	resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return "", parseYouTubeError(resp.StatusCode, raw)
	}
	session := resp.Header.Get("Location")
	if session == "" {
		return "", errors.New("youtube: upload session has no Location")
	}
	var video struct {
		ID string `json:"id"`
	}
	if _, err := c.do(ctx, http.MethodPut, session, "video/mp4", f, info.Size(), &video); err != nil {
		return "", err
	}
	if video.ID == "" {
		return "", errors.New("youtube: upload response has no video ID")
	}
	return video.ID, nil
}

// SetThumbnail sets a video's custom thumbnail. Channels must be verified to
// use custom thumbnails.
func (c *YouTubeClient) SetThumbnail(ctx context.Context, videoID, path string) error {
	f, err := os.Open(path)
	if err != nil {
		return err
	}
	defer f.Close()
	info, err := f.Stat()
	if err != nil {
		return err
	}
	_, err = c.do(ctx, http.MethodPost, orDefault(c.UploadBase, youTubeUploadBase)+"/thumbnails/set?videoId="+url.QueryEscape(videoID),
		"image/jpeg", f, info.Size(), nil)
	return err
}

// CreatePlaylist creates a playlist and returns its ID.
func (c *YouTubeClient) CreatePlaylist(ctx context.Context, title, description, privacy string) (string, error) {
	body, err := json.Marshal(map[string]any{
		"snippet": map[string]any{"title": title, "description": description},
		"status":  map[string]any{"privacyStatus": privacy},
	})
	if err != nil {
		return "", err
	}
	var playlist struct {
		ID string `json:"id"`
	}
	if _, err := c.do(ctx, http.MethodPost, orDefault(c.APIBase, youTubeAPIBase)+"/playlists?part=snippet,status",
		"application/json", bytes.NewReader(body), int64(len(body)), &playlist); err != nil {
		return "", err
	}
	if playlist.ID == "" {
		return "", errors.New("youtube: playlist response has no ID")
	}
	return playlist.ID, nil
}

// AddToPlaylist appends a video to a playlist.
func (c *YouTubeClient) AddToPlaylist(ctx context.Context, playlistID, videoID string) error {
	body, err := json.Marshal(map[string]any{
		"snippet": map[string]any{
			"playlistId": playlistID,
			"resourceId": map[string]any{"kind": "youtube#video", "videoId": videoID},
		},
	})
	if err != nil {
		return err
	}
	_, err = c.do(ctx, http.MethodPost, orDefault(c.APIBase, youTubeAPIBase)+"/playlistItems?part=snippet",
		"application/json", bytes.NewReader(body), int64(len(body)), nil)
	return err
}

// VideoStats reads the view count and privacy of up to 50 videos with one
// videos.list call (1 quota unit). Videos YouTube no longer returns, because
// they were deleted or removed, are missing from the result.
func (c *YouTubeClient) VideoStats(ctx context.Context, ids []string) (map[string]YouTubeVideoStats, error) {
	if len(ids) == 0 {
		return map[string]YouTubeVideoStats{}, nil
	}
	if len(ids) > 50 {
		return nil, errors.New("youtube: videos.list takes at most 50 IDs")
	}
	target := orDefault(c.APIBase, youTubeAPIBase) + "/videos?part=statistics,status&maxResults=50&id=" + url.QueryEscape(strings.Join(ids, ","))
	var body struct {
		Items []struct {
			ID         string `json:"id"`
			Statistics struct {
				ViewCount string `json:"viewCount"`
			} `json:"statistics"`
			Status struct {
				PrivacyStatus string `json:"privacyStatus"`
			} `json:"status"`
		} `json:"items"`
	}
	if _, err := c.do(ctx, http.MethodGet, target, "", nil, -1, &body); err != nil {
		return nil, err
	}
	out := make(map[string]YouTubeVideoStats, len(body.Items))
	for _, item := range body.Items {
		// viewCount is a decimal string; it is absent when the owner hides it.
		views, _ := strconv.ParseInt(item.Statistics.ViewCount, 10, 64)
		out[item.ID] = YouTubeVideoStats{ViewCount: views, PrivacyStatus: item.Status.PrivacyStatus}
	}
	return out, nil
}

// YouTubeVideoStats is one video's view count and privacy.
type YouTubeVideoStats struct {
	ViewCount     int64
	PrivacyStatus string
}
