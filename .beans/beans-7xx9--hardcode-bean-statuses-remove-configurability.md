---
title: Hardcode bean statuses (remove configurability)
status: completed
type: task
created_at: 2025-12-08T13:07:39Z
updated_at: 2025-12-08T13:26:36Z
links:
    - parent: beans-58hm
---

## Summary

Remove the configurability of bean statuses from config.yaml and hardcode the following statuses:

- **not-ready**: Not yet ready to be worked on (blocked, needs more info, etc.)
- **ready**: Ready to be worked on
- **in-progress**: Currently being worked on
- **completed**: Finished successfully
- **scrapped**: Will not be done

## Rationale

Simplifies the system by removing unnecessary configurability. The hardcoded statuses cover all workflow states.

## Checklist

- [x] Update `internal/config/config.go` to hardcode statuses instead of reading from config
- [x] Remove statuses section from config.yaml handling
- [x] Update `beans init` to not create statuses in config (init saves config without statuses now)
- [x] Update any validation logic to use hardcoded statuses
- [x] Remove any `beans statuses` command if planned (there wasn't one)
- [x] Update prompt.md to reflect the hardcoded statuses
- [x] Update tests
- [x] Consider migration path for existing beans using old status names (open → ready, done → completed) - updated all existing beans
