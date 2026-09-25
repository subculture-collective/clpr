---
title: "UNCLAIMED ACCOUNTS"
summary: "The Unclaimed Accounts feature allows the platform to automatically create user profiles for Twitch clip creators who don't have accounts yet. This enables the community to see who created clips and a"
tags: ["features"]
area: "features"
status: "stable"
owner: "team-core"
version: "1.0"
last_reviewed: 2026-01-29
---

# Unclaimed Accounts Feature

## Overview

The Unclaimed Accounts feature allows the platform to automatically create user profiles for Twitch clip creators who don't have accounts yet. This enables the community to see who created clips and allows these creators to claim their profiles later.

## Implementation Summary

### Database Schema Changes

**Migration: `000093_add_account_status`**

Added `account_status` field to the `users` table with three possible values:
- `active` - Normal registered user account
- `unclaimed` - Auto-created account from Twitch clip creator (not yet claimed)
- `pending` - Account in the process of being claimed (transitional state)

Also made `twitch_id` nullable to support unclaimed accounts that don't have OAuth authentication.

### Backend Components

#### 1. Clip Sync Service Enhancement

**File:** `backend/internal/services/clip_sync_service.go`

- Added `UserRepository` to `ClipSyncService`
- Implemented `ensureUnclaimedUser()` method that:
  - Checks if a user exists with the given Twitch ID
  - Creates an unclaimed account if not found
  - Generates a normalized username from the Twitch display name
  - Sets `account_status` to "unclaimed"

- Modified `processClip()` to call `ensureUnclaimedUser()` for both clip creators and broadcasters before saving clips

**Username Generation Logic:**
- Normalizes display name to lowercase
- Replaces non-alphanumeric characters with underscores
- Falls back to `user_{twitchID}` if display name is too short
- Truncates to 50 characters maximum

#### 2. User Repository Updates

**File:** `backend/internal/repository/user_repository.go`

Added methods:
- `UpdateAccountStatus()` - Updates user's account_status field
- `UpdateDisplayName()` - Updates user's display_name field
- Modified `Create()` to include account_status in INSERT query with default value "active"

#### 3. Claim Account Endpoint

**File:** `backend/internal/handlers/user_handler.go`

Created `ClaimAccount()` handler:
- **Route:** `POST /api/v1/users/claim-account`
- **Authentication:** Required
- **Request Body:**
  ```json
  {
    "twitch_id": "12345678"
  }
  ```

**Validation:**
- Verifies authenticated user's Twitch ID matches the claim request
- Ensures the target account exists and has status "unclaimed"
- Prevents claiming if authenticated user already has an active account

**Claim Process:**
1. Finds unclaimed account by Twitch ID
2. Transfers display name from unclaimed account to authenticated user
3. Updates authenticated user's account_status to "active"
4. Marks unclaimed account as "pending" to prevent reuse

#### 4. Main Application Updates

**File:** `backend/cmd/api/main.go`

- Updated `NewClipSyncService()` call to include `userRepo` parameter
- Added route for claim account endpoint

### Frontend Integration (Planned)

**Required Changes:**

1. **Profile Pages** - Update to show unclaimed status
   - Display banner: "This is an unclaimed profile"
   - Show "Claim This Profile" button for authenticated users matching the Twitch ID
   - Disable certain features (following, messaging) for unclaimed accounts

2. **Clip Attribution** - Already displaying creator information
   - Creator name links to profile even if unclaimed
   - Show indicator that account is unclaimed

3. **Claim Account Flow**
   - Create claim account modal/page
   - Call `POST /api/v1/users/claim-account` endpoint
   - Show success message and redirect to now-active profile

## API Endpoints

### Claim Account

**POST** `/api/v1/users/claim-account`

Claims an unclaimed profile for the authenticated user.

**Headers:**
```
Authorization: Bearer <token>
```

**Request Body:**
```json
{
  "twitch_id": "12345678"
}
```

**Response (Success):**
```json
{
  "success": true,
  "message": "account claimed successfully"
}
```

**Error Responses:**

- `401 Unauthorized` - Not authenticated
- `403 Forbidden` - Twitch ID doesn't match authenticated user
- `404 Not Found` - No unclaimed account found
- `400 Bad Request` - Account is not unclaimed or user already has active account

## Future Enhancements

### TODO Items

1. **Content Transfer**
   - Transfer clips from unclaimed account to claimed account
   - Transfer votes, favorites, and other user activity
   - Update clip ownership records

2. **Notification System**
   - Notify users when someone claims a profile they've interacted with
   - Email notification to Twitch email (if available) when clips are posted

3. **Profile Merging**
   - Handle cases where user already has content before claiming
   - Merge karma points, badges, and reputation

4. **Analytics**
   - Track unclaimed account creation rate
   - Monitor claim conversion rate
   - Identify popular unclaimed creators for outreach

5. **Admin Tools**
   - View all unclaimed accounts
   - Manually link unclaimed accounts if needed
   - Handle disputed claims

## Testing Considerations

1. **Unit Tests Needed:**
   - `ensureUnclaimedUser()` edge cases
   - Username generation with special characters
   - Claim account validation logic

2. **Integration Tests:**
   - Clip sync creates unclaimed users
   - Claim account endpoint workflow
   - Concurrent claims of same account

3. **E2E Tests:**
   - Full claim flow from frontend
   - Profile display for unclaimed accounts
   - Navigation between unclaimed profiles

## Migration Path

The migration is backward compatible:
- Existing users automatically get `account_status = 'active'`
- Existing `twitch_id` values remain non-null
- No data migration required for existing records

## Performance Considerations

- Username generation is performed inline during clip sync
- Database has index on `account_status` for filtering
- Partial unique index on `twitch_id` allows multiple NULL values but ensures uniqueness for non-NULL values
- Check constraint on `account_status` ensures data integrity

## Security Notes

- Claim endpoint verifies Twitch ID match to prevent account takeover
- Only authenticated users can claim accounts
- Unclaimed accounts have limited functionality (no auth, no posting)
- Prevents duplicate claims through status transitions
