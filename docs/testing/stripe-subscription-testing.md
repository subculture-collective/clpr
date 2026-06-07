---
title: "Stripe Subscription Testing"
summary: "This guide provides comprehensive testing procedures for Stripe subscription flows, including automated integration tests, manual test cases, edge cases, and dashboard reconciliation procedures."
tags: ["testing"]
area: "testing"
status: "stable"
owner: "team-core"
version: "1.0"
last_reviewed: 2026-01-29
---

# Stripe Subscription Lifecycle Testing Guide

This guide provides comprehensive testing procedures for Stripe subscription flows, including automated integration tests, manual test cases, edge cases, and dashboard reconciliation procedures.

## Automated Integration Tests

**New**: Integration tests are available to validate subscription infrastructure, database schema, and handler wiring:

```bash
# Run all Stripe integration tests
make test-integration-stripe

# Run specific test categories
cd backend
go test -v -tags=integration ./tests/integration/premium/ -run TestWebhookIdempotency
go test -v -tags=integration ./tests/integration/premium/ -run TestEntitlementUpdates
go test -v -tags=integration ./tests/integration/premium/ -run TestProration
go test -v -tags=integration ./tests/integration/premium/ -run TestPaymentFailure
```

**Automated Test Coverage (Current Scope)**:
1. **Webhook Idempotency**: Verifies basic webhook handler wiring and database state (for example, schema/migration expectations and that services are constructed and invoked without errors for sample events).
2. **Entitlement Updates**: Verifies that subscription/entitlement handlers can be invoked and that core persistence logic executes against the expected schema.
3. **Payment Failures**: Verifies that payment‑failure handlers are wired correctly and can persist basic failure state to the database.
4. **Proration**: Verifies that proration‑related handlers and invoice processing paths execute successfully against the current schema.
5. **Retry Logic (Infrastructure Only)**: Verifies that retry‑related services are instantiated and callable; it does **not** fully simulate or assert on a real webhook retry queue or exponential backoff behavior.

These automated tests do **not** by themselves validate full Stripe lifecycle behavior such as real duplicate event detection, production‑grade user permissions syncing, Stripe's dunning process, or actual webhook retry timing/exponential backoff. Validating those behaviors requires end‑to‑end flows driven by valid Stripe webhooks and/or the manual procedures described later in this guide. See the [Integration Test Implementation](#integration-test-details) section for more details about what is and is not covered.

## Table of Contents

1. [Test Prerequisites](#test-prerequisites)
2. [New Subscription Creation](#new-subscription-creation)
3. [Subscription Cancellation](#subscription-cancellation)
4. [Payment Method Updates](#payment-method-updates)
5. [Payment Failure Handling](#payment-failure-handling)
6. [Proration Calculations](#proration-calculations)
7. [Subscription Reactivation](#subscription-reactivation)
8. [Dispute/Chargeback Handling](#disputechargeback-handling)
9. [Stripe Dashboard Reconciliation](#stripe-dashboard-reconciliation)
10. [Edge Cases and Known Issues](#edge-cases-and-known-issues)

## Test Prerequisites

### Stripe Test Environment Setup

1. **Stripe Test Mode**: Ensure you're using Stripe test keys
   ```
   STRIPE_SECRET_KEY=sk_test_...
   STRIPE_PUBLISHABLE_KEY=pk_test_...
   STRIPE_WEBHOOK_SECRET=whsec_test_...
   ```

2. **Test Cards**: Use Stripe test card numbers
   - **Success**: `4242 4242 4242 4242`
   - **Decline**: `4000 0000 0000 0002`
   - **Requires Authentication (3D Secure)**: `4000 0025 0000 3155`
   - **Insufficient Funds**: `4000 0000 0000 9995`

3. **Webhook Configuration**: Configure webhook endpoint in Stripe Dashboard
   ```
   https://your-domain.com/api/v1/webhooks/stripe
   ```

4. **Required Webhook Events**:
   - `customer.subscription.created`
   - `customer.subscription.updated`
   - `customer.subscription.deleted`
   - `invoice.paid`
   - `invoice.payment_failed`
   - `invoice.finalized`
   - `payment_intent.succeeded`
   - `payment_intent.payment_failed`
   - `charge.dispute.created`

## New Subscription Creation

### Test Case 1.1: Create Monthly Subscription

**Steps**:
1. Log in as a test user
2. Navigate to pricing page
3. Click "Subscribe Monthly" button
4. Complete Stripe Checkout with test card `4242 4242 4242 4242`
5. Verify redirect to success page

**Expected Results**:
- ✅ Checkout session created successfully
- ✅ Customer record created in Stripe
- ✅ Subscription status is "active"
- ✅ User tier updated to "pro"
- ✅ Webhook `customer.subscription.created` received
- ✅ Subscription record created in database
- ✅ Current period dates set correctly
- ✅ Invoice generated and paid

**Verification Queries**:
```sql
-- Check subscription record
SELECT * FROM subscriptions WHERE user_id = 'USER_UUID';

-- Check subscription events
SELECT * FROM subscription_events WHERE subscription_id = 'SUBSCRIPTION_UUID' ORDER BY created_at DESC;

-- Check audit logs
SELECT * FROM audit_logs WHERE user_id = 'USER_UUID' AND action = 'subscription_created';
```

### Test Case 1.2: Create Yearly Subscription

**Steps**:
1. Log in as a test user
2. Navigate to pricing page
3. Click "Subscribe Yearly" button
4. Complete Stripe Checkout with test card `4242 4242 4242 4242`
5. Verify redirect to success page

**Expected Results**:
- ✅ Checkout session created successfully
- ✅ Annual pricing applied correctly
- ✅ Subscription status is "active"
- ✅ Current period end is ~365 days from now
- ✅ Correct price ID stored in database

### Test Case 1.3: Create Subscription with Coupon Code

**Steps**:
1. Create a test coupon in Stripe Dashboard (e.g., "TESTDISCOUNT", 25% off)
2. Log in as a test user
3. Navigate to pricing page
4. Enter coupon code during checkout
5. Complete payment

**Expected Results**:
- ✅ Discount applied to checkout total
- ✅ Subscription created with discount
- ✅ Coupon code logged in audit trail
- ✅ Invoice shows discount line item

### Test Case 1.4: Failed Subscription Creation

**Steps**:
1. Attempt checkout with declined card `4000 0000 0000 0002`

**Expected Results**:
- ✅ Checkout shows error message
- ✅ No subscription created
- ✅ User remains on free tier
- ✅ Failed payment logged in audit trail

## Subscription Cancellation

### Test Case 2.1: Cancel Immediately

**Steps**:
1. Log in as user with active subscription
2. Navigate to settings/subscription page
3. Click "Cancel Subscription"
4. Choose "Cancel Immediately" option
5. Confirm cancellation

**Expected Results**:
- ✅ Subscription canceled in Stripe
- ✅ Webhook `customer.subscription.deleted` received
- ✅ Subscription status changed to "canceled"
- ✅ User tier downgraded to "free"
- ✅ Access to pro features removed immediately
- ✅ Refund processed (if within refund window)

**Verification**:
```sql
SELECT status, tier, canceled_at, cancel_at_period_end 
FROM subscriptions 
WHERE user_id = 'USER_UUID';
```

### Test Case 2.2: Cancel at Period End

**Steps**:
1. Log in as user with active subscription
2. Navigate to Stripe Customer Portal
3. Click "Cancel Subscription"
4. Choose "Cancel at period end" option
5. Confirm cancellation

**Expected Results**:
- ✅ Subscription still active in Stripe
- ✅ Webhook `customer.subscription.updated` received
- ✅ `cancel_at_period_end` set to `true`
- ✅ Current period end date unchanged
- ✅ User retains pro access until period end
- ✅ Cancellation date recorded

**Verification**:
```sql
SELECT status, cancel_at_period_end, current_period_end, canceled_at 
FROM subscriptions 
WHERE user_id = 'USER_UUID';
```

### Test Case 2.3: Portal Session Creation

**Steps**:
1. API call: `POST /api/v1/subscriptions/portal`
2. Verify portal URL returned
3. Access portal URL in browser

**Expected Results**:
- ✅ Portal session created successfully
- ✅ Portal URL valid and accessible
- ✅ Portal shows subscription details
- ✅ Cancellation options available

## Payment Method Updates

### Test Case 3.1: Update Payment Method via Portal

**Steps**:
1. Access Stripe Customer Portal
2. Click "Update payment method"
3. Enter new test card details
4. Save changes

**Expected Results**:
- ✅ Payment method updated in Stripe
- ✅ New card becomes default payment method
- ✅ Next invoice uses new payment method
- ✅ Update logged in audit trail

### Test Case 3.2: Add Additional Payment Method

**Steps**:
1. Access Stripe Customer Portal
2. Add second payment method
3. Set as default

**Expected Results**:
- ✅ Multiple payment methods stored
- ✅ Default method updated
- ✅ Old method retained as backup

## Payment Failure Handling

### Test Case 4.1: First Payment Failure

**Steps**:
1. Update payment method to card `4000 0000 0000 9995` (insufficient funds)
2. Wait for next billing cycle or trigger renewal manually
3. Verify dunning process starts

**Expected Results**:
- ✅ Webhook `invoice.payment_failed` received
- ✅ Subscription status changed to "past_due"
- ✅ Dunning record created in database
- ✅ Email notification sent to user
- ✅ Grace period initiated (3 days default)
- ✅ User retains pro access during grace period
- ✅ Retry attempt scheduled

**Verification**:
```sql
-- Check dunning status
SELECT * FROM dunning_attempts WHERE subscription_id = 'SUBSCRIPTION_UUID';

-- Check grace period
SELECT status, grace_period_end FROM subscriptions WHERE user_id = 'USER_UUID';
```

### Test Case 4.2: Multiple Payment Failures

**Steps**:
1. Let payment retry attempts fail (usually 3-4 attempts)
2. Verify escalating notifications

**Expected Results**:
- ✅ Multiple dunning attempts recorded
- ✅ Escalating email notifications sent
- ✅ Grace period maintained between retries
- ✅ After final failure, subscription canceled
- ✅ User downgraded to free tier

### Test Case 4.3: Payment Success After Failure

**Steps**:
1. Start with failed payment (past_due status)
2. Update to valid payment method `4242 4242 4242 4242`
3. Trigger retry or wait for automatic retry

**Expected Results**:
- ✅ Webhook `invoice.paid` received
- ✅ Subscription status changed to "active"
- ✅ Dunning record cleared
- ✅ Grace period removed
- ✅ Success email sent to user
- ✅ Pro access continues uninterrupted

### Test Case 4.4: Payment Intent Failures

**Steps**:
1. Use test card requiring authentication `4000 0025 0000 3155`
2. Abandon authentication flow

**Expected Results**:
- ✅ Webhook `payment_intent.payment_failed` received
- ✅ Failure reason logged (authentication_required)
- ✅ User notified to complete authentication
- ✅ Logged in audit trail

## Proration Calculations

### Test Case 5.1: Upgrade from Monthly to Yearly

**Steps**:
1. User has active monthly subscription ($9.99/month)
2. API call: `POST /api/v1/subscriptions/change-plan` with yearly price ID
3. Verify proration invoice created

**Expected Results**:
- ✅ Immediate proration invoice generated
- ✅ Credit issued for unused monthly time
- ✅ Charge applied for yearly subscription
- ✅ Net amount correct (yearly - prorated monthly)
- ✅ Subscription updated to yearly billing
- ✅ Current period end updated to +365 days
- ✅ Webhook `customer.subscription.updated` received
- ✅ Invoice with billing_reason "subscription_update"

**Proration Formula**:
```
Credit = (Monthly Price * Unused Days) / Days in Period
Charge = Yearly Price
Net Amount = Charge - Credit
```

**Verification**:
```sql
-- Check subscription events
SELECT event_type, data FROM subscription_events 
WHERE subscription_id = 'SUBSCRIPTION_UUID' 
AND event_type = 'subscription_updated' 
ORDER BY created_at DESC LIMIT 1;
```

### Test Case 5.2: Downgrade from Yearly to Monthly

**Steps**:
1. User has active yearly subscription ($99.99/year)
2. API call: `POST /api/v1/subscriptions/change-plan` with monthly price ID
3. Verify proration handled correctly

**Expected Results**:
- ✅ Subscription updated to monthly billing
- ✅ Proration behavior set to "always_invoice"
- ✅ Proration invoice/credit note generated
- ✅ Next billing date recalculated
- ✅ Remaining time credited as applicable

### Test Case 5.3: Invalid Plan Change

**Steps**:
1. Attempt to change to current plan
2. Attempt to change to invalid price ID

**Expected Results**:
- ✅ API returns 400 Bad Request for same plan
- ✅ API returns 400 Bad Request for invalid price ID
- ✅ Error message is user-friendly
- ✅ Subscription remains unchanged

## Subscription Reactivation

### Test Case 6.1: Reactivate Scheduled Cancellation

**Steps**:
1. User has subscription set to cancel at period end
2. Access Stripe Customer Portal
3. Click "Reactivate Subscription"
4. Confirm reactivation

**Expected Results**:
- ✅ `cancel_at_period_end` changed to `false`
- ✅ Webhook `customer.subscription.updated` received
- ✅ Subscription continues beyond current period
- ✅ User notified of reactivation
- ✅ Next invoice scheduled normally

**Verification**:
```sql
SELECT cancel_at_period_end, current_period_end, canceled_at 
FROM subscriptions 
WHERE user_id = 'USER_UUID';
```

### Test Case 6.2: Create New Subscription After Cancellation

**Steps**:
1. User's subscription fully canceled (status: canceled)
2. Navigate to pricing page
3. Subscribe again with same or different plan

**Expected Results**:
- ✅ New checkout session created
- ✅ New subscription created in Stripe
- ✅ Database records updated with new subscription ID
- ✅ Old subscription remains canceled
- ✅ User tier upgraded to pro
- ✅ New billing cycle starts

## Dispute/Chargeback Handling

### Test Case 7.1: Dispute Created

**Steps**:
1. Simulate dispute in Stripe Dashboard (Test mode)
2. Select reason (fraudulent, unrecognized, etc.)
3. Verify webhook handling

**Expected Results**:
- ✅ Webhook `charge.dispute.created` received
- ✅ Dispute logged in database
- ✅ Email notification sent to user
- ✅ Admin notification sent
- ✅ Subscription remains active pending resolution
- ✅ Audit log entry created

**Verification**:
```sql
-- Check subscription events for disputes
SELECT * FROM subscription_events 
WHERE event_type = 'dispute_created' 
ORDER BY created_at DESC;

-- Check audit logs
SELECT * FROM audit_logs 
WHERE action = 'dispute_created' 
ORDER BY created_at DESC;
```

### Test Case 7.2: Dispute Won

**Steps**:
1. Resolve dispute as "won" in Stripe Dashboard
2. Verify webhook handling

**Expected Results**:
- ✅ Webhook `charge.dispute.closed` received (status: won)
- ✅ Dispute status updated in database
- ✅ Subscription remains active
- ✅ Funds restored
- ✅ User notified of resolution

### Test Case 7.3: Dispute Lost

**Steps**:
1. Resolve dispute as "lost" in Stripe Dashboard
2. Verify subscription handling

**Expected Results**:
- ✅ Webhook `charge.dispute.closed` received (status: lost)
- ✅ Subscription potentially canceled (depends on policy)
- ✅ Funds not restored
- ✅ User notified of outcome
- ✅ May require manual intervention

### Test Case 7.4: Multiple Disputes (Fraud Pattern)

**Steps**:
1. Simulate multiple disputes from same user
2. Verify fraud detection

**Expected Results**:
- ✅ Pattern detected
- ✅ User flagged for review
- ✅ Additional verification required for future purchases
- ✅ Admin alert sent

## Stripe Dashboard Reconciliation

### Daily Reconciliation Checklist

1. **Revenue Verification**
   ```
   - Compare Stripe revenue reports with database
   - Check subscription_events table
   - Verify payment intent amounts
   - Reconcile refunds and disputes
   ```

2. **Subscription Status Sync**
   ```sql
   -- Query to check status mismatches
   SELECT s.id, s.user_id, s.status, s.stripe_subscription_id, s.updated_at
   FROM subscriptions s
   WHERE s.status IN ('active', 'past_due', 'trialing')
   ORDER BY s.updated_at DESC;
   ```
   - Cross-reference with Stripe Dashboard
   - Investigate any mismatches
   - Check for missed webhooks

3. **Failed Webhook Investigation**
   ```sql
   -- Check webhook retry queue
   SELECT * FROM webhook_retry_queue 
   WHERE status = 'failed' 
   ORDER BY created_at DESC;
   ```
   - Review failed webhook events
   - Manually replay if necessary
   - Fix root cause (endpoint down, invalid data, etc.)

4. **Dunning Status Review**
   ```sql
   -- Check active dunning attempts
   SELECT d.*, s.user_id, s.status 
   FROM dunning_attempts d
   JOIN subscriptions s ON d.subscription_id = s.id
   WHERE d.status = 'active'
   ORDER BY d.next_retry_at;
   ```
   - Review pending payment retries
   - Check grace period expirations
   - Verify email notifications sent

5. **Monthly Metrics Reconciliation**
   - MRR (Monthly Recurring Revenue)
   - ARR (Annual Recurring Revenue)
   - Churn rate
   - Payment success rate
   - Average subscription value
   - Customer lifetime value

### Automated Reconciliation Script

```bash
# Run daily reconciliation report
cd backend
go run cmd/backfill-stripe-metrics/main.go --date=$(date +%Y-%m-%d)
```

## Edge Cases and Known Issues

### Edge Case 1: Trial Period Subscriptions

**Scenario**: User subscribes with trial period

**Behavior**:
- Subscription status: "trialing"
- No immediate charge
- Invoice generated at trial end
- Webhook `customer.subscription.trial_will_end` sent 3 days before

**Test Steps**:
1. Create checkout session with trial period
2. Verify trial status set correctly
3. Wait for trial to end (or simulate time)
4. Verify conversion to active subscription

### Edge Case 2: Timezone Handling

**Issue**: Period start/end times may differ between Stripe (UTC) and local timezone

**Solution**:
- Store all dates in UTC in database
- Convert to user timezone for display
- Use Unix timestamps from Stripe webhooks

**Verification**:
```sql
SELECT 
    current_period_start,
    current_period_end,
    EXTRACT(TIMEZONE FROM current_period_start) as start_tz,
    EXTRACT(TIMEZONE FROM current_period_end) as end_tz
FROM subscriptions;
```

### Edge Case 3: Concurrent Subscription Updates

**Scenario**: User initiates two subscription changes simultaneously

**Behavior**:
- Idempotency keys prevent duplicate charges
- Later webhook overrides earlier one
- Race condition possible

**Mitigation**:
- Use database transactions
- Implement optimistic locking
- Check event timestamps

### Edge Case 4: Webhook Delivery Failures

**Scenario**: Webhook endpoint unavailable or returns error

**Behavior**:
- Stripe retries with exponential backoff
- Up to 3 days of retries
- Events may arrive out of order

**Mitigation**:
- Implement webhook retry queue
- Log all webhook events
- Use event IDs for idempotency
- Manual replay capability

**Recovery Process**:
```sql
-- Find missing events
SELECT stripe_event_id FROM subscription_events 
WHERE created_at > NOW() - INTERVAL '7 days';

-- Compare with Stripe Dashboard events
-- Manually replay missing events via API
```

### Edge Case 5: Payment Method Expiration

**Scenario**: Card expires before next billing date

**Behavior**:
- Stripe sends `customer.source.expiring` webhook ~30 days before
- Email notification sent to user
- Payment may fail if not updated

**Test Steps**:
1. Set card expiration to current month + 1
2. Verify expiration notification sent
3. Test payment failure handling
4. Verify grace period applied

### Edge Case 6: Duplicate Customer Accounts

**Scenario**: User creates multiple accounts with same email

**Behavior**:
- Multiple Stripe customers created
- Potential for duplicate subscriptions

**Mitigation**:
- Check for existing customer by email
- Merge customers if necessary
- Implement email uniqueness constraint

### Edge Case 7: Refund Requests

**Scenario**: User requests refund mid-cycle

**Process**:
1. Review refund policy (e.g., 30-day guarantee)
2. Issue refund via Stripe Dashboard
3. Cancel subscription
4. Downgrade user to free tier
5. Log refund in audit trail

**Expected Behavior**:
- Webhook `charge.refunded` received
- Full or partial refund processed
- Subscription canceled or adjusted
- User notified

### Edge Case 8: Tax Calculation Errors

**Scenario**: Stripe Tax enabled but fails to calculate

**Behavior**:
- Checkout may fail
- Invoice amount incorrect

**Mitigation**:
- Fallback to manual tax calculation
- Verify billing address completeness
- Test with various locales

**Test Steps**:
1. Enable Stripe Tax in settings
2. Test checkout with various countries
3. Verify tax rates applied correctly
4. Check invoice line items

### Edge Case 9: Zero-Dollar Invoices

**Scenario**: 100% discount coupon or credit balance covers invoice

**Behavior**:
- Invoice marked as paid
- No payment intent created
- Subscription remains active

**Verification**:
```sql
SELECT * FROM subscription_events 
WHERE event_type = 'invoice_paid' 
AND data->>'amount_paid' = '0';
```

### Edge Case 10: Subscription in Incomplete State

**Scenario**: Payment requires additional authentication but not completed

**Behavior**:
- Subscription status: "incomplete"
- Payment intent status: "requires_action"
- User must complete authentication

**Resolution**:
- Send reminder email with payment link
- After 23 hours, subscription canceled if not completed
- Webhook `customer.subscription.updated` sent

## Testing Checklist Summary

- [ ] **New Subscription Creation**
  - [ ] Monthly subscription
  - [ ] Yearly subscription
  - [ ] Subscription with coupon
  - [ ] Failed payment on creation
  - [ ] Trial period subscription

- [ ] **Subscription Cancellation**
  - [ ] Cancel immediately
  - [ ] Cancel at period end
  - [ ] Portal session access

- [ ] **Payment Method Updates**
  - [ ] Update via portal
  - [ ] Add additional methods
  - [ ] Handle expired cards

- [ ] **Payment Failure Handling**
  - [ ] First failure (dunning start)
  - [ ] Multiple failures
  - [ ] Recovery after failure
  - [ ] Grace period behavior

- [ ] **Proration Calculations**
  - [ ] Upgrade monthly to yearly
  - [ ] Downgrade yearly to monthly
  - [ ] Invalid plan changes

- [ ] **Subscription Reactivation**
  - [ ] Cancel scheduled cancellation
  - [ ] Create new after cancellation

- [ ] **Dispute/Chargeback Handling**
  - [ ] Dispute created
  - [ ] Dispute won
  - [ ] Dispute lost
  - [ ] Multiple disputes

- [ ] **Dashboard Reconciliation**
  - [ ] Daily revenue sync
  - [ ] Subscription status sync
  - [ ] Failed webhook review
  - [ ] Dunning status check
  - [ ] Monthly metrics

- [ ] **Edge Cases**
  - [ ] Trial periods
  - [ ] Timezone handling
  - [ ] Concurrent updates
  - [ ] Webhook failures
  - [ ] Payment method expiration
  - [ ] Duplicate customers
  - [ ] Refund requests
  - [ ] Tax calculations
  - [ ] Zero-dollar invoices
  - [ ] Incomplete subscriptions

## Reporting Issues

When you discover issues during testing:

1. **Document the Issue**:
   - Exact steps to reproduce
   - Expected vs actual behavior
   - Screenshots/logs
   - Stripe event IDs
   - User/subscription IDs

2. **Check Logs**:
   ```bash
   # Backend logs
   docker logs clpr-backend | grep -i "stripe\|subscription\|webhook"
   
   # Database logs
   SELECT * FROM subscription_events WHERE created_at > NOW() - INTERVAL '1 hour';
   ```

3. **Stripe Dashboard Investigation**:
   - Events log
   - Customer timeline
   - Payment attempts
   - Webhook delivery status

4. **Create GitHub Issue**:
   - Label: `bug`, `stripe`, `subscription`
   - Include all documentation
   - Reference this testing guide

## Success Metrics

After completing all tests, verify:

- ✅ 100% webhook delivery success rate
- ✅ 0 subscription status mismatches
- ✅ Payment success rate > 95%
- ✅ Dunning recovery rate tracked
- ✅ All dispute notifications sent
- ✅ Revenue reconciliation matches within $0.01
- ✅ No failed tests in CI/CD pipeline
- ✅ All edge cases documented

## Next Steps

1. Complete all test cases in this guide
2. Document any new edge cases discovered
3. Update automated tests based on findings
4. Schedule production cutover
5. Set up monitoring and alerts
6. Plan for ongoing testing and maintenance

## Integration Test Details

The automated integration tests are located in `backend/tests/integration/premium/subscription_webhook_integration_test.go`.

### Test Implementation

**TestWebhookIdempotencyWithDatabaseAssertion**
- Validates webhook endpoint exists and responds
- Tests database schema supports event tracking (`stripe_webhooks_log` table)
- Verifies handler infrastructure for concurrent requests
- **Note**: Does not test actual idempotency logic (requires valid Stripe signatures)

**TestEntitlementUpdatesOnSubscriptionStatusChanges**
- Tests subscription creation and database persistence
- Validates `IsProUser` logic for different subscription statuses
- Tests grace period handling infrastructure
- **Note**: Does not test webhook-driven entitlement updates (requires valid Stripe signatures)

**TestWebhookRetryLogic**
- Validates retry queue database tables exist (`webhook_retry_queue`)
- Verifies schema supports exponential backoff (`next_retry_at` column)
- Tests max retries column exists
- **Note**: Does not test actual retry processing or backoff behavior

**TestProrationCalculationsWithValidation**
- Tests plan change endpoint exists and responds
- Validates webhook handlers can receive invoice events
- Verifies database schema supports proration tracking
- **Note**: Does not validate actual proration amounts (requires valid Stripe API calls)

**TestPaymentFailureHandlingWithAlerts**
- Tests payment failure webhook handlers exist
- Validates database can track failure events
- Verifies handler infrastructure for multiple failure scenarios
- **Note**: Does not test actual dunning process or escalation logic (requires valid Stripe signatures)

### Running Individual Tests

```bash
# Test webhook idempotency
go test -v -tags=integration ./tests/integration/premium/ -run TestWebhookIdempotencyWithDatabaseAssertion

# Test entitlement updates
go test -v -tags=integration ./tests/integration/premium/ -run TestEntitlementUpdatesOnSubscriptionStatusChanges

# Test retry logic
go test -v -tags=integration ./tests/integration/premium/ -run TestWebhookRetryLogic

# Test proration
go test -v -tags=integration ./tests/integration/premium/ -run TestProrationCalculationsWithValidation

# Test payment failures
go test -v -tags=integration ./tests/integration/premium/ -run TestPaymentFailureHandlingWithAlerts
```

### Test Database

Tests use the test database configured in `docker-compose.test.yml`:
- **Host**: localhost:5437
- **Database**: clpr_test
- **User**: clpr
- **Password**: clpr_password

The Makefile target `test-integration-stripe` automatically:
1. Starts the test database
2. Runs migrations
3. Executes integration tests
4. Stops the test database

### CI/CD Integration

For CI pipelines, set environment variables:
```bash
export TEST_STRIPE_SECRET_KEY=sk_test_your_test_key
export TEST_STRIPE_WEBHOOK_SECRET=whsec_test_your_webhook_secret
```

Tests will run with mock Stripe clients if keys are not provided, but with actual Stripe test keys, tests will validate against real Stripe API behavior in test mode.

---

**Last Updated**: December 26, 2025  
**Version**: 1.1  
**Maintained By**: Clipper Engineering Team
