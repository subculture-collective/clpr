package services

import "testing"

func TestStreamerTagIdentity(t *testing.T) {
	cases := []struct {
		label, slug, name string
		ok                bool
	}{
		{"CommunityOriented", "streamer/community-oriented", "Streamer: Community Oriented", true},
		{"English", "streamer/english", "Streamer: English", true},
		{"ADHD", "streamer/adhd", "Streamer: ADHD", true},
		{"  ", "", "", false},
	}
	for _, tc := range cases {
		slug, name, ok := streamerTagIdentity(tc.label)
		if slug != tc.slug || name != tc.name || ok != tc.ok {
			t.Errorf("streamerTagIdentity(%q) = %q, %q, %v; want %q, %q, %v", tc.label, slug, name, ok, tc.slug, tc.name, tc.ok)
		}
	}
}
