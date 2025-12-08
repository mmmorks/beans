---
title: Add priority field
status: backlog
type: feature
tags:
    - schema
created_at: 2025-12-06T22:04:39Z
updated_at: 2025-12-08T14:09:11Z
links:
    - parent: beans-7lmv
---


Add a `priority` field to bean frontmatter with hard-coded allowed values.

## Requirements
- Hard-coded values: `low`, `medium`, `high`, `critical`
- Validation should reject unknown priority values
- Display priority in list/show commands with appropriate styling

## Checklist
- [ ] Add `Priority string` field to Bean struct in `internal/bean/bean.go`
- [ ] Update frontmatter parsing/rendering
- [ ] Add `--priority` flag to `beans create` command
- [ ] Add `--priority` flag to `beans update` command
- [ ] Add `priority` to JSON output
- [ ] Add validation for allowed priority values
- [ ] Unit tests for priority field handling

## Context
Part of the issue metadata expansion. See original planning bean: beans-v8qj