# clpr Design System — "Tally · Ultraviolet"

> **Idea**: clpr is a broadcast control room for Twitch moments. Clips are
> monitors, rank is read across the room, and every tag says where it came from.
> Discussion stays beside the clip.

---

## Table of Contents

1. [Design Principles](#1-design-principles)
2. [Color Tokens](#2-color-tokens)
3. [Typography](#3-typography)
4. [Spacing & Layout](#4-spacing--layout)
5. [Layout Architecture](#5-layout-architecture)
6. [Component Specifications](#6-component-specifications)
7. [Animation & Motion](#7-animation--motion)
8. [Accessibility](#8-accessibility)
9. [Active theme configuration](#9-active-theme-configuration)
10. [Maintenance and verification](#10-maintenance-and-verification)

---

## 1. Design Principles

### Control room

1. **The tally light means "this one"**: violet marks the item that is on air: the active nav item, the selected tab, the top rank, and the primary action. Use it once per region.
2. **Rank is big, metadata is mono**: rank numerals and titles use condensed uppercase display type. Counts, durations, timecodes and dates use the mono face with tabular figures.
3. **Rules, not cards**: surfaces are separated by 1px rules on a violet-tinted ink. Corners are square or nearly square. No gradients, glass or soft drop shadows.
4. **Clpr's own purple**: the violet sits near Twitch's purple to suggest the relationship without copying it. Never use Twitch's exact brand purple or their logo.
5. **No gaming visual language**: clpr covers every kind of Twitch moment. Avoid gamer tropes such as HUD meters, neon RGB or esports scoreboards.

### Tags say where they came from

6. **One colour per tag origin, not per tag**: Twitch category, what clpr saw (the vision tagger), community tags, and the streamer's own Twitch channel tags each have one treatment. Duration and language are metadata, not chips.
7. **Evidence is visible**: tags from the vision tagger show the evidence level the backend assigned (`visible`, `contextual`, or `strong`; see `backend/internal/services/content_tags.go`).

### Comment-forward

8. **Comments beside, not below**: on wide desktop, discussion is visible next to the clip. In playlists, the current clip's comments are one tab away.
9. **Input always visible**: the comment form is sticky, never buried at the bottom of a thread.
10. **Two density modes**: compact (sidebar, playlist) and expanded (full page). Same data, different spacing.

---

## 2. Color Tokens

All colours are CSS custom properties holding space-separated RGB so Tailwind alpha modifiers work (`bg-surface/80`). Values live in `frontend/src/index.css`.

### Surfaces and rules

| Token                    | RGB          | Hex       | Usage                                  |
| ------------------------ | ------------ | --------- | -------------------------------------- |
| `--clpr-background`      | `14 12 19`   | `#0E0C13` | Page background (violet-tinted ink)    |
| `--clpr-surface`         | `23 20 31`   | `#17141F` | Panels, clip cards, inputs             |
| `--clpr-surface-raised`  | `30 26 40`   | `#1E1A28` | Menus, modals, toasts                  |
| `--clpr-surface-hover`   | `38 33 51`   | `#262133` | Hover state for interactive surfaces   |
| `--clpr-border`          | `39 34 47`   | `#27222F` | Default 1px rule                       |
| `--clpr-border-subtle`   | `29 25 37`   | `#1D1925` | Separators inside a panel              |
| `--clpr-border-strong`   | `59 53 72`   | `#3B3548` | Outlined buttons, chips, inputs        |

### Text

| Token                    | Hex       | Usage                                           |
| ------------------------ | --------- | ----------------------------------------------- |
| `--clpr-text-primary`    | `#EEEDF7` | Body, titles (cool "ice" white)                 |
| `--clpr-text-secondary`  | `#ACA7BB` | Metadata, secondary labels                      |
| `--clpr-text-tertiary`   | `#8F8A9C` | Mono labels (`.kicker`), counts                 |
| `--clpr-text-disabled`   | `#534D60` | Disabled states and decoration only             |

### Signal colours

| Token                  | Hex       | Meaning                                                              |
| ---------------------- | --------- | -------------------------------------------------------------------- |
| `--clpr-brand` (tally) | `#8C5CFF` | On air: active item, top rank, primary action, upvote                |
| `--clpr-link`          | `#B79BFF` | Links and focus ring                                                 |
| `--clpr-seen`          | `#3DDC97` | Tag evidence `visible` (seen in the frame); success                  |
| `--clpr-context`       | `#FFC24A` | Tag evidence `contextual` (metadata or transcript plus frame); warning |
| `--clpr-category`      | `#9FD8FF` | Twitch category lane; info; downvote                                 |

Primary buttons use `primary-400` (`#A07CFF`) with ink text. Ink on `#8C5CFF` measures about 4.7:1; white on it measures about 4.1:1 and fails AA, so text on solid violet is always ink (`text-background`). Hover lightens (`primary-300`) instead of darkening.

The Tailwind `gray`, `neutral`, `zinc`, `slate` and `stone` scales are remapped to the violet-ink ramp, so older ad-hoc classes stay on palette until each surface is rebuilt. `success`, `warning`, `error` and `info` scales are retuned around the signal colours above.

The other Tailwind colour families point at those scales by meaning, so legacy classes cannot introduce an off-palette hue:

| Legacy families                                   | Resolves to            |
| ------------------------------------------------- | ---------------------- |
| `blue`, `purple`, `violet`, `indigo`, `fuchsia`, `pink` | `primary` (violet)     |
| `red`, `rose`                                     | `error`                |
| `green`, `emerald`, `lime`                        | `success` (seen)       |
| `yellow`, `amber`, `orange`                       | `warning` (context)    |
| `sky`, `cyan`, `teal`                             | `info` (category)      |

New code should use the semantic names (`primary-*`, `tally`, `seen`, `context`, `category`, `error-*`) rather than the legacy families. There are no gradients: media captions use solid `bg-black/70`–`/80` bars.

### Thread colours

| Depth | Token             | Hex       |
| ----- | ----------------- | --------- |
| 0     | `--clpr-thread-0` | `#8C5CFF` |
| 1     | `--clpr-thread-1` | `#9FD8FF` |
| 2     | `--clpr-thread-2` | `#3DDC97` |
| 3     | `--clpr-thread-3` | `#FFC24A` |
| 4     | `--clpr-thread-4` | `#8F8A9C` |

---

## 3. Typography

### Families

| Role    | Family            | Tailwind        | Use                                                       |
| ------- | ----------------- | --------------- | --------------------------------------------------------- |
| Display | Barlow Condensed  | `font-heading`  | Headings (uppercase), rank numerals, buttons, nav, tabs   |
| Text    | Barlow            | `font-sans`     | Body, comments, descriptions                              |
| Data    | IBM Plex Mono     | `font-mono`     | Counts, durations, timecodes, dates, tags, `.kicker` labels |

```css
@import url('https://fonts.googleapis.com/css2?family=Barlow+Condensed:wght@600;700;800&family=Barlow:ital,wght@0,400;0,500;0,600;0,700;1,400&family=IBM+Plex+Mono:wght@400;500;600&display=swap');
```

`h1`–`h6` are uppercase Barlow Condensed at weight 700 by default. Uppercase is visual only; keep source text in sentence case so screen readers and tests see normal text.

### Helper classes

| Class        | Purpose                                                              |
| ------------ | -------------------------------------------------------------------- |
| `.kicker`    | 11px mono uppercase label with 0.12em tracking (section labels, form labels) |
| `.display`   | Condensed 800 uppercase for titles and rank numerals                 |
| `.burn-in`   | Timecode burned into a media corner (mono on 72% black)              |
| `.tally-bar` | 3px violet rule across the top of an active panel, menu or modal     |

### Type scale

| Element             | Family           | Size    | Weight | Notes                          |
| ------------------- | ---------------- | ------- | ------ | ------------------------------ |
| Page title (h1)     | Barlow Condensed | 24–36px | 700    | Uppercase, line-height 1.02    |
| Section (h2)        | Barlow Condensed | 20–30px | 700    | Uppercase                      |
| Rank numeral        | Barlow Condensed | 28–88px | 800    | Violet for #1, dim otherwise   |
| Button / nav / tab  | Barlow Condensed | 13–17px | 700    | Uppercase, 0.04–0.06em tracking |
| Body                | Barlow           | 15px    | 400    | Line-height 1.6                |
| Comment body        | Barlow           | 14px    | 400    | `.comment-body`                |
| Metadata / counts   | IBM Plex Mono    | 11–12px | 400–500 | Tabular figures               |
| Tag chip            | IBM Plex Mono    | 12px    | 500    | See tag lanes in §6            |

---

## 4. Spacing & Layout

### Spacing scale

A 4px base unit, unchanged: `0.5` 2px, `1` 4px, `1.5` 6px, `2` 8px, `3` 12px, `4` 16px, `6` 24px, `8` 32px, `12` 48px.

### Corners

| Token          | Value | Usage                                  |
| -------------- | ----- | -------------------------------------- |
| `rounded-none` | 0     | Buttons, inputs, cards, modals, chips  |
| `rounded-sm`   | 1px   | Legacy classes resolve here            |
| `rounded`–`lg` | 2px   | Legacy classes resolve here            |
| `rounded-xl`+  | 3–4px | Legacy classes resolve here            |
| `rounded-full` | pill  | Avatars and status dots only           |

### Breakpoints

Unchanged from current config:

| Name  | Min Width | Layout Behavior                                |
| ----- | --------- | ---------------------------------------------- |
| `xs`  | 375px     | Mobile — single column, stacked layout         |
| `sm`  | 640px     | Wide mobile — minor padding adjustments        |
| `md`  | 768px     | Tablet — playback and discussion remain stacked |
| `lg`  | 1024px    | Desktop — feed sidebar; stacked clip discussion         |
| `xl`  | 1280px    | Wide desktop — playback beside discussion             |
| `2xl` | 1536px    | Ultra-wide — max-width container               |

### Container

```css
.container {
    width: 100%;
    max-width: 1440px; /* shared with Container and page-container */
    margin: 0 auto;
    padding-inline: 1rem; /* mobile */
}
@media (min-width: 640px) {
    .container {
        padding-inline: 1.5rem;
    }
}
@media (min-width: 1024px) {
    .container {
        padding-inline: 2rem;
    }
}
```

---

## 5. Layout Architecture

### 5.1 ClipDetailPage — Side-by-Side

The most important layout change. Comments move from below the video to beside it.

**Desktop (lg+)**:

```
┌──────────────────────────────────────────────────────────────┐
│  Header / Nav                                                │
├────────────────────────────────┬─────────────────────────────┤
│  Video Player                  │  Comment Panel              │
│  ┌──────────────────────────┐  │  ┌─────────────────────────┐│
│  │                          │  │  │ 47 comments  Sort ▾     ││
│  │       16:9 Player        │  │  ├─────────────────────────┤│
│  │                          │  │  │ ▲ 42 ▼ @user · 2h ago  ││
│  └──────────────────────────┘  │  │ This play was insane,   ││
│                                │  │ the way they...          ││
│  Clip Title                    │  │   └─ ▲ 12 ▼ @reply · 1h││
│  @broadcaster · Game · 2h ago  │  │     Agreed, timing...   ││
│                                │  │   └─ ▲ 5 ▼ @reply2     ││
│  ┌────┬────────┬──────────┐   │  │     + 3 more replies    ││
│  │ ▲▼ │ 💬 47  │ ♥ Save  │   │  ├─────────────────────────┤│
│  └────┴────────┴──────────┘   │  │ ▲ 38 ▼ @user2 · 1h ago ││
│                                │  │ Context: this was       ││
│  Share · Report · ...          │  │ during the finals...    ││
│                                │  ├─────────────────────────┤│
│                                │  │ [sticky] Write a        ││
│                                │  │ comment...        [Post]││
│                                │  └─────────────────────────┘│
└────────────────────────────────┴─────────────────────────────┘
```

```css
/* ClipDetailPage grid */
.clip-detail-layout {
    display: grid;
    gap: 0;
}

/* Mobile: stacked */
.clip-detail-layout {
    grid-template-columns: 1fr;
    grid-template-rows: auto 1fr;
}

/* Desktop: side-by-side */
@media (min-width: 1024px) {
    .clip-detail-layout {
        grid-template-columns: 1fr 420px;
        grid-template-rows: 1fr;
        max-height: calc(100vh - var(--nav-height));
    }
}

@media (min-width: 1280px) {
    .clip-detail-layout {
        grid-template-columns: 1fr 480px;
    }
}
```

**Left panel (clip)**: `position: sticky; top: var(--nav-height)` — stays visible while scrolling comments on mobile fallback.

**Right panel (comments)**: `overflow-y: auto; height: calc(100vh - var(--nav-height))` — independently scrollable comment stream.

**Mobile (< lg)**: Single column, video on top. Comment form becomes a **sticky bottom bar**:

```
┌──────────────────────────┐
│  Video Player (16:9)     │
│  Title + metadata        │
│  Vote / Fav / Share      │
├──────────────────────────┤
│  Comments (scrollable)   │
│  ...                     │
│  ...                     │
├──────────────────────────┤
│  [sticky] Add comment... │  ← fixed to bottom of viewport
└──────────────────────────┘
```

### 5.2 PlaylistTheatreMode — Comment Tab in Sidebar

Add a tab system to the existing sidebar so users can discuss clips without leaving the playlist.

```
┌─────────────────────────────────────────────────────────────┐
│  ┌───────────────────────────────────┐  ┌─────────────────┐ │
│  │                                   │  │ Queue │ Chat 💬47│ │
│  │                                   │  ├─────────────────┤ │
│  │         Video Player              │  │                 │ │
│  │                                   │  │  (Tab Content)  │ │
│  │                                   │  │                 │ │
│  │                                   │  │                 │ │
│  │                                   │  ├─────────────────┤ │
│  └───────────────────────────────────┘  │ Comment input   │ │
│                                         └─────────────────┘ │
└─────────────────────────────────────────────────────────────┘
```

**Tab bar**:

- `Queue` — current playlist items (existing behavior)
- `Chat 💬 47` — comments for the currently playing clip, with count badge

**Sidebar comments use compact density** (see section 6.2). Max thread depth: 2 levels in sidebar, with "View full thread" links.

**Keyboard shortcuts**:

- `C` — switch to Chat tab
- `Q` — switch to Queue tab
- `N` — next clip (existing)
- `S` — toggle sidebar (existing)

### 5.3 PlaylistCard — Comment Preview

Add comment count and top-comment preview to playlist cards in list/grid views.

```
┌───────────────────────────────────────┐
│  [Thumbnail ─────────── 16:9 ──────] │
│                                       │
│  Clip Title That Might Be Long...     │
│  @broadcaster · Game Name             │
│  ▲ 142 · 💬 47                        │
│                                       │
│  "This play was absolutely insane..." │  ← top comment (1 line, line-clamp-1)
│   — @topcommenter                     │
└───────────────────────────────────────┘
```

The top-comment preview acts as a hook — it signals active discussion and gives visitors a reason to click beyond the thumbnail.

---

## 6. Component Specifications

### 6.1 CommentItem — Expanded (ClipDetailPage)

Used on the full ClipDetailPage comment panel.

```
┌──────────────────────────────────────────────────┐
│ ▲                                                │
│ 42   @username · 2h ago · edited                 │
│ ▼                                                │
│                                                  │
│      Comment body text goes here. This can be    │
│      multiple lines with full markdown support   │
│      including **bold**, *italic*, `code`,       │
│      > blockquotes, and [links](#).              │
│                                                  │
│      Reply · Share · Report                      │
│                                                  │
│  ┃   ▲                                           │  ← thread line (colored by depth)
│  ┃   8    @replier · 1h ago                      │
│  ┃   ▼                                           │
│  ┃        Reply body text...                     │
│  ┃        Reply · Share                          │
│  ┃                                               │
│  ┃   + 3 more replies                            │
└──────────────────────────────────────────────────┘
```

**Structure**:

| Element        | Spec                                                               |
| -------------- | ------------------------------------------------------------------ |
| Vote column    | 32px wide, vertical stack: ▲ button + score + ▼ button             |
| Author         | `font-heading`, 13px, weight 600, `color-brand` (links to profile) |
| Separator dot  | `·` in `text-tertiary`                                             |
| Timestamp      | 12px, `text-secondary`, relative time ("2h ago")                   |
| Edit indicator | "edited" in `text-tertiary`, shows tooltip with edit timestamp     |
| Body           | `comment-body` class (see Typography section), 14px Barlow          |
| Actions        | 12px, `text-secondary`, hover `text-cta` for Reply                 |
| Thread indent  | `margin-left: 12px` per depth level                                |
| Thread line    | `border-left: 2px solid`, color from thread color table            |
| Thread padding | `padding-left: 12px` inside the thread line                        |

**Spacing**:

- Between comments: `padding-y: 12px`, separated by `border-bottom: 1px solid var(--clpr-border-subtle)`
- Vote button touch targets: 32x32px minimum
- Gap between vote column and content: 8px

### 6.2 CommentItem — Compact (Playlist Sidebar)

Used in the PlaylistTheatreMode sidebar Chat tab.

```
┌───────────────────────────────────┐
│ ▲42▼  @user · 2h                  │
│       Comment text truncated to   │
│       two lines max with clamp... │
│       Reply                       │
│                                   │
│  │ ▲8▼  @reply · 1h              │  ← thread line, single indent
│  │      Reply text one line...    │
│  │      + 2 more                  │
└───────────────────────────────────┘
```

**Differences from Expanded**:

| Property         | Expanded                   | Compact                         |
| ---------------- | -------------------------- | ------------------------------- |
| Vote layout      | Vertical stack (▲ score ▼) | Inline horizontal (▲42▼)        |
| Vote font        | 13px weight 600            | 12px weight 500                 |
| Body line clamp  | None (full)                | `line-clamp-2` with "show more" |
| Thread depth max | 10                         | 2 (then "View thread" link)     |
| Padding-y        | 12px                       | 8px                             |
| Thread indent    | 12px/level                 | 8px/level                       |
| Actions shown    | Reply, Share, Report       | Reply only                      |
| Author font size | 13px                       | 12px                            |

### 6.3 CommentForm

**Desktop (in side panel, sticky bottom)**:

```
┌──────────────────────────────────────────┐
│  ┌────────────────────────────────────┐  │
│  │  Write a comment...                │  │  ← single line, expands on focus
│  └────────────────────────────────────┘  │
│  B  I  ~~  🔗  "  `  😊  ?    [Post]   │  ← toolbar appears on focus
└──────────────────────────────────────────┘
```

**Behavior**:

- Default: Single-line input, placeholder "Write a comment..."
- On focus: Expands to 3-line minimum textarea, toolbar appears with fade-in
- On typing: Grows to max 8 lines, then scrolls internally
- Post button: `color-cta` blue, disabled until content exists
- `Ctrl+Enter` to submit, `Escape` to collapse (if empty)
- Markdown preview: Toggle via toolbar icon, not a separate tab (saves vertical space)

**Mobile (sticky bottom bar)**:

```
┌────────────────────────────────────────┐
│  [avatar] Add a comment...      [▶]   │  ← tappable bar
└────────────────────────────────────────┘
```

On tap: Slides up a bottom sheet with full editor, keyboard, and toolbar.

**Compact (playlist sidebar)**:

```
┌────────────────────────────────┐
│  Comment on this clip... [▶]   │  ← minimal, single line
└────────────────────────────────┘
```

### 6.4 CommentVoteButtons

**Expanded (vertical)**:

```
  [▲]      ← 32x32 touch target, transparent bg
   42      ← score, 13px weight 600
  [▼]      ← 32x32 touch target
```

| State            | Icon Color                | Score Color      | Background                    |
| ---------------- | ------------------------- | ---------------- | ----------------------------- |
| Neutral          | `text-tertiary`           | `text-tertiary`  | none                          |
| Hover (up)       | `text-primary`            | —                | `upvote-hover` (15% orange)   |
| Hover (down)     | `text-primary`            | —                | `downvote-hover` (15% indigo) |
| Active upvoted   | `color-upvote` (filled)   | `color-upvote`   | none                          |
| Active downvoted | `color-downvote` (filled) | `color-downvote` | none                          |

**Compact (inline)**:

```
  ▲ 42 ▼   ← all on one line, 12px
```

Same color states, smaller touch targets (28x28), no background on hover.

### 6.5 CommentSection Header

```
┌──────────────────────────────────────────┐
│  47 comments              Sort: Best ▾   │
└──────────────────────────────────────────┘
```

- Comment count: `font-heading`, 14px, weight 600
- Sort dropdown: `text-secondary`, 12px, with current sort highlighted in `text-primary`
- Sort options: Best (default), New, Top, Old, Controversial
- Divider below: `border-bottom: 1px solid var(--clpr-border)`

### 6.6 ClipCard and tag chips

Feed card (`frontend/src/components/clip/ClipCard.tsx`):

```
┌─[tally rule on #1 or the playing card]───────────────────────┐
│ 01   STRANGER JOINS THE STREET PERFORMANCE          ← display │
│  ▲   LUNAWAVES · MUSIC · CLIPPED BY OKAYSAM · 2H    ← mono    │
│ 42   ┌──────────────── 16:9 ───────────────┐                  │
│  ▼   │               [▶]                    │                  │
│      │                               0:20  │ ← burn-in       │
│      └──────────────────────────────────────┘                  │
│      [Singing CTX] [Music] [#goosebumps] +2  ← tag lanes      │
│      💬 418 COMMENTS  ♥ 12  SHARE     👁 2.1K  PLAYLIST QUEUE  │
└───────────────────────────────────────────────────────────────┘
```

| Element        | Spec                                                                               |
| -------------- | ---------------------------------------------------------------------------------- |
| Card           | `bg-card`, 1px `border-border`, square; hover raises the rule to `line-strong`     |
| Tally rule     | `.tally-bar` on rank 1 and on the active (playing) card                            |
| Rank           | Ranked sorts only (not "new"). `.display` 48px; #1 in tally violet, others tertiary. Mobile shows `#01` in the metadata line |
| Title          | Uppercase Barlow Condensed, 24–28px, `line-clamp-2`                                |
| Metadata       | IBM Plex Mono 11px uppercase, `·` separators                                       |
| Duration       | `.burn-in` in the bottom-right media corner                                        |
| Play control   | Square `primary-400` tile with ink icon; no blur or glow                           |
| Tags           | `TagList`, ordered by lane; length and language are not chips                      |

Tag chips (`frontend/src/components/tag/TagChip.tsx`) are 12px mono with a 1px border. Colour depends on the lane, never on per-tag colour:

| Lane       | Treatment                                                             |
| ---------- | --------------------------------------------------------------------- |
| Detected   | Ice text, `line-strong` border, evidence mark: Seen (mint), Ctx (amber), Outcome (violet) |
| Category   | Sky text and border                                                   |
| Community  | `#` prefix, link violet                                               |
| Streamer   | Dashed border, tertiary text                                          |

The evidence mark describes what the tag definition requires, not a verdict about the clip.

### 6.7 Playlist Sidebar Tabs

```
┌────────────────────────────────┐
│  [Queue]    [Chat 💬 47]       │  ← tab bar
├────────────────────────────────┤
│                                │
│  (tab content)                 │
│                                │
└────────────────────────────────┘
```

| Element               | Spec                                                               |
| --------------------- | ------------------------------------------------------------------ |
| Tab bar height        | 40px                                                               |
| Tab bar background    | `color-surface`                                                    |
| Tab font              | `font-heading`, 13px, weight 600                                   |
| Inactive tab          | `text-secondary`                                                   |
| Active tab            | `text-primary`, bottom border 2px `color-brand`                    |
| Badge (count)         | 11px, weight 600, `color-brand` background, white text, pill shape |
| Tab bar border-bottom | `1px solid var(--clpr-border)`                                    |

---

## 7. Animation & Motion

### Principles

1. **Purposeful only** — animate to communicate state change, not for decoration
2. **Fast** — 150ms for micro-interactions, 200ms for reveals, 300ms max for page transitions
3. **Respect preferences** — all animations wrapped in `prefers-reduced-motion` check

### Timing Functions

| Use Case                      | Easing      | CSS                            |
| ----------------------------- | ----------- | ------------------------------ |
| Element entering              | ease-out    | `cubic-bezier(0, 0, 0.2, 1)`   |
| Element exiting               | ease-in     | `cubic-bezier(0.4, 0, 1, 1)`   |
| State change (color, opacity) | ease-in-out | `cubic-bezier(0.4, 0, 0.2, 1)` |

### Specific Animations

| Element             | Animation                  | Duration          |
| ------------------- | -------------------------- | ----------------- |
| Vote state change   | Color transition           | 150ms ease-in-out |
| Comment form expand | Height + opacity           | 200ms ease-out    |
| New comment appear  | Fade-in + slide-down (4px) | 200ms ease-out    |
| Comment collapse    | Height to 0 + opacity      | 150ms ease-in     |
| Tab switch content  | Fade cross-dissolve        | 150ms ease-in-out |
| Card hover          | Background color           | 150ms ease-in-out |
| Mobile bottom sheet | Slide-up                   | 300ms ease-out    |
| Skeleton shimmer    | translateX(-100% to 100%)  | 2s infinite       |

### Reduced Motion

```css
@media (prefers-reduced-motion: reduce) {
    *,
    *::before,
    *::after {
        animation-duration: 0.01ms !important;
        animation-iteration-count: 1 !important;
        transition-duration: 0.01ms !important;
    }
}
```

---

## 8. Accessibility

### Contrast Ratios

All text tokens verified against their intended background:

| Combination                    | Ratio  | WCAG Level                                 |
| ------------------------------ | ------ | ------------------------------------------ |
| `text-primary` on `surface`    | ~12:1  | AAA                                        |
| `text-secondary` on `surface`  | ~5.5:1 | AA                                         |
| `text-tertiary` on `surface`   | ~6.0:1 | AA Large only (use for non-essential info) |
| `brand` on `surface`           | ~4.8:1 | AA                                         |
| `upvote` on `surface`          | ~5.2:1 | AA                                         |
| `cta` on `surface`             | ~4.6:1 | AA                                         |
| `text-primary` on `background` | ~13:1  | AAA                                        |

### Focus Management

- All interactive elements show `outline: 2px solid var(--clpr-focus-ring)` with `outline-offset: 2px` on `:focus-visible`
- Comment threads are navigable with `Tab` / `Shift+Tab`
- Vote buttons announce score change to screen readers via `aria-live="polite"`
- Comment form uses `aria-label="Write a comment"` when placeholder-only

### Touch Targets

- Minimum 44x44px for primary actions (vote buttons in expanded mode)
- Minimum 32x32px for compact mode vote buttons (acceptable per WCAG for supplementary controls)
- 8px minimum gap between adjacent touch targets

### Keyboard Shortcuts

| Key          | Context              | Action                   |
| ------------ | -------------------- | ------------------------ |
| `C`          | PlaylistTheatreMode  | Switch to Chat tab       |
| `Q`          | PlaylistTheatreMode  | Switch to Queue tab      |
| `N`          | PlaylistTheatreMode  | Next clip                |
| `S`          | PlaylistTheatreMode  | Toggle sidebar           |
| `Ctrl+Enter` | Comment form focused | Submit comment           |
| `Escape`     | Comment form focused | Collapse form (if empty) |

### Screen Reader

- Comment count announced: "47 comments" (not just "47")
- Vote score announced: "42 points" with vote state "upvoted" / "downvoted"
- Thread depth announced: "Reply, depth 2 of 10"
- Collapsed thread: "Collapsed thread, 3 replies, press Enter to expand"

---

## 9. Active theme configuration

`frontend/src/index.css` is the single active Tailwind 4 theme entry. There is no legacy JavaScript Tailwind configuration file. It declares:

- `@import 'tailwindcss' source('./')` with test sources excluded from utility discovery.
- `@custom-variant dark` bound to the application's `.dark` root class.
- A rem-based `xs` breakpoint (23.4375rem), ordered with Tailwind's standard breakpoints.
- Native font, semantic color, spacing-independent layering, and animation tokens.
- Raw RGB channels named `--clpr-*`; generated Tailwind colors retain their `--color-*` namespace. Do not use a complete CSS color where channel values are expected.

Use `Container` for page wrappers and `page-container` in the shared navigation. Both cap content at 1440px and use 16px, 24px, and 32px responsive gutters. Avoid Tailwind's built-in `container` utility where the shared page width is intended.

Use `text-link` for small accent text and links. Solid violet belongs on filled controls and the tally light, and text on it is always ink (`text-background`). Links within paragraphs need a persistent underline. Provider-supplied tag colors use linearized sRGB luminance to choose black or white text; unsupported color strings fall back to the standard violet.

Clip detail switches to a two-column grid at `xl` (1280px): flexible playback and a 24rem discussion panel. At smaller widths the discussion follows playback. On wide screens the discussion list scrolls independently while its composer remains visible. Full comment text and author/moderation actions remain available in that panel.

The feed uses measured two-way virtualization, stable clip IDs, and an overscan region. Keep the focused row mounted. A failed later page must retain loaded clips and offer a retry; returning upward must restore earlier rows.

## 10. Maintenance and verification

Treat the implementation and executed browser evidence as authoritative. Design examples elsewhere in this document describe visual intent; they do not establish that a feature is enabled or released.

- Verify 390px, 768px, and 1440px layouts with populated public, member, moderator, and administrator fixtures. Check narrow reflow and magnification separately.
- Exercise keyboard navigation, Escape and focus restoration, disclosure semantics, and cookie-banner focus visibility. Automated accessibility checks complement these interactions.
- Keep touch controls comfortably sized: primary actions, voting, playlist actions, and search controls use 44px targets. Inline text links and dense tags still need adequate separation.
- Preserve entered settings and comment drafts across background refreshes and recoverable failures. Distinguish unavailable data from a valid empty state.
- Run production typechecking, lint, source reachability, ownership, component coverage, browser contracts, route checks, and the bundle gate.
- Application CSS is bounded at 128 KiB raw and 21 KiB gzip. The raw ceiling accounts for restored native theme utilities; compressed size is independently enforced. Initial application JavaScript remains bounded at 550 KiB.
- Keep screenshots and test reports outside tracked source. Record the exact source revision and distinguish source-browser checks from immutable-image qualification.
