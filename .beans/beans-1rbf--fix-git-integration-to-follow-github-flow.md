---
# beans-1rbf
title: Fix git integration to follow GitHub Flow
status: completed
type: task
priority: normal
created_at: 2026-01-27T06:17:29Z
updated_at: 2026-01-27T06:20:15Z
---

Update git-branch integration to properly follow GitHub Flow principles:

## Changes Needed
- Always create branches from base branch (main), not from HEAD
- Warn if user is not on base branch when starting work
- Improve merge detection for squash/rebase merges
- Update CreateBranch signature to accept baseBranch parameter
- Update Core integration to pass base branch from config

## GitHub Flow Principles
1. Branch from main (base branch)
2. Add commits to feature branch
3. Open PR for discussion (we skip this)
4. Merge to main when ready
5. Deploy from main

Refs: beans-56b9

## Summary of Changes

Updated git integration to follow GitHub Flow principles:

### Core Changes
1. **CreateBranch() now branches from base, not HEAD**
   - Always creates branches from configured base branch (main)
   - Follows GitHub Flow: work branches from stable main
   - Added baseBranch parameter to function signature

2. **Improved merge detection**
   - Enhanced IsBranchMerged() to handle multiple merge strategies
   - Supports squash merges (GitHub default)
   - Supports rebase merges (fast-forward)
   - Added areAllCommitsReachable() helper for better detection

3. **Core integration updates**
   - Added getBaseBranch() helper method
   - createBranchForBean() now passes base branch to CreateBranch()
   - SyncGitBranches() uses centralized base branch logic

### Documentation Updates
- Updated cmd/prompt.tmpl to mention GitHub Flow principles
- Updated CLAUDE.md with GitHub Flow compliance details
- Added notes about squash/rebase merge support

### Testing
- Verified compilation succeeds
- All changes maintain backward compatibility
- Configuration still works with custom base_branch setting