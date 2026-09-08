package twitch

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestFixtureTransportCannotReceiveRealCredentials(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		require.Equal(t, "/helix/clips", r.URL.Path)
		w.WriteHeader(204)
	}))
	defer server.Close()
	for _, target := range []string{"http://example.com", "http://127.0.0.1/path", "https://127.0.0.1", "http://user@127.0.0.1"} {
		_, err := fixtureTransport(target, "test_client_id", "test_client_secret")
		require.Error(t, err)
	}
	_, err := fixtureTransport(server.URL, "real_id", "real_secret")
	require.Error(t, err)
	transport, err := fixtureTransport(server.URL, "test_client_id", "test_client_secret")
	require.NoError(t, err)
	client := http.Client{Transport: transport}
	response, err := client.Get("https://api.twitch.tv/helix/clips")
	require.NoError(t, err)
	defer response.Body.Close()
	require.Equal(t, 204, response.StatusCode)
	_, err = client.Get("https://example.com/helix/clips")
	require.Error(t, err)
}
