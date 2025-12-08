---
title: Add priority field
status: completed
type: feature
tags:
    - schema
created_at: 2025-12-06T22:04:39Z
updated_at: 2025-12-08T17:38:41Z
links:
    - parent: beans-7lmv
---

Add a 5-level priority system to beans: `critical`, `high`, `normal`, `low`, `deferred`.

## Requirements
- Priority is optional (beans without priority are treated as `normal` for sorting)
- Hard-coded values: `critical`, `high`, `normal`, `low`, `deferred`
- Validation should reject unknown priority values
- Status sorts first, then priority within each status group
- Display priority in list/show commands with colored styling

## Checklist
- [x] Add PriorityConfig and validation in `internal/config/config.go`
- [x] Add `Priority` field to Bean struct in `internal/bean/bean.go`
- [x] Update frontmatter parsing/rendering
- [x] Update sorting to include priority in `internal/bean/sort.go`
- [x] Add `--priority` flag to `beans create` command
- [x] Add `--priority` flag to `beans update` command
- [x] Add `--priority` / `--no-priority` filtering to `beans list` command
- [x] Display priority in `beans show` command
- [x] Update agent prompt in `cmd/prompt.go`
- [x] Add tests for priority functionality

## Priority Colors
| Priority | Color |
|----------|-------|
| critical | red |
| high | yellow |
| normal | white |
| low | gray |
| deferred | gray/dim |