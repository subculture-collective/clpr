# Clipper Targeted Scraping Script

## Overview

The `scrape_clips.go` script is a targeted Twitch clip scraper that focuses on broadcasters who have had clips submitted on the Clipper platform. This improves content relevance and reduces noise compared to broad scraping.

## Features

- ✅ Queries broadcasters from recent clip submissions
- ✅ Fetches clips only from targeted broadcasters
- ✅ Deduplication to prevent duplicate clips
- ✅ Configurable filters (view count, clip age)
- ✅ Dry-run mode for testing
- ✅ Manual broadcaster list support
- ✅ Detailed logging and metrics
- ✅ Rate limiting via Twitch API client

## Prerequisites

- Go 1.24.7 or higher
- PostgreSQL database with Clipper schema
- Redis instance
- Twitch API credentials (Client ID and Secret)

## Configuration

The script uses environment variables from `.env` file:

```bash
# Twitch API
TWITCH_CLIENT_ID=your_client_id
TWITCH_CLIENT_SECRET=your_client_secret

# Database
DB_HOST=localhost
DB_PORT=5432
DB_USER=clpr
DB_PASSWORD=your_password
DB_NAME=clpr_db
DB_SSLMODE=disable

# Redis
REDIS_HOST=localhost
REDIS_PORT=6379
REDIS_PASSWORD=
REDIS_DB=0
```

## Usage

### Build the Script

```bash
cd backend
go build -o bin/scrape_clips ./scripts/scrape_clips.go
```

### Run the Script

#### Basic Usage

```bash
./bin/scrape_clips
```

#### With Custom Options

```bash
# Dry run mode (no database inserts)
./bin/scrape_clips --dry-run

# Custom batch size (clips per broadcaster)
./bin/scrape_clips --batch-size 100

# Filter by minimum views
./bin/scrape_clips --min-views 500

# Custom lookback period for submissions
./bin/scrape_clips --lookback-days 60

# Scrape specific broadcasters
./bin/scrape_clips --broadcasters "xQc,Pokimane,shroud"

# Combine options
./bin/scrape_clips --dry-run --batch-size 50 --min-views 200 --max-age-days 14
```

### Command-Line Flags

| Flag | Type | Default | Description |
|------|------|---------|-------------|
| `--dry-run` | bool | false | Dry run mode - don't insert clips into database |
| `--batch-size` | int | 50 | Number of clips to fetch per broadcaster |
| `--max-age-days` | int | 30 | Maximum age of clips to scrape (in days) |
| `--min-views` | int | 100 | Minimum view count for clips |
| `--lookback-days` | int | 30 | Number of days to look back for submissions |
| `--broadcasters` | string | "" | Comma-separated list of broadcaster names (overrides DB query) |

## How It Works

1. **Query Broadcasters**: The script queries the `clip_submissions` table to find broadcasters with submissions in the last N days (default: 30)

2. **Fetch Clips**: For each broadcaster:
   - Retrieves broadcaster ID from Twitch API
   - Fetches recent clips within the specified time window
   - Applies filters (min views, max age)

3. **Deduplication**: Checks if each clip already exists in the database by `twitch_clip_id`

4. **Insert Clips**: Inserts new clips into the `clips` table with metadata from Twitch

5. **Logging**: Provides detailed progress and summary statistics

## Scheduling with Cron

### Daily Run (2 AM UTC)

Add to your crontab:

```bash
# Clipper clip scraper - runs daily at 2 AM UTC
0 2 * * * cd /path/to/clpr/backend && ./bin/scrape_clips >> /var/log/clpr/scraper.log 2>&1
```

### Weekly Run (Sunday 2 AM UTC)

```bash
# Clipper clip scraper - runs weekly on Sunday at 2 AM UTC
0 2 * * 0 cd /path/to/clpr/backend && ./bin/scrape_clips --batch-size 100 >> /var/log/clpr/scraper.log 2>&1
```

### With Systemd Timer

Create `/etc/systemd/system/clpr-scraper.service`:

```ini
[Unit]
Description=Clipper Clip Scraper
After=network.target postgresql.service redis.service

[Service]
Type=oneshot
User=clpr
WorkingDirectory=/opt/clpr/backend
Environment="PATH=/usr/local/go/bin:/usr/bin:/bin"
ExecStart=/opt/clpr/backend/bin/scrape_clips
StandardOutput=journal
StandardError=journal
```

Create `/etc/systemd/system/clpr-scraper.timer`:

```ini
[Unit]
Description=Run Clipper Scraper Daily
Requires=clpr-scraper.service

[Timer]
OnCalendar=daily
OnCalendar=02:00:00
Persistent=true

[Install]
WantedBy=timers.target
```

Enable and start:

```bash
sudo systemctl daemon-reload
sudo systemctl enable clpr-scraper.timer
sudo systemctl start clpr-scraper.timer
```

## Monitoring

### View Logs

```bash
# Systemd journal
sudo journalctl -u clpr-scraper.service -f

# Log file
tail -f /var/log/clpr/scraper.log
```

### Key Metrics

The script outputs the following metrics:

- **Duration**: Total execution time
- **Broadcasters fetched**: Number of broadcasters found
- **Broadcasters scraped**: Number of broadcasters successfully processed
- **Clips checked**: Total clips examined
- **Clips inserted**: New clips added to database
- **Clips skipped**: Duplicates or filtered clips
- **Errors**: Number of errors encountered
- **API calls made**: Twitch API call count

### Example Output

```
=== Clipper Targeted Scraper ===
Configuration:
  Dry Run: false
  Batch Size: 50 clips per broadcaster
  Max Age: 30 days
  Min Views: 100
  Lookback: 30 days

✓ Database connection established
✓ Redis connection established
✓ Twitch API client initialized

Found 15 broadcasters with submissions in the last 30 days

Starting clip scraping...
[1/15] Processing broadcaster: xQc
  Broadcaster ID: 71092938
  Retrieved 50 clips from Twitch API
  ✓ Processed: 12 clips added, 38 skipped
[2/15] Processing broadcaster: Pokimane
  ...

=== Scraping Summary ===
Duration: 2m15s
Broadcasters fetched: 15
Broadcasters scraped: 15
Clips checked: 750
Clips inserted: 127
Clips skipped: 623
Errors: 0
API calls made: 30
Average time per clip inserted: 1.06s
✓ Scraping completed successfully!
```

## Troubleshooting

### Common Issues

#### "TWITCH_CLIENT_ID and TWITCH_CLIENT_SECRET must be set"

**Solution**: Ensure your `.env` file contains valid Twitch API credentials or set them as environment variables.

#### "Failed to connect to database"

**Solution**: 
- Verify database is running: `psql -h localhost -U clpr -d clpr_db`
- Check connection settings in `.env`
- Ensure database migrations have been run

#### "Rate limited by Twitch"

**Solution**: The Twitch client has built-in rate limiting (800 req/min). If you hit the limit:
- Reduce `--batch-size`
- Run less frequently
- The script will automatically retry with exponential backoff

#### "No broadcasters to scrape"

**Solution**: 
- Check if there are submissions in the database: `SELECT COUNT(*) FROM clip_submissions WHERE created_at > NOW() - INTERVAL '30 days';`
- Try increasing `--lookback-days`
- Use `--broadcasters` flag to manually specify broadcasters

## Best Practices

1. **Start with Dry Run**: Always test with `--dry-run` first
2. **Monitor API Usage**: Keep track of Twitch API quota (800 req/min)
3. **Gradual Rollout**: Start with small batch sizes and increase gradually
4. **Log Rotation**: Set up log rotation to prevent disk space issues
5. **Error Alerting**: Configure alerting for scraper failures (e.g., via Sentry)
6. **Database Backups**: Ensure backups before running large scrapes

## Performance Considerations

- **Batch Size**: Higher batch sizes = fewer API calls but more data per call
- **Lookback Days**: More days = more broadcasters = longer execution time
- **Min Views Filter**: Higher threshold = fewer clips to process
- **Concurrent Execution**: Do not run multiple instances simultaneously (no locking mechanism)

## Database Impact

The script performs the following database operations:

- **Read**: 1 query to fetch broadcasters from `clip_submissions`
- **Write**: 1 insert per new clip into `clips` table
- **Check**: 1 existence check per clip via `twitch_clip_id` index

The `twitch_clip_id` index ensures fast duplicate detection.

## Future Enhancements

Potential improvements (not yet implemented):

- [ ] Parallel processing of broadcasters
- [ ] Game name lookup and caching
- [ ] Progress persistence (resume from interruption)
- [ ] Webhook notifications on completion
- [ ] Prometheus metrics export
- [ ] Automatic retry on failure
- [ ] Priority queuing for high-value broadcasters

## Support

For issues or questions:
- Check the troubleshooting section above
- Review logs for detailed error messages
- Open an issue in the repository
