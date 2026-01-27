---
# beans-m4cx
title: Implement git filters in GraphQL ApplyFilter
status: completed
type: bug
priority: normal
created_at: 2026-01-27T06:48:31Z
updated_at: 2026-01-27T07:09:30Z
---

The GraphQL schema defines `hasGitBranch` and `gitBranchMerged` filters in the BeanFilter type, but these filters are not implemented in the `ApplyFilter` function in `internal/graph/filters.go`.

## Current State
- Filters are defined in the schema (`internal/graph/schema.graphqls`)
- Filter fields are in the generated model (`internal/graph/model/models_gen.go`)
- Tests exist but fail because the filters don't actually filter (`internal/graph/schema.resolvers_test.go`)

## Expected Behavior
- `hasGitBranch: true` should return only beans with non-empty `GitBranch` field
- `hasGitBranch: false` should return only beans with empty `GitBranch` field
- `gitBranchMerged: true` should return only beans where `GitMergedAt` is not nil
- `gitBranchMerged: false` should return beans without merged branches (GitMergedAt is nil)

## Implementation
Add filter logic in `internal/graph/filters.go` ApplyFilter function around line 77 (after blocking filters):

```go
// Git filters
if filter.HasGitBranch != nil {
    if *filter.HasGitBranch {
        result = filterByHasGitBranch(result)
    } else {
        result = filterByNoGitBranch(result)
    }
}
if filter.GitBranchMerged != nil {
    if *filter.GitBranchMerged {
        result = filterByGitBranchMerged(result)
    } else {
        result = filterByGitBranchNotMerged(result)
    }
}
```

Then implement the helper functions following the existing pattern.

## Testing
Run: `go test ./internal/graph -run TestQueryBeansFilter.*Git -v`
Both tests should pass after implementation.


## Summary of Changes

Implemented the missing git filter logic in `internal/graph/filters.go`:

- Added `hasGitBranch` filter support in ApplyFilter function (filters.go:80-86)
  - `true`: Returns only beans with non-empty GitBranch field
  - `false`: Returns only beans with empty GitBranch field

- Added `gitBranchMerged` filter support in ApplyFilter function (filters.go:87-93)
  - `true`: Returns only beans where GitMergedAt is not nil (merged branches)
  - `false`: Returns beans where GitMergedAt is nil (unmerged or no git branch)

- Implemented helper functions (filters.go:299-341):
  - `filterByHasGitBranch()`: Filter beans with git branches
  - `filterByNoGitBranch()`: Filter beans without git branches
  - `filterByGitBranchMerged()`: Filter beans with merged branches
  - `filterByGitBranchNotMerged()`: Filter beans without merged branches

All tests now pass:
- `TestQueryBeansFilter_HasGitBranch` ✓
- `TestQueryBeansFilter_GitBranchMerged` ✓
- Full graph package test suite ✓