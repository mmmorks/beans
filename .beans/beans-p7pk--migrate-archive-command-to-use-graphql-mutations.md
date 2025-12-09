---
title: Migrate 'archive' command to use GraphQL mutations
status: scrapped
type: task
priority: normal
created_at: 2025-12-09T12:04:26Z
updated_at: 2025-12-09T13:18:36Z
links:
    - blocks: beans-wp2o
    - parent: beans-7ao1
---

## Summary

Migrate the `beans archive` command to use GraphQL instead of directly calling core methods.

## Current Implementation

In `cmd/archive.go`:
- Line 25: `core.All()` - fetch all beans
- Line 48: `core.FindIncomingLinks(id)` - for each archivable bean
- Line 97: `core.RemoveLinksTo(id)` - remove incoming links
- Line 110: `core.Delete(id)` - delete the bean

## Target Implementation

Once GraphQL mutations are available (beans-wp2o):
```go
resolver := &graph.Resolver{Core: core}

// Query all beans with completed/scrapped status
filter := &model.BeanFilter{
  Status: []string{"completed", "scrapped"},
}
beans, err := resolver.Query().Beans(ctx, filter)

// For each bean, use mutations to clean up and delete
for _, b := range beans {
  resolver.Mutation().RemoveLinksTo(ctx, b.ID)
  resolver.Mutation().DeleteBean(ctx, b.ID)
}
```

## Blocked By

- beans-wp2o (Add GraphQL mutations for bean CRUD operations)

## Checklist

- [ ] Wait for GraphQL mutations to be implemented
- [ ] Replace `core.All()` with filtered GraphQL query
- [ ] Replace `core.FindIncomingLinks()` with GraphQL (or keep direct)
- [ ] Replace `core.RemoveLinksTo()` with mutation
- [ ] Replace `core.Delete()` with mutation
- [ ] Run tests