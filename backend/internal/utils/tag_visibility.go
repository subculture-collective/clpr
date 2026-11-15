package utils

// VisibleTagPredicate excludes blacklisted and suppressed tags for the tags
// row aliased as alias. It matches is_tag_blacklisted() but is written inline
// so PostgreSQL hash-joins exact-match patterns (blacklisted_tags.literal_slug,
// migration 000144) instead of calling the function once per tag.
func VisibleTagPredicate(alias string) string {
	return `NOT EXISTS (SELECT 1 FROM blacklisted_tags b WHERE b.literal_slug = lower(` + alias + `.slug))
		  AND NOT EXISTS (SELECT 1 FROM blacklisted_tags b WHERE b.literal_slug IS NULL AND ` + alias + `.slug ILIKE b.like_pattern ESCAPE '\')
		  AND NOT EXISTS (SELECT 1 FROM tag_suppressions s WHERE s.tag_id = ` + alias + `.id)`
}
