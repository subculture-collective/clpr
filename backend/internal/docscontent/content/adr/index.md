---
title: "Architecture Decision Records"
summary: "Architecture Decision Records (ADRs) documenting significant architectural decisions made for the Clipper project."
tags: ["adr", "hub", "index", "architecture"]
area: "docs"
status: "stable"
owner: "team-core"
version: "1.0"
last_reviewed: 2026-01-29
aliases: ["adrs", "adr hub", "architecture decisions"]
---

# Architecture Decision Records (ADR)

This directory contains Architecture Decision Records (ADRs) documenting significant architectural decisions made for the Clipper project.

> **Note**: The main decision records are now consolidated in [[../decisions/index|decisions/]] folder. This folder contains legacy ADRs for reference.

## What is an ADR?

An Architecture Decision Record (ADR) is a document that captures an important architectural decision made along with its context and consequences.

## ADR Format

Each ADR follows this structure:

- **Title**: Short descriptive title
- **Status**: Proposed, Accepted, Deprecated, or Superseded
- **Context**: The situation prompting the decision
- **Decision**: The architectural decision made
- **Consequences**: Results and trade-offs of the decision
- **Alternatives Considered**: Other options that were evaluated

## Index of ADRs

```dataview
TABLE title, summary, status, last_reviewed
FROM "docs/adr"
WHERE file.name != "index" AND file.name != "README"
SORT file.name ASC
```

### Available ADRs

| ADR | Title | Status | Date |
|-----|-------|--------|------|
| [[001-semantic-search-vector-db\|ADR-001]] | Semantic Search Vector Database Selection | Accepted | 2025-11-03 |

## Creating a New ADR

1. Copy the template from an existing ADR
2. Number it sequentially (e.g., ADR-002)
3. Fill in all sections with relevant information
4. Update this index with the new ADR
5. Commit and submit for review

---

**See also:**
[[../decisions/index|Current Decision Records]] ·
[[../backend/architecture|Backend Architecture]] ·
[[../index|Documentation Home]]
