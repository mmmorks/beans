---
title: Investigate and improve Claude Code hooks integration
status: open
created_at: 2025-12-06T18:35:29Z
updated_at: 2025-12-06T18:35:29Z
---

Currently requires verbose JSON configuration in .claude/settings.json with PreCompact and SessionStart hooks. Explore better mechanisms for hooking into Claude Code.

## Current approach

```json
"hooks": {
  "PreCompact": [
    {
      "matcher": "",
      "hooks": [
        {
          "type": "command",
          "command": "beans prompt"
        }
      ]
    }
  ],
  "SessionStart": [
    {
      "matcher": "",
      "hooks": [
        {
          "type": "command",
          "command": "beans prompt"
        }
      ]
    }
  ]
}
```

## Investigation areas

- Can we simplify the hook setup?
- Is there a way to auto-configure this?
- Could `beans init` set this up automatically?
- Are there alternative integration points we should explore?
- Should we support a `beans hooks` command that outputs the required config?