ALTER TABLE playlist_scripts
    DROP CONSTRAINT IF EXISTS chk_playlist_scripts_strategy;

ALTER TABLE playlist_scripts
    ADD CONSTRAINT chk_playlist_scripts_strategy
    CHECK (strategy IN (
        'standard', 'sleeper_hits', 'viral_velocity', 'community_favorites',
        'deep_cuts', 'fresh_faces', 'one_per_creator', 'diversity_roulette',
        'clip_of_the_day', 'weekend_mix', 'similar_vibes', 'cross_game_hits',
        'controversial', 'binge_worthy', 'rising_stars', 'twitch_top_game',
        'twitch_top_broadcaster', 'twitch_trending', 'twitch_discovery'
    ));
