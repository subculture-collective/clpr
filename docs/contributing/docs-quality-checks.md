---
title: Documentation Quality Checks
summary: Comprehensive guide to documentation quality enforcement in CI
tags: [documentation, ci, quality, testing]
area: contributing
status: active
version: 1.0
last_reviewed: 2026-01-29
---

# Documentation Quality Checks

This document describes the automated documentation quality checks that run in CI to ensure high-quality, consistent documentation throughout the Clipper project.

**Related Issues:** [#803](https://git.subcult.tv/subculture-collective/clpr/issues/803), [#845](https://git.subcult.tv/subculture-collective/clpr/issues/845), [#846](https://git.subcult.tv/subculture-collective/clpr/issues/846), [#805](https://git.subcult.tv/subculture-collective/clpr/issues/805)

## Overview

The documentation quality enforcement system includes six automated checks that run on every pull request:

1. **Markdown Linting** - Ensures consistent markdown formatting
2. **Spell Checking** - Catches typos while respecting Obsidian patterns
3. **Link Validation** - Verifies all links are valid and reachable
4. **Anchor Validation** - Ensures heading anchors exist
5. **Orphan Detection** - Finds unreachable documentation pages
6. **Asset Checking** - Detects unused assets

All checks **exclude the `/vault/**` directory**, which is used for HashiCorp Vault configuration, not documentation.

## Running Checks Locally

### All Checks at Once

```bash
npm run docs:check
```

This runs all six checks sequentially. Use this before submitting a PR to ensure it will pass CI.

### Individual Checks

Run specific checks for faster iteration:

```bash
# Markdown linting
npm run docs:lint

# Spell checking
npm run docs:spell

# Link validation
npm run docs:links

# Anchor validation
npm run docs:anchors

# Orphan detection
npm run docs:orphans

# Asset checking
npm run docs:assets
```

## Check Details

### 1. Markdown Linting

**Tool:** [markdownlint-cli2](https://github.com/DavidAnson/markdownlint-cli2)  
**Config:** `.markdownlint.jsonc`  
**Command:** `npm run docs:lint`

Validates markdown formatting including:
- Heading styles and hierarchy
- List formatting
- Code block formatting

**Obsidian-Friendly Settings:**
- Line length checking is disabled (MD013 off) to allow flexible formatting for docs and Obsidian exports
- Allows HTML (for complex tables and Obsidian features)
- Allows frontmatter (YAML at top of files)
- Flexible heading spacing
- Mixed emphasis styles (bold/italic)
- Flexible table pipe styles

**Exclusions:**
- `/vault/**` - HashiCorp Vault configuration
- `/docs/.obsidian/**` - Obsidian configuration files
- `/node_modules/**` - Dependencies

**Common Fixes:**
```bash
# View specific errors
npm run docs:lint

# Most errors can be fixed by:
# 1. Breaking long lines
# 2. Adding blank lines around headings
# 3. Fixing list indentation
```

### 2. Spell Checking

**Tool:** [cspell](https://cspell.org/)  
**Config:** `.cspell.json`  
**Command:** `npm run docs:spell`

Checks spelling across all documentation while respecting Obsidian syntax.

**Ignored Patterns:**
- Wikilinks: `[[page]]`, `[[page|alias]]`
- Block references: `^block-id`
- Tags: `#tag`, `#tag/subtag`

**Exclusions:**
- `/vault/**` - HashiCorp Vault configuration
- `/docs/.obsidian/**` - Obsidian configuration
- Image files (`.png`, `.svg`, etc.)

**Adding Words to Dictionary:**

Edit `.cspell.json` and add words to the `words` array:

```json
{
  "words": [
    "myword",
    "techterm",
    "productname"
  ]
}
```

**Common Fixes:**
```bash
# Find spelling errors
npm run docs:spell

# Fix by either:
# 1. Correcting the typo
# 2. Adding legitimate words to .cspell.json
```

### 3. Link Validation

**Tool:** [lychee](https://github.com/lycheeverse/lychee)  
**Config:** `.lycheeignore`  
**Command:** `npm run docs:links`

Validates all HTTP(S) and file links in documentation.

**Ignored Patterns:**
- `http://localhost` - Local development URLs
- `https://localhost` - Local development URLs
- `http://127.0.0.1` - Local development URLs
- `clips.twitch.tv` - May be rate-limited

**Exclusions:**
- `/vault/**` - HashiCorp Vault configuration

**Accepted Status Codes:**
- `200` - OK
- `206` - Partial Content
- `429` - Too Many Requests (retry after delay)

**Common Fixes:**
```bash
# Find broken links
npm run docs:links

# Fix by:
# 1. Updating the URL
# 2. Removing dead links
# 3. Adding to .lycheeignore if intentional
```

### 4. Anchor Validation

**Script:** `scripts/check-anchors.js`  
**Command:** `npm run docs:anchors`

Ensures that all anchor links (e.g., `[text](#heading)`) point to existing headings.

**How It Works:**
1. Extracts all headings from markdown files
2. Converts headings to GitHub-style anchors (lowercase, hyphens, no special chars)
3. Validates that all `#anchor` references exist

**Exclusions:**
- `/vault/**` - HashiCorp Vault configuration
- `/docs/archive/**` - Archived documentation (legacy issues expected)

**Common Fixes:**
```bash
# Find broken anchors
npm run docs:anchors

# Fix by:
# 1. Creating the missing heading
# 2. Updating the anchor reference
# 3. Removing the dead link
```

**Example:**
```markdown
# Database Setup

[See the setup section](#database-setup)  <!-- ✓ Valid -->
[See the config](#database-configuration)  <!-- ✗ Invalid - heading doesn't exist -->
```

### 5. Orphan Detection

**Script:** `scripts/check-orphans.js`  
**Command:** `npm run docs:orphans`

Finds documentation pages that are not reachable from `/docs/index.md` using breadth-first search (BFS).

**How It Works:**
1. Starts from `/docs/index.md`
2. Follows all links (wikilinks and markdown links)
3. Recursively discovers reachable pages
4. Reports pages that cannot be reached

**Exclusions:**
- `/vault/**` - HashiCorp Vault configuration
- `/docs/.obsidian/**` - Obsidian configuration
- `/docs/archive/**` - Archived documentation
- `/docs/adr/**` - Architecture Decision Records (index tracked separately)

**Allowlist:**
- `changelog.md` - Special file, not linked from index
- `contributing.md` - Special file, not linked from index
- `index.md` - Starting point, self-referenced

**Common Fixes:**
```bash
# Find orphaned pages
npm run docs:orphans

# Fix by:
# 1. Adding link from index.md or another page
# 2. Moving to docs/archive/ if deprecated
# 3. Adding to ALLOWLIST in check-orphans.js if intentional
```

**Why This Matters:**
- Ensures all documentation is discoverable
- Prevents abandoned documentation
- Maintains clean information architecture

### 6. Asset Checking

**Script:** `scripts/check-unused-assets.js`  
**Command:** `npm run docs:assets`

Detects unreferenced assets in `/docs/_assets/` and warns about large files (>500KB).

**Checked Asset Types:**
- Images: `.png`, `.jpg`, `.jpeg`, `.gif`, `.svg`, `.webp`, `.ico`
- Documents: `.pdf`
- Videos: `.mp4`, `.webm`

**Exclusions:**
- `/vault/**` - HashiCorp Vault configuration
- `/docs/archive/**` - Archived documentation

**Common Fixes:**
```bash
# Find unused assets
npm run docs:assets

# Fix by:
# 1. Removing unused assets
# 2. Adding references to the assets
# 3. Optimizing large assets (compress images, etc.)
```

## CI/CD Integration

### GitHub Actions Workflow

File: `.github/workflows/docs.yml`

The workflow runs on:
- All PRs that modify documentation
- Pushes to `main` branch

**Trigger Paths:**
- `docs/**`
- `README.md`
- `scripts/check-*.js`
- `.markdownlint.jsonc`
- `.cspell.json`
- `.lycheeignore`
- `package.json`

### Workflow Steps

1. **Checkout** - Clones repository
2. **Setup Node.js** - Installs Node.js 20
3. **Install Dependencies** - Runs `npm ci`
4. **Install Lychee** - Installs link checker
5. **Run All Checks** - Executes all six quality checks
6. **Report Results** - Comments on PR if checks fail

### Failure Handling

The quality checks are handled differently depending on whether they are **blocking** or **non-blocking**:

- **Blocking checks** (linting, anchors, orphans, assets):
  - If any of these checks fail, the workflow fails and blocks PR merge.
  - A comment is posted to the PR with troubleshooting tips.
  - Check logs provide detailed information about failures.
- **Non-blocking checks** (spelling, external link validation):
  - These checks may fail without blocking PR merge.
  - Failures are still reported in the workflow logs and PR comments so they can be addressed.

While only the blocking checks are enforced by CI for merging, contributors are expected to fix issues reported by all checks whenever possible.

## Configuration Files

### `.markdownlint.jsonc`

Markdownlint configuration with Obsidian-compatible settings.

**Key Settings:**
- `MD013`: Line length rule disabled (no enforced line length limit)
- `MD033`: Allow HTML
- `MD041`: Allow frontmatter before first heading
- `MD022`: Allow flexible heading spacing
- `MD050`: Allow mixed emphasis styles
- `MD055`: Allow flexible table pipes

### `.cspell.json`

Spell checker configuration with custom dictionary.

**Key Settings:**
- `language`: "en" (English)
- `ignorePaths`: Excludes vault, .obsidian, node_modules
- `ignoreRegExpList`: Patterns for wikilinks, block refs, tags
- `words`: Custom dictionary of technical terms

### `.lycheeignore`

Link checker exclusion patterns.

**Ignored:**
- Local development URLs (localhost, 127.0.0.1)
- Rate-limited external sites (clips.twitch.tv)

## Troubleshooting

### Spelling Errors

**Problem:** Technical term flagged as misspelled  
**Fix:** Add to `.cspell.json`:
```json
{
  "words": [
    "techterm"
  ]
}
```

### Broken Links

**Problem:** Link returns 404  
**Fix:** Update or remove the link, or add to `.lycheeignore` if it's a known issue.

### Invalid Anchors

**Problem:** Anchor doesn't match heading  
**Fix:** Ensure the anchor matches the heading exactly (GitHub-style: lowercase, hyphens, no special chars).

### Orphaned Pages

**Problem:** Page not reachable from index  
**Fix:** Add a link from index.md or another connected page.

### Unused Assets

**Problem:** Asset not referenced anywhere  
**Fix:** Remove the asset or add a reference to it.

## Best Practices

1. **Run checks locally** before pushing
2. **Fix issues incrementally** - don't wait for CI
3. **Add words to dictionary** when adding technical terms
4. **Keep documentation connected** - all pages should be reachable
5. **Optimize assets** - compress images before committing
6. **Use frontmatter** - helps with organization and searchability

## Related Documentation

- [[obsidian-guide]] - Obsidian vault setup and conventions
- [[contributing]] - General contribution guidelines
- [markdownlint Rules](https://github.com/DavidAnson/markdownlint/blob/main/doc/Rules.md)
- [cspell Documentation](https://cspell.org/)
- [lychee Documentation](https://lychee.cli.rs/)

## Version History

- **1.0** (2026-01-29) - Initial documentation of quality checks
