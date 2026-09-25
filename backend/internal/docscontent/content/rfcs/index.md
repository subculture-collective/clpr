---
title: "Request for Comments (RFCs)"
summary: "Request for Comments (RFCs) for proposing and discussing major features and architectural changes."
tags: ["rfcs", "hub", "index", "proposals"]
area: "docs"
status: "stable"
owner: "team-core"
version: "1.0"
last_reviewed: 2026-01-29
aliases: ["rfc hub", "proposals", "feature proposals"]
---

# Request for Comments (RFCs)

Request for Comments (RFCs) for proposing and discussing major features and architectural changes.

## What is an RFC?

An RFC (Request for Comments) is a detailed proposal for a significant feature, architectural change, or process improvement. RFCs allow the team to:

- Propose and discuss major changes before implementation
- Document the reasoning behind decisions
- Gather feedback from stakeholders
- Build consensus on technical direction

## Documentation Index

```dataview
TABLE title, summary, status, last_reviewed
FROM "docs/rfcs"
WHERE file.name != "index" AND file.name != "README"
SORT file.name ASC
```

## Available RFCs

- [[001-mobile-framework-selection|RFC-001: Mobile Framework Selection]] - React Native + Expo decision (Accepted, 2025-11)
- [[002-advanced-query-language|RFC-002: Advanced Query Language]] - Advanced search syntax (Accepted, 2025-11)
- [[003-infrastructure-modernization|RFC-003: Infrastructure Modernization]] - Infrastructure modernization proposal (Proposed, 2025-11)

| RFC | Title | Status | Date |
|-----|-------|--------|------|
| [[001-mobile-framework-selection\|RFC-001]] | Mobile Framework Selection | Accepted | 2025-11 |
| [[002-advanced-query-language\|RFC-002]] | Advanced Query Language | Accepted | 2025-11 |
| [[003-infrastructure-modernization\|RFC-003]] | Infrastructure Modernization | Proposed | 2025-11 |

## RFC Process

### 1. Draft Phase

- Author writes RFC in this directory
- Include: Problem, Proposal, Alternatives, Open Questions
- Share with team for initial feedback

### 2. Discussion Phase

- RFC is reviewed by relevant stakeholders
- Questions and concerns are addressed
- Alternatives are debated
- RFC may be revised based on feedback

### 3. Decision Phase

- Team decides to Accept, Reject, or Defer
- Decision is documented in RFC status
- If accepted, implementation begins

### 4. Implementation Phase

- RFC guides implementation work
- Implementation may reveal need for RFC updates
- RFC is updated to reflect final design

## RFC Template

```markdown
---
title: "RFC-NNN: Title"
summary: "Brief summary"
tags: ["rfc", "proposal"]
area: "rfcs"
status: "Proposed | Accepted | Rejected | Implemented"
owner: "Author Name"
version: "1.0"
last_reviewed: YYYY-MM-DD
---

# RFC-NNN: Title

**Status**: Proposed | Accepted | Rejected | Implemented
**Author**: Name
**Date**: YYYY-MM-DD
**Related Issues**: Links

## Problem

What problem does this solve?

## Proposal

Detailed proposal including:
- Technical design
- User experience
- API changes
- Migration plan

## Alternatives Considered

What other approaches were evaluated and why were they not chosen?

## Open Questions

What questions remain to be answered?

## Implementation Plan

High-level implementation steps and timeline.

## Success Criteria

How will we know if this is successful?
```

## Creating a New RFC

1. Copy the template above
2. Number it sequentially (e.g., RFC-004)
3. Fill in all sections thoroughly
4. Update this index page
5. Open a PR and tag relevant reviewers
6. Discuss and iterate

---

**See also:**
[[../decisions/index|Architecture Decisions]] ·
[[../product/roadmap|Product Roadmap]] ·
[[../index|Documentation Home]]
