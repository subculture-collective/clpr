-- Roll back creator account foundations.

DROP INDEX IF EXISTS idx_clip_submissions_creator_account_id;
DROP INDEX IF EXISTS idx_clips_creator_account_id;
DROP INDEX IF EXISTS idx_creator_platform_accounts_creator_id;
DROP INDEX IF EXISTS idx_creator_accounts_slug;
DROP INDEX IF EXISTS idx_creator_accounts_owner_user_id;

ALTER TABLE clip_submissions
    DROP COLUMN IF EXISTS creator_account_id;

ALTER TABLE clips
    DROP COLUMN IF EXISTS creator_account_id;

DROP TABLE IF EXISTS creator_platform_accounts;
DROP TABLE IF EXISTS creator_accounts;
