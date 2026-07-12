---
title: "ENABLING PREMIUM SUBSCRIPTION TESTS"
summary: "The premium subscription checkout E2E tests are fully implemented but **4 out of 13 tests are currently skipped** because they require Stripe test mode configuration. This document explains how to ena"
tags: ["testing"]
area: "testing"
status: "stable"
owner: "team-core"
version: "1.0"
last_reviewed: 2026-01-29
---

# Enabling Premium Subscription E2E Tests

## Overview

The premium subscription checkout E2E tests are fully implemented but **4 out of 13 tests are currently skipped** because they require Stripe test mode configuration. This document explains how to enable these tests.

## Current Test Status

### ✅ Active Tests (9 tests - No Stripe Required)

These tests run without Stripe configuration:

1. **Pricing Page Display** - Verifies page loads with pricing tiers
2. **Billing Toggle** - Tests monthly/yearly switching
3. **Login Redirect** - Ensures unauthenticated users go to login
4. **Checkout Cancellation** - Tests user can cancel checkout
5. **Navigation** - Verifies navigation between pages
6. **Success Page UI** - Tests success page renders correctly
7. **Upgrade Prompts** - Verifies free users see upgrade CTAs
8. **Pricing Links** - Confirms pricing links are visible

### ⏭️ Skipped Tests (4 tests - Require Stripe)

These tests are conditionally skipped when `VITE_STRIPE_PUBLISHABLE_KEY` is missing or doesn't start with `pk_test_`:

1. **Successful Checkout** (`line 83`) - Tests complete purchase flow
2. **Declined Card** (`line 118`) - Tests error handling for declined payments
3. **Insufficient Funds** (`line 148`) - Tests error handling for insufficient funds
4. **Pro Feature Access** (`line 246`) - Tests entitlements after purchase

## Prerequisites

### 1. Install Dependencies

```bash
cd frontend
npm install
npx playwright install  # Install browser binaries
```

### 2. Set Up Stripe Test Mode

#### A. Create/Access Stripe Account

1. Go to https://dashboard.stripe.com
2. Toggle to **Test Mode** (top-right corner)
3. Navigate to **Developers** > **API keys**

#### B. Get API Keys

Copy these values:

- **Publishable key**: Starts with `pk_test_`
- **Secret key**: Starts with `sk_test_`

#### C. Create Products and Prices

1. Go to **Products** > **Add product**
2. Create product: **clpr Pro**
3. Add two prices:
   - **Monthly**: `$9.99/month` (recurring)
   - **Yearly**: `$99.99/year` (recurring)
4. Copy the Price IDs (format: `price_xxxxx`)

#### D. Set Up Webhook (Optional for E2E)

For full integration testing:

1. Go to **Developers** > **Webhooks**
2. Add endpoint: `https://your-domain.com/api/v1/webhooks/stripe`
3. Select events:
   - `customer.subscription.created`
   - `customer.subscription.updated`
   - `customer.subscription.deleted`
   - `invoice.paid`
   - `invoice.payment_failed`
   - `payment_intent.succeeded`
4. Copy the **Webhook signing secret** (format: `whsec_xxxxx`)

### 3. Configure Environment Variables

#### Frontend Configuration

Edit `frontend/.env`:

```bash
# Stripe Test Mode Keys
VITE_STRIPE_PUBLISHABLE_KEY=pk_test_51XXXXXXXXXXXXX
VITE_STRIPE_PRO_MONTHLY_PRICE_ID=price_XXXXXXXXXXXXX
VITE_STRIPE_PRO_YEARLY_PRICE_ID=price_YYYYYYYYYYYYY
```

#### Backend Configuration (Optional)

Edit `backend/.env`:

```bash
# Stripe Test Mode Keys
STRIPE_SECRET_KEY=sk_test_51XXXXXXXXXXXXX
STRIPE_WEBHOOK_SECRET=whsec_XXXXXXXXXXXXX
STRIPE_PRO_MONTHLY_PRICE_ID=price_XXXXXXXXXXXXX
STRIPE_PRO_YEARLY_PRICE_ID=price_YYYYYYYYYYYYY

# Redirect URLs
STRIPE_SUCCESS_URL=http://localhost:5173/subscription/success
STRIPE_CANCEL_URL=http://localhost:5173/pricing
```

## Running the Tests

### Run All Tests

```bash
cd frontend
npm run test:e2e -- e2e/tests/premium-subscription-checkout.spec.ts
```

This will run all 13 tests (39 total with 3 browsers).

### Run Only Active Tests (No Stripe)

```bash
npm run test:e2e -- e2e/tests/premium-subscription-checkout.spec.ts \
  --grep-invert "test card|declined card|insufficient funds|pro features after"
```

### Run Only Stripe Tests

```bash
npm run test:e2e -- e2e/tests/premium-subscription-checkout.spec.ts \
  --grep "test card|declined card|insufficient funds"
```

### Run with UI (Debug Mode)

```bash
npm run test:e2e:ui -- e2e/tests/premium-subscription-checkout.spec.ts
```

### Run Single Browser

```bash
# Chromium only
npm run test:e2e -- e2e/tests/premium-subscription-checkout.spec.ts --project=chromium

# Firefox only
npm run test:e2e -- e2e/tests/premium-subscription-checkout.spec.ts --project=firefox

# WebKit only
npm run test:e2e -- e2e/tests/premium-subscription-checkout.spec.ts --project=webkit
```

## Test Cards

The following Stripe test cards are pre-configured in `stripe-helpers.ts`:

| Card Number | Description | Expected Result |
|------------|-------------|-----------------|
| `4242 4242 4242 4242` | Success | Payment succeeds |
| `4000 0000 0000 0002` | Declined | Generic decline |
| `4000 0000 0000 9995` | Insufficient Funds | Insufficient funds error |
| `4000 0025 0000 3155` | 3D Secure | Requires authentication |

**Expiry**: Any future date (e.g., `12/34`)  
**CVC**: Any 3 digits (e.g., `123`)  
**ZIP**: Any 5 digits (e.g., `12345`)

## Troubleshooting

### Tests Still Skipped

**Problem**: Tests show as skipped even with Stripe keys configured.

**Solution**: Verify your publishable key starts with `pk_test_`. Check the test output:

```
test.skip() condition: !stripeKey || !stripeKey.startsWith('pk_test_')
```

### Checkout Redirect Fails

**Problem**: Test fails at `await page.waitForURL(/checkout\.stripe\.com/)`

**Solution**: 
1. Check Price IDs are correct
2. Verify backend is running and accessible
3. Check browser console for API errors

### Webhook Tests Fail

**Problem**: Subscription activation tests fail

**Solution**:
1. Ensure backend is running
2. Configure webhook secret in backend `.env`
3. Check webhook endpoint is accessible

### Browser Not Installed

**Problem**: `Error: Executable doesn't exist at /home/.../.cache/ms-playwright/...`

**Solution**:
```bash
npx playwright install
```

## CI/CD Configuration

### GitHub Actions

Add Stripe test keys as repository secrets:

```yaml
env:
  VITE_STRIPE_PUBLISHABLE_KEY: ${{ secrets.STRIPE_TEST_PUBLISHABLE_KEY }}
  VITE_STRIPE_PRO_MONTHLY_PRICE_ID: ${{ secrets.STRIPE_TEST_MONTHLY_PRICE_ID }}
  VITE_STRIPE_PRO_YEARLY_PRICE_ID: ${{ secrets.STRIPE_TEST_YEARLY_PRICE_ID }}
```

### Skip Tests in CI Without Stripe

Tests automatically skip when keys aren't available:

```typescript
test('should complete successful checkout with test card', async ({ authenticatedPage }) => {
  // Skip this test if Stripe is not configured
  const stripeKey = process.env.VITE_STRIPE_PUBLISHABLE_KEY;
  if (!stripeKey || !stripeKey.startsWith('pk_test_')) {
    test.skip();  // Gracefully skip if not configured
  }
  // ... test implementation
});
```

## Next Steps

1. ✅ Configure Stripe test mode keys
2. ✅ Run non-Stripe tests to verify baseline
3. ✅ Run Stripe tests to verify checkout flow
4. ⏭️ Configure CI/CD with Stripe secrets (optional)
5. ⏭️ Set up staging environment webhook endpoint (for integration testing)

## Related Documentation

- [Stripe Subscription Testing Guide](./stripe-subscription-testing.md) - Comprehensive testing procedures
- [Stripe CI Secrets](./stripe-ci-secrets.md) - CI/CD configuration
- Premium Subscription Checkout Status - Implementation status

## Support

If you encounter issues:

1. Check the [Stripe Documentation](https://stripe.com/docs/testing)
2. Review test output and error messages
3. Verify environment variables are loaded correctly
4. Check browser console for API errors
5. Ensure backend is running and accessible
