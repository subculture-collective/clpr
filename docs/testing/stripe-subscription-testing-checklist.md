---
title: "Stripe Subscription Testing Checklist"
summary: "Quick reference checklist for manual testing of Stripe subscription flows before production deployment."
tags: ["testing"]
area: "testing"
status: "stable"
owner: "team-core"
version: "1.0"
last_reviewed: 2026-01-29
---

# Stripe Subscription Manual Testing Checklist

Quick reference checklist for manual testing of Stripe subscription flows before production deployment.

## Pre-Testing Setup

- [ ] Stripe Test Mode enabled
- [ ] Test API keys configured
- [ ] Webhook endpoint configured and verified
- [ ] Test cards available (4242 4242 4242 4242, etc.)
- [ ] Test user accounts created
- [ ] Database access for verification queries

## Core Flows

### 1. New Subscription Creation ⏱️ Est: 30 min

- [ ] **Monthly Subscription**
  - [ ] Navigate to pricing page
  - [ ] Click "Subscribe Monthly"
  - [ ] Complete checkout with card 4242...4242
  - [ ] Verify success redirect
  - [ ] Check subscription status in database
  - [ ] Verify user has pro access
  - [ ] Check webhook logs for `subscription.created`

- [ ] **Yearly Subscription**
  - [ ] Navigate to pricing page
  - [ ] Click "Subscribe Yearly"
  - [ ] Complete checkout with card 4242...4242
  - [ ] Verify correct period end date (~365 days)
  - [ ] Check yearly price applied

- [ ] **Subscription with Coupon**
  - [ ] Create test coupon in Stripe (e.g., TESTDISCOUNT)
  - [ ] Apply coupon during checkout
  - [ ] Verify discount in checkout total
  - [ ] Verify discount in database

- [ ] **Failed Creation**
  - [ ] Attempt checkout with declined card 4000...0002
  - [ ] Verify error message displayed
  - [ ] Verify no subscription created
  - [ ] Verify user remains on free tier

**Verification SQL**:
```sql
SELECT id, user_id, status, tier, stripe_subscription_id, current_period_end 
FROM subscriptions WHERE user_id = 'USER_UUID';
```

### 2. Subscription Cancellation ⏱️ Est: 20 min

- [ ] **Cancel Immediately**
  - [ ] Access Stripe Customer Portal
  - [ ] Select "Cancel Immediately"
  - [ ] Confirm cancellation
  - [ ] Verify subscription status = "canceled"
  - [ ] Verify user downgraded to free
  - [ ] Check webhook logs for `subscription.deleted`

- [ ] **Cancel at Period End**
  - [ ] Access Stripe Customer Portal
  - [ ] Select "Cancel at period end"
  - [ ] Confirm cancellation
  - [ ] Verify `cancel_at_period_end = true`
  - [ ] Verify user still has pro access
  - [ ] Verify access expires at period end
  - [ ] Check webhook logs for `subscription.updated`

**Verification SQL**:
```sql
SELECT status, tier, cancel_at_period_end, current_period_end, canceled_at 
FROM subscriptions WHERE user_id = 'USER_UUID';
```

### 3. Payment Method Update ⏱️ Est: 15 min

- [ ] **Update via Portal**
  - [ ] Access Stripe Customer Portal
  - [ ] Click "Update payment method"
  - [ ] Add new test card
  - [ ] Set as default
  - [ ] Verify update in Stripe Dashboard
  - [ ] Check webhook logs for `customer.updated`

- [ ] **Add Backup Payment Method**
  - [ ] Add second payment method
  - [ ] Keep first as default
  - [ ] Verify both methods stored

### 4. Payment Failure Handling ⏱️ Est: 45 min

- [ ] **First Payment Failure**
  - [ ] Update payment method to card 4000...9995 (insufficient funds)
  - [ ] Wait for/trigger billing cycle
  - [ ] Verify webhook `invoice.payment_failed` received
  - [ ] Verify subscription status = "past_due"
  - [ ] Verify dunning email sent to user
  - [ ] Verify grace period initiated (3 days)
  - [ ] Verify user retains pro access during grace
  - [ ] Check dunning_attempts table

- [ ] **Payment Recovery**
  - [ ] Update to valid card 4242...4242
  - [ ] Trigger payment retry
  - [ ] Verify webhook `invoice.paid` received
  - [ ] Verify subscription status = "active"
  - [ ] Verify dunning cleared
  - [ ] Verify success email sent

- [ ] **Payment Intent Failure**
  - [ ] Test with card requiring auth 4000...3155
  - [ ] Abandon authentication
  - [ ] Verify webhook `payment_intent.payment_failed`
  - [ ] Verify failure logged

**Verification SQL**:
```sql
SELECT * FROM dunning_attempts WHERE subscription_id = 'SUBSCRIPTION_UUID' ORDER BY created_at DESC;
SELECT status, grace_period_end FROM subscriptions WHERE user_id = 'USER_UUID';
```

### 5. Proration Calculations ⏱️ Est: 30 min

- [ ] **Upgrade Monthly to Yearly**
  - [ ] User has active monthly subscription
  - [ ] Call API: `POST /api/v1/subscriptions/change-plan` with yearly price
  - [ ] Verify immediate proration invoice
  - [ ] Verify credit for unused monthly time
  - [ ] Verify charge for yearly subscription
  - [ ] Verify net amount correct
  - [ ] Verify subscription updated to yearly
  - [ ] Verify period end updated to +365 days
  - [ ] Check webhook logs for `subscription.updated`

- [ ] **Downgrade Yearly to Monthly**
  - [ ] User has active yearly subscription
  - [ ] Call API with monthly price
  - [ ] Verify proration invoice/credit
  - [ ] Verify subscription updated to monthly
  - [ ] Verify billing recalculated

- [ ] **Invalid Changes**
  - [ ] Attempt change to current plan → expect 400 error
  - [ ] Attempt change to invalid price → expect 400 error

**Verification SQL**:
```sql
SELECT event_type, data FROM subscription_events 
WHERE subscription_id = 'SUBSCRIPTION_UUID' AND event_type = 'subscription_updated' 
ORDER BY created_at DESC LIMIT 1;
```

### 6. Subscription Reactivation ⏱️ Est: 20 min

- [ ] **Reactivate Scheduled Cancellation**
  - [ ] User has `cancel_at_period_end = true`
  - [ ] Access Stripe Customer Portal
  - [ ] Click "Reactivate Subscription"
  - [ ] Verify `cancel_at_period_end = false`
  - [ ] Verify subscription continues beyond period
  - [ ] Verify user notified

- [ ] **New Subscription After Cancellation**
  - [ ] User's subscription fully canceled
  - [ ] Navigate to pricing page
  - [ ] Subscribe again
  - [ ] Verify new subscription created
  - [ ] Verify old subscription still canceled
  - [ ] Verify user upgraded to pro

**Verification SQL**:
```sql
SELECT cancel_at_period_end, canceled_at, stripe_subscription_id 
FROM subscriptions WHERE user_id = 'USER_UUID';
```

### 7. Dispute/Chargeback Handling ⏱️ Est: 30 min

- [ ] **Dispute Created**
  - [ ] Create test dispute in Stripe Dashboard
  - [ ] Select reason (e.g., "fraudulent")
  - [ ] Verify webhook `charge.dispute.created` received
  - [ ] Verify dispute logged in database
  - [ ] Verify email sent to user
  - [ ] Verify admin notification sent
  - [ ] Verify subscription still active

- [ ] **Dispute Resolved (Won)**
  - [ ] Resolve dispute as "won" in Stripe
  - [ ] Verify webhook `charge.dispute.closed` (won)
  - [ ] Verify dispute status updated
  - [ ] Verify subscription remains active

- [ ] **Dispute Resolved (Lost)**
  - [ ] Resolve dispute as "lost" in Stripe
  - [ ] Verify webhook `charge.dispute.closed` (lost)
  - [ ] Verify appropriate action taken
  - [ ] Verify user notified

**Verification SQL**:
```sql
SELECT * FROM subscription_events WHERE event_type = 'dispute_created' ORDER BY created_at DESC;
SELECT * FROM audit_logs WHERE action = 'dispute_created' ORDER BY created_at DESC;
```

### 8. Stripe Dashboard Reconciliation ⏱️ Est: 60 min

- [ ] **Daily Revenue Check**
  - [ ] Export revenue from Stripe Dashboard
  - [ ] Run revenue report from database
  - [ ] Compare totals
  - [ ] Investigate discrepancies

- [ ] **Subscription Status Sync**
  - [ ] Query active subscriptions from database
  - [ ] Cross-reference with Stripe Dashboard
  - [ ] Verify status matches
  - [ ] Check for missed webhooks

- [ ] **Failed Webhooks**
  - [ ] Check webhook retry queue in database
  - [ ] Review failed events in Stripe Dashboard
  - [ ] Manually replay if needed
  - [ ] Fix root causes

- [ ] **Dunning Status**
  - [ ] Query active dunning attempts
  - [ ] Verify retry schedules
  - [ ] Check email notification logs
  - [ ] Verify grace periods

- [ ] **Monthly Metrics**
  - [ ] Calculate MRR (Monthly Recurring Revenue)
  - [ ] Calculate ARR (Annual Recurring Revenue)
  - [ ] Calculate churn rate
  - [ ] Calculate payment success rate
  - [ ] Verify against Stripe analytics

**Verification Queries**:
```sql
-- Active subscriptions
SELECT COUNT(*), status FROM subscriptions GROUP BY status;

-- Failed webhooks
SELECT * FROM webhook_retry_queue WHERE status = 'failed' ORDER BY created_at DESC LIMIT 20;

-- Active dunning
SELECT d.*, s.user_id FROM dunning_attempts d 
JOIN subscriptions s ON d.subscription_id = s.id 
WHERE d.status = 'active' ORDER BY d.next_retry_at;

-- Recent revenue
SELECT SUM(CAST(data->>'amount_paid' AS INTEGER)) / 100.0 AS total_revenue
FROM subscription_events 
WHERE event_type = 'invoice_paid' 
AND created_at > NOW() - INTERVAL '30 days';
```

## Edge Cases Testing ⏱️ Est: 90 min

- [ ] **Trial Period**
  - [ ] Create subscription with trial
  - [ ] Verify status = "trialing"
  - [ ] Verify no immediate charge
  - [ ] Wait for trial end (or simulate)
  - [ ] Verify conversion to active

- [ ] **Timezone Handling**
  - [ ] Verify all dates stored in UTC
  - [ ] Test with users in different timezones
  - [ ] Verify display converts correctly

- [ ] **Webhook Order**
  - [ ] Simulate out-of-order webhook delivery
  - [ ] Verify idempotency handling
  - [ ] Verify event timestamps respected

- [ ] **Payment Method Expiration**
  - [ ] Set card expiration to next month
  - [ ] Verify expiration warning sent
  - [ ] Test payment failure on expiration
  - [ ] Verify grace period applied

- [ ] **Duplicate Customers**
  - [ ] Attempt to create customer with existing email
  - [ ] Verify handling (merge or error)

- [ ] **Refund Request**
  - [ ] Issue refund via Stripe Dashboard
  - [ ] Verify webhook `charge.refunded` received
  - [ ] Verify subscription canceled
  - [ ] Verify user downgraded

- [ ] **Tax Calculations**
  - [ ] Enable Stripe Tax
  - [ ] Test checkout with various countries
  - [ ] Verify tax rates applied
  - [ ] Verify invoice line items correct

- [ ] **Zero-Dollar Invoice**
  - [ ] Apply 100% discount coupon
  - [ ] Complete checkout
  - [ ] Verify invoice marked paid
  - [ ] Verify subscription active

- [ ] **Incomplete Subscription**
  - [ ] Use card requiring auth 4000...3155
  - [ ] Abandon authentication
  - [ ] Verify status = "incomplete"
  - [ ] Verify reminder email sent
  - [ ] Verify auto-cancellation after 23 hours

## Post-Testing Validation

- [ ] **Automated Tests Pass**
  ```bash
  make test-integration-premium
  ```

- [ ] **Code Review**
  - [ ] Webhook handlers reviewed
  - [ ] Error handling verified
  - [ ] Logging sufficient
  - [ ] Idempotency implemented

- [ ] **Documentation Updated**
  - [ ] API documentation
  - [ ] Webhook event handling
  - [ ] Edge cases documented
  - [ ] Runbooks created

- [ ] **Monitoring Setup**
  - [ ] Webhook delivery alerts
  - [ ] Payment failure alerts
  - [ ] Revenue anomaly alerts
  - [ ] Dashboard for key metrics

- [ ] **Production Readiness**
  - [ ] Production Stripe keys configured
  - [ ] Webhook secrets rotated
  - [ ] Rate limits tested
  - [ ] Backup/recovery tested
  - [ ] Rollback plan documented

## Testing Sign-off

**Tester**: ___________________________  
**Date**: ___________________________  
**Environment**: Test / Staging / Production  
**Test Duration**: ___________ hours

**Overall Result**: ☐ Pass ☐ Fail ☐ Pass with Issues

**Issues Found**:
- Issue 1: ___________________________
- Issue 2: ___________________________
- Issue 3: ___________________________

**Notes**:
```
[Add any additional notes, observations, or recommendations here]
```

**Approvals**:
- Engineering Lead: ___________________________
- Product Manager: ___________________________
- QA Lead: ___________________________

---

## Quick Reference Links

- Full Testing Guide
- [Stripe Dashboard](https://dashboard.stripe.com/test/dashboard)
- [Stripe Test Cards](https://stripe.com/docs/testing#cards)
- [Webhook Events](https://dashboard.stripe.com/test/webhooks)
- API Documentation

## Emergency Contacts

- **Engineering Lead**: [Name] - [Email/Slack]
- **DevOps**: [Name] - [Email/Slack]
- **Stripe Support**: <support@stripe.com>

---

**Last Updated**: December 24, 2025  
**Version**: 1.0
