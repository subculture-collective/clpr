-- Drop indexes
DROP INDEX IF EXISTS idx_clips_recommendation_lookup;
DROP INDEX IF EXISTS idx_user_interactions_type;
DROP INDEX IF EXISTS idx_user_interactions_clip;
DROP INDEX IF EXISTS idx_user_interactions_user;
DROP INDEX IF EXISTS idx_user_preferences_tags;
DROP INDEX IF EXISTS idx_user_preferences_categories;
DROP INDEX IF EXISTS idx_user_preferences_streamers;
DROP INDEX IF EXISTS idx_user_preferences_games;

-- Drop triggers before the functions they execute
DROP TRIGGER IF EXISTS trigger_update_user_preferences_timestamp ON user_preferences;

-- Drop functions
DROP FUNCTION IF EXISTS update_user_preferences_from_interactions(UUID);
DROP FUNCTION IF EXISTS update_user_preferences_timestamp();

-- Drop tables
DROP TABLE IF EXISTS user_clip_interactions;
DROP TABLE IF EXISTS user_preferences;
