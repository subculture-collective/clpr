-- Tag listings filtered every row through is_tag_blacklisted(), which ran the
-- PL/pgSQL glob converter once per tag per pattern. With ~29k tags and 25
-- patterns, GET /api/v1/tags spent over four seconds counting visible tags.
-- Store the converted pattern once, and store exact-match patterns as a
-- lowercase literal so listings can hash-join against them.

ALTER TABLE blacklisted_tags
    ADD COLUMN like_pattern TEXT
        GENERATED ALWAYS AS (tag_blacklist_glob_to_like(pattern)) STORED,
    ADD COLUMN literal_slug TEXT
        GENERATED ALWAYS AS (
            CASE WHEN strpos(pattern, '*') = 0 AND strpos(pattern, '?') = 0
                 THEN lower(pattern)
            END
        ) STORED;

CREATE OR REPLACE FUNCTION is_tag_blacklisted(candidate_slug TEXT)
RETURNS BOOLEAN
LANGUAGE sql
STABLE
STRICT
AS $$
    SELECT EXISTS (
        SELECT 1 FROM blacklisted_tags
        WHERE literal_slug = lower(candidate_slug)
    ) OR EXISTS (
        SELECT 1 FROM blacklisted_tags
        WHERE literal_slug IS NULL
          AND candidate_slug ILIKE like_pattern ESCAPE '\'
    );
$$;
