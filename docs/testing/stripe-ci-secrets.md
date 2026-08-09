---
title: "Stripe Ci Secrets"
summary: "This document describes how to configure Stripe test keys for E2E testing in CI/CD pipelines."
tags: ["testing"]
area: "testing"
status: "stable"
owner: "team-core"
version: "1.0"
last_reviewed: 2026-01-29
---

# Stripe Test Keys Configuration for CI/CD

This document describes how to configure Stripe test keys for E2E testing in CI/CD pipelines.

## Required GitHub Secrets

The following secrets must be configured in GitHub repository settings for premium subscription E2E tests to run with Stripe integration:

### Secrets to Configure

Go to: `Settings → Secrets and variables → Actions → New repository secret`

#### 1. STRIPE_TEST_PUBLISHABLE_KEY
- **Description**: Stripe publishable key for test mode (frontend)
- **Format**: Starts with `pk_test_`
- **Where to find**: Stripe Dashboard → Developers → API keys → Test mode
- **Example**: `pk_test_51A1B2C3D4E5F6G7H8I9J0K1L2M3N4O5P6Q7R8S9T0U1V2W3X4Y5Z6`

#### 2. STRIPE_TEST_SECRET_KEY
- **Description**: Stripe secret key for test mode (backend)
- **Format**: Starts with `sk_test_`
- **Where to find**: Stripe Dashboard → Developers → API keys → Test mode
- **Example**: `sk_test_<test-key-from-dashboard>`
- **⚠️ IMPORTANT**: This is sensitive. Never commit or expose.

#### 3. STRIPE_TEST_WEBHOOK_SECRET
- **Description**: Webhook signing secret for test mode (backend)
- **Format**: Starts with `whsec_`
- **Where to find**: Stripe Dashboard → Developers → Webhooks → Add endpoint → Get signing secret
- **Example**: `whsec_1A2B3C4D5E6F7G8H9I0J1K2L3M4N5O6P7Q8R9S0T1U2V3W4X5Y6Z`
- **Note**: Create a webhook endpoint for `https://test.example.com/api/v1/webhooks/stripe` or use Stripe CLI

#### 4. STRIPE_TEST_MONTHLY_PRICE_ID
- **Description**: Price ID for monthly Pro subscription in test mode
- **Format**: Starts with `price_`
- **Where to find**: Stripe Dashboard → Products → Pro subscription → Pricing → Copy price ID
- **Example**: `price_1A2B3C4D5E6F7G8H9I0J1K2L`

#### 5. STRIPE_TEST_YEARLY_PRICE_ID
- **Description**: Price ID for yearly Pro subscription in test mode
- **Format**: Starts with `price_`
- **Where to find**: Stripe Dashboard → Products → Pro subscription → Pricing → Copy price ID
- **Example**: `price_1K2L3M4N5O6P7Q8R9S0T1U2V`

## Setting Up Stripe Test Environment

### 1. Create Test Mode Products

1. Go to Stripe Dashboard
2. Switch to **Test mode** (toggle in top right)
3. Navigate to Products
4. Create a product called "Pro Subscription" or similar
5. Add two prices:
   - Monthly: $9.99/month (or your pricing)
   - Yearly: $99.99/year (or your pricing)
6. Copy the price IDs for both

### 2. Configure Webhook Endpoint (Optional)

If you want to test webhooks locally or in staging:

1. Go to Developers → Webhooks
2. Click "Add endpoint"
3. URL: Your backend webhook endpoint (e.g., `https://staging.yourdomain.com/api/v1/webhooks/stripe`)
4. Select events to listen for:
   - `customer.subscription.created`
   - `customer.subscription.updated`
   - `customer.subscription.deleted`
   - `invoice.payment_succeeded`
   - `invoice.payment_failed`
5. Copy the signing secret

### 3. Add Secrets to GitHub

```bash
# Using GitHub CLI
gh secret set STRIPE_TEST_PUBLISHABLE_KEY -b "pk_test_..."
gh secret set STRIPE_TEST_SECRET_KEY -b "sk_test_..."
gh secret set STRIPE_TEST_WEBHOOK_SECRET -b "whsec_..."
gh secret set STRIPE_TEST_MONTHLY_PRICE_ID -b "price_..."
gh secret set STRIPE_TEST_YEARLY_PRICE_ID -b "price_..."
```

Or use the GitHub web interface:
1. Go to repository Settings
2. Click "Secrets and variables" → "Actions"
3. Click "New repository secret"
4. Add each secret with its name and value

## Testing Without Stripe Keys

The E2E tests are designed to gracefully handle missing Stripe keys:

- **Without keys**: Tests verify UI/UX only. Tests requiring actual Stripe integration are skipped.
- **With keys**: Full integration tests run, including actual Stripe Checkout flows.

This allows tests to run in environments where Stripe integration isn't fully configured.

## Security Best Practices

### ✅ DO:
- ✅ Always use test mode keys (starting with `pk_test_`, `sk_test_`)
- ✅ Store keys in GitHub Secrets, never in code
- ✅ Rotate keys periodically
- ✅ Use different keys for staging and production
- ✅ Limit access to secrets to necessary team members
- ✅ Monitor Stripe Dashboard for unusual test mode activity

### ❌ DON'T:
- ❌ NEVER commit keys to version control
- ❌ NEVER use production keys (`pk_live_`, `sk_live_`) in CI/CD
- ❌ NEVER expose secret keys in logs or error messages
- ❌ NEVER share keys via insecure channels (email, Slack, etc.)
- ❌ NEVER use the same keys across multiple environments

## Verifying Configuration

### Check Secrets Are Set

```bash
# Using GitHub CLI
gh secret list
```

You should see:
```
STRIPE_TEST_PUBLISHABLE_KEY
STRIPE_TEST_SECRET_KEY
STRIPE_TEST_WEBHOOK_SECRET
STRIPE_TEST_MONTHLY_PRICE_ID
STRIPE_TEST_YEARLY_PRICE_ID
```

### Test Locally

Before pushing, test with local Stripe keys:

```bash
cd frontend

# Set environment variables
export VITE_STRIPE_PUBLISHABLE_KEY="pk_test_..."
export VITE_STRIPE_PRO_MONTHLY_PRICE_ID="price_..."
export VITE_STRIPE_PRO_YEARLY_PRICE_ID="price_..."

# Run E2E tests
npm run test:e2e -- premium-subscription
```

## Troubleshooting

### Tests Skip Stripe Flows

**Problem**: Subscription tests skip checkout scenarios

**Solution**: Verify Stripe keys are configured in GitHub Secrets and accessible in the workflow

### Invalid Stripe Keys Error

**Problem**: Tests fail with "Invalid API key" or "No such price"

**Solution**: 
1. Verify keys are from test mode (start with `pk_test_`, `sk_test_`, `price_`)
2. Check price IDs exist in your Stripe Dashboard
3. Ensure secrets are named exactly as shown above

### Webhook Signature Verification Fails

**Problem**: Backend rejects webhooks with signature errors

**Solution**:
1. Verify `STRIPE_TEST_WEBHOOK_SECRET` is set correctly
2. For local testing, use Stripe CLI: `stripe listen --forward-to localhost:8080/api/v1/webhooks/stripe`
3. Check webhook endpoint URL matches configuration

### Tests Timeout

**Problem**: E2E tests timeout waiting for Stripe Checkout

**Solution**:
1. Increase timeout in test configuration
2. Check backend is running and accessible
3. Verify Stripe API is reachable from CI environment

## CI Workflow Integration

The Stripe secrets are used in `.github/workflows/ci.yml`:

```yaml
- name: Run E2E tests
  env:
    VITE_STRIPE_PUBLISHABLE_KEY: ${{ secrets.STRIPE_TEST_PUBLISHABLE_KEY }}
    VITE_STRIPE_PRO_MONTHLY_PRICE_ID: ${{ secrets.STRIPE_TEST_MONTHLY_PRICE_ID }}
    VITE_STRIPE_PRO_YEARLY_PRICE_ID: ${{ secrets.STRIPE_TEST_YEARLY_PRICE_ID }}
  run: |
    cd frontend
    npm run test:e2e
```

Backend also needs keys for webhook handling:

```yaml
- name: Start backend server
  env:
    STRIPE_SECRET_KEY: ${{ secrets.STRIPE_TEST_SECRET_KEY }}
    STRIPE_WEBHOOK_SECRET: ${{ secrets.STRIPE_TEST_WEBHOOK_SECRET }}
    # ... other env vars
```

## Maintenance

### Key Rotation

Rotate Stripe test keys periodically:

1. Generate new keys in Stripe Dashboard
2. Update GitHub Secrets
3. Trigger a test CI run to verify
4. Revoke old keys in Stripe Dashboard

### Monitoring

Monitor your Stripe test mode activity:

1. Check Stripe Dashboard → Developers → Logs
2. Review API request volume
3. Look for unusual patterns or errors
4. Set up alerts for failed requests

## Related Documentation

- Premium Subscription E2E Tests README
- [Stripe Webhook Testing Guide](./STRIPE_WEBHOOK_TESTING.md)
- Stripe Integration Documentation
- [GitHub Secrets Documentation](https://docs.github.com/en/actions/security-guides/encrypted-secrets)
- [Stripe Test Mode Documentation](https://stripe.com/docs/test-mode)

---

**Last Updated**: December 25, 2025  
**Maintained By**: Clipper Engineering Team
