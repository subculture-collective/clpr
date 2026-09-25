---
title: "Analytics Dashboard Setup"
summary: "**Time Required**: 4-6 hours"
tags: ["mobile"]
area: "mobile"
status: "stable"
owner: "team-core"
version: "1.0"
last_reviewed: 2026-01-29
---

# Mobile Analytics Dashboard - Quick Setup Guide

**Time Required**: 4-6 hours  
**Prerequisites**: PostHog project with mobile SDK integration complete

## Table of Contents

- [Overview](#overview)
- [Before You Start](#before-you-start)
- [Setup Order](#setup-order)
- [Dashboard 1: Mobile DAU/MAU](#dashboard-1-mobile-daumau-30-min)
- [Dashboard 2: Mobile Screen Views](#dashboard-2-mobile-screen-views-45-min)
- [Dashboard 3: Mobile User Funnels](#dashboard-3-mobile-user-funnels-60-min)
- [Dashboard 4: Mobile Retention Analysis](#dashboard-4-mobile-retention-analysis-90-min)
- [Dashboard 5: Mobile Stability](#dashboard-5-mobile-stability-60-min)
- [Post-Setup Tasks](#post-setup-tasks)
- [Validation Checklist](#validation-checklist)
- [Troubleshooting](#troubleshooting)
- [Next Steps](#next-steps)
- [Support](#support)

## Overview

This guide provides step-by-step instructions to set up the five core mobile analytics dashboards in PostHog.

## Before You Start

✅ **Verify Prerequisites**:
- [ ] PostHog project exists (app.posthog.com or self-hosted)
- [ ] Mobile app has PostHog SDK integrated (see POSTHOG_ANALYTICS.md)
- [ ] Test events are flowing to PostHog (check Live Events tab)
- [ ] You have dashboard creation permissions

## Setup Order

Follow this order for most efficient setup:

1. ✅ **Mobile DAU/MAU** (30 min) - Foundational metrics
2. ✅ **Mobile Screen Views** (45 min) - Navigation tracking
3. ✅ **Mobile User Funnels** (60 min) - Conversion analysis
4. ✅ **Mobile Retention Analysis** (90 min) - Cohort tracking
5. ✅ **Mobile Stability** (60 min) - Error monitoring

**Total Time**: ~4.5 hours

---

## Dashboard 1: Mobile DAU/MAU (30 min)

### Step 1: Create Dashboard
1. Go to PostHog → Dashboards → New Dashboard
2. Name: "Mobile DAU/MAU"
3. Description: "Daily and monthly active users for mobile platform"
4. Tags: `mobile`, `analytics`, `kpi`

### Step 2: Add DAU Insight
1. Click "Add insight" → New insight
2. Configuration:
   - Visualization: Trends (Line graph)
   - Series: Select "All events"
   - Filters: Add `platform = mobile`
   - Aggregation: Unique users
   - Time range: Last 90 days
   - Interval: Day
3. Name: "Daily Active Users (DAU)"
4. Add to dashboard

### Step 3: Add MAU Insight
1. Click "Add insight" → New insight
2. Configuration:
   - Visualization: Trends (Line graph with area)
   - Series: Select "All events"
   - Filters: Add `platform = mobile`
   - Aggregation: Unique users
   - Time range: Last 12 months
   - Interval: Month
3. Name: "Monthly Active Users (MAU)"
4. Add to dashboard

### Step 4: Add Stickiness Insight
1. Click "Add insight" → New insight
2. Configuration:
   - Visualization: Stickiness
   - Event: Any event
   - Filters: Add `platform = mobile`
   - Time range: Current month
3. Name: "Stickiness Ratio (DAU/MAU)"
4. Add to dashboard

### Step 5: Add Summary Cards
1. Add 4 "Big Number" insights:
   - Current DAU (today)
   - Current MAU (last 30 days)
   - Stickiness % (DAU/MAU * 100)
   - Week-over-week DAU change %
2. Position at top of dashboard

**Dashboard 1 Complete** ✅

---

## Dashboard 2: Mobile Screen Views (45 min)

### Step 1: Create Dashboard
1. Go to PostHog → Dashboards → New Dashboard
2. Name: "Mobile Screen Views"
3. Description: "Screen navigation patterns and engagement"
4. Tags: `mobile`, `ux`, `navigation`

### Step 2: Top Screens Insight
1. Click "Add insight" → New insight
2. Configuration:
   - Visualization: Bar chart (horizontal)
   - Event: `screen_viewed`
   - Breakdown: `screen_name`
   - Filters: `platform = mobile`
   - Time range: Last 7 days
   - Display: Top 20
3. Name: "Top 20 Screens by Views"
4. Add to dashboard

### Step 3: Screen Trend Over Time
1. Click "Add insight" → New insight
2. Configuration:
   - Visualization: Line graph
   - Event: `screen_viewed`
   - Breakdown: `screen_name` (select top 10 only)
   - Filters: `platform = mobile`
   - Time range: Last 30 days
   - Interval: Daily
3. Name: "Screen Views Over Time (Top 10)"
4. Add to dashboard

### Step 4: Screen Engagement Time
1. Click "Add insight" → New insight
2. Configuration:
   - Visualization: Table
   - Event: `screen_viewed`
   - Property: Average of `time_on_screen` (if tracked)
   - Breakdown: `screen_name`
   - Filters: `platform = mobile`
   - Time range: Last 7 days
   - Sort: By average descending
3. Name: "Average Time on Screen"
4. Add to dashboard

### Step 5: Navigation Paths (Optional)
1. Click "Add insight" → New insight
2. Configuration:
   - Visualization: Path (if available in your PostHog version)
   - Start event: `screen_viewed`
   - End event: `screen_viewed`
   - Path depth: 3-4 steps
   - Filters: `platform = mobile`
   - Time range: Last 7 days
3. Name: "Common Navigation Paths"
4. Add to dashboard

**Dashboard 2 Complete** ✅

---

## Dashboard 3: Mobile User Funnels (60 min)

### Step 1: Create Dashboard
1. Go to PostHog → Dashboards → New Dashboard
2. Name: "Mobile User Funnels"
3. Description: "User journey conversion analysis"
4. Tags: `mobile`, `conversion`, `funnel`

### Step 2: Onboarding Funnel
1. Click "Add insight" → New insight
2. Select "Funnels" visualization
3. Configure steps:
   - Step 1: `signup_started`
   - Step 2: `signup_completed`
   - Step 3: `login_completed`
   - Step 4: `submission_viewed`
4. Settings:
   - Conversion window: 24 hours
   - Filters: `platform = mobile`
   - Breakdown: `device_os`
   - Time range: Last 30 days
5. Name: "New User Onboarding Funnel"
6. Add to dashboard

### Step 3: Content Engagement Funnel
1. Click "Add insight" → New insight
2. Select "Funnels" visualization
3. Configure steps:
   - Step 1: `submission_viewed`
   - Step 2: `submission_play_started`
   - Step 3: `upvote_clicked` OR `comment_create_started` OR `follow_clicked`
4. Settings:
   - Conversion window: 1 hour
   - Filters: `platform = mobile`
   - Breakdown: `screen_name` (optional)
   - Time range: Last 7 days
5. Name: "Content Engagement Funnel"
6. Add to dashboard

### Step 4: Submission Creation Funnel
1. Click "Add insight" → New insight
2. Select "Funnels" visualization
3. Configure steps:
   - Step 1: `submission_create_started`
   - Step 2: `submission_create_completed`
   - Step 3: `submission_shared` OR `submission_viewed`
4. Settings:
   - Conversion window: 30 minutes
   - Filters: `platform = mobile`
   - Breakdown: `app_version` (optional)
   - Time range: Last 30 days
5. Name: "Submission Creation Funnel"
6. Add to dashboard

### Step 5: Add Funnel Summary
1. Add text panel at top with:
   - Current conversion rates
   - Improvement targets
   - Key drop-off points to watch

**Dashboard 3 Complete** ✅

---

## Dashboard 4: Mobile Retention Analysis (90 min)

### Step 1: Create Dashboard
1. Go to PostHog → Dashboards → New Dashboard
2. Name: "Mobile Retention Analysis"
3. Description: "User retention cohort tracking"
4. Tags: `mobile`, `retention`, `cohorts`

### Step 2: Weekly Retention
1. Click "Add insight" → New insight
2. Select "Retention" visualization
3. Configuration:
   - Cohort action: `signup_completed` OR `login_completed`
   - Return action: `screen_viewed`
   - Cohort size: Weekly
   - Retention period: 8 weeks
   - Filters: `platform = mobile`
4. Name: "Weekly User Retention"
5. Add to dashboard

### Step 3: Monthly Retention
1. Click "Add insight" → New insight
2. Select "Retention" visualization
3. Configuration:
   - Cohort action: `signup_completed`
   - Return action: Any engagement event (use multiple OR conditions)
   - Cohort size: Monthly
   - Retention period: 6 months
   - Filters: `platform = mobile`
4. Name: "Monthly User Retention"
5. Add to dashboard

### Step 4: Feature-Specific Retention
1. Click "Add insight" → New insight
2. Select "Retention" visualization
3. Configuration:
   - Cohort action: `submission_create_completed`
   - Return action: `submission_create_started`
   - Cohort size: Weekly
   - Filters: `platform = mobile`
4. Name: "Content Creator Retention"
5. Add to dashboard

### Step 5: Retention Summary Cards
1. Add 4 "Big Number" insights:
   - Day 1 retention %
   - Day 7 retention %
   - Day 30 retention %
   - Current month retention vs last month
2. Position at top of dashboard

**Dashboard 4 Complete** ✅

---

## Dashboard 5: Mobile Stability (60 min)

### Step 1: Create Dashboard
1. Go to PostHog → Dashboards → New Dashboard
2. Name: "Mobile Stability"
3. Description: "Crash-free sessions and error monitoring"
4. Tags: `mobile`, `stability`, `errors`

### Step 2: Crash-Free Session Rate
1. Click "Add insight" → New insight
2. Configuration:
   - Visualization: Number (with trend)
   - Formula (if supported): 
     ```
     (count(distinct $session_id) - count(distinct $session_id where error_occurred)) 
     / count(distinct $session_id) * 100
     ```
   - Or use two separate counts and calculate manually
   - Filters: `platform = mobile`
   - Time range: Last 7 days
3. Name: "Crash-Free Session Rate %"
4. Add alert: Alert when < 99%
5. Add to dashboard

### Step 3: Error Rate Trends
1. Click "Add insight" → New insight
2. Configuration:
   - Visualization: Line graph
   - Event: `error_occurred`
   - Breakdown: `error_type`
   - Filters: `platform = mobile`
   - Time range: Last 30 days
   - Interval: Daily
3. Name: "Error Rate by Type"
4. Add to dashboard

### Step 4: Errors by Screen
1. Click "Add insight" → New insight
2. Configuration:
   - Visualization: Table
   - Event: `error_occurred`
   - Rows: Group by `screen_name`
   - Columns: Breakdown by `error_type`
   - Filters: `platform = mobile`
   - Time range: Last 7 days
3. Name: "Errors by Screen and Type"
4. Add to dashboard

### Step 5: Critical Errors Table
1. Click "Add insight" → New insight
2. Configuration:
   - Visualization: Table
   - Event: `error_occurred`
   - Columns: 
     - `error_message`
     - Count
     - Unique users affected
   - Filters: `platform = mobile`, `error_type IN ['ApiError', 'NetworkError', 'VideoPlaybackError']`
   - Time range: Last 7 days
   - Sort: By count descending
3. Name: "Top Critical Errors"
4. Add to dashboard

### Step 6: Stability Summary
1. Add 3 "Big Number" insights at top:
   - Crash-free rate % (last 7d)
   - Total errors (last 24h)
   - Error-free users % (last 7d)

**Dashboard 5 Complete** ✅

---

## Post-Setup Tasks

### 1. Configure Alerts (15 min)

Set up alerts for critical metrics:

1. **DAU Drop Alert**:
   - Go to Dashboard → Mobile DAU/MAU
   - Click on DAU insight → Create alert
   - Condition: Decreases by >20% day-over-day
   - Recipients: Product team, engineering lead

2. **Crash Rate Alert**:
   - Go to Dashboard → Mobile Stability
   - Click on Crash-free rate insight → Create alert
   - Condition: Drops below 99%
   - Recipients: Mobile team, on-call engineer

3. **Error Spike Alert**:
   - Go to Dashboard → Mobile Stability
   - Click on Error rate insight → Create alert
   - Condition: Increases by >100% hour-over-hour
   - Recipients: Mobile team

### 2. Share Dashboards (10 min)

1. For each dashboard:
   - Click "Share" button
   - Add team members with appropriate permissions
   - Generate shareable link for stakeholders
   - Add to team wiki/documentation

2. Create dashboard collection:
   - Go to PostHog → Dashboards
   - Create tag "Mobile Analytics"
   - Tag all 5 dashboards
   - Makes them easy to find

### 3. Set Refresh Rates (5 min)

Configure automatic refresh:

1. **DAU/MAU**: Refresh every 1 hour
2. **Screen Views**: Refresh every 30 minutes
3. **Funnels**: Refresh every 1 hour
4. **Retention**: Refresh every 4 hours (less frequently needed)
5. **Stability**: Refresh every 15 minutes (critical monitoring)

### 4. Schedule Reports (10 min)

Set up automated email reports:

1. **Daily Stability Report**:
   - Dashboard: Mobile Stability
   - Frequency: Daily at 9 AM
   - Recipients: Mobile team

2. **Weekly Metrics Summary**:
   - Dashboard: Mobile DAU/MAU
   - Frequency: Monday 10 AM
   - Recipients: Product team, executives

3. **Monthly Deep Dive**:
   - Dashboard: All dashboards (combined)
   - Frequency: First of month
   - Recipients: All stakeholders

### 5. Documentation (15 min)

1. Add dashboard links to internal wiki
2. Update team onboarding docs
3. Create quick reference card with:
   - Dashboard names and purposes
   - Key metrics and their targets
   - How to access dashboards
   - Who to contact for questions

---

## Validation Checklist

Before considering setup complete, verify:

### Data Quality
- [ ] All insights showing data (not empty)
- [ ] Event counts seem reasonable
- [ ] No unusual spikes or drops (unless expected)
- [ ] Breakdown dimensions working correctly
- [ ] Filters applied properly

### Dashboard Functionality
- [ ] All insights load within 5 seconds
- [ ] Refresh works correctly
- [ ] Time range selectors work
- [ ] Breakdowns display properly
- [ ] Mobile view is readable

### Team Access
- [ ] All team members can access dashboards
- [ ] Permissions set correctly (view vs edit)
- [ ] Dashboard links work
- [ ] Alerts configured and tested

### Integration
- [ ] Sentry links work (if applicable)
- [ ] Slack notifications working
- [ ] Email reports being received
- [ ] Data matches expectations

---

## Troubleshooting

### No Data Showing

**Check**:
1. PostHog SDK initialized correctly in mobile app
2. Events being sent (check PostHog Live Events)
3. Filters not too restrictive (try removing `platform = mobile` temporarily)
4. Time range includes data
5. User granted analytics consent

**Solution**:
- Review POSTHOG_ANALYTICS.md integration guide
- Check mobile app logs for analytics errors
- Verify PostHog API key configured

### Slow Dashboard Loading

**Optimize**:
1. Reduce time range (90 days → 30 days)
2. Limit breakdowns (top 10 instead of all)
3. Use sampling for exploratory analysis
4. Simplify complex formulas
5. Cache results when possible

### Incorrect Metrics

**Verify**:
1. Event names match exactly (case-sensitive)
2. Property names correct
3. Aggregation function appropriate
4. Filters working as expected
5. No duplicate event sends

**Debug**:
- Check event examples in PostHog
- Compare with raw event data
- Verify calculation formulas
- Review recent SDK changes

---

## Next Steps

After setup is complete:

1. **Week 1**: Monitor daily, familiarize team with dashboards
2. **Week 2**: Establish baseline metrics, set targets
3. **Week 3**: Identify optimization opportunities
4. **Week 4**: Review and iterate on dashboards
5. **Monthly**: Review retention, adjust as needed

## Support

- **Technical Questions**: See [analytics-dashboards.md](./analytics-dashboards.md)
- **Query Help**: See [posthog-dashboard-queries.md](./posthog-dashboard-queries.md)
- **PostHog Docs**: https://posthog.com/docs
- **Team Contact**: [Your team channel/email]

---

**Estimated Total Time**: 4.5 hours + 1 hour for post-setup = **5.5 hours**

**Success Criteria**: All 5 dashboards created, data flowing, team has access, alerts configured.
