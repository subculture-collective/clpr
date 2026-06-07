---
title: "NSFW ENV VARS"
summary: "This file documents the environment variables required for NSFW (Not Safe For Work) image detection."
tags: ["docs"]
area: "docs"
status: "stable"
owner: "team-core"
version: "1.0"
last_reviewed: 2026-01-29
---

# NSFW Detection Environment Variables

This file documents the environment variables required for NSFW (Not Safe For Work) image detection.

## Configuration Variables

### NSFW_ENABLED
- **Type**: Boolean
- **Default**: `false`
- **Description**: Master switch to enable/disable NSFW detection system
- **Example**: `NSFW_ENABLED=true`

### NSFW_API_KEY
- **Type**: String
- **Required**: Yes (when enabled)
- **Description**: API key for the NSFW detection service provider
- **Example**: `NSFW_API_KEY=your_api_key_here`
- **Security**: Keep this secret and never commit to version control

### NSFW_API_URL
- **Type**: String (URL)
- **Required**: Yes (when enabled)
- **Description**: API endpoint URL for the NSFW detection service
- **Example**: `NSFW_API_URL=https://api.sightengine.com/1.0/check.json`
- **Supported Services**:
  - Sightengine: `https://api.sightengine.com/1.0/check.json`
  - AWS Rekognition: Configure via AWS SDK
  - Google Cloud Vision: `https://vision.googleapis.com/v1/images:annotate`
  - Azure Content Moderator: Your Azure endpoint

### NSFW_THRESHOLD
- **Type**: Float (0.0 to 1.0)
- **Default**: `0.80`
- **Description**: Confidence threshold above which content is flagged as NSFW
- **Example**: `NSFW_THRESHOLD=0.85`
- **Tuning Guide**:
  - `0.70-0.75`: More lenient, higher false positives
  - `0.80-0.85`: Balanced (recommended)
  - `0.90-0.95`: Strict, may miss some NSFW content

### NSFW_SCAN_THUMBNAILS
- **Type**: Boolean
- **Default**: `true`
- **Description**: Enable automatic scanning of thumbnails at upload time
- **Example**: `NSFW_SCAN_THUMBNAILS=true`

### NSFW_AUTO_FLAG
- **Type**: Boolean
- **Default**: `true`
- **Description**: Automatically add detected NSFW content to moderation queue
- **Example**: `NSFW_AUTO_FLAG=true`
- **Note**: Set to `false` to only log detections without flagging

### NSFW_MAX_LATENCY_MS
- **Type**: Integer (milliseconds)
- **Default**: `200`
- **Description**: Target maximum latency for NSFW detection (p95)
- **Example**: `NSFW_MAX_LATENCY_MS=200`
- **Note**: Used for monitoring and alerting, not enforcement

### NSFW_TIMEOUT_SECONDS
- **Type**: Integer (seconds)
- **Default**: `5`
- **Description**: HTTP request timeout for NSFW detection API calls
- **Example**: `NSFW_TIMEOUT_SECONDS=5`
- **Range**: 1-30 seconds recommended

## Example .env Configuration

### Development
```bash
# NSFW Detection - Development
NSFW_ENABLED=true
NSFW_API_KEY=dev_test_key_12345
NSFW_API_URL=https://api.sightengine.com/1.0/check.json
NSFW_THRESHOLD=0.75
NSFW_SCAN_THUMBNAILS=true
NSFW_AUTO_FLAG=false  # Log only in dev
NSFW_MAX_LATENCY_MS=200
NSFW_TIMEOUT_SECONDS=5
```

### Production
```bash
# NSFW Detection - Production
NSFW_ENABLED=true
NSFW_API_KEY=${SECRET_NSFW_API_KEY}  # From secrets manager
NSFW_API_URL=https://api.sightengine.com/1.0/check.json
NSFW_THRESHOLD=0.80
NSFW_SCAN_THUMBNAILS=true
NSFW_AUTO_FLAG=true
NSFW_MAX_LATENCY_MS=200
NSFW_TIMEOUT_SECONDS=5
```

### Disabled (Default)
```bash
# NSFW Detection - Disabled
NSFW_ENABLED=false
# Other variables not required when disabled
```

## Database Migration

Before enabling NSFW detection, run the database migration:

```bash
# Apply the NSFW detection migration
migrate -path ./migrations -database "postgresql://user:pass@localhost:5432/clpr_db?sslmode=disable" up

# Or using make
make migrate-up
```

This creates:
- `nsfw_detection_metrics` table
- `nsfw_scan_jobs` table
- Moderation queue NSFW columns
- Required indexes

## Monitoring & Metrics

When enabled, NSFW detection exposes Prometheus metrics at `/metrics`:

- `nsfw_detection_total{result="safe|nsfw|error"}`: Detection counts
- `nsfw_detection_latency_ms`: Detection latency histogram
- `nsfw_content_flagged_total{content_type}`: Flagged content counts
- `nsfw_detection_errors_total{error_type}`: Error counts

## Testing

To test the NSFW detection system:

1. Enable with development settings
2. Call the health check endpoint:
   ```bash
   curl -H "Authorization: Bearer $ADMIN_TOKEN" \
        http://localhost:8080/admin/nsfw/health
   ```
3. Test detection on a sample image:
   ```bash
   curl -X POST -H "Authorization: Bearer $ADMIN_TOKEN" \
        -H "Content-Type: application/json" \
        -d '{"image_url":"https://example.com/test.jpg","content_type":"thumbnail"}' \
        http://localhost:8080/admin/nsfw/detect
   ```

## Security Notes

1. **API Key Protection**: 
   - Never commit API keys to version control
   - Use environment variables or secrets management
   - Rotate keys regularly

2. **API Rate Limits**:
   - Check your provider's rate limits
   - Implement backoff strategies if needed
   - Monitor API usage metrics

3. **Privacy**:
   - External API calls send image URLs to third-party services
   - Ensure compliance with privacy policies
   - Consider on-premise solutions for sensitive content

4. **False Positives**:
   - Review flagged content regularly
   - Adjust threshold based on false positive rate
   - Implement appeals process for users

## Troubleshooting

### NSFW detection not working
1. Check `NSFW_ENABLED=true`
2. Verify API key and URL are correct
3. Check network connectivity to API endpoint
4. Review logs for error messages
5. Test with health check endpoint

### High latency
1. Check API provider status
2. Reduce `NSFW_TIMEOUT_SECONDS` if too high
3. Consider caching results for frequently checked images
4. Monitor network latency to API endpoint

### Too many false positives
1. Increase `NSFW_THRESHOLD` (e.g., from 0.80 to 0.85)
2. Review detection categories
3. Consider different API provider
4. Implement manual review workflow

### Too many false negatives
1. Decrease `NSFW_THRESHOLD` (e.g., from 0.80 to 0.75)
2. Enable multiple detection categories
3. Review API provider documentation
4. Consider ensemble of multiple APIs

## Related Documentation

- [NSFW Detection API](./NSFW_DETECTION.md)
- [Moderation Queue](./MODERATION_QUEUE.md)
- [Environment Variables](./ENVIRONMENT_VARIABLES.md)
