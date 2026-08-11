package services

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestTwitchClipDownloadServiceGetsOfficialAuthorizedURL(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "/helix/clips/downloads", r.URL.Path)
		assert.Equal(t, "broadcaster-123", r.URL.Query().Get("broadcaster_id"))
		assert.Equal(t, "broadcaster-123", r.URL.Query().Get("editor_id"))
		assert.Equal(t, []string{"clip-456"}, r.URL.Query()["clip_id"])
		assert.Equal(t, "Bearer streamer-access-token", r.Header.Get("Authorization"))
		assert.Equal(t, "clpr-client-id", r.Header.Get("Client-Id"))

		require.NoError(t, json.NewEncoder(w).Encode(map[string]interface{}{
			"data": []map[string]interface{}{
				{
					"clip_id":                "clip-456",
					"landscape_download_url": "https://production.assets.clips.twitchcdn.net/authorized.mp4",
					"portrait_download_url":  nil,
				},
			},
		}))
	}))
	defer server.Close()

	service := NewTwitchClipDownloadService("clpr-client-id", server.URL+"/helix", server.Client())
	downloadURL, err := service.GetDownloadURL(
		context.Background(), "broadcaster-123", "clip-456", "streamer-access-token",
	)

	require.NoError(t, err)
	assert.Equal(t, "https://production.assets.clips.twitchcdn.net/authorized.mp4", downloadURL)
}
