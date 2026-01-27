---
# beans-f8je
title: Add CLI tests for beans update with git integration
status: completed
type: task
priority: normal
created_at: 2026-01-27T07:45:48Z
updated_at: 2026-01-27T07:55:18Z
---

Add end-to-end CLI tests for `beans update` command to verify git integration hooks work correctly.

## Test Coverage Needed

### Status Transition Tests
- Test `beans update <id> -s in-progress` creates git branch for parent beans
- Test `beans update <id> -s in-progress` does NOT create branch for non-parent beans
- Test `beans update <id> -s completed` when branch exists but not merged (should fail if require_merge is true)
- Test multiple status transitions on the same bean

### Error Handling Tests
- Test update when working tree is dirty (should fail or auto-commit based on config)
- Test update when git integration is disabled
- Test update when branch creation fails

### Output Tests
- Test that git branch info is displayed in success messages
- Test `--json` output includes git fields

## Implementation Notes
- Create new test file `cmd/update_test.go` or add to existing
- Use CLI command execution (not just GraphQL mutation calls)
- Test both human-readable and JSON output

## Summary of Changes

Added comprehensive CLI tests for `beans update` command with git integration in `cmd/update_test.go`:

### Tests Added

1. **TestUpdateCommand_GitBranchAutoCreate_ParentBean** - Verifies that updating a parent bean to in-progress creates a git branch
2. **TestUpdateCommand_GitBranchNoAutoCreate_NonParentBean** - Verifies that non-parent beans do NOT get branches
3. **TestUpdateCommand_GitDisabled** - Verifies that updates work correctly when git integration is disabled
4. **TestUpdateCommand_DirtyWorkingTree** - Verifies that branch creation fails with appropriate error when working tree has uncommitted changes
5. **TestUpdateCommand_MultipleStatusTransitions** - Verifies that multiple status transitions on the same bean work correctly (branch persists)
6. **TestUpdateCommand_JSONOutput** - Verifies that JSON output includes git fields
7. **TestUpdateCommand_ResponseOutput** - Verifies that CLI response output includes git fields

### Test Infrastructure

- Added `setupUpdateTestEnv()` helper that sets up git repository with beans core
- Added `setupUpdateTestEnvNoGit()` helper for testing without git integration
- Tests use GraphQL resolver layer (same pattern as sync_test.go) since CLI commands internally use GraphQL

All tests pass successfully.