---
title: "Admin System Configuration Ui"
summary: "Build system configuration UI allowing admins to toggle feature flags, edit rate limits, manage notification settings, and control premium tiers without code deployment."
tags: ["operations"]
area: "operations"
status: "stable"
owner: "team-core"
version: "1.0"
last_reviewed: 2026-01-29
---

## Summary

Build system configuration UI allowing admins to toggle feature flags, edit rate limits, manage notification settings, and control premium tiers without code deployment.

## Scope

### Backend Requirements

- Config API: `GET /api/v1/admin/config` returns all editable settings
- Config update API: `PATCH /api/v1/admin/config` with key-value pairs
- Settings schema: feature_flags (object), rate_limits (object), email_templates (object), premium_tiers (array), maintenance_mode (bool)
- Config cache: changes propagate to all services within 30s
- Audit log: all config changes tracked with admin ID and timestamp
- Rollback: ability to revert config to previous version

### Frontend Requirements

- Configuration panel with tabs:
  - **Feature Flags**: toggles for experimental features (e.g., theatre_mode, watch_parties, webhooks)
  - **Rate Limits**: input fields for submission rate limit (per user), API rate limit (per IP), search rate limit
  - **Email Templates**: editor for notification emails (account verification, password reset, moderation notice) with variables ({{user_name}}, {{reason}})
  - **Premium Tiers**: editor for tier definitions (free, pro, premium) with features, pricing, storage limits
  - **Maintenance Mode**: toggle (blue banner shows maintenance message to users when enabled)
  - **System Alerts**: enable/disable alerts to admins for errors, high latency, etc.
- Change history: list of recent config changes with revert button
- Preview: show email template preview before saving
- Validation: warn about potentially dangerous settings (e.g., setting rate limits too low)

### Acceptance Criteria

- [ ] Feature flag toggle takes effect within 30s globally
- [ ] Rate limit changes apply to new requests immediately
- [ ] Email template changes reflected in next email sent
- [ ] Premium tier changes accessible to frontend within 1 minute
- [ ] Maintenance mode toggle hides login/submission for all users
- [ ] All config changes logged with admin ID and old/new values
- [ ] Revert function works for recent changes (last 7 days)
- [ ] No syntax errors in email template prevent save
- [ ] Premium tier pricing changes don't affect current subscriptions
- [ ] Invalid config (e.g., negative rate limit) rejected with helpful error

## Definition of Done

- Admins can manage feature rollout without code deploy
- Config changes reach all services quickly
- No production incidents from config changes
- Audit trail proves compliance
