-- Restore the 000142 matcher before removing the columns it no longer reads.
CREATE OR REPLACE FUNCTION is_tag_blacklisted(candidate_slug TEXT)
RETURNS BOOLEAN
LANGUAGE sql
STABLE
STRICT
AS $$
    SELECT EXISTS (
        SELECT 1
        FROM blacklisted_tags
        WHERE candidate_slug ILIKE tag_blacklist_glob_to_like(pattern) ESCAPE '\'
    );
$$;

ALTER TABLE blacklisted_tags
    DROP COLUMN IF EXISTS literal_slug,
    DROP COLUMN IF EXISTS like_pattern;
