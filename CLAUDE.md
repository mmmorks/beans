# What we're building

You already know what beans is. This is the beans repository.

# Commits

- Use conventional commit messages ("feat", "fix", "chore", etc.) when making commits.
- Include the relevant bean ID(s) in the commit message (please follow conventional commit conventions, e.g. `Refs: bean-xxxx`).
- Mark commits as "breaking" using the `!` notation when applicable (e.g., `feat!: ...`).
- When making commits, provide a meaningful commit message. The description should be a concise bullet point list of changes made.

# Pull Requests

- When we're working in a PR branch, make separate commits, and update the PR description to reflect the changes made.
- Include the relevant bean ID(s) in the PR title (please follow conventional commit conventions, e.g. `Refs: bean-xxxx`).

# Project Specific

- When making changes to the GraphQL schema, run `mise codegen` to regenerate the code.
- The `internal/graph/` package provides a GraphQL resolver that can be used to query and mutate beans.
- All CLI commands that interact with beans should internally use GraphQL queries/mutations.
- `mise build` to build a `./beans` executable

# Git Integration

Beans includes git-branch integration that automatically creates and manages git branches for parent beans (beans with children), following **GitHub Flow** principles.

**Key Features:**
- Auto-creates git branches when parent beans transition to `in-progress` status
- **GitHub Flow**: Always branches from base branch (main), not from HEAD
- Branch naming: `{bean-id}/{slug}` (e.g., `beans-abc123/user-authentication`)
- Bidirectional sync: merged branches → completed status, deleted branches → scrapped status
- Handles squash merges (GitHub default), rebase merges, and fast-forward merges
- Configuration in `.beans.yml` under `beans.git` section

**Commands:**
- `beans sync` - Synchronize bean status with git branch lifecycle (use `--apply` to make changes)
- `beans update <id> --status in-progress` - Auto-creates branch if bean has children

**Technical Details:**
- Git operations are in `internal/gitflow/` package using go-git library
- Core integration hooks in `internal/beancore/core.go` handle status transitions
- GitHub Flow compliance: branches created from base branch, merge detection supports squash/rebase
- Base branch auto-detection: reads `origin/HEAD`, falls back to "main"/"master"
- GraphQL schema includes git fields: `gitBranch`, `gitCreatedAt`, `gitMergedAt`, `gitMergeCommit`
- Git metadata is stored in bean frontmatter

# Extra rules for our own beans/issues

- Use the `idea` tag for ideas and proposals.

# Testing

## Unit Tests

- Always write or update tests for the changes you make.
- Run all tests: `mise test`
- Run specific package: `go test ./internal/bean/`
- Use table-driven tests following Go conventions

## Git Integration Test Coverage

The git integration feature has comprehensive test coverage across multiple layers:

### Test Architecture

1. **Git Operations Layer** (`internal/gitflow/*_test.go`)
   - Unit tests for low-level git operations
   - Tests branch creation, detection, merging, status checks
   - Includes error scenario tests for edge cases

2. **Core Integration Layer** (`internal/beancore/core_test.go`, `core_git_error_test.go`)
   - Tests lifecycle hooks (status transitions triggering branch creation)
   - Tests auto-commit functionality
   - Tests configuration options (enabled/disabled, base branch, etc.)

3. **GraphQL Layer** (`internal/graph/schema.resolvers_test.go`)
   - Tests GraphQL queries with git filters
   - Tests mutations that interact with git (updateBean, syncGitBranches)
   - Tests git field exposure in API

4. **CLI Layer** (`cmd/*_test.go`)
   - Tests beans update command with git integration
   - Tests beans sync command (merged/deleted branch detection)
   - Tests beans list/show with git field display
   - Tests beans query with git filters

### Running Git Integration Tests

```bash
# Run all git-related tests
go test ./internal/gitflow/... ./internal/beancore/... ./internal/graph/... ./cmd/... -run Git -v

# Run specific test suites
go test ./internal/gitflow/... -v                    # Git operations (unit tests)
go test ./internal/beancore/... -run GitFlow -v     # Core integration tests
go test ./internal/graph/... -run Git -v            # GraphQL API tests
go test ./cmd/... -run "Update|Sync|List|Show" -v  # CLI command tests

# Run error scenario tests specifically
go test ./internal/gitflow/... -run Error -v        # Git error handling
go test ./internal/beancore/... -run GitError -v   # Core git errors
```

### Test Coverage Summary

| Feature | Test File | Status |
|---------|-----------|--------|
| Branch creation/naming | `internal/gitflow/git_test.go` | ✅ Passing |
| Branch merge detection | `internal/gitflow/git_test.go` | ✅ Passing |
| Sync functionality | `internal/gitflow/sync_test.go` | ✅ Passing |
| Error scenarios | `internal/gitflow/git_error_test.go` | ✅ Passing |
| Core status transitions | `internal/beancore/core_test.go` | ✅ Passing |
| Auto-commit beans | `internal/beancore/core_test.go` | ✅ Passing |
| Core error scenarios | `internal/beancore/core_git_error_test.go` | ✅ Passing |
| GraphQL git fields | `internal/graph/schema.resolvers_test.go` | ✅ Passing |
| GraphQL mutations | `internal/graph/schema.resolvers_test.go` | ✅ Passing |
| CLI update command | `cmd/update_test.go` | ✅ Passing |
| CLI sync command | `cmd/sync_test.go` | ✅ Passing |
| CLI list/show display | `cmd/list_test.go`, `cmd/show_test.go` | ✅ Passing |
| CLI query filters | `cmd/query_git_test.go` | ✅ Passing |

### Error Scenarios Covered

The test suite includes comprehensive error scenario testing:
- Detached HEAD state handling
- Invalid base branch configuration
- Missing .git directory
- Dirty working tree (uncommitted changes)
- Git integration disabled in config
- Branch already exists
- Nonexistent branches
- Permission errors
- Auto-commit edge cases

### Test Helpers

Key test helper functions (defined in test files):
- `setupTestRepo(t)` - Creates a git repo with initial commit
- `setupTestCoreWithGit(t, cfg)` - Creates beancore with git integration
- `commitFile(t, repo, filename, content, message)` - Makes a commit
- `commitAll(t, repo, message)` - Commits all changes
- `mergeToMain(t, repo, hash)` - Fast-forward merges to main
- `deleteBranch(t, repo, name)` - Deletes a git branch

### Adding New Tests

When adding new git integration tests:
1. **Unit tests** go in `internal/gitflow/*_test.go` for pure git operations
2. **Integration tests** go in `internal/beancore/core_test.go` for bean lifecycle interactions
3. **API tests** go in `internal/graph/schema.resolvers_test.go` for GraphQL layer
4. **CLI tests** go in `cmd/*_test.go` for command-line interface

Always use the test helper functions to set up git repositories and avoid code duplication.

## Manual CLI Testing

- `mise beans` will compile and run the beans CLI. Use it instead of building and running `./beans` manually.
- When testing read-only functionality, feel free to use this project's own `.beans/` directory. But for anything that modifies data, create a separate test project directory. All commands support the `--beans-path` flag to specify a custom path.
