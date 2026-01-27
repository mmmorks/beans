---
# beans-aei5
title: Add CLI tests for git field display in list/show commands
status: completed
type: task
priority: normal
created_at: 2026-01-27T07:45:48Z
updated_at: 2026-01-27T08:03:35Z
---

Add tests to verify that `beans list` and `beans show` correctly display git integration fields.

## Test Coverage Needed

### beans list Tests
- Test that beans with git branches show branch name in output
- Test `--json` output includes git fields (gitBranch, gitCreatedAt, gitMergedAt, gitMergeCommit)
- Test filtering by git status (already has GraphQL tests, but need CLI tests)

### beans show Tests
- Test that git section is displayed for beans with git branches
- Test that git fields are omitted for beans without git integration
- Test `--json` output includes all git fields with correct formatting

### beans query Tests
- Test GraphQL queries with git filters work from CLI
- Test example queries from documentation work correctly

## Implementation Notes
- Add to existing test files (cmd/list_test.go, cmd/show_test.go)
- Create actual beans with git metadata
- Verify both human-readable and JSON output

## Summary of Changes

Added comprehensive CLI tests for git field display in list/show/query commands:

### Files Created

1. **cmd/show_test.go** (new file)
   - TestShowCommand_GitFields_WithBranch - Verifies git fields are present for beans with branches
   - TestShowCommand_GitFields_WithoutBranch - Verifies git fields are empty for beans without branches
   - TestShowCommand_JSONOutput_WithGitFields - Verifies JSON output includes git fields
   - TestShowCommand_JSONOutput_WithoutGitFields - Verifies JSON output handles beans without git fields
   - TestShowCommand_HumanReadableOutput - Verifies human-readable output displays git information
   - TestShowCommand_MultipleBeansWithMixedGitStatus - Verifies mixed scenarios

2. **cmd/query_git_test.go** (new file)
   - TestQueryCommand_GitBranchFilter - Tests hasGitBranch filter
   - TestQueryCommand_GitBranchMergedFilter - Tests gitBranchMerged filter
   - TestQueryCommand_CombinedGitAndStatusFilters - Tests combining git and status filters
   - TestQueryCommand_GitFieldsInResult - Verifies all git fields are present in query results
   - TestQueryCommand_DocumentationExampleQueries - Tests example queries from documentation

### Files Modified

3. **cmd/list_test.go**
   - TestListCommand_GitFieldsInBeans - Verifies beans with git fields are handled correctly in lists
   - TestListCommand_SortBeansWithGitFields - Verifies sorting works with beans that have git fields

### Coverage

All tests verify both human-readable and JSON output formats, covering:
- Git branch display
- Git creation/merge timestamps
- Git merge commits
- Filtering by git status (hasGitBranch, gitBranchMerged)
- Combined filters (git + status)
- Edge cases (beans without git, mixed scenarios)

All tests pass successfully.