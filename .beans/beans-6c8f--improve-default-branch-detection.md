---
# beans-6c8f
title: Improve default branch detection
status: completed
type: task
created_at: 2026-01-27T06:23:03Z
updated_at: 2026-01-27T06:23:03Z
---

Enhanced GetMainBranch() to automatically infer the default branch from git instead of just guessing.

## Implementation
1. Read origin/HEAD (remote's default branch) - most reliable
2. Fall back to common names (main, master)
3. Last resort: use first available branch

This respects what the remote repository considers the default branch, which is more reliable than hardcoding 'main' or 'master'.

## Benefits
- Works with custom default branch names
- Respects remote repository configuration
- No manual configuration needed in most cases
- Still allows override via base_branch config

Refs: beans-1rbf