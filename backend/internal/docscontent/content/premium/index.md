---
title: "Premium & Monetization"
summary: "Subscription model, pricing tiers, and payment integration."
tags: ["premium", "hub", "index"]
area: "product"
status: "stable"
owner: "team-core"
version: "1.0"
last_reviewed: 2025-12-01
aliases: ["premium hub", "monetization"]
---

# Premium & Monetization

This section covers Clipper's subscription model and premium features.

## Quick Links

- [[overview|Premium Overview]] - Subscription philosophy
- [[tiers|Pricing Tiers]] - Free vs Pro comparison
- [[entitlements|Entitlements]] - Feature gating
- [[entitlement-matrix|Entitlement Matrix]] - Feature access matrix
- [[subscription-privileges-matrix|Subscription Privileges Matrix]] - Privilege mapping
- [[subscriptions|Subscriptions]] - Subscription management
- [[stripe|Stripe Integration]] - Payment processing
- [[stripe-testing|Stripe Testing]] - Testing Stripe integration
- [[dunning|Dunning]] - Payment recovery process
- [[trials-and-discounts|Trials & Discounts]] - Trial periods and discount codes
- [[paywall-analytics|Paywall Analytics]] - Conversion analytics

## Documentation Index

```dataview
TABLE title, summary, status, last_reviewed
FROM "docs/premium"
WHERE file.name != "index"
SORT title ASC
```

## Subscription Tiers

| Tier | Price | Key Features |
|------|-------|--------------|
| Free | $0 | Core features, basic search, 10 clips/day |
| Pro | $4.99/mo | Unlimited favorites, advanced search, 50 clips/day |

See [[tiers|Pricing Tiers]] for complete comparison.

---

**See also:** [[../backend/api|API Reference]] · [[../index|Documentation Home]]
