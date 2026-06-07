-- Add local creator moderation tables.

CREATE TABLE creator_moderators (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    creator_id UUID NOT NULL REFERENCES creator_accounts(id) ON DELETE CASCADE,
    user_id UUID REFERENCES users(id) ON DELETE CASCADE,
    platform VARCHAR(30),
    platform_user_id VARCHAR(255),
    permissions TEXT[] NOT NULL DEFAULT ARRAY[]::TEXT[],
    source VARCHAR(30) NOT NULL DEFAULT 'manual',
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    CONSTRAINT creator_moderators_source_check CHECK (source IN ('manual', 'imported')),
    CONSTRAINT creator_moderators_platform_check CHECK (platform IS NULL OR platform IN ('twitch', 'kick', 'youtube', 'tiktok')),
    CONSTRAINT creator_moderators_permissions_check CHECK (permissions <@ ARRAY['manage_creator_clips', 'approve_creator_submissions', 'remove_creator_comments', 'ban_creator_users', 'sync_platform_bans']::TEXT[]),
    CONSTRAINT creator_moderators_identity_check CHECK (
        (user_id IS NOT NULL OR (platform IS NOT NULL AND platform_user_id IS NOT NULL))
        AND ((platform IS NULL) = (platform_user_id IS NULL))
    )
);

CREATE UNIQUE INDEX IF NOT EXISTS uq_creator_moderators_creator_user
    ON creator_moderators (creator_id, user_id)
    WHERE user_id IS NOT NULL;

CREATE UNIQUE INDEX IF NOT EXISTS uq_creator_moderators_creator_platform_user
    ON creator_moderators (creator_id, platform, platform_user_id)
    WHERE platform IS NOT NULL AND platform_user_id IS NOT NULL;

CREATE INDEX IF NOT EXISTS idx_creator_moderators_creator_id
    ON creator_moderators (creator_id);

CREATE TABLE creator_bans (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    creator_id UUID NOT NULL REFERENCES creator_accounts(id) ON DELETE CASCADE,
    target_user_id UUID REFERENCES users(id) ON DELETE CASCADE,
    target_platform VARCHAR(30),
    target_platform_user_id VARCHAR(255),
    source VARCHAR(30) NOT NULL DEFAULT 'manual',
    reason TEXT,
    scopes TEXT[] NOT NULL DEFAULT ARRAY['interact', 'submit', 'comment']::TEXT[],
    expires_at TIMESTAMPTZ,
    created_by_user_id UUID REFERENCES users(id) ON DELETE SET NULL,
    sync_status VARCHAR(30) NOT NULL DEFAULT 'local_only',
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    CONSTRAINT creator_bans_source_check CHECK (source IN ('manual', 'imported')),
    CONSTRAINT creator_bans_platform_check CHECK (target_platform IS NULL OR target_platform IN ('twitch', 'kick', 'youtube', 'tiktok')),
    CONSTRAINT creator_bans_scopes_check CHECK (scopes <@ ARRAY['interact', 'submit', 'comment']::TEXT[]),
    CONSTRAINT creator_bans_sync_status_check CHECK (sync_status = 'local_only'),
    CONSTRAINT creator_bans_identity_check CHECK (
        (target_user_id IS NOT NULL OR (target_platform IS NOT NULL AND target_platform_user_id IS NOT NULL))
        AND ((target_platform IS NULL) = (target_platform_user_id IS NULL))
    )
);

CREATE INDEX IF NOT EXISTS idx_creator_bans_active_user_lookup
    ON creator_bans (creator_id, target_user_id, expires_at)
    WHERE target_user_id IS NOT NULL;

CREATE INDEX IF NOT EXISTS idx_creator_bans_active_platform_lookup
    ON creator_bans (creator_id, target_platform, target_platform_user_id, expires_at)
    WHERE target_platform IS NOT NULL AND target_platform_user_id IS NOT NULL;

CREATE INDEX IF NOT EXISTS idx_creator_bans_scopes
    ON creator_bans USING GIN (scopes);
