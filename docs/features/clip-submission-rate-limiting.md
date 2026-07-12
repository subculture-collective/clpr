# Clip Submission Rate Limiting & Duplicate Detection

## Overview

This document describes the implementation of rate limiting and duplicate detection for clip submissions in Clipper.

## Features

### 1. Rate Limiting

Users are limited in how many clips they can submit to prevent spam and abuse:

- **Hourly Limit**: 10 submissions per hour
- **Daily Limit**: 20 submissions per 24 hours
- **Admin Bypass**: Administrators are not subject to rate limits

#### User Experience

When a user hits the rate limit:

1. **Backend Response**: HTTP 429 (Too Many Requests) with JSON body:
   ```json
   {
     "error": "rate_limit_exceeded",
     "limit": 10,
     "window": 3600,
     "retry_after": 1738352400
   }
   ```

2. **Frontend Display**: 
   - Warning alert with countdown timer
   - Submit button disabled
   - Human-readable time remaining (e.g., "5 minutes 23 seconds")
   - Link to help documentation

3. **State Persistence**:
   - Rate limit state stored in localStorage
   - Persists across page reloads
   - Automatically cleared when timer expires

### 2. Duplicate Detection

Prevents users from submitting clips that already exist in the database:

- **Proactive Detection**: Checks when URL is entered (before submission)
- **Backend Validation**: Double-checks during submission
- **User Feedback**: Clear error message with link to existing clip

#### User Experience

When a duplicate is detected:

1. **Frontend Check** (on URL entry):
   - Calls `/submissions/check/:clip_id` endpoint
   - Shows error immediately if clip exists
   - Provides link to view existing clip

2. **Backend Check** (on submission):
   - Returns HTTP 400 with validation error
   - Error message: "This clip has already been added to our database and cannot be submitted again"

3. **Display**:
   - Error alert with explanation
   - Link to view the existing clip
   - Educational message about duplicate prevention

## Technical Implementation

### Backend

**Files Modified:**
- `backend/internal/services/submission_service.go`
- `backend/internal/handlers/submission_handler.go`

**Key Components:**

1. **RateLimitError Type**:
   ```go
   type RateLimitError struct {
       Error      string `json:"error"`
       Limit      int    `json:"limit"`
       Window     int    `json:"window"`
       RetryAfter int64  `json:"retry_after"`
   }
   ```

2. **Rate Limit Check**:
   - Queries submission count from database
   - Compares against limits
   - Returns RateLimitError if exceeded

3. **Handler Response**:
   - Detects RateLimitError
   - Returns HTTP 429 with proper JSON structure

### Frontend

**Files:**
- `frontend/src/pages/SubmitClipPage.tsx`
- `frontend/src/components/clip/RateLimitError.tsx`
- `frontend/src/components/clip/DuplicateClipError.tsx`

**Key Components:**

1. **RateLimitError Component**:
   - Displays countdown timer
   - Updates every second
   - Shows time remaining in human-readable format
   - Dismissible after expiration

2. **DuplicateClipError Component**:
   - Shows error message
   - Provides link to existing clip
   - Educational content

3. **SubmitClipPage Integration**:
   - Handles 429 responses
   - Stores rate limit in localStorage
   - Proactively checks for duplicates
   - Disables submit button when errors present

## Configuration

### Rate Limits

Configured in `submission_service.go`:

```go
// Hourly limit
if hourlyCount >= 10 {
    // Return rate limit error
}

// Daily limit
if dailyCount >= 20 {
    // Return rate limit error
}
```

### Test Bypass

For E2E testing, rate limits can be bypassed:

```go
bypassRateLimits := testFixturesEnabled && 
    strings.EqualFold(os.Getenv("SUBMISSION_BYPASS_RATE_LIMIT"), "true")
```

## Testing

### E2E Tests

**Location**: `frontend/e2e/tests/clip-submission-flow.spec.ts`

**Test Cases**:

1. **Rate Limiting Test**:
   - Seeds 10 submissions for user
   - Attempts another submission
   - Expects 429 response
   - Verifies countdown timer displayed
   - Confirms submit button disabled

2. **Duplicate Detection Test**:
   - Seeds existing clip
   - Enters duplicate URL
   - Expects error message
   - Verifies form disabled

**Browsers Tested**:
- Chromium
- Firefox
- WebKit

**Total Tests**: 2 unique tests × 3 browsers = 6 test cases

### Unit Tests

Backend service tests should verify:
- Rate limit calculation
- Correct error type returned
- Admin bypass works
- Timestamp calculations

## Analytics

Events tracked:
- `submission.rate_limit_hit`: When user hits rate limit
- `submission.rate_limit_expired`: When rate limit expires
- `submission.duplicate_attempted`: When duplicate submission attempted

## Security Considerations

1. **Rate Limiting**:
   - Prevents submission spam
   - Mitigates abuse/automation
   - Protects database resources

2. **Duplicate Detection**:
   - Prevents duplicate content
   - Reduces moderation overhead
   - Improves data quality

3. **Admin Bypass**:
   - Only applies to rate limits
   - Verified through user role check
   - Does not bypass duplicate detection

## Future Enhancements

Potential improvements:

1. **Dynamic Rate Limits**:
   - Adjust based on user karma/reputation
   - Premium users get higher limits
   - Graduated limits for new users

2. **Better Retry After**:
   - Calculate based on oldest submission in window
   - More accurate countdown
   - Sliding window implementation

3. **Duplicate Claiming**:
   - Allow users to claim scraped clips
   - Transfer ownership of unclaimed clips
   - Karma rewards for claiming

## Related Documentation

- Submission API Documentation
- E2E Testing Guide
- Rate Limiting Strategy
