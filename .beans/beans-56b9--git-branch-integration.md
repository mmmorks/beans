---
# beans-56b9
title: Git-Branch Integration
status: completed
type: feature
priority: normal
created_at: 2026-01-27T05:57:11Z
updated_at: 2026-01-27T06:09:38Z
---

Add comprehensive git-branch integration where parent beans automatically create git branches when transitioning to in-progress.

## Scope
- Auto-create git branches for parent beans (beans with children) on status transition to in-progress
- Track git branch metadata in bean frontmatter (branch name, created_at, merged_at, merge_commit)
- Bidirectional sync: merged branches → completed, deleted branches → scrapped
- New sync command to reconcile git state
- GraphQL extensions for git fields and mutations
- CLI enhancements for branch management

## Architecture
- New internal/gitflow/ package for git operations using go-git
- Core integration with lifecycle hooks
- Branch naming: {bean-id}/{slug}
- Configuration support in .beans.yml

## Checklist
- [x] Phase 1: Core Infrastructure (git operations package, data model)
- [x] Phase 2: Core Integration (lifecycle hooks, hasChildren detection)
- [x] Phase 3: GraphQL Extensions (schema updates, resolvers)
- [x] Phase 4: CLI Commands (update, sync, show, list enhancements)
- [x] Phase 5: Configuration & Documentation (config, tests, docs)