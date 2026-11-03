package services

import (
	"context"
	"encoding/json"
	"testing"
	"time"

	"git.subcult.tv/subculture-collective/clpr/internal/models"
	"github.com/google/uuid"
)

func newTestWatchPartyHub() *WatchPartyHub {
	return &WatchPartyHub{
		PartyID:    uuid.New(),
		Clients:    make(map[uuid.UUID]*WatchPartyClient),
		Broadcast:  make(chan *models.WatchPartySyncEvent, 256),
		Register:   make(chan *WatchPartyClient),
		Unregister: make(chan *WatchPartyClient),
		stopChan:   make(chan struct{}),
	}
}

func newTestWatchPartyClient(hub *WatchPartyHub, userID uuid.UUID) *WatchPartyClient {
	return &WatchPartyClient{
		Hub:    hub,
		UserID: userID,
		Role:   "host",
		Send:   make(chan []byte, 256),
		User:   &models.User{ID: userID, DisplayName: "tester"},
	}
}

func isSendClosed(c *WatchPartyClient) bool {
	c.sendMu.Lock()
	defer c.sendMu.Unlock()
	return c.sendClosed
}

// Regression: commands used the client-supplied party_id, so a host of one
// party could control or post into another party.
func TestHandleCommand_RejectsOtherPartyID(t *testing.T) {
	hub := newTestWatchPartyHub()
	client := newTestWatchPartyClient(hub, uuid.New())
	position := 42

	// The repository is nil: if the command were not rejected before touching
	// it, this would panic.
	for _, cmdType := range []string{"seek", "play", "pause", "skip", "chat", "reaction", "typing", "sync-request"} {
		client.handleCommand(context.Background(), &models.WatchPartyCommand{
			Type:     cmdType,
			PartyID:  uuid.New().String(),
			Position: &position,
			Message:  "hi",
			Emoji:    "x",
		})
	}
	client.handleCommand(context.Background(), &models.WatchPartyCommand{Type: "typing", PartyID: "not-a-uuid"})

	if len(hub.Broadcast) != 0 || len(client.Send) != 0 {
		t.Fatal("commands for another party must not produce events")
	}
}

func TestHandleCommand_UsesHubPartyID(t *testing.T) {
	hub := newTestWatchPartyHub()
	client := newTestWatchPartyClient(hub, uuid.New())

	for _, partyID := range []string{hub.PartyID.String(), ""} {
		client.handleCommand(context.Background(), &models.WatchPartyCommand{Type: "typing", PartyID: partyID, IsTyping: true})
		select {
		case event := <-hub.Broadcast:
			if event.PartyID != hub.PartyID.String() {
				t.Fatalf("event party = %s, want %s", event.PartyID, hub.PartyID)
			}
		default:
			t.Fatalf("expected typing event for party_id %q", partyID)
		}
	}
}

// Regression: a second connection from the same user replaced the map entry,
// and the first connection's unregister then deleted and closed the new one.
func TestWatchPartyHub_StaleUnregisterKeepsNewConnection(t *testing.T) {
	hub := newTestWatchPartyHub()
	userID := uuid.New()
	oldTab := newTestWatchPartyClient(hub, userID)
	newTab := newTestWatchPartyClient(hub, userID)

	hub.handleRegister(oldTab)
	hub.handleRegister(newTab)
	if !isSendClosed(oldTab) {
		t.Fatal("displaced connection should be closed")
	}

	hub.handleUnregister(oldTab)
	if hub.Clients[userID] != newTab {
		t.Fatal("stale unregister removed the replacement connection")
	}
	if isSendClosed(newTab) {
		t.Fatal("stale unregister closed the replacement connection")
	}

	// Drain and verify no participant-left was emitted for the stale tab.
	for len(newTab.Send) > 0 {
		var event models.WatchPartySyncEvent
		if err := json.Unmarshal(<-newTab.Send, &event); err != nil {
			t.Fatal(err)
		}
		if event.Type == "participant-left" {
			t.Fatal("stale unregister must not announce the user as having left")
		}
	}

	hub.handleUnregister(newTab)
	if _, ok := hub.Clients[userID]; ok || !isSendClosed(newTab) {
		t.Fatal("current connection should be removed and closed on unregister")
	}
}

// Regression: participant events were sent to h.Broadcast from the hub's own
// goroutine, deadlocking once the buffer was full.
func TestWatchPartyHub_ParticipantEventsDoNotBlockOnFullBroadcast(t *testing.T) {
	hub := newTestWatchPartyHub()
	for i := 0; i < cap(hub.Broadcast); i++ {
		hub.Broadcast <- &models.WatchPartySyncEvent{Type: "filler"}
	}
	observer := newTestWatchPartyClient(hub, uuid.New())
	hub.Clients[observer.UserID] = observer
	joiner := newTestWatchPartyClient(hub, uuid.New())

	done := make(chan struct{})
	go func() {
		hub.handleRegister(joiner)
		hub.handleUnregister(joiner)
		close(done)
	}()
	select {
	case <-done:
	case <-time.After(2 * time.Second):
		t.Fatal("participant events blocked on a full Broadcast channel")
	}
	if len(observer.Send) != 2 {
		t.Fatalf("observer should get joined and left events inline, got %d", len(observer.Send))
	}
}

// Regression: one sync.Once guarded both close(Send) and Conn.Close(), so the
// first to run prevented the other.
func TestWatchPartyClient_SendAndConnCloseAreIndependent(t *testing.T) {
	client := newTestWatchPartyClient(newTestWatchPartyHub(), uuid.New())
	client.closeSend()
	ran := false
	client.connCloseOnce.Do(func() { ran = true })
	if !ran {
		t.Fatal("closing Send must not consume the connection close")
	}

	other := newTestWatchPartyClient(newTestWatchPartyHub(), uuid.New())
	other.closeConn() // nil Conn is tolerated
	other.closeSend()
	if !isSendClosed(other) {
		t.Fatal("closing the connection must not prevent closing Send")
	}
	if other.trySend([]byte("late")) {
		t.Fatal("trySend must fail after closeSend")
	}
}

// Regression: after the hub stopped, ReadPump's unregister and the handler's
// register blocked forever; a full Send buffer blocked sync replies.
func TestWatchPartyHub_StoppedHubDoesNotBlockClients(t *testing.T) {
	hub := newTestWatchPartyHub()
	close(hub.stopChan)
	client := newTestWatchPartyClient(hub, uuid.New())
	client.Send = make(chan []byte) // unbuffered: a blocking send would hang

	done := make(chan struct{})
	go func() {
		if hub.RegisterClient(client) {
			t.Error("RegisterClient should fail on a stopped hub")
		}
		client.leaveHub()
		client.trySend([]byte("sync"))
		client.handleCommand(context.Background(), &models.WatchPartyCommand{Type: "typing"})
		close(done)
	}()
	select {
	case <-done:
	case <-time.After(2 * time.Second):
		t.Fatal("client operations blocked on a stopped hub")
	}
}
