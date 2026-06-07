---
title: User-Submitted Clip System Implementation
summary: This document describes the implementation of the user-submitted clip system with moderation queue for the Clipper application.
tags: ["archive", "implementation", "summary"]
area: docs
status: archived
owner: team-core
version: "1.0"
last_reviewed: 2026-01-29
---


# User-Submitted Clip System Implementation

## Overview

This document describes the implementation of the user-submitted clip system with moderation queue for the Clipper application.

## Features Implemented

### Backend Implementation

#### 1. Database Schema

- **New Table: `clip_submissions`**
  - Tracks user-submitted clips pending moderation
  - Includes metadata cached from Twitch for review
  - Stores submission status (pending, approved, rejected)
  - Records rejection reasons and reviewer information
  
- **View: `submission_stats`**
  - Aggregates user submission statistics
  - Calculates approval rates
  - Tracks total, approved, rejected, and pending counts

#### 2. Repository Layer (`submission_repository.go`)

- `Create()` - Create new submission
- `GetByID()` - Retrieve submission by ID
- `GetByTwitchClipID()` - Find submission by Twitch clip ID (for duplicate detection)
- `ListByUser()` - List all submissions for a user with pagination
- `ListPending()` - List pending submissions for moderation with user info
- `UpdateStatus()` - Update submission status (approve/reject)
- `CountUserSubmissions()` - Count recent submissions for rate limiting
- `GetUserStats()` - Get submission statistics for a user

#### 3. Service Layer (`submission_service.go`)

- **Quality Checks:**
  - Clip age validation (must be < 6 months)
  - Duration validation (must be ≥ 5 seconds)
  - Metadata validation (title, broadcaster name required)
  
- **Duplicate Detection:**
  - Check against existing clips in database
  - Check against pending/approved submissions
  - Allow resubmission of clips rejected > 7 days ago
  
- **Rate Limiting:**
  - 5 submissions per hour per user
  - 20 submissions per day per user
  
- **Auto-Approval Logic:**
  - Admins and moderators auto-approved
  - Users with ≥1000 karma auto-approved
  - Auto-approved submissions immediately create clip and award karma
  
- **Karma System:**
  - +10 karma for approved submissions
  - -5 karma for rejected submissions
  - Minimum 100 karma required to submit

#### 4. API Endpoints (`submission_handler.go`)

**User Endpoints:**

- `POST /api/v1/submissions` - Submit a clip (rate limited: 5/hour)
- `GET /api/v1/submissions` - List user's submissions
- `GET /api/v1/submissions/stats` - Get submission statistics

**Admin Endpoints:**

- `GET /api/v1/admin/submissions` - List pending submissions (moderation queue)
- `POST /api/v1/admin/submissions/:id/approve` - Approve a submission
- `POST /api/v1/admin/submissions/:id/reject` - Reject a submission with reason

### Frontend Implementation

#### 1. Submission Form (`SubmitClipPage.tsx`)

- URL input with validation
- Auto-fetch metadata from backend
- Optional custom title override
- Tag suggestions input
- NSFW checkbox
- Submission reason/context field
- Real-time karma check (requires 100 karma)
- Recent submissions preview

#### 2. User Submission Tracking (`UserSubmissionsPage.tsx`)

- List all user submissions with status
- Status badges (pending/approved/rejected)
- Submission statistics dashboard
- Rejection reasons displayed
- Pagination support
- Thumbnail previews
- Tag display

#### 3. Admin Moderation Queue (`ModerationQueuePage.tsx`)

- List pending submissions
- Embedded clip previews
- Submitter information (karma, role)
- Quick approve/reject actions
- Rejection reason modal
- View on Twitch link
- Real-time updates after actions
- Pending count display

#### 4. API Integration (`submission-api.ts`)

- Type-safe API client methods
- Error handling
- Axios integration with auto-refresh

#### 5. Type Definitions (`submission.ts`)

- `ClipSubmission` - Main submission type
- `ClipSubmissionWithUser` - Submission with user data
- `SubmissionStats` - User statistics
- Request/response types

### Routes Added

**Frontend Routes:**

- `/submit` - Submission form (protected, requires auth)
- `/submissions` - User's submission history (protected, requires auth)
- `/admin/submissions` - Moderation queue (protected, requires admin/moderator role)

**Backend Routes:**

- `POST /api/v1/submissions` - Submit clip
- `GET /api/v1/submissions` - List user submissions
- `GET /api/v1/submissions/stats` - Get user stats
- `GET /api/v1/admin/submissions` - List pending (admin only)
- `POST /api/v1/admin/submissions/:id/approve` - Approve (admin only)
- `POST /api/v1/admin/submissions/:id/reject` - Reject (admin only)

## Security Considerations

### Authentication & Authorization

- All submission endpoints require authentication
- Admin endpoints require admin or moderator role
- Rate limiting prevents spam
- Karma requirement prevents low-quality submissions

### Input Validation

- URL validation and sanitization
- Twitch API validation before acceptance
- Custom title length limits
- Tag input sanitization

### Data Protection

- User IDs validated before database operations
- SQL injection prevented through parameterized queries
- Rate limiting prevents abuse

## Testing

### Backend Tests

- Auto-approval logic tests
- URL extraction and parsing tests
- All existing tests passing

### Security Scanning

- CodeQL analysis passed with 0 alerts
- No vulnerabilities detected in Go or JavaScript code

## Database Migration

To apply the database schema changes:

```sql
-- Run the migration
psql -d clpr -f backend/migrations/000004_add_clip_submissions.up.sql

-- To rollback
psql -d clpr -f backend/migrations/000004_add_clip_submissions.down.sql
```

## Configuration

No additional configuration required. The feature uses existing:

- Authentication system
- Rate limiting infrastructure
- Twitch API integration
- Database connection

## Future Enhancements

The following features were identified but not implemented in this phase:

1. **Notification System**
   - Email/in-app notifications for status changes
   - Moderator alerts for new submissions
   - Daily digest for moderators

2. **Advanced Features**
   - Bulk approval/rejection
   - Submission filters (by user, date, game)
   - Similar clip detection (beyond exact duplicates)
   - Submission appeal system
   - Temporary submission ban system

3. **Analytics**
   - Submission trends
   - Approval rate tracking
   - Popular submitters leaderboard
   - Game/category insights

## Usage Examples

### Submitting a Clip (User)

1. Navigate to `/submit`
2. Paste Twitch clip URL
3. Optionally add custom title, tags, NSFW flag
4. Click "Submit Clip"
5. Clip goes to pending if user has < 1000 karma
6. Auto-approved if user is admin/moderator or has ≥1000 karma

### Moderating Submissions (Admin)

1. Navigate to `/admin/submissions`
2. Review clip preview and metadata
3. Check submitter karma and history
4. Click "Approve" to publish or "Reject" to deny
5. If rejecting, provide reason for user

### Tracking Submissions (User)

1. Navigate to `/submissions`
2. View all submitted clips
3. See status badges (pending/approved/rejected)
4. Read rejection reasons if applicable
5. View approval rate and karma earned

## Performance Considerations

- Pagination implemented for all list endpoints
- Database indexes on frequently queried fields
- Rate limiting prevents API abuse
- Cached Twitch metadata reduces API calls

## Conclusion

The user-submitted clip system is fully implemented and ready for use. It provides a complete workflow from submission through moderation with proper security, validation, and user experience considerations. The system integrates seamlessly with existing authentication, authorization, and karma systems.
