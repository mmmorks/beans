---
# beans-vjmq
title: Debug and fix core git integration tests
status: completed
type: task
priority: normal
created_at: 2026-01-27T06:48:52Z
updated_at: 2026-01-27T07:01:49Z
---

The git integration tests in `internal/beancore/core_test.go` are failing because auto-branch creation is not working in the test environment.

## Problem
Tests like `TestGitFlow_AutoCreateBranch_ParentBean` create parent beans and transition them to in-progress, expecting a git branch to be auto-created, but:
- `parent.GitBranch` remains empty after update
- `parent.GitCreatedAt` is nil
- The git branch is not created in the repository

## Possible Causes
1. **Config not properly set**: Git integration config might not be applied correctly
2. **Git flow not enabled**: `EnableGitFlow` might not be working in tests
3. **Working tree check**: Tests commit beans but something might still make the tree "dirty"
4. **hasChildren detection**: The `hasChildren` check might not find the child bean
5. **Timing issue**: The child might not be persisted when the parent is updated

## Investigation Steps
- [x] Add debug logging to `handleGitTransition` in core.go
- [ ] Verify `IsGitFlowEnabled()` returns true in test
- [ ] Verify `hasChildren()` returns true for the parent bean
- [ ] Check if `IsWorkingTreeClean()` passes in test
- [ ] Verify bean files are actually committed before transition

## Test Files
- `internal/beancore/core_test.go`: Lines 1653-1740 (TestGitFlow_AutoCreateBranch_ParentBean)
- Other failing tests follow the same pattern

## Expected Outcome
All `TestGitFlow_*` tests should pass, validating:
- Auto-branch creation for parent beans
- Branch creation from base branch (GitHub Flow)
- Rejection when working tree is dirty
- Sync operations for merged/deleted branches
- Configuration options

## Summary of Changes

The git integration tests were failing because of two main issues:

### Issue 1: Missing AutoCreateBranch config check
The `handleGitTransition` function was checking if git flow was enabled, but not checking the `AutoCreateBranch` config setting. Added a check for `c.config.Beans.Git.AutoCreateBranch` before attempting to create branches.

### Issue 2: Bean state comparison bug  
When users called `core.Get()` to retrieve a bean, modified it, and passed it to `core.Update()`, both the "old" and "new" bean references pointed to the same object. This meant status transitions weren't detected.

**Solution:** Modified the `Update` method to reload the bean's old state from disk before comparing states. This ensures we always have the true previous state for git transition detection.

### Issue 3: Test pattern updates
The tests needed to:
1. Commit bean files to git before calling Update (to ensure clean working tree)
2. Reload beans after Update to get updated git fields (like GitBranch)  
3. Commit bean files after each Update when testing multiple transitions

All `TestGitFlow_*` tests now pass successfully.

## Files Modified
- `internal/beancore/core.go`: Fixed Update method to reload old state from disk, added AutoCreateBranch config check
- `internal/beancore/core_test.go`: Updated tests to follow correct git workflow patterns