# Beans - Agentic Issue Tracker

This project uses **beans**, an agentic-first issue tracker. Issues are called "beans", and you can
use the "beans" CLI to manage them.

All commands support --json for machine-readable output. Use this flag to parse responses easily.

## CRITICAL: Track All Work With Beans

**BEFORE starting any task the user asks you to do:**

1. FIRST: Create a bean with `beans create "Title" -d "Description..." -s in-progress --no-edit`
2. THEN: Do the work
3. FINALLY: Mark done with `beans update <bean-id> --status done` (do this before committing)

**Do NOT use the TodoWrite tool or markdown TODOs.** Use beans for all task tracking.

If you identify something that should be tracked during your work, create a bean for it.

## Core Rules

- After compaction or clear, run `beans prompt` to re-sync
- All bean commands support the `--json` flag for machine-readable output.
- Lean towards using sub-agents for interacting with beans.
- You can inspect `.beans/config.yaml` to learn about the different issue types and statuses configured for this project.

## Finding work

- `beans list --json` to list all beans (descriptions not included by default)
- `beans list --json --full` to include full description content

## Working on a bean

- `beans update <bean-id> --status in-progress --json` to mark a bean as in-progress
- `beans show <bean-id> --json` to see full details including description
- Adhere to the instructions in the bean's description when working on it

## Creating new beans

- `beans create --help`
- Example: `beans create "Fix login bug" -d "Users cannot log in when..." -s open --no-edit`
- When creating new beans, include a useful description. If you're not sure what to write, ask the user.
- Make the description as detailed as possible, similar to a plan that you would create for yourself.
- If possible, split the work into a checklist of GitHub-Formatted-Markdown tasks. Use a `## Checklist` header to precede it.

## Cleaning up beans

- `beans archive` will archive (delete) beans marked as done.
