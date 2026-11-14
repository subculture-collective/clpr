package services

import (
	"bufio"
	"context"
	"errors"
	"net"
	"strings"
	"sync/atomic"
	"testing"
	"time"

	"git.subcult.tv/subculture-collective/clpr/internal/models"
	"github.com/google/uuid"
)

func TestParseTwitchIRCPrivmsg(t *testing.T) {
	raw := "@id=msg-1;user-id=42;display-name=Viewer :viewer!viewer@viewer.tmi.twitch.tv PRIVMSG #moonmoon :watch https://clpr.tv/clip/123e4567-e89b-12d3-a456-426614174000"
	msg, ok := ParseTwitchIRCPrivmsg(raw)
	if !ok {
		t.Fatal("expected PRIVMSG")
	}
	if msg.MessageID != "msg-1" || msg.UserID != "42" || msg.Username != "Viewer" || msg.Channel != "moonmoon" {
		t.Fatalf("unexpected parsed message: %#v", msg)
	}
	if msg.Text != "watch https://clpr.tv/clip/123e4567-e89b-12d3-a456-426614174000" {
		t.Fatalf("unexpected parsed text: %#v", msg.Text)
	}
}

func TestParseTwitchIRCPrivmsgRejectsNonPrivmsg(t *testing.T) {
	if _, ok := ParseTwitchIRCPrivmsg("PING :tmi.twitch.tv"); ok {
		t.Fatal("expected non-PRIVMSG to be rejected")
	}
}

func TestParseTwitchIRCPrivmsgStripsActionWrapper(t *testing.T) {
	raw := "@id=msg-2;user-id=42 :viewer!viewer@viewer.tmi.twitch.tv PRIVMSG #moonmoon :\x01ACTION https://clips.twitch.tv/SomeSlug\x01"
	msg, ok := ParseTwitchIRCPrivmsg(raw)
	if !ok {
		t.Fatal("expected PRIVMSG")
	}
	if msg.Text != "https://clips.twitch.tv/SomeSlug" {
		t.Fatalf("unexpected parsed text: %q", msg.Text)
	}
	links := ExtractClipLinks(msg.Text)
	if len(links) != 1 || links[0].TwitchClipID != "SomeSlug" {
		t.Fatalf("expected the /me clip link to be detected, got %#v", links)
	}
}

// fakeIRCServer is a local TCP server standing in for Twitch chat.
type fakeIRCServer struct {
	t        *testing.T
	listener net.Listener
}

func newFakeIRCServer(t *testing.T) *fakeIRCServer {
	t.Helper()
	listener, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatalf("failed to listen: %v", err)
	}
	t.Cleanup(func() { _ = listener.Close() })
	return &fakeIRCServer{t: t, listener: listener}
}

func (s *fakeIRCServer) dial(ctx context.Context) (net.Conn, error) {
	var dialer net.Dialer
	return dialer.DialContext(ctx, "tcp", s.listener.Addr().String())
}

type fakeIRCConn struct {
	t      *testing.T
	conn   net.Conn
	reader *bufio.Reader
}

func (s *fakeIRCServer) accept() *fakeIRCConn {
	s.t.Helper()
	_ = s.listener.(*net.TCPListener).SetDeadline(time.Now().Add(5 * time.Second))
	conn, err := s.listener.Accept()
	if err != nil {
		s.t.Fatalf("expected the listener to connect: %v", err)
	}
	s.t.Cleanup(func() { _ = conn.Close() })
	return &fakeIRCConn{t: s.t, conn: conn, reader: bufio.NewReader(conn)}
}

func (c *fakeIRCConn) expect(prefix string) string {
	c.t.Helper()
	_ = c.conn.SetReadDeadline(time.Now().Add(5 * time.Second))
	line, err := c.reader.ReadString('\n')
	if err != nil {
		c.t.Fatalf("expected a line starting with %q: %v", prefix, err)
	}
	line = strings.TrimRight(line, "\r\n")
	if !strings.HasPrefix(line, prefix) {
		c.t.Fatalf("expected a line starting with %q, got %q", prefix, line)
	}
	return line
}

func (c *fakeIRCConn) send(line string) {
	c.t.Helper()
	if _, err := c.conn.Write([]byte(line + "\r\n")); err != nil {
		c.t.Fatalf("failed to send %q: %v", line, err)
	}
}

// login plays the server side of a successful sign-in and join.
func (c *fakeIRCConn) login() {
	c.t.Helper()
	c.expect("PASS oauth:bot-token")
	c.expect("NICK clprbot")
	c.expect("CAP REQ")
	c.send(":tmi.twitch.tv 001 clprbot :Welcome, GLHF!")
	c.expect("JOIN #moonmoon")
}

func (c *fakeIRCConn) expectClosed() {
	c.t.Helper()
	_ = c.conn.SetReadDeadline(time.Now().Add(5 * time.Second))
	for {
		if _, err := c.reader.ReadString('\n'); err != nil {
			if netErr, ok := err.(net.Error); ok && netErr.Timeout() {
				c.t.Fatal("expected the listener to close the connection")
			}
			return
		}
	}
}

type listenerFixture struct {
	manager *TwitchChatListenerManager
	rooms   *fakeStreamerClipRoomRepository
	clips   *fakeClipLookupRepository
	server  *fakeIRCServer
	roomID  uuid.UUID
}

func newListenerFixture(t *testing.T) *listenerFixture {
	t.Helper()
	rooms := newFakeStreamerClipRoomRepository()
	room := &models.StreamerClipRoom{
		ID:              uuid.New(),
		OwnerUserID:     uuid.New(),
		TwitchChannel:   "moonmoon",
		ApprovalMode:    models.StreamerClipRoomApprovalManual,
		IsActive:        true,
		SubmissionsOpen: true,
	}
	rooms.seedRoom(room)
	clips := newFakeClipLookupRepository()
	server := newFakeIRCServer(t)

	manager := NewTwitchChatListenerManager(newStreamerClipRoomService(rooms, clips))
	manager.dial = server.dial
	manager.backoff = func(int) time.Duration { return time.Millisecond }
	manager.loginTimeout = 5 * time.Second
	t.Cleanup(manager.Shutdown)

	return &listenerFixture{manager: manager, rooms: rooms, clips: clips, server: server, roomID: room.ID}
}

func (f *listenerFixture) start(t *testing.T) {
	t.Helper()
	if err := f.manager.StartWithUsername(f.roomID, "moonmoon", "clprbot", "bot-token"); err != nil {
		t.Fatalf("failed to start listener: %v", err)
	}
}

func waitFor(t *testing.T, what string, condition func() bool) {
	t.Helper()
	deadline := time.Now().Add(5 * time.Second)
	for !condition() {
		if time.Now().After(deadline) {
			t.Fatalf("timed out waiting for %s", what)
		}
		time.Sleep(5 * time.Millisecond)
	}
}

func TestTwitchChatListenerRunsUntilShutdownAndLeavesRoomActive(t *testing.T) {
	f := newListenerFixture(t)
	f.start(t)

	conn := f.server.accept()
	conn.login()
	conn.send("PING :tmi.twitch.tv")
	conn.expect("PONG :tmi.twitch.tv")

	if !f.manager.IsRunning(f.roomID) {
		t.Fatal("expected the listener to keep running after Start returned")
	}

	f.manager.Shutdown()
	conn.expectClosed()

	if f.manager.IsRunning(f.roomID) {
		t.Fatal("expected no listener after shutdown")
	}
	room := f.rooms.room(f.roomID)
	if !room.IsActive || room.LastListenerError != nil {
		t.Fatalf("shutdown must leave the room active for the next process, got active=%v error=%v", room.IsActive, room.LastListenerError)
	}
	if err := f.manager.StartWithUsername(f.roomID, "moonmoon", "clprbot", "bot-token"); err == nil {
		t.Fatal("expected Start to fail after shutdown")
	}
}

func TestTwitchChatListenerReconnectsAfterDropAndReconnectCommand(t *testing.T) {
	f := newListenerFixture(t)
	clipID := uuid.New()
	f.clips.clipsByID[clipID] = &models.Clip{ID: clipID, Title: "Clip"}
	f.start(t)

	first := f.server.accept()
	first.login()
	_ = first.conn.Close() // connection drops: the listener sees EOF

	second := f.server.accept()
	second.login()
	second.send(":tmi.twitch.tv RECONNECT")

	third := f.server.accept()
	third.login()
	third.send("@id=msg-1;user-id=42;display-name=Viewer :viewer!viewer@viewer.tmi.twitch.tv PRIVMSG #moonmoon :\x01ACTION https://clpr.tv/clip/" + clipID.String() + "\x01")

	waitFor(t, "the chat message to be queued", func() bool {
		items, _ := f.rooms.ListItems(context.Background(), f.roomID, models.StreamerClipRoomItemStatusPending, 10)
		return len(items) == 1
	})

	room := f.rooms.room(f.roomID)
	if !room.IsActive || room.LastListenerError != nil {
		t.Fatalf("reconnects must keep the room active, got active=%v error=%v", room.IsActive, room.LastListenerError)
	}
	if !f.manager.IsRunning(f.roomID) {
		t.Fatal("expected the listener to still be running")
	}
}

func TestTwitchChatListenerStopsOnFatalNotice(t *testing.T) {
	tests := []struct {
		name   string
		script func(conn *fakeIRCConn)
		want   string
	}{
		{
			name: "rejected login",
			script: func(conn *fakeIRCConn) {
				conn.expect("PASS")
				conn.expect("NICK")
				conn.expect("CAP REQ")
				conn.send(":tmi.twitch.tv NOTICE * :Login authentication failed")
				_ = conn.conn.Close()
			},
			want: "The Clpr bot could not sign in to Twitch. This is a problem on our side.",
		},
		{
			name: "banned in channel",
			script: func(conn *fakeIRCConn) {
				conn.login()
				conn.send("@msg-id=msg_banned :tmi.twitch.tv NOTICE #moonmoon :You are permanently banned from talking in moonmoon.")
			},
			want: "The Clpr bot is banned in this channel.",
		},
		{
			name: "suspended channel",
			script: func(conn *fakeIRCConn) {
				conn.login()
				conn.send("@msg-id=msg_channel_suspended :tmi.twitch.tv NOTICE #moonmoon :This channel does not exist or has been suspended.")
			},
			want: "This Twitch channel is suspended, so the Clpr bot cannot join it.",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			f := newListenerFixture(t)
			f.start(t)
			tt.script(f.server.accept())

			waitFor(t, "the listener to stop", func() bool { return !f.manager.IsRunning(f.roomID) })
			room := f.rooms.room(f.roomID)
			if room.IsActive {
				t.Fatal("expected the room to be marked stopped")
			}
			if room.LastListenerError == nil || *room.LastListenerError != tt.want {
				t.Fatalf("unexpected listener error: %v", room.LastListenerError)
			}
		})
	}
}

func TestTwitchChatListenerStopsAfterRepeatedFailedReconnects(t *testing.T) {
	f := newListenerFixture(t)
	var dials atomic.Int32
	f.manager.dial = func(context.Context) (net.Conn, error) {
		dials.Add(1)
		return nil, errors.New("connection refused")
	}
	f.manager.maxFailures = 3
	f.start(t)

	waitFor(t, "the listener to give up", func() bool { return !f.manager.IsRunning(f.roomID) })
	if got := dials.Load(); got != 3 {
		t.Fatalf("expected 3 connection attempts, got %d", got)
	}
	room := f.rooms.room(f.roomID)
	if room.IsActive || room.LastListenerError == nil || *room.LastListenerError != listenerConnectionLostMessage {
		t.Fatalf("expected a readable stop message, got active=%v error=%v", room.IsActive, room.LastListenerError)
	}
}

func TestTwitchChatListenerStopsWhenSignedInConnectionsKeepDropping(t *testing.T) {
	f := newListenerFixture(t)
	f.manager.maxFailures = 3
	f.start(t)

	// Twitch accepts the login and drops the connection straight away, every time.
	for i := 0; i < 3; i++ {
		conn := f.server.accept()
		conn.login()
		_ = conn.conn.Close()
	}

	waitFor(t, "the listener to give up", func() bool { return !f.manager.IsRunning(f.roomID) })
	room := f.rooms.room(f.roomID)
	if room.IsActive || room.LastListenerError == nil || *room.LastListenerError != listenerConnectionLostMessage {
		t.Fatalf("expected a readable stop message, got active=%v error=%v", room.IsActive, room.LastListenerError)
	}
}

func TestTwitchChatListenerHealthyConnectionResetsFailureCount(t *testing.T) {
	f := newListenerFixture(t)
	f.manager.maxFailures = 2
	f.manager.stableAfter = 0
	f.start(t)

	for i := 0; i < 4; i++ {
		conn := f.server.accept()
		conn.login()
		_ = conn.conn.Close()
	}
	f.server.accept().login()

	room := f.rooms.room(f.roomID)
	if !room.IsActive || !f.manager.IsRunning(f.roomID) {
		t.Fatalf("drops after healthy connections must not stop the room, got active=%v", room.IsActive)
	}
}

func TestTwitchChatListenerIgnoresLateFailureFromReplacedListener(t *testing.T) {
	f := newListenerFixture(t)
	f.start(t)
	first := f.server.accept()
	first.login()

	f.manager.mu.Lock()
	staleToken := f.manager.listenerTokens[f.roomID]
	f.manager.mu.Unlock()

	f.start(t) // the creator presses Start again
	first.expectClosed()
	f.server.accept().login()

	// The replaced listener reports its failure only now.
	f.manager.finishListener(f.roomID, staleToken, errors.New("EOF"))

	room := f.rooms.room(f.roomID)
	if !room.IsActive || room.LastListenerError != nil {
		t.Fatalf("a replaced listener must not stop the restarted room, got active=%v error=%v", room.IsActive, room.LastListenerError)
	}
	if !f.manager.IsRunning(f.roomID) {
		t.Fatal("expected the new listener to keep running")
	}
}

func TestTwitchChatListenerResumesActiveRooms(t *testing.T) {
	f := newListenerFixture(t)
	inactive := &models.StreamerClipRoom{ID: uuid.New(), OwnerUserID: uuid.New(), TwitchChannel: "idle"}
	f.rooms.seedRoom(inactive)

	resumed, err := f.manager.ResumeActiveRooms(context.Background(), "clprbot", "bot-token")
	if err != nil || resumed != 1 {
		t.Fatalf("expected one resumed room, got %d (%v)", resumed, err)
	}
	f.server.accept().login()

	if !f.manager.IsRunning(f.roomID) || f.manager.IsRunning(inactive.ID) {
		t.Fatal("expected only the active room to have a listener")
	}
	room := f.rooms.room(f.roomID)
	if !room.IsActive || !room.SubmissionsOpen {
		t.Fatalf("resume must keep the room live with submissions open, got %#v", room)
	}
}

func TestTwitchChatListenerResumeWithoutBotMarksRoomsStopped(t *testing.T) {
	f := newListenerFixture(t)

	resumed, err := f.manager.ResumeActiveRooms(context.Background(), "", "")
	if err != nil || resumed != 0 {
		t.Fatalf("expected no resumed rooms, got %d (%v)", resumed, err)
	}
	room := f.rooms.room(f.roomID)
	if room.IsActive || room.LastListenerError == nil || *room.LastListenerError != listenerBotNotConfiguredMessage {
		t.Fatalf("expected the room to be stopped with a clear message, got active=%v error=%v", room.IsActive, room.LastListenerError)
	}
	if f.manager.IsRunning(f.roomID) {
		t.Fatal("expected no listener without bot credentials")
	}
}
