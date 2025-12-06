---
title: Implement TUI with bubbletea
status: done
created_at: 2025-12-06T19:07:38Z
updated_at: 2025-12-06T19:09:18Z
---


Add a `beans tui` command that provides an interactive terminal UI.

## Features
- List view showing all beans
- Press Enter to view bean details
- Navigation with j/k keys
- Quit with q

## Files to Create
1. internal/tui/keys.go - Key bindings
2. internal/tui/styles.go - TUI-specific styles
3. internal/tui/list.go - List view using bubbles/list
4. internal/tui/detail.go - Detail view with viewport + glamour
5. internal/tui/tui.go - Main app model
6. cmd/beans/tui.go - Cobra command