# Clipper Design System — "Swiss Dark, Comment-Forward"

> **Philosophy**: The clip is the prompt. The discussion is the product.
> Every layout decision prioritizes making comments visible, accessible, and frictionless.

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

### Comment-Forward

1. **Comments beside, not below** — On desktop, comments are always visible alongside the clip. Never require scrolling past the video to reach discussion.
2. **Comments inside playlists** — When watching a playlist, the current clip's comments are one tab away in the sidebar. No navigation required.
3. **Input always visible** — The comment form is sticky. Never buried at the bottom of a thread.
4. **Discussion as discovery** — Playlist cards preview the top comment. Comment count is a first-class metric alongside votes.

### Swiss Dark

5. **Readability over decoration** — Typography, contrast, and spacing are optimized for reading dense comment threads. No gradients, no glassmorphism, no aurora effects.
6. **Grid discipline** — Consistent spacing scale, clear visual hierarchy, predictable layout.
7. **Personality through restraint** — Brand violet as a single accent color. Let the conversations provide the energy.

### Density

8. **Two density modes** — Compact (sidebar, playlist) and Expanded (full page). Same data, different spacing.
9. **Scannable by default** — Inline metadata (author, time, score) on one line. Thread structure visible at a glance.

---

## 2. Color Tokens

### Core Palette

All colors defined as CSS custom properties using space-separated RGB for Tailwind alpha support.

| Token                    | RGB        | Hex       | Usage                                                     |
| ------------------------ | ---------- | --------- | --------------------------------------------------------- |
| `--clpr-background`     | `15 15 20` | `#0F0F14` | Page background — warm dark, less sterile than pure black |
| `--clpr-surface`        | `26 26 36` | `#1A1A24` | Cards, panels, comment containers                         |
| `--clpr-surface-raised` | `34 34 51` | `#222233` | Comment input, modals, elevated panels, dropdowns         |
| `--clpr-surface-hover`  | `42 42 60` | `#2A2A3C` | Hover state for interactive surfaces                      |
| `--clpr-border`         | `42 42 58` | `#2A2A3A` | Borders, dividers, thread lines                           |
| `--clpr-border-subtle`  | `34 34 48` | `#222230` | Subtle separators within cards                            |

### Text Hierarchy

| Token                    | RGB           | Hex       | Contrast on Surface | Usage                                                 |
| ------------------------ | ------------- | --------- | ------------------- | ----------------------------------------------------- |
| `--clpr-text-primary`   | `232 232 237` | `#E8E8ED` | ~12:1               | Comment body, headings, primary content               |
| `--clpr-text-secondary` | `152 152 168` | `#9898A8` | ~5.5:1              | Timestamps, usernames, metadata                       |
| `--clpr-text-tertiary`  | `144 144 160` | `#9090A0` | ~6.0:1              | Neutral vote counts, placeholders (not for body text) |
| `--clpr-text-disabled`  | `68 68 82`    | `#444452` | ~2:1                | Disabled states only                                  |

### Brand & Accent

| Token                 | Hex         | Usage                                                      |
| --------------------- | ----------- | ---------------------------------------------------------- |
| `--clpr-brand`       | `#7C3AED`   | Brand violet — links, active tab indicators, user mentions |
| `--clpr-brand-hover` | `#6D28D9`   | Hover state for brand elements                             |
| `--clpr-brand-muted` | `#7C3AED26` | 15% opacity — subtle brand tint backgrounds                |

### Interaction Colors

| Token                    | Hex         | Usage                                   |
| ------------------------ | ----------- | --------------------------------------- |
| `--clpr-upvote`         | `#F97316`   | Upvote active state — warm orange       |
| `--clpr-upvote-hover`   | `#F9731626` | Upvote hover background (15% opacity)   |
| `--clpr-downvote`       | `#6366F1`   | Downvote active state — indigo          |
| `--clpr-downvote-hover` | `#6366F126` | Downvote hover background (15% opacity) |
| `--clpr-cta`            | `#818CF8`   | Reply button, primary actions — indigo  |
| `--clpr-cta-hover`      | `#6366F1`   | CTA hover state                         |
| `--clpr-focus-ring`     | `#7C3AED`   | Focus outline for keyboard navigation   |

### Semantic Colors

| Token             | Hex       | Usage                                         |
| ----------------- | --------- | --------------------------------------------- |
| `--clpr-success` | `#22C55E` | Success states, positive feedback             |
| `--clpr-warning` | `#F59E0B` | Warnings, edit indicators                     |
| `--clpr-error`   | `#EF4444` | Errors, delete confirmations, removed content |
| `--clpr-info`    | `#6366F1` | Informational, tips (indigo — matches brand)  |

### Thread Colors

Nested comment threads use progressively subtle left-border colors:

| Depth    | Border Color       | Hex       |
| -------- | ------------------ | --------- |
| 0 (root) | `--clpr-brand`    | `#7C3AED` |
| 1        | `--clpr-thread-1` | `#A855F7` |
| 2        | `--clpr-thread-2` | `#C084FC` |
| 3        | `--clpr-thread-3` | `#E879A8` |
| 4        | `--clpr-thread-4` | `#F0ABAB` |
| 5+       | `--clpr-border`   | `#2A2A3A` |

Thread colors stay in the violet-pink warm family for cohesion with the brand palette. Colored thread lines help users visually trace reply chains in deep threads.

---

## 3. Typography

### Font Stack

**Headings & UI Labels**: Space Grotesk — geometric, techy, distinctive character
**Body & Comments**: Inter — designed for screens, exceptional readability at 14px
**Accent & Stats**: Syne — bold, artistic character for hero numbers, achievements, and empty states
**Code Blocks**: JetBrains Mono — clear distinction between similar characters

```css
@import url('https://fonts.googleapis.com/css2?family=Inter:wght@400;500;600&family=Space+Grotesk:wght@500;600;700&family=JetBrains+Mono:wght@400;500&family=Syne:wght@600;700;800&display=swap');
```

```css
/* frontend/src/index.css */
@theme inline {
  --font-sans: Inter, system-ui, sans-serif;
  --font-heading: "Space Grotesk", system-ui, sans-serif;
  --font-accent: Syne, "Space Grotesk", system-ui, sans-serif;
  --font-mono: "JetBrains Mono", ui-monospace, monospace;
}
```

**Accent font usage**: Leaderboard positions, achievement titles, empty state headlines, stat counters on profile pages. Use `font-accent` with weights 600-800 and tight letter spacing (-0.02em).

### Type Scale

| Element                | Font          | Size             | Weight | Line Height | Letter Spacing | Color Token       |
| ---------------------- | ------------- | ---------------- | ------ | ----------- | -------------- | ----------------- |
| Page title (h1)        | Space Grotesk | 28px / 1.75rem   | 700    | 1.2         | -0.02em        | `text-primary`    |
| Section heading (h2)   | Space Grotesk | 22px / 1.375rem  | 700    | 1.25        | -0.015em       | `text-primary`    |
| Card heading (h3)      | Space Grotesk | 18px / 1.125rem  | 600    | 1.3         | -0.01em        | `text-primary`    |
| Subsection (h4)        | Space Grotesk | 16px / 1rem      | 600    | 1.35        | -0.005em       | `text-primary`    |
| **Comment body**       | Inter         | 14px / 0.875rem  | 400    | 1.6         | 0              | `text-primary`    |
| **Comment author**     | Space Grotesk | 13px / 0.8125rem | 600    | 1.3         | 0              | `link` (linked)  |
| **Comment timestamp**  | Inter         | 12px / 0.75rem   | 400    | 1.4         | 0.01em         | `text-secondary`  |
| **Comment score**      | Inter         | 13px / 0.8125rem | 600    | 1           | 0              | context-dependent |
| **Reply action**       | Inter         | 12px / 0.75rem   | 500    | 1.4         | 0.01em         | `cta`             |
| **Thread "view more"** | Inter         | 12px / 0.75rem   | 500    | 1.4         | 0              | `brand`           |
| Body text              | Inter         | 15px / 0.9375rem | 400    | 1.6         | 0              | `text-primary`    |
| Small / caption        | Inter         | 12px / 0.75rem   | 400    | 1.4         | 0.01em         | `text-secondary`  |
| Button label           | Inter         | 14px / 0.875rem  | 500    | 1           | 0.01em         | context-dependent |
| Badge / tag            | Inter         | 11px / 0.6875rem | 600    | 1.2         | 0.03em         | context-dependent |

### Comment Body Markdown Rendering

```css
.comment-body {
    font-family: 'Inter', system-ui, sans-serif;
    font-size: 0.875rem; /* 14px */
    line-height: 1.6;
    color: var(--clpr-text-primary);
}

.comment-body p + p {
    margin-top: 0.5rem;
}
.comment-body blockquote {
    border-left: 2px solid var(--clpr-border);
    padding-left: 0.75rem;
    color: var(--clpr-text-secondary);
    font-style: italic;
}
.comment-body code {
    font-family: 'JetBrains Mono', monospace;
    font-size: 0.8125rem; /* 13px */
    background: var(--clpr-surface-raised);
    padding: 0.125rem 0.375rem;
    border-radius: 4px;
}
.comment-body pre code {
    display: block;
    padding: 0.75rem;
    overflow-x: auto;
}
.comment-body a {
    color: var(--clpr-brand);
    text-decoration: underline;
    text-underline-offset: 2px;
}
```

---

## 4. Spacing & Layout

### Spacing Scale

Based on a 4px base unit for tight control over comment density.

| Token       | Value | Common Use                               |
| ----------- | ----- | ---------------------------------------- |
| `space-0.5` | 2px   | Inline icon gap                          |
| `space-1`   | 4px   | Tight inner padding                      |
| `space-1.5` | 6px   | Compact comment gap                      |
| `space-2`   | 8px   | Default inner padding, compact mode gaps |
| `space-3`   | 12px  | Expanded comment gap                     |
| `space-4`   | 16px  | Card padding, section gaps               |
| `space-6`   | 24px  | Between sections                         |
| `space-8`   | 32px  | Major section breaks                     |
| `space-12`  | 48px  | Page-level spacing                       |

### Border Radius

| Token        | Value | Usage                             |
| ------------ | ----- | --------------------------------- |
| `rounded-sm` | 4px   | Badges, inline code               |
| `rounded`    | 6px   | Buttons, inputs, small cards      |
| `rounded-md` | 8px   | Cards, panels, comment containers |
| `rounded-lg` | 12px  | Modals, large panels              |

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
| Body           | `comment-body` class (see Typography section), 14px Inter          |
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

### 6.6 ClipCard (with comment preview)

Standard card used in feeds and search results:

```
┌────────────────────────────────────┐
│ [Thumbnail ─────────── 16:9 ────] │
│  0:32                         ▶   │  ← duration overlay, play icon
├────────────────────────────────────┤
│ Clip Title That Might Wrap to      │
│ Two Lines Maximum                  │
│                                    │
│ @broadcaster · Game Name           │  ← text-secondary
│ ▲ 142  ·  💬 47                    │  ← votes + comment count
│                                    │
│ "This play was absolutely..."      │  ← text-secondary, italic, line-clamp-1
│  — @topcommenter                   │  ← text-tertiary
└────────────────────────────────────┘
```

| Element             | Spec                                                                      |
| ------------------- | ------------------------------------------------------------------------- |
| Card background     | `color-surface`                                                           |
| Card border         | `1px solid var(--clpr-border-subtle)`                                    |
| Card radius         | `rounded-md` (8px)                                                        |
| Card hover          | `background: var(--clpr-surface-hover)`, transition 150ms                |
| Thumbnail           | `aspect-ratio: 16/9`, `object-fit: cover`, `rounded-md` top corners       |
| Title               | `font-heading`, 15px, weight 600, `line-clamp-2`                          |
| Metadata            | 12px, `text-secondary`                                                    |
| Vote count          | 13px, weight 600, `text-primary`                                          |
| Comment count       | 13px, weight 500, `text-secondary`, with 💬 icon (Lucide `MessageSquare`) |
| Top comment preview | 12px, italic, `text-secondary`, `line-clamp-1`                            |
| Top comment author  | 11px, `text-tertiary`                                                     |
| Padding             | 12px (thumbnail area: 0)                                                  |

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

Use `text-link` for small accent text and links. Brand violet belongs on filled controls and decorative elements. Links within paragraphs need a persistent underline. Provider-supplied tag colors use linearized sRGB luminance to choose black or white text; unsupported color strings fall back to the standard violet.

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
