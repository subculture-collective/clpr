-- Publish every staged discovery clip into the main feed. The staging table is
-- retained empty for one compatibility release so rollback and old binaries do
-- not fail during a rolling deployment.
BEGIN;

INSERT INTO clips (
    id, twitch_clip_id, twitch_clip_url, embed_url, title,
    creator_name, creator_id, broadcaster_name, broadcaster_id,
    game_id, game_name, language, thumbnail_url, duration,
    view_count, created_at, imported_at, vote_score, comment_count,
    favorite_count, is_featured, is_nsfw, is_removed, is_hidden,
    submitted_by_user_id, submitted_at
)
SELECT
    id, twitch_clip_id, twitch_clip_url, embed_url, title,
    creator_name, creator_id, broadcaster_name, broadcaster_id,
    game_id, game_name, language, thumbnail_url, duration,
    view_count, created_at, imported_at, 0, 0,
    0, FALSE, is_nsfw, is_removed, is_hidden,
    NULL, NULL
FROM discovery_clips
ON CONFLICT (twitch_clip_id) DO UPDATE SET
    twitch_clip_url = EXCLUDED.twitch_clip_url,
    embed_url = EXCLUDED.embed_url,
    title = EXCLUDED.title,
    creator_name = EXCLUDED.creator_name,
    creator_id = EXCLUDED.creator_id,
    broadcaster_name = EXCLUDED.broadcaster_name,
    broadcaster_id = EXCLUDED.broadcaster_id,
    game_id = EXCLUDED.game_id,
    game_name = COALESCE(EXCLUDED.game_name, clips.game_name),
    language = EXCLUDED.language,
    thumbnail_url = EXCLUDED.thumbnail_url,
    duration = EXCLUDED.duration,
    view_count = GREATEST(clips.view_count, EXCLUDED.view_count);

DELETE FROM discovery_clips;

COMMIT;
