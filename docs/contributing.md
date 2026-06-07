---
title: "Contributing"
summary: "How to contribute to the Clipper project."
tags: ["docs", "contributing", "guide"]
area: "docs"
status: "stable"
owner: "team-core"
version: "1.0"
last_reviewed: 2026-01-29
aliases: ["contribute", "contributing guide"]
---

# Contributing to Clipper

See the main [Contributing Guide](../CONTRIBUTING.md) in the repository root for:

- Development workflow
- Code standards
- Testing requirements
- Pull request process
- Code of Conduct

## Documentation Contributions

To improve these docs:

1. Edit markdown files in `/docs/`
2. Add required frontmatter using the fields listed below
3. Follow [[index|documentation conventions]]
4. Run validation: `npm run docs:check`
5. Submit PR with `documentation` label

### Frontmatter Requirements

All documentation pages must include YAML frontmatter with:
- `title` - Human-readable page title
- `summary` - One-sentence description
- `tags` - Array of categorization tags
- `area` - Documentation section (backend, frontend, mobile, etc.)
- `status` - Document status (draft, review, stable, deprecated, archived)
- `owner` - Responsible team or individual
- `version` - Document version
- `last_reviewed` - Last review date (YYYY-MM-DD)

Keep frontmatter concise and aligned with nearby documentation pages.

## Quick Links

- [[setup/development|Development Setup]]
- [[backend/testing|Testing Guide]]
- [[index|Documentation Index]]

---

[[index|← Back to Index]]
