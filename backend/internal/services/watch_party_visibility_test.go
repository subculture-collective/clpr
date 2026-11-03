package services

import (
	"testing"
	"time"

	"git.subcult.tv/subculture-collective/clpr/internal/models"
	"github.com/google/uuid"
)

func TestCanViewWatchParty(t *testing.T) {
	host := uuid.New()
	member := uuid.New()
	stranger := uuid.New()
	left := time.Now()

	active := &models.WatchPartyParticipant{UserID: member}
	departed := &models.WatchPartyParticipant{UserID: member, LeftAt: &left}

	for _, visibility := range []string{"private", "friends", "invite"} {
		party := &models.WatchParty{HostUserID: host, Visibility: visibility}
		if CanViewWatchParty(party, nil, nil) {
			t.Errorf("%s: anonymous viewer allowed", visibility)
		}
		if CanViewWatchParty(party, &stranger, nil) {
			t.Errorf("%s: non-participant allowed", visibility)
		}
		if CanViewWatchParty(party, &member, departed) {
			t.Errorf("%s: departed participant allowed", visibility)
		}
		if CanViewWatchParty(party, &stranger, active) {
			t.Errorf("%s: participant record for another user accepted", visibility)
		}
		if !CanViewWatchParty(party, &member, active) {
			t.Errorf("%s: active participant denied", visibility)
		}
		if !CanViewWatchParty(party, &host, nil) {
			t.Errorf("%s: host denied", visibility)
		}
	}

	public := &models.WatchParty{HostUserID: host, Visibility: "public"}
	if !CanViewWatchParty(public, nil, nil) {
		t.Error("public party should be readable anonymously")
	}
}

func TestRedactWatchPartyForViewer(t *testing.T) {
	host := uuid.New()
	member := uuid.New()

	for _, visibility := range []string{"private", "friends", "invite"} {
		party := &models.WatchParty{HostUserID: host, Visibility: visibility, InviteCode: "ABC123"}
		RedactWatchPartyForViewer(party, &member)
		if party.InviteCode != "" {
			t.Errorf("%s: invite code exposed to non-host", visibility)
		}

		party.InviteCode = "ABC123"
		RedactWatchPartyForViewer(party, nil)
		if party.InviteCode != "" {
			t.Errorf("%s: invite code exposed anonymously", visibility)
		}

		party.InviteCode = "ABC123"
		RedactWatchPartyForViewer(party, &host)
		if party.InviteCode != "ABC123" {
			t.Errorf("%s: invite code hidden from host", visibility)
		}
	}

	public := &models.WatchParty{HostUserID: host, Visibility: "public", InviteCode: "ABC123"}
	RedactWatchPartyForViewer(public, nil)
	if public.InviteCode != "ABC123" {
		t.Error("public parties need the invite code to be joinable")
	}
}
