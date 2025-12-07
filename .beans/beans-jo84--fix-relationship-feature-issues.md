---
title: Fix relationship feature issues
status: done
created_at: 2025-12-07T11:09:47Z
updated_at: 2025-12-07T11:11:35Z
---


## Tasks

- [ ] Add `--link` filter to list command (opposite of `--linked`, filters by active/outgoing links)
- [ ] Sort link types for deterministic display order in show command
- [ ] Use shared `KnownLinkTypes` constant in update command
- [ ] Clean up the beans-h0nh tracking bean's stale content