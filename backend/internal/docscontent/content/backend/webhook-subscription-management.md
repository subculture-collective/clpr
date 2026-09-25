---
title: "Webhook Subscription Management"
summary: "The webhook subscription management feature allows users to create and manage webhook subscriptions for receiving real-time notifications when events occur in the system."
tags: ["backend"]
area: "backend"
status: "stable"
owner: "team-core"
version: "1.0"
last_reviewed: 2026-01-29
---

# Webhook Subscription Management

## Overview

The webhook subscription management feature allows users to create and manage webhook subscriptions for receiving real-time notifications when events occur in the system.

## Quick Links

- **Webhook Signature Verification Guide** - Complete guide with examples in 7+ languages
- **Working Examples** - Test servers and sample code you can run immediately
- **Test Payloads** - Pre-signed sample payloads for testing your integration

## Features

### User Interface

- **Access**: Navigate to Settings > Webhooks or directly to `/settings/webhooks`
- **CRUD Operations**: Create, read, update, and delete webhook subscriptions
- **Event Selection**: Subscribe to specific events (clip.submitted, clip.approved, clip.rejected)
- **Secret Management**: Secure secret generation and rotation
- **Delivery History**: View audit log of webhook deliveries with status and error messages

### API Endpoints

All webhook endpoints require authentication.

#### Get Supported Events

```
GET /api/v1/webhooks/events
```
Returns a list of available webhook events.

#### Create Webhook Subscription

```
POST /api/v1/webhooks
Content-Type: application/json

{
  "url": "https://example.com/webhook",
  "events": ["clip.submitted", "clip.approved"],
  "description": "My webhook for clip notifications"
}
```
Returns the created subscription and the secret (only shown once).

#### List Webhook Subscriptions

```
GET /api/v1/webhooks
```
Returns all webhook subscriptions for the authenticated user.

#### Get Webhook Subscription

```
GET /api/v1/webhooks/:id
```
Returns details of a specific webhook subscription.

#### Update Webhook Subscription

```
PATCH /api/v1/webhooks/:id
Content-Type: application/json

{
  "url": "https://example.com/webhook-new",
  "events": ["clip.submitted"],
  "is_active": false,
  "description": "Updated description"
}
```
All fields are optional.

#### Delete Webhook Subscription

```
DELETE /api/v1/webhooks/:id
```
Deletes the webhook subscription.

#### Regenerate Secret

```
POST /api/v1/webhooks/:id/regenerate-secret
```
Generates a new secret for the webhook subscription. The old secret becomes invalid.

#### Get Delivery History

```
GET /api/v1/webhooks/:id/deliveries?page=1&limit=20
```
Returns delivery history with pagination.

## Admin Endpoints (Admin/Moderator Only)

### Get Dead-Letter Queue Items

```
GET /api/v1/admin/webhooks/dlq?page=1&limit=20
```
Returns paginated list of failed webhook deliveries.

Response:
```json
{
  "items": [
    {
      "id": "uuid",
      "subscription_id": "uuid",
      "delivery_id": "uuid",
      "event_type": "clip.submitted",
      "event_id": "uuid",
      "payload": "{...}",
      "error_message": "HTTP 500: Internal Server Error",
      "http_status_code": 500,
      "attempt_count": 5,
      "original_created_at": "2024-01-01T12:00:00Z",
      "moved_to_dlq_at": "2024-01-01T12:10:00Z",
      "replayed_at": null,
      "replay_successful": null
    }
  ],
  "pagination": {
    "page": 1,
    "limit": 20,
    "total": 5,
    "total_pages": 1
  }
}
```

### Replay Dead-Letter Queue Item

```
POST /api/v1/admin/webhooks/dlq/:id/replay
```
Attempts to redeliver a failed webhook. The webhook will be sent with a special `X-Webhook-Replay: true` header.

Response:
```json
{
  "message": "Webhook replayed successfully"
}
```

### Delete Dead-Letter Queue Item

```
DELETE /api/v1/admin/webhooks/dlq/:id
```
Permanently deletes a DLQ item.

Response:
```json
{
  "message": "DLQ item deleted successfully"
}
```

## Security Features

### Secret Management

- Secrets are 32-byte cryptographically secure random values
- Secrets are only shown once upon creation or regeneration
- Secrets should be stored securely by the webhook consumer
- Secrets are used for HMAC-SHA256 signature verification

### SSRF Protection

- Webhook URLs are validated to prevent Server-Side Request Forgery
- Private IP ranges and localhost are blocked
- Only HTTP and HTTPS schemes are allowed

### Rate Limiting

- Create webhook: 10 requests per hour
- Regenerate secret: 5 requests per hour
- Get events: 60 requests per minute

## Webhook Payload Format

Webhooks are sent as POST requests with the following format:

```json
{
  "event": "clip.submitted",
  "timestamp": "2024-01-01T12:00:00Z",
  "data": {
    "submission_id": "uuid",
    "clip_id": "uuid",
    "user_id": "uuid",
    // ... additional event-specific data
  }
}
```

### Signature Verification

Each webhook request includes an `X-Webhook-Signature` header containing an HMAC-SHA256 signature. You must verify this signature to ensure the webhook is authentic.

**Important:** See the Webhook Signature Verification Guide for:
- Detailed explanation of the signature verification process
- Complete code examples in 7+ programming languages (Node.js, Python, Go, Ruby, PHP, Java, C#)
- Working test servers you can run locally
- Sample payloads with pre-computed signatures
- Security best practices and troubleshooting

Quick example (Node.js):

```javascript
const crypto = require('crypto');

function verifySignature(payload, signature, secret) {
  const hmac = crypto.createHmac('sha256', secret);
  hmac.update(payload);
  const expectedSignature = hmac.digest('hex');
  return crypto.timingSafeEqual(
    Buffer.from(signature),
    Buffer.from(expectedSignature)
  );
}
```

**For production-ready implementations and examples in your preferred language, see WEBHOOK_SIGNATURE_VERIFICATION.md.**

## Delivery Guarantees

### Retry Logic

- Failed deliveries are automatically retried with exponential backoff
- Maximum 5 retry attempts
- Retry intervals: 30s, 60s, 120s, 240s, 480s
- After exhausting all retries, deliveries are moved to the dead-letter queue (DLQ)

### Delivery Status

- **pending**: Delivery is scheduled or in progress
- **delivered**: Successfully delivered (HTTP 2xx response)
- **failed**: All retry attempts exhausted (moved to DLQ)

### Dead-Letter Queue (DLQ)

When a webhook delivery fails after all retry attempts, it is moved to the dead-letter queue. Administrators can:
- View all failed deliveries in the admin panel at `/admin/webhooks/dlq`
- Inspect the full payload and error details
- Manually replay failed deliveries to retry them
- Delete entries that are no longer needed

**Admin Features:**
- **View DLQ**: See all permanently failed webhook deliveries
- **Replay**: Attempt to redeliver a failed webhook
- **Delete**: Remove entries from the DLQ
- **Audit Trail**: Track replay attempts and their success/failure status

### Audit Log

The delivery history provides a complete audit log including:
- Event type and timestamp
- HTTP status code
- Response body (truncated)
- Error messages
- Retry count
- Delivery status

## Best Practices

1. **Verify Signatures**: Always verify the webhook signature before processing
2. **Handle Idempotency**: Use event IDs to prevent duplicate processing
3. **Respond Quickly**: Return 2xx status within 10 seconds
4. **Secure Your Endpoint**: Use HTTPS and validate requests
5. **Monitor Deliveries**: Check delivery history regularly
6. **Rotate Secrets**: Periodically regenerate secrets for security
7. **Handle Failures**: Implement proper error handling and logging

## Troubleshooting

### Webhook Not Receiving Events

1. Check that the subscription is active
2. Verify the URL is correct and accessible
3. Check delivery history for error messages
4. Ensure your endpoint responds within 10 seconds
5. Verify you're subscribed to the correct events
6. **If deliveries are in DLQ**: Check admin panel for failed deliveries and replay them after fixing issues

### Authentication Failures

1. Verify signature verification is implemented correctly
2. Ensure you're using the current secret
3. Check that the payload hasn't been modified

### Persistent Failures

If webhooks consistently fail and appear in the DLQ:
1. **Check endpoint availability**: Ensure your webhook endpoint is accessible from the internet
2. **Review error messages**: Check the DLQ for specific error details
3. **Verify SSL/TLS**: Ensure your endpoint has a valid SSL certificate
4. **Check response time**: Webhooks must respond within 10 seconds
5. **Fix and replay**: After resolving issues, use the admin panel to replay failed deliveries

### Rate Limiting

If you hit rate limits:
1. Reduce creation frequency
2. Implement exponential backoff
3. Cache supported events list

### Monitoring and Observability

For production deployments:
1. **Monitor DLQ size**: Keep track of failed deliveries
2. **Set up alerts**: Get notified when DLQ items exceed threshold
3. **Review regularly**: Check admin panel periodically for patterns
4. **Track replay success**: Monitor whether replayed webhooks succeed
