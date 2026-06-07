package services

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"sync"
	"time"

	"github.com/google/uuid"
	"github.com/gorilla/websocket"
	"git.subcult.tv/subculture-collective/clpr/internal/models"
	"git.subcult.tv/subculture-collective/clpr/internal/repository"
)

// SimpleRateLimiter provides a simple in-memory rate limiter
type SimpleRateLimiter struct {
	requests sync.Map // map[string]*watchPartyRateLimitWindow
	limit    int
	window   time.Duration
}

type watchPartyRateLimitWindow struct {
	timestamps []time.Time
	mu         sync.Mutex
}

// NewSimpleRateLimiter creates a new rate limiter
func NewSimpleRateLimiter(limit int, window time.Duration) *SimpleRateLimiter {
	return &SimpleRateLimiter{
		limit:  limit,
		window: window,
	}
}

// Allow checks if a request should be allowed
func (r *SimpleRateLimiter) Allow(key string) bool {
	now := time.Now()
	cutoff := now.Add(-r.window)

	val, _ := r.requests.LoadOrStore(key, &watchPartyRateLimitWindow{
		timestamps: make([]time.Time, 0),
	})
	w := val.(*watchPartyRateLimitWindow)

	w.mu.Lock()
	defer w.mu.Unlock()

	// Remove old timestamps
	valid := make([]time.Time, 0, len(w.timestamps))
	for _, ts := range w.timestamps {
		if ts.After(cutoff) {
			valid = append(valid, ts)
		}
	}
	w.timestamps = valid

	// Check limit
	if len(w.timestamps) >= r.limit {
		return false
	}

	// Add current timestamp
	w.timestamps = append(w.timestamps, now)
	return true
}

// WatchPartyHub manages WebSocket connections for a single watch party
type WatchPartyHub struct {
	PartyID          uuid.UUID
	watchPartyRepo   *repository.WatchPartyRepository
	Clients          map[uuid.UUID]*WatchPartyClient
	Broadcast        chan *models.WatchPartySyncEvent
	Register         chan *WatchPartyClient
	Unregister       chan *WatchPartyClient
	mutex            sync.RWMutex
	stopChan         chan struct{}
	wg               sync.WaitGroup
	chatRateLimiter  RateLimiter // Distributed rate limiter: 10 messages per minute per user
	reactRateLimiter RateLimiter // Distributed rate limiter: 30 reactions per minute per user
}

// WatchPartyClient represents a connected client
type WatchPartyClient struct {
	Hub       *WatchPartyHub
	Conn      *websocket.Conn
	UserID    uuid.UUID
	Role      string
	Send      chan []byte
	User      *models.User
	closeOnce sync.Once
}

// WatchPartyHubManager manages multiple watch party hubs
type WatchPartyHubManager struct {
	hubs             map[uuid.UUID]*WatchPartyHub
	mutex            sync.RWMutex
	watchPartyRepo   *repository.WatchPartyRepository
	chatRateLimiter  RateLimiter
	reactRateLimiter RateLimiter
}

// NewWatchPartyHubManager creates a new hub manager with distributed rate limiting
func NewWatchPartyHubManager(watchPartyRepo *repository.WatchPartyRepository, chatLimiter, reactLimiter RateLimiter) *WatchPartyHubManager {
	return &WatchPartyHubManager{
		hubs:             make(map[uuid.UUID]*WatchPartyHub),
		watchPartyRepo:   watchPartyRepo,
		chatRateLimiter:  chatLimiter,
		reactRateLimiter: reactLimiter,
	}
}

// GetOrCreateHub gets an existing hub or creates a new one
func (m *WatchPartyHubManager) GetOrCreateHub(partyID uuid.UUID) *WatchPartyHub {
	m.mutex.Lock()
	defer m.mutex.Unlock()

	if hub, exists := m.hubs[partyID]; exists {
		return hub
	}

	hub := &WatchPartyHub{
		PartyID:          partyID,
		watchPartyRepo:   m.watchPartyRepo,
		Clients:          make(map[uuid.UUID]*WatchPartyClient),
		Broadcast:        make(chan *models.WatchPartySyncEvent, 256),
		Register:         make(chan *WatchPartyClient),
		Unregister:       make(chan *WatchPartyClient),
		stopChan:         make(chan struct{}),
		chatRateLimiter:  m.chatRateLimiter,  // Use shared distributed rate limiter
		reactRateLimiter: m.reactRateLimiter, // Use shared distributed rate limiter
	}

	m.hubs[partyID] = hub
	hub.wg.Add(1)
	go hub.Run()

	return hub
}

// RemoveHub removes a hub when it's no longer needed
func (m *WatchPartyHubManager) RemoveHub(partyID uuid.UUID) {
	m.mutex.Lock()
	hub, exists := m.hubs[partyID]
	if !exists {
		m.mutex.Unlock()
		return
	}
	delete(m.hubs, partyID)
	m.mutex.Unlock()

	// Signal hub to stop and wait for it to finish
	close(hub.stopChan)
	hub.wg.Wait()
}

// Run starts the hub's main loop
func (h *WatchPartyHub) Run() {
	defer h.wg.Done()
	defer func() {
		// Clean up all clients when hub stops
		h.mutex.Lock()
		for _, client := range h.Clients {
			client.closeOnce.Do(func() {
				close(client.Send)
			})
		}
		h.mutex.Unlock()
	}()

	for {
		select {
		case client := <-h.Register:
			h.mutex.Lock()
			h.Clients[client.UserID] = client
			h.mutex.Unlock()

			// Send participant-joined event to all other clients
			h.broadcastParticipantEvent("participant-joined", client.UserID, client.User, client.Role)

		case client := <-h.Unregister:
			h.mutex.Lock()
			if _, ok := h.Clients[client.UserID]; ok {
				delete(h.Clients, client.UserID)
				client.closeOnce.Do(func() {
					close(client.Send)
				})
			}
			h.mutex.Unlock()

			// Send participant-left event to remaining clients
			h.broadcastParticipantEvent("participant-left", client.UserID, client.User, client.Role)

		case event := <-h.Broadcast:
			h.mutex.RLock()
			for _, client := range h.Clients {
				select {
				case client.Send <- mustMarshalJSON(event):
				default:
					// Client's send channel is full, skip this message
					log.Printf("Failed to send to client %s, channel full", client.UserID)
				}
			}
			h.mutex.RUnlock()

		case <-h.stopChan:
			return
		}
	}
}

// broadcastParticipantEvent sends a participant joined/left event
func (h *WatchPartyHub) broadcastParticipantEvent(eventType string, userID uuid.UUID, user *models.User, role string) {
	if user == nil {
		return
	}

	event := &models.WatchPartySyncEvent{
		Type:            eventType,
		PartyID:         h.PartyID.String(),
		ServerTimestamp: time.Now().Unix(),
		Participant: &models.WatchPartyParticipantInfo{
			UserID:      userID,
			DisplayName: user.DisplayName,
			AvatarURL:   user.AvatarURL,
			Role:        role,
		},
	}

	h.Broadcast <- event
}

// ReadPump reads messages from the WebSocket connection
func (c *WatchPartyClient) ReadPump(ctx context.Context) {
	defer func() {
		c.Hub.Unregister <- c
		c.closeOnce.Do(func() {
			c.Conn.Close()
		})
	}()

	c.Conn.SetReadDeadline(time.Now().Add(60 * time.Second))
	c.Conn.SetPongHandler(func(string) error {
		c.Conn.SetReadDeadline(time.Now().Add(60 * time.Second))
		return nil
	})

	for {
		_, message, err := c.Conn.ReadMessage()
		if err != nil {
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
				log.Printf("WebSocket error: %v", err)
			}
			break
		}

		var cmd models.WatchPartyCommand
		if err := json.Unmarshal(message, &cmd); err != nil {
			log.Printf("Failed to unmarshal command: %v", err)
			continue
		}

		// Allow chat, reaction, typing, and sync-request for all users
		// Only host/co-host can control playback
		if c.Role != "host" && c.Role != "co-host" {
			if cmd.Type != "sync-request" && cmd.Type != "chat" && cmd.Type != "reaction" && cmd.Type != "typing" {
				continue
			}
		}

		// Handle command
		c.handleCommand(ctx, &cmd)
	}
}

// handleCommand processes a command from a client
func (c *WatchPartyClient) handleCommand(ctx context.Context, cmd *models.WatchPartyCommand) {
	partyID, err := uuid.Parse(cmd.PartyID)
	if err != nil {
		log.Printf("Invalid party ID: %v", err)
		return
	}

	var event *models.WatchPartySyncEvent

	switch cmd.Type {
	case "play":
		// Get current party state
		party, err := c.Hub.watchPartyRepo.GetByID(ctx, partyID)
		if err != nil {
			log.Printf("Failed to get party: %v", err)
			return
		}

		// Update database
		err = c.Hub.watchPartyRepo.UpdatePlaybackState(ctx, partyID, true, party.CurrentPositionSeconds)
		if err != nil {
			log.Printf("Failed to update playback state: %v", err)
			return
		}

		event = &models.WatchPartySyncEvent{
			Type:            "play",
			PartyID:         cmd.PartyID,
			Position:        party.CurrentPositionSeconds,
			IsPlaying:       true,
			ServerTimestamp: time.Now().Unix(),
		}

	case "pause":
		party, err := c.Hub.watchPartyRepo.GetByID(ctx, partyID)
		if err != nil {
			log.Printf("Failed to get party: %v", err)
			return
		}

		err = c.Hub.watchPartyRepo.UpdatePlaybackState(ctx, partyID, false, party.CurrentPositionSeconds)
		if err != nil {
			log.Printf("Failed to update playback state: %v", err)
			return
		}

		event = &models.WatchPartySyncEvent{
			Type:            "pause",
			PartyID:         cmd.PartyID,
			Position:        party.CurrentPositionSeconds,
			IsPlaying:       false,
			ServerTimestamp: time.Now().Unix(),
		}

	case "seek":
		if cmd.Position == nil {
			return
		}

		// Validate position is not negative
		if *cmd.Position < 0 {
			log.Printf("Invalid seek position (negative): %d", *cmd.Position)
			return
		}

		err := c.Hub.watchPartyRepo.UpdatePlaybackState(ctx, partyID, false, *cmd.Position)
		if err != nil {
			log.Printf("Failed to update playback state: %v", err)
			return
		}

		event = &models.WatchPartySyncEvent{
			Type:            "seek",
			PartyID:         cmd.PartyID,
			Position:        *cmd.Position,
			IsPlaying:       false,
			ServerTimestamp: time.Now().Unix(),
		}

	case "skip":
		if cmd.ClipID == nil {
			return
		}

		err := c.Hub.watchPartyRepo.UpdateCurrentClip(ctx, partyID, *cmd.ClipID, 0)
		if err != nil {
			log.Printf("Failed to skip clip: %v", err)
			return
		}

		event = &models.WatchPartySyncEvent{
			Type:            "skip",
			PartyID:         cmd.PartyID,
			ClipID:          cmd.ClipID,
			Position:        0,
			IsPlaying:       true,
			ServerTimestamp: time.Now().Unix(),
		}

	case "sync-request":
		// Send current party state to requesting client
		party, err := c.Hub.watchPartyRepo.GetByID(ctx, partyID)
		if err != nil {
			log.Printf("Failed to get party: %v", err)
			return
		}

		event = &models.WatchPartySyncEvent{
			Type:            "sync",
			PartyID:         cmd.PartyID,
			ClipID:          party.CurrentClipID,
			Position:        party.CurrentPositionSeconds,
			IsPlaying:       party.IsPlaying,
			ServerTimestamp: time.Now().Unix(),
		}

		// Send only to requesting client
		c.Send <- mustMarshalJSON(event)
		return

	case "chat":
		// Rate limit check - 10 messages per minute
		rateLimitKey := fmt.Sprintf("chat:%s:%s", partyID.String(), c.UserID.String())
		allowed, err := c.Hub.chatRateLimiter.Allow(ctx, rateLimitKey)
		if err != nil {
			// Log infrastructure errors but don't expose to client
			// This could indicate Redis connectivity issues
			log.Printf("Rate limiter error for user %s (falling back to deny): %v", c.UserID, err)
			// Fail closed: deny the request on infrastructure errors
			return
		}
		if !allowed {
			log.Printf("Chat rate limit exceeded for user %s", c.UserID)
			return
		}

		// Validate message
		if cmd.Message == "" || len(cmd.Message) > 1000 {
			log.Printf("Invalid message length: %d", len(cmd.Message))
			return
		}

		// Create message in database
		msg := &models.WatchPartyMessage{
			ID:           uuid.New(),
			WatchPartyID: partyID,
			UserID:       c.UserID,
			Message:      cmd.Message,
		}

		if err := c.Hub.watchPartyRepo.CreateMessage(ctx, msg); err != nil {
			log.Printf("Failed to create message: %v", err)
			return
		}

		// Add user info for broadcast
		msg.Username = c.User.Username
		msg.DisplayName = c.User.DisplayName
		msg.AvatarURL = c.User.AvatarURL

		event = &models.WatchPartySyncEvent{
			Type:            "chat_message",
			PartyID:         cmd.PartyID,
			ServerTimestamp: time.Now().Unix(),
			ChatMessage:     msg,
		}

	case "reaction":
		// Rate limit check - 30 reactions per minute
		rateLimitKey := fmt.Sprintf("reaction:%s:%s", partyID.String(), c.UserID.String())
		allowed, err := c.Hub.reactRateLimiter.Allow(ctx, rateLimitKey)
		if err != nil {
			// Log infrastructure errors but don't expose to client
			// This could indicate Redis connectivity issues
			log.Printf("Rate limiter error for user %s (falling back to deny): %v", c.UserID, err)
			// Fail closed: deny the request on infrastructure errors
			return
		}
		if !allowed {
			log.Printf("Reaction rate limit exceeded for user %s", c.UserID)
			return
		}

		// Validate emoji
		if cmd.Emoji == "" || len(cmd.Emoji) > 10 {
			log.Printf("Invalid emoji length: %d", len(cmd.Emoji))
			return
		}

		// Create reaction in database
		reaction := &models.WatchPartyReaction{
			ID:             uuid.New(),
			WatchPartyID:   partyID,
			UserID:         c.UserID,
			Emoji:          cmd.Emoji,
			VideoTimestamp: cmd.VideoTimestamp,
		}

		if err := c.Hub.watchPartyRepo.CreateReaction(ctx, reaction); err != nil {
			log.Printf("Failed to create reaction: %v", err)
			return
		}

		// Add username for broadcast
		reaction.Username = c.User.Username

		event = &models.WatchPartySyncEvent{
			Type:            "reaction",
			PartyID:         cmd.PartyID,
			ServerTimestamp: time.Now().Unix(),
			Reaction:        reaction,
		}

	case "typing":
		// Broadcast typing indicator (no rate limit, no persistence)
		event = &models.WatchPartySyncEvent{
			Type:            "typing",
			PartyID:         cmd.PartyID,
			UserID:          &c.UserID,
			IsTyping:        cmd.IsTyping,
			ServerTimestamp: time.Now().Unix(),
		}
	}

	if event != nil {
		c.Hub.Broadcast <- event
	}
}

// WritePump writes messages to the WebSocket connection
func (c *WatchPartyClient) WritePump() {
	ticker := time.NewTicker(54 * time.Second)
	defer func() {
		ticker.Stop()
		c.closeOnce.Do(func() {
			c.Conn.Close()
		})
	}()

	for {
		select {
		case message, ok := <-c.Send:
			c.Conn.SetWriteDeadline(time.Now().Add(10 * time.Second))
			if !ok {
				c.Conn.WriteMessage(websocket.CloseMessage, []byte{})
				return
			}

			// Send each message as a separate WebSocket frame
			if err := c.Conn.WriteMessage(websocket.TextMessage, message); err != nil {
				return
			}

		case <-ticker.C:
			c.Conn.SetWriteDeadline(time.Now().Add(10 * time.Second))
			if err := c.Conn.WriteMessage(websocket.PingMessage, nil); err != nil {
				return
			}
		}
	}
}

// mustMarshalJSON marshals v to JSON or panics
func mustMarshalJSON(v interface{}) []byte {
	b, err := json.Marshal(v)
	if err != nil {
		panic(err)
	}
	return b
}
