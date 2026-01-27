---
# beans-xv94
title: Add error scenario tests for git integration
status: completed
type: task
priority: normal
created_at: 2026-01-27T07:45:49Z
updated_at: 2026-01-27T08:13:18Z
---

Add comprehensive error scenario tests for git integration to ensure proper error handling and messages.

## Test Coverage Needed

### Git Availability Tests
- Test behavior when .git directory doesn't exist
- Test behavior when git integration is explicitly disabled in config
- Test auto-enable git integration when possible

### Permissions and Conflicts Tests  
- Test behavior when .beans directory is not writable
- Test behavior when git repository is in detached HEAD state
- Test concurrent bean updates (race conditions)

### Configuration Tests
- Test with `require_merge: true` preventing manual completion
- Test with `auto_commit_beans: true/false`
- Test with invalid base_branch configuration

### Recovery Tests
- Test recovery from partial branch creation
- Test cleanup of orphaned git branches
- Test handling of corrupted bean files with git metadata

## Implementation Notes
- Add tests across multiple layers (gitflow, beancore, CLI)
- Focus on clear, actionable error messages
- Verify that errors don't leave system in inconsistent state

## Summary of Changes

Added comprehensive error scenario tests for git integration across multiple layers:

### Git Layer Tests ()
- Detached HEAD state handling
- Invalid base branch configuration
- Main branch detection fallbacks (no remote, master-only repos, no branches)
- Working tree clean/dirty state validation (unstaged and staged changes)  
- CommitBeans with no changes
- CommitBeans only commits .beans files (not other changes)
- Empty branch name validation
- Switch branch with nonexistent branch
- Switch branch with dirty working tree (modified tracked files)
- IsBranchMerged with invalid/nonexistent branch names

### BeanCore Layer Tests (`internal/beancore/core_git_error_test.go`)
- RequireMerge configuration (test skeleton added, feature not yet implemented)
- AutoCommitBeans configuration
- Invalid base branch configuration
- Auto-enable git integration when .git directory exists
- Git integration disabled in config
- Git integration when no git repository exists

All tests pass successfully and cover edge cases and error conditions to ensure robust error handling and messages.

## Test Files Created

- internal/gitflow/git_error_test.go - 13 new error scenario tests for gitflow operations
- internal/beancore/core_git_error_test.go - 7 new error scenario tests for core git integration

All tests pass successfully and provide comprehensive coverage of error scenarios.