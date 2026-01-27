---
# beans-gjv8
title: Document git integration test coverage
status: completed
type: task
priority: low
created_at: 2026-01-27T06:49:33Z
updated_at: 2026-01-27T08:15:30Z
---

Create comprehensive documentation explaining the test coverage for the GitHub Flow git integration feature.

## What to Document

### 1. Test Architecture Overview
Explain the multi-layer testing approach:
- **Unit tests** (internal/gitflow): Git operations in isolation
- **Integration tests** (internal/beancore): Core lifecycle hooks
- **GraphQL tests** (internal/graph): API layer
- **CLI tests** (cmd): End-to-end command testing

### 2. Running Tests
Document how to run tests at each layer:
```bash
# All git-related tests
go test ./internal/gitflow ./internal/beancore -run Git -v
go test ./internal/graph -run Git -v

# Specific test suites
go test ./internal/gitflow -v                    # Git operations
go test ./internal/beancore -run TestGitFlow -v  # Core integration
go test ./internal/graph -run TestBeanGit -v     # GraphQL fields
go test ./cmd -run TestSync -v                   # CLI sync command
```

### 3. Test Coverage Summary
Create a table showing:
- What's tested
- Where it's tested  
- Current status (passing/failing/skipped)
- Line numbers for reference

### 4. Known Issues
Document:
- Missing git filter implementation in ApplyFilter
- Core integration tests needing debugging
- CLI tests marked as skipped

### 5. Adding New Tests
Guide for future test additions:
- When to add unit vs integration tests
- How to use the test helpers (setupTestRepo, commitAll, etc.)
- Common patterns and gotchas

## Suggested Location
Either:
- Add to CLAUDE.md under a "Testing" section
- Create docs/testing.md
- Add to README.md

## Success Criteria
- Clear overview of test architecture
- Commands to run each test suite
- Known issues documented
- Easy for contributors to understand test coverage

## Summary of Changes

Added comprehensive git integration test documentation to CLAUDE.md under the Testing section.

### Documentation Added

The new "Git Integration Test Coverage" section includes:

1. **Test Architecture Overview** - Explains the multi-layer testing approach across gitflow, beancore, GraphQL, and CLI layers

2. **Running Tests** - Provides specific commands to run git-related tests at each layer

3. **Test Coverage Summary** - Table showing all tested features, their test files, and current status (all passing)

4. **Error Scenarios Covered** - Lists the comprehensive error scenarios that are tested

5. **Test Helpers** - Documents the key test helper functions available for writing new tests

6. **Adding New Tests** - Guide for where to add new tests based on what layer is being tested

All existing and new git integration tests are passing. The documentation provides a clear reference for understanding the test coverage and for contributors adding new tests.