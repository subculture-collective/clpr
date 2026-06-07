---
title: "Obsidian Setup Guide"
summary: "Complete guide to using the Clipper documentation as an Obsidian vault, including setup, navigation, and best practices."
tags: ["docs", "meta", "obsidian", "guide"]
area: "docs"
status: "stable"
owner: "team-core"
version: "1.0"
last_reviewed: 2026-01-29
aliases: ["obsidian guide", "vault setup"]
---

# Obsidian Setup Guide

This guide explains how to use the Clipper documentation as an Obsidian vault for enhanced navigation, search, and knowledge management.

## Prerequisites

- [Obsidian](https://obsidian.md/) desktop app (free)
- Git clone of the Clipper repository

## Opening the Vault

1. **Clone the Repository** (if not already done):
   ```bash
   git clone https://git.subcult.tv/subculture-collective/clpr.git
   cd clpr
   ```

2. **Open in Obsidian**:
   - Launch Obsidian
   - Click "Open folder as vault"
   - Navigate to `/path/to/clpr/docs`
   - Click "Open"

3. **Trust the Vault**:
   - Obsidian will ask if you trust the vault
   - Click "Trust author and enable plugins" to enable community plugins

## Installed Plugins

The vault comes pre-configured with the following community plugins:

### Core Plugins
- **Dataview**: Query and display documentation dynamically
- **Tag Wrangler**: Manage and organize tags
- **Omnisearch**: Enhanced search across all documents
- **Calendar**: Calendar view for daily notes
- **Table Editor**: Edit markdown tables easily
- **Obsidian Git**: Git integration (optional)
- **Style Settings**: Customize vault appearance
- **Homepage**: Set a custom homepage

### Essential Dataview Features

Dataview queries are used throughout hub pages to automatically list related documentation. For example:

```dataview
TABLE title, summary, status, last_reviewed
FROM "docs/backend"
WHERE file.name != "index"
SORT title ASC
```

This automatically displays all backend documentation with key metadata.

## Navigation Tips

### Quick Navigation

- **⌘/Ctrl + O**: Quick switcher - find any page by title
- **⌘/Ctrl + P**: Command palette - access all commands
- **⌘/Ctrl + Shift + F**: Global search - find text across all docs
- **⌘/Ctrl + G**: Graph view - visualize document connections

### Using Wikilinks

Navigate between pages using wikilinks:
- `[[page-name]]` - Link to a page by name
- `[[page-name|Display Text]]` - Link with custom display text
- `[[page-name#section]]` - Link to a specific section
- `[[../relative/path]]` - Relative path links

### Backlinks

- Open the **Backlinks** pane (right sidebar) to see all pages that link to the current page
- Use backlinks to discover related content

### Tags

- Click any tag (e.g., `#backend`, `#api`) to see all documents with that tag
- Use the **Tag Pane** (right sidebar) to browse all tags
- See [[.obsidian/tag-taxonomy|Tag Taxonomy]] for approved tags

### Graph View

- Press **⌘/Ctrl + G** to open the graph view
- Visualize relationships between documentation pages
- Filter by tags, folders, or search terms
- Zoom and pan to explore document connections

## Creating New Documentation

### Using the Frontmatter Template

1. Create a new markdown file in the appropriate directory
2. Add required frontmatter (see [[.obsidian/templates/frontmatter-template|Frontmatter Template]])
3. Write your content using standard markdown
4. Add wikilinks to related pages
5. Tag appropriately using [[.obsidian/tag-taxonomy|approved tags]]

### Example New Document

```markdown
---
title: "My New Feature"
summary: "Brief description of the new feature documentation"
tags: ["backend", "feature", "api"]
area: "backend"
status: "draft"
owner: "team-core"
version: "1.0"
last_reviewed: 2026-01-29
---

# My New Feature

Content goes here with [[wikilinks]] to related pages.
```

## Frontmatter Reference

All documentation pages must include YAML frontmatter with these required fields:

| Field | Type | Description |
|-------|------|-------------|
| `title` | string | Human-readable page title |
| `summary` | string | One-sentence description |
| `tags` | array | Categorization tags (see [[.obsidian/tag-taxonomy|Tag Taxonomy]]) |
| `area` | string | Documentation section (backend, frontend, mobile, etc.) |
| `status` | string | Document status (draft, review, stable, deprecated, archived) |
| `owner` | string | Responsible team or individual |
| `version` | string | Document version (semantic versioning) |
| `last_reviewed` | date | Last review date (YYYY-MM-DD) |

See [[.obsidian/templates/frontmatter-template|complete template]] for optional fields and examples.

## Search and Discovery

### Omnisearch

The Omnisearch plugin provides powerful search capabilities:
- Full-text search across all documents
- Search in frontmatter fields
- Fuzzy matching
- Search by tags

### Dataview Queries

Hub pages use Dataview queries to automatically list related documentation. Common patterns:

**List all pages in a directory:**
```dataview
LIST
FROM "docs/backend"
WHERE file.name != "index"
SORT title ASC
```

**Table of pages with metadata:**
```dataview
TABLE title, summary, status, last_reviewed
FROM "docs"
WHERE contains(tags, "api")
SORT last_reviewed DESC
```

**Find outdated documentation:**
```dataview
TABLE title, area, last_reviewed
FROM "docs"
WHERE last_reviewed < date(today) - dur(90 days)
SORT last_reviewed ASC
```

## Customization

### Themes

The vault includes several themes in `.obsidian/themes/`. To change:
1. Open Settings → Appearance
2. Select a theme from the dropdown
3. Toggle between light/dark mode

### CSS Snippets

Custom CSS snippets are in `.obsidian/snippets/`:
- Enable/disable in Settings → Appearance → CSS snippets
- Add your own custom styling as needed

### Hotkeys

Customize keyboard shortcuts in Settings → Hotkeys. Recommended additions:
- Toggle reading/editing mode
- Open graph view
- Switch between panes
- Insert template

## Sync and Collaboration

### Git Integration

The Obsidian Git plugin is installed for automatic commits:
1. Configure in Settings → Obsidian Git
2. Set auto-backup interval (e.g., every 30 minutes)
3. Enable auto-pull on startup

### Best Practices

- **Always pull before editing**: Get latest changes from the repository
- **Commit frequently**: Small, focused commits with clear messages
- **Update last_reviewed**: Always update when reviewing/editing a document
- **Follow conventions**: Use the [[.obsidian/templates/frontmatter-template|frontmatter template]] and [[.obsidian/tag-taxonomy|tag taxonomy]]
- **Test wikilinks**: Ensure all links resolve correctly before committing

## Troubleshooting

### Wikilinks Not Resolving

1. Check that the target file exists
2. Verify the file name matches exactly (case-sensitive)
3. Try using relative path: `[[../path/to/file]]`
4. Refresh the index: Reload Obsidian (⌘/Ctrl + R)

### Dataview Queries Not Working

1. Verify Dataview plugin is enabled (Settings → Community Plugins)
2. Check query syntax in Dataview documentation
3. Ensure frontmatter fields exist on target documents
4. Reload Obsidian to refresh queries

### Plugin Issues

1. Update plugins: Settings → Community Plugins → Check for updates
2. Disable/re-enable problematic plugin
3. Check plugin settings for configuration issues
4. Consult plugin documentation or GitHub issues

### Performance Issues

If Obsidian is slow:
1. Disable unused plugins
2. Close graph view when not in use
3. Reduce Dataview query complexity
4. Exclude large directories in Settings → Files & Links
5. Clear cache and restart Obsidian

## Additional Resources

- [[.obsidian/templates/frontmatter-template|Frontmatter Template]] - Required metadata fields
- [[.obsidian/tag-taxonomy|Tag Taxonomy]] - Approved tags and usage guidelines
- [[contributing|Contributing Guide]] - Documentation contribution workflow
- [[index|Documentation Home]] - Main documentation index
- [Obsidian Help](https://help.obsidian.md/) - Official Obsidian documentation
- [Dataview Plugin](https://blacksmithgu.github.io/obsidian-dataview/) - Dataview documentation

## Related Issues

This Obsidian vault setup fulfills the requirements from:
- Issue #803: Docs Structure & Canonical Pages
- Issue #845: Docs Structure & Canonical Pages (duplicate ref)
- Issue #846: Obsidian Frontmatter & Metadata

---

**Last Updated**: 2026-01-29  
**Maintained by**: Team Core
