---
title: New beans created outside TUI don't appear automatically
status: backlog
type: bug
created_at: 2025-12-12T23:02:58Z
updated_at: 2025-12-12T23:02:58Z
---

## Problem

When the TUI is running, new beans created outside of it (e.g., via CLI in another terminal, or by an agent) don't appear in the TUI until it's restarted.

## Expected Behavior

The TUI should automatically detect new bean files in the `.beans/` directory and add them to the list.

## Likely Cause

This is probably related to how we're watching for file changes. The file watcher may only be watching for modifications to existing files, not for new files being created in the directory.

## Investigation Areas

- [ ] Check how the file watcher is configured in the TUI
- [ ] Verify if we're watching the directory itself vs individual files
- [ ] Ensure CREATE events are being handled, not just MODIFY events