---
title: "API Reference - Baseline"
summary: "Baseline API reference for version 1.0.0"
tags: ["api", "changelog", "openapi"]
area: "openapi"
generated: 2026-11-07T06:50:55.098Z
---

# API Baseline - Version 1.0.0

**Total Endpoints:** 410

## Endpoints by Category

### Health (10)

- **GET /sitemap.xml** - Get sitemap XML
- **GET /robots.txt** - Get robots.txt
- **GET /health/ready** - Readiness check
- **GET /health/live** - Liveness check
- **GET /api/v1/health** - API health check
- **GET /api/v1/ping** - Ping endpoint
- **GET /api/v1/config** - Get public configuration
- **POST /api/v1/logs** - Submit application logs
- **GET /api/v1/logs/stats** - Get log statistics
- **GET /health** - Get process health

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

### Clips (19)

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
- **GET /api/v1/creators/{creator}/clips** - List clips for a Twitch creator ID
- **GET /api/v1/clips/{id}/media** - Resolve direct clip media

### Tags (8)

- **GET /api/v1/clips/{id}/tags** - Get clip tags
- **POST /api/v1/clips/{id}/tags** - Add tags to clip
- **DELETE /api/v1/clips/{id}/tags/{slug}** - Remove tag from clip
- **GET /api/v1/tags** - List tags
- **GET /api/v1/tags/search** - Search tags
- **GET /api/v1/tags/tree** - Get tag hierarchy tree
- **GET /api/v1/tags/{slug}** - Get tag by slug
- **GET /api/v1/tags/{slug}/clips** - Get clips by tag

### Topics (1)

- **GET /api/v1/clips/{id}/topics** - Get directly classified clip topics

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

### Submissions (6)

- **GET /api/v1/submissions** - Get user submissions
- **POST /api/v1/submissions** - Submit clip
- **POST /api/v1/submissions/upload** - Upload a clip for moderation
- **GET /api/v1/submissions/stats** - Get submission statistics
- **GET /api/v1/submissions/metadata** - Get clip metadata
- **GET /api/v1/submissions/check/{clip_id}** - Check clip status

### Reports (1)

- **POST /api/v1/reports** - Submit report

### Moderation (20)

- **GET /api/v1/moderation/appeals** - Get user appeals
- **POST /api/v1/moderation/appeals** - Create appeal
- **GET /api/v1/moderation/bans** - List channel bans
- **POST /api/v1/moderation/bans** - Create ban
- **GET /api/v1/moderation/ban/{id}** - Get ban details
- **DELETE /api/v1/moderation/ban/{id}** - Revoke ban
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
- **POST /api/v1/moderation/ban** - Ban a user from a community

### Users (32)

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
- **GET /api/v1/users/me/consent** - Get cookie consent
- **POST /api/v1/users/me/consent** - Save cookie consent
- **GET /api/v1/users/me/email-logs** - Get email logs
- **GET /api/v1/watch-history** - Get the current user's watch history
- **POST /api/v1/watch-history** - Record clip watch progress
- **DELETE /api/v1/watch-history** - Clear the current user's watch history

### Webhooks (9)

- **GET /api/v1/webhooks** - List webhook subscriptions
- **POST /api/v1/webhooks** - Create a webhook subscription
- **GET /api/v1/webhooks/events** - List supported webhook events
- **GET /api/v1/webhooks/{id}** - Get a webhook subscription
- **PATCH /api/v1/webhooks/{id}** - Update a webhook subscription
- **DELETE /api/v1/webhooks/{id}** - Delete a webhook subscription
- **GET /api/v1/webhooks/{id}/deliveries** - List webhook delivery attempts
- **POST /api/v1/webhooks/{id}/regenerate-secret** - Regenerate a webhook signing secret
- **POST /api/v1/webhooks/sendgrid** - Ingest signed SendGrid event notifications

### Support (1)

- **POST /api/v1/contact** - Submit a contact message

### Admin (89)

- **GET /api/v1/admin/contact** - List contact messages
- **PUT /api/v1/admin/contact/{id}/status** - Update a contact message status
- **GET /api/v1/admin/reports** - List content reports
- **GET /api/v1/admin/reports/{id}** - Get a content report
- **PUT /api/v1/admin/reports/{id}** - Resolve a content report
- **GET /api/v1/feeds/analytics** - Get aggregate feed event metrics
- **GET /api/v1/feeds/analytics/hourly** - Get hourly metrics for an event type
- **GET /api/v1/admin/webhooks/dlq** - List outbound webhook dead-letter items
- **DELETE /api/v1/admin/webhooks/dlq/{id}** - Permanently delete a dead-letter item
- **POST /api/v1/admin/webhooks/dlq/{id}/replay** - Replay a dead-lettered webhook
- **POST /api/v1/admin/sync/clips** - Run a clip synchronization
- **GET /api/v1/admin/sync/status** - Get persisted clip synchronization status
- **GET /api/v1/admin/audit-logs** - List immutable moderation audit events
- **GET /api/v1/admin/audit-logs/export** - Export filtered audit events as CSV
- **GET /api/v1/admin/analytics/alerts** - List engagement alerts
- **GET /api/v1/admin/analytics/content** - Get platform content metrics
- **GET /api/v1/admin/analytics/health** - Get platform engagement health
- **GET /api/v1/admin/analytics/overview** - Get platform analytics overview
- **GET /api/v1/admin/analytics/trending** - Get trending engagement metrics
- **GET /api/v1/admin/analytics/trends** - Get platform metric trends
- **GET /api/v1/admin/email/alerts** - List unresolved email alerts
- **POST /api/v1/admin/email/alerts/{id}/acknowledge** - Acknowledge an email alert
- **POST /api/v1/admin/email/alerts/{id}/resolve** - Resolve an email alert
- **GET /api/v1/admin/email/logs** - Search email delivery logs
- **GET /api/v1/admin/email/metrics** - Get aggregate email metrics
- **GET /api/v1/admin/email/metrics/dashboard** - Get the email operations dashboard
- **GET /api/v1/admin/email/metrics/templates** - Get email metrics by template
- **GET /api/v1/admin/moderation/abuse/{userId}** - Get abuse statistics for a user
- **GET /api/v1/admin/moderation/appeals** - List moderation appeals
- **POST /api/v1/admin/moderation/appeals/{id}/resolve** - Resolve a moderation appeal
- **GET /api/v1/admin/moderation/audit** - List moderation decisions
- **POST /api/v1/admin/moderation/bulk** - Moderate multiple queue items
- **GET /api/v1/admin/moderation/events** - List pending moderation events
- **POST /api/v1/admin/moderation/events/{id}/process** - Process a moderation event
- **POST /api/v1/admin/moderation/events/{id}/review** - Mark a moderation event reviewed
- **GET /api/v1/admin/moderation/events/{type}** - List moderation events by type
- **GET /api/v1/admin/moderation/queue** - List moderation queue items
- **GET /api/v1/admin/moderation/queue/stats** - Get moderation queue statistics
- **GET /api/v1/admin/moderation/stats** - Get moderation event statistics
- **GET /api/v1/admin/moderation/toxicity/metrics** - Get toxicity-classification metrics
- **POST /api/v1/admin/moderation/{id}/approve** - Approve a pending moderation item
- **POST /api/v1/admin/moderation/{id}/reject** - Reject a pending moderation item
- **POST /api/v1/admin/nsfw/batch-detect** - Detect NSFW content in an image batch
- **GET /api/v1/admin/nsfw/config** - Get NSFW detector availability
- **POST /api/v1/admin/nsfw/detect** - Detect NSFW content in one image
- **GET /api/v1/admin/nsfw/health** - Check NSFW detector readiness
- **GET /api/v1/admin/nsfw/metrics** - Get NSFW detection metrics
- **GET /api/v1/admin/playlist-scripts** - List playlist automation scripts
- **POST /api/v1/admin/playlist-scripts** - Create a playlist automation script
- **PUT /api/v1/admin/playlist-scripts/{id}** - Update a playlist automation script
- **DELETE /api/v1/admin/playlist-scripts/{id}** - Delete a playlist automation script
- **POST /api/v1/admin/playlist-scripts/{id}/generate** - Generate a playlist from a script
- **GET /api/v1/admin/submissions** - List pending clip submissions
- **POST /api/v1/admin/submissions/bulk-approve** - Approve multiple submissions
- **POST /api/v1/admin/submissions/bulk-reject** - Reject multiple submissions
- **GET /api/v1/admin/submissions/rejection-reasons** - List rejection-reason templates
- **POST /api/v1/admin/submissions/{id}/approve** - Approve a pending submission
- **POST /api/v1/admin/submissions/{id}/reject** - Reject a pending submission
- **GET /api/v1/admin/tags** - List tags including suppressed records
- **POST /api/v1/admin/tags** - Create a tag
- **PUT /api/v1/admin/clips/{id}/topics** - Replace a clip's topic assignments
- **POST /api/v1/admin/topics/{id}/merge** - Merge one topic into another
- **POST /api/v1/admin/topics/{id}/split** - Split selected clips into a new topic
- **GET /api/v1/admin/tags/blacklist** - List blacklisted tag slugs
- **POST /api/v1/admin/tags/blacklist** - Blacklist a tag slug
- **DELETE /api/v1/admin/tags/blacklist/{id}** - Remove a blacklisted tag slug
- **GET /api/v1/admin/tags/promotion-queue** - List tag promotion candidates
- **POST /api/v1/admin/tags/promotion-queue/{id}/approve** - Approve a tag promotion
- **POST /api/v1/admin/tags/promotion-queue/{id}/reject** - Reject a tag promotion
- **PUT /api/v1/admin/tags/{id}** - Replace a tag
- **DELETE /api/v1/admin/tags/{id}** - Delete a tag
- **POST /api/v1/admin/tags/suppressions/{id}** - Suppress a tag from public surfaces
- **DELETE /api/v1/admin/tags/suppressions/{id}** - Restore a suppressed tag
- **GET /api/v1/admin/users** - Search users for administration
- **GET /api/v1/admin/moderators** - List platform moderators
- **POST /api/v1/admin/moderators** - Add a platform moderator
- **GET /api/v1/admin/openapi.json** - Get the running backend OpenAPI contract
- **PATCH /api/v1/admin/moderators/{id}** - Confirm or update platform moderator access
- **DELETE /api/v1/admin/moderators/{id}** - Revoke platform moderator access
- **POST /api/v1/admin/users/{id}/badges** - Award a badge to a user
- **DELETE /api/v1/admin/users/{id}/badges/{badgeId}** - Remove a badge from a user
- **POST /api/v1/admin/users/{id}/ban** - Ban a user
- **GET /api/v1/admin/users/{id}/comment-suspension-history** - Get comment-suspension history
- **PATCH /api/v1/admin/users/{id}/karma** - Set user karma
- **POST /api/v1/admin/users/{id}/lift-comment-suspension** - Lift a comment suspension
- **PATCH /api/v1/admin/users/{id}/role** - Change a user role
- **POST /api/v1/admin/users/{id}/suspend-comments** - Suspend comment privileges
- **POST /api/v1/admin/users/{id}/toggle-comment-review** - Set comment-review requirement
- **POST /api/v1/admin/users/{id}/unban** - Unban a user

### Analytics (1)

- **POST /api/v1/events** - Queue one or more analytics events

### Operations (3)

- **GET /internal/operations/webhooks** - Get webhook delivery and retry health
- **GET /internal/operations/cache** - Get typed Redis operational statistics
- **GET /internal/operations/cache/check** - Check Redis connectivity

### Broadcasters (9)

- **GET /api/v1/broadcasters/live** - List currently live broadcasters
- **GET /api/v1/broadcasters/{id}/live-status** - Get a broadcaster's live status
- **GET /api/v1/broadcasters/discover** - Discover trending, rising, and new creators
- **GET /api/v1/broadcasters/popular** - List popular broadcasters
- **GET /api/v1/broadcasters/rankings** - List broadcaster rankings
- **GET /api/v1/broadcasters/{id}** - Get a visible broadcaster profile
- **GET /api/v1/broadcasters/{id}/clips** - List visible broadcaster clips
- **POST /api/v1/broadcasters/{id}/follow** - Follow a broadcaster
- **DELETE /api/v1/broadcasters/{id}/follow** - Unfollow a broadcaster

### Pages (4)

- **GET /api/v1/share/clips/{id}** - Render a clip link preview
- **GET /clips/game/{gameSlug}** - Render a game clip collection page
- **GET /clips/streamer/{broadcasterName}** - Render a broadcaster clip profile page
- **GET /clips/streamer/{broadcasterName}/{gameSlug}** - Render a broadcaster-by-game clip page

### Documentation (4)

- **GET /api/v1/docs** - List available Markdown documentation
- **GET /api/v1/docs/search** - Search documentation text
- **GET /api/v1/docs/{path}** - Get a top-level Markdown document
- **GET /api/v1/docs/content/{path}** - Get a nested Markdown document

### Creator Exports (4)

- **POST /api/v1/creators/me/export/request** - Request a creator data export
- **GET /api/v1/creators/me/export/status/{id}** - Get a creator export's status
- **GET /api/v1/creators/me/export/download/{id}** - Download a completed creator export
- **GET /api/v1/creators/me/exports** - List the current creator's export requests

### Creator Analytics (4)

- **GET /api/v1/creators/{creator}/analytics/overview** - Get public creator analytics totals
- **GET /api/v1/creators/{creator}/analytics/clips** - List a creator's top-performing clips
- **GET /api/v1/creators/{creator}/analytics/trends** - Get a creator metric time series
- **GET /api/v1/creators/{creator}/analytics/audience** - Get aggregate creator audience device insights

### Discovery Lists (8)

- **GET /api/v1/discovery-lists** - List active discovery lists
- **GET /api/v1/discovery-lists/{id}** - Get a discovery list by UUID or slug
- **GET /api/v1/discovery-lists/{id}/clips** - List clips in a discovery list
- **POST /api/v1/discovery-lists/{id}/follow** - Follow a discovery list
- **DELETE /api/v1/discovery-lists/{id}/follow** - Stop following a discovery list
- **POST /api/v1/discovery-lists/{id}/bookmark** - Bookmark a discovery list
- **DELETE /api/v1/discovery-lists/{id}/bookmark** - Remove a discovery-list bookmark
- **GET /api/v1/users/me/discovery-list-follows** - List discovery lists followed by the current user

### Admin Discovery Lists (9)

- **GET /api/v1/admin/discovery-lists** - List all active and inactive curated discovery lists
- **POST /api/v1/admin/discovery-lists** - Create a curated discovery list
- **GET /api/v1/admin/discovery-lists/{id}** - Get a curated discovery list
- **PUT /api/v1/admin/discovery-lists/{id}** - Update a curated discovery list
- **DELETE /api/v1/admin/discovery-lists/{id}** - Soft-delete a curated discovery list
- **GET /api/v1/admin/discovery-lists/{id}/clips** - List clips in a curated discovery list
- **POST /api/v1/admin/discovery-lists/{id}/clips** - Add an active clip to a curated discovery list
- **DELETE /api/v1/admin/discovery-lists/{id}/clips/{clipId}** - Remove a clip from a curated discovery list
- **PUT /api/v1/admin/discovery-lists/{id}/clips/reorder** - Atomically reorder every clip in a curated discovery list

### Admin Broadcasters (1)

- **POST /api/v1/admin/broadcasters/refresh-rankings** - Refresh broadcaster ranking materialization

### Reputation (2)

- **GET /api/v1/badges** - List badge definitions
- **GET /api/v1/leaderboards/{type}** - Get a public reputation leaderboard

### Categories (4)

- **GET /api/v1/categories** - List categories
- **GET /api/v1/categories/{slug}** - Get a category
- **GET /api/v1/categories/{slug}/clips** - List category clips
- **GET /api/v1/categories/{slug}/games** - List category games

### Communities (21)

- **GET /api/v1/communities** - List public communities
- **POST /api/v1/communities** - Create a community
- **GET /api/v1/communities/search** - Search public communities
- **GET /api/v1/communities/{id}** - Get a visible community
- **PUT /api/v1/communities/{id}** - Update a community
- **DELETE /api/v1/communities/{id}** - Delete a community
- **POST /api/v1/communities/{id}/ban** - Ban a community member
- **DELETE /api/v1/communities/{id}/ban/{userId}** - Unban a community member
- **GET /api/v1/communities/{id}/bans** - List community bans
- **POST /api/v1/communities/{id}/clips** - Add a clip to a community
- **DELETE /api/v1/communities/{id}/clips/{clipId}** - Remove a community clip
- **GET /api/v1/communities/{id}/discussions** - List visible community discussions
- **POST /api/v1/communities/{id}/discussions** - Create a community discussion
- **GET /api/v1/communities/{id}/discussions/{discussionId}** - Get a visible community discussion
- **PUT /api/v1/communities/{id}/discussions/{discussionId}** - Update a community discussion
- **DELETE /api/v1/communities/{id}/discussions/{discussionId}** - Delete a community discussion
- **GET /api/v1/communities/{id}/feed** - Get a visible community feed
- **POST /api/v1/communities/{id}/join** - Join a public community
- **POST /api/v1/communities/{id}/leave** - Leave a community
- **GET /api/v1/communities/{id}/members** - List visible community members
- **PUT /api/v1/communities/{id}/members/{userId}/role** - Update a community member role

### Feeds (15)

- **GET /api/v1/feeds/clips** - List filtered feed clips
- **GET /api/v1/feeds/discover** - Discover public feeds
- **GET /api/v1/feeds/following** - List clips from followed accounts
- **GET /api/v1/feeds/search** - Search public feeds
- **GET /api/v1/users/{id}/feeds** - List a user's visible feeds
- **POST /api/v1/users/{id}/feeds** - Create a feed for the authenticated user
- **GET /api/v1/users/{id}/feeds/{feedId}** - Get a visible feed
- **PUT /api/v1/users/{id}/feeds/{feedId}** - Update an owned feed
- **DELETE /api/v1/users/{id}/feeds/{feedId}** - Delete an owned feed
- **GET /api/v1/users/{id}/feeds/{feedId}/clips** - List clips in a visible feed
- **POST /api/v1/users/{id}/feeds/{feedId}/clips** - Add a clip to an owned feed
- **PUT /api/v1/users/{id}/feeds/{feedId}/clips/reorder** - Reorder every clip in an owned feed
- **DELETE /api/v1/users/{id}/feeds/{feedId}/clips/{clipId}** - Remove a clip from an owned feed
- **POST /api/v1/users/{id}/feeds/{feedId}/follow** - Follow a public feed
- **DELETE /api/v1/users/{id}/feeds/{feedId}/follow** - Unfollow a public feed

### Games (6)

- **GET /api/v1/games/trending** - List trending games
- **GET /api/v1/games/{gameId}** - Get a game
- **GET /api/v1/games/{gameId}/clips** - List visible game clips
- **POST /api/v1/games/{gameId}/follow** - Follow a game
- **DELETE /api/v1/games/{gameId}/follow** - Unfollow a game
- **GET /api/v1/users/{id}/games/following** - List games followed by a user

### Twitch Categories (6)

- **GET /api/v1/twitch-categories/trending** - List trending Twitch categories
- **GET /api/v1/twitch-categories/{gameId}** - Get a Twitch category
- **GET /api/v1/twitch-categories/{gameId}/clips** - List visible clips in a Twitch category
- **POST /api/v1/twitch-categories/{gameId}/follow** - Follow a Twitch category
- **DELETE /api/v1/twitch-categories/{gameId}/follow** - Unfollow a Twitch category
- **GET /api/v1/users/{id}/twitch-categories/following** - List Twitch categories followed by a user

### Notifications (11)

- **GET /api/v1/notifications** - List notifications for the authenticated user
- **GET /api/v1/notifications/count** - Get the authenticated user's unread notification count
- **GET /api/v1/notifications/preferences** - Get notification preferences
- **PUT /api/v1/notifications/preferences** - Replace notification preferences
- **POST /api/v1/notifications/preferences/reset** - Reset notification preferences to defaults
- **PUT /api/v1/notifications/read-all** - Mark every notification as read
- **POST /api/v1/notifications/register** - Register a push-notification device token
- **DELETE /api/v1/notifications/unregister** - Unregister the matching push-notification device token
- **GET /api/v1/notifications/unsubscribe** - Consume a one-time email unsubscribe token
- **DELETE /api/v1/notifications/{id}** - Delete an owned notification
- **PUT /api/v1/notifications/{id}/read** - Mark an owned notification as read

### Playlist Scripts (5)

- **GET /api/v1/playlist-scripts** - List playlist scripts owned by the authenticated user
- **POST /api/v1/playlist-scripts** - Create an owned standard playlist script
- **PUT /api/v1/playlist-scripts/{id}** - Update an owned playlist script
- **DELETE /api/v1/playlist-scripts/{id}** - Delete an owned playlist script
- **POST /api/v1/playlist-scripts/{id}/generate** - Atomically generate a playlist from an owned active script

### Playlists (24)

- **GET /api/v1/playlists** - List playlists owned by the authenticated user
- **POST /api/v1/playlists** - Create a playlist
- **GET /api/v1/playlists/bookmarks** - List the authenticated user's bookmarked playlists
- **GET /api/v1/playlists/featured** - List current public featured or curated playlists
- **GET /api/v1/playlists/public** - List public playlists
- **GET /api/v1/playlists/share/{token}** - Get a playlist through its share token
- **GET /api/v1/playlists/today** - Get the most recently generated active daily public playlist
- **GET /api/v1/playlists/{id}** - Get a visible playlist and paginated clips
- **PATCH /api/v1/playlists/{id}** - Update a playlist when the caller has edit permission
- **DELETE /api/v1/playlists/{id}** - Soft-delete an owned playlist
- **POST /api/v1/playlists/{id}/bookmark** - Idempotently bookmark a visible playlist
- **DELETE /api/v1/playlists/{id}/bookmark** - Idempotently remove the caller's playlist bookmark
- **POST /api/v1/playlists/{id}/clips** - Atomically add clips to a playlist
- **PUT /api/v1/playlists/{id}/clips/order** - Atomically reorder every clip in a playlist
- **DELETE /api/v1/playlists/{id}/clips/{clip_id}** - Atomically remove and reindex a playlist clip
- **GET /api/v1/playlists/{id}/collaborators** - List collaborators for a visible playlist
- **POST /api/v1/playlists/{id}/collaborators** - Add or replace a collaborator permission
- **PATCH /api/v1/playlists/{id}/collaborators/{user_id}** - Update an existing collaborator permission
- **DELETE /api/v1/playlists/{id}/collaborators/{user_id}** - Remove an existing collaborator
- **POST /api/v1/playlists/{id}/copy** - Atomically copy a visible playlist and its ordered clips
- **POST /api/v1/playlists/{id}/like** - Idempotently like a visible playlist
- **DELETE /api/v1/playlists/{id}/like** - Idempotently remove the caller's playlist like
- **GET /api/v1/playlists/{id}/share-link** - Get or atomically create the canonical share link
- **POST /api/v1/playlists/{id}/track-share** - Atomically record a public playlist share event

### Queue (8)

- **GET /api/v1/queue** - Get the active queue
- **POST /api/v1/queue** - Add a clip to the queue
- **DELETE /api/v1/queue** - Clear the active queue
- **POST /api/v1/queue/convert-to-playlist** - Convert queue items to a private playlist
- **GET /api/v1/queue/count** - Count active queue items
- **PATCH /api/v1/queue/reorder** - Move an active queue item
- **DELETE /api/v1/queue/{id}** - Remove a queue item
- **POST /api/v1/queue/{id}/played** - Mark a queue item as played

### Recommendations (3)

- **POST /api/v1/recommendations/onboarding** - Complete recommendation onboarding
- **GET /api/v1/recommendations/preferences** - Get recommendation preferences
- **PUT /api/v1/recommendations/preferences** - Update recommendation preferences

### Streamer Clip Rooms (9)

- **GET /api/v1/streamer-clip-rooms/{channel}** - Get or create a channel clip room
- **GET /api/v1/streamer-clip-rooms/{channel}/items** - List room items
- **PUT /api/v1/streamer-clip-rooms/{channel}/items/order** - Reorder every approved room item
- **POST /api/v1/streamer-clip-rooms/{channel}/items/{itemId}/approve** - Approve a room item
- **POST /api/v1/streamer-clip-rooms/{channel}/items/{itemId}/reject** - Reject a room item
- **POST /api/v1/streamer-clip-rooms/{channel}/start** - Start the Twitch chat listener
- **POST /api/v1/streamer-clip-rooms/{channel}/stop** - Stop the Twitch chat listener
- **PUT /api/v1/streamer-clip-rooms/{channel}/submissions** - Open or close viewer clip submissions
- **GET /api/v1/streamer-clip-rooms/{channel}/ws** - Open the read-only room event websocket

### Twitch OAuth (4)

- **DELETE /api/v1/twitch/auth** - Revoke stored Twitch authentication
- **GET /api/v1/twitch/auth/status** - Get Twitch authentication status
- **GET /api/v1/twitch/oauth/authorize** - Begin user-bound Twitch OAuth
- **GET /api/v1/twitch/oauth/callback** - Complete user-bound Twitch OAuth

### Filter Presets (5)

- **GET /api/v1/users/{id}/filter-presets** - List the authenticated user's filter presets
- **POST /api/v1/users/{id}/filter-presets** - Create a filter preset
- **GET /api/v1/users/{id}/filter-presets/{presetId}** - Get a filter preset
- **PUT /api/v1/users/{id}/filter-presets/{presetId}** - Update a filter preset
- **DELETE /api/v1/users/{id}/filter-presets/{presetId}** - Delete a filter preset

### SEO (1)

- **GET /clips/best/{path}** - Render a best-clips period page

### Debug (11)

- **GET /debug/pprof** - Open the Go profiler index in debug mode
- **GET /debug/pprof/allocs** - Get allocation profile in debug mode
- **GET /debug/pprof/block** - Get block profile in debug mode
- **GET /debug/pprof/cmdline** - Get process command line in debug mode
- **GET /debug/pprof/goroutine** - Get goroutine profile in debug mode
- **GET /debug/pprof/heap** - Get heap profile in debug mode
- **GET /debug/pprof/mutex** - Get mutex profile in debug mode
- **GET /debug/pprof/profile** - Capture CPU profile in debug mode
- **GET /debug/pprof/symbol** - Resolve profiler symbols in debug mode
- **GET /debug/pprof/threadcreate** - Get thread creation profile in debug mode
- **GET /debug/pprof/trace** - Capture execution trace in debug mode

### Service Status (2)

- **GET /internal/operations/database** - Get database pool statistics
- **GET /internal/operations/metrics** - Scrape operational metrics
