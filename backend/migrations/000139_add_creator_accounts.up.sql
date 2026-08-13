-- Add creator account foundations.
--
-- NOTE: clips and clip_submissions already have a legacy source metadata column
-- named creator_id (Twitch/source creator ID). To preserve that existing data
-- model, the creator-account foreign key uses creator_account_id here.

CREATE TABLE creator_accounts (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    owner_user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    display_name VARCHAR(255) NOT NULL,
    slug VARCHAR(255) UNIQUE NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_creator_accounts_owner_user_id
    ON creator_accounts(owner_user_id);
CREATE INDEX IF NOT EXISTS idx_creator_accounts_slug
    ON creator_accounts(slug);

CREATE TABLE creator_platform_accounts (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    creator_id UUID NOT NULL REFERENCES creator_accounts(id) ON DELETE CASCADE,
    platform VARCHAR(30) NOT NULL,
    platform_user_id VARCHAR(255) NOT NULL,
    platform_display_name VARCHAR(255) NOT NULL,
    profile_url TEXT,
    can_import_bans BOOLEAN NOT NULL DEFAULT false,
    can_sync_bans_outbound BOOLEAN NOT NULL DEFAULT false,
    can_import_moderators BOOLEAN NOT NULL DEFAULT false,
    can_verify_ownership BOOLEAN NOT NULL DEFAULT false,
    can_fetch_metadata BOOLEAN NOT NULL DEFAULT true,
    access_token_encrypted TEXT,
    refresh_token_encrypted TEXT,
    token_expires_at TIMESTAMPTZ,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    CONSTRAINT creator_platform_accounts_platform_check
        CHECK (platform IN ('twitch', 'kick', 'youtube', 'tiktok')),
    CONSTRAINT creator_platform_accounts_unique_platform_user
        UNIQUE (platform, platform_user_id)
);

CREATE INDEX IF NOT EXISTS idx_creator_platform_accounts_creator_id
    ON creator_platform_accounts(creator_id);

ALTER TABLE clips
    ADD COLUMN IF NOT EXISTS creator_account_id UUID REFERENCES creator_accounts(id) ON DELETE SET NULL;

ALTER TABLE clip_submissions
    ADD COLUMN IF NOT EXISTS creator_account_id UUID REFERENCES creator_accounts(id) ON DELETE SET NULL;

CREATE INDEX IF NOT EXISTS idx_clips_creator_account_id
    ON clips(creator_account_id)
    WHERE creator_account_id IS NOT NULL;
CREATE INDEX IF NOT EXISTS idx_clip_submissions_creator_account_id
    ON clip_submissions(creator_account_id)
    WHERE creator_account_id IS NOT NULL;
