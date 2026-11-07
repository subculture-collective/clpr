package websocket

import (
	"sync"
	"time"

	"github.com/google/uuid"
	"github.com/gorilla/websocket"
)

// ClientMessage represents a message from client to server
type ClientMessage struct {
	Type      string     `json:"type"`       // message, typing, join, leave
	ChannelID string     `json:"channel_id"` // UUID as string
	Content   *string    `json:"content,omitempty"`
	MessageID *string    `json:"message_id,omitempty"` // UUID for deduplication
	Timestamp *time.Time `json:"timestamp,omitempty"`
}

// ServerMessage represents a message from server to client
type ServerMessage struct {
	Type         string     `json:"type"`                    // message, typing, presence, error
	ChannelID    string     `json:"channel_id"`              // UUID as string
	UserID       *string    `json:"user_id,omitempty"`       // UUID as string
	Username     *string    `json:"username,omitempty"`      // Username for display
	DisplayName  *string    `json:"display_name,omitempty"`  // Display name
	AvatarURL    *string    `json:"avatar_url,omitempty"`    // Avatar URL
	Content      *string    `json:"content,omitempty"`       // Message content
	MessageID    *string    `json:"message_id,omitempty"`    // UUID as string
	Timestamp    *time.Time `json:"timestamp,omitempty"`     // Message timestamp
	PresenceType *string    `json:"presence_type,omitempty"` // joined, left
	Error        *string    `json:"error,omitempty"`         // Error message
}

// ChatClient represents a connected WebSocket client
type ChatClient struct {
	Hub      *ChannelHub
	Conn     *websocket.Conn
	UserID   uuid.UUID
	Username string
	Send     chan []byte

	// sendMu guards sendClosed so that writers (hub broadcasts, history
	// goroutines, error replies) never send on a channel that the hub has
	// already closed during unregister or shutdown.
	sendMu     sync.Mutex
	sendClosed bool
}

// trySend queues data for the client without blocking. It returns false when
// the client's send channel is closed or its buffer is full.
func (c *ChatClient) trySend(data []byte) bool {
	c.sendMu.Lock()
	defer c.sendMu.Unlock()
	if c.sendClosed {
		return false
	}
	select {
	case c.Send <- data:
		return true
	default:
		return false
	}
}

// closeSend closes the client's send channel exactly once. It is safe to call
// concurrently with trySend.
func (c *ChatClient) closeSend() {
	c.sendMu.Lock()
	defer c.sendMu.Unlock()
	if !c.sendClosed {
		c.sendClosed = true
		close(c.Send)
	}
}

// MessageType constants
const (
	MessageTypeMessage  = "message"
	MessageTypeTyping   = "typing"
	MessageTypeJoin     = "join"
	MessageTypeLeave    = "leave"
	MessageTypePresence = "presence"
	MessageTypeError    = "error"
)

// PresenceType constants
const (
	PresenceTypeJoined = "joined"
	PresenceTypeLeft   = "left"
)
