package services

import (
	"bufio"
	"context"
	"crypto/tls"
	"errors"
	"fmt"
	"net"
	"os"
	"strings"
	"sync"
	"time"

	"github.com/google/uuid"
)

type TwitchIRCMessage struct {
	MessageID string
	UserID    string
	Username  string
	Channel   string
	Text      string
}

const (
	twitchIRCAddress = "irc.chat.twitch.tv:6697"
	// Twitch sends a PING about every five minutes, so a quiet socket past
	// this deadline is treated as dead and reconnected.
	twitchIRCReadTimeout  = 6 * time.Minute
	twitchIRCLoginTimeout = 30 * time.Second
	// Consecutive failed connection attempts before the room is marked stopped
	// (about two minutes with the default backoff).
	twitchIRCMaxFailures = 8
	// A connection counts as healthy only after it has stayed signed in this
	// long. Shorter ones count as failures, so a connection that Twitch
	// accepts and drops straight away cannot loop forever.
	twitchIRCStableAfter = time.Minute
	// Room status writes happen under the manager lock, so they are bounded.
	listenerStatusWriteTimeout = 10 * time.Second
)

const (
	listenerSignInFailedMessage     = "The Clpr bot could not sign in to Twitch. This is a problem on our side."
	listenerBannedMessage           = "The Clpr bot is banned in this channel."
	listenerSuspendedMessage        = "This Twitch channel is suspended, so the Clpr bot cannot join it."
	listenerConnectionLostMessage   = "The Clpr bot lost its connection to Twitch chat and could not reconnect. Start the listener again."
	listenerBotNotConfiguredMessage = "The Clpr bot is not set up, so the listener stopped. This is a problem on our side."
)

// listenerFatalError stops a listener for good; retrying would not help.
type listenerFatalError struct{ message string }

func (e *listenerFatalError) Error() string { return e.message }

type TwitchChatListenerManager struct {
	mu             sync.Mutex
	listeners      map[uuid.UUID]context.CancelFunc
	listenerTokens map[uuid.UUID]uint64
	nextToken      uint64
	service        *StreamerClipRoomService

	// Listeners run on baseCtx, never on a request context, so they outlive
	// the HTTP request that started them. Shutdown cancels it.
	baseCtx    context.Context
	baseCancel context.CancelFunc

	dial         func(ctx context.Context) (net.Conn, error)
	backoff      func(failures int) time.Duration
	readTimeout  time.Duration
	loginTimeout time.Duration
	maxFailures  int
	stableAfter  time.Duration
}

func NewTwitchChatListenerManager(service *StreamerClipRoomService) *TwitchChatListenerManager {
	baseCtx, baseCancel := context.WithCancel(context.Background())
	return &TwitchChatListenerManager{
		listeners:      make(map[uuid.UUID]context.CancelFunc),
		listenerTokens: make(map[uuid.UUID]uint64),
		service:        service,
		baseCtx:        baseCtx,
		baseCancel:     baseCancel,
		dial:           dialTwitchIRC,
		backoff:        twitchIRCBackoff,
		readTimeout:    twitchIRCReadTimeout,
		loginTimeout:   twitchIRCLoginTimeout,
		maxFailures:    twitchIRCMaxFailures,
		stableAfter:    twitchIRCStableAfter,
	}
}

// TwitchBotCredentialsFromEnv returns the bot login and chat token. Either is
// empty when the bot is not configured.
func TwitchBotCredentialsFromEnv() (username string, oauthToken string) {
	username = strings.TrimSpace(os.Getenv("TWITCH_BOT_USERNAME"))
	oauthToken = strings.TrimSpace(strings.TrimPrefix(strings.TrimSpace(os.Getenv("TWITCH_BOT_OAUTH_TOKEN")), "oauth:"))
	return username, oauthToken
}

func dialTwitchIRC(ctx context.Context) (net.Conn, error) {
	dialer := &tls.Dialer{
		NetDialer: &net.Dialer{Timeout: 10 * time.Second},
		Config: &tls.Config{
			ServerName: "irc.chat.twitch.tv",
			MinVersion: tls.VersionTLS12,
		},
	}
	return dialer.DialContext(ctx, "tcp", twitchIRCAddress)
}

// twitchIRCBackoff waits 1s, 2s, 4s ... capped at 30s.
func twitchIRCBackoff(failures int) time.Duration {
	if failures < 0 {
		failures = 0
	}
	if failures > 5 {
		return 30 * time.Second
	}
	return time.Second << failures
}

func ParseTwitchIRCPrivmsg(raw string) (TwitchIRCMessage, bool) {
	if !strings.Contains(raw, " PRIVMSG ") {
		return TwitchIRCMessage{}, false
	}

	msg := TwitchIRCMessage{}
	line := strings.TrimRight(raw, "\r\n")

	if strings.HasPrefix(line, "@") {
		tagSection, remainder, ok := strings.Cut(line, " ")
		if !ok {
			return TwitchIRCMessage{}, false
		}
		for _, entry := range strings.Split(strings.TrimPrefix(tagSection, "@"), ";") {
			if entry == "" {
				continue
			}
			key, value, found := strings.Cut(entry, "=")
			if !found {
				continue
			}
			switch key {
			case "id":
				msg.MessageID = unescapeTwitchIRCTag(value)
			case "user-id":
				msg.UserID = unescapeTwitchIRCTag(value)
			case "display-name":
				msg.Username = unescapeTwitchIRCTag(value)
			}
		}
		line = remainder
	}

	prefixAndChannel, text, ok := strings.Cut(line, " :")
	if !ok {
		return TwitchIRCMessage{}, false
	}

	_, channel, ok := strings.Cut(prefixAndChannel, " PRIVMSG #")
	if !ok {
		return TwitchIRCMessage{}, false
	}

	channel = strings.TrimSpace(channel)
	if channel == "" || text == "" {
		return TwitchIRCMessage{}, false
	}

	// A /me message arrives wrapped as "\x01ACTION text\x01".
	if rest, isAction := strings.CutPrefix(text, "\x01ACTION "); isAction {
		text = strings.TrimSuffix(rest, "\x01")
		if text == "" {
			return TwitchIRCMessage{}, false
		}
	}

	msg.Channel = channel
	msg.Text = text
	return msg, true
}

func (m *TwitchChatListenerManager) Start(roomID uuid.UUID, channel string, oauthToken string) error {
	return m.StartWithUsername(roomID, channel, deriveFallbackTwitchNickname(channel), oauthToken)
}

// StartWithUsername starts (or replaces) the chat listener for a room. The
// listener runs until Stop, Shutdown or a fatal Twitch error.
func (m *TwitchChatListenerManager) StartWithUsername(roomID uuid.UUID, channel string, username string, oauthToken string) error {
	if m == nil {
		return fmt.Errorf("twitch chat listener manager is nil")
	}
	channel = normalizeTwitchChannel(channel)
	if channel == "" {
		return fmt.Errorf("channel is required")
	}
	if strings.TrimSpace(oauthToken) == "" {
		return fmt.Errorf("oauth token is required")
	}
	username = normalizeTwitchNickname(username)
	if username == "" {
		username = deriveFallbackTwitchNickname(channel)
	}

	m.mu.Lock()
	defer m.mu.Unlock()
	if m.baseCtx == nil || m.baseCtx.Err() != nil {
		return fmt.Errorf("twitch chat listener manager is shut down")
	}
	// Mark the room active under the same lock that guards the listener
	// tokens. A listener being replaced may be writing "stopped" right now;
	// it holds this lock while it does, so this write always lands last.
	if m.service != nil {
		startCtx, cancelStart := context.WithTimeout(m.baseCtx, listenerStatusWriteTimeout)
		err := m.service.markListenerStarted(startCtx, roomID)
		cancelStart()
		if err != nil {
			return err
		}
	}

	listenerCtx, cancel := context.WithCancel(m.baseCtx)
	m.nextToken++
	token := m.nextToken
	if existingCancel, ok := m.listeners[roomID]; ok {
		existingCancel()
	}
	m.listeners[roomID] = cancel
	m.listenerTokens[roomID] = token

	go m.runListener(listenerCtx, roomID, channel, username, oauthToken, token)
	return nil
}

// ResumeActiveRooms restarts listeners for rooms that were active when the
// previous process exited. Without bot credentials those rooms are marked
// stopped so the page does not show a listener that is not there.
//
// During a blue/green overlap two processes listen to the same chat. That is
// safe: the unique index on (room_id, twitch_message_id, source_url) rejects
// the second copy of each submission.
func (m *TwitchChatListenerManager) ResumeActiveRooms(ctx context.Context, botUsername string, botOAuthToken string) (int, error) {
	if m == nil || m.service == nil {
		return 0, nil
	}
	rooms, err := m.service.activeRooms(ctx)
	if err != nil {
		return 0, err
	}

	resumed := 0
	for _, room := range rooms {
		if botUsername == "" || botOAuthToken == "" {
			m.service.MarkListenerStopped(room.ID, &listenerFatalError{message: listenerBotNotConfiguredMessage})
			continue
		}
		if err := m.StartWithUsername(room.ID, room.TwitchChannel, botUsername, botOAuthToken); err != nil {
			m.service.MarkListenerStopped(room.ID, &listenerFatalError{message: listenerConnectionLostMessage})
			continue
		}
		resumed++
	}
	return resumed, nil
}

// Shutdown stops every listener without marking rooms inactive, so the next
// process resumes them.
func (m *TwitchChatListenerManager) Shutdown() {
	if m == nil {
		return
	}

	m.mu.Lock()
	defer m.mu.Unlock()
	if m.baseCancel != nil {
		m.baseCancel()
	}
	m.listeners = make(map[uuid.UUID]context.CancelFunc)
	m.listenerTokens = make(map[uuid.UUID]uint64)
}

func (m *TwitchChatListenerManager) Stop(roomID uuid.UUID) {
	if m == nil {
		return
	}

	m.mu.Lock()
	cancel, ok := m.listeners[roomID]
	if ok {
		delete(m.listeners, roomID)
		delete(m.listenerTokens, roomID)
	}
	m.mu.Unlock()

	if ok {
		cancel()
	}
}

func (m *TwitchChatListenerManager) IsRunning(roomID uuid.UUID) bool {
	if m == nil {
		return false
	}

	m.mu.Lock()
	defer m.mu.Unlock()
	_, ok := m.listeners[roomID]
	return ok
}

func (m *TwitchChatListenerManager) runListener(ctx context.Context, roomID uuid.UUID, channel string, username string, oauthToken string, token uint64) {
	failures := 0
	for {
		connectedAt := time.Now()
		welcomed, err := m.runConnection(ctx, roomID, channel, username, oauthToken)
		if ctx.Err() != nil {
			m.finishListener(roomID, token, nil)
			return
		}

		var fatal *listenerFatalError
		if errors.As(err, &fatal) {
			m.finishListener(roomID, token, fatal)
			return
		}

		// A drop after a healthy connection (including a Twitch RECONNECT)
		// starts the count again.
		if welcomed && time.Since(connectedAt) >= m.stableAfter {
			failures = 0
		} else {
			failures++
			if failures >= m.maxFailures {
				m.finishListener(roomID, token, &listenerFatalError{message: listenerConnectionLostMessage})
				return
			}
		}

		timer := time.NewTimer(m.backoff(failures))
		select {
		case <-ctx.Done():
			timer.Stop()
			m.finishListener(roomID, token, nil)
			return
		case <-timer.C:
		}
	}
}

// runConnection handles one IRC connection. It reports whether Twitch accepted
// the login, and returns a *listenerFatalError when reconnecting cannot help.
func (m *TwitchChatListenerManager) runConnection(ctx context.Context, roomID uuid.UUID, channel string, username string, oauthToken string) (bool, error) {
	conn, err := m.dial(ctx)
	if err != nil {
		return false, err
	}
	defer conn.Close()

	closed := make(chan struct{})
	defer close(closed)
	go func() {
		select {
		case <-ctx.Done():
			_ = conn.Close()
		case <-closed:
		}
	}()

	writeLine := func(format string, args ...any) error {
		_ = conn.SetWriteDeadline(time.Now().Add(m.loginTimeout))
		_, err := fmt.Fprintf(conn, format+"\r\n", args...)
		return err
	}

	if err := writeLine("PASS oauth:%s", strings.TrimSpace(oauthToken)); err != nil {
		return false, err
	}
	if err := writeLine("NICK %s", username); err != nil {
		return false, err
	}
	if err := writeLine("CAP REQ :twitch.tv/tags twitch.tv/commands twitch.tv/membership"); err != nil {
		return false, err
	}

	welcomed := false
	reader := bufio.NewReader(conn)
	for {
		timeout := m.readTimeout
		if !welcomed {
			timeout = m.loginTimeout
		}
		_ = conn.SetReadDeadline(time.Now().Add(timeout))

		line, err := reader.ReadString('\n')
		if err != nil {
			return welcomed, err
		}

		line = strings.TrimRight(line, "\r\n")
		if line == "" {
			continue
		}

		if strings.HasPrefix(line, "PING ") {
			if err := writeLine("PONG %s", strings.TrimPrefix(line, "PING ")); err != nil {
				return welcomed, err
			}
			continue
		}

		tags, command, text := splitTwitchIRCLine(line)
		switch command {
		case "001":
			// Join only once Twitch has accepted the login.
			if !welcomed {
				welcomed = true
				if err := writeLine("JOIN #%s", channel); err != nil {
					return welcomed, err
				}
			}
		case "RECONNECT":
			return welcomed, fmt.Errorf("twitch asked the listener to reconnect")
		case "NOTICE":
			if message := fatalTwitchNotice(tags, text); message != "" {
				return welcomed, &listenerFatalError{message: message}
			}
		case "PRIVMSG":
			ircMsg, ok := ParseTwitchIRCPrivmsg(line)
			if !ok || m.service == nil {
				continue
			}
			_, _ = m.service.IngestChatMessage(ctx, TwitchChatClipMessage{
				RoomID:          roomID,
				TwitchMessageID: ircMsg.MessageID,
				TwitchUserID:    ircMsg.UserID,
				TwitchUsername:  ircMsg.Username,
				MessageText:     ircMsg.Text,
			})
		}
	}
}

// splitTwitchIRCLine returns the raw tag section, the command and the
// trailing text of an IRC line.
func splitTwitchIRCLine(line string) (tags string, command string, text string) {
	if strings.HasPrefix(line, "@") {
		tags, line, _ = strings.Cut(line, " ")
	}
	if strings.HasPrefix(line, ":") {
		_, line, _ = strings.Cut(line, " ")
	}
	head, text, _ := strings.Cut(line, " :")
	command, _, _ = strings.Cut(strings.TrimSpace(head), " ")
	return tags, command, text
}

// fatalTwitchNotice maps a Twitch NOTICE to a message for the creator, or ""
// when the notice does not stop the listener.
func fatalTwitchNotice(tags string, text string) string {
	for _, entry := range strings.Split(strings.TrimPrefix(tags, "@"), ";") {
		key, value, _ := strings.Cut(entry, "=")
		if key != "msg-id" {
			continue
		}
		switch value {
		case "msg_banned":
			return listenerBannedMessage
		case "msg_channel_suspended":
			return listenerSuspendedMessage
		}
	}

	lower := strings.ToLower(text)
	if strings.Contains(lower, "login authentication failed") ||
		strings.Contains(lower, "improperly formatted auth") ||
		strings.Contains(lower, "login unsuccessful") {
		return listenerSignInFailedMessage
	}
	return ""
}

// finishListener removes the listener and, when stopErr is set, marks the room
// stopped. Both happen under the lock and only while this listener is still
// the current one for the room, so a late failure from a replaced listener
// cannot mark a freshly restarted room inactive.
func (m *TwitchChatListenerManager) finishListener(roomID uuid.UUID, token uint64, stopErr error) {
	if m == nil {
		return
	}

	m.mu.Lock()
	defer m.mu.Unlock()
	currentToken, ok := m.listenerTokens[roomID]
	if !ok || currentToken != token {
		return
	}
	delete(m.listeners, roomID)
	delete(m.listenerTokens, roomID)
	if stopErr != nil && m.service != nil {
		m.service.MarkListenerStopped(roomID, stopErr)
	}
}

func deriveFallbackTwitchNickname(channel string) string {
	channel = normalizeTwitchChannel(channel)
	if channel == "" {
		return "clpr_bot"
	}
	return normalizeTwitchNickname(channel)
}

func normalizeTwitchChannel(channel string) string {
	channel = strings.TrimSpace(strings.TrimPrefix(channel, "#"))
	channel = strings.ToLower(channel)
	return normalizeTwitchNickname(channel)
}

func normalizeTwitchNickname(username string) string {
	username = strings.TrimSpace(strings.ToLower(strings.TrimPrefix(username, "#")))
	if username == "" {
		return ""
	}

	var b strings.Builder
	b.Grow(len(username))
	for _, r := range username {
		switch {
		case r >= 'a' && r <= 'z':
			b.WriteRune(r)
		case r >= '0' && r <= '9':
			b.WriteRune(r)
		case r == '_':
			b.WriteRune(r)
		}
	}

	nick := b.String()
	if nick == "" {
		return ""
	}
	if len(nick) > 25 {
		nick = nick[:25]
	}
	return nick
}

var twitchIRCTagReplacer = strings.NewReplacer(
	`\\`, `\`,
	`\s`, " ",
	`\r`, "\r",
	`\n`, "\n",
	`\:`, ";",
)

func unescapeTwitchIRCTag(value string) string {
	return twitchIRCTagReplacer.Replace(value)
}
