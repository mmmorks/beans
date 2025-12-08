---
title: Remove creation TUI from beans create command
status: completed
type: task
created_at: 2025-12-08T21:57:13Z
updated_at: 2025-12-08T22:01:38Z
---

Remove the interactive form and editor launching from `beans create`. The creation TUI should be part of `beans tui` instead, not the create command itself.

## Context

GitHub issue #8 reports that agents encounter 'Vim: Warning: Output is not to a terminal' errors when using `beans create` without `--no-edit`. The resolution is to remove the TUI entirely from create and make it a pure non-interactive command.

## Changes

- Remove the interactive form (huh form) from create.go
- Remove the $EDITOR launching logic
- Remove the `--no-edit` flag (it becomes the default behavior)
- Update prompt.md to remove reference to `--no-edit`