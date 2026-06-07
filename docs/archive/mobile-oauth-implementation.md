---
title: "OAuth PKCE Implementation Summary"
summary: "This document describes the OAuth PKCE (Proof Key for Code Exchange) authentication implementation f"
tags: ["mobile"]
area: "mobile"
status: "stable"
owner: "team-core"
version: "1.0"
last_reviewed: 2025-12-11
---

# OAuth PKCE Implementation Summary

This document describes the OAuth PKCE (Proof Key for Code Exchange) authentication implementation for the Clipper mobile app.

## Overview

The mobile app now supports secure authentication with Twitch using the OAuth 2.0 Authorization Code flow with PKCE. This implementation provides:

- Secure login via Twitch OAuth
- Persistent sessions across app restarts
- Automatic token refresh
- Clean logout with token revocation
- Protected routes

## Architecture

### Components

#### 1. Auth Context (`contexts/AuthContext.tsx`)

The central authentication state management using React Context.

**Features:**

- Manages user state and authentication status
- Provides auth tokens storage in expo-secure-store
- Exposes auth methods: `setAuthTokens`, `setUser`, `logout`, `getAccessToken`, `getRefreshToken`
- Loads saved auth state on app start

**Storage Keys:**

- `auth_token` - Access token (15 min expiry)
- `refresh_token` - Refresh token (7 day expiry)
- `user_data` - User profile JSON

#### 2. Auth Service (`services/auth.ts`)

Handles the OAuth PKCE flow and backend communication.

**Key Functions:**

- `initiateOAuthFlow()` - Starts OAuth flow with PKCE
  - Generates code verifier (random 32 bytes, base64url)
  - Generates code challenge (SHA256 hash of verifier)
  - Generates state for CSRF protection
  - Opens Twitch authorization page
  - Returns authorization code, state, and code verifier

- `exchangeCodeForTokens()` - Exchanges auth code for tokens
  - Calls `POST /api/v1/auth/twitch/callback`
  - Sends code, state, and code_verifier
  - Backend sets HTTP-only secure cookies

- `getCurrentUser()` - Fetches user profile
  - Calls `GET /api/v1/auth/me`
  - Returns user data for context storage

- `refreshAccessToken()` - Refreshes expired token
  - Calls `POST /api/v1/auth/refresh`
  - Backend refreshes cookies automatically

- `logoutUser()` - Revokes tokens
  - Calls `POST /api/v1/auth/logout`
  - Backend revokes refresh token

#### 3. API Interceptor (`lib/api.ts`)

Axios interceptors for automatic token handling.

**Request Interceptor:**

- Reads access token from expo-secure-store
- Adds `Authorization: Bearer {token}` header

**Response Interceptor:**

- Detects 401 (Unauthorized) responses
- Automatically calls refresh token endpoint
- Queues failed requests during refresh
- Retries queued requests after successful refresh
- Clears auth state if refresh fails

#### 4. Login Screen (`app/auth/login.tsx`)

User-facing login interface.

**Flow:**

1. User clicks "Login with Twitch"
2. Opens Twitch OAuth page in browser
3. User authorizes app
4. Redirects back to app with auth code
5. Exchanges code for tokens
6. Fetches user profile
7. Saves auth state
8. Navigates to main app

#### 5. Profile Screen (`app/(tabs)/profile.tsx`)

Displays user profile and logout option.

**Features:**

- Shows user avatar, display name, username
- Shows reputation score
- Logout button with confirmation
- Guest view when not authenticated

#### 6. Route Guard Hook (`hooks/use-require-auth.ts`)

Helper hook for protecting screens.

**Usage:**

```typescript
function ProtectedScreen() {
  const { isAuthenticated, isLoading } = useRequireAuth();

  if (isLoading) return <LoadingSpinner />;

  // Screen content - auto redirects if not authenticated
}
```

## Security Features

### PKCE (Proof Key for Code Exchange)

PKCE adds security for mobile apps where client secrets cannot be securely stored:

1. **Code Verifier**: Random 43-128 character string (we use 32 bytes = 43 chars base64url)
2. **Code Challenge**: Base64url-encoded SHA256 hash of the verifier
3. **Challenge Method**: `S256` (SHA256 hashing)

The verifier stays on the device, while only the challenge is sent to Twitch. When exchanging the auth code, the verifier is sent to prove the same client is completing the flow.

### State Parameter

Random state value prevents CSRF attacks:

- Generated on auth initiation (32 random bytes)
- Sent to Twitch and returned in callback
- Verified to match before accepting auth code

### Secure Storage

All sensitive data stored in expo-secure-store:

- iOS: Encrypted in device Keychain
- Android: Encrypted in KeyStore
- Not accessible by other apps
- Persists across app restarts
- Cleared on app uninstall

### Token Refresh

Automatic token refresh prevents session interruption:

- Access token expires after 15 minutes
- Refresh token valid for 7 days
- 401 responses trigger automatic refresh
- Failed requests queued and retried
- Multiple simultaneous requests handled gracefully

## Configuration

### App Configuration (`app.json`)

```json
{
  "scheme": "clpr"
}
```

This enables deep linking for OAuth callbacks via `clpr://` URLs.

### Environment Variables

Required in `.env` file:

```bash
EXPO_PUBLIC_API_URL=http://localhost:8080/api/v1
EXPO_PUBLIC_TWITCH_CLIENT_ID=your_twitch_client_id
```

### Twitch App Configuration

In the [Twitch Developer Console](https://dev.twitch.tv/console/apps):

1. Create or edit your application
2. Add OAuth Redirect URL: `clpr://`
3. Note your Client ID (no client secret needed for PKCE)

## Backend Integration

The backend provides these endpoints:

### `POST /api/v1/auth/twitch/callback`

Exchange auth code for tokens.

**Request:**

```json
{
  "code": "auth_code_from_twitch",
  "state": "csrf_state_token",
  "code_verifier": "pkce_verifier"
}
```

**Response:**

```json
{
  "message": "Authentication successful"
}
```

Backend sets HTTP-only secure cookies:

- `access_token` (15 min)
- `refresh_token` (7 days)

### `GET /api/v1/auth/me`

Get current user profile.

**Response:**

```json
{
  "id": "user_uuid",
  "twitch_user_id": "12345",
  "username": "user123",
  "display_name": "User Name",
  "email": "user@example.com",
  "profile_image_url": "https://...",
  "role": "user",
  "is_banned": false,
  "reputation_score": 100,
  "created_at": "2024-01-01T00:00:00Z"
}
```

### `POST /api/v1/auth/refresh`

Refresh access token.

**Request:** None (uses refresh_token cookie)

**Response:**

```json
{
  "message": "Token refreshed successfully"
}
```

Backend updates `access_token` cookie.

### `POST /api/v1/auth/logout`

Revoke tokens and logout.

**Request:** None (uses cookies)

**Response:**

```json
{
  "message": "Logged out successfully"
}
```

Backend revokes refresh token and clears cookies.

## Testing

### Manual Testing Checklist

- [ ] User can initiate login
- [ ] Twitch OAuth page opens in browser
- [ ] User can authorize the app
- [ ] App receives callback and exchanges code
- [ ] User profile is fetched and displayed
- [ ] Session persists after app restart
- [ ] Access token auto-refreshes on 401
- [ ] User can logout
- [ ] Tokens are revoked on logout
- [ ] Auth state is cleared after logout
- [ ] Protected routes redirect when not authenticated

### Test Scenarios

1. **Fresh Install Flow**
   - Install app
   - Open app → see login screen
   - Login with Twitch → see profile
   - Close and reopen app → still logged in

2. **Token Refresh Flow**
   - Login to app
   - Wait 15 minutes (access token expiry)
   - Make API request → auto-refresh
   - Request succeeds after refresh

3. **Logout Flow**
   - Login to app
   - Navigate to profile
   - Click logout
   - Confirm logout
   - App clears local state
   - Backend revokes tokens

4. **Error Handling**
   - Cancel OAuth → show error, stay on login
   - Network error during exchange → show error
   - Invalid credentials → show error
   - Expired refresh token → redirect to login

## Troubleshooting

### "Authentication failed" error

**Cause:** Code exchange failed
**Solution:**

- Check API_URL is correct
- Verify backend is running
- Check Twitch Client ID matches backend

### "State mismatch" error

**Cause:** CSRF state doesn't match
**Solution:**

- Clear app data and try again
- This could indicate a security issue

### App doesn't redirect after Twitch auth

**Cause:** Deep link not configured
**Solution:**

- Verify `scheme: "clpr"` in app.json
- Rebuild app with `npx expo prebuild`
- Check Twitch app has `clpr://` redirect

### Session doesn't persist

**Cause:** Secure store not working
**Solution:**

- Check device permissions
- Test on different device/emulator
- Check for expo-secure-store errors in logs

### API requests return 401 repeatedly

**Cause:** Token refresh failing
**Solution:**

- Check refresh token is valid (< 7 days old)
- Verify backend refresh endpoint works
- Check network connectivity
- Try logout and login again

## Future Enhancements

Potential improvements for future iterations:

1. **Biometric Authentication**: Add Face ID/Touch ID for re-authentication
2. **Token Expiry Warnings**: Notify user before refresh token expires
3. **Multiple Account Support**: Allow switching between accounts
4. **Offline Mode**: Better handling of offline scenarios
5. **Session Analytics**: Track session duration and refresh patterns

## References

- [OAuth 2.0 RFC](https://tools.ietf.org/html/rfc6749)
- [PKCE RFC](https://tools.ietf.org/html/rfc7636)
- [Twitch OAuth Documentation](https://dev.twitch.tv/docs/authentication)
- [Expo Auth Session](https://docs.expo.dev/versions/latest/sdk/auth-session/)
- [Expo Secure Store](https://docs.expo.dev/versions/latest/sdk/securestore/)
