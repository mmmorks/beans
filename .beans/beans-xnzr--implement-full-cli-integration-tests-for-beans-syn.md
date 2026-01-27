---
# beans-xnzr
title: Implement full CLI integration tests for beans sync
status: todo
type: task
priority: low
created_at: 2026-01-27T06:49:12Z
updated_at: 2026-01-27T06:49:12Z
---

Add end-to-end integration tests for the `beans sync` command in `cmd/sync_test.go`.

## Current State
- Test file exists with skeleton structure and skip placeholders
- Tests are marked as `t.Skip("Requires full git integration test setup")`
- Underlying functionality is tested at lower layers (gitflow, beancore, graph)

## Scenarios to Test

### 1. Dry-run mode (default)
```bash
beans sync
```
- Should display preview of changes
- Should NOT actually update bean status
- Should show "Run with --apply to update beans" message

### 2. Apply mode
```bash
beans sync --apply
```
- Should actually update bean status
- Should report number of synced beans
- Should persist changes to disk

### 3. JSON output
```bash
beans sync --json
beans sync --apply --json
```
- Should output valid JSON
- Should include updated bean IDs and new statuses

### 4. Error cases
- Git integration not enabled → helpful error message
- No git repository → helpful error message  
- No beans with git branches → empty result, not an error

### 5. Mixed branch states
Create test scenario with:
- Bean with merged branch → should become completed
- Bean with deleted branch → should become scrapped
- Bean with active branch → no change

## Implementation Notes
- Tests need full git repository setup (like core_test.go)
- Can leverage `setupTestCoreWithGit` pattern
- May need to execute actual CLI commands or test the sync function directly
- Consider using golden files for output validation

## Success Criteria
- Remove all `t.Skip()` calls
- All test cases pass
- Coverage for both dry-run and apply modes
- Both human-readable and JSON output validated