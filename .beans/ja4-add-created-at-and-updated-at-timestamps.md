---
title: Add created_at and updated_at timestamps
status: in-progress
---


Add automatic timestamp tracking to beans via `created_at` and `updated_at` front matter fields.

Implementation:
- Add `created_at` and `updated_at` fields to Bean struct and frontMatter
- Use RFC3339 format (e.g., `2025-01-15T14:30:00Z`)
- Set `created_at` when a bean is first created (in `beans new`)
- Update `updated_at` whenever a bean is saved (in `store.Save`)
- Preserve `created_at` on updates (don't overwrite)
- Support `--sort created` and `--sort updated` in `beans list`
- Include timestamps in JSON output

Example front matter:
```yaml
title: Fix login bug
status: open
created_at: 2025-01-15T14:30:00Z
updated_at: 2025-01-16T09:15:00Z
```

This enables sorting by recency and tracking when work was started/modified.