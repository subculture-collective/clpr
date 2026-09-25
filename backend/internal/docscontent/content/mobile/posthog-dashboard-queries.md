---
title: "Posthog Dashboard Queries"
summary: "Quick reference for PostHog query configurations used in mobile analytics dashboards."
tags: ["mobile"]
area: "mobile"
status: "stable"
owner: "team-core"
version: "1.0"
last_reviewed: 2026-01-29
---

# PostHog Dashboard Queries Reference

Quick reference for PostHog query configurations used in mobile analytics dashboards.

## Query Patterns

### Active Users Queries

#### Daily Active Users (DAU)
```
Event: Any event
Aggregation: Unique users
Filters: platform = mobile
Time Range: Last 90 days
Interval: Daily
```

#### Monthly Active Users (MAU)
```
Event: Any event
Aggregation: Unique users
Filters: platform = mobile
Time Range: Last 12 months
Interval: Monthly (rolling 30-day window)
```

#### Stickiness Ratio
```
Formula: (DAU / MAU) * 100
Where:
  DAU = Count distinct users (daily)
  MAU = Count distinct users (30-day rolling)
Filters: platform = mobile
Display: Percentage
```

### Funnel Queries

#### Signup Funnel
```
Funnel Steps:
  1. Event: signup_started
  2. Event: signup_completed
  3. Event: login_completed
  4. Event: submission_viewed
  
Conversion Window: 24 hours
Breakdown By: device_os
Filters: platform = mobile
Time Range: Last 30 days
```

#### Content Engagement Funnel
```
Funnel Steps:
  1. Event: submission_viewed
  2. Event: submission_play_started
  3. Event: upvote_clicked OR comment_create_started OR follow_clicked
  
Conversion Window: 1 hour
Breakdown By: screen_name
Filters: platform = mobile, is_authenticated = true
Time Range: Last 7 days
```

### Retention Queries

#### Weekly Cohort Retention
```
Cohort Action: signup_completed OR login_completed
Return Action: screen_viewed
Cohort Size: Weekly
Retention Period: 8 weeks
Filters: platform = mobile
Display: Retention table (Day 1, Day 7, Day 14, Day 30)
```

#### Monthly Cohort Retention
```
Cohort Action: signup_completed
Return Action: upvote_clicked OR comment_create_completed OR submission_create_started
Cohort Size: Monthly
Retention Period: 6 months
Filters: platform = mobile
Display: Retention curve
```

### Screen View Queries

#### Top Screens
```
Event: screen_viewed
Aggregation: Total count
Breakdown By: screen_name
Filters: platform = mobile
Time Range: Last 7 days
Display: Bar chart (Top 20)
Order By: Count descending
```

#### Screen Engagement Time
```
Event: screen_viewed
Aggregation: Average of time_on_screen property
Breakdown By: screen_name
Filters: platform = mobile, time_on_screen > 0
Time Range: Last 7 days
Display: Table
Order By: Average time descending
```

#### Navigation Flow
```
Insight Type: Path Analysis
Start Event: screen_viewed
End Event: screen_viewed
Path Depth: 3-4 steps
Filters: platform = mobile
Time Range: Last 7 days
Minimum Path Frequency: 10 users
Display: Sankey diagram
```

### Stability Queries

#### Crash-Free Session Rate
```
Formula: ((total_sessions - sessions_with_errors) / total_sessions) * 100

Where:
  total_sessions = Count distinct $session_id
  sessions_with_errors = Count distinct $session_id where error_occurred event exists

Filters: platform = mobile
Time Range: Last 7 days
Display: Gauge with percentage
Thresholds:
  - Green: >=99.5%
  - Yellow: 99.0-99.5%
  - Red: <99.0%
```

#### Error Rate by Type
```
Event: error_occurred
Aggregation: Count
Breakdown By: error_type
Filters: platform = mobile
Time Range: Last 30 days
Display: Line graph (daily)
```

#### Errors by Screen
```
Event: error_occurred
Aggregation: Count
Row Dimension: screen_name
Column Dimension: error_type
Filters: platform = mobile
Time Range: Last 7 days
Display: Heatmap or table
```

### Performance Queries

#### Average Page Load Time
```
Event: page_load_time
Aggregation: Average of duration property
Breakdown By: screen_name
Filters: platform = mobile, duration > 0
Time Range: Last 7 days
Display: Bar chart
Unit: Milliseconds
```

#### API Response Time (P95)
```
Event: api_response_time
Aggregation: 95th percentile of duration property
Breakdown By: endpoint property
Filters: platform = mobile
Time Range: Last 7 days
Display: Line graph
Unit: Milliseconds
Threshold: 1000ms (warning)
```

### User Segmentation Queries

#### New vs Returning Users
```
Daily Breakdown:
  - New Users: user_id with first event = today
  - Returning Users: user_id with first event < today, active today
  
Filters: platform = mobile
Time Range: Last 30 days
Display: Stacked bar chart
```

#### User Activity Distribution
```
Calculation: Days active per user in last 30 days
Buckets:
  - 1 day
  - 2-5 days
  - 6-10 days
  - 11-20 days
  - 21-30 days
  
Filters: platform = mobile
Display: Histogram
Metric: Count of users per bucket
```

#### Platform Distribution
```
Event: Any event
Aggregation: Unique users
Breakdown By: device_os
Filters: platform = mobile
Time Range: Last 30 days
Display: Pie chart or bar chart
```

## Property Filters

### Common Filters

```
Platform Filter:
  platform = mobile

Authenticated Users:
  is_authenticated = true

iOS Only:
  platform = mobile AND device_os = iOS

Android Only:
  platform = mobile AND device_os = Android

Specific App Version:
  platform = mobile AND app_version = 1.2.0

Recent Users (Last 7 Days):
  platform = mobile AND last_seen >= now() - interval '7 days'
```

### Advanced Filters

```
Power Users (High Activity):
  platform = mobile AND 
  event_count_last_7d >= 50

Content Creators:
  platform = mobile AND 
  submission_create_completed_count > 0

Engaged Users (Comments/Votes):
  platform = mobile AND 
  (comment_create_completed_count > 0 OR upvote_clicked_count > 0)

Users with Errors:
  platform = mobile AND 
  error_occurred_count > 0

Session Duration Filter:
  platform = mobile AND 
  session_duration_seconds >= 60
```

## Breakdown Dimensions

Useful dimensions for breaking down metrics:

```
By Platform:
  - device_os (iOS, Android)
  - device_model
  - device_os_version

By User Attributes:
  - user_role
  - subscription_tier
  - account_age_days
  - locale
  - timezone

By App Context:
  - app_version
  - app_build
  - screen_name
  - feature_flag values

By Time:
  - hour_of_day
  - day_of_week
  - week_of_year
  - month

By Behavior:
  - user_cohort (signup month)
  - engagement_level
  - content_preference
```

## Calculated Properties

### Session Metrics

```
Session Length:
  last_event_timestamp - first_event_timestamp
  Unit: Seconds
  
Sessions Per User:
  count(distinct session_id) / count(distinct user_id)
  
Average Events Per Session:
  count(events) / count(distinct session_id)
```

### Engagement Metrics

```
Engagement Score:
  (views * 1) + (votes * 2) + (comments * 3) + (submissions * 5)
  
Days Active (Last 30d):
  count(distinct date) where any event occurred
  
Time to First Action:
  first_meaningful_event_timestamp - signup_timestamp
  Unit: Minutes
```

### Conversion Metrics

```
Signup Conversion Rate:
  (signup_completed / signup_started) * 100
  
Content Engagement Rate:
  (engaged_users / total_viewers) * 100
  Where engaged = voted OR commented OR shared
  
Submission Success Rate:
  (submission_create_completed / submission_create_started) * 100
```

## Time Ranges

Recommended time ranges for different use cases:

```
Real-Time Monitoring:
  - Last 1 hour (refresh every 5 minutes)
  - Use for: Crash monitoring, error spikes

Daily Operations:
  - Last 7 days (refresh daily)
  - Use for: DAU trends, screen views, errors

Weekly Reviews:
  - Last 30 days (refresh weekly)
  - Use for: Funnels, retention, MAU trends

Monthly/Quarterly Planning:
  - Last 90 days or 12 months
  - Use for: Long-term trends, cohort analysis

Historical Analysis:
  - All time or specific date ranges
  - Use for: Year-over-year comparisons
```

## PostHog-Specific Features

### Session Recording

```
Query: Sessions with specific events
Filter: Contains event "error_occurred"
Purpose: Watch user sessions where errors happened
Time Range: Last 7 days
Display: List of session recordings
```

### Feature Flags

```
Query: Event breakdown by feature flag
Event: Any event
Breakdown By: Feature flag "new_ui_enabled"
Purpose: A/B test performance comparison
```

### Experiments

```
Query: Funnel by experiment variant
Funnel: Standard engagement funnel
Breakdown By: Experiment variant (control vs test)
Purpose: Measure experiment impact
Statistical Significance: 95% confidence
```

## Tips & Best Practices

### Query Optimization

1. **Use Time Ranges Wisely**: Shorter ranges = faster queries
2. **Limit Breakdowns**: Too many breakdowns slow down queries
3. **Use Sampling**: For large datasets, enable sampling (10% sample for exploratory analysis)
4. **Cache Results**: PostHog caches insight results; use cached results when possible
5. **Aggregate When Possible**: Use daily aggregation instead of hourly for long time ranges

### Data Quality

1. **Filter Test Users**: Exclude test/internal users with email filter
2. **Remove Duplicates**: Use distinct user/session counts appropriately
3. **Handle Null Values**: Filter out null properties when calculating averages
4. **Timezone Consistency**: Use UTC or user's local timezone consistently
5. **Version Filtering**: Compare same app versions or filter recent versions only

### Dashboard Performance

1. **Limit Insights Per Dashboard**: Max 10-15 insights for good performance
2. **Use Summary Stats**: Add stat cards for quick overview
3. **Progressive Loading**: Place most important insights at the top
4. **Set Refresh Rates**: Balance freshness with performance (5-15 min refresh)
5. **Archive Old Dashboards**: Keep dashboard list manageable

## Example Dashboard Layouts

### Executive Dashboard (High-Level)
```
Row 1: Big Numbers (4 stat cards)
  - DAU | MAU | Stickiness | Crash-Free Rate

Row 2: Trends (2 graphs)
  - DAU Trend (30d) | MAU Trend (90d)

Row 3: Key Funnel (1 funnel)
  - Onboarding Funnel

Row 4: Health Indicators (2 gauges + 1 table)
  - Stability | Retention D7 | Top Errors
```

### Operational Dashboard (Detailed)
```
Row 1: Real-Time Stats (5 stat cards)
  - Active Now | Errors (1h) | Avg Session | Events/min | Crash Rate

Row 2: Detailed Trends (3 graphs)
  - Screen Views | Error Rate | API Response Time

Row 3: Breakdowns (2 tables)
  - Top Screens | Top Errors

Row 4: Paths (1 sankey)
  - Navigation Flow
```

---

**Quick Start**: Copy a query from this guide, paste into PostHog Insights, adjust filters for your needs.

**Need Help?** Refer to [analytics-dashboards.md](./analytics-dashboards.md) for detailed setup instructions.
