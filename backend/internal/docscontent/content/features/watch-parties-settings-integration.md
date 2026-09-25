---
title: "Watch Parties Settings Integration"
summary: "This document provides guidance for integrating the Watch Party Settings & History feature into the application UI."
tags: ["features"]
area: "features"
status: "stable"
owner: "team-core"
version: "1.0"
last_reviewed: 2026-01-29
---

# Watch Party Settings & History - Integration Guide

## Overview

This document provides guidance for integrating the Watch Party Settings & History feature into the application UI.

## What's Been Implemented

### Backend (Complete ✅)

- Database migration for password storage and visibility indexing
- PATCH `/api/v1/watch-parties/{id}/settings` endpoint
- GET `/api/v1/watch-parties/history` endpoint  
- Password protection with bcrypt hashing
- Enhanced join validation with password checking
- Comprehensive unit tests

### Frontend (Complete ✅)

- `WatchPartySettings` component
- `WatchPartyHistory` component
- TypeScript types and API client methods
- Dark mode support

## Integration Steps

### 1. Add Settings Panel to Watch Party UI

In your watch party player/room component, import and use the settings component:

```typescript
import { WatchPartySettings } from '@/components/watch-party';

function WatchPartyRoom({ partyId, currentUser, party }) {
  const [showSettings, setShowSettings] = useState(false);
  const isHost = currentUser.id === party.host_user_id;

  return (
    <div>
      {/* Your existing watch party UI */}
      
      {/* Add settings button (host only) */}
      {isHost && (
        <button onClick={() => setShowSettings(!showSettings)}>
          Settings
        </button>
      )}

      {/* Settings panel */}
      {showSettings && (
        <WatchPartySettings
          partyId={partyId}
          currentPrivacy={party.visibility}
          isHost={isHost}
          onSettingsUpdated={() => {
            // Reload party data to get updated settings
            refetchParty();
            setShowSettings(false);
          }}
        />
      )}
    </div>
  );
}
```

### 2. Add History Page Route

Create a new page at `/watch-parties/history`:

```typescript
// app/watch-parties/history/page.tsx or pages/watch-parties/history.tsx
import { WatchPartyHistory } from '@/components/watch-party';

export default function WatchPartyHistoryPage() {
  return (
    <div className="container mx-auto px-4 py-8">
      <WatchPartyHistory />
    </div>
  );
}
```

Add a navigation link in your user menu or watch party section:

```typescript
<Link href="/watch-parties/history">
  Watch Party History
</Link>
```

### 3. Update Join Flow for Password Protection

Update your join watch party flow to handle password prompts:

```typescript
import { joinWatchParty } from '@/lib/watch-party-api';

async function handleJoinParty(inviteCode: string) {
  try {
    const party = await joinWatchParty(inviteCode);
    // Success - navigate to party
    router.push(`/watch-parties/${party.id}`);
  } catch (error) {
    if (error.message.includes('password')) {
      // Show password input dialog
      const password = await showPasswordPrompt();
      if (password) {
        const party = await joinWatchParty(inviteCode, { password });
        router.push(`/watch-parties/${party.id}`);
      }
    } else {
      // Show error
      toast.error(error.message);
    }
  }
}
```

### 4. Update Create Party Form

Update your create party form to include privacy and password options:

```typescript
import { createWatchParty } from '@/lib/watch-party-api';

function CreatePartyForm() {
  const [title, setTitle] = useState('');
  const [visibility, setVisibility] = useState<'public' | 'friends' | 'invite'>('public');
  const [password, setPassword] = useState('');

  const handleCreate = async () => {
    const request: any = { title, visibility };
    
    // Only include password for invite-only parties
    if (visibility === 'invite' && password) {
      request.password = password;
    }

    const party = await createWatchParty(request);
    // Navigate to party
    router.push(`/watch-parties/${party.id}`);
  };

  return (
    <form onSubmit={handleCreate}>
      <input
        type="text"
        value={title}
        onChange={(e) => setTitle(e.target.value)}
        placeholder="Party title"
      />
      
      <select value={visibility} onChange={(e) => setVisibility(e.target.value as any)}>
        <option value="public">Public</option>
        <option value="friends">Friends Only</option>
        <option value="invite">Invite Only (Password Protected)</option>
      </select>

      {visibility === 'invite' && (
        <input
          type="password"
          value={password}
          onChange={(e) => setPassword(e.target.value)}
          placeholder="Password (optional)"
        />
      )}

      <button type="submit">Create Party</button>
    </form>
  );
}
```

## API Reference

### Update Settings

```typescript
PATCH /api/v1/watch-parties/{id}/settings
Headers: Authorization: Bearer {token}
Body: {
  privacy?: 'public' | 'friends' | 'invite',
  password?: string  // omit to keep, "" to remove, value to update
}
```

### Get History

```typescript
GET /api/v1/watch-parties/history?page=1&limit=20
Headers: Authorization: Bearer {token}
Response: {
  success: true,
  data: {
    history: WatchPartyHistoryEntry[],
    pagination: {
      page: number,
      limit: number,
      total_count: number,
      total_pages: number
    }
  }
}
```

### Join with Password

```typescript
POST /api/v1/watch-parties/{code}/join
Headers: Authorization: Bearer {token}
Body: {
  password?: string
}
```

## Testing Checklist

Before deploying to production:

- [ ] Test settings update as host
- [ ] Verify non-host cannot access settings
- [ ] Test each privacy level (public, friends, invite)
- [ ] Test password protection on join
- [ ] Test password removal (set to empty string)
- [ ] Test history pagination
- [ ] Test history for user with no parties
- [ ] Verify party duration calculations
- [ ] Test dark mode appearance
- [ ] Verify mobile responsiveness
- [ ] Test with expired/invalid tokens
- [ ] Load test settings endpoint (should be < 100ms)

## Database Migration

Run the migration before deploying:

```bash
migrate -path ./backend/migrations -database "postgres://..." up
```

The migration adds:
- `password` column (VARCHAR 255, nullable)
- `idx_parties_visibility` index for performance

## Performance Considerations

- Settings update uses indexed queries: < 100ms typical
- History query is paginated to prevent large data transfers
- Password hashing uses bcrypt default cost (10 rounds)
- Rate limits: 20/hour for settings, standard for history

## Security Notes

- Passwords are bcrypt-hashed, never stored in plain text
- Password hashes are never sent to clients (JSON tag: `json:"-"`)
- Only party host can update settings
- Password validation prevents unauthorized joins
- All endpoints require authentication
- Rate limiting prevents abuse

## Troubleshooting

**Settings update fails with 403:**
- Check user is the party host
- Verify party hasn't ended

**History returns empty:**
- User must have participated in ended parties
- Check pagination parameters

**Join fails with password error:**
- Ensure password matches for invite-only parties
- Check party visibility settings

**Performance issues:**
- Verify database indexes are in place
- Check database connection pool settings
- Monitor bcrypt hashing time (should be ~100ms)

## Future Enhancements

Potential improvements for later:
1. Co-host management (promote viewers to co-host)
2. Invite link expiration
3. Replay functionality for saved parties
4. Export party stats/history
5. Scheduled parties with automatic start
6. Party templates with saved settings
7. Friend list integration for friends-only mode
8. Analytics dashboard for party metrics

## Support

For issues or questions:
- Check API error responses for detailed messages
- Review backend logs for server-side errors
- Verify migration ran successfully
- Test with curl/Postman to isolate frontend issues
