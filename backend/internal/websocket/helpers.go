package websocket

import "time"

// Pointer helpers for websocket messages.

func strPtr(s string) *string {
	return &s
}

func timePtr(t time.Time) *time.Time {
	return &t
}
