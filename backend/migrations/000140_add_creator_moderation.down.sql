-- Roll back local creator moderation tables.

DROP INDEX IF EXISTS idx_creator_bans_scopes;
DROP INDEX IF EXISTS idx_creator_bans_active_platform_lookup;
DROP INDEX IF EXISTS idx_creator_bans_active_user_lookup;
DROP INDEX IF EXISTS idx_creator_moderators_creator_id;
DROP INDEX IF EXISTS uq_creator_moderators_creator_platform_user;
DROP INDEX IF EXISTS uq_creator_moderators_creator_user;

DROP TABLE IF EXISTS creator_bans;
DROP TABLE IF EXISTS creator_moderators;
