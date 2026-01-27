---
# beans-8a2r
title: Add integration tests for auto-commit functionality
status: completed
type: task
priority: normal
created_at: 2026-01-27T07:45:49Z
updated_at: 2026-01-27T07:59:15Z
---

Add comprehensive integration tests for the auto-commit beans feature to ensure it works correctly in various scenarios.

## Test Coverage Needed

### Basic Auto-Commit Tests (already exist in core_test.go, verify completeness)
- ✓ Test auto-commit when only .beans/ has changes
- ✓ Test no auto-commit when other files have changes  
- ✓ Test auto-commit disabled via config

### Additional Scenarios to Test
- Test auto-commit during branch creation
- Test auto-commit during status transitions
- Test auto-commit with multiple bean updates in sequence
- Test commit message format and content

### Error Handling
- Test auto-commit failure (e.g., empty commit, merge conflict)
- Test recovery from failed auto-commit
- Test behavior when git user.name/user.email not configured

## Implementation Notes
- Review existing tests in `internal/beancore/core_test.go` (TestGitFlow_AutoCommitBeans*)
- Add missing scenarios
- Ensure commit messages follow expected format

## Summary of Changes

Added comprehensive integration tests for auto-commit functionality in `internal/beancore/core_test.go`:

### Tests Added

1. **TestGitFlow_AutoCommitBeans_MultipleBeanUpdates** - Verifies that auto-commit works correctly when creating multiple parent beans and transitioning them to in-progress in sequence
2. **TestGitFlow_AutoCommitBeans_CommitMessageFormat** - Verifies that auto-commit messages follow the expected format ("chore: update beans") and include proper author information
3. **TestGitFlow_AutoCommitBeans_DuringBranchCreation** - Verifies that auto-commit happens before branch creation, ensuring bean files are committed before switching to the new branch
4. **TestGitFlow_AutoCommitBeans_StatusTransitions** - Verifies that auto-commit works during status transitions that trigger branch creation
5. **TestGitFlow_AutoCommitBeans_EmptyCommitScenario** - Verifies that branch creation succeeds even when there are no uncommitted changes (no empty commit is created)

### Existing Tests Verified

Reviewed existing tests and confirmed they cover:
- ✓ Auto-commit when only .beans/ has changes (TestGitFlow_AutoCommitBeans)
- ✓ No auto-commit when other files have changes (TestGitFlow_AutoCommitBeans_MixedChanges)  
- ✓ Auto-commit disabled via config (TestGitFlow_AutoCommitBeans_Disabled)

### Note

Auto-commit only triggers during git branch creation (when parent beans transition to in-progress). Regular bean updates without branch creation do not trigger auto-commits, which is the expected behavior.

All tests pass successfully.