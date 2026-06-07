package services

import (
	"bufio"
	"context"
	"crypto/tls"
	"fmt"
	"net"
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

type TwitchChatListenerManager struct {
	mu             sync.Mutex
	listeners      map[uuid.UUID]context.CancelFunc
	listenerTokens map[uuid.UUID]uint64
	nextToken      uint64
	service        *StreamerClipRoomService
}

func NewTwitchChatListenerManager(service *StreamerClipRoomService) *TwitchChatListenerManager {
	return &TwitchChatListenerManager{
		listeners:      make(map[uuid.UUID]context.CancelFunc),
		listenerTokens: make(map[uuid.UUID]uint64),
		service:        service,
	}
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

	msg.Channel = channel
	msg.Text = text
	return msg, true
}

func (m *TwitchChatListenerManager) Start(ctx context.Context, roomID uuid.UUID, channel string, oauthToken string) error {
	return m.StartWithUsername(ctx, roomID, channel, deriveFallbackTwitchNickname(channel), oauthToken)
}

func (m *TwitchChatListenerManager) StartWithUsername(ctx context.Context, roomID uuid.UUID, channel string, username string, oauthToken string) error {
	if m == nil {
		return fmt.Errorf("twitch chat listener manager is nil")
	}
	if ctx == nil {
		return fmt.Errorf("context is required")
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

	listenerCtx, cancel := context.WithCancel(ctx)

	m.mu.Lock()
	m.nextToken++
	token := m.nextToken
	if m.listeners == nil {
		m.listeners = make(map[uuid.UUID]context.CancelFunc)
	}
	if m.listenerTokens == nil {
		m.listenerTokens = make(map[uuid.UUID]uint64)
	}
	if existingCancel, ok := m.listeners[roomID]; ok {
		existingCancel()
	}
	m.listeners[roomID] = cancel
	m.listenerTokens[roomID] = token
	m.mu.Unlock()

	go m.runListener(listenerCtx, roomID, channel, username, oauthToken, token)
	return nil
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
	defer m.cleanupListener(roomID, token)
	var listenerErr error
	defer func() {
		if listenerErr != nil && ctx.Err() == nil && m.service != nil {
			m.service.MarkListenerStopped(roomID, listenerErr)
		}
	}()

	dialer := &net.Dialer{Timeout: 10 * time.Second}
	conn, err := tls.DialWithDialer(dialer, "tcp", "irc.chat.twitch.tv:6697", &tls.Config{
		ServerName:         "irc.chat.twitch.tv",
		MinVersion:         tls.VersionTLS12,
		InsecureSkipVerify: false,
	})
	if err != nil {
		listenerErr = err
		return
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
		_, err := fmt.Fprintf(conn, format+"\r\n", args...)
		return err
	}

	if err := writeLine("PASS oauth:%s", strings.TrimSpace(oauthToken)); err != nil {
		listenerErr = err
		return
	}
	if err := writeLine("NICK %s", username); err != nil {
		listenerErr = err
		return
	}
	_ = writeLine("CAP REQ :twitch.tv/tags twitch.tv/commands twitch.tv/membership")
	if err := writeLine("JOIN #%s", channel); err != nil {
		listenerErr = err
		return
	}

	reader := bufio.NewReader(conn)
	for {
		select {
		case <-ctx.Done():
			return
		default:
		}

		line, err := reader.ReadString('\n')
		if err != nil {
			listenerErr = err
			return
		}

		line = strings.TrimRight(line, "\r\n")
		if line == "" {
			continue
		}

		if strings.HasPrefix(line, "PING ") {
			_ = writeLine("PONG %s", strings.TrimPrefix(line, "PING "))
			continue
		}

		ircMsg, ok := ParseTwitchIRCPrivmsg(line)
		if !ok {
			continue
		}

		if m.service == nil {
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

func (m *TwitchChatListenerManager) cleanupListener(roomID uuid.UUID, token uint64) {
	if m == nil {
		return
	}

	m.mu.Lock()
	defer m.mu.Unlock()
	if currentToken, ok := m.listenerTokens[roomID]; ok && currentToken == token {
		delete(m.listeners, roomID)
		delete(m.listenerTokens, roomID)
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
