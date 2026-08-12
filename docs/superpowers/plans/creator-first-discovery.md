# Creator-First Discovery

## Product direction

Clpr should be a place to discover **creators and moments**, not a catalog of games.

The organizing hierarchy should be:

1. **Creator** — who made the moment
2. **Topic** — what the moment is about
3. **Moment type** — why the clip is interesting
4. **Twitch category** — what Twitch classified the stream as at that time

Twitch category data remains useful metadata, but it should no longer define the product's identity, navigation, or recommendation model.

A concise positioning statement:

> Discover the people and moments shaping live culture.

## Why this fits the current catalog

Production snapshot on August 11, 2026:

- 1,176 visible clips
- 522 distinct creators/broadcasters
- 106 distinct Twitch category IDs
- The most-viewed creators commonly span multiple Twitch categories: `xQc` spans 6, `ohnePixel` 5, `caseoh_` 7, and `Maximilian_DOOD` 4.
- All current visible clips have a missing `game_name`, even though most have a `game_id`. Creator identity is currently the more complete discovery primitive.

The codebase already contains much of the needed creator infrastructure:

- Broadcaster profiles, clips, follows, rankings, and live status
- Creator analytics and profile routes
- Popular creator navigation
- Topics, tags, playlists, and curated collections

However, the user experience still frames Clpr as gaming-first in several places:

- The sidebar has a prominent Games section.
- Search treats Games as a top-level result type and says “Search clips, games, creators.”
- Onboarding models `favorite_games` before creators and interests.
- Category pages primarily list games within a category.
- Clip metadata links Twitch categories as games.
- Homepage, feed, About, Terms, Community Rules, submission, and SEO copy repeatedly say “gaming moments” or “gaming community.”
- Several playlist strategies use game diversity as their main definition of variety.

## Approaches considered

### 1. Copy-only rebrand

Replace “gaming” with “streaming,” rename Games to Categories, and leave the information architecture unchanged.

This is fast, but cosmetic. Users would still be sent through a game-shaped discovery system, and non-gaming topics would remain constrained by game mappings.

### 2. Creator-first presentation with compatible internals — recommended

Make creators and topics primary in the interface and rankings while retaining Twitch `game_id` fields internally for API compatibility. Introduce additive Twitch-category terminology and direct clip-to-topic classification over time.

This provides an immediate product shift without a dangerous all-at-once database rename.

### 3. Full taxonomy replacement

Rename the games table and every API, route, filter, preference, and index to Twitch categories in one migration.

This produces the cleanest eventual schema but has a large blast radius and does not itself solve clip-level topics. It should be the end state of a gradual compatibility migration, not the first move.

## Recommended information architecture

### Primary discovery axes

- **Creators**: trending, rising, live, followed, and newly discovered
- **Topics**: News & Politics, IRL & Travel, Reactions & Commentary, Gaming, Music & Performance, Sports, Creative & Making, Tech, and Culture & Drama
- **Moments**: Funny, Heated, Insightful, Wholesome, Wild, Clutch, Fail, Debate, Breaking, and Performance
- **Collections**: editorial and automated playlists that can combine creators, topics, and moment types

“Gaming” remains a valid topic. It stops being the container for everything else.

### Naming rules

- Use **Creator** in user-facing text.
- Retain **broadcaster** where it is a Twitch-specific technical identifier.
- Use **Twitch category** for Twitch's `game_id`/`game_name` concept.
- Use **Topic** for Clpr's semantic classification of what a clip is about.
- Use **Tag** for specific, community-readable descriptors.

This avoids conflating three different people: the channel creator, the person who created the Twitch clip, and the user who submitted it to Clpr.

## Experience changes

### Navigation and homepage

- Make Creators the default tab in the horizontal discovery rail.
- Rename Streamers to Creators everywhere user-facing.
- Replace the Games sidebar block with one of:
  - Live Creators
  - Trending Creators
  - Rising Creators
- Move Topics above Tags in discovery surfaces.
- Keep Twitch categories available as a secondary filter, not a primary homepage section.
- Use creator avatar, creator name, live state, and follow action as the strongest metadata on clip cards.
- Show the topic or moment label after creator identity; show Twitch category only when helpful.

### Search

Primary result types should be:

- Clips
- Creators
- Topics
- Tags

Twitch categories can remain searchable under an advanced filter or a secondary “Categories” result group during migration.

Examples should shift from `game:valorant` to combinations such as:

- `creator:xqc topic:politics`
- `topic:irl tag:funny`
- `creator:caseoh tag:reaction`
- `category:valorant` for the explicit Twitch-category filter

### Onboarding and personalization

Ask new users to:

1. Follow creators
2. Choose topics
3. Choose moment types or tags

Favorite games should become an optional Twitch-category preference, not the lead question.

Recommendation weighting should prioritize:

1. Followed and adjacent creators
2. Topic affinity
3. Moment/tag affinity
4. Community response and velocity
5. Twitch-category affinity

### Creator pages

Creator pages should become the main content hubs and include:

- Best, newest, and rising clips
- Current live status
- Topic and moment breakdown
- Recent category/activity changes
- Related creators based on audience/topic overlap
- Creator-specific collections

The existing broadcaster profile is a strong base, but its user-facing title and actions should consistently say Creator.

### Curated collections

Automated collections should enforce creator diversity first and topic diversity second.

- **Diversity Roulette** should become **Creator & Topic Roulette** or **Across the Culture**.
- **Weekend Mix** should keep one clip per creator and use a soft topic cap instead of a game cap.
- Add **Creator Spotlight**, **Rising Creators**, **Outside Your Bubble**, and topic-driven collections such as **News & Politics Today** or **IRL This Week**.
- Twitch category can remain one input to ranking, but not the diversity definition.

## Data model evolution

### Do not immediately delete or rename `game_id`

Twitch still calls this field `game_id`, and the ingestion stack depends on it. Keep it as a compatibility field initially.

Add clearer aliases and concepts:

- `twitch_category_id`
- `twitch_category_name`
- `clip_topics` many-to-many relationship
- optional `clip_moments` relationship, or reuse structured tags for moment types

The API can expose additive `twitch_category_*` fields while retaining `game_*` fields through a deprecation period.

### Direct clip-to-topic classification is essential

The existing topic model is not enough. Category feeds currently derive clip membership through `category_games`, which means a News or Politics clip is classified according to the Twitch category of the entire stream. That fails for variety creators and mixed streams.

Introduce direct topic membership:

```text
clip
 ├── creator/broadcaster
 ├── Twitch category (upstream metadata)
 ├── topics[] (what this clip is about)
 └── tags[] / moments[] (why it is interesting)
```

Topic classification can combine:

- Twitch category
- Existing tags
- Clip title
- Authorized transcript/Whisper output
- Stream title at clip time, when available
- Manual moderation corrections

Store classification confidence and source so low-confidence automated assignments can be reviewed or omitted.

## Rollout plan

### Phase 1 — Reposition without schema risk

- Replace gaming-focused global copy and SEO descriptions.
- Rename user-facing Streamers/Broadcasters to Creators.
- Default the discovery rail to Creators.
- Remove Games from the homepage sidebar and emphasize Creators and Topics.
- Rename game labels on clip metadata to Twitch Category where they remain visible.
- Update collection descriptions and diversity labels.
- Preserve all existing routes and API fields.

This is the fastest visible change and can ship independently.

### Phase 2 — Creator-first discovery and personalization

- Add dedicated Trending, Rising, Live, and New Creator rails.
- Reorder search around Clips, Creators, Topics, and Tags.
- Update onboarding to creators/topics/tags.
- Increase followed-creator and creator-adjacency weight in recommendations.
- Change automated collection constraints from game-first to creator/topic-first.

### Phase 3 — Real topic intelligence

- Add direct `clip_topics` storage with source and confidence.
- Classify new clips during enrichment using metadata, tags, and transcripts where authorized.
- Backfill existing clips.
- Make Topic pages query direct clip membership.
- Add moderation tools for topic correction and merge/split operations.

### Phase 4 — Technical terminology migration

- Add `twitch_category_*` API fields and query parameters.
- Add `/twitch-category/:id` or `/category/:id` routes while redirecting `/game/:id`.
- Rename search index concepts and internal types gradually.
- Deprecate game-named API fields only after all clients use the new aliases.

## Initial implementation slice

The first deployable slice should include:

1. Creator-first copy across global SEO, About, submission, feeds, rules, and search.
2. Creators as the first/default discovery rail tab.
3. Games removed from the feed sidebar; Trending/Live Creators promoted.
4. “Streamers” and “Broadcasters” changed to “Creators” in user-facing UI.
5. Search copy and tabs reordered to Clips, Creators, Topics, Tags, with Games relabeled Twitch Categories temporarily.
6. Playlist descriptions and admin labels changed from game diversity to creator/topic diversity where the strategy actually supports it.

The last item should not falsely relabel current SQL behavior. Ranking queries must be changed before their descriptions claim topic diversity.

## Success measures

- Creator profile click-through rate from the feed
- Creator follows per active user
- Percentage of sessions that include clips from three or more creators
- Topic-page engagement and continuation rate
- Repeat viewing of the same creator across sessions
- Reduction in discovery interactions that depend on a game/category page
- New-user activation after creator/topic onboarding
- Coverage and correction rate of automated clip topics

## Risks and guardrails

- **Creator concentration:** creator-first ranking can over-amplify already large creators. Retain per-creator caps and explicit Fresh Faces/Rising Creators surfaces.
- **Topic ambiguity:** a clip can span multiple topics. Use many-to-many membership rather than one required category.
- **Automated political classification:** News and Politics labels need confidence thresholds and transparent sources; avoid inferring political beliefs about creators or users.
- **Compatibility:** do not break Twitch ingestion or old URLs merely to clean up terminology.
- **Empty topics:** do not feature topic pages until they have direct, visible clip membership.
- **Gaming audience:** gaming remains a first-class topic and filter; it is not removed, only placed alongside the rest of live culture.

## Decision

Adopt creator-first presentation with compatible internals.

The product should treat **people as durable**, **topics as semantic**, and **Twitch categories as transient upstream metadata**. This reflects how variety creators actually behave and gives Clpr room to cover IRL, reactions, news, politics, music, sports, and gaming without pretending they are all games.
