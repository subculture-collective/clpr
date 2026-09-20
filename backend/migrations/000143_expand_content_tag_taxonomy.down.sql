-- Preserve any tags already attached to clips. A rollback removes only newly
-- seeded, unused vocabulary; model validation is rolled back with the binary.
DELETE FROM tags t
WHERE t.slug IN (
    'content/gameplay', 'content/facecam', 'content/vtuber',
    'content/screen-share', 'content/conversation', 'content/interview',
    'content/podcast', 'content/commentary', 'content/review',
    'content/demonstration', 'content/event', 'content/crowd', 'content/stage',
    'content/charity', 'content/music-performance', 'content/singing',
    'content/instrument', 'content/dance', 'content/art', 'content/drawing',
    'content/cosplay', 'content/cooking', 'content/baking', 'content/crafting',
    'content/maker', 'content/coding', 'content/fitness', 'content/sports',
    'content/travel', 'content/outdoors', 'content/animals', 'content/unboxing',
    'content/pack-opening',
    'content/esports', 'content/tournament', 'content/ranked',
    'content/casual-play', 'content/multiplayer', 'content/cooperative',
    'content/pvp', 'content/pve', 'content/boss-fight', 'content/puzzle',
    'content/exploration', 'content/building', 'content/roleplay',
    'content/simulation', 'content/racing-game', 'content/fighting-game',
    'content/sports-game', 'content/strategy-game', 'content/horror-game',
    'content/retro-game', 'content/shooter-game', 'content/moba',
    'content/battle-royale', 'content/survival-game', 'content/sandbox-game',
    'content/card-game', 'content/cutscene', 'content/victory', 'content/defeat',
    'content/comeback', 'content/close-call', 'content/elimination',
    'content/team-wipe', 'content/record', 'content/personal-best',
    'content/boss-kill', 'content/jump-scare', 'content/trick-shot',
    'content/achievement', 'content/discovery', 'content/challenge',
    'content/reveal', 'content/celebration', 'content/wholesome',
    'content/emotional', 'content/surprise', 'content/scary'
)
AND NOT EXISTS (SELECT 1 FROM clip_tags ct WHERE ct.tag_id = t.id);
