package websocket

import (
	"encoding/json"
	"sync"
	"testing"
	"time"

	"github.com/google/uuid"
)

func newSafetyTestClient(hub *ChannelHub) *ChatClient {
	return &ChatClient{
		Hub:      hub,
		UserID:   uuid.New(),
		Username: "tester",
		Send:     make(chan []byte, 4),
	}
}

// A message whose channel_id differs from the connection's channel must be
// rejected before anything else happens (regression: client-chosen channel
// bypassed the access and ban checks done for the URL channel).
func TestHandleChatMessage_RejectsMismatchedChannel(t *testing.T) {
	hub := NewChannelHub(uuid.New().String(), nil, nil)
	client := newSafetyTestClient(hub)
	content := "hello"

	client.handleChatMessage(&ClientMessage{
		Type:      MessageTypeMessage,
		ChannelID: uuid.New().String(),
		Content:   &content,
	})

	select {
	case raw := <-client.Send:
		var msg ServerMessage
		if err := json.Unmarshal(raw, &msg); err != nil {
			t.Fatalf("unmarshal: %v", err)
		}
		if msg.Type != MessageTypeError || msg.Error == nil || *msg.Error != "Channel ID does not match this connection" {
			t.Fatalf("expected channel mismatch error, got %+v", msg)
		}
	default:
		t.Fatal("expected an error reply")
	}
	if len(hub.Broadcast) != 0 {
		t.Fatal("mismatched message must not be broadcast")
	}
}

func TestHandleTypingIndicator_UsesHubChannel(t *testing.T) {
	hub := NewChannelHub(uuid.New().String(), nil, nil)
	client := newSafetyTestClient(hub)
	hub.Clients[client] = true

	client.handleTypingIndicator(&ClientMessage{Type: MessageTypeTyping, ChannelID: uuid.New().String()})

	var msg ServerMessage
	if err := json.Unmarshal(<-client.Send, &msg); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}
	if msg.ChannelID != hub.ID {
		t.Fatalf("typing indicator channel = %s, want hub channel %s", msg.ChannelID, hub.ID)
	}
}

// Regression: sendMessageHistory runs in its own goroutine and could send on
// client.Send after handleUnregister/shutdown closed it, panicking the process.
func TestChatClient_SendAfterCloseDoesNotPanic(t *testing.T) {
	client := newSafetyTestClient(NewChannelHub(uuid.New().String(), nil, nil))

	var wg sync.WaitGroup
	for i := 0; i < 8; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for j := 0; j < 100; j++ {
				client.trySend([]byte("x"))
				select {
				case <-client.Send:
				default:
				}
			}
		}()
	}
	client.closeSend()
	client.closeSend() // idempotent
	wg.Wait()

	if client.trySend([]byte("late")) {
		t.Fatal("trySend must fail after closeSend")
	}
	client.sendError("late error") // must not panic
}

// Regression: handleRegister/handleUnregister ran on the hub goroutine and
// performed a blocking send on the hub's own Broadcast channel, deadlocking
// once the buffer was full.
func TestChannelHub_RegisterUnregisterDoNotBlockOnFullBroadcast(t *testing.T) {
	hub := NewChannelHub(uuid.New().String(), nil, nil)
	for i := 0; i < cap(hub.Broadcast); i++ {
		hub.Broadcast <- []byte("filler")
	}
	observer := newSafetyTestClient(hub)
	hub.Clients[observer] = true
	client := newSafetyTestClient(hub)

	done := make(chan struct{})
	go func() {
		hub.handleRegister(client)
		hub.handleUnregister(client)
		close(done)
	}()

	select {
	case <-done:
	case <-time.After(2 * time.Second):
		t.Fatal("register/unregister blocked on a full Broadcast channel")
	}
	if len(observer.Send) != 2 {
		t.Fatalf("observer should receive joined and left presence inline, got %d messages", len(observer.Send))
	}
}

func TestChatClient_LeaveHubReturnsAfterHubStopped(t *testing.T) {
	hub := &ChannelHub{ID: uuid.New().String(), Unregister: make(chan *ChatClient), Stop: make(chan struct{})}
	close(hub.Stop)
	client := newSafetyTestClient(hub)

	done := make(chan struct{})
	go func() {
		client.leaveHub()
		close(done)
	}()
	select {
	case <-done:
	case <-time.After(2 * time.Second):
		t.Fatal("leaveHub blocked after the hub stopped")
	}
}
