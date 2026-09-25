# Site Freshness Automation

Clipper already has a built-in scheduler for playlist scripts. This guide explains the easiest way to use it so the public playlist catalog keeps updating without manual curation every day.

## What is already automated

The backend starts several recurring jobs when the API boots. For freshness, the key one is the **playlist script scheduler**, which:

- checks every 5 minutes for due playlist scripts
- generates new playlists for scripts on a schedule
- cleans up old generated playlists after their retention window
- uses a per-script PostgreSQL advisory lock to prevent duplicate generation across API replicas

Relevant backend pieces:

- `backend/cmd/api/schedulers.go`
- `backend/internal/scheduler/playlist_script_scheduler.go`
- `backend/internal/services/playlist_script_service.go`

## The quickest win

Use the default **site freshness** bootstrap command. It creates a starter pack of public smart playlists owned by the system bot user.

Included presets:

- `Viral Velocity`
- `Fresh Faces`
- `Deep Cuts Weekly`
- `Trending Now` *(only when Twitch credentials are configured)*
- `Discovery Mix` *(only when Twitch credentials are configured)*

These rules are intentionally idempotent: rerunning the bootstrap skips presets that already exist for the configured owner.

## How to enable it

From the repository root:

```bash
make site-freshness-seed
```

If you want an immediate batch of playlists right away instead of waiting for the next scheduler pass:

```bash
make site-freshness-generate
```

## How it behaves

- Creates default public playlist scripts if missing
- Uses the well-known `clpr-bot` owner by default
- Adds Twitch-backed presets only when `TWITCH_CLIENT_ID` and `TWITCH_CLIENT_SECRET` are configured
- When `-generate-now` is used, it tries to generate the first playlists immediately
- Otherwise, the API scheduler will pick them up on its next due run

## Custom owner

You can assign the presets to a different user if needed:

```bash
cd backend
go run ./cmd/seed-site-freshness -owner <user-uuid>
```

## Recommended operating model

For most environments:

1. run migrations
2. ensure the backend API is running with schedulers enabled
3. run `make site-freshness-seed`
4. optionally run `make site-freshness-generate`

After that, the scheduler keeps the public playlist shelf refreshed automatically.

Suppressed tags never include or exclude clips from generated playlists. If a
script includes only a suppressed tag, the run is acknowledged as empty and
the previous generated playlist remains available until normal retention.

## Notes

- Generated playlists are additive; retention windows are what prevent the public catalog from filling with stale entries forever.
- If Twitch credentials are unavailable, the database-only presets still provide rotating public playlists from existing content.
- You can still manage or extend rules manually in the admin playlist scripts UI.
