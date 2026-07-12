---
title: "TOXICITY CLASSIFICATION"
summary: "The Toxicity Classification System is an ML-based moderation feature that automatically detects and flags toxic comments in the Clipper platform. This system integrates with Google's Perspective API ("
tags: ["docs"]
area: "docs"
status: "stable"
owner: "team-core"
version: "1.0"
last_reviewed: 2026-01-29
---

# Toxicity Classification System

## Overview

The Toxicity Classification System is an ML-based moderation feature that automatically detects and flags toxic comments in the Clipper platform. This system integrates with Google's Perspective API (or similar services) to provide real-time toxicity analysis with high precision and recall.

## Features

- **Real-time Classification**: Comments are analyzed for toxicity immediately upon submission
- **Async Processing**: Classification runs asynchronously to avoid blocking comment creation
- **Multi-category Detection**: Detects various forms of toxicity including:
  - General toxicity
  - Severe toxicity
  - Identity attacks
  - Insults
  - Profanity
  - Threats
  - Sexually explicit content
- **Automatic Flagging**: Comments exceeding the confidence threshold are automatically added to the moderation queue
- **Graceful Degradation**: Falls back to safe defaults if the ML service is unavailable
- **Metrics Tracking**: Comprehensive metrics for monitoring model performance

## Configuration

Add the following environment variables to your `.env` file:

```bash
# Enable toxicity detection
TOXICITY_ENABLED=false

# API key for Perspective API (get from https://developers.perspectiveapi.com/s/docs-get-started)
TOXICITY_API_KEY=your-api-key-here

# API URL (default is Perspective API)
TOXICITY_API_URL=https://commentanalyzer.googleapis.com/v1alpha1/comments:analyze

# Confidence threshold for auto-flagging (0.0 to 1.0)
# Recommended: 0.85 (85% confidence)
TOXICITY_THRESHOLD=0.85
```

## Database Schema

The system uses two primary tables:

### toxicity_predictions
Stores ML model predictions for each comment:
- `comment_id`: Reference to the comment
- `toxic`: Boolean flag indicating if the comment is toxic
- `confidence_score`: Model confidence (0.0 to 1.0)
- `categories`: JSONB object with scores per toxicity category
- `reason_codes`: Array of triggered categories
- `model_version`: Version identifier of the model used

### toxicity_review_feedback
Tracks moderator feedback for model improvement:
- `prediction_id`: Reference to the prediction
- `reviewer_id`: Moderator who reviewed the prediction
- `actual_toxic`: Ground truth label from human review
- `feedback_notes`: Optional notes from the reviewer

## API Endpoints

### Get Toxicity Metrics
```
GET /api/v1/admin/moderation/toxicity/metrics
```

Query Parameters:
- `start_date` (optional): Start date in YYYY-MM-DD format (default: 30 days ago)
- `end_date` (optional): End date in YYYY-MM-DD format (default: today)

Response:
```json
{
  "success": true,
  "data": {
    "total_predictions": 1250,
    "total_toxic": 87,
    "toxicity_rate": 0.0696,
    "precision": 0.891,
    "recall": 0.873,
    "false_positive_rate": 0.048,
    "true_positives": 76,
    "false_positives": 11,
    "false_negatives": 15
  },
  "start_date": "2024-01-01",
  "end_date": "2024-01-31"
}
```

## Workflow

1. **Comment Submission**: User submits a comment
2. **Validation**: Comment passes standard validation checks
3. **Database Storage**: Comment is saved to the database
4. **Async Classification**: Toxicity classification runs in the background
5. **Prediction Storage**: Classification results are stored in `toxicity_predictions`
6. **Auto-Flagging**: If confidence ≥ threshold, comment is added to moderation queue
7. **Human Review**: Moderators review flagged comments
8. **Feedback Loop**: Review decisions update metrics for monitoring

## Integration with Perspective API

### Getting Started

1. Visit [Perspective API documentation](https://developers.perspectiveapi.com/s/docs-get-started)
2. Request an API key
3. Add the API key to your `.env` file
4. Enable the feature with `TOXICITY_ENABLED=true`

### API Usage

The system makes requests to Perspective API with the following attributes:
- `TOXICITY`: General toxicity detection
- `SEVERE_TOXICITY`: Severe forms of toxicity
- `IDENTITY_ATTACK`: Attacks on identity groups
- `INSULT`: Insulting content
- `PROFANITY`: Profane language
- `THREAT`: Threatening content
- `SEXUALLY_EXPLICIT`: Sexual content

### Rate Limits

Perspective API has the following rate limits:
- Free tier: 1 QPS (query per second)
- Standard tier: 10 QPS
- Enterprise tier: Custom limits

The system handles rate limiting gracefully by:
- Processing comments asynchronously
- Not blocking comment creation
- Logging errors for monitoring

## Metrics and Monitoring

### Key Metrics

The system tracks the following metrics:

1. **Precision**: TP / (TP + FP)
   - Percentage of flagged comments that are actually toxic
   - Target: ≥85%

2. **Recall**: TP / (TP + FN)
   - Percentage of toxic comments that are flagged
   - Target: ≥85%

3. **False Positive Rate**: FP / (FP + TN)
   - Percentage of non-toxic comments incorrectly flagged
   - Target: <5%

4. **Toxicity Rate**: Total toxic / Total predictions
   - Overall percentage of toxic comments

### Viewing Metrics

Access the metrics dashboard:
```bash
GET /api/v1/admin/moderation/toxicity/metrics?start_date=2024-01-01&end_date=2024-01-31
```

### Database Views

The system includes a `toxicity_metrics` view for easy querying:
```sql
SELECT * FROM toxicity_metrics
WHERE date >= '2024-01-01'
ORDER BY date DESC;
```

## Moderation Queue Integration

Toxic comments are automatically added to the moderation queue with:
- `content_type`: "comment"
- `reason`: Mapped from toxicity categories (e.g., "toxic", "harassment", "offensive")
- `priority`: Based on confidence score (50-100 scale)
- `auto_flagged`: true
- `confidence_score`: Model confidence value
- `metadata`: JSONB with model version and category details

Moderators can:
1. View flagged comments in the moderation queue
2. Review content and make decisions (approve/reject)
3. See confidence scores and reason codes
4. Provide feedback for model improvement

## Fallback Behavior

If the ML service is unavailable or disabled:
1. Comment creation proceeds normally
2. No classification is performed
3. Comments are not auto-flagged
4. Manual moderation remains available
5. No errors are shown to users

This ensures the platform remains functional even when the ML service is down.

## Best Practices

### Threshold Tuning

The confidence threshold balances precision and recall:
- **Lower threshold (0.75-0.80)**: Higher recall, more false positives
- **Medium threshold (0.80-0.85)**: Balanced approach (recommended)
- **Higher threshold (0.85-0.90)**: Higher precision, fewer false positives

Monitor metrics regularly and adjust the threshold based on:
- Moderator workload
- Community feedback
- Precision/recall targets

### Human Review

Always maintain human oversight:
1. Review auto-flagged content regularly
2. Provide feedback on predictions
3. Use feedback to improve the system
4. Don't rely solely on automated moderation

### Privacy and Ethics

- Comments are analyzed only for moderation purposes
- No personal data is sent to external services beyond comment text
- Classification results are stored securely
- Users are notified of moderation actions
- Appeals process is available

## Troubleshooting

### Common Issues

**Issue**: High false positive rate
- **Solution**: Increase the confidence threshold
- **Check**: Review categories in metadata to identify problematic patterns

**Issue**: Missing toxic comments (low recall)
- **Solution**: Decrease the confidence threshold
- **Check**: Verify API key is valid and service is responding

**Issue**: API errors in logs
- **Solution**: Check API key, rate limits, and network connectivity
- **Check**: Verify TOXICITY_API_URL is correct

**Issue**: Classification not running
- **Solution**: Ensure TOXICITY_ENABLED=true in environment
- **Check**: Verify service is initialized in logs

### Logs

Classification errors are logged but don't block comment creation:
```
Warning: failed to classify comment {id} for toxicity: {error}
Warning: failed to add comment {id} to moderation queue: {error}
```

## Performance Considerations

- Classification runs asynchronously (doesn't block comment creation)
- Typical response time: 200-500ms per comment
- Database queries are optimized with indexes
- Metrics queries use efficient CTEs
- No impact on comment submission latency

## Roadmap

Future improvements:
- [ ] Support for additional ML providers
- [ ] Custom model training on platform-specific data
- [ ] Multi-language support
- [ ] Real-time confidence adjustment
- [ ] Advanced analytics dashboard
- [ ] A/B testing framework for threshold optimization

## Support

For issues or questions:
1. Check logs for error messages
2. Verify configuration in `.env`
3. Test API connectivity manually
4. Review metrics for anomalies
5. Contact support with reproduction steps

## References

- [Perspective API Documentation](https://developers.perspectiveapi.com/s/docs)
- [Perspective API Best Practices](https://developers.perspectiveapi.com/s/docs-best-practices)
- Moderation Queue System
- [Database Migrations](../../backend/migrations/000087_add_toxicity_classification.up.sql)
