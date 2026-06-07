package websocket

import (
	"context"
	"encoding/json"
	"log"
	"os"
	"strconv"
	"time"

	"github.com/google/uuid"
	"github.com/gorilla/websocket"
	"golang.org/x/time/rate"
)

const (
	// Time allowed to write a message to the peer
	writeWait = 10 * time.Second

	// Time allowed to read the next pong message from the peer
	pongWait = 60 * time.Second

	// Send pings to peer with this period. Must be less than pongWait
	pingPeriod = 54 * time.Second

	// Maximum message size allowed from peer (500 chars as per requirements)
	maxMessageSize = 512 // 500 chars + overhead
)

// NewChatClient creates a new chat client
func NewChatClient(hub *ChannelHub, conn *websocket.Conn, userID uuid.UUID, username string) *ChatClient {
	// Allow disabling rate limiting by leaving env vars unset (integration tests expect
	// no rate limiting in most scenarios). When values are provided, they will be used
	// to construct the limiter.
	limitPerMinute := getEnvInt("CHAT_RATE_LIMIT_PER_MINUTE", 0)
	burst := getEnvInt("CHAT_RATE_LIMIT_BURST", 0)

	var limiter *rate.Limiter
	if limitPerMinute > 0 {
		limiter = rate.NewLimiter(rate.Limit(float64(limitPerMinute)/60.0), burst)
	}

	return &ChatClient{
		Hub:       hub,
		Conn:      conn,
		UserID:    userID,
		Username:  username,
		Send:      make(chan []byte, 256),
		RateLimit: limiter,
	}
}

// ReadPump pumps messages from the WebSocket connection to the hub
func (c *ChatClient) ReadPump() {
	defer func() {
		c.Hub.Unregister <- c
		c.Conn.Close()
	}()

	c.Conn.SetReadLimit(maxMessageSize)
	c.Conn.SetReadDeadline(time.Now().Add(pongWait))
	c.Conn.SetPongHandler(func(string) error {
		c.Conn.SetReadDeadline(time.Now().Add(pongWait))
		return nil
	})

	if c.ReadOnly {
		for {
			if _, _, err := c.Conn.ReadMessage(); err != nil {
				if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
					log.Printf("WebSocket error: user_id=%s, error=%v", c.UserID, err)
				}
				break
			}
		}
		return
	}

	for {
		var msg ClientMessage
		err := c.Conn.ReadJSON(&msg)
		if err != nil {
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
				log.Printf("WebSocket error: user_id=%s, error=%v", c.UserID, err)
			}
			break
		}

		c.handleMessage(&msg)
	}
}

// WritePump pumps messages from the hub to the WebSocket connection
func (c *ChatClient) WritePump() {
	ticker := time.NewTicker(pingPeriod)
	defer func() {
		ticker.Stop()
		c.Conn.Close()
	}()

	for {
		select {
		case message, ok := <-c.Send:
			c.Conn.SetWriteDeadline(time.Now().Add(writeWait))
			if !ok {
				// The hub closed the channel
				c.Conn.WriteMessage(websocket.CloseMessage, []byte{})
				return
			}

			if err := c.Conn.WriteMessage(websocket.TextMessage, message); err != nil {
				return
			}

		case <-ticker.C:
			c.Conn.SetWriteDeadline(time.Now().Add(writeWait))
			if err := c.Conn.WriteMessage(websocket.PingMessage, nil); err != nil {
				return
			}
		}
	}
}

// handleMessage processes incoming messages from the client
func (c *ChatClient) handleMessage(msg *ClientMessage) {
	switch msg.Type {
	case MessageTypeMessage:
		c.handleChatMessage(msg)
	case MessageTypeTyping:
		c.handleTypingIndicator(msg)
	default:
		// Unknown message type
		c.sendError("Unknown message type")
	}
}

// handleChatMessage processes a chat message
func (c *ChatClient) handleChatMessage(msg *ClientMessage) {
	start := time.Now()

	// Check if DB is available (skip if in test mode)
	if c.Hub.DB == nil {
		c.sendError("Database not available")
		return
	}

	// Check rate limit (if configured)
	if c.RateLimit != nil && !c.RateLimit.Allow() {
		RecordRateLimitHit(c.Hub.ID)
		c.sendError("Rate limit exceeded. Maximum 20 messages per minute.")
		return
	}

	// Validate message
	if msg.Content == nil || *msg.Content == "" {
		RecordError(c.Hub.ID, "empty_message")
		c.sendError("Message content cannot be empty")
		return
	}

	if len(*msg.Content) > 500 {
		RecordError(c.Hub.ID, "message_too_long")
		c.sendError("Message content exceeds maximum size of 500 characters")
		return
	}

	// Generate message ID if not provided
	messageID := uuid.New()
	if msg.MessageID != nil {
		// Use client-provided ID for deduplication
		parsedID, err := uuid.Parse(*msg.MessageID)
		if err == nil {
			messageID = parsedID
		}
	}

	channelUUID, err := uuid.Parse(msg.ChannelID)
	if err != nil {
		RecordError(c.Hub.ID, "invalid_channel_id")
		c.sendError("Invalid channel ID")
		return
	}

	// Save message to database
	now := time.Now()
	query := `
		INSERT INTO chat_messages (id, channel_id, user_id, content, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6)
		ON CONFLICT (id) DO NOTHING
	`

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	result, err := c.Hub.DB.Exec(ctx,
		query, messageID, channelUUID, c.UserID, *msg.Content, now, now)
	if err != nil {
		log.Printf("Failed to save message to database: %v", err)
		RecordError(c.Hub.ID, "db_save_error")
		c.sendError("Failed to save message")
		return
	}

	// Deduplicate: if no rows were inserted, the message already exists
	if result.RowsAffected() == 0 {
		return
	}

	// Broadcast message to all clients
	messageIDStr := messageID.String()
	userIDStr := c.UserID.String()

	// Fetch user details for the broadcast (use fresh context)
	var displayName string
	var avatarURL *string
	userQuery := `SELECT display_name, avatar_url FROM users WHERE id = $1`
	userCtx, userCancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer userCancel()

	err = c.Hub.DB.QueryRow(userCtx, userQuery, c.UserID).Scan(&displayName, &avatarURL)
	if err != nil {
		displayName = c.Username
	}

	serverMsg := ServerMessage{
		Type:        MessageTypeMessage,
		ChannelID:   msg.ChannelID,
		UserID:      &userIDStr,
		Username:    &c.Username,
		DisplayName: &displayName,
		AvatarURL:   avatarURL,
		Content:     msg.Content,
		MessageID:   &messageIDStr,
		Timestamp:   &now,
	}

	data, err := json.Marshal(serverMsg)
	if err != nil {
		log.Printf("Failed to marshal message: %v", err)
		RecordError(c.Hub.ID, "marshal_error")
		return
	}

	c.Hub.Broadcast <- data

	// Record metrics
	RecordMessage(c.Hub.ID, MessageTypeMessage)
	latency := time.Since(start).Seconds()
	RecordMessageLatency(c.Hub.ID, latency)
}

func getEnvInt(key string, defaultVal int) int {
	if valStr := os.Getenv(key); valStr != "" {
		if parsed, err := strconv.Atoi(valStr); err == nil && parsed >= 0 {
			return parsed
		}
	}
	return defaultVal
}

// handleTypingIndicator processes a typing indicator
func (c *ChatClient) handleTypingIndicator(msg *ClientMessage) {
	// Typing indicators are not persisted, just broadcasted
	userIDStr := c.UserID.String()
	serverMsg := ServerMessage{
		Type:      MessageTypeTyping,
		ChannelID: msg.ChannelID,
		UserID:    &userIDStr,
		Username:  &c.Username,
		Timestamp: timePtr(time.Now()),
	}

	data, err := json.Marshal(serverMsg)
	if err != nil {
		log.Printf("Failed to marshal typing indicator: %v", err)
		return
	}

	// Broadcast directly to local clients only (no Redis pub/sub for typing)
	c.Hub.broadcastToClients(data)

	// Record metrics
	RecordMessage(c.Hub.ID, MessageTypeTyping)
}

// sendError sends an error message to the client
func (c *ChatClient) sendError(errorMsg string) {
	serverMsg := ServerMessage{
		Type:      MessageTypeError,
		ChannelID: c.Hub.ID,
		Error:     &errorMsg,
		Timestamp: timePtr(time.Now()),
	}

	data, err := json.Marshal(serverMsg)
	if err != nil {
		log.Printf("Failed to marshal error message: %v", err)
		return
	}

	select {
	case c.Send <- data:
	default:
		// Send channel is full, client might be slow
	}
}
