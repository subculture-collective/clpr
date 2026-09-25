---
title: "Self Qa Flow Checklist"
summary: "Use this checklist for quick, repeatable validation before sharing with other testers. Update the date and environment at the top of each run."
tags: ["testing"]
area: "testing"
status: "stable"
owner: "team-core"
version: "1.0"
last_reviewed: 2026-01-29
---

# Self QA Flow Checklist

Use this checklist for quick, repeatable validation before sharing with other testers. Update the date and environment at the top of each run.

**Date**:
**Environment**: (local / staging / prod)
**Build/Commit**:
**Tester**:

---

## 1) Preflight
- [ ] Environment up and reachable (web, API, DB, cache, search)
- [ ] Feature flags set as expected
- [ ] Backend log level set to debug (GIN_MODE=debug)
- [ ] Frontend logging in debug mode (dev build / NODE_ENV=development)
- [ ] Test data seeded or known accounts available
- [ ] Error monitoring/logs visible
- [ ] Known test accounts confirmed (creator, moderator, viewer)

## 2) Authentication & Account
- [ ] Sign up (new account)
- [ ] Login (existing account)
- [ ] Logout
- [ ] Password reset flow
- [ ] Session persistence (refresh / reopen browser)
- [ ] Role-based access (viewer vs moderator)

## 3) Core Navigation & Layout
- [ ] Landing/home loads quickly
- [ ] Global nav links work
- [ ] Responsive layout (desktop + mobile)
- [ ] Settings page reachable

## 4) Clips (Core Product)
- [ ] Create/upload clip (happy path)
- [ ] View clip detail page
- [ ] Share clip link
- [ ] Clip list/pagination loads
- [ ] Clip metadata edits save
- [ ] Delete clip (if allowed)

## 5) Search & Discovery
- [ ] Search returns results
- [ ] Empty state for no results
- [ ] Filters/sorting (if present)
- [ ] Search result click-through

## 6) Comments & Engagement
- [ ] Add comment
- [ ] Edit comment
- [ ] Delete comment
- [ ] Report / moderation actions (if available)

## 7) Moderation (High Risk)
- [ ] Twitch moderation actions (ban/unban/etc.)
- [ ] Error alert behavior (single vs multiple alerts)
- [ ] Sync bans modal (start, progress, completion)
- [ ] Rate limit behavior (UI + API)
- [ ] Audit log updates (if available)

## 8) Notifications & Messaging
- [ ] In-app notification appears
- [ ] Notification click-through
- [ ] Unread indicator updates

## 9) Premium / Billing (if enabled)
- [ ] Start checkout
- [ ] Cancel checkout
- [ ] Complete subscription
- [ ] Manage subscription settings
- [ ] Webhook effect reflected in UI

## 10) Watch Parties / Live Features (if enabled)
- [ ] Create/join session
- [ ] Sync status updates
- [ ] End session cleanly

## 11) Settings & Profile
- [ ] Update profile info
- [ ] Update privacy or preferences
- [ ] Deactivate/delete account (if enabled)

## 12) Error Handling & Resilience
- [ ] API error surfaces readable message
- [ ] Retry action works after transient error
- [ ] Offline/slow network behavior acceptable

---

## Bug Log (quick capture)
| Time | Area | Steps | Expected | Actual | Severity | Link |
|------|------|-------|----------|--------|----------|------|
|      |      |       |          |        |          |      |

---

## Notes
- Focus extra attention on **Twitch moderation actions** and **Sync bans modal** (known flaky areas in tests).
- If any P0/P1 bug is found, capture repro steps immediately and run the relevant targeted test before hotfix.
