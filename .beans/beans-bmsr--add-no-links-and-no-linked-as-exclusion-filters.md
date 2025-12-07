---
title: Add --no-links and --no-linked-as exclusion filters
status: done
type: feature
created_at: 2025-12-07T11:44:16Z
updated_at: 2025-12-07T11:49:16Z
---


## Summary

Add exclusion filters to `beans list` to find beans that are NOT linked by a relationship type.

## Use Case

Find actionable work — beans not blocked by anything:
```bash
beans list --status open --no-linked-as blocks
```

## Implementation

- Add `--no-links` flag (exclude by outgoing relationship)
- Add `--no-linked-as` flag (exclude by incoming relationship)
- Same format as positive filters: `type` or `type:id`

## Checklist

- [ ] Add flag variables and register in init()
- [ ] Implement excludeByLinks() function
- [ ] Implement excludeByLinkedAs() function  
- [ ] Apply exclusion filters after positive filters
- [ ] Update cmd/prompt.md documentation
- [ ] Add tests