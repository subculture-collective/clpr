package websocket

import (
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"git.subcult.tv/subculture-collective/clpr/config"
)

func TestNewServer(t *testing.T) {
	cfg := &config.WebSocketConfig{
		AllowedOrigins: []string{"http://localhost:5173"},
	}
	server := NewServer(nil, nil, cfg)

	assert.NotNil(t, server)
	assert.NotNil(t, server.Hubs)
	assert.NotNil(t, server.Upgrader)
	assert.Equal(t, 1024, server.Upgrader.ReadBufferSize)
	assert.Equal(t, 1024, server.Upgrader.WriteBufferSize)
	assert.Equal(t, []string{"http://localhost:5173"}, server.allowedOrigins)
}

func TestServer_GetOrCreateHub(t *testing.T) {
	cfg := &config.WebSocketConfig{
		AllowedOrigins: []string{"http://localhost:5173"},
	}
	server := NewServer(nil, nil, cfg)
	channelID := uuid.New().String()

	// Manually create a hub without starting goroutine
	hub := &ChannelHub{
		ID:         channelID,
		Clients:    make(map[*ChatClient]bool),
		Broadcast:  make(chan []byte, 256),
		Register:   make(chan *ChatClient),
		Unregister: make(chan *ChatClient),
		Stop:       make(chan struct{}),
	}

	server.HubsMux.Lock()
	server.Hubs[channelID] = hub
	server.HubsMux.Unlock()

	// Second call should return the same hub
	hub2 := server.GetOrCreateHub(channelID)
	assert.Equal(t, hub, hub2)

	// Verify only one hub exists
	server.HubsMux.RLock()
	hubCount := len(server.Hubs)
	server.HubsMux.RUnlock()
	assert.Equal(t, 1, hubCount)
}

func TestServer_GetStats(t *testing.T) {
	cfg := &config.WebSocketConfig{
		AllowedOrigins: []string{"http://localhost:5173"},
	}
	server := NewServer(nil, nil, cfg)

	// Initially no connections
	stats := server.GetStats()
	assert.Equal(t, 0, stats["total_connections"])
	assert.Equal(t, 0, stats["active_channels"])

	// Manually create a hub without starting goroutines
	channelID := uuid.New().String()
	hub := &ChannelHub{
		ID:      channelID,
		Clients: make(map[*ChatClient]bool),
	}

	server.HubsMux.Lock()
	server.Hubs[channelID] = hub
	server.HubsMux.Unlock()

	client1 := &ChatClient{UserID: uuid.New()}
	client2 := &ChatClient{UserID: uuid.New()}

	hub.Mutex.Lock()
	hub.Clients[client1] = true
	hub.Clients[client2] = true
	hub.Mutex.Unlock()

	// Check stats again
	stats = server.GetStats()
	assert.Equal(t, 2, stats["total_connections"])
	assert.Equal(t, 1, stats["active_channels"])

	channelStats := stats["channel_stats"].(map[string]int)
	assert.Equal(t, 2, channelStats[channelID])
}

func TestServer_GetStats_MultipleChannels(t *testing.T) {
	cfg := &config.WebSocketConfig{
		AllowedOrigins: []string{"http://localhost:5173"},
	}
	server := NewServer(nil, nil, cfg)

	// Create multiple hubs
	channel1 := uuid.New().String()
	channel2 := uuid.New().String()

	// Manually create hubs without starting goroutines (to avoid nil DB/Redis issues)
	hub1 := &ChannelHub{
		ID:      channel1,
		Clients: make(map[*ChatClient]bool),
	}
	hub2 := &ChannelHub{
		ID:      channel2,
		Clients: make(map[*ChatClient]bool),
	}

	server.HubsMux.Lock()
	server.Hubs[channel1] = hub1
	server.Hubs[channel2] = hub2
	server.HubsMux.Unlock()

	// Add clients to hub1
	hub1.Mutex.Lock()
	hub1.Clients[&ChatClient{UserID: uuid.New()}] = true
	hub1.Clients[&ChatClient{UserID: uuid.New()}] = true
	hub1.Mutex.Unlock()

	// Add clients to hub2
	hub2.Mutex.Lock()
	hub2.Clients[&ChatClient{UserID: uuid.New()}] = true
	hub2.Mutex.Unlock()

	// Check stats
	stats := server.GetStats()
	assert.Equal(t, 3, stats["total_connections"])
	assert.Equal(t, 2, stats["active_channels"])

	channelStats := stats["channel_stats"].(map[string]int)
	assert.Equal(t, 2, channelStats[channel1])
	assert.Equal(t, 1, channelStats[channel2])
}

func TestServer_Shutdown(t *testing.T) {
	cfg := &config.WebSocketConfig{
		AllowedOrigins: []string{"http://localhost:5173"},
	}
	server := NewServer(nil, nil, cfg)

	// Manually create some hubs without starting goroutines
	channel1 := uuid.New().String()
	channel2 := uuid.New().String()

	hub1 := &ChannelHub{
		ID:      channel1,
		Clients: make(map[*ChatClient]bool),
		Stop:    make(chan struct{}),
	}
	hub2 := &ChannelHub{
		ID:      channel2,
		Clients: make(map[*ChatClient]bool),
		Stop:    make(chan struct{}),
	}

	server.HubsMux.Lock()
	server.Hubs[channel1] = hub1
	server.Hubs[channel2] = hub2
	server.HubsMux.Unlock()

	// Add clients
	hub1.Mutex.Lock()
	hub1.Clients[&ChatClient{UserID: uuid.New(), Send: make(chan []byte, 1)}] = true
	hub1.Mutex.Unlock()

	hub2.Mutex.Lock()
	hub2.Clients[&ChatClient{UserID: uuid.New(), Send: make(chan []byte, 1)}] = true
	hub2.Mutex.Unlock()

	// Verify hubs exist
	assert.Equal(t, 2, len(server.Hubs))

	// Shutdown server
	server.Shutdown()

	// Verify hubs are cleared
	server.HubsMux.RLock()
	hubCount := len(server.Hubs)
	server.HubsMux.RUnlock()
	assert.Equal(t, 0, hubCount)
}
