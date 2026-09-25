---
title: "Analytics Dashboards"
summary: "This document provides comprehensive setup instructions for PostHog dashboards to visualize mobile app metrics for the Clipper mobile application."
tags: ["mobile"]
area: "mobile"
status: "stable"
owner: "team-core"
version: "1.0"
last_reviewed: 2026-01-29
---

# Mobile Analytics Dashboards

This document provides comprehensive setup instructions for PostHog dashboards to visualize mobile app metrics for the Clipper mobile application.

## Overview

The mobile analytics dashboards provide insights into user behavior, engagement, retention, and app stability. These dashboards are built on top of the PostHog SDK integration documented in mobile/POSTHOG_ANALYTICS.md.

## Dashboard Suite

We maintain five core dashboards for mobile analytics:

1. **Mobile User Funnels** - User journey conversion analysis
2. **Mobile Retention Analysis** - Cohort retention tracking
3. **Mobile Screen Views** - Navigation patterns and screen engagement
4. **Mobile Stability** - Crash-free sessions and error tracking
5. **Mobile DAU/MAU** - Daily and monthly active user metrics

## Prerequisites

- PostHog project configured with API key
- Mobile app deployed with PostHog SDK integration
- Access to PostHog dashboard (app.posthog.com or self-hosted instance)
- Admin or team member access to create dashboards

## Event Schema Reference

All dashboards use the standardized mobile event schema defined in `mobile/lib/analytics.ts`:

### Core Event Categories

- **AuthEvents**: `signup_started`, `signup_completed`, `login_completed`, `logout`, `oauth_callback`
- **SubmissionEvents**: `submission_viewed`, `submission_create_started`, `submission_create_completed`, `submission_play_started`
- **EngagementEvents**: `upvote_clicked`, `downvote_clicked`, `comment_create_completed`, `follow_clicked`, `search_performed`
- **NavigationEvents**: `screen_viewed`, `page_viewed`, `tab_clicked`
- **ErrorEvents**: `error_occurred`, `api_error`, `network_error`, `video_playback_error`
- **PerformanceEvents**: `page_load_time`, `api_response_time`, `video_load_time`

### Common Event Properties

- `user_id`: Unique user identifier
- `screen_name`: Current screen name
- `timestamp`: Event timestamp (ISO 8601)
- `app_version`: Mobile app version
- `device_os`: Operating system (iOS/Android)
- `device_model`: Device model name
- `platform`: Always "mobile" for mobile events

## Dashboard Setup Instructions

### Dashboard 1: Mobile User Funnels

**Purpose**: Track user conversion through key journeys: signup → first clip play → engagement actions.

**Setup Steps**:

1. Navigate to PostHog → Insights → New Insight
2. Select "Funnel" visualization type
3. Create the following funnels:

#### Funnel A: New User Onboarding
```
Step 1: signup_started
Step 2: signup_completed
Step 3: login_completed
Step 4: submission_viewed (first clip view)
```

**Filters**:
- Time range: Last 30 days
- Platform: `platform = mobile`
- Breakdown by: `device_os` (iOS vs Android)

**Insight Configuration**:
- Conversion window: 24 hours
- Visualization: Funnel chart
- Show conversion rate percentages

#### Funnel B: Content Engagement Journey
```
Step 1: submission_viewed
Step 2: submission_play_started
Step 3: upvote_clicked OR comment_create_started OR follow_clicked
```

**Filters**:
- Time range: Last 7 days
- Platform: `platform = mobile`
- User property: `is_authenticated = true`

#### Funnel C: Submission Flow
```
Step 1: submission_create_started
Step 2: submission_create_completed
Step 3: submission_shared OR submission_viewed (own submission)
```

**Filters**:
- Time range: Last 30 days
- Platform: `platform = mobile`
- Breakdown by: `app_version`

4. Save all funnels to a dashboard named "Mobile User Funnels"
5. Add text panel with dashboard description and key metrics

**Key Metrics to Monitor**:
- Overall funnel conversion rate (target: >40% for onboarding)
- Drop-off points (identify friction)
- Platform comparison (iOS vs Android)
- Version-over-version improvements

---

### Dashboard 2: Mobile Retention Analysis

**Purpose**: Measure user retention over time with weekly and monthly cohorts.

**Setup Steps**:

1. Navigate to PostHog → Insights → New Insight
2. Select "Retention" visualization type

#### Cohort A: Weekly Retention
**Configuration**:
- Cohort action: `signup_completed` OR `login_completed`
- Return action: `screen_viewed` (any screen view indicates active use)
- Cohort size: Weekly
- Retention period: 8 weeks
- Filters: `platform = mobile`

**Visualization**:
- Display as: Retention table with percentages
- Show: Day 1, Day 7, Day 14, Day 30 retention rates
- Color code: Green (>40%), Yellow (20-40%), Red (<20%)

#### Cohort B: Monthly Retention
**Configuration**:
- Cohort action: `signup_completed`
- Return action: Any engagement event (`upvote_clicked`, `comment_create_completed`, `submission_create_started`)
- Cohort size: Monthly
- Retention period: 6 months
- Filters: `platform = mobile`

**Visualization**:
- Display as: Retention curve chart
- Highlight: Month 1, Month 3, Month 6 retention

#### Additional Retention Insights

**Power User Retention**:
- Cohort action: Users with `>=10` events in first week
- Return action: `>=10` events in subsequent weeks
- Filters: `platform = mobile`

**Feature-Specific Retention**:
- Cohort action: `submission_create_completed`
- Return action: `submission_create_started` (returning submitters)
- Cohort size: Weekly
- Filters: `platform = mobile`

3. Save all retention insights to "Mobile Retention Analysis" dashboard
4. Add trend cards showing:
   - Current D1 retention rate
   - Current D7 retention rate
   - Current D30 retention rate
   - Week-over-week retention change

**Target Retention Rates**:
- Day 1: >40%
- Day 7: >25%
- Day 30: >15%
- Month 3: >10%

---

### Dashboard 3: Mobile Screen Views

**Purpose**: Understand navigation patterns, popular screens, and time spent per screen.

**Setup Steps**:

1. Navigate to PostHog → Insights → New Insight
2. Create the following insights:

#### Insight A: Top Screens by Views
**Configuration**:
- Visualization: Bar chart
- Event: `screen_viewed`
- Breakdown by: `screen_name`
- Filters: `platform = mobile`
- Time range: Last 7 days
- Display: Top 20 screens

**Calculated Metrics**:
- Total views per screen
- Unique users per screen
- Views per user ratio

#### Insight B: Screen View Trends
**Configuration**:
- Visualization: Line graph
- Event: `screen_viewed`
- Breakdown by: `screen_name` (top 10 only)
- Filters: `platform = mobile`
- Time range: Last 30 days
- Interval: Daily

#### Insight C: Screen Engagement Time
**Configuration**:
- Visualization: Table
- Event: `screen_viewed`
- Aggregation: Average of `time_on_screen` property
- Breakdown by: `screen_name`
- Filters: `platform = mobile`, `time_on_screen > 0`
- Time range: Last 7 days
- Sort by: Average time descending

**Important Screens to Track**:
- `HomeScreen` / Feed
- `SubmissionDetailsScreen`
- `SubmitScreen`
- `ProfileScreen`
- `SearchScreen`
- `SettingsScreen`
- `LoginScreen`
- `SignupScreen`

#### Insight D: Navigation Flow (Sankey Diagram)
**Configuration**:
- Visualization: Path analysis / Sankey
- Start event: `screen_viewed`
- End event: `screen_viewed`
- Filters: `platform = mobile`
- Time range: Last 7 days
- Path depth: 3-4 screens
- Minimum path occurrence: 10 users

This shows common navigation patterns like:
```
HomeScreen → SubmissionDetails → Comments
HomeScreen → Search → SubmissionDetails
ProfileScreen → SubmissionDetails → Share
```

#### Insight E: Screen Exit Points
**Configuration**:
- Event: `screen_viewed` where next event is app background/close
- Breakdown by: `screen_name`
- Shows which screens users leave from most often

3. Save all insights to "Mobile Screen Views" dashboard
4. Add summary cards:
   - Total screens viewed (last 7d)
   - Average screens per session
   - Most popular screen
   - Highest engagement time screen

**Key Metrics**:
- Screens per session (target: >5)
- Average time per screen (target: >10 seconds)
- Home screen return rate
- Search usage rate

---

### Dashboard 4: Mobile Stability

**Purpose**: Monitor app stability with crash-free session rates and error tracking, integrated with Sentry signals.

**Setup Steps**:

1. Navigate to PostHog → Insights → New Insight
2. Create the following insights:

#### Insight A: Crash-Free Session Rate
**Configuration**:
- Visualization: Gauge / Percentage
- Formula: `(total_sessions - sessions_with_errors) / total_sessions * 100`
- Where:
  - `total_sessions` = Count of distinct sessions (use `$session_id`)
  - `sessions_with_errors` = Count of distinct sessions with `error_occurred` event
- Filters: `platform = mobile`
- Time range: Last 7 days
- Target: >99.5% (critical threshold: <99%)

**Color Thresholds**:
- Green: ≥99.5%
- Yellow: 99.0-99.5%
- Red: <99.0%

#### Insight B: Error Rate Trends
**Configuration**:
- Visualization: Line graph
- Event: `error_occurred`
- Metric: Count
- Breakdown by: `error_type`
- Filters: `platform = mobile`
- Time range: Last 30 days
- Show: Daily granularity

**Track Error Types**:
- `NetworkError`: Network/API failures
- `VideoPlaybackError`: Video player issues
- `FormValidationError`: Input validation issues
- `ApiError`: Backend API errors
- `ImageLoadError`: Image loading failures

#### Insight C: Errors by Screen
**Configuration**:
- Visualization: Heatmap or table
- Event: `error_occurred`
- Row dimension: `screen_name`
- Column dimension: `error_type`
- Cell value: Count of errors
- Filters: `platform = mobile`
- Time range: Last 7 days

#### Insight D: Critical Errors
**Configuration**:
- Visualization: Table
- Event: `error_occurred`
- Filters: 
  - `platform = mobile`
  - `error_type IN ('ApiError', 'NetworkError', 'VideoPlaybackError')`
- Columns:
  - Error message
  - Count
  - Affected users
  - First seen
  - Last seen
- Sort by: Count descending
- Time range: Last 7 days

#### Insight E: Error Recovery Rate
**Configuration**:
- Visualization: Percentage
- Calculation: Users who encountered error but continued using app
- Events sequence:
  1. `error_occurred`
  2. Followed by any engagement event within 5 minutes
- Metric: Percentage of error events followed by recovery
- Filters: `platform = mobile`
- Time range: Last 7 days

#### Insight F: Sentry Integration Metrics
**Configuration**:
- Visualization: Stat cards
- Metrics:
  - Total Sentry issues (link to Sentry dashboard)
  - Unresolved critical issues
  - Average time to resolution
  - Crash rate by app version
- Data source: Combined PostHog + Sentry (via Sentry integration or manual sync)

**Note**: Sentry-PostHog integration can be configured via:
- PostHog Data Pipeline for Sentry events
- Sentry webhook → PostHog ingestion
- Manual correlation using `sentry_event_id` property

3. Save all insights to "Mobile Stability" dashboard
4. Add alert panels:
   - Alert when crash-free rate drops below 99%
   - Alert when error rate increases >50% week-over-week
   - Alert for new critical error types

**Key Stability Metrics**:
- Crash-free session rate (target: >99.5%)
- Error rate per session (target: <0.1 errors/session)
- Mean time between failures (MTBF)
- Error resolution time (target: <24 hours for critical)

---

### Dashboard 5: Mobile DAU/MAU

**Purpose**: Track daily and monthly active users with trends and growth metrics.

**Setup Steps**:

1. Navigate to PostHog → Insights → New Insight
2. Create the following insights:

#### Insight A: DAU Trend
**Configuration**:
- Visualization: Line graph
- Formula: Count distinct `user_id` where any event occurred
- Alternative: Use PostHog's built-in "Active Users" insight with Daily interval
- Filters: `platform = mobile`
- Time range: Last 90 days
- Granularity: Daily

**Enhancements**:
- Add trend line (7-day moving average)
- Show week-over-week comparison
- Highlight weekends vs weekdays

#### Insight B: MAU Trend
**Configuration**:
- Visualization: Line graph with area fill
- Formula: Count distinct `user_id` in rolling 30-day window
- Alternative: Use PostHog's built-in "Active Users" insight with Monthly interval
- Filters: `platform = mobile`
- Time range: Last 12 months
- Granularity: Weekly or monthly

#### Insight C: WAU (Weekly Active Users)
**Configuration**:
- Visualization: Line graph
- Formula: Count distinct `user_id` in rolling 7-day window
- Filters: `platform = mobile`
- Time range: Last 90 days
- Granularity: Weekly

#### Insight D: Stickiness Ratio (DAU/MAU)
**Configuration**:
- Visualization: Gauge / Percentage
- Formula: `(DAU / MAU) * 100`
- Filters: `platform = mobile`
- Time range: Current month
- Benchmark: Industry standard is 20-25% for social/content apps

**Color Thresholds**:
- Green: ≥25%
- Yellow: 15-25%
- Red: <15%

#### Insight E: DAU Breakdown
**Configuration**:
- Visualization: Stacked area chart
- Metric: Daily active users
- Breakdown by: `device_os` (iOS vs Android)
- Filters: `platform = mobile`
- Time range: Last 30 days
- Shows platform-specific growth

#### Insight F: New vs Returning Users
**Configuration**:
- Visualization: Stacked bar chart
- Daily data showing:
  - New users (first `signup_completed` or first app open)
  - Returning users (users active but not new)
- Filters: `platform = mobile`
- Time range: Last 30 days

#### Insight G: User Growth Rate
**Configuration**:
- Visualization: Stat card with sparkline
- Calculation: `((Current Month MAU - Previous Month MAU) / Previous Month MAU) * 100`
- Shows month-over-month growth percentage
- Time range: Last 6 months
- Display: Percentage with up/down indicator

#### Insight H: Cohort Size Evolution
**Configuration**:
- Visualization: Line graph (multiple lines)
- Shows MAU contribution from signup cohorts
- Lines for:
  - Users who signed up 0-1 months ago (recent)
  - Users who signed up 1-3 months ago
  - Users who signed up 3-6 months ago
  - Users who signed up 6+ months ago
- Filters: `platform = mobile`
- Time range: Last 6 months

#### Insight I: Activity Distribution
**Configuration**:
- Visualization: Histogram
- Metric: Number of days active per user in last 30 days
- Buckets: 1, 2-5, 6-10, 11-20, 21-30 days
- Shows distribution of user engagement levels
- Filters: `platform = mobile`

#### Insight J: Peak Usage Hours
**Configuration**:
- Visualization: Heatmap
- X-axis: Hour of day (0-23)
- Y-axis: Day of week (Mon-Sun)
- Cell value: Average DAU during that hour/day
- Filters: `platform = mobile`
- Time range: Last 30 days
- Timezone: User's local timezone

3. Save all insights to "Mobile DAU/MAU" dashboard
4. Add summary cards at the top:
   - Current DAU (with % change vs yesterday)
   - Current MAU (with % change vs last month)
   - Current Stickiness Ratio
   - Total registered mobile users
   - New users this month
   - Average session length

**Key Growth Metrics**:
- DAU growth (target: >10% month-over-month)
- MAU growth (target: >15% month-over-month)
- Stickiness ratio (target: >20%)
- New user acquisition rate
- Activation rate (new users who complete key action)

---

## Dashboard Organization

### Recommended Layout

```
PostHog Projects
└── Clipper Mobile
    └── Dashboards
        ├── 📊 Mobile User Funnels
        ├── 📈 Mobile Retention Analysis
        ├── 🧭 Mobile Screen Views
        ├── 🛡️ Mobile Stability
        └── 👥 Mobile DAU/MAU
```

### Dashboard Naming Convention

- Prefix all mobile dashboards with "Mobile" for easy filtering
- Use emojis for visual identification (optional)
- Keep names concise but descriptive
- Tag all dashboards with: `mobile`, `analytics`, `production`

### Access Control

**Recommended Access Levels**:
- **View**: All team members
- **Edit**: Product managers, analytics team, engineering leads
- **Admin**: Product owner, analytics lead

### Sharing

1. Generate shareable links for stakeholder reviews
2. Set up Slack notifications for key metric changes
3. Schedule weekly dashboard email digests
4. Embed dashboards in internal tools (e.g., Notion, Confluence)

---

## Data Validation & Testing

### Testing with Staging Environment

Before relying on production data, validate dashboards in staging:

1. **Generate Test Events**:
   ```bash
   # Use mobile app in staging with PostHog configured
   # Perform typical user flows:
   - Sign up new account
   - Browse feed
   - Play clips
   - Submit content
   - Navigate between screens
   - Trigger intentional errors
   ```

2. **Verify Event Capture**:
   - Navigate to PostHog → Live Events
   - Filter by `platform = mobile`
   - Confirm events appear with correct properties
   - Check for missing or malformed data

3. **Cross-Check Counts**:
   - Compare event counts with app usage logs
   - Verify user counts match authentication logs
   - Check session counts align with app analytics
   - Validate timestamp accuracy

4. **Dashboard Refresh**:
   - Refresh dashboards after test events
   - Verify new data appears correctly
   - Check filters are working
   - Confirm breakdowns display properly

### Data Freshness Validation

- PostHog typically has <5 minute data delay
- Check "Last updated" timestamp on insights
- If data seems stale, check PostHog ingestion status
- Verify mobile app is sending events (check PostHog debugger)

### Common Data Quality Issues

**Missing Events**:
- Symptom: Events not appearing in PostHog
- Causes: Network issues, consent not granted, SDK initialization failed
- Solution: Check app logs, verify API key, test in PostHog debugger

**Incorrect Event Properties**:
- Symptom: Properties null or wrong type
- Causes: Code bug, schema mismatch, version differences
- Solution: Review analytics.ts, check event tracking calls

**Duplicate Events**:
- Symptom: Inflated event counts
- Causes: Double initialization, retry logic, dev/staging mixed
- Solution: Add distinct ID checks, separate environments

**Session Misattribution**:
- Symptom: Sessions incorrectly counted
- Causes: Session timeout issues, app backgrounding
- Solution: Review session management, adjust timeout

---

## Maintenance & Ownership

### Dashboard Ownership

| Dashboard | Primary Owner | Secondary Owner | Update Frequency |
|-----------|---------------|-----------------|------------------|
| Mobile User Funnels | Product Manager | Mobile Lead | Weekly |
| Mobile Retention | Product Manager | Growth Lead | Weekly |
| Mobile Screen Views | UX Designer | Mobile Lead | Bi-weekly |
| Mobile Stability | Mobile Lead | DevOps | Daily |
| Mobile DAU/MAU | Growth Lead | Product Manager | Daily |

### Regular Review Schedule

- **Daily**: Stability dashboard for crash monitoring
- **Weekly**: DAU/MAU trends, funnel performance
- **Monthly**: Retention analysis, comprehensive review
- **Quarterly**: Dashboard optimization, new metrics addition

### Dashboard Evolution

As the product evolves, update dashboards to reflect:
- New features and their tracking
- Deprecated screens or flows
- Changed event schemas
- New business metrics
- A/B test results

**Change Log**:
Maintain a dashboard changelog in this document:
- Date: [YYYY-MM-DD]
- Dashboard: [Name]
- Change: [Description]
- Reason: [Why it was changed]

---

## Alerting & Monitoring

### Critical Alerts to Configure

1. **Crash-Free Rate Alert**:
   - Metric: Crash-free session rate
   - Threshold: <99%
   - Recipients: Mobile team, on-call engineer
   - Action: Immediate investigation

2. **DAU Drop Alert**:
   - Metric: DAU
   - Threshold: >20% drop day-over-day (excluding weekends)
   - Recipients: Product manager, engineering lead
   - Action: Check for service issues, investigate

3. **Error Spike Alert**:
   - Metric: Error rate
   - Threshold: >2x normal rate
   - Recipients: Mobile team
   - Action: Check recent deployments, review errors

4. **Funnel Drop Alert**:
   - Metric: Key funnel conversion rate
   - Threshold: <10% drop week-over-week
   - Recipients: Product manager, growth team
   - Action: Analyze funnel steps, check for UX issues

### PostHog Alerting Setup

PostHog supports alerts via:
- Slack webhooks
- Email notifications
- Webhooks to custom services
- Integration with PagerDuty/Opsgenie

Configure in PostHog → Project Settings → Alerts

---

## Integration with Other Tools

### Sentry Integration

Correlate PostHog events with Sentry errors:

1. Add `sentry_event_id` to error tracking events
2. Link from PostHog dashboard to Sentry issues
3. Create combined view for stability monitoring

**Code Example**:
```typescript
import * as Sentry from '@sentry/react-native';
import { trackError } from '@/lib/analytics';

try {
  // Some code that might fail
} catch (error) {
  const sentryEventId = Sentry.captureException(error);
  trackError(error as Error, {
    errorType: 'UnexpectedError',
    sentry_event_id: sentryEventId,
  });
}
```

### Grafana Integration

For unified monitoring:
- Create Grafana dashboard that pulls PostHog metrics via API
- Display mobile metrics alongside backend metrics
- Correlate mobile errors with API performance

### Slack Integration

Set up Slack notifications:
- Daily digest of key metrics
- Alert notifications for critical thresholds
- Weekly dashboard screenshot shares

---

## Troubleshooting

### Dashboard Not Showing Data

**Check**:
1. ✅ PostHog API key configured correctly in mobile app
2. ✅ User has granted analytics consent
3. ✅ Events are being sent (check PostHog Live Events)
4. ✅ Time range includes data
5. ✅ Filters not overly restrictive

### Metrics Seem Wrong

**Verify**:
1. Event schema matches expectations (check event examples)
2. Property names and types are correct
3. No duplicate event sends
4. Correct calculation formulas
5. Timezone settings

### Dashboard Performance Issues

**Optimize**:
1. Reduce time range for complex queries
2. Use sampling for large datasets
3. Cache frequently accessed insights
4. Simplify breakdowns
5. Use PostHog recording rules for pre-aggregation

### Missing Events

**Debug**:
1. Check PostHog debugger in mobile app (enable in dev mode)
2. Review app logs for analytics errors
3. Verify network connectivity
4. Check PostHog service status
5. Review recent SDK version changes

---

## Resources

### Documentation Links

- PostHog Analytics Integration
- [Mobile Architecture](../ARCHITECTURE.md)
- Mobile Analytics Event Schema
- Sentry Integration

### PostHog Resources

- [PostHog Documentation](https://posthog.com/docs)
- [PostHog Insights Guide](https://posthog.com/docs/user-guides/insights)
- [PostHog Dashboards](https://posthog.com/docs/user-guides/dashboards)
- [PostHog Funnels](https://posthog.com/docs/user-guides/funnels)
- [PostHog Retention](https://posthog.com/docs/user-guides/retention)

### Team Contacts

- **Product Owner**: [Name] - Questions about metrics and KPIs
- **Mobile Lead**: [Name] - Technical implementation questions
- **Analytics Lead**: [Name] - Dashboard setup and optimization
- **DevOps**: [Name] - Infrastructure and integration issues

---

## Changelog

### 2026-01-01
- Initial dashboard suite documentation created
- Five core dashboards defined with setup instructions
- Event schema reference added
- Testing and validation guidelines documented
- Ownership and maintenance procedures established

---

**Document Version**: 1.0  
**Last Updated**: 2026-01-01  
**Maintained By**: Mobile Analytics Team  
**Review Cycle**: Monthly
