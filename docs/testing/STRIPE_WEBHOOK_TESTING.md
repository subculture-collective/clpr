---
title: "STRIPE WEBHOOK TESTING"
summary: "This guide explains how to test Stripe webhook handlers using the Stripe CLI, automated integration tests, and verify webhook functionality in development and production environments."
tags: ["docs","testing"]
area: "docs"
status: "stable"
owner: "team-core"
version: "1.0"
last_reviewed: 2026-01-29
---

# Stripe Webhook Testing Guide

This guide explains how to test Stripe webhook handlers using the Stripe CLI, automated integration tests, and verify webhook functionality in development and production environments.

## Quick Start: Automated Integration Tests

**New**: Run integration tests that validate subscription infrastructure, database schema, and handler wiring:

```bash
# Run all Stripe integration tests
make test-integration-stripe

# Or run from backend directory
cd backend
go test -v -tags=integration ./tests/integration/premium/ -run "TestWebhook|TestEntitlement|TestProration|TestPaymentFailure"
```

**What these automated tests cover**:
- ✅ Infrastructure wiring for webhook endpoints and handlers
- ✅ Basic entitlement and subscription status update flows
- ✅ Grace period and subscription state transitions at the database level
- ✅ Presence of retry/queue mechanisms for webhook processing
- ✅ Proration- and invoice-related database interactions
- ✅ Payment failure–related database interactions and escalation paths
- ✅ Database schema and migration validation

> **Note**: These integration tests focus on infrastructure, schema, and handler wiring. They do **not** use valid Stripe webhook signatures and therefore do **not** fully validate end-to-end Stripe webhook verification or business logic. For full webhook behavior testing, use the Stripe CLI and the flows described below.

**Environment Variables** (for future full end-to-end testing):
```bash
export TEST_STRIPE_SECRET_KEY=sk_test_your_key
export TEST_STRIPE_WEBHOOK_SECRET=whsec_test_your_secret
```

See `backend/tests/integration/premium/subscription_webhook_integration_test.go` for test implementation details.

## Prerequisites

1. **Stripe Account**: You need a Stripe account with test mode enabled
2. **Stripe CLI**: Install the Stripe CLI tool
3. **Backend Running**: The backend server must be running locally or be accessible via a public URL

## Installing Stripe CLI

### macOS
```bash
brew install stripe/stripe-cli/stripe
```

### Linux
```bash
# Download and extract
wget https://github.com/stripe/stripe-cli/releases/latest/download/stripe_linux_x86_64.tar.gz
tar -xvf stripe_linux_x86_64.tar.gz
sudo mv stripe /usr/local/bin/
```

### Windows
Download from: https://github.com/stripe/stripe-cli/releases

## Authenticating with Stripe

```bash
# Login to your Stripe account
stripe login

# This will open a browser window for authentication
# After successful login, you'll receive a confirmation
```

## Testing Webhooks Locally

### Step 1: Start the Backend Server

```bash
cd backend
go run ./cmd/api
```

Verify the server is running on `http://localhost:8080` (or your configured port).

### Step 2: Forward Webhooks to Local Server

```bash
# Forward all webhook events to local endpoint
stripe listen --forward-to localhost:8080/api/v1/webhooks/stripe

# Or forward specific events only
stripe listen --events customer.subscription.created,customer.subscription.updated,customer.subscription.deleted,invoice.payment_succeeded,invoice.payment_failed,charge.dispute.created --forward-to localhost:8080/api/v1/webhooks/stripe
```

**Important**: When you run `stripe listen`, it will output a webhook signing secret like:
```
> Ready! Your webhook signing secret is whsec_xxxxxxxxxxxxx (^C to quit)
```

Copy this secret and update your `.env` file:
```bash
STRIPE_WEBHOOK_SECRET=whsec_xxxxxxxxxxxxx
```

Then restart your backend server to use the new secret.

### Step 3: Trigger Test Events

In a new terminal window, trigger test webhook events:

#### Subscription Events

```bash
# Test subscription creation
stripe trigger customer.subscription.created

# Test subscription update
stripe trigger customer.subscription.updated

# Test subscription deletion
stripe trigger customer.subscription.deleted
```

#### Invoice Events

```bash
# Test successful payment
stripe trigger invoice.payment_succeeded

# Test failed payment
stripe trigger invoice.payment_failed

# Test invoice finalization
stripe trigger invoice.finalized
```

#### Dispute Events

```bash
# Test dispute creation
stripe trigger charge.dispute.created
```

### Step 4: Monitor Webhook Processing

Watch the terminal where `stripe listen` is running to see:
- Incoming webhook events
- HTTP status codes returned by your endpoint
- Any errors or warnings

Check your backend logs for detailed processing information:
```bash
# Backend should log webhook events like:
[WEBHOOK] Received event: evt_xxxxx (type: customer.subscription.created)
[WEBHOOK] Processing subscription.created for customer: cus_xxxxx
[WEBHOOK] Successfully processed event: evt_xxxxx
```

## Testing Specific Event Types

### Testing Subscription Creation
```bash
# Trigger a subscription creation with specific data
stripe trigger customer.subscription.created --add subscription:status=active --add subscription:items:0:price:id=price_1234567890
```

### Testing Payment Failure with Dunning
```bash
# Trigger payment failure to test dunning logic
stripe trigger invoice.payment_failed

# Check that:
# 1. Subscription status is updated
# 2. Dunning entry is created
# 3. Email notification is sent (if enabled)
```

### Testing Dispute
```bash
# Trigger dispute to test dispute handling
stripe trigger charge.dispute.created

# Verify:
# 1. Dispute is logged in audit log
# 2. Email notification is sent to user
# 3. Subscription event is recorded
```

## Verifying Webhook Functionality

### 1. Check Idempotency

Send the same event twice and verify it's only processed once:

```bash
# Get a real event ID from Stripe Dashboard or logs
stripe events resend evt_xxxxxxxxxxxxx
stripe events resend evt_xxxxxxxxxxxxx  # Send again

# Check logs - second event should be skipped:
# [WEBHOOK] Duplicate event evt_xxxxx, skipping
```

### 2. Test Signature Verification

Try sending a webhook without proper signature:

```bash
# This should fail signature verification
curl -X POST http://localhost:8080/api/v1/webhooks/stripe \
  -H "Content-Type: application/json" \
  -d '{"type": "customer.subscription.created", "data": {}}'

# Expected response: 400 Bad Request - "Missing signature"
```

### 3. Verify All Event Types

Create a test script to verify all supported event types:

```bash
#!/bin/bash
# test-all-webhooks.sh

echo "Testing all Stripe webhook event types..."

events=(
  "customer.subscription.created"
  "customer.subscription.updated"
  "customer.subscription.deleted"
  "invoice.payment_succeeded"
  "invoice.payment_failed"
  "charge.dispute.created"
)

for event in "${events[@]}"; do
  echo "Testing $event..."
  stripe trigger "$event"
  sleep 2  # Wait between events
done

echo "All webhook tests completed!"
```

Make it executable and run:
```bash
chmod +x test-all-webhooks.sh
./test-all-webhooks.sh
```

## Testing Retry Mechanism

### Simulate Webhook Failure

1. **Temporarily break the database connection** to cause webhook processing to fail
2. Send a webhook event
3. Verify the event is added to retry queue:
   ```bash
   # Check webhook retry queue in your database
   psql -d clpr_db -c "SELECT * FROM webhook_retry_queue ORDER BY created_at DESC LIMIT 10;"
   ```
4. Restore database connection
5. Wait for retry scheduler to process the queued event (runs every 1 minute by default)

## Production Testing

### Setting Up Production Webhooks

1. **Create Webhook Endpoint in Stripe Dashboard**:
   - Go to: https://dashboard.stripe.com/webhooks
   - Click "Add endpoint"
   - URL: `https://your-domain.com/api/v1/webhooks/stripe`
   - Select events to listen to (or select "all events")
   - Click "Add endpoint"

2. **Get Webhook Signing Secret**:
   - After creating the endpoint, click "Reveal" next to "Signing secret"
   - Copy the secret (starts with `whsec_`)

3. **Configure Production Environment**:
   ```bash
   # In your production .env or via Vault
   STRIPE_WEBHOOK_SECRET=whsec_your_production_secret
   ```

### Testing Production Webhooks

1. **Use Stripe Dashboard**:
   - Go to Developers → Webhooks
   - Click on your endpoint
   - Click "Send test webhook"
   - Select event type and send

2. **Monitor Production Events**:
   - Check webhook attempt logs in Stripe Dashboard
   - View response codes and timing
   - Check for failed attempts

3. **Verify Webhook Signature in Production**:
   - Ensure production webhook secret is configured
   - Test with a real subscription flow
   - Check application logs for successful processing

## Monitoring and Debugging

### Check Webhook Logs

```bash
# View recent webhook processing logs
tail -f /var/log/clpr/backend.log | grep WEBHOOK

# Or use your logging system (e.g., CloudWatch, Datadog)
```

### Common Issues and Solutions

#### 1. Signature Verification Fails
**Problem**: `webhook signature verification failed`
**Solution**:
- Verify `STRIPE_WEBHOOK_SECRET` matches the secret from Stripe CLI or Dashboard
- Ensure you're using the correct secret for test vs live mode
- Check that the secret is correctly loaded in your application

#### 2. Duplicate Event Processing
**Problem**: Same event processed multiple times
**Solution**:
- Verify idempotency check is working
- Check `GetEventByStripeEventID` is querying correctly
- Ensure database transaction is properly committed

#### 3. Timeout or Slow Processing
**Problem**: Webhooks timing out or processing slowly
**Solution**:
- Optimize database queries
- Use async processing for non-critical tasks (emails, etc.)
- Ensure webhook handler returns quickly (<5 seconds)

#### 4. Missing Events
**Problem**: Some events not being processed
**Solution**:
- Check event type is in the switch statement
- Verify endpoint is listening for the event type in Stripe Dashboard
- Check application logs for any errors during processing

## Best Practices

1. **Always Use Signature Verification**: Never skip signature verification in production
2. **Implement Idempotency**: Always check for duplicate events using event ID
3. **Return Quickly**: Acknowledge webhooks within 5 seconds, defer heavy processing
4. **Log Everything**: Log all webhook events, successes, and failures
5. **Monitor Retry Queue**: Set up alerts for growing retry queue
6. **Test Thoroughly**: Test all event types before deploying to production
7. **Handle Gracefully**: Don't crash on unexpected event types or data
8. **Use Multiple Secrets**: Support webhook secret rotation by configuring multiple secrets

## Webhook Monitoring Dashboard

Monitor webhook health using these endpoints:

```bash
# Get webhook retry statistics
curl http://localhost:8080/health/webhooks

# Response includes:
# - Total events processed
# - Failed events
# - Events in retry queue
# - Average processing time
```

## Further Reading

- [Stripe Webhook Documentation](https://stripe.com/docs/webhooks)
- [Stripe CLI Documentation](https://stripe.com/docs/stripe-cli)
- [Testing Webhooks](https://stripe.com/docs/webhooks/test)
- [Webhook Best Practices](https://stripe.com/docs/webhooks/best-practices)
