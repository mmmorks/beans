---
title: Migrate 'check' command to use GraphQL
status: scrapped
type: task
priority: normal
created_at: 2025-12-09T12:04:04Z
updated_at: 2025-12-09T13:18:36Z
links:
    - parent: beans-7ao1
---

## Summary

Migrate the `beans check` command to use GraphQL instead of directly calling core link checking methods.

## Current Implementation

In `cmd/check.go`:
- Line 114: `core.CheckAllLinks()`
- Line 118: `core.FixBrokenLinks()`

## Considerations

The `check` and `fix` operations are specialized link integrity operations. Options:

1. **Add to GraphQL as queries/mutations:**
   - `checkLinks` query that returns broken links
   - `fixBrokenLinks` mutation

2. **Keep as beancore calls but standardize access:**
   - These are administrative/maintenance operations
   - May be acceptable to keep outside GraphQL

## Recommendation

Add GraphQL support for these operations to maintain consistency:

```graphql
type Query {
  """Check for broken links across all beans"""
  checkLinks: [BrokenLink!]!
}

type Mutation {
  """Remove all broken links"""
  fixBrokenLinks: Int!
}

type BrokenLink {
  beanId: String!
  linkType: String!
  targetId: String!
}
```

## Checklist

- [ ] Decide: Add to GraphQL or keep as direct core calls
- [ ] If GraphQL: Add types to schema
- [ ] If GraphQL: Implement resolvers
- [ ] Update check command to use chosen approach
- [ ] Run tests