---
title: Fix --linked-as filter logic for type:id format
status: completed
type: bug
created_at: 2025-12-08T11:12:25Z
updated_at: 2025-12-08T11:15:24Z
---

## Problem

The `--linked-as type:id` filter has inverted logic. Currently:

- `--linked-as parent:beans-58hm` checks if `beans-58hm` has outgoing parent links to the bean
- But it SHOULD find beans that have `parent: beans-58hm` in their own links (beans that link TO beans-58hm)

## Expected behavior

`--linked-as parent:milestone-id` should return all beans that have `parent: milestone-id` in their links - i.e., children of the milestone.

## Current behavior  

It returns nothing because it's checking the wrong direction - looking for outgoing links FROM the milestone instead of incoming links TO the milestone.

## Fix

In `filterByLinkedAs` (cmd/list.go:400-434), the type:id case should check if the current bean (`b`) has a link of `type` pointing to `targetID`, not the reverse.

Change:
```go
source, exists := idx.byID[f.targetID]
if source.Links.HasLink(f.linkType, b.ID) { ... }
```

To:
```go  
if b.Links.HasLink(f.linkType, f.targetID) { ... }
```

Same fix needed for `excludeByLinkedAs`.