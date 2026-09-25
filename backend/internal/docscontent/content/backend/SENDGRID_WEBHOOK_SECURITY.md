---
title: "SENDGRID WEBHOOK SECURITY"
summary: "The SendGrid webhook handler implements ECDSA signature verification to prevent webhook spoofing attacks. This ensures that all webhook events received by the application are genuinely from SendGrid."
tags: ["docs"]
area: "docs"
status: "stable"
owner: "team-core"
version: "1.0"
last_reviewed: 2026-01-29
---

# SendGrid Webhook Signature Verification

## Overview

The SendGrid webhook handler implements ECDSA signature verification to prevent webhook spoofing attacks. This ensures that all webhook events received by the application are genuinely from SendGrid.

## Security Implementation

### How It Works

1. **Signature Headers**: SendGrid includes two headers with each webhook request:
   - `X-Twilio-Email-Event-Webhook-Signature`: Base64-encoded ECDSA signature
   - `X-Twilio-Email-Event-Webhook-Timestamp`: Unix timestamp of when the webhook was sent

2. **Verification Process**:
   - Extract signature and timestamp from headers
   - Validate timestamp is within 5 minutes (prevents replay attacks)
   - Construct signed payload: `timestamp + request_body`
   - Hash the signed payload with SHA-256
   - Verify the signature using SendGrid's ECDSA public key
   - Reject request with 401 Unauthorized if verification fails

3. **Logging**: All verification attempts (success and failure) are logged for security monitoring

## Configuration

### Obtaining the Public Key

1. Log in to [SendGrid Dashboard](https://app.sendgrid.com/)
2. Navigate to **Settings** → **Mail Settings** → **Event Webhook**
3. Enable **Signed Event Webhook** if not already enabled
4. Copy the **Public Key** (PEM format)

### Environment Configuration

Add the public key to your environment configuration:

```bash
# .env or environment variables
SENDGRID_WEBHOOK_PUBLIC_KEY='-----BEGIN PUBLIC KEY-----
MFkwEwYHKoZIzj0CAQYIKoZIzj0DAQcDQgAE...
-----END PUBLIC KEY-----'
```

**Important Notes:**
- The public key must be in PEM format (PKIX/SubjectPublicKeyInfo)
- Include the `-----BEGIN PUBLIC KEY-----` and `-----END PUBLIC KEY-----` headers
- The key uses the ECDSA P-256 curve
- For multi-line keys in environment files, use `\n` for newlines or quote the entire key

### Example Multi-line Format

```bash
SENDGRID_WEBHOOK_PUBLIC_KEY="-----BEGIN PUBLIC KEY-----\nMFkwEwYHKoZIzj0CAQYIKoZIzj0DAQcDQgAEexample...\n-----END PUBLIC KEY-----"
```

## Behavior

### With Public Key Configured

When `SENDGRID_WEBHOOK_PUBLIC_KEY` is set:
- **All** webhook requests MUST include valid signature headers
- Missing or invalid signatures are rejected with 401 Unauthorized
- Only requests with valid signatures are processed

### Without Public Key Configured

When `SENDGRID_WEBHOOK_PUBLIC_KEY` is empty or not set:
- Webhook signature verification is **disabled**
- A warning is logged on application startup
- All webhook requests are processed without verification
- **NOT RECOMMENDED for production**

## Key Rotation

If you need to rotate your SendGrid webhook signing key:

1. Generate a new key pair in SendGrid dashboard
2. Update the `SENDGRID_WEBHOOK_PUBLIC_KEY` environment variable
3. Restart the application to load the new key
4. Old signatures will be rejected immediately after restart

**Best Practice**: Use a zero-downtime deployment strategy (blue-green or rolling update) to minimize webhook delivery issues during key rotation.

## Testing

### Generate Test Signatures

The test suite includes utilities for generating valid ECDSA signatures:

```go
// From sendgrid_webhook_handler_test.go
privateKey, publicKeyPEM, err := generateTestKeyPair()
signature, err := signPayload(privateKey, timestamp, payload)
```

### Run Tests

```bash
cd backend
go test -v ./internal/handlers -run TestWebhookSignature
```

### Manual Testing with curl

```bash
# This will fail without a valid signature
curl -X POST http://localhost:8080/api/v1/webhooks/sendgrid \
  -H "Content-Type: application/json" \
  -d '[{"email":"test@example.com","timestamp":1234567890,"event":"delivered"}]'

# Response: 401 Unauthorized
```

## Security Considerations

### Protection Against

✅ **Webhook Spoofing**: Attackers cannot forge SendGrid webhook events  
✅ **Replay Attacks**: 5-minute timestamp window prevents replay of old webhooks  
✅ **MITM Attacks**: Signature verification ensures payload integrity  

### Additional Security Layers

1. **HTTPS**: Always use HTTPS endpoints for production webhooks
2. **IP Whitelisting**: Consider additional IP-based filtering for SendGrid's webhook IPs
3. **Rate Limiting**: Standard application rate limiting applies to webhook endpoints
4. **Monitoring**: Failed verification attempts are logged for security monitoring

## Troubleshooting

### Common Issues

#### "Missing signature headers"
- **Cause**: SendGrid webhook signature is not enabled
- **Solution**: Enable "Signed Event Webhook" in SendGrid dashboard

#### "Invalid signature"
- **Cause**: Wrong public key, payload tampering, or timestamp mismatch
- **Solution**: Verify public key matches SendGrid dashboard, check logs for details

#### "Timestamp too old"
- **Cause**: Webhook delivery delayed > 5 minutes
- **Solution**: Investigate network issues or webhook queue backlog in SendGrid

#### "Invalid timestamp format"
- **Cause**: Malformed timestamp header
- **Solution**: Contact SendGrid support if persists

### Debug Logging

Failed verification attempts are logged with details:

```json
{
  "level": "warn",
  "message": "Webhook signature verification failed: timestamp too old",
  "fields": {
    "timestamp_age_seconds": 360.5,
    "max_age_seconds": 300
  }
}
```

Successful verifications are also logged:

```json
{
  "level": "info",
  "message": "Webhook signature verification successful",
  "fields": {
    "timestamp": "1234567890"
  }
}
```

## References

- [SendGrid Event Webhook Security](https://docs.sendgrid.com/for-developers/tracking-events/getting-started-event-webhook-security-features)
- [Go crypto/ecdsa Package](https://pkg.go.dev/crypto/ecdsa)
- [ECDSA on Wikipedia](https://en.wikipedia.org/wiki/Elliptic_Curve_Digital_Signature_Algorithm)

## Production Checklist

Before deploying to production:

- [ ] SendGrid webhook signing enabled
- [ ] Public key configured in environment
- [ ] Tests passing (including signature verification tests)
- [ ] HTTPS endpoint configured
- [ ] Monitoring/alerting set up for failed verifications
- [ ] Security scan completed (CodeQL)
- [ ] Documentation reviewed by security team
