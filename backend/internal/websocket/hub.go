package websocket

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"sync"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/redis/go-redis/v9"
)

// ChannelHub maintains the set of active clients and broadcasts messages to them
type ChannelHub struct {
	ID         string
	Clients    map[*ChatClient]bool
	Broadcast  chan []byte
	Register   chan *ChatClient
	Unregister chan *ChatClient
	Mutex      sync.RWMutex
	DB         *pgxpool.Pool
	Redis      *redis.Client
	Stop       chan struct{}
}

// NewChannelHub creates a new channel hub
func NewChannelHub(channelID string, db *pgxpool.Pool, redisClient *redis.Client) *ChannelHub {
	return &ChannelHub{
		ID:         channelID,
		Clients:    make(map[*ChatClient]bool),
		Broadcast:  make(chan []byte, 256),
		Register:   make(chan *ChatClient, 10), // Buffered to prevent blocking
		Unregister: make(chan *ChatClient, 10), // Buffered to prevent blocking
		DB:         db,
		Redis:      redisClient,
		Stop:       make(chan struct{}),
	}
}

// Run starts the hub's main loop
func (h *ChannelHub) Run() {
	ticker := time.NewTicker(30 * time.Second)
	defer ticker.Stop()

	// Subscribe to Redis pub/sub for this channel if Redis is available
	var pubsub *redis.PubSub
	var redisChan <-chan *redis.Message

	if h.Redis != nil {
		// Use a cancellable context for Redis subscription
		ctx, cancel := context.WithCancel(context.Background())
		defer cancel()

		pubsub = h.Redis.Subscribe(ctx, fmt.Sprintf("chat:%s", h.ID))
		defer pubsub.Close()
		redisChan = pubsub.Channel()
	}
	// Note: If Redis is nil, redisChan will be nil. Reading from a nil channel blocks forever,
	// so the select case will never trigger - this is the intended behavior for non-Redis setups.

	for {
		select {
		case client := <-h.Register:
			h.handleRegister(client)

		case client := <-h.Unregister:
			h.handleUnregister(client)

		case message := <-h.Broadcast:
			h.handleBroadcast(message)

		case redisMsg := <-redisChan:
			// Received message from Redis pub/sub (from another instance or this instance)
			if redisMsg != nil {
				h.broadcastToClients([]byte(redisMsg.Payload))
			}

		case <-ticker.C:
			// Periodic health check (can be used for monitoring)

		case <-h.Stop:
			// Graceful shutdown
			h.shutdown()
			return
		}
	}
}

// handleRegister registers a new client
func (h *ChannelHub) handleRegister(client *ChatClient) {
	h.Mutex.Lock()
	h.Clients[client] = true
	clientCount := len(h.Clients)
	h.Mutex.Unlock()

	// Record metrics
	RecordConnection(h.ID)

	log.Printf("Client registered: user_id=%s, channel=%s, total_clients=%d",
		client.UserID, h.ID, clientCount)

	// Broadcast user joined presence
	userIDStr := client.UserID.String()
	presenceType := PresenceTypeJoined
	presence := ServerMessage{
		Type:         MessageTypePresence,
		ChannelID:    h.ID,
		UserID:       &userIDStr,
		Username:     &client.Username,
		PresenceType: &presenceType,
		Timestamp:    timePtr(time.Now()),
	}

	// Fan out inline: Run is the only reader of h.Broadcast, so sending to it
	// from here would deadlock the hub once the buffer is full.
	data, err := json.Marshal(presence)
	if err == nil {
		h.handleBroadcast(data)
	}

}

// handleUnregister unregisters a client
func (h *ChannelHub) handleUnregister(client *ChatClient) {
	h.Mutex.Lock()
	if _, ok := h.Clients[client]; ok {
		delete(h.Clients, client)
		client.closeSend()
	}
	totalClients := len(h.Clients)
	h.Mutex.Unlock()

	// Record metrics
	RecordDisconnection(h.ID)

	log.Printf("Client unregistered: user_id=%s, channel=%s, total_clients=%d",
		client.UserID, h.ID, totalClients)

	// Broadcast user left presence
	userIDStr := client.UserID.String()
	presenceType := PresenceTypeLeft
	presence := ServerMessage{
		Type:         MessageTypePresence,
		ChannelID:    h.ID,
		UserID:       &userIDStr,
		Username:     &client.Username,
		PresenceType: &presenceType,
		Timestamp:    timePtr(time.Now()),
	}

	// Fan out inline (see handleRegister).
	data, err := json.Marshal(presence)
	if err == nil {
		h.handleBroadcast(data)
	}
}

// handleBroadcast broadcasts a message to all clients via Redis Pub/Sub
// The message is published to Redis and will be received by all instances (including this one)
// This ensures consistent message delivery across all instances
func (h *ChannelHub) handleBroadcast(message []byte) {
	start := time.Now()

	// Publish to Redis for all instances (if Redis is available)
	if h.Redis != nil {
		ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
		defer cancel()

		err := h.Redis.Publish(ctx, fmt.Sprintf("chat:%s", h.ID), string(message)).Err()
		if err != nil {
			log.Printf("Failed to publish message to Redis: %v", err)
			RecordError(h.ID, "redis_publish_error")
			// Fallback to local broadcast if Redis fails
			h.broadcastToClients(message)
		}
	} else {
		// No Redis available, broadcast locally only
		h.broadcastToClients(message)
	}

	// Record broadcast duration metric
	duration := time.Since(start).Seconds()
	RecordBroadcastDuration(h.ID, duration)
}

// broadcastToClients sends message to all connected clients
func (h *ChannelHub) broadcastToClients(message []byte) {
	h.Mutex.RLock()
	defer h.Mutex.RUnlock()

	for client := range h.Clients {
		if !client.trySend(message) {
			// Client's send channel is full or closed, skip
			log.Printf("Client send buffer full, skipping message for user_id=%s", client.UserID)
		}
	}
}

// shutdown gracefully shuts down the hub
func (h *ChannelHub) shutdown() {
	h.Mutex.Lock()
	defer h.Mutex.Unlock()

	log.Printf("Shutting down hub for channel: %s", h.ID)

	// Close all client connections
	for client := range h.Clients {
		client.closeSend()
		if client.Conn != nil {
			client.Conn.Close()
		}
	}

	// Clear clients map
	h.Clients = make(map[*ChatClient]bool)
}

// GetClientCount returns the number of connected clients
func (h *ChannelHub) GetClientCount() int {
	h.Mutex.RLock()
	defer h.Mutex.RUnlock()
	return len(h.Clients)
}
