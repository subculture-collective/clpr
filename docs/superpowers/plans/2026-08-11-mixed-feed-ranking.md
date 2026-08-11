# Mixed Feed Ranking Exploration

## Seed

The main feed now combines community submissions and automatically discovered
Twitch clips. Automated clips need a changing but coherent order based on
audience evidence, while community submissions should retain a meaningful
distribution advantage.

## Defaults enumerated

- Sort only by total Twitch views.
- Sort by creation time.
- Use PostgreSQL `random()` on every request.
- Keep the existing engagement-over-age trending score.
- Put every community submission ahead of every automated clip.

## Core functions

- Reward demonstrated audience interest without letting channel size fully
  determine the feed.
- Detect clips whose view count is increasing quickly.
- Give automated clips bounded opportunities for exploration.
- Preserve a material community-submission boost.
- Keep ordering stable within a cursor-paginated feed session.

## Alternatives explored

| Approach | Strength | Failure mode |
| --- | --- | --- |
| Request-time random shuffle | Maximum variety | Duplicate/missing cursor pages and noisy refreshes |
| Fixed weighted raw score | Simple | A few very large channels dominate because signals have different scales |
| Editorial slot interleaving | Strong source guarantees | More query machinery and abrupt quality cliffs between slots |
| Percentile mixed rank | Scale-independent and tunable | Requires periodic score refresh and view history |

## Selected design

Use a periodically materialized quality rank plus a stateless feed-session
shuffle:

- 30% log-scaled Twitch view-count percentile.
- 35% view-velocity percentile, measured from successive Twitch observations.
- 15% freshness with a 24-hour smooth decay.
- 20% deterministic exploration derived from the automated clip ID and a
  per-viewer feed-session seed.
- A 0.35 additive boost for user-submitted clips.

New clips use lifetime-average views/hour until a second observation exists.
Each first-page request creates a fresh seed mixed with the authenticated viewer
ID. The seed is carried in the pagination cursor, requiring no Redis or database
session state. User-submitted clips receive no random adjustment, while the same
seed keeps automated clip ordering stable throughout infinite scroll.

## Stress test

- A huge but stale clip keeps some view strength but loses velocity/freshness.
- A fresh small-channel clip can rise through velocity and exploration.
- A low-signal user post gets distribution help but does not permanently outrank
  every high-signal automated clip.
- Refreshing the first page changes the automated clip arrangement.
- Following the returned cursor preserves the arrangement without duplicate or
  missing clips.
- User-submitted scores are identical across seeds.

## Compatibility

The API exposes the effective session score as `trending_score` and carries the
shuffle seed in an extended backward-compatible cursor format. The frontend
infinite-query hook passes that cursor back verbatim.
