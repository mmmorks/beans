---
title: Add issue relationships
status: done
type: feature
created_at: 2025-12-06T22:04:39Z
updated_at: 2025-12-07T11:01:22Z
---

Add relationship fields to beans for expressing dependencies and connections.

## Implementation

- Links stored as `links: {type: [ids]}` in YAML frontmatter
- Supports flexible input: `blocks: abc` or `blocks: [abc, def]`
- Relationship types: `blocks`, `duplicates`, `parent`, `relates-to`
- CLI: `--link`/`--unlink` on update command
- CLI: `--link` (outgoing) and `--linked` (incoming) filters on list command

## Context

Part of the issue metadata expansion. See original planning bean: beans-v8qj
