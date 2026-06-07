package websocket

import (
	"log"
	"net/http"
	"sync"

	"github.com/google/uuid"
	"github.com/gorilla/websocket"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/redis/go-redis/v9"
	"git.subcult.tv/subculture-collective/clpr/config"
)

// Server represents the WebSocket server
type Server struct {
	DB             *pgxpool.Pool
	Redis          *redis.Client
	Upgrader       websocket.Upgrader
	Hubs           map[string]*ChannelHub
	HubsMux        sync.RWMutex
	shutdownOnce   sync.Once
	allowedOrigins []string
}

// NewServer creates a new WebSocket server
func NewServer(db *pgxpool.Pool, redisClient *redis.Client, cfg *config.WebSocketConfig) *Server {
	// Validate allowed origins configuration
	validateAllowedOrigins(cfg.AllowedOrigins)

	server := &Server{
		DB:             db,
		Redis:          redisClient,
		allowedOrigins: cfg.AllowedOrigins,
		Upgrader: websocket.Upgrader{
			ReadBufferSize:  1024,
			WriteBufferSize: 1024,
			CheckOrigin: func(r *http.Request) bool {
				// Get the origin from the request
				origin := r.Header.Get("Origin")
				if origin == "" {
					// No origin header, reject for security
					log.Printf("WebSocket connection rejected: no Origin header")
					return false
				}

				// Check if origin is allowed
				allowed := isOriginAllowed(origin, cfg.AllowedOrigins)
				if !allowed {
					// Log rejected origins for debugging
					log.Printf("WebSocket connection rejected: origin %s not in allowed list", origin)
				}
				return allowed
			},
		},
		Hubs: make(map[string]*ChannelHub),
	}

	return server
}

// GetOrCreateHub gets an existing hub or creates a new one
func (s *Server) GetOrCreateHub(channelID string) *ChannelHub {
	s.HubsMux.Lock()
	defer s.HubsMux.Unlock()

	hub, exists := s.Hubs[channelID]
	if !exists {
		hub = NewChannelHub(channelID, s.DB, s.Redis)
		s.Hubs[channelID] = hub
		go hub.Run()
		log.Printf("Created new hub for channel: %s", channelID)

		// Update active channels metric
		SetActiveChannels(len(s.Hubs))
	}

	return hub
}

// HandleWebSocket handles WebSocket connection requests
func (s *Server) HandleWebSocket(w http.ResponseWriter, r *http.Request, userID uuid.UUID, username, channelID string) error {
	// Upgrade HTTP connection to WebSocket
	conn, err := s.Upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Printf("Failed to upgrade connection: %v", err)
		return err
	}

	// Get or create hub for the channel
	hub := s.GetOrCreateHub(channelID)

	// Create client
	client := NewChatClient(hub, conn, userID, username)

	// Register client with hub
	hub.Register <- client

	// Start read and write pumps
	go client.WritePump()
	go client.ReadPump()

	return nil
}

// Shutdown gracefully shuts down all hubs
func (s *Server) Shutdown() {
	s.shutdownOnce.Do(func() {
		s.HubsMux.Lock()
		defer s.HubsMux.Unlock()

		log.Println("Shutting down WebSocket server...")

		for channelID, hub := range s.Hubs {
			log.Printf("Shutting down hub for channel: %s", channelID)
			close(hub.Stop)
		}

		// Clear hubs map
		s.Hubs = make(map[string]*ChannelHub)

		log.Println("WebSocket server shutdown complete")
	})
}

// GetStats returns statistics about the WebSocket server
func (s *Server) GetStats() map[string]interface{} {
	s.HubsMux.RLock()
	defer s.HubsMux.RUnlock()

	totalConnections := 0
	channelStats := make(map[string]int)

	for channelID, hub := range s.Hubs {
		count := hub.GetClientCount()
		totalConnections += count
		channelStats[channelID] = count
	}

	return map[string]interface{}{
		"total_connections": totalConnections,
		"active_channels":   len(s.Hubs),
		"channel_stats":     channelStats,
	}
}
