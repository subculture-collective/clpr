# Analytics Module

Comprehensive event tracking system for Clipper web and mobile applications.

## Features

- **Unified Tracking**: Single API for both Google Analytics and PostHog
- **GDPR Compliant**: Respects user consent preferences via ConsentContext
- **Auto-enrichment**: Automatically adds user properties, device context, and page context
- **Type-safe**: Full TypeScript support with 40+ predefined events
- **React Integration**: Easy-to-use `useAnalytics` hook for components
- **Mobile Support**: Cross-platform analytics for React Native

## Quick Start

### Web (React)

```typescript
import { trackEvent, AuthEvents, useAnalytics } from '@/lib/telemetry';

// In a component
function MyComponent() {
  const { trackLogin, trackSubmissionCreate } = useAnalytics();
  
  const handleLogin = async () => {
    // ... login logic
    trackLogin('twitch');
  };
  
  return <button onClick={handleLogin}>Login</button>;
}

// Outside of components
trackEvent(AuthEvents.LOGIN_COMPLETED, {
  method: 'twitch'
});
```

### Mobile (React Native)

```typescript
import { trackEvent, AuthEvents, trackScreenView } from '@/lib/telemetry';

// Track screen views
trackScreenView('Home');

// Track events
trackEvent(AuthEvents.LOGIN_COMPLETED, {
  method: 'twitch'
});
```

## Event Categories

The analytics module includes 40+ predefined events organized into 8 categories:

### 1. Authentication Events

```typescript
import { AuthEvents } from '@/lib/telemetry';

// Signup
trackEvent(AuthEvents.SIGNUP_STARTED);
trackEvent(AuthEvents.SIGNUP_COMPLETED, { method: 'twitch' });
trackEvent(AuthEvents.SIGNUP_FAILED, { error: 'Email already exists' });

// Login
trackEvent(AuthEvents.LOGIN_STARTED);
trackEvent(AuthEvents.LOGIN_COMPLETED, { method: 'twitch' });
trackEvent(AuthEvents.LOGIN_FAILED, { error: 'Invalid credentials' });

// Logout
trackEvent(AuthEvents.LOGOUT);
```

### 2. Submission Events (Clips)

```typescript
import { SubmissionEvents } from '@/lib/telemetry';

// Viewing
trackEvent(SubmissionEvents.SUBMISSION_VIEWED, { 
  clip_id: 'abc123',
  title: 'Amazing Play',
  creator_name: 'streamer123'
});

// Creating
trackEvent(SubmissionEvents.SUBMISSION_CREATE_STARTED);
trackEvent(SubmissionEvents.SUBMISSION_CREATE_COMPLETED, {
  clip_id: 'abc123',
  is_nsfw: false,
  tags: ['gaming', 'valorant']
});
trackEvent(SubmissionEvents.SUBMISSION_CREATE_FAILED, {
  error: 'Invalid clip URL'
});

// Sharing
trackEvent(SubmissionEvents.SUBMISSION_SHARED, {
  clip_id: 'abc123',
  share_platform: 'twitter'
});
```

### 3. Engagement Events

```typescript
import { EngagementEvents } from '@/lib/telemetry';

// Voting
trackEvent(EngagementEvents.UPVOTE_CLICKED, {
  target_id: 'clip123',
  target_type: 'clip'
});

// Comments
trackEvent(EngagementEvents.COMMENT_CREATE_COMPLETED, {
  target_id: 'clip123',
  target_type: 'clip'
});

// Following
trackEvent(EngagementEvents.FOLLOW_CLICKED, {
  target_type: 'creator',
  target_id: 'creator123'
});

// Search
trackEvent(EngagementEvents.SEARCH_PERFORMED, {
  search_query: 'valorant clips',
  search_results_count: 42
});

// Favorites
trackEvent(EngagementEvents.FAVORITE_ADDED, {
  target_id: 'clip123',
  target_type: 'clip'
});
```

### 5. Navigation Events

```typescript
import { NavigationEvents } from '@/lib/telemetry';

// Page views (auto-tracked by useAnalytics hook)
trackEvent(NavigationEvents.PAGE_VIEWED, {
  page_path: '/clips',
  page_title: 'Browse Clips'
});

// Feature clicks
trackEvent(NavigationEvents.FEATURE_CLICKED, {
  feature_name: 'dark_mode_toggle'
});

// Links
trackEvent(NavigationEvents.NAV_LINK_CLICKED, {
  link_text: 'Submit Clip',
  link_url: '/submit'
});
```

### 6. Settings Events

```typescript
import { SettingsEvents } from '@/lib/telemetry';

// Theme
trackEvent(SettingsEvents.THEME_CHANGED, {
  setting_name: 'theme',
  new_value: 'dark'
});

// Privacy
trackEvent(SettingsEvents.CONSENT_UPDATED, {
  consent_type: 'analytics',
  new_value: 'true'
});
```

### 7. Error Events

```typescript
import { ErrorEvents, trackError } from '@/lib/telemetry';

// Generic errors
trackEvent(ErrorEvents.ERROR_OCCURRED, {
  error_type: 'NetworkError',
  error_message: 'Failed to fetch'
});

// API errors
trackEvent(ErrorEvents.API_ERROR, {
  api_endpoint: '/api/v1/clips',
  http_status: 500,
  error_message: 'Internal server error'
});

// Convenience method
try {
  // ... code that might throw
} catch (error) {
  trackError(error, {
    errorType: 'ValidationError',
    apiEndpoint: '/api/v1/submit'
  });
}
```

### 8. Performance Events

```typescript
import { PerformanceEvents, trackPerformance } from '@/lib/telemetry';

// Track performance metrics
trackPerformance('api_response_time', 245, 'ms');
trackPerformance('page_load_time', 1250, 'ms');
```

## User Identification

### Web

User identification is handled automatically by the `AuthContext`. When a user logs in, their properties are synced to analytics:

```typescript
// In AuthContext - done automatically
identifyUser('user123', {
  user_id: 'user123',
  username: 'john_doe',
  signup_date: '2024-01-15',
  is_verified: true
});

// On logout - done automatically
resetUser();
```

### Mobile

```typescript
import { identifyUser, resetUser } from '@/lib/telemetry';

// After login
identifyUser('user123', {
  user_id: 'user123',
  username: 'john_doe',
});

// On logout
resetUser();
```

## Configuration

### Environment Variables

#### Web (.env)

```bash
# Enable analytics
VITE_ENABLE_ANALYTICS=true

# Google Analytics
VITE_GA_MEASUREMENT_ID=G-XXXXXXXXXX

# PostHog
VITE_POSTHOG_API_KEY=phc_xxxxxxxxxxxxx
VITE_POSTHOG_HOST=https://app.posthog.com
```

#### Mobile (.env)

```bash
# Enable analytics
EXPO_PUBLIC_ENABLE_ANALYTICS=true

# PostHog
EXPO_PUBLIC_POSTHOG_API_KEY=phc_xxxxxxxxxxxxx
EXPO_PUBLIC_POSTHOG_HOST=https://app.posthog.com
```

## GDPR Compliance

Analytics tracking is **disabled by default** and requires explicit user consent. The system:

1. **Respects browser Do Not Track**: If DNT is enabled, analytics are disabled regardless of consent
2. **Cookie Consent Integration**: Uses `ConsentContext` to check analytics consent
3. **Auto-initialization**: Automatically initializes when consent is granted
4. **Auto-cleanup**: Cleans up user data on logout

### Consent Flow

```typescript
import { enableAnalytics, disableAnalytics } from '@/lib/telemetry';

// When user grants consent
await enableAnalytics();

// When user revokes consent
await disableAnalytics();
```

## Testing

### Debug Mode

Enable debug logging to see all tracked events in the console:

```bash
# Web
VITE_ENABLE_DEBUG=true

# Mobile (automatically enabled in __DEV__)
```

### Manual Testing

```typescript
import { getAnalyticsConfig, isAnalyticsEnabled } from '@/lib/telemetry';

// Check if analytics is enabled
console.log('Analytics enabled:', isAnalyticsEnabled());

// View current configuration (web only)
console.log('Config:', getAnalyticsConfig());
```

## PostHog Integration

### Web

PostHog is already integrated via the `posthog-analytics.ts` module. It loads dynamically when consent is granted.

### Mobile

To enable full PostHog integration on mobile:

1. Install the SDK:
   ```bash
   cd mobile
   npm install posthog-react-native
   ```

2. Uncomment the PostHog integration code in `mobile/lib/telemetry.ts`

3. Initialize in your app:
   ```typescript
   import { PostHogProvider } from 'posthog-react-native';
   
   function App() {
     return (
       <PostHogProvider
         apiKey={process.env.EXPO_PUBLIC_POSTHOG_API_KEY}
         options={{
           host: process.env.EXPO_PUBLIC_POSTHOG_HOST,
         }}
       >
         {/* Your app */}
       </PostHogProvider>
     );
   }
   ```

## Architecture

```
frontend/src/lib/telemetry/
├── events.ts         # Event schema and type definitions
├── tracker.ts        # Unified tracking logic
└── index.ts          # Public API exports

frontend/src/hooks/
└── useAnalytics.ts   # React hook for components

mobile/lib/
└── analytics.ts      # Mobile analytics module

Integration Points:
├── ConsentContext    # GDPR consent management
├── AuthContext       # User identification
└── Components        # Event tracking via useAnalytics
```

## Best Practices

1. **Use predefined events**: Prefer using constants from event schema over custom strings
2. **Include context**: Add relevant properties to help with analysis
3. **Track failures**: Don't just track successes, track errors too
4. **Respect privacy**: Never track PII without explicit consent
5. **Be consistent**: Use the same event names across web and mobile

## Funnels

### Onboarding Funnel

```typescript
// 1. Signup started
trackEvent(AuthEvents.SIGNUP_STARTED);

// 2. OAuth redirect
trackEvent(AuthEvents.OAUTH_REDIRECT, { provider: 'twitch' });

// 3. Signup completed
trackEvent(AuthEvents.SIGNUP_COMPLETED, { method: 'twitch' });

// 4. First submission
trackEvent(SubmissionEvents.SUBMISSION_CREATE_COMPLETED);
```

### Submission Funnel

```typescript
// 1. View submissions list
trackEvent(SubmissionEvents.SUBMISSION_LIST_VIEWED);

// 2. Click to view a submission
trackEvent(SubmissionEvents.SUBMISSION_VIEWED, { clip_id: 'abc' });

// 3. Engage with submission
trackEvent(EngagementEvents.UPVOTE_CLICKED, { target_id: 'abc' });

// 4. Share submission
trackEvent(SubmissionEvents.SUBMISSION_SHARED, { 
  clip_id: 'abc',
  share_platform: 'twitter' 
});
```

## Support

For issues or questions about analytics:

1. Check this README for usage examples
2. Review the event schema in `events.ts`
3. Enable debug mode to see what's being tracked
4. Open an issue on GitHub with the `analytics` label
