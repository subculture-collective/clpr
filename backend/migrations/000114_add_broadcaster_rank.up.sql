-- Materialized view for broadcaster ranking
-- Weighs human engagement (submissions, comments, votes) heavier than auto-scraped content
CREATE MATERIALIZED VIEW IF NOT EXISTS broadcaster_rankings AS
SELECT
    c.broadcaster_id,
    MAX(c.broadcaster_name) AS broadcaster_name,
    COUNT(DISTINCT c.id) AS total_clips,
    COUNT(DISTINCT CASE WHEN c.submitted_by_user_id IS NOT NULL THEN c.id END) AS human_submitted_clips,
    COALESCE(SUM(c.vote_score), 0) AS total_vote_score,
    COALESCE(SUM(c.view_count), 0) AS total_views,
    COALESCE(comment_counts.total_comments, 0) AS total_comments,
    COALESCE(comment_counts.unique_commenters, 0) AS unique_commenters,
    -- Ranking formula: human submissions 5x, comments 2x, votes 1x, auto-scraped 0.5x
    (
        COUNT(DISTINCT CASE WHEN c.submitted_by_user_id IS NOT NULL THEN c.id END) * 5.0
        + COALESCE(comment_counts.total_comments, 0) * 2.0
        + GREATEST(COALESCE(SUM(c.vote_score), 0), 0) * 1.0
        + COUNT(DISTINCT c.id) * 0.5
    ) AS engagement_score,
    COUNT(DISTINCT bf.user_id) AS follower_count,
    NOW() AS last_calculated
FROM clips c
LEFT JOIN broadcaster_follows bf ON bf.broadcaster_id = c.broadcaster_id
LEFT JOIN (
    SELECT
        cl.broadcaster_id,
        COUNT(cm.id) AS total_comments,
        COUNT(DISTINCT cm.user_id) AS unique_commenters
    FROM comments cm
    JOIN clips cl ON cl.id = cm.clip_id
    WHERE cm.is_removed = false
    GROUP BY cl.broadcaster_id
) comment_counts ON comment_counts.broadcaster_id = c.broadcaster_id
WHERE c.broadcaster_id IS NOT NULL
  AND c.broadcaster_id != ''
GROUP BY c.broadcaster_id, comment_counts.total_comments, comment_counts.unique_commenters
ORDER BY engagement_score DESC;

CREATE UNIQUE INDEX idx_broadcaster_rankings_id ON broadcaster_rankings(broadcaster_id);
CREATE INDEX idx_broadcaster_rankings_score ON broadcaster_rankings(engagement_score DESC);

-- Function to refresh the materialized view
CREATE OR REPLACE FUNCTION refresh_broadcaster_rankings()
RETURNS void AS $$
BEGIN
    REFRESH MATERIALIZED VIEW CONCURRENTLY broadcaster_rankings;
END;
$$ LANGUAGE plpgsql;
