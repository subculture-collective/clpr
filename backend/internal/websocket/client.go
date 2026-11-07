package websocket

import (
	"log"
	"time"

	"github.com/google/uuid"
	"github.com/gorilla/websocket"
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

// NewChatClient creates a client that receives the hub's broadcasts.
func NewChatClient(hub *ChannelHub, conn *websocket.Conn, userID uuid.UUID, username string) *ChatClient {
	return &ChatClient{
		Hub:      hub,
		Conn:     conn,
		UserID:   userID,
		Username: username,
		Send:     make(chan []byte, 256),
	}
}

// ReadPump reads from the connection until it closes, then leaves the hub.
func (c *ChatClient) ReadPump() {
	defer func() {
		c.leaveHub()
		c.Conn.Close()
	}()

	c.Conn.SetReadLimit(maxMessageSize)
	c.Conn.SetReadDeadline(time.Now().Add(pongWait))
	c.Conn.SetPongHandler(func(string) error {
		c.Conn.SetReadDeadline(time.Now().Add(pongWait))
		return nil
	})

	// Clients only receive. Reading keeps pong handling and close detection
	// working; anything a client sends is discarded.
	for {
		if _, _, err := c.Conn.ReadMessage(); err != nil {
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
				log.Printf("WebSocket error: user_id=%s, error=%v", c.UserID, err)
			}
			break
		}
	}
}

// leaveHub asks the hub to unregister the client. If the hub has already
// stopped, its shutdown has closed every client, so there is nothing to wait for.
func (c *ChatClient) leaveHub() {
	select {
	case c.Hub.Unregister <- c:
	case <-c.Hub.Stop:
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
