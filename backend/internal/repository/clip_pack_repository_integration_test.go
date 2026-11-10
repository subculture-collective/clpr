//go:build integration

package repository

import (
	"context"
	"errors"
	"strings"
	"testing"
	"time"

	"git.subcult.tv/subculture-collective/clpr/internal/models"
	"git.subcult.tv/subculture-collective/clpr/internal/testutil"
	"github.com/google/uuid"
)

func TestClipPackLifecycle(t *testing.T) {
	pool := testutil.SetupTestDB(t)
	defer testutil.CleanupTestDB(t, pool)
	testutil.TruncateTables(t, pool, "clip_packs", "countdown_exclusions", "broadcaster_feature_optins", "clips", "users")
	ctx := context.Background()
	repo := NewCountdownRepository(pool)

	weekStart := time.Date(2026, 11, 2, 0, 0, 0, 0, time.UTC)
	weekEnd := weekStart.AddDate(0, 0, 7)
	ids := insertCountdownClips(t, pool, weekEnd, []countdownClip{
		{name: "a-top", broadcaster: "alpha", velocity: 1, duration: 30, age: 24 * time.Hour},
		{name: "a-second", broadcaster: "alpha", velocity: 1, duration: 30, age: 48 * time.Hour},
		{name: "a-third", broadcaster: "alpha", velocity: 1, duration: 30, age: 72 * time.Hour},
		{name: "a-hidden", broadcaster: "alpha", velocity: 1, duration: 30, age: time.Hour, hidden: true},
		{name: "a-last-week", broadcaster: "alpha", velocity: 1, duration: 30, age: 8 * 24 * time.Hour},
		{name: "b-clip", broadcaster: "bravo", velocity: 1, duration: 30, age: time.Hour},
		{name: "c-clip", broadcaster: "charlie", velocity: 1, duration: 30, age: time.Hour},
	})
	// Views decide the order: a-second has the most.
	for name, views := range map[string]int{"a-top": 500, "a-second": 900, "a-third": 10} {
		if _, err := pool.Exec(ctx, `UPDATE clips SET view_count = $2 WHERE id = $1`, ids[name], views); err != nil {
			t.Fatal(err)
		}
	}
	for _, b := range []string{"alpha", "bravo"} { // charlie does not feature their clips
		userID := uuid.New()
		if _, err := pool.Exec(ctx, `INSERT INTO users (id, twitch_id, username, display_name, role, account_type)
			VALUES ($1, $2, $2, $2, 'user', 'member')`, userID, b); err != nil {
			t.Fatal(err)
		}
		if _, err := pool.Exec(ctx, `INSERT INTO broadcaster_feature_optins (broadcaster_id, user_id, scope, consent_text_version)
			VALUES ($1, $2, 'channel:manage:clips', '2026-11-07')`, b, userID); err != nil {
			t.Fatal(err)
		}
	}
	bravo := "bravo"
	if _, err := repo.CreateExclusion(ctx, &models.CountdownExclusionInput{BroadcasterID: &bravo, Reason: "request"}, "admin", nil); err != nil {
		t.Fatal(err)
	}

	if _, err := repo.EnsureWeeklyPacks(ctx, weekStart.AddDate(0, 0, 1), 2); err == nil {
		t.Fatal("a Tuesday week start was accepted")
	}
	created, err := repo.EnsureWeeklyPacks(ctx, weekStart, 2)
	if err != nil || created != 1 {
		t.Fatalf("created = %d, %v (only alpha is due)", created, err)
	}
	if created, err := repo.EnsureWeeklyPacks(ctx, weekStart, 2); err != nil || created != 0 {
		t.Fatalf("second ensure created %d, %v", created, err)
	}

	pack, err := repo.ClaimPackForBuild(ctx, time.Hour)
	if err != nil || pack == nil || pack.BroadcasterID != "alpha" || pack.WeekStart != "2026-11-02" || pack.Status != models.ClipPackBuilding {
		t.Fatalf("claimed = %+v, %v", pack, err)
	}
	if len(pack.Clips) != 2 || pack.Clips[0].ClipID != ids["a-second"] || pack.Clips[1].ClipID != ids["a-top"] || pack.Clips[0].ViewCount != 900 {
		t.Fatalf("pack clips = %+v", pack.Clips)
	}
	if again, err := repo.ClaimPackForBuild(ctx, time.Hour); err != nil || again != nil {
		t.Fatalf("a building pack was claimed again: %+v, %v", again, err)
	}
	if again, err := repo.ClaimPackForBuild(ctx, 0); err != nil || again == nil || again.ID != pack.ID {
		t.Fatalf("a stale claim was not taken again: %+v, %v", again, err)
	}

	expires := time.Now().Add(7 * 24 * time.Hour)
	name := pack.ID.String() + "/clpr-pack-alpha-2026-11-02.zip"
	if err := repo.CompletePack(ctx, pack.ID, name, strings.Repeat("c", 64), 1234, 2, expires); err != nil {
		t.Fatal(err)
	}
	if err := repo.FailPack(ctx, pack.ID, "late"); !errors.Is(err, ErrClipPackNotBuilding) {
		t.Fatalf("failing a ready pack = %v", err)
	}
	file, err := repo.PackFile(ctx, pack.ID)
	if err != nil || file == nil || file.FileName != name || file.Status != models.ClipPackReady || file.Deleted || file.ExpiresAt == nil {
		t.Fatalf("pack file = %+v, %v", file, err)
	}
	if missing, err := repo.PackFile(ctx, uuid.New()); err != nil || missing != nil {
		t.Fatalf("unknown pack file = %+v, %v", missing, err)
	}
	packs, err := repo.ListBroadcasterPacks(ctx, "alpha", 10)
	if err != nil || len(packs) != 1 || packs[0].Status != models.ClipPackReady || *packs[0].ClipCount != 2 || *packs[0].Bytes != 1234 {
		t.Fatalf("alpha packs = %+v, %v", packs, err)
	}

	// A clip taken down after the pack was made withdraws the pack.
	if _, err := pool.Exec(ctx, `UPDATE clips SET dmca_removed = true WHERE id = $1`, ids["a-top"]); err != nil {
		t.Fatal(err)
	}
	if file, _ := repo.PackFile(ctx, pack.ID); file == nil || !file.Withdrawn {
		t.Fatalf("pack with a removed clip = %+v", file)
	}
	if packs, _ := repo.ListBroadcasterPacks(ctx, "alpha", 10); len(packs) != 1 || !packs[0].Withdrawn {
		t.Fatalf("listed pack with a removed clip = %+v", packs)
	}
	if _, err := pool.Exec(ctx, `UPDATE clips SET dmca_removed = false WHERE id = $1`, ids["a-top"]); err != nil {
		t.Fatal(err)
	}

	if due, err := repo.ListPacksToDelete(ctx, time.Now()); err != nil || len(due) != 0 {
		t.Fatalf("due before expiry = %+v, %v", due, err)
	}
	due, err := repo.ListPacksToDelete(ctx, expires.Add(time.Minute))
	if err != nil || len(due) != 1 || due[0].ID != pack.ID || *due[0].FileName != name {
		t.Fatalf("due after expiry = %+v, %v", due, err)
	}
	if err := repo.MarkPackDeleted(ctx, pack.ID); err != nil {
		t.Fatal(err)
	}
	if file, _ := repo.PackFile(ctx, pack.ID); file == nil || !file.Deleted || file.Status != models.ClipPackExpired {
		t.Fatalf("deleted pack file = %+v", file)
	}
	if due, err := repo.ListPacksToDelete(ctx, expires.Add(time.Minute)); err != nil || len(due) != 0 {
		t.Fatalf("due after deletion = %+v, %v", due, err)
	}
}

func TestClipPackClaimRechecksEligibility(t *testing.T) {
	pool := testutil.SetupTestDB(t)
	defer testutil.CleanupTestDB(t, pool)
	testutil.TruncateTables(t, pool, "clip_packs", "countdown_exclusions", "broadcaster_feature_optins", "clips", "users")
	ctx := context.Background()
	repo := NewCountdownRepository(pool)

	weekStart := time.Date(2026, 11, 2, 0, 0, 0, 0, time.UTC)
	ids := insertCountdownClips(t, pool, weekStart.AddDate(0, 0, 7), []countdownClip{
		{name: "kept", broadcaster: "alpha", velocity: 1, duration: 30, age: 24 * time.Hour},
		{name: "hidden-later", broadcaster: "alpha", velocity: 1, duration: 30, age: 48 * time.Hour},
	})
	userID := uuid.New()
	if _, err := pool.Exec(ctx, `INSERT INTO users (id, twitch_id, username, display_name, role, account_type)
		VALUES ($1, 'alpha', 'alpha', 'alpha', 'user', 'member')`, userID); err != nil {
		t.Fatal(err)
	}
	if _, err := pool.Exec(ctx, `INSERT INTO broadcaster_feature_optins (broadcaster_id, user_id, scope, consent_text_version)
		VALUES ('alpha', $1, 'channel:manage:clips', '2026-11-07')`, userID); err != nil {
		t.Fatal(err)
	}
	if created, err := repo.EnsureWeeklyPacks(ctx, weekStart, 10); err != nil || created != 1 {
		t.Fatalf("created = %d, %v", created, err)
	}
	if _, err := pool.Exec(ctx, `UPDATE clips SET is_hidden = true WHERE id = $1`, ids["hidden-later"]); err != nil {
		t.Fatal(err)
	}
	pack, err := repo.ClaimPackForBuild(ctx, time.Hour)
	if err != nil || pack == nil || len(pack.Clips) != 1 || pack.Clips[0].ClipID != ids["kept"] {
		t.Fatalf("claimed = %+v, %v", pack, err)
	}

	// A broadcaster who stopped featuring their clips gets an empty claim.
	if _, err := pool.Exec(ctx, `DELETE FROM broadcaster_feature_optins`); err != nil {
		t.Fatal(err)
	}
	pack, err = repo.ClaimPackForBuild(ctx, 0)
	if err != nil || pack == nil || len(pack.Clips) != 0 {
		t.Fatalf("claim after opting out = %+v, %v", pack, err)
	}
}
