---
# beans-bfsi
title: Improve git integration workflow guidance in beans prime prompt
status: completed
type: task
priority: normal
created_at: 2026-01-27T07:47:50Z
updated_at: 2026-02-02T00:26:41Z
---

The current git integration documentation in the beans prime prompt (system reminder) is good for feature reference, but lacks practical workflow guidance for agents. Enhance it to be more agent-oriented.

## What's Currently Good
- ✓ Features and capabilities documented
- ✓ Commands listed
- ✓ Technical details provided
- ✓ Basic usage examples

## What's Missing

### 1. Workflow Steps
Add a step-by-step workflow guide:
- **Before starting work**: Create parent bean (epic/feature) with child beans
- **Committing beans**: Commit bean files to git before transitioning parent to in-progress
- **Starting work**: Run `beans update <parent-id> -s in-progress` to auto-create branch
- **During work**: Make commits that include both code changes AND bean updates
- **After merging PR**: Run `beans sync --apply` to update bean statuses

### 2. Common Patterns
Document common scenarios:
- How to work with parent beans (epics with children)
- Why parent beans get branches but child beans don't
- When branch creation happens (only on in-progress transition)
- What to do if you want to work without git integration

### 3. Error Handling
Document what to do when:
- "dirty working tree" error → commit your changes first
- Branch creation fails → check git config and working tree
- Sync finds no changes → normal, means no branches merged/deleted

### 4. Best Practices
- Always commit beans before status transitions
- Include bean IDs in commit messages
- Run sync after merging PRs to keep beans in sync
- Use parent/child relationships to get auto-branching

## Implementation Notes
- Update the git integration section in beans prime output
- Move it earlier in the prompt (before "Extra rules" section)
- Add practical examples with actual commands
- Use a "Workflow" subsection that's easy to scan
- Keep existing technical details but make workflow primary focus

## Summary of Changes

Improved the beans prime prompt template based on bd's approach:

1. **Added Context Recovery note** - Tells agent when to re-run `beans prime`
2. **Added Session Close Protocol** - Prominent checklist ensuring work is committed
3. **Reorganized by workflow** - Sections: Finding Work, Creating & Updating, Relationships, Git Integration, Common Workflows
4. **Streamlined git integration** - Focus on workflow steps, not feature documentation
5. **Added Common Workflows section** - Step-by-step practical examples
6. **Added `beans onboard` command** - Minimal snippet for AGENTS.md pointing to `beans prime`
7. **Added `beans prime --minimal` flag** - Compact output (~50 tokens) for context-limited scenarios
8. **Added `beans prime --export` flag** - Output default template for customization
9. **Added comprehensive tests** for template rendering

Files changed:
- `cmd/prompt.tmpl` - Rewritten with improved structure
- `cmd/prompt_minimal.tmpl` - New minimal template
- `cmd/prime.go` - Added --minimal and --export flags
- `cmd/onboard.go` - New onboard command
- `cmd/prime_test.go` - New tests
