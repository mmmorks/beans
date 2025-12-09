---
title: GraphQL Migration Epic
status: completed
type: epic
priority: normal
created_at: 2025-12-09T12:04:53Z
updated_at: 2025-12-09T13:18:40Z
---

## Summary

Migrate all CLI commands to use GraphQL internally for data access. This ensures a consistent data access pattern throughout the codebase and enables future features like remote beans servers.

## Current State

Already migrated to GraphQL:
- ✅ `beans list` - uses GraphQL resolver
- ✅ `beans query` - is the GraphQL query interface
- ✅ `beans tui` - uses GraphQL resolver

## Migration Plan

### Phase 1: Query-Only Commands (Can migrate now)
These commands only read data and can use existing GraphQL queries:
- `beans show` - needs `Query.Bean()`
- `beans roadmap` - needs `Query.Beans()`
- `beans check` - needs new query for broken links (optional)

### Phase 2: Mutation Infrastructure
Before migrating write commands, we need:
- Add `Mutation` type to GraphQL schema
- Implement `createBean`, `updateBean`, `deleteBean` mutations
- Add `removeLinksTo` mutation for link cleanup

### Phase 3: Write Commands (Blocked on mutations)
These commands require GraphQL mutations:
- `beans create` - needs `createBean` mutation
- `beans update` - needs `updateBean` mutation
- `beans delete` - needs `deleteBean` + `removeLinksTo` mutations
- `beans archive` - needs `deleteBean` + `removeLinksTo` mutations

### Phase 4: Special Cases
- `beans init` - may stay as direct beancore call (bootstrapping)
- `beans prompt` - utility command, no data mutation needed
- `beans version` - no data access

## Benefits of Migration

1. **Consistency** - All data access through one interface
2. **Testability** - GraphQL layer can be mocked/tested independently
3. **Future-proofing** - Enables remote beans server scenarios
4. **Type safety** - GraphQL provides strong typing