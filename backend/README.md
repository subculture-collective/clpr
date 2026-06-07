# Backend

Go-based REST API backend for Clipper.

## Tech Stack

- **Go**: 1.24+
- **Framework**: Gin (HTTP web framework)
- **Database**: PostgreSQL with pgx driver
- **Cache**: Redis with go-redis client
- **Authentication**: JWT with golang-jwt
- **Configuration**: godotenv for environment variables

## Project Structure

```
backend/
├── cmd/api/          # Application entry point
│   └── main.go      # Main server file
├── internal/        # Private application code
│   ├── handlers/    # HTTP request handlers
│   ├── models/      # Domain models and DTOs
│   ├── repository/  # Database access layer
│   ├── services/    # Business logic layer
│   ├── middleware/  # HTTP middleware (auth, CORS, logging)
│   └── scheduler/   # Background job schedulers
├── pkg/             # Public packages
│   ├── database/    # Database connection pool
│   ├── jwt/         # JWT utilities
│   ├── redis/       # Redis client wrapper
│   └── twitch/      # Twitch API client
├── config/          # Configuration management
├── migrations/      # Database migrations
│   ├── *.up.sql     # Migration up files
│   ├── *.down.sql   # Migration down files
│   ├── seed.sql     # Development seed data
│   └── README.md    # Migration documentation
├── docs/            # Documentation
│   ├── authentication.md      # Auth documentation
│   ├── FFMPEG_JOB_QUEUE.md    # FFmpeg job queue system
│   └── TWITCH_INTEGRATION.md  # Twitch API docs
├── go.mod           # Go module dependencies
└── go.sum           # Dependency checksums
```

## Getting Started

### Prerequisites

- Go 1.24 or higher
- PostgreSQL 17 (via Docker Compose)
- Redis 8 (via Docker Compose)
- golang-migrate CLI tool (for database migrations)

### Setup

1. Install dependencies:

   ```bash
   go mod download
   ```

2. Install golang-migrate (if not already installed):

   ```bash
   # macOS
   brew install golang-migrate

   # Linux
   curl -L https://github.com/golang-migrate/migrate/releases/download/v4.17.0/migrate.linux-amd64.tar.gz | tar xvz
   sudo mv migrate /usr/local/bin/

   # Or using Go
   go install -tags 'postgres' github.com/golang-migrate/migrate/v4/cmd/migrate@latest
   ```

3. Copy environment configuration:

   ```bash
   cp .env.example .env
   ```

4. Edit `.env` with your configuration (database, Redis, JWT secret, Twitch API keys)

5. Start database and Redis:

   ```bash
   cd .. && docker compose up -d
   ```

6. Run database migrations:

   ```bash
   # From project root
   make migrate-up

   # Or from backend directory
   cd backend
   migrate -path migrations -database "postgresql://clpr:clpr_password@localhost:5436/clpr_db?sslmode=disable" up
   ```

7. (Optional) Seed database with sample data:

   ```bash
   make migrate-seed
   # or run the helper script (respects backend/.env)
   ./backend/scripts/seed_db.sh
   # include load-test dataset (more clips/users):
   ./backend/scripts/seed_db.sh --load-test
   ```

8. Run the server:

   ```bash
  go run ./cmd/api
   ```

The server will start on `http://localhost:8080`

## Development

### Building

```bash
# Build binary
go build -o bin/api ./cmd/api

# Run binary
./bin/api
```

### Testing

```bash
# Run all tests
go test ./...

# Run with coverage
go test -cover ./...

# Run specific package tests
go test ./internal/handlers
```

### Code Quality

```bash
# Format code
go fmt ./...

# Run linter (if installed)
golangci-lint run

# Check for common mistakes
go vet ./...
```

## Dependencies

The following dependencies will be automatically added when imported in code:

### Core Dependencies

- `github.com/gin-gonic/gin` - HTTP web framework
- `github.com/jackc/pgx/v5` - PostgreSQL driver with connection pooling
- `github.com/google/uuid` - UUID generation and parsing
- `github.com/golang-jwt/jwt/v5` - JWT authentication
- `github.com/redis/go-redis/v9` - Redis client
- `github.com/joho/godotenv` - Environment variable management
- `github.com/golang-migrate/migrate/v4` - Database migrations

### Development Tools

- `github.com/go-delve/delve` - Debugger (optional)

To add a dependency, simply import it in your code and run:

```bash
go mod tidy
```

## Database Management

### Migrations

Database migrations are located in the `migrations/` directory. See [migrations/README.md](migrations/README.md) for detailed documentation.

**Common commands:**

```bash
# Run all pending migrations
make migrate-up

# Rollback the last migration
make migrate-down

# Rollback all migrations
make migrate-down-all

# Check current migration version
make migrate-status

# Create a new migration
make migrate-create NAME=add_new_feature

# Seed database with sample data (development only)
make migrate-seed
```

### Database Schema

The database includes the following tables:

- **users** - User accounts and profiles
- **clips** - Twitch clips with metadata
- **votes** - User votes on clips
- **comments** - User comments on clips
- **comment_votes** - User votes on comments
- **favorites** - User favorite clips
- **tags** - Categorization tags
- **clip_tags** - Many-to-many clip-tag relationships
- **reports** - Content moderation reports
- **refresh_tokens** - JWT refresh token storage

See [docs/DATABASE-SCHEMA.md](../docs/DATABASE-SCHEMA.md) for complete schema documentation including:

- Entity relationship diagram
- Table structures
- Triggers and functions
- Views for common queries
- Indexes and performance optimization

### Database Models

Go models for all database tables are defined in `internal/models/models.go`:

- Type-safe representations of database entities
- JSON serialization tags for API responses
- UUID types for all primary keys
- Proper handling of nullable fields

## API Endpoints

### Health Check

- `GET /health` - Basic server health check
- `GET /health/ready` - Readiness check (includes database and Redis connectivity)
- `GET /health/live` - Liveness check
- `GET /health/stats` - Database connection pool statistics

### Authentication

- `GET /api/v1/auth/twitch` - Initiate Twitch OAuth flow
- `GET /api/v1/auth/twitch/callback` - OAuth callback handler
- `POST /api/v1/auth/refresh` - Refresh access token
- `POST /api/v1/auth/logout` - Logout user
- `GET /api/v1/auth/me` - Get current user (requires auth)

See [docs/authentication.md](docs/authentication.md) for complete authentication documentation.

### Clips (NEW!)

- `POST /api/v1/clips/request` - Submit a clip by URL (requires auth, rate limited: 5/hour)

### Admin Endpoints (NEW!)

- `POST /api/v1/admin/sync/clips` - Manually trigger clip sync (requires auth)
- `GET /api/v1/admin/sync/status` - Get sync job status (requires auth)

See [docs/TWITCH_INTEGRATION.md](docs/TWITCH_INTEGRATION.md) for complete Twitch API integration documentation.

### Comments (NEW!)

- `GET /api/v1/clips/:clipId/comments` - List comments for a clip with sorting and pagination
- `POST /api/v1/clips/:clipId/comments` - Create a new comment or reply (requires auth)
- `GET /api/v1/comments/:id/replies` - Get replies to a specific comment
- `PUT /api/v1/comments/:id` - Edit a comment (requires auth)
- `DELETE /api/v1/comments/:id` - Delete a comment (requires auth)
- `POST /api/v1/comments/:id/vote` - Vote on a comment (requires auth)

**Features:**
- Reddit-style nested threading (up to 10 levels deep)
- Markdown support with XSS protection
- Voting system with optimistic updates
- Soft-delete preserves thread structure
- Performance optimized for 1000+ comments per clip

**Examples:**

```bash
# List comments for a clip
curl "http://localhost:8080/api/v1/clips/{clipId}/comments?sort=best&limit=20"

# Create a comment
curl -X POST "http://localhost:8080/api/v1/clips/{clipId}/comments" \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer YOUR_TOKEN" \
  -d '{"content": "Great clip!"}'

# Create a nested reply
curl -X POST "http://localhost:8080/api/v1/clips/{clipId}/comments" \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer YOUR_TOKEN" \
  -d '{"content": "I agree!", "parent_comment_id": "parent-uuid"}'

# Vote on a comment (1 = upvote, -1 = downvote, 0 = remove vote)
curl -X POST "http://localhost:8080/api/v1/comments/{commentId}/vote" \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer YOUR_TOKEN" \
  -d '{"vote": 1}'
```

See [docs/backend/comment-api.md](../docs/backend/comment-api.md) for complete comment API documentation.
See [docs/features/comments.md](../docs/features/comments.md) for comprehensive feature overview and E2E testing procedures.

### API v1

- `GET /api/v1/ping` - API ping test

More endpoints will be added as features are implemented.

## Environment Variables

See `.env.example` for all available configuration options:

- **Server**: Port, Gin mode
- **Database**: Host, port, credentials, database name
- **Redis**: Host, port, password
- **JWT**: Private/public keys for authentication
- **Twitch**: OAuth credentials
- **Clip**: Submission quality and storage settings (`CLIP_*` env vars)
- **Stripe**: Payment integration
- **Email**: SendGrid integration
- **OpenSearch**: Search service connection
- **Feature Flags**: Toggle features on/off
- **Rate Limiting**: Request limits per tier
- **Recommendations**: Algorithm tuning parameters (see below)

### Recommendation Configuration

Fine-tune the hybrid recommendation algorithm:

```bash
# Hybrid algorithm weights (should sum to ~1.0)
REC_CONTENT_WEIGHT=0.5          # Content-based filtering weight (default: 0.5)
REC_COLLABORATIVE_WEIGHT=0.3    # Collaborative filtering weight (default: 0.3)
REC_TRENDING_WEIGHT=0.2         # Trending signal weight (default: 0.2)

# Collaborative filtering parameters
REC_CF_FACTORS=50               # Latent factors (default: 50)
REC_CF_REGULARIZATION=0.01      # L2 regularization (default: 0.01)
REC_CF_LEARNING_RATE=0.01       # SGD learning rate (default: 0.01)
REC_CF_ITERATIONS=20            # Training iterations (default: 20)

# General settings
REC_ENABLE_HYBRID=true          # Enable hybrid recommendations (default: true)
REC_CACHE_TTL_HOURS=24          # Cache TTL in hours (default: 24)
```

For optimization guidance, see `../docs/CF-OPTIMIZATION-RESULTS.md`.

### Clip Submission Storage Configuration

These settings are wired for backend configuration and validation; the app does not ship a direct upload flow here.

```bash
CLIP_MAX_DURATION_SECONDS=60
CLIP_RECOMMENDED_DURATION_SECONDS=60
CLIP_MAX_UPLOAD_BYTES=104857600
CLIP_ALLOWED_UPLOAD_MIME_TYPES=video/mp4,video/webm,video/quicktime
CLIP_REQUIRE_MODERATION_FOR_UPLOAD=false
CLIP_STORAGE_PROVIDER=local
CLIP_STORAGE_ENDPOINT=
CLIP_STORAGE_BUCKET=
CLIP_STORAGE_REGION=us-east-1
CLIP_STORAGE_ACCESS_KEY=
CLIP_STORAGE_SECRET_KEY=
CLIP_STORAGE_FORCE_PATH_STYLE=false
CLIP_STORAGE_PUBLIC_BASE_URL=
CLIP_MEDIA_PUBLIC_BASE_URL=
```

Direct clip media is exposed through the app-owned redirect endpoint
`/api/v1/clips/{id}/media`. When `video_url` points at object storage, clip API
responses rewrite it to that endpoint, and the endpoint redirects to the resolved
storage URL without proxying video bytes through the backend. Set
`CLIP_MEDIA_PUBLIC_BASE_URL` (for example, `https://clpr.tv/api/v1/clips`) when
clients need absolute app-owned media URLs; otherwise responses use a relative
API path.

- **Redis**: Host, port, password
- **JWT**: Secret key, token expiration
- **Twitch API**: Client ID, secret, redirect URI
- **CORS**: Allowed origins

## Project Conventions

### Code Style

- Follow standard Go conventions
- Use `gofmt` for formatting
- Use meaningful variable and function names
- Keep functions small and focused
- Document exported functions with comments

### Package Organization

- `cmd/` - Application entry points
- `internal/` - Private application code (cannot be imported by other projects)
  - `handlers/` - HTTP request handlers
  - `models/` - Domain models
  - `repository/` - Database access layer
  - `services/` - Business logic layer
  - `middleware/` - HTTP middleware
  - `scheduler/` - Background jobs
- `pkg/` - Public libraries (can be imported by other projects)
  - `database/` - Database connection pool
  - `jwt/` - JWT utilities
  - `redis/` - Redis client wrapper
  - `twitch/` - Twitch API client
- `config/` - Configuration management

### Error Handling

- Always check and handle errors
- Return errors to caller when appropriate
- Log errors with context
- Use custom error types for domain-specific errors

### Logging

- Use structured logging
- Log at appropriate levels (debug, info, warn, error)
- Include context in log messages

## Scripts

The `scripts/` directory contains standalone utilities for backend operations.

### Clip Scraper

**Purpose**: Targeted scraping of Twitch clips from broadcasters with submissions on the platform.

**Build:**
```bash
go build -o bin/scrape_clips ./scripts/scrape_clips.go
```

**Usage:**
```bash
# Basic usage (queries broadcasters from submissions)
./bin/scrape_clips

# Dry run mode (no database inserts)
./bin/scrape_clips --dry-run

# With custom options
./bin/scrape_clips --batch-size 100 --min-views 200 --max-age-days 14

# Scrape specific broadcasters
./bin/scrape_clips --broadcasters "xQc,Pokimane,shroud"
```

**Documentation**: See [scripts/README_SCRAPER.md](scripts/README_SCRAPER.md) for:
- Complete usage guide
- Configuration options
- Cron/systemd scheduling
- Troubleshooting

**Scheduling**:
- Cron examples: [scripts/cron.example](scripts/cron.example)
- Systemd service: [scripts/systemd/](scripts/systemd/)

## Next Steps

1. ✅ ~~Implement database schema and migrations~~
2. ✅ ~~Create database connection pool~~
3. ✅ ~~Define Go models for all tables~~
4. ✅ ~~Create repository layer for data access~~
5. ✅ ~~Implement authentication with Twitch OAuth~~
6. ✅ ~~Integrate Twitch API for clip fetching~~
7. ✅ ~~Implement clip sync service with scheduler~~
8. ✅ ~~Create targeted clip scraping script~~
9. ✅ ~~Add search evaluation framework~~
10. ✅ ~~Add recommendation evaluation framework~~
11. Add more business logic in services layer
12. Create HTTP handlers for remaining API endpoints
13. Add comprehensive tests for all components
14. Add monitoring and metrics

## Algorithm Evaluation

The backend includes evaluation frameworks for assessing the quality of search and recommendation algorithms:

### Search Evaluation
- **Metrics**: nDCG, MRR, Precision@k, Recall@k
- **Dataset**: `testdata/search_evaluation_dataset.yaml`
- **CLI tool**: `cmd/evaluate-search`
- **Makefile**: `make evaluate-search` or `make evaluate-search-json`

### Recommendation Evaluation
- **Metrics**: Precision@k, Recall@k, nDCG, Diversity, Serendipity, Cold-start performance
- **Dataset**: `testdata/recommendation_evaluation_dataset.yaml`
- **CLI tools**:
  - `cmd/evaluate-recommendations` - Run evaluations
  - `cmd/grid-search-recommendations` - Parameter optimization
- **Makefile**:
  - `make evaluate-recommendations` or `make evaluate-recommendations-json`
  - `make grid-search-recommendations` or `make grid-search-recommendations-full`
- **Documentation**:
  - `../docs/RECOMMENDATION-EVALUATION.md` - Evaluation framework
  - `../docs/CF-OPTIMIZATION-RESULTS.md` - Optimization results and A/B test plan
- **CI**: Runs nightly via GitHub Actions (`.github/workflows/recommendation-evaluation.yml`)
- **Configuration**: See environment variables section for tuning parameters

Both frameworks support:
- Automated testing via CI/CD
- Baseline measurement tracking
- Target/threshold monitoring
- JSON output for analysis and trend tracking

## Resources

- [Go Documentation](https://go.dev/doc/)
- [Gin Framework](https://gin-gonic.com/docs/)
- [pgx Documentation](https://pkg.go.dev/github.com/jackc/pgx/v5)
- [golang-migrate](https://github.com/golang-migrate/migrate)
- [golang-jwt](https://github.com/golang-jwt/jwt)
- [Database Schema Documentation](../docs/DATABASE-SCHEMA.md)
- [Authentication Documentation](docs/authentication.md)
- [Twitch API Integration Documentation](docs/TWITCH_INTEGRATION.md)
- [FFmpeg Job Queue Documentation](docs/FFMPEG_JOB_QUEUE.md)
- [Twitch API Reference](https://dev.twitch.tv/docs/api/)
