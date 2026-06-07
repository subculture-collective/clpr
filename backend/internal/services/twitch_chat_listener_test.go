package services

import "testing"

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
