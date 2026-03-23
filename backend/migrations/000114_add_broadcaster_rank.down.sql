DROP FUNCTION IF EXISTS refresh_broadcaster_rankings();
DROP INDEX IF EXISTS idx_broadcaster_rankings_score;
DROP INDEX IF EXISTS idx_broadcaster_rankings_id;
DROP MATERIALIZED VIEW IF EXISTS broadcaster_rankings;
