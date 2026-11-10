# clpr

![clpr: Somebody clipped it. Clip-stack artwork in purple on dark ink.](docs/assets/readme/banner.png)

**Somebody clipped it. It's probably in here.**

clpr sorts Twitch clips by creator, topic and tag. Search for the one you half
remember, save clips into playlists, and send a playlist to the people who
missed the stream.

clpr is a responsive web app, live at [clpr.tv](https://clpr.tv). The
[launch feature inventory](docs/LAUNCH_FEATURE_INVENTORY.md) lists what is
supported today and which features were removed.

![clpr home and clip discovery](https://subcult.tv/screenshots/clpr-home-1440.webp)

## Features

- **Browse:** clips by creator, topic, tag and Twitch category, with trending,
  hot and new feeds and leaderboards.
- **Search:** full-text search over titles, creators and tags, plus transcript
  search for clips whose speech has been transcribed. The `said` filter keeps
  only clips that contain a phrase.
- **Playlists:** create, share and bookmark playlists, browse public ones, and
  play a playlist or your queue in theatre mode. Watch history is kept per
  account.
- **Submit:** send in an existing Twitch clip from the site or from the
  [browser extension](extension/README.md).
- **Vote and comment:** upvotes and comments move clips in the feed.
- **Countdowns:** a daily ranked list of clips per series, each with a public
  page. Operators can render editions as videos and Shorts.
- **Streamer tools:** a clip room that collects clip links posted in a
  streamer's Twitch chat, a creator dashboard and analytics, and weekly clip
  packs for connected streamers.
- **Webhooks:** subscribe to clip submission and review events, daily top clips
  and published countdowns, with optional Discord delivery.
- **Accounts:** Twitch sign-in, optional two-factor authentication,
  notification preferences and cookie settings.
- **Moderation:** submission review, reports, site bans, appeals, an audit log
  and DMCA handling.

Automated account deletion and automated data export are not available.

## Quick start

You need Go (version in `backend/go.mod`), Node.js 24, Docker,
[Task](https://taskfile.dev/) and
[golang-migrate](https://github.com/golang-migrate/migrate).

Run the test suite against disposable PostgreSQL, Redis and OpenSearch
containers. No production credentials are needed:

```bash
cp backend/.env.example backend/.env
cp frontend/.env.example frontend/.env
task test:setup
task test
npm --prefix frontend ci
npm --prefix frontend run test:coverage
npm --prefix frontend run build
task test:teardown
```

To run the app locally, start PostgreSQL 17 with pgvector
(`Dockerfile.postgres` builds one), Redis and OpenSearch, point `backend/.env`
at them, then:

```bash
task install
task migrate:up      # uses the database URL in Taskfile.yml
task backend-dev     # API on PORT from backend/.env
task frontend-dev    # Vite dev server on port 5173
```

`task --list` shows every task. `make <target>` forwards to the same tasks.

## Configuration

The backend reads environment variables; `backend/.env.example` lists them with
comments, and `backend/config/config.go` has the full set and defaults. The
main groups:

- **Server:** `ENVIRONMENT`, `GIN_MODE`, `PORT`, `BASE_URL`,
  `CORS_ALLOWED_ORIGINS`, `WEBSOCKET_ALLOWED_ORIGINS`, `TRUSTED_PROXIES`
- **Data:** `DB_HOST`, `DB_PORT`, `DB_NAME`, `DB_USER`, `DB_PASSWORD`,
  `DB_SSLMODE`, `REDIS_*`, `OPENSEARCH_*`
- **Auth:** `JWT_PRIVATE_KEY`, `JWT_PUBLIC_KEY`, `MFA_ENCRYPTION_KEY`,
  `TWITCH_CLIENT_ID`, `TWITCH_CLIENT_SECRET`, `TWITCH_REDIRECT_URI`,
  `TWITCH_BOT_USERNAME`, `TWITCH_BOT_OAUTH_TOKEN`
- **Storage:** `CLIP_STORAGE_*`, `CLIP_MEDIA_PUBLIC_BASE_URL`
- **Optional features:** `FEATURE_*` flags, `EMAIL_*` and `SENDGRID_*`,
  `EMBEDDING_*`, `VISION_*`, `WHISPER_*`, `SENTRY_*`, `TELEMETRY_*`,
  `CLIP_PACK_*`

The frontend reads `VITE_*` variables listed in `frontend/.env.example`. The
root `.env.example` is the combined file used by `docker-compose.yml`.

Production profiles refuse to start without HTTPS, durable object storage and
verified database TLS; the rules are in `backend/config/validate.go`.

## Deploy

clpr runs as container images built from this repository:

- `backend/Dockerfile`: the Go API (port 8080)
- `backend/Dockerfile.crawler`: the clip scraper
- `backend/Dockerfile.render`: the countdown media and transcription worker
- `frontend/Dockerfile`: the built web app served by nginx (port 8080)

`docker-compose.yml` shows how the API, scraper and frontend fit together. It
expects PostgreSQL with pgvector, Redis and OpenSearch to be reachable, an
`.env` file, and two existing Docker networks (`projects` and `dev`). Run
`task migrate:up` (or `migrate` against your database) before starting a new
release.

## Project layout

```text
backend/     Go API, workers, migrations (backend/migrations) and tests
frontend/    React, TypeScript and Vite web app
extension/   Browser extension for submitting Twitch clips
docs/        Product, API (OpenAPI) and developer documentation
scripts/     Test, documentation and release scripts
```

More reading: the [development guide](DEVELOPMENT.md), the
[backend README](backend/README.md) and the
[API reference](docs/openapi/README.md).

## Contributing

Open a pull request on GitHub against `main`. CI runs the backend and frontend
test suites, linting, end-to-end browser tests and documentation checks on every
pull request; the `Mirror CI gate` check must pass before a pull request can
merge. For documentation changes, run `npm run docs:check` first. See the
[contributor guide](docs/contributing.md).

## License

clpr is released under the [MIT License](LICENSE). Built by
[SUBCULT](https://subcult.tv).
