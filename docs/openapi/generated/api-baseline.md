---
title: "API Reference - Baseline"
summary: "Baseline API reference for version 1.0.0"
tags: ["api", "changelog", "openapi"]
area: "openapi"
generated: 2026-01-30T03:02:25.951Z
---

# API Baseline - Version 1.0.0

**Total Endpoints:** 129

## Endpoints by Category

### Health (14)

- **GET /sitemap.xml** - Get sitemap XML
- **GET /robots.txt** - Get robots.txt
- **GET /health/ready** - Readiness check
- **GET /health/live** - Liveness check
- **GET /health/stats** - Health statistics
- **GET /health/cache** - Cache statistics
- **GET /health/cache/check** - Cache health check
- **GET /health/webhooks** - Webhook retry statistics
- **GET /debug/metrics** - Prometheus metrics
- **GET /api/v1/health** - API health check
- **GET /api/v1/ping** - Ping endpoint
- **GET /api/v1/config** - Get public configuration
- **POST /api/v1/logs** - Submit application logs
- **GET /api/v1/logs/stats** - Get log statistics

### Authentication (8)

- **GET /api/v1/auth/twitch** - Initiate Twitch OAuth
- **GET /api/v1/auth/twitch/callback** - Twitch OAuth callback
- **POST /api/v1/auth/twitch/callback** - PKCE OAuth callback
- **POST /api/v1/auth/test-login** - Test login (development only)
- **POST /api/v1/auth/refresh** - Refresh access token
- **POST /api/v1/auth/logout** - Logout
- **GET /api/v1/auth/me** - Get current user
- **POST /api/v1/auth/twitch/reauthorize** - Reauthorize Twitch

### MFA (8)

- **POST /api/v1/auth/mfa/enroll** - Start MFA enrollment
- **POST /api/v1/auth/mfa/verify-enrollment** - Verify MFA enrollment
- **GET /api/v1/auth/mfa/status** - Get MFA status
- **POST /api/v1/auth/mfa/regenerate-backup-codes** - Regenerate backup codes
- **POST /api/v1/auth/mfa/disable** - Disable MFA
- **GET /api/v1/auth/mfa/trusted-devices** - Get trusted devices
- **DELETE /api/v1/auth/mfa/trusted-devices/{id}** - Revoke trusted device
- **POST /api/v1/auth/mfa/verify-login** - Verify MFA login

### Clips (17)

- **GET /api/v1/clips** - List clips
- **GET /api/v1/clips/{id}** - Get clip by ID
- **PUT /api/v1/clips/{id}** - Update clip (Admin)
- **DELETE /api/v1/clips/{id}** - Delete clip (Admin)
- **GET /api/v1/clips/{id}/related** - Get related clips
- **POST /api/v1/clips/batch-media** - Batch get clip media URLs
- **GET /api/v1/clips/{id}/analytics** - Get clip analytics
- **POST /api/v1/clips/{id}/track-view** - Track clip view
- **GET /api/v1/clips/{id}/engagement** - Get clip engagement score
- **POST /api/v1/clips/{id}/vote** - Vote on clip
- **POST /api/v1/clips/{id}/favorite** - Add clip to favorites
- **DELETE /api/v1/clips/{id}/favorite** - Remove clip from favorites
- **PUT /api/v1/clips/{id}/metadata** - Update clip metadata
- **PUT /api/v1/clips/{id}/visibility** - Update clip visibility
- **POST /api/v1/clips/request** - Request clip sync
- **GET /api/v1/scraped-clips** - List scraped clips
- **GET /api/v1/favorites** - List user favorites

### Tags (7)

- **GET /api/v1/clips/{id}/tags** - Get clip tags
- **POST /api/v1/clips/{id}/tags** - Add tags to clip
- **DELETE /api/v1/clips/{id}/tags/{slug}** - Remove tag from clip
- **GET /api/v1/tags** - List tags
- **GET /api/v1/tags/search** - Search tags
- **GET /api/v1/tags/{slug}** - Get tag by slug
- **GET /api/v1/tags/{slug}/clips** - Get clips by tag

### Watch History (1)

- **GET /api/v1/clips/{id}/progress** - Get resume position

### Comments (6)

- **GET /api/v1/clips/{id}/comments** - List clip comments
- **POST /api/v1/clips/{id}/comments** - Create comment
- **GET /api/v1/comments/{id}/replies** - Get comment replies
- **PUT /api/v1/comments/{id}** - Update comment
- **DELETE /api/v1/comments/{id}** - Delete comment
- **POST /api/v1/comments/{id}/vote** - Vote on comment

### Search (7)

- **GET /api/v1/search** - Search clips
- **GET /api/v1/search/suggestions** - Get search suggestions
- **GET /api/v1/search/scores** - Search with scores
- **GET /api/v1/search/trending** - Get trending searches
- **GET /api/v1/search/history** - Get search history
- **GET /api/v1/search/failed** - Get failed searches (Admin)
- **GET /api/v1/search/analytics** - Get search analytics (Admin)

### Submissions (5)

- **GET /api/v1/submissions** - Get user submissions
- **POST /api/v1/submissions** - Submit clip
- **GET /api/v1/submissions/stats** - Get submission statistics
- **GET /api/v1/submissions/metadata** - Get clip metadata
- **GET /api/v1/submissions/check/{clip_id}** - Check clip status

### Reports (1)

- **POST /api/v1/reports** - Submit report

### Moderation (22)

- **GET /api/v1/moderation/appeals** - Get user appeals
- **POST /api/v1/moderation/appeals** - Create appeal
- **POST /api/v1/moderation/sync-bans** - Sync bans from Twitch
- **GET /api/v1/moderation/bans** - Get bans
- **POST /api/v1/moderation/bans** - Create ban
- **GET /api/v1/moderation/ban/{id}** - Get ban details
- **DELETE /api/v1/moderation/ban/{id}** - Revoke ban
- **POST /api/v1/moderation/twitch/ban** - Ban user on Twitch
- **DELETE /api/v1/moderation/twitch/ban** - Unban user on Twitch
- **GET /api/v1/moderation/moderators** - List moderators
- **POST /api/v1/moderation/moderators** - Add moderator
- **PATCH /api/v1/moderation/moderators/{id}** - Update moderator permissions
- **DELETE /api/v1/moderation/moderators/{id}** - Remove moderator
- **GET /api/v1/moderation/audit-logs** - List moderation audit logs
- **GET /api/v1/moderation/audit-logs/export** - Export moderation audit logs
- **GET /api/v1/moderation/audit-logs/{id}** - Get audit log
- **GET /api/v1/moderation/ban-templates** - List ban reason templates
- **POST /api/v1/moderation/ban-templates** - Create ban template
- **GET /api/v1/moderation/ban-templates/stats** - Get template usage statistics
- **GET /api/v1/moderation/ban-templates/{id}** - Get ban template
- **PATCH /api/v1/moderation/ban-templates/{id}** - Update ban template
- **DELETE /api/v1/moderation/ban-templates/{id}** - Delete ban template

### Users (33)

- **GET /api/v1/users/by-username/{username}** - Get user by username
- **GET /api/v1/users/autocomplete** - User autocomplete
- **GET /api/v1/users/{id}** - Get user profile
- **POST /api/v1/users/claim-account** - Claim account
- **GET /api/v1/users/{id}/reputation** - Get user reputation
- **GET /api/v1/users/{id}/karma** - Get user karma
- **GET /api/v1/users/{id}/badges** - Get user badges
- **GET /api/v1/users/{id}/comments** - Get user comments
- **GET /api/v1/users/{id}/clips** - Get user clips
- **GET /api/v1/users/{id}/activity** - Get user activity
- **GET /api/v1/users/{id}/upvoted** - Get user upvoted clips
- **GET /api/v1/users/{id}/downvoted** - Get user downvoted clips
- **GET /api/v1/users/{id}/followers** - Get user followers
- **GET /api/v1/users/{id}/following** - Get following
- **GET /api/v1/users/{id}/following/broadcasters** - Get followed broadcasters
- **POST /api/v1/users/{id}/follow** - Follow user
- **DELETE /api/v1/users/{id}/follow** - Unfollow user
- **POST /api/v1/users/{id}/block** - Block user
- **DELETE /api/v1/users/{id}/block** - Unblock user
- **GET /api/v1/users/me/blocked** - Get blocked users
- **GET /api/v1/users/me/stats** - Get personal statistics
- **GET /api/v1/users/{id}/engagement** - Get user engagement score
- **PUT /api/v1/users/me/profile** - Update profile
- **PUT /api/v1/users/me/social-links** - Update social links
- **GET /api/v1/users/me/settings** - Get user settings
- **PUT /api/v1/users/me/settings** - Update settings
- **GET /api/v1/users/me/export** - Export user data
- **GET /api/v1/users/me/consent** - Get cookie consent
- **POST /api/v1/users/me/consent** - Save cookie consent
- **GET /api/v1/users/me/email-logs** - Get email logs
