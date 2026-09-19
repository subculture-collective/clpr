DROP VIEW IF EXISTS tag_promotion_candidates;

CREATE VIEW tag_promotion_candidates AS
SELECT t.slug, t.name, COUNT(DISTINCT ct.clip_id) AS clip_count,
       COUNT(DISTINCT c.submitted_by_user_id) AS unique_users, t.parent_slug
FROM clip_tags ct
JOIN tags t ON t.id = ct.tag_id
JOIN clips c ON c.id = ct.clip_id
WHERE (t.parent_slug IS NULL OR t.parent_slug = 'community')
  AND c.submitted_by_user_id IS NOT NULL
GROUP BY t.slug, t.name, t.parent_slug
HAVING COUNT(DISTINCT c.submitted_by_user_id) >= 3
   AND COUNT(DISTINCT ct.clip_id) >= 5;

DROP TABLE IF EXISTS clip_score_refresh_stage;
DROP TABLE IF EXISTS tag_aliases;
DROP FUNCTION IF EXISTS is_tag_blacklisted(TEXT);
DROP FUNCTION IF EXISTS tag_blacklist_glob_to_like(TEXT);
